package completion

import (
	"os"

	"github.com/spf13/cobra"
)

// Register adds shell completion subcommands.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "生成 shell 自动补全脚本",
	}
	cmd.AddCommand(bashCmd(), zshCmd(), powershellCmd())
	root.AddCommand(cmd)
}

func bashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "生成 Bash 补全脚本",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletion(os.Stdout)
		},
	}
}

func zshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "生成 Zsh 补全脚本",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(os.Stdout)
		},
	}
}

func powershellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "powershell",
		Short: "生成 PowerShell 补全脚本",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		},
	}
}
