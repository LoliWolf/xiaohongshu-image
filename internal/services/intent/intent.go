package intent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/xiaohongshu-image/internal/config"
)

type IntentResult struct {
	HasRequest  bool            `json:"has_request"`
	RequestType string          `json:"request_type"`
	Prompt      string          `json:"prompt"`
	Email       *string         `json:"email"`
	Confidence  float64         `json:"confidence"`
	Reason      string          `json:"reason"`
	RawJSON     json.RawMessage `json:"-"`
}

type Service struct {
	cfg        *config.LLMConfig
	httpClient HTTPClient
}

type HTTPClient interface {
	Do(req interface{}) (*interface{}, error)
}

type LLMRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMResponse struct {
	Choices []Choice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

const (
	SystemPrompt = `你是一个意图抽取器。你只能输出 JSON，不能输出任何解释、Markdown、代码块。请从评论中判断是否存在明确的"生成图片/生成视频"请求，并抽取用于生成模型的 prompt，同时抽取邮箱（如果存在）。不确定时必须返回 has_request=false。

输出字段必须严格为：
has_request(boolean), request_type("image"|"video"|"unknown"), prompt(string), email(string|null), confidence(number 0..1), reason(string)`

	UserPromptTemplate = `评论文本如下：
<<<COMMENT>>>

规则：
- 如果评论没有明确要求生成图片/视频，has_request=false
- 如果无法可靠判断类型，request_type="unknown"，has_request=false
- prompt 必须是可直接用于生成模型的描述，去掉邮箱和无关寒暄
- 只要邮箱缺失或疑似无效，email=null，has_request=false
- 仅当非常确定时 confidence 才能 >=0.7`
)

var (
	imageKeywords = []string{
		"出图", "生成图", "做图片", "帮我画", "AI生成", "来一张", "画一张", "生成一张",
		"画个", "做个图", "出个图", "生成个", "画一幅", "生成一幅",
	}

	videoKeywords = []string{
		"做视频", "生成视频", "做个视频", "生成个视频", "出视频", "来个视频",
		"做短片", "生成短片", "做个短片",
	}

	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
)

func NewService(cfg *config.LLMConfig) *Service {
	return &Service{
		cfg: cfg,
		httpClient: &defaultHTTPClient{
			timeout: cfg.Timeout,
		},
	}
}

func (s *Service) ExtractIntent(ctx context.Context, comment string, threshold float64) (*IntentResult, error) {
	email := s.extractEmail(comment)

	if !s.hasGenerationKeywords(comment) {
		return &IntentResult{
			HasRequest:  false,
			RequestType: "unknown",
			Prompt:      "",
			Email:       email,
			Confidence:  0,
			Reason:      "评论不包含生成图片/视频的关键词",
		}, nil
	}

	if email == nil {
		return &IntentResult{
			HasRequest:  false,
			RequestType: "unknown",
			Prompt:      "",
			Email:       nil,
			Confidence:  0,
			Reason:      "评论未包含有效邮箱",
		}, nil
	}

	userPrompt := strings.Replace(UserPromptTemplate, "<<<COMMENT>>>", comment, 1)

	intentResult, err := s.callLLM(ctx, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	intentResult.Email = email
	intentResult.RawJSON = nil

	if !s.isClearIntent(intentResult, threshold) {
		intentResult.HasRequest = false
		intentResult.Reason = fmt.Sprintf("意图不明确: %s", intentResult.Reason)
	}

	return intentResult, nil
}

func (s *Service) extractEmail(comment string) *string {
	matches := emailRegex.FindAllString(comment, -1)
	if len(matches) == 0 {
		return nil
	}

	email := strings.TrimSpace(matches[0])
	if !s.isValidEmail(email) {
		return nil
	}

	if len(matches) > 1 {
		return &email
	}

	return &email
}

func (s *Service) isValidEmail(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if len(parts[0]) == 0 || len(parts[0]) > 64 {
		return false
	}

	domainParts := strings.Split(parts[1], ".")
	if len(domainParts) < 2 {
		return false
	}

	for _, part := range domainParts {
		if len(part) == 0 {
			return false
		}
	}

	return true
}

func (s *Service) hasGenerationKeywords(comment string) bool {
	lowerComment := strings.ToLower(comment)

	for _, kw := range imageKeywords {
		if strings.Contains(lowerComment, strings.ToLower(kw)) {
			return true
		}
	}

	for _, kw := range videoKeywords {
		if strings.Contains(lowerComment, strings.ToLower(kw)) {
			return true
		}
	}

	return false
}

func (s *Service) isClearIntent(result *IntentResult, threshold float64) bool {
	if !result.HasRequest {
		return false
	}

	if result.RequestType != "image" && result.RequestType != "video" {
		return false
	}

	if len(result.Prompt) < 8 {
		return false
	}

	if result.Email == nil {
		return false
	}

	if result.Confidence < threshold {
		return false
	}

	return true
}

func (s *Service) callLLM(ctx context.Context, userPrompt string) (*IntentResult, error) {
	reqBody := LLMRequest{
		Model: s.cfg.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: SystemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LLM request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", s.cfg.BaseURL)

	var lastErr error
	for i := 0; i < s.cfg.MaxRetries; i++ {
		result, err := s.httpClient.Do(reqJSON)
		if err == nil {
			break
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if lastErr != nil {
		return nil, lastErr
	}

	var intentResult IntentResult
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", result)), &intentResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal intent result: %w", err)
	}

	return &intentResult, nil
}

type defaultHTTPClient struct {
	timeout time.Duration
}

func (c *defaultHTTPClient) Do(req interface{}) (*interface{}, error) {
	return nil, fmt.Errorf("not implemented - use real HTTP client in production")
}
