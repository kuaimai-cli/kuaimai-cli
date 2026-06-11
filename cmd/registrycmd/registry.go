package registrycmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/config"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/spf13/cobra"
)

// Register adds registry sync commands.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Registry 同步与管理",
	}
	cmd.AddCommand(syncCmd())
	cmd.AddCommand(watchCmd())
	root.AddCommand(cmd)
}

func syncCmd() *cobra.Command {
	var source string
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "从远程 HTTP 服务拉取 registry.json 到本地缓存",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(source)
		},
	}
	cmd.Flags().StringVar(&source, "source", "", "registry.json URL（默认读取 config registry.source）")
	return cmd
}

func runSync(sourceFlag string) error {
	source := sourceFlag
	if source == "" {
		cfg, err := config.New()
		if err != nil {
			return err
		}
		source = cfg.RegistrySource()
	}

	result, err := registry.Sync(context.Background(), source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	msg := "registry 已同步到本地"
	if result.NotModified || !result.Changed {
		msg = "registry 无变化，本地缓存已是最新"
	}
	return f.Printer().Success(map[string]any{
		"message":       msg,
		"source":        result.Source,
		"path":          result.Path,
		"version":       result.Version,
		"api_count":     result.APICount,
		"changed":       result.Changed,
		"not_modified":  result.NotModified,
		"hint":          "使用 kuaimai-cli schema --output json 查看接口；web call <apiId> 调用",
	})
}
