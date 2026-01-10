package xhsconnector

import (
	"context"
	"time"
)

type Comment struct {
	CommentID        string    `json:"comment_id"`
	UserName         string    `json:"user_name"`
	Content          string    `json:"content"`
	CommentCreatedAt time.Time `json:"comment_created_at"`
}

type ListCommentsResult struct {
	Comments   []Comment `json:"comments"`
	NextCursor string    `json:"next_cursor"`
	HasMore    bool      `json:"has_more"`
}

type Connector interface {
	ListComments(ctx context.Context, noteIDOrURL string, cursor string) (*ListCommentsResult, error)
}

type ConnectorConfig struct {
	Mode         string
	MCPServerCmd *string
	MCPServerURL *string
	MCPAuth      *string
}

func NewConnector(cfg *ConnectorConfig) (Connector, error) {
	if cfg.Mode == "mcp" {
		return NewMCPConnector(cfg)
	}
	return NewMockConnector(), nil
}
