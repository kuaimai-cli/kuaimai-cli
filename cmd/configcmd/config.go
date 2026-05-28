package configcmd

import (
	"fmt"
	"os"

	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/config"
	"github.com/kuaimai/kuaimai-cli/internal/output"
	"github.com/spf13/cobra"
)

// Register adds config subcommands to root.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理",
	}
	cmd.AddCommand(initCmd(), getCmd(), setCmd())
	root.AddCommand(cmd)
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "初始化本地配置文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			created, err := config.Init()
			if err != nil {
				return friendly(err)
			}
			f, err := cmdutil.NewFactory()
			if err != nil {
				return friendly(err)
			}
			p := f.Printer()
			msg := "配置已初始化"
			if !created {
				msg = "配置文件已存在，未覆盖"
			}
			return p.Success(map[string]string{
				"message": msg,
				"path":    config.ConfigPath(),
			})
		},
	}
}

func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "查看配置项，省略 key 时输出全部",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return friendly(err)
			}
			p := f.Printer()
			if len(args) == 0 {
				return p.Success(f.Config.AllSettings())
			}
			val, err := f.Config.Get(args[0])
			if err != nil {
				return p.Fail(err.Error(), "使用 kuaimai-cli config init 初始化配置")
			}
			return p.Success(map[string]any{args[0]: val})
		},
	}
}

func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "设置配置项",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := cmdutil.NewFactory()
			if err != nil {
				return friendly(err)
			}
			if err := f.Config.Set(args[0], args[1]); err != nil {
				return friendly(err)
			}
			return f.Printer().Success(map[string]string{
				"message": "配置已更新",
				"key":     args[0],
				"value":   args[1],
			})
		},
	}
}

func friendly(err error) error {
	fmt.Fprintln(os.Stderr, err.Error())
	return err
}

// HandleRun wraps RunE with stdout envelope on failure.
func HandleRun(run func() error, p *output.Printer) error {
	if err := run(); err != nil {
		if p != nil {
			_ = p.Fail(err.Error(), "")
		}
		return err
	}
	return nil
}
