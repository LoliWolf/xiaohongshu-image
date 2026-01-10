package models

import (
	"time"
)

type ConnectorMode string

const (
	ConnectorModeMock ConnectorMode = "mock"
	ConnectorModeMCP  ConnectorMode = "mcp"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusExtracted TaskStatus = "EXTRACTED"
	TaskStatusSubmitted TaskStatus = "SUBMITTED"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusSucceeded TaskStatus = "SUCCEEDED"
	TaskStatusEmailed   TaskStatus = "EMAILED"
	TaskStatusFailed    TaskStatus = "FAILED"
)

type RequestType string

const (
	RequestTypeImage RequestType = "image"
	RequestTypeVideo RequestType = "video"
)

type DeliveryStatus string

const (
	DeliveryStatusSent   DeliveryStatus = "SENT"
	DeliveryStatusFailed DeliveryStatus = "FAILED"
)

type Setting struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConnectorMode      string    `gorm:"type:varchar(20);not null;default:'mock'" json:"connector_mode"`
	MCPServerCmd       *string   `gorm:"type:text" json:"mcp_server_cmd,omitempty"`
	MCPServerURL       *string   `gorm:"type:varchar(500)" json:"mcp_server_url,omitempty"`
	MCPAuth            *string   `gorm:"type:text" json:"mcp_auth,omitempty"`
	NoteTarget         string    `gorm:"type:varchar(500);not null" json:"note_target"`
	PollingIntervalSec int       `gorm:"not null;default:120" json:"polling_interval_sec"`
	LLMBaseURL         *string   `gorm:"type:varchar(500)" json:"llm_base_url,omitempty"`
	LLMAPIKey          *string   `gorm:"type:varchar(200)" json:"llm_api_key,omitempty"`
	LLMModel           *string   `gorm:"type:varchar(100)" json:"llm_model,omitempty"`
	LLMTimeoutSec      int       `gorm:"default:15" json:"llm_timeout_sec"`
	IntentThreshold    float64   `gorm:"type:decimal(3,2);not null;default:0.70" json:"intent_threshold"`
	SMTPHost           *string   `gorm:"type:varchar(200)" json:"smtp_host,omitempty"`
	SMTPPort           *int      `json:"smtp_port,omitempty"`
	SMTPUser           *string   `gorm:"type:varchar(200)" json:"smtp_user,omitempty"`
	SMTPPass           *string   `gorm:"type:varchar(200)" json:"smtp_pass,omitempty"`
	SMTPFrom           *string   `gorm:"type:varchar(200)" json:"smtp_from,omitempty"`
	ProviderJSON       string    `gorm:"type:json" json:"provider_json"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (Setting) TableName() string {
	return "settings"
}

type Note struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	NoteTarget   string     `gorm:"type:varchar(500);uniqueIndex;not null" json:"note_target"`
	LastCursor   *string    `gorm:"type:text" json:"last_cursor,omitempty"`
	LastPolledAt *time.Time `json:"last_polled_at,omitempty"`
	LastError    *string    `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (Note) TableName() string {
	return "notes"
}

type Comment struct {
	ID               uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	NoteTarget       string     `gorm:"type:varchar(500);not null;index:idx_note_target" json:"note_target"`
	CommentUID       string     `gorm:"type:varchar(100);uniqueIndex:uk_comment_uid;not null" json:"comment_uid"`
	UserName         *string    `gorm:"type:varchar(200)" json:"user_name,omitempty"`
	Content          string     `gorm:"type:text" json:"content"`
	CommentCreatedAt *time.Time `json:"comment_created_at,omitempty"`
	IngestedAt       time.Time  `json:"ingested_at"`
}

func (Comment) TableName() string {
	return "comments"
}

type Task struct {
	ID              uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentID       uint        `gorm:"not null;uniqueIndex:uk_comment_id" json:"comment_id"`
	Status          TaskStatus  `gorm:"type:enum('PENDING','EXTRACTED','SUBMITTED','RUNNING','SUCCEEDED','EMAILED','FAILED');not null;default:'PENDING'" json:"status"`
	RequestType     RequestType `gorm:"type:enum('image','video');not null" json:"request_type"`
	Email           *string     `gorm:"type:varchar(200);index:idx_email" json:"email,omitempty"`
	Prompt          *string     `gorm:"type:text" json:"prompt,omitempty"`
	Confidence      *float64    `gorm:"type:decimal(3,2)" json:"confidence,omitempty"`
	ProviderName    *string     `gorm:"type:varchar(100)" json:"provider_name,omitempty"`
	ProviderJobID   *string     `gorm:"type:varchar(200)" json:"provider_job_id,omitempty"`
	ResultObjectKey *string     `gorm:"type:varchar(500)" json:"result_object_key,omitempty"`
	ResultURL       *string     `gorm:"type:varchar(1000)" json:"result_url,omitempty"`
	Error           *string     `gorm:"type:text" json:"error,omitempty"`
	RetryCount      int         `gorm:"default:0" json:"retry_count"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Comment         *Comment    `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
	Deliveries      []Delivery  `gorm:"foreignKey:TaskID" json:"deliveries,omitempty"`
}

func (Task) TableName() string {
	return "tasks"
}

type Delivery struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID    uint           `gorm:"not null;index:idx_task_id" json:"task_id"`
	EmailTo   string         `gorm:"type:varchar(200);not null" json:"email_to"`
	Status    DeliveryStatus `gorm:"type:enum('SENT','FAILED');not null;index:idx_status" json:"status"`
	SentAt    *time.Time     `json:"sent_at,omitempty"`
	Error     *string        `gorm:"type:text" json:"error,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

func (Delivery) TableName() string {
	return "deliveries"
}

type AuditLog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Level       string    `gorm:"type:varchar(20);not null;index:idx_event" json:"level"`
	Event       string    `gorm:"type:varchar(100);not null;index:idx_event" json:"event"`
	PayloadJSON string    `gorm:"type:json" json:"payload_json"`
	CreatedAt   time.Time `gorm:"index:idx_created_at" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
