package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kuaimai/kuaimai-cli/internal/auth"
	"github.com/kuaimai/kuaimai-cli/internal/config"
	"github.com/kuaimai/kuaimai-cli/pkg/convert"
	"github.com/kuaimai/kuaimai-cli/pkg/logger"
	"github.com/kuaimai/kuaimai-cli/pkg/sanitize"
)

// Client is the unified HTTP client for kuaimai API.
type Client struct {
	baseURL      string
	httpClient   *http.Client
	authStore    *auth.Store
	dryRun       bool
	maxRetry int
}

// New builds a client from config and auth store.
func New(cfg *config.Manager, store *auth.Store, dryRun bool) (*Client, error) {
	timeout := cfg.APITimeout()
	transport := newPooledTransport(TransportOptions{
		MaxIdleConns:        cfg.APIPoolMaxIdle(),
		MaxIdleConnsPerHost: cfg.APIPoolMaxIdlePerHost(),
		CircuitThreshold:    cfg.APICircuitThreshold(),
		CircuitCooldown:     cfg.APICircuitCooldown(),
	})
	profile := "default"
	if store != nil {
		profile = store.Profile()
	}
	base := cfg.ProfileAPIURL(profile)
	return &Client{
		baseURL: strings.TrimRight(base, "/"),
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		authStore: store,
		dryRun:    dryRun,
		maxRetry:  cfg.APIRetry(),
	}, nil
}

// Request performs an HTTP call and returns parsed JSON body when possible.
func (c *Client) Request(ctx context.Context, method, path string, body []byte) (any, int, error) {
	return c.requestWithRetry(ctx, method, path, body, "")
}

func (c *Client) requestWithRetry(ctx context.Context, method, path string, body []byte, contentType string) (any, int, error) {
	method = strings.ToUpper(method)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	urlStr := c.baseURL + path

	if c.dryRun {
		logger.Info("dry-run: %s %s", method, urlStr)
		return map[string]any{
			"dry_run": true,
			"method":  method,
			"url":     urlStr,
		}, 200, nil
	}

	var lastErr error
	attempts := c.maxRetry + 1
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(500*(1<<(attempt-1))) * time.Millisecond
			logger.Debug("retry %d/%d after %v", attempt, c.maxRetry, backoff)
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(backoff):
			}
		}

		data, code, err := c.doOnce(ctx, method, urlStr, body, contentType)
		if err == nil {
			return data, code, nil
		}
		lastErr = err

		if !isRetryable(err) {
			return data, code, err
		}
	}
	return nil, 0, lastErr
}

func (c *Client) doOnce(ctx context.Context, method, urlStr string, body []byte, contentType string) (any, int, error) {
	token, err := c.authStore.GetToken()
	if err != nil {
		return nil, 0, fmt.Errorf("未登录，请先执行 kuaimai-cli auth login <accessToken>")
	}

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set(auth.HeaderAccessToken, token)
	if bodyReader != nil {
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	}

	preview := auth.TokenPreview(token)
	logger.Debug("%s %s", method, urlStr)
	logger.Debug("request header %s: %s (len=%d)", auth.HeaderAccessToken, preview, len(token))
	if bodyReader != nil {
		logger.Debug("request Content-Type: %s", contentType)
		if len(body) > 0 && len(body) <= 4096 {
			logger.Debug("request body: %s", sanitize.JSONString(string(body)))
		} else if len(body) > 4096 {
			logger.Debug("request body: (%d bytes, truncated in log)", len(body))
		}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, friendlyNetworkErr(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %w", err)
	}

	var parsed any
	if len(raw) > 0 {
		parsed, err = convert.ToMap(raw)
		if err != nil {
			parsed = string(raw)
		}
	}

	if looksLikeHTML(parsed) {
		return parsed, resp.StatusCode, &APIError{
			Message: "接口返回了 HTML 页面（快麦通前端），不是业务 JSON",
			Status:  resp.StatusCode,
			Hint:    "请确认 api.url 为 https://erp1.superboss.cc/ 且路径与浏览器 DevTools 中一致",
		}
	}

	if resp.StatusCode >= 400 {
		msg := fmt.Sprintf("接口返回 %d", resp.StatusCode)
		if m, ok := parsed.(map[string]any); ok {
			if e, ok := m["message"].(string); ok && e != "" {
				msg = e
			}
		}
		hint := statusHint(resp.StatusCode)
		return parsed, resp.StatusCode, &APIError{Message: msg, Status: resp.StatusCode, Hint: hint}
	}

	if biz := ParseBusinessError(parsed); biz != nil {
		return parsed, resp.StatusCode, EnrichBusinessError(biz, preview)
	}

	return parsed, resp.StatusCode, nil
}

func looksLikeHTML(body any) bool {
	s, ok := body.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "<!DOCTYPE") ||
		strings.HasPrefix(s, "<!doctype") ||
		strings.HasPrefix(s, "<html") ||
		strings.HasPrefix(s, "<HTML")
}

// RequestAllPages fetches all pages when page-all is enabled.
func (c *Client) RequestAllPages(ctx context.Context, method, path string, body []byte) (any, error) {
	if c.dryRun {
		data, _, err := c.Request(ctx, method, path, body)
		return data, err
	}

	allItems := make([]map[string]any, 0)
	page := 1
	const pageSize = 100

	for {
		pagedPath, err := withPageParams(path, page, pageSize)
		if err != nil {
			return nil, err
		}
		data, _, err := c.Request(ctx, method, pagedPath, body)
		if err != nil {
			return nil, err
		}
		items, hasMore := extractPageItems(data, page)
		allItems = append(allItems, items...)
		if !hasMore || len(items) == 0 {
			break
		}
		page++
		if page > 1000 {
			break
		}
	}
	return allItems, nil
}

func withPageParams(path string, page, pageSize int) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return path, err
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()
	if u.Scheme == "" {
		return u.Path + "?" + u.RawQuery, nil
	}
	return u.String(), nil
}

func extractPageItems(data any, page int) ([]map[string]any, bool) {
	items := extractItems(data)
	hasMore := false
	if m, ok := data.(map[string]any); ok {
		if p, ok := m["pagination"].(map[string]any); ok {
			if hp, ok := p["has_more"].(bool); ok {
				hasMore = hp
			} else if tp, ok := p["total_pages"].(float64); ok {
				hasMore = float64(page) < tp
			}
		}
		if _, ok := m["next_page"]; ok {
			hasMore = true
		}
		if !hasMore && len(items) >= 100 {
			hasMore = true
		}
	}
	return items, hasMore
}

func extractItems(data any) []map[string]any {
	var raw []any
	switch v := data.(type) {
	case []any:
		raw = v
	case map[string]any:
		for _, key := range []string{"data", "items", "list", "records"} {
			if inner, ok := v[key]; ok {
				return extractItems(inner)
			}
		}
		return []map[string]any{v}
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// APIError represents a non-2xx API response.
type APIError struct {
	Message string
	Status  int
	Hint    string
}

func (e *APIError) Error() string {
	return e.Message
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
	}
	return false
}

func friendlyNetworkErr(err error) error {
	if err == nil {
		return nil
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return fmt.Errorf("请求超时，请检查网络或调大 api.timeout 配置")
		}
	}
	return fmt.Errorf("无法连接 API 服务，请检查 api.url 配置与服务是否可用")
}

func statusHint(code int) string {
	switch code {
	case http.StatusNotFound:
		return "请检查路径是否正确，或使用 kuaimai-cli --help 查看可用命令"
	case http.StatusUnauthorized:
		return "凭证无效或已过期，请执行 kuaimai-cli auth login <accessToken> 重新登录"
	case http.StatusInternalServerError:
		return "服务端异常，请稍后重试或联系平台管理员"
	default:
		return "请使用 --verbose 查看详细日志"
	}
}

// PostJSON marshals v and POSTs to path.
func (c *Client) PostJSON(ctx context.Context, path string, v any) (any, int, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, 0, err
	}
	return c.Request(ctx, http.MethodPost, path, b)
}

// PostForm POSTs application/x-www-form-urlencoded body to path.
func (c *Client) PostForm(ctx context.Context, path string, form url.Values) (any, int, error) {
	return c.requestWithRetry(ctx, http.MethodPost, path, []byte(form.Encode()), "application/x-www-form-urlencoded")
}

// IsWriteMethod reports whether HTTP method mutates server state.
func IsWriteMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
