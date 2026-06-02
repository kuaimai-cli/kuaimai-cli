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
		Short: "Skill 文档（list / install，从 GitHub 安装到各 Agent 目录）",
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
	c := &cobra.Command{
		Use:   "install [name...]",
		Short: "从 GitHub 安装 Skill 到各 Agent 目录（无参数时安装 kuaimai-shared、kuaimai-item）",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			names := args
			if len(names) == 0 {
				names = skill.DefaultSkillNames
			}
			opts := skill.InstallOptions{Repo: repo, Ref: gitRef}
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
	c.Flags().StringVar(&gitRef, "ref", "main", "GitHub 分支或 tag")
	return c
}
