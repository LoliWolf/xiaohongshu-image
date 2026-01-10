package intent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xiaohongshu-image/internal/config"
)

type realHTTPClient struct {
	client  *http.Client
	apiKey  string
	baseURL string
	timeout time.Duration
}

func NewRealHTTPClient(baseURL, apiKey string, timeout time.Duration) HTTPClient {
	return &realHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		apiKey:  apiKey,
		baseURL: baseURL,
		timeout: timeout,
	}
}

func (c *realHTTPClient) Do(reqBody interface{}) (*interface{}, error) {
	reqBytes, ok := reqBody.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid request body type")
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
	}

	var llmResp LLMResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal LLM response: %w", err)
	}

	if llmResp.Error != nil {
		return nil, fmt.Errorf("LLM API error: %s", llmResp.Error.Message)
	}

	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	content := llmResp.Choices[0].Message.Content

	var intentResult IntentResult
	if err := json.Unmarshal([]byte(content), &intentResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal intent result from LLM: %w", err)
	}

	intentResult.RawJSON = []byte(content)

	var result interface{} = intentResult
	return &result, nil
}

func (s *Service) SetHTTPClient(client HTTPClient) {
	s.httpClient = client
}

func NewServiceWithClient(cfg *config.LLMConfig, client HTTPClient) *Service {
	return &Service{
		cfg:        cfg,
		httpClient: client,
	}
}
