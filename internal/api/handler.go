package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/xiaohongshu-image/internal/db"
	"github.com/xiaohongshu-image/internal/worker"
	"go.uber.org/zap"
)

type Handler struct {
	db     *db.Database
	redis  *asynq.Client
	worker *worker.Worker
	logger *zap.Logger
}

func NewHandler(
	db *db.Database,
	redis *asynq.Client,
	worker *worker.Worker,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		db:     db,
		redis:  redis,
		worker: worker,
		logger: logger,
	}
}

type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", h.HealthCheck)
	r.GET("/metrics", h.Metrics)

	api := r.Group("/api")
	{
		api.GET("/settings", h.GetSettings)
		api.PUT("/settings", h.UpdateSettings)
		api.POST("/poll/run", h.RunPoll)
		api.GET("/tasks", h.ListTasks)
		api.GET("/tasks/:id", h.GetTask)
		api.GET("/files/:key", h.GetFile)
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (h *Handler) Metrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"metrics": "prometheus metrics would be here",
	})
}

func (h *Handler) GetSettings(c *gin.Context) {
	setting, err := h.db.GetSetting()
	if err != nil {
		h.logger.Error("failed to get settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get settings",
		})
		return
	}

	c.JSON(http.StatusOK, setting)
}

type UpdateSettingsRequest struct {
	ConnectorMode      *string  `json:"connector_mode" binding:"omitempty,oneof=mock mcp"`
	MCPServerCmd       *string  `json:"mcp_server_cmd" binding:"omitempty"`
	MCPServerURL       *string  `json:"mcp_server_url" binding:"omitempty"`
	MCPAuth            *string  `json:"mcp_auth" binding:"omitempty"`
	NoteTarget         *string  `json:"note_target" binding:"omitempty"`
	PollingIntervalSec *int     `json:"polling_interval_sec" binding:"omitempty,min=10"`
	LLMBaseURL         *string  `json:"llm_base_url" binding:"omitempty"`
	LLMAPIKey          *string  `json:"llm_api_key" binding:"omitempty"`
	LLMModel           *string  `json:"llm_model" binding:"omitempty"`
	LLMTimeoutSec      *int     `json:"llm_timeout_sec" binding:"omitempty,min=5,max=300"`
	IntentThreshold    *float64 `json:"intent_threshold" binding:"omitempty,min=0,max=1"`
	SMTPHost           *string  `json:"smtp_host" binding:"omitempty"`
	SMTPPort           *int     `json:"smtp_port" binding:"omitempty,min=1,max=65535"`
	SMTPUser           *string  `json:"smtp_user" binding:"omitempty"`
	SMTPPass           *string  `json:"smtp_pass" binding:"omitempty"`
	SMTPFrom           *string  `json:"smtp_from" binding:"omitempty,email"`
	ProviderJSON       *string  `json:"provider_json" binding:"omitempty"`
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	setting, err := h.db.GetSetting()
	if err != nil {
		h.logger.Error("failed to get settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get settings",
		})
		return
	}

	if req.ConnectorMode != nil {
		setting.ConnectorMode = *req.ConnectorMode
	}
	if req.MCPServerCmd != nil {
		setting.MCPServerCmd = req.MCPServerCmd
	}
	if req.MCPServerURL != nil {
		setting.MCPServerURL = req.MCPServerURL
	}
	if req.MCPAuth != nil {
		setting.MCPAuth = req.MCPAuth
	}
	if req.NoteTarget != nil {
		setting.NoteTarget = *req.NoteTarget
	}
	if req.PollingIntervalSec != nil {
		setting.PollingIntervalSec = *req.PollingIntervalSec
	}
	if req.LLMBaseURL != nil {
		setting.LLMBaseURL = req.LLMBaseURL
	}
	if req.LLMAPIKey != nil {
		setting.LLMAPIKey = req.LLMAPIKey
	}
	if req.LLMModel != nil {
		setting.LLMModel = req.LLMModel
	}
	if req.LLMTimeoutSec != nil {
		setting.LLMTimeoutSec = *req.LLMTimeoutSec
	}
	if req.IntentThreshold != nil {
		setting.IntentThreshold = *req.IntentThreshold
	}
	if req.SMTPHost != nil {
		setting.SMTPHost = req.SMTPHost
	}
	if req.SMTPPort != nil {
		setting.SMTPPort = req.SMTPPort
	}
	if req.SMTPUser != nil {
		setting.SMTPUser = req.SMTPUser
	}
	if req.SMTPPass != nil {
		setting.SMTPPass = req.SMTPPass
	}
	if req.SMTPFrom != nil {
		setting.SMTPFrom = req.SMTPFrom
	}
	if req.ProviderJSON != nil {
		setting.ProviderJSON = *req.ProviderJSON
	}

	if err := h.db.UpdateSetting(setting); err != nil {
		h.logger.Error("failed to update settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update settings",
		})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (h *Handler) RunPoll(c *gin.Context) {
	setting, err := h.db.GetSetting()
	if err != nil {
		h.logger.Error("failed to get settings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get settings",
		})
		return
	}

	payload, _ := json.Marshal(worker.PollCommentsPayload{
		NoteTarget: setting.NoteTarget,
	})

	_, err = h.redis.Enqueue(
		asynq.NewTask(worker.TypePollComments, payload, asynq.Queue("critical")),
	)
	if err != nil {
		h.logger.Error("failed to enqueue poll task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to enqueue poll task",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Poll task enqueued",
	})
}

func (h *Handler) ListTasks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 100
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	tasks, err := h.db.ListTasks(limit, offset)
	if err != nil {
		h.logger.Error("failed to list tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list tasks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":  tasks,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid task ID",
		})
		return
	}

	task, err := h.db.GetTaskByID(uint(id))
	if err != nil {
		h.logger.Error("failed to get task", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) GetFile(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_KEY",
			Message: "File key is required",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File download would be implemented here",
		"key":     key,
	})
}
