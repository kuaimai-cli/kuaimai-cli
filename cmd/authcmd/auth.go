package authcmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/kuaimai/kuaimai-cli/internal/auth"
	"github.com/kuaimai/kuaimai-cli/internal/client"
	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/config"
	"github.com/kuaimai/kuaimai-cli/pkg/logger"
	"github.com/kuaimai/kuaimai-cli/shortcuts/item"
	"github.com/spf13/cobra"
)

// Register adds auth subcommands.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "鉴权管理",
	}
	cmd.AddCommand(loginCmd(), logoutCmd(), statusCmd(), checkCmd(), listCmd(), useCmd())
	root.AddCommand(cmd)
}

func loginCmd() *cobra.Command {
	var profile string
	var apiURL string
	c := &cobra.Command{
		Use:   "login <token>",
		Short: "将 accessToken 写入系统密钥链（唯一鉴权方式）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.New()
			if err != nil {
				return err
			}
			name := cfg.ActiveProfile()
			if profile != "" {
				name = auth.NormalizeProfile(profile)
			}
			if err := auth.LoginWithToken(name, args[0]); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				f, _ := cmdutil.NewFactory()
				if f != nil {
					_ = f.Printer().Fail(err.Error(), "用法: kuaimai-cli auth login <accessToken>")
				}
				return err
			}
			if apiURL != "" {
				if err := cfg.SetProfileAPIURL(name, apiURL); err != nil {
					return err
				}
			}
			logger.Info("登录成功，profile=%s，accessToken 已写入系统密钥链", name)
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			return f.Printer().Success(map[string]any{
				"message": "登录成功",
				"profile": name,
				"storage": "system keyring",
				"header":  auth.HeaderAccessToken,
				"api_url": cfg.ProfileAPIURL(name),
			})
		},
	}
	c.Flags().StringVar(&profile, "profile", "", "账号 profile 名称（默认当前 auth.profile）")
	c.Flags().StringVar(&apiURL, "api-url", "", "该 profile 的 API 地址（可选，覆盖全局 api.url）")
	return c
}

func logoutCmd() *cobra.Command {
	var profile string
	c := &cobra.Command{
		Use:   "logout",
		Short: "登出并清除密钥链中的 accessToken",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			name := f.Auth.Profile()
			if profile != "" {
				st, err := auth.NewStoreForProfile(profile)
				if err != nil {
					return err
				}
				name = st.Profile()
				if err := st.DeleteToken(); err != nil {
					fmt.Fprintln(os.Stderr, "清除凭证时出现问题（可能尚未登录）")
				}
				return f.Printer().Success(map[string]string{
					"message": "已登出",
					"profile": name,
				})
			}
			if err := f.Auth.DeleteToken(); err != nil {
				fmt.Fprintln(os.Stderr, "清除凭证时出现问题（可能尚未登录）")
			}
			return f.Printer().Success(map[string]string{
				"message": "已登出",
				"profile": name,
			})
		},
	}
	c.Flags().StringVar(&profile, "profile", "", "仅登出指定 profile（默认当前 profile）")
	return c
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "查看登录状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			loggedIn := f.Auth.IsLoggedIn()
			status := "未登录"
			if loggedIn {
				status = "已登录"
			}
			result := map[string]any{
				"logged_in": loggedIn,
				"status":    status,
				"profile":   f.Auth.Profile(),
				"storage":   "system keyring",
				"header":    auth.HeaderAccessToken,
				"api_url":   f.Config.ProfileAPIURL(f.Auth.Profile()),
			}
			if loggedIn {
				if tok, err := f.Auth.GetToken(); err == nil {
					result["token_preview"] = auth.TokenPreview(tok)
				}
			}
			return f.Printer().Success(result)
		},
	}
}

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "探测当前 profile 的 accessToken 是否可用",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			if err := f.RequireAuth(); err != nil {
				_ = f.Printer().Fail(err.Error(), "请先执行 kuaimai-cli auth login <accessToken>")
				return err
			}
			httpClient, err := f.HTTPClient()
			if err != nil {
				return err
			}
			body := map[string]any{"pageNo": 1, "pageSize": 1}
			item.ApplyStockListDefaults(body)
			_, _, err = httpClient.PostForm(context.Background(), item.QueryCountPath, client.MapToFormValues(body))
			if err != nil {
				hint := "请确认 accessToken 有效且 api.url 正确"
				var bizErr *client.BusinessError
				if errors.As(err, &bizErr) && bizErr.Hint != "" {
					hint = bizErr.Hint
				}
				_ = f.Printer().Fail(err.Error(), hint)
				return err
			}
			return f.Printer().Success(map[string]any{
				"valid":   true,
				"profile": f.Auth.Profile(),
				"api_url": f.Config.ProfileAPIURL(f.Auth.Profile()),
				"probe":   "POST " + item.QueryCountPath,
			})
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出已注册的 profile 及登录状态",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			profiles := auth.ListProfilesWithLoginStatus(f.Config)
			return f.Printer().Success(map[string]any{
				"active_profile": f.Config.ActiveProfile(),
				"profiles":       profiles,
			})
		},
	}
}

func useCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "切换当前使用的 profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			name := auth.NormalizeProfile(args[0])
			if err := f.Config.SetActiveProfile(name); err != nil {
				return err
			}
			st, _ := auth.NewStoreForProfile(name)
			loggedIn := st != nil && st.IsLoggedIn()
			return f.Printer().Success(map[string]any{
				"message":   "已切换 profile",
				"profile":   name,
				"logged_in": loggedIn,
				"api_url":   f.Config.ProfileAPIURL(name),
				"hint":      hintForUse(loggedIn),
			})
		},
	}
}

func hintForUse(loggedIn bool) string {
	if loggedIn {
		return ""
	}
	return "该 profile 尚未登录，请执行 kuaimai-cli auth login --profile <name> <accessToken>"
}
