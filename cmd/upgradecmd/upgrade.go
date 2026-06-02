package upgradecmd

import (
	"fmt"
	"os"

	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/skill"
	"github.com/kuaimai-cli/kuaimai-cli/internal/upgrade"
	"github.com/spf13/cobra"
)

// Register adds upgrade command.
func Register(root *cobra.Command) {
	root.AddCommand(upgradeCmd())
}

func upgradeCmd() *cobra.Command {
	var repo string
	var checkOnly bool
	c := &cobra.Command{
		Use:   "upgrade",
		Short: "检查并升级到 GitHub / npm 最新版本",
		Long:  "默认行为：若有新版本则通过 npm 全局安装 @kuaimai-cli/cli@latest，并同步 Skills；仅检查请加 --check-only。",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			res, err := upgrade.CheckLatest(repo)
			if err != nil {
				_ = f.Printer().Fail(err.Error(), "请检查网络或稍后重试")
				return err
			}
			if checkOnly || !res.UpdateAvail {
				return f.Printer().Success(res)
			}

			fmt.Fprintln(os.Stderr, "[kuaimai-cli] 正在升级…")
			if err := upgrade.Apply(); err != nil {
				_ = f.Printer().Fail(err.Error(), res.Hint)
				return err
			}
			fmt.Fprintln(os.Stderr, upgrade.ApplyHint(res.Latest))

			opts := skill.InstallOptions{Repo: repo}
			syncErr := skill.SyncDefaultsAfterUpgrade(res.Latest, opts)
			skillsSynced := syncErr == nil
			if syncErr != nil {
				fmt.Fprintf(os.Stderr, "[kuaimai-cli] 升级完成，但 Skill 同步失败: %v\n", syncErr)
			} else {
				fmt.Fprintln(os.Stderr, "[kuaimai-cli] Skills 已同步至最新 Release")
			}

			res.Hint = "升级命令已执行，请在新进程中运行 kuaimai-cli --version 确认"
			return f.Printer().Success(map[string]any{
				"message":         "upgrade triggered",
				"check":           res,
				"skills_synced":   skillsSynced,
				"reopen_terminal": true,
			})
		},
	}
	c.Flags().StringVar(&repo, "repo", "", "GitHub 仓库 owner/name（默认 kuaimai-cli/kuaimai-cli）")
	c.Flags().BoolVar(&checkOnly, "check-only", false, "仅检查版本，不执行 npm 安装")
	return c
}
