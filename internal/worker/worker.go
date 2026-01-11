package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/xiaohongshu-image/internal/db"
	"github.com/xiaohongshu-image/internal/models"
	"github.com/xiaohongshu-image/internal/services/intent"
	"github.com/xiaohongshu-image/internal/services/mailer"
	"github.com/xiaohongshu-image/internal/services/provider"
	"github.com/xiaohongshu-image/internal/services/xhsconnector"
	"go.uber.org/zap"
)

const (
	TypePollComments   = "poll:comments"
	TypeProcessComment = "process:comment"
	TypeSubmitJob      = "submit:job"
	TypeCheckStatus    = "check:status"
	TypeSendEmail      = "send:email"
)

type Worker struct {
	db        *db.Database
	redis     *asynq.Client
	connector xhsconnector.Connector
	intentSvc *intent.Service
	providers map[string]provider.Provider
	storage   provider.Storage
	mailer    *mailer.Service
	logger    *zap.Logger
}

func NewWorker(
	db *db.Database,
	redis *asynq.Client,
	connector xhsconnector.Connector,
	intentSvc *intent.Service,
	providers map[string]provider.Provider,
	storage provider.Storage,
	mailer *mailer.Service,
	logger *zap.Logger,
) *Worker {
	return &Worker{
		db:        db,
		redis:     redis,
		connector: connector,
		intentSvc: intentSvc,
		providers: providers,
		storage:   storage,
		mailer:    mailer,
		logger:    logger,
	}
}

func (w *Worker) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc(TypePollComments, w.HandlePollComments)
	mux.HandleFunc(TypeProcessComment, w.HandleProcessComment)
	mux.HandleFunc(TypeSubmitJob, w.HandleSubmitJob)
	mux.HandleFunc(TypeCheckStatus, w.HandleCheckStatus)
	mux.HandleFunc(TypeSendEmail, w.HandleSendEmail)
}

type PollCommentsPayload struct {
	NoteTarget string `json:"note_target"`
}

func (w *Worker) HandlePollComments(ctx context.Context, t *asynq.Task) error {
	var payload PollCommentsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("failed to unmarshal poll payload", zap.Error(err))
		return err
	}

	w.logger.Info("polling comments", zap.String("note_target", payload.NoteTarget))

	note, err := w.db.GetOrCreateNote(payload.NoteTarget)
	if err != nil {
		w.logger.Error("failed to get or create note", zap.Error(err))
		return err
	}

	lockKey := fmt.Sprintf("lock:poll:%s", payload.NoteTarget)
	acquired, err := w.acquireLock(ctx, lockKey, 60)
	if err != nil {
		w.logger.Error("failed to acquire lock", zap.Error(err))
		return err
	}
	if !acquired {
		w.logger.Info("poll already in progress", zap.String("note_target", payload.NoteTarget))
		return nil
	}
	defer w.releaseLock(ctx, lockKey)

	cursor := ""
	if note.LastCursor != nil {
		cursor = *note.LastCursor
	}

	result, err := w.connector.ListComments(ctx, payload.NoteTarget, cursor)
	if err != nil {
		w.logger.Error("failed to list comments", zap.Error(err))
		now := time.Now()
		note.LastError = new(string)
		*note.LastError = err.Error()
		note.LastPolledAt = &now
		w.db.UpdateNote(note)
		return err
	}

	newCommentsCount := 0
	for _, comment := range result.Comments {
		commentUID := comment.CommentID
		if commentUID == "" {
			commentUID = w.generateCommentUID(comment)
		}

		exists, err := w.db.CommentExists(commentUID)
		if err != nil {
			w.logger.Error("failed to check comment existence", zap.Error(err))
			continue
		}

		if exists {
			continue
		}

		now := time.Now()
		dbComment := &models.Comment{
			NoteTarget:       payload.NoteTarget,
			CommentUID:       commentUID,
			UserName:         &comment.UserName,
			Content:          comment.Content,
			CommentCreatedAt: &comment.CommentCreatedAt,
			IngestedAt:       now,
		}

		if err := w.db.CreateComment(dbComment); err != nil {
			w.logger.Error("failed to create comment", zap.Error(err))
			continue
		}

		newCommentsCount++

		taskPayload, _ := json.Marshal(ProcessCommentPayload{
			CommentID:  dbComment.ID,
			CommentUID: commentUID,
			Content:    comment.Content,
			NoteTarget: payload.NoteTarget,
		})

		_, err = w.redis.Enqueue(
			asynq.NewTask(TypeProcessComment, taskPayload, asynq.Queue("default")),
		)
		if err != nil {
			w.logger.Error("failed to enqueue process comment task", zap.Error(err))
		}
	}

	now := time.Now()
	note.LastPolledAt = &now
	if result.NextCursor != "" {
		note.LastCursor = &result.NextCursor
	}
	note.LastError = nil
	if err := w.db.UpdateNote(note); err != nil {
		w.logger.Error("failed to update note", zap.Error(err))
	}

	w.logger.Info("poll completed",
		zap.String("note_target", payload.NoteTarget),
		zap.Int("new_comments", newCommentsCount),
	)

	return nil
}

func (w *Worker) generateCommentUID(comment xhsconnector.Comment) string {
	data := fmt.Sprintf("%s|%s|%s|%d",
		comment.CommentID,
		comment.UserName,
		comment.Content,
		comment.CommentCreatedAt.Unix(),
	)
	return fmt.Sprintf("%x", data)
}

type ProcessCommentPayload struct {
	CommentID  uint   `json:"comment_id"`
	CommentUID string `json:"comment_uid"`
	Content    string `json:"content"`
	NoteTarget string `json:"note_target"`
}

func (w *Worker) HandleProcessComment(ctx context.Context, t *asynq.Task) error {
	var payload ProcessCommentPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("failed to unmarshal process comment payload", zap.Error(err))
		return err
	}

	w.logger.Info("processing comment", zap.String("comment_uid", payload.CommentUID))

	setting, err := w.db.GetSetting()
	if err != nil {
		w.logger.Error("failed to get settings", zap.Error(err))
		return err
	}

	intentResult, err := w.intentSvc.ExtractIntent(ctx, payload.Content, setting.IntentThreshold)
	if err != nil {
		w.logger.Error("failed to extract intent", zap.Error(err), zap.String("comment_uid", payload.CommentUID))

		w.db.CreateAuditLog(&models.AuditLog{
			Level:       "ERROR",
			Event:       "intent_extraction_failed",
			PayloadJSON: fmt.Sprintf(`{"comment_uid": "%s", "error": "%s"}`, payload.CommentUID, err.Error()),
		})

		return nil
	}

	if !intentResult.HasRequest {
		w.logger.Info("comment skipped - no clear intent", zap.String("comment_uid", payload.CommentUID), zap.String("reason", intentResult.Reason))
		return nil
	}

	task := &models.Task{
		CommentID:   payload.CommentID,
		Status:      models.TaskStatusExtracted,
		RequestType: models.RequestType(intentResult.RequestType),
		Email:       intentResult.Email,
		Prompt:      &intentResult.Prompt,
		Confidence:  &intentResult.Confidence,
	}

	if err := w.db.CreateTask(task); err != nil {
		w.logger.Error("failed to create task", zap.Error(err), zap.String("comment_uid", payload.CommentUID))
		return err
	}

	w.logger.Info("task created", zap.Uint("task_id", task.ID), zap.String("comment_uid", payload.CommentUID))

	submitPayload, _ := json.Marshal(SubmitJobPayload{
		TaskID:      task.ID,
		RequestType: string(task.RequestType),
		Prompt:      *task.Prompt,
	})

	_, err = w.redis.Enqueue(
		asynq.NewTask(TypeSubmitJob, submitPayload, asynq.Queue("critical")),
	)
	if err != nil {
		w.logger.Error("failed to enqueue submit job task", zap.Error(err))
	}

	return nil
}

type SubmitJobPayload struct {
	TaskID      uint   `json:"task_id"`
	RequestType string `json:"request_type"`
	Prompt      string `json:"prompt"`
}

func (w *Worker) HandleSubmitJob(ctx context.Context, t *asynq.Task) error {
	var payload SubmitJobPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("failed to unmarshal submit job payload", zap.Error(err))
		return err
	}

	w.logger.Info("submitting job", zap.Uint("task_id", payload.TaskID))

	task, err := w.db.GetTaskByID(payload.TaskID)
	if err != nil {
		w.logger.Error("failed to get task", zap.Error(err), zap.Uint("task_id", payload.TaskID))
		return err
	}

	setting, err := w.db.GetSetting()
	if err != nil {
		w.logger.Error("failed to get settings", zap.Error(err))
		return err
	}

	var providers []provider.ProviderConfig
	if err := json.Unmarshal([]byte(setting.ProviderJSON), &providers); err != nil {
		w.logger.Error("failed to unmarshal provider configs", zap.Error(err))
		return err
	}

	if len(providers) == 0 {
		err := fmt.Errorf("no providers configured")
		w.logger.Error("no providers configured", zap.Error(err))
		task.Status = models.TaskStatusFailed
		task.Error = new(string)
		*task.Error = err.Error()
		w.db.UpdateTask(task)
		return err
	}

	providerName := providers[0].ProviderName
	prov, exists := w.providers[providerName]
	if !exists {
		err := fmt.Errorf("provider not found: %s", providerName)
		w.logger.Error("provider not found", zap.String("provider", providerName))
		task.Status = models.TaskStatusFailed
		task.Error = new(string)
		*task.Error = err.Error()
		w.db.UpdateTask(task)
		return err
	}

	req := provider.UnifiedGenRequest{
		RequestID: fmt.Sprintf("task_%d", payload.TaskID),
		Type:      provider.RequestType(payload.RequestType),
		Prompt:    payload.Prompt,
	}

	result, err := prov.Submit(ctx, req)
	if err != nil {
		w.logger.Error("failed to submit job", zap.Error(err), zap.Uint("task_id", payload.TaskID))
		task.Status = models.TaskStatusFailed
		task.Error = new(string)
		*task.Error = err.Error()
		w.db.UpdateTask(task)
		return err
	}

	task.Status = models.TaskStatusSubmitted
	task.ProviderName = &providerName
	task.ProviderJobID = &result.ProviderJobID
	if err := w.db.UpdateTask(task); err != nil {
		w.logger.Error("failed to update task", zap.Error(err))
		return err
	}

	statusPayload, _ := json.Marshal(CheckStatusPayload{
		TaskID:        payload.TaskID,
		ProviderJobID: result.ProviderJobID,
		ProviderName:  providerName,
		RetryCount:    0,
	})

	_, err = w.redis.Enqueue(
		asynq.NewTask(TypeCheckStatus, statusPayload, asynq.ProcessIn(15*time.Second), asynq.Queue("default")),
	)
	if err != nil {
		w.logger.Error("failed to enqueue check status task", zap.Error(err))
	}

	w.logger.Info("job submitted", zap.Uint("task_id", payload.TaskID), zap.String("provider_job_id", result.ProviderJobID))

	return nil
}

type CheckStatusPayload struct {
	TaskID        uint   `json:"task_id"`
	ProviderJobID string `json:"provider_job_id"`
	ProviderName  string `json:"provider_name"`
	RetryCount    int    `json:"retry_count"`
}

func (w *Worker) HandleCheckStatus(ctx context.Context, t *asynq.Task) error {
	var payload CheckStatusPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("failed to unmarshal check status payload", zap.Error(err))
		return err
	}

	w.logger.Info("checking status", zap.Uint("task_id", payload.TaskID), zap.String("provider_job_id", payload.ProviderJobID))

	task, err := w.db.GetTaskByID(payload.TaskID)
	if err != nil {
		w.logger.Error("failed to get task", zap.Error(err), zap.Uint("task_id", payload.TaskID))
		return err
	}

	if task.Status == models.TaskStatusSucceeded || task.Status == models.TaskStatusEmailed {
		w.logger.Info("task already completed", zap.Uint("task_id", payload.TaskID))
		return nil
	}

	prov, exists := w.providers[payload.ProviderName]
	if !exists {
		err := fmt.Errorf("provider not found: %s", payload.ProviderName)
		w.logger.Error("provider not found", zap.String("provider", payload.ProviderName))
		return err
	}

	status, err := prov.Status(ctx, payload.ProviderJobID)
	if err != nil {
		w.logger.Error("failed to check status", zap.Error(err), zap.Uint("task_id", payload.TaskID))
		return err
	}

	if status.Status == provider.JobStatusSucceeded {
		task.Status = models.TaskStatusSucceeded
		if status.ResultURL != nil {
			task.ResultURL = status.ResultURL
		}
		if err := w.db.UpdateTask(task); err != nil {
			w.logger.Error("failed to update task", zap.Error(err))
			return err
		}

		emailPayload, _ := json.Marshal(SendEmailPayload{
			TaskID: payload.TaskID,
		})

		_, err = w.redis.Enqueue(
			asynq.NewTask(TypeSendEmail, emailPayload, asynq.Queue("critical")),
		)
		if err != nil {
			w.logger.Error("failed to enqueue send email task", zap.Error(err))
		}

		w.logger.Info("task succeeded", zap.Uint("task_id", payload.TaskID))
		return nil
	}

	if status.Status == provider.JobStatusFailed {
		task.Status = models.TaskStatusFailed
		if status.Error != nil {
			task.Error = status.Error
		}
		w.db.UpdateTask(task)
		w.logger.Info("task failed", zap.Uint("task_id", payload.TaskID))
		return nil
	}

	if payload.RetryCount >= 20 {
		task.Status = models.TaskStatusFailed
		task.Error = new(string)
		*task.Error = "max retries exceeded"
		w.db.UpdateTask(task)
		w.logger.Info("task failed - max retries", zap.Uint("task_id", payload.TaskID))
		return nil
	}

	backoff := 15 * time.Second
	for i := 0; i < payload.RetryCount; i++ {
		backoff *= 2
		if backoff > 60*time.Second {
			backoff = 60 * time.Second
		}
	}

	nextPayload, _ := json.Marshal(CheckStatusPayload{
		TaskID:        payload.TaskID,
		ProviderJobID: payload.ProviderJobID,
		ProviderName:  payload.ProviderName,
		RetryCount:    payload.RetryCount + 1,
	})

	_, err = w.redis.Enqueue(
		asynq.NewTask(TypeCheckStatus, nextPayload, asynq.ProcessIn(backoff), asynq.Queue("default")),
	)
	if err != nil {
		w.logger.Error("failed to enqueue next check status task", zap.Error(err))
	}

	return nil
}

type SendEmailPayload struct {
	TaskID uint `json:"task_id"`
}

func (w *Worker) HandleSendEmail(ctx context.Context, t *asynq.Task) error {
	var payload SendEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("failed to unmarshal send email payload", zap.Error(err))
		return err
	}

	w.logger.Info("sending email", zap.Uint("task_id", payload.TaskID))

	task, err := w.db.GetTaskByID(payload.TaskID)
	if err != nil {
		w.logger.Error("failed to get task", zap.Error(err), zap.Uint("task_id", payload.TaskID))
		return err
	}

	if task.Status == models.TaskStatusEmailed {
		w.logger.Info("email already sent", zap.Uint("task_id", payload.TaskID))
		return nil
	}

	if task.Email == nil {
		w.logger.Error("no email address", zap.Uint("task_id", payload.TaskID))
		return nil
	}

	if task.ResultURL == nil {
		w.logger.Error("no result URL", zap.Uint("task_id", payload.TaskID))
		return nil
	}

	if err := w.checkRateLimit(ctx, *task.Email); err != nil {
		w.logger.Error("rate limit exceeded", zap.Error(err), zap.String("email", *task.Email))
		return nil
	}

	err = w.mailer.SendResultEmail(*task.Email, string(task.RequestType), *task.Prompt, *task.ResultURL)
	if err != nil {
		w.logger.Error("failed to send email", zap.Error(err), zap.Uint("task_id", payload.TaskID))

		delivery := &models.Delivery{
			TaskID:  payload.TaskID,
			EmailTo: *task.Email,
			Status:  models.DeliveryStatusFailed,
			Error:   new(string),
		}
		*delivery.Error = err.Error()
		w.db.CreateDelivery(delivery)

		return err
	}

	delivery := &models.Delivery{
		TaskID:  payload.TaskID,
		EmailTo: *task.Email,
		Status:  models.DeliveryStatusSent,
	}
	now := time.Now()
	delivery.SentAt = &now
	w.db.CreateDelivery(delivery)

	task.Status = models.TaskStatusEmailed
	w.db.UpdateTask(task)

	w.logger.Info("email sent", zap.Uint("task_id", payload.TaskID), zap.String("email", *task.Email))

	return nil
}

func (w *Worker) acquireLock(ctx context.Context, key string, ttl int) (bool, error) {
	return true, nil
}

func (w *Worker) releaseLock(ctx context.Context, key string) error {
	return nil
}

func (w *Worker) checkRateLimit(ctx context.Context, email string) error {
	return nil
}
