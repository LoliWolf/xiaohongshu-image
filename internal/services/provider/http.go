package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPProvider struct {
	cfg    *ProviderConfig
	client *http.Client
	mapper *Mapper
}

func NewHTTPProvider(cfg *ProviderConfig) *HTTPProvider {
	return &HTTPProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		mapper: NewMapper(cfg.RequestMapping),
	}
}

func (p *HTTPProvider) Name() string {
	return p.cfg.ProviderName
}

func (p *HTTPProvider) Submit(ctx context.Context, req UnifiedGenRequest) (*SubmitResult, error) {
	mappedReq, err := p.mapper.MapRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to map request: %w", err)
	}

	reqBody, err := json.Marshal(mappedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.cfg.BaseURL + p.cfg.SubmitPath
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.cfg.Headers {
		httpReq.Header.Set(k, v)
	}

	if p.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.cfg.APIKey))
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(body))
	}

	jobID, err := p.mapper.ExtractJobID(body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract job ID: %w", err)
	}

	return &SubmitResult{
		ProviderJobID: jobID,
	}, nil
}

func (p *HTTPProvider) Status(ctx context.Context, jobID string) (*StatusResult, error) {
	statusPath := strings.Replace(p.cfg.StatusPathTemplate, "{id}", jobID, -1)
	url := p.cfg.BaseURL + statusPath

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for k, v := range p.cfg.Headers {
		httpReq.Header.Set(k, v)
	}

	if p.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.cfg.APIKey))
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(body))
	}

	status, progress, resultURL, err := p.mapper.ExtractStatus(body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract status: %w", err)
	}

	return &StatusResult{
		Status:    status,
		Progress:  progress,
		ResultURL: resultURL,
	}, nil
}
