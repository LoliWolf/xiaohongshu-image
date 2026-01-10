package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/xiaohongshu-image/internal/config"
	"github.com/xiaohongshu-image/internal/db"
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

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Asynq.RedisDB,
	})

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")
	asynqServer.Shutdown()
	logger.Info("Worker exited")
}
