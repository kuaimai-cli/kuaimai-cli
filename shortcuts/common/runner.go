package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kuaimai/kuaimai-cli/pkg/sanitize"

	"github.com/kuaimai/kuaimai-cli/internal/audit"
	"github.com/kuaimai/kuaimai-cli/internal/auth"
	"github.com/kuaimai/kuaimai-cli/internal/client"
	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/core"
	"github.com/kuaimai/kuaimai-cli/internal/output"
	"github.com/kuaimai/kuaimai-cli/pkg/logger"
)

// RunFunc performs the business HTTP call and returns data for stdout envelope.
type RunFunc func(ctx context.Context, c *client.Client) (any, error)

// ListOptions configures list-style shortcut execution.
type ListOptions struct {
	Method string
	Path   string
}

// WriteOptions configures write-style shortcut execution with dry-run support.
type WriteOptions struct {
	Method       string
	Path         string
	Body         any
	FormEncoded  bool // 与浏览器一致：application/x-www-form-urlencoded
}

// Runner executes shortcuts with auth, dry-run, and unified output.
type Runner struct {
	Factory *cmdutil.Factory
}

// NewRunner creates a runner from factory.
func NewRunner(f *cmdutil.Factory) *Runner {
	return &Runner{Factory: f}
}

// Execute runs the shortcut pipeline.
func (r *Runner) Execute(ctx context.Context, fn RunFunc) error {
	p := r.Factory.Printer()

	if err := r.Factory.RequireAuth(); err != nil {
		var authErr *cmdutil.AuthRequiredError
		if errors.As(err, &authErr) {
			_ = p.Fail(authErr.Error(), authErr.Hint())
			fmt.Fprintln(os.Stderr, authErr.Error()+": "+authErr.Hint())
			return err
		}
		_ = p.Fail(err.Error(), "")
		return err
	}

	httpClient, err := r.Factory.HTTPClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		_ = p.Fail(err.Error(), "请检查配置")
		return err
	}

	data, err := fn(ctx, httpClient)
	if err != nil {
		_ = audit.Record(auditCommandName(), "", "", "error")
		return r.handleError(p, err)
	}
	_ = audit.Record(auditCommandName(), "", "", "ok")
	return p.Success(data)
}

// ExecuteList runs a list command with optional --page-all pagination.
func (r *Runner) ExecuteList(ctx context.Context, opts ListOptions) error {
	method := opts.Method
	if method == "" {
		method = "GET"
	}
	return r.Execute(ctx, func(ctx context.Context, c *client.Client) (any, error) {
		if core.Ctx.PageAll {
			logger.Info("page-all: 自动拉取全部分页")
			data, err := c.RequestAllPages(ctx, method, opts.Path, nil)
			if err != nil {
				return nil, err
			}
			return NormalizeList(data), nil
		}
		data, _, err := c.Request(ctx, method, opts.Path, nil)
		if err != nil {
			return nil, err
		}
		return NormalizeList(data), nil
	})
}

// ExecuteWrite runs a write command; dry-run returns simulated payload without sending.
func (r *Runner) ExecuteWrite(ctx context.Context, opts WriteOptions) error {
	method := strings.ToUpper(opts.Method)
	if core.Ctx.DryRun {
		p := r.Factory.Printer()
		fullURL := opts.Path
		if cfg := r.Factory.Config; cfg != nil {
			fullURL = client.ResolveAPIURL(cfg.APIURL(), opts.Path)
		}
		out := map[string]any{
			"dry_run":      true,
			"method":       method,
			"path":         opts.Path,
			"url":          fullURL,
			"content_type": "application/json",
			"body":         opts.Body,
		}
		if opts.FormEncoded {
			out["content_type"] = "application/x-www-form-urlencoded"
			if m, ok := opts.Body.(map[string]any); ok {
				out["body_form"] = client.MapToFormValues(sanitize.Map(m)).Encode()
			}
		}
		if m, ok := opts.Body.(map[string]any); ok && !opts.FormEncoded {
			out["body"] = sanitize.Map(m)
		}
		appendAuthDryRunInfo(r, out)
		logger.Info("dry-run: %s %s", method, fullURL)
		return p.Success(out)
	}
	return r.Execute(ctx, func(ctx context.Context, c *client.Client) (any, error) {
		var data any
		var err error
		if opts.FormEncoded {
			m, ok := opts.Body.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("form 请求体必须是 map")
			}
			_ = audit.Record(auditCommandName(), opts.Method, opts.Path, "request")
			if core.Ctx.PageAll {
				logger.Info("page-all: 自动拉取全部分页（form pageNo/pageSize）")
				data, err = c.PostFormAllPages(ctx, opts.Path, m)
			} else {
				data, _, err = c.PostForm(ctx, opts.Path, client.MapToFormValues(m))
			}
		} else if opts.Body != nil {
			data, _, err = c.PostJSON(ctx, opts.Path, opts.Body)
		} else {
			data, _, err = c.Request(ctx, method, opts.Path, nil)
		}
		if err != nil {
			return nil, err
		}
		return data, nil
	})
}

func appendAuthDryRunInfo(r *Runner, out map[string]any) {
	if r == nil || r.Factory == nil || r.Factory.Auth == nil {
		return
	}
	if !r.Factory.Auth.IsLoggedIn() {
		out["auth_header"] = auth.HeaderAccessToken
		out["auth_status"] = "未登录（实际请求将失败，请先 auth login）"
		return
	}
	tok, err := r.Factory.Auth.GetToken()
	if err != nil {
		out["auth_status"] = "读取密钥链失败"
		return
	}
	out["auth_header"] = auth.HeaderAccessToken
	out["token_preview"] = auth.TokenPreview(tok)
	out["auth_status"] = "已登录，实际请求将携带 accessToken 头"
}

func (r *Runner) handleError(p *output.Printer, err error) error {
	var bizErr *client.BusinessError
	if errors.As(err, &bizErr) {
		if bizErr.TokenPreview != "" {
			logger.Info("请求头 %s 已设置（预览: %s）", auth.HeaderAccessToken, bizErr.TokenPreview)
		}
		logger.Error("%s (result=%d)", bizErr.Message, bizErr.Result)
		_ = p.Fail(bizErr.Error(), bizErr.Hint)
		return err
	}
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		logger.Error("%s", apiErr.Message)
		_ = p.Fail(apiErr.Message, apiErr.Hint)
		return err
	}
	fmt.Fprintln(os.Stderr, err.Error())
	hint := "请使用 --verbose 查看详情"
	if strings.Contains(err.Error(), "无法连接 API") {
		hint = "请执行 kuaimai-cli config set api.url <地址> 配置正确的 API 地址"
	}
	_ = p.Fail(err.Error(), hint)
	return err
}

// NormalizeList converts common API list payloads to table-friendly rows.
func NormalizeList(body any) any {
	if body == nil {
		return []map[string]any{}
	}
	if list, ok := body.([]any); ok {
		rows := make([]map[string]any, 0, len(list))
		for _, item := range list {
			if m, ok := item.(map[string]any); ok {
				rows = append(rows, m)
			}
		}
		return rows
	}
	if list, ok := body.([]map[string]any); ok {
		return list
	}
	if m, ok := body.(map[string]any); ok {
		for _, key := range []string{"data", "items", "list", "records"} {
			if v, ok := m[key]; ok {
				return NormalizeList(v)
			}
		}
		return []map[string]any{m}
	}
	return body
}

// ParseBodyJSON parses a JSON string flag into a map for write operations.
func ParseBodyJSON(raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("body 必须是合法 JSON: %w", err)
	}
	return m, nil
}

// EnsurePrinter is used by tests.
func EnsurePrinter(f *cmdutil.Factory) *output.Printer {
	return f.Printer()
}

// Friendly wraps errors without stack.
func Friendly(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s", err.Error())
}

func auditCommandName() string {
	if len(os.Args) < 2 {
		return "kuaimai-cli"
	}
	end := 3
	if end > len(os.Args) {
		end = len(os.Args)
	}
	return strings.Join(os.Args[1:end], " ")
}
