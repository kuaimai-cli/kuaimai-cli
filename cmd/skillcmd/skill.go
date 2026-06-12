package skillcmd

import (
	"fmt"
	"os"

	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/skill"
	"github.com/spf13/cobra"
)

// Register adds skill platform commands.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Skill 文档（list / install，优先 npm 包内置 skills，回退 GitHub）",
	}
	cmd.AddCommand(listCmd())
	cmd.AddCommand(installCmd())
	root.AddCommand(cmd)
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出已安装的 Skill（~/.agents/skills、~/.cursor/skills 等）",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			entries, err := skill.List()
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return err
			}
			return f.Printer().Success(entries)
		},
	}
}

func installCmd() *cobra.Command {
	var repo, gitRef string
	var ifStale bool
	var force bool
	c := &cobra.Command{
		Use:   "install [name...]",
		Short: "安装 Skill 到各 Agent 目录（无参数时安装 kuaimai-shared、kuaimai-erp-item、kuaimai-scm-item）",
		Long: "覆盖写入各 Agent 的 skill 目录；默认 Skills 会先删除旧默认目录再重装（不删除非默认用户 skill）。\n" +
			"优先从 npm 包内置 skills/ 或仓库 skills/ 复制；不可用时回退 GitHub。\n" +
			"无参数时默认仅在未安装、Release 或 CLI 版本变化时更新（等同 --if-stale）；" +
			"任意命令结束后也会在后台自动同步（24h 缓存，见 KUAIMAI_CLI_SKIP_SKILL_SYNC）。",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			opts := skill.InstallOptions{Repo: repo, Ref: gitRef}

			useStale := (ifStale || len(args) == 0) && !force
			if useStale && len(args) == 0 {
				updated, results, err := skill.InstallIfStale(opts)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					return err
				}
				if !updated {
					return f.Printer().Success(map[string]any{
						"message": "Skills 已是最新，无需更新",
						"stale":   false,
					})
				}
				return f.Printer().Success(map[string]any{
					"message": "Skills 已更新",
					"stale":   true,
					"count":   len(results),
					"skills":  results,
				})
			}

			names := args
			if len(names) == 0 {
				names = skill.DefaultSkillNames
			}
			var results []skill.InstallResult
			for _, name := range names {
				res, err := skill.Install(name, opts)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					return err
				}
				results = append(results, res)
			}
			return f.Printer().Success(map[string]any{
				"message": "Skills 已安装",
				"count":   len(results),
				"skills":  results,
			})
		},
	}
	c.Flags().StringVar(&repo, "repo", skill.DefaultGitHubRepo(), "GitHub 仓库（owner/repo）")
	c.Flags().StringVar(&gitRef, "ref", "main", "GitHub 分支或 tag（--if-stale 时默认用最新 Release tag）")
	c.Flags().BoolVar(&ifStale, "if-stale", false, "仅在未安装或 Release/CLI 版本变化时安装默认 Skills（无参数时已是默认行为）")
	c.Flags().BoolVar(&force, "force", false, "无参数时强制重装默认 Skills（忽略是否已最新）")
	return c
}
