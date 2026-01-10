package xhsconnector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MCPConnector struct {
	serverURL string
	auth      *string
	client    *http.Client
}

type MCPRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MCPToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func NewMCPConnector(cfg *ConnectorConfig) (*MCPConnector, error) {
	if cfg.MCPServerURL == nil || *cfg.MCPServerURL == "" {
		return nil, fmt.Errorf("MCP server URL is required for MCP mode")
	}

	return &MCPConnector{
		serverURL: *cfg.MCPServerURL,
		auth:      cfg.MCPAuth,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (m *MCPConnector) callTool(ctx context.Context, toolName string, args map[string]interface{}) (json.RawMessage, error) {
	params := MCPToolCallParams{
		Name:      toolName,
		Arguments: args,
	}

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  []interface{}{params},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MCP request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.serverURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if m.auth != nil && *m.auth != "" {
		httpReq.Header.Set("Authorization", *m.auth)
	}

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call MCP server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server returned status %d: %s", resp.StatusCode, string(body))
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP response: %w", err)
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", mcpResp.Error.Message)
	}

	return mcpResp.Result, nil
}

func (m *MCPConnector) ListComments(ctx context.Context, noteIDOrURL string, cursor string) (*ListCommentsResult, error) {
	args := map[string]interface{}{
		"note_id_or_url": noteIDOrURL,
	}

	if cursor != "" {
		args["cursor"] = cursor
	}

	result, err := m.callTool(ctx, "xhs_list_comments", args)
	if err != nil {
		return nil, fmt.Errorf("failed to call xhs_list_comments: %w", err)
	}

	var commentsResult struct {
		Comments   []Comment `json:"comments"`
		NextCursor string    `json:"next_cursor"`
		HasMore    bool      `json:"has_more"`
	}

	if err := json.Unmarshal(result, &commentsResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal comments result: %w", err)
	}

	return &ListCommentsResult{
		Comments:   commentsResult.Comments,
		NextCursor: commentsResult.NextCursor,
		HasMore:    commentsResult.HasMore,
	}, nil
}
