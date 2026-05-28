package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kuaimai/kuaimai-cli/internal/auth"
	"github.com/kuaimai/kuaimai-cli/internal/client"
	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/core"
	"github.com/kuaimai/kuaimai-cli/pkg/logger"
	"github.com/spf13/cobra"
)

// Register adds raw API commands: kuaimai-cli api GET /path
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "原始 API 兜底调用",
	}
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		m := method
		sub := &cobra.Command{
			Use:   fmt.Sprintf("%s <path>", m),
			Short: fmt.Sprintf("HTTP %s 请求", m),
			Args:  cobra.MinimumNArgs(1),
			RunE: func(c *cobra.Command, args []string) error {
				return runAPI(m, strings.Join(args, " "), c)
			},
		}
		sub.Flags().String("body", "", "请求体 JSON（POST/PUT）")
		cmd.AddCommand(sub)
	}
	root.AddCommand(cmd)
}

func runAPI(method, path string, c *cobra.Command) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	p := f.Printer()

	if err := f.RequireAuth(); err != nil {
		hint := "请先执行 kuaimai-cli auth login <accessToken> 完成登录"
		var authErr *cmdutil.AuthRequiredError
		if errors.As(err, &authErr) {
			hint = authErr.Hint()
		}
		_ = p.Fail(err.Error(), hint)
		return err
	}

	httpClient, err := f.HTTPClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	var body []byte
	if bodyStr, _ := c.Flags().GetString("body"); bodyStr != "" {
		body = []byte(bodyStr)
	}

	if client.IsWriteMethod(method) && core.Ctx.DryRun {
		fullURL := client.ResolveAPIURL(f.Config.APIURL(), path)
		out := map[string]any{
			"dry_run": true,
			"method":  method,
			"url":     fullURL,
			"body":    string(body),
		}
		if f.Auth.IsLoggedIn() {
			if tok, err := f.Auth.GetToken(); err == nil {
				out["auth_header"] = auth.HeaderAccessToken
				out["token_preview"] = auth.TokenPreview(tok)
			}
		}
		return p.Success(out)
	}

	var data any
	if core.Ctx.PageAll && method == "GET" {
		data, err = httpClient.RequestAllPages(context.Background(), method, path, body)
	} else {
		data, _, err = httpClient.Request(context.Background(), method, path, body)
	}
	if err != nil {
		var bizErr *client.BusinessError
		if errors.As(err, &bizErr) {
			if bizErr.TokenPreview != "" {
				logger.Info("请求头 %s 已设置（预览: %s）", auth.HeaderAccessToken, bizErr.TokenPreview)
			}
			_ = p.Fail(bizErr.Error(), bizErr.Hint)
			return err
		}
		if apiErr, ok := err.(*client.APIError); ok {
			_ = p.Fail(apiErr.Message, apiErr.Hint)
			return err
		}
		fmt.Fprintln(os.Stderr, err.Error())
		_ = p.Fail(err.Error(), "请检查网络与 api.url 配置")
		return err
	}
	return p.Success(data)
}
