package upgradecmd

import (
	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/upgrade"
	"github.com/spf13/cobra"
)

// Register adds upgrade command.
func Register(root *cobra.Command) {
	root.AddCommand(upgradeCmd())
}

func upgradeCmd() *cobra.Command {
	var repo string
	c := &cobra.Command{
		Use:   "upgrade",
		Short: "检查 GitHub 是否有新版本",
		Long:  "对比当前 kuaimai-cli 与 GitHub Release 最新版本；安装请使用 npx @kuaimai/cli@latest install 或从 Release 页下载二进制。",
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
			return f.Printer().Success(res)
		},
	}
	c.Flags().StringVar(&repo, "repo", "", "GitHub 仓库 owner/name（默认 kuaimai/kuaimai-cli）")
	return c
}
