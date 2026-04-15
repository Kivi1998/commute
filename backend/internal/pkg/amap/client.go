package amap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client 高德 Web Service 客户端
type Client struct {
	key     string
	baseURL string
	http    *http.Client
}

type Config struct {
	Key     string
	BaseURL string
	Timeout time.Duration
}

func New(c Config) *Client {
	if c.BaseURL == "" {
		c.BaseURL = "https://restapi.amap.com"
	}
	if c.Timeout == 0 {
		c.Timeout = 10 * time.Second
	}
	return &Client{
		key:     c.Key,
		baseURL: c.BaseURL,
		http:    &http.Client{Timeout: c.Timeout},
	}
}

// APIError 高德业务错误（infocode 非 10000）
type APIError struct {
	InfoCode string
	Info     string
	Status   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("amap: infocode=%s status=%s info=%s", e.InfoCode, e.Status, e.Info)
}

// IsQuotaExceeded 配额超限
func (e *APIError) IsQuotaExceeded() bool { return e.InfoCode == "10003" }

// ErrKeyNotConfigured 未配置 Key
var ErrKeyNotConfigured = errors.New("amap: key not configured")

// doGet 拼接 key + params，请求并解析到 out。只检查 status 字段，业务错误抛 APIError。
func (c *Client) doGet(ctx context.Context, path string, params url.Values, out any) error {
	if c.key == "" {
		return ErrKeyNotConfigured
	}
	params.Set("key", c.key)
	fullURL := c.baseURL + path + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("amap request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("amap read body: %w", err)
	}

	// 先探测状态字段
	var envelope struct {
		Status   string `json:"status"`
		Info     string `json:"info"`
		InfoCode string `json:"infocode"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("amap decode envelope: %w; body=%s", err, truncate(string(body), 200))
	}
	if envelope.Status != "1" {
		return &APIError{
			InfoCode: envelope.InfoCode, Info: envelope.Info, Status: envelope.Status,
		}
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("amap decode result: %w; body=%s", err, truncate(string(body), 200))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
