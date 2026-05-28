package skillcmd

import (
	"fmt"
	"os"

	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/skill"
	"github.com/spf13/cobra"
)

// Register adds skill platform commands.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Skill 文档仓库（list / install / install-all）",
	}
	cmd.AddCommand(listCmd())
	cmd.AddCommand(installCmd())
	cmd.AddCommand(installAllCmd())
	root.AddCommand(cmd)
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出本地 Skill（./skills 与 ~/.agents/skills）",
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
	var fromDir, repo, gitRef string
	c := &cobra.Command{
		Use:   "install <name>",
		Short: "安装单个 Skill 到 ~/.agents/skills/<name>/（默认从 GitHub 仓库拉取）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			name := args[0]
			dest, err := skill.Install(name, skill.InstallOptions{
				FromDir: fromDir,
				Repo:    repo,
				Ref:     gitRef,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return err
			}
			return f.Printer().Success(map[string]string{
				"message": "Skill 已安装",
				"name":    name,
				"path":    dest,
			})
		},
	}
	c.Flags().StringVar(&fromDir, "from", "", "本地 skills 目录（安装 <name>/SKILL.md，如 ./skills）")
	c.Flags().StringVar(&repo, "repo", skill.DefaultGitHubRepo(), "GitHub 仓库（未指定 --from 时使用）")
	c.Flags().StringVar(&gitRef, "ref", "main", "GitHub 分支或 tag")
	return c
}

func installAllCmd() *cobra.Command {
	var fromDir, repo, gitRef string
	c := &cobra.Command{
		Use:   "install-all",
		Short: "批量安装 Skill 到 ~/.agents/skills/（本地目录或 GitHub 仓库）",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return err
			}
			var results []skill.InstallResult
			switch {
			case fromDir != "":
				results, err = skill.InstallAllFromDir(fromDir)
			default:
				results, err = skill.InstallAllFromGitHub(repo, gitRef)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return err
			}
			return f.Printer().Success(map[string]any{
				"message": "Skills 已安装",
				"count":   len(results),
				"skills":  results,
			})
		},
	}
	c.Flags().StringVar(&fromDir, "from", "", "本地 skills 目录（如 ./skills）")
	c.Flags().StringVar(&repo, "repo", skill.DefaultGitHubRepo(), "GitHub 仓库（默认 kuaimai/kuaimai-cli）")
	c.Flags().StringVar(&gitRef, "ref", "main", "GitHub 分支或 tag")
	return c
}
