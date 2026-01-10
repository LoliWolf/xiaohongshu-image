package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/xiaohongshu-image/internal/api"
	"github.com/xiaohongshu-image/internal/config"
	"github.com/xiaohongshu-image/internal/db"
	"github.com/xiaohongshu-image/internal/models"
	"github.com/xiaohongshu-image/internal/services/intent"
	"github.com/xiaohongshu-image/internal/services/mailer"
	"github.com/xiaohongshu-image/internal/services/provider"
	"github.com/xiaohongshu-image/internal/services/storage"
	"github.com/xiaohongshu-image/internal/services/xhsconnector"
	"github.com/xiaohongshu-image/internal/worker"
	"github.com/xiaohongshu-image/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logger.New(cfg.Server.Mode)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	database, err := db.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Asynq.RedisDB,
	})

	minioService, err := storage.NewMinIOService(&cfg.MinIO)
	if err != nil {
		logger.Fatal("Failed to create MinIO service", zap.Error(err))
	}

	mailerService := mailer.NewService(&cfg.SMTP)

	llmHTTPClient := intent.NewRealHTTPClient(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Timeout)
	intentService := intent.NewServiceWithClient(&cfg.LLM, llmHTTPClient)

	setting, err := database.GetSetting()
	if err != nil {
		logger.Fatal("Failed to get settings", zap.Error(err))
	}

	connectorCfg := &xhsconnector.ConnectorConfig{
		Mode:         setting.ConnectorMode,
		MCPServerCmd: setting.MCPServerCmd,
		MCPServerURL: setting.MCPServerURL,
		MCPAuth:      setting.MCPAuth,
	}

	connector, err := xhsconnector.NewConnector(connectorCfg)
	if err != nil {
		logger.Fatal("Failed to create connector", zap.Error(err))
	}

	var providers []provider.ProviderConfig
	if err := json.Unmarshal([]byte(setting.ProviderJSON), &providers); err != nil {
		logger.Fatal("Failed to unmarshal provider configs", zap.Error(err))
	}

	providersMap := make(map[string]provider.Provider)
	for _, pCfg := range providers {
		if pCfg.ProviderName == "mock" {
			providersMap[pCfg.ProviderName] = provider.NewMockProvider(pCfg.ProviderName, minioService)
		} else {
			providersMap[pCfg.ProviderName] = provider.NewHTTPProvider(&pCfg)
		}
	}

	workerInstance := worker.NewWorker(
		database,
		asynqClient,
		connector,
		intentService,
		providersMap,
		minioService,
		mailerService,
		logger,
	)

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Asynq.RedisDB,
		},
		asynq.Config{
			Concurrency: cfg.Asynq.Concurrency,
			Queues:      cfg.Asynq.Queues,
		},
	)

	mux := asynq.NewServeMux()
	workerInstance.RegisterHandlers(mux)

	go func() {
		logger.Info("Starting worker server")
		if err := asynqServer.Run(mux); err != nil {
			logger.Fatal("Failed to run worker server", zap.Error(err))
		}
	}()

	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware(logger))

	handler := api.NewHandler(database, asynqClient, workerInstance, logger)
	handler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info("Starting API server", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	go startScheduler(ctx, database, asynqClient, setting, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	asynqServer.Shutdown()

	logger.Info("Server exited")
}

func startScheduler(ctx context.Context, database *db.Database, client *asynq.Client, setting *models.Setting, logger *zap.Logger) {
	ticker := time.NewTicker(time.Duration(setting.PollingIntervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			payload, _ := json.Marshal(worker.PollCommentsPayload{
				NoteTarget: setting.NoteTarget,
			})

			_, err := client.Enqueue(
				asynq.NewTask(worker.TypePollComments, payload, asynq.Queue("critical")),
			)
			if err != nil {
				logger.Error("Failed to enqueue poll task", zap.Error(err))
			}
		}
	}
}

func loggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("HTTP request",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Duration("latency", latency),
		)
	}
}
