package doubao

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 火山方舟（豆包）Chat Completions 客户端
type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

type Config struct {
	APIKey  string
	BaseURL string
	Model   string // 接入点 ID，形如 ep-xxxx
	Timeout time.Duration
}

func New(c Config) *Client {
	if c.BaseURL == "" {
		c.BaseURL = "https://ark.cn-beijing.volces.com"
	}
	if c.Timeout == 0 {
		c.Timeout = 60 * time.Second
	}
	return &Client{
		apiKey: c.APIKey, baseURL: c.BaseURL, model: c.Model,
		http: &http.Client{Timeout: c.Timeout},
	}
}

var ErrNotConfigured = errors.New("doubao: api key or model not configured")

// Configured 是否已配置（key + model 都有）
func (c *Client) Configured() bool {
	return c.apiKey != "" && c.model != ""
}

// Message 消息
type Message struct {
	Role    string `json:"role"` // system / user / assistant
	Content string `json:"content"`
}

// ResponseFormat 响应格式（json_object 宽松模式）
type ResponseFormat struct {
	Type string `json:"type"` // "json_object"
}

type ChatRequest struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	Temperature    *float64        `json:"temperature,omitempty"`
	TopP           *float64        `json:"top_p,omitempty"`
	MaxTokens      *int            `json:"max_tokens,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	Stream         bool            `json:"stream"`
}

type ChatChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

// APIError 豆包错误响应
type APIError struct {
	StatusCode int
	Type       string `json:"type"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Param      string `json:"param"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("doubao: http=%d type=%s code=%s: %s", e.StatusCode, e.Type, e.Code, e.Message)
}

// ChatCompletion 调用 Chat Completions 接口
func (c *Client) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	if req.Model == "" {
		req.Model = c.model
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("doubao marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v3/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("doubao request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("doubao read: %w", err)
	}

	if resp.StatusCode >= 400 {
		var envelope struct {
			Error APIError `json:"error"`
		}
		if err := json.Unmarshal(respBody, &envelope); err == nil && envelope.Error.Message != "" {
			envelope.Error.StatusCode = resp.StatusCode
			return nil, &envelope.Error
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: truncate(string(respBody), 200)}
	}

	var chat ChatResponse
	if err := json.Unmarshal(respBody, &chat); err != nil {
		return nil, fmt.Errorf("doubao decode: %w; body=%s", err, truncate(string(respBody), 200))
	}
	return &chat, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
