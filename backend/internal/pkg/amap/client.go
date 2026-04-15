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

// IsQuotaExceeded 日配额超限
func (e *APIError) IsQuotaExceeded() bool { return e.InfoCode == "10003" }

// IsQPSExceeded 并发/每秒限流（常见于免费 Key）
func (e *APIError) IsQPSExceeded() bool { return e.InfoCode == "10021" }

// ErrKeyNotConfigured 未配置 Key
var ErrKeyNotConfigured = errors.New("amap: key not configured")

// doGet 拼接 key + params，请求并解析到 out。对 QPS 超限做最多 2 次退避重试。
func (c *Client) doGet(ctx context.Context, path string, params url.Values, out any) error {
	if c.key == "" {
		return ErrKeyNotConfigured
	}
	params.Set("key", c.key)
	fullURL := c.baseURL + path + "?" + params.Encode()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			// 退避：500ms, 1200ms
			backoff := time.Duration(500+attempt*700) * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := c.doGetOnce(ctx, fullURL, out)
		if err == nil {
			return nil
		}
		lastErr = err

		var apiErr *APIError
		if !(errorsAs(err, &apiErr) && apiErr.IsQPSExceeded()) {
			return err // 非 QPS 错误直接抛
		}
	}
	return lastErr
}

func (c *Client) doGetOnce(ctx context.Context, fullURL string, out any) error {
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

// errorsAs 适配 errors.As 避免文件导入膨胀
func errorsAs(err error, target any) bool { return errors.As(err, target) }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
