package provider

import (
	"context"
)

type RequestType string

const (
	RequestTypeImage RequestType = "image"
	RequestTypeVideo RequestType = "video"
)

type UnifiedGenRequest struct {
	RequestID      string                 `json:"request_id"`
	Type           RequestType            `json:"type"`
	Prompt         string                 `json:"prompt"`
	NegativePrompt string                 `json:"negative_prompt,omitempty"`
	Style          string                 `json:"style,omitempty"`
	Width          *int                   `json:"width,omitempty"`
	Height         *int                   `json:"height,omitempty"`
	DurationSec    *int                   `json:"duration_sec,omitempty"`
	Ratio          string                 `json:"ratio,omitempty"`
	Seed           *int                   `json:"seed,omitempty"`
	Extra          map[string]interface{} `json:"extra,omitempty"`
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusSucceeded JobStatus = "succeeded"
	JobStatusFailed    JobStatus = "failed"
)

type SubmitResult struct {
	ProviderJobID string `json:"provider_job_id"`
}

type StatusResult struct {
	Status    JobStatus `json:"status"`
	Progress  int       `json:"progress"`
	ResultURL *string   `json:"result_url,omitempty"`
	Error     *string   `json:"error,omitempty"`
}

type Provider interface {
	Submit(ctx context.Context, req UnifiedGenRequest) (*SubmitResult, error)
	Status(ctx context.Context, jobID string) (*StatusResult, error)
	Name() string
}

type ProviderConfig struct {
	ProviderName       string                 `json:"provider_name"`
	Type               string                 `json:"type"`
	BaseURL            string                 `json:"base_url"`
	APIKey             string                 `json:"api_key"`
	SubmitPath         string                 `json:"submit_path"`
	StatusPathTemplate string                 `json:"status_path_template"`
	Headers            map[string]string      `json:"headers"`
	RequestMapping     map[string]interface{} `json:"request_mapping"`
	ResponseMapping    map[string]string      `json:"response_mapping"`
	StatusMapping      map[string]string      `json:"status_mapping"`
}
