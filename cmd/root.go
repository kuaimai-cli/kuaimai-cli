package cmd

import (
	"os"

	"github.com/kuaimai/kuaimai-cli/cmd/api"
	"github.com/kuaimai/kuaimai-cli/cmd/authcmd"
	"github.com/kuaimai/kuaimai-cli/cmd/completion"
	"github.com/kuaimai/kuaimai-cli/cmd/configcmd"
	"github.com/kuaimai/kuaimai-cli/cmd/doctorcmd"
	"github.com/kuaimai/kuaimai-cli/cmd/schemacmd"
	"github.com/kuaimai/kuaimai-cli/cmd/servicecmd"
	"github.com/kuaimai/kuaimai-cli/cmd/skillcmd"
	"github.com/kuaimai/kuaimai-cli/cmd/upgradecmd"
	"github.com/kuaimai/kuaimai-cli/internal/build"
	"github.com/kuaimai/kuaimai-cli/internal/config"
	"github.com/kuaimai/kuaimai-cli/internal/core"
	"github.com/kuaimai/kuaimai-cli/internal/output"
	"github.com/kuaimai/kuaimai-cli/pkg/logger"
	"github.com/kuaimai/kuaimai-cli/shortcuts/item"
	"github.com/spf13/cobra"
)

func resolveOutputFormat(cmd *cobra.Command) {
	format := output.FormatTable
	if cfg, err := config.New(); err == nil {
		if o := cfg.CLIOutput(); o != "" {
			if f, ok := output.ParseFormat(o); ok {
				format = f
			}
		}
	}
	if cmd.Flags().Changed("output") {
		out, _ := cmd.Flags().GetString("output")
		if f, ok := output.ParseFormat(out); ok {
			format = f
		}
	}
	core.Ctx.Output = format
}

func resolveColor(cmd *cobra.Command) {
	if core.Ctx.NoColor || os.Getenv("NO_COLOR") != "" {
		core.Ctx.NoColor = true
		return
	}
	if cfg, err := config.New(); err == nil {
		core.Ctx.NoColor = !cfg.CLIColorEnabled()
	}
	if cmd.Flags().Changed("no-color") {
		noColor, _ := cmd.Flags().GetBool("no-color")
		if noColor {
			core.Ctx.NoColor = true
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "kuaimai-cli",
	Short: "快麦业务专属命令行工具",
	Long:  "kuaimai-cli 快麦业务专属 CLI 工具，提供配置、鉴权、API 与 erp-items-core 商品快捷命令。",
	Version: build.Version + " (" + build.Date + ")",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&core.Ctx.Verbose, "verbose", false, "开启详细调试日志")
	rootCmd.PersistentFlags().BoolVar(&core.Ctx.DryRun, "dry-run", false, "试运行，不实际发送写请求")
	rootCmd.PersistentFlags().BoolVar(&core.Ctx.PageAll, "page-all", false, "列表命令自动拉取全部分页")
	rootCmd.PersistentFlags().BoolVar(&core.Ctx.NoColor, "no-color", false, "禁用终端彩色输出")
	rootCmd.PersistentFlags().String("output", "table", "输出格式: table|json|csv|ndjson")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		logger.SetVerbose(core.Ctx.Verbose)
		if core.Ctx.Verbose {
			logger.Debug("verbose mode enabled")
		}
		resolveOutputFormat(cmd)
		resolveColor(cmd)
	}

	configcmd.Register(rootCmd)
	authcmd.Register(rootCmd)
	api.Register(rootCmd)
	schemacmd.Register(rootCmd)
	servicecmd.Register(rootCmd)
	item.Register(rootCmd)
	skillcmd.Register(rootCmd)
	completion.Register(rootCmd)
	upgradecmd.Register(rootCmd)
	doctorcmd.Register(rootCmd)
}
