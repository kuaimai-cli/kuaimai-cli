package schemacmd

import (
	"fmt"
	"os"

	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/registry"
	"github.com/spf13/cobra"
)

// Register adds schema introspection command.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "API 元数据自省（OpenAPI 注册表）",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSchema()
		},
	}
	root.AddCommand(cmd)
}

func runSchema() error {
	meta, err := registry.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	rows := make([]map[string]any, 0)
	for _, item := range meta.AllOperations() {
		op := item.Operation
		rows = append(rows, map[string]any{
			"service":     item.Service,
			"operation":   op.Name,
			"method":      op.Method,
			"path":        op.Path,
			"description": op.Description,
			"paginated":   op.Paginated,
			"write":       op.Write,
		})
	}
	return f.Printer().Success(map[string]any{
		"version":      meta.Version,
		"generated_at": meta.GeneratedAt,
		"meta_path":    registry.MetaDataPath(),
		"operations":   rows,
		"hint":         fmt.Sprintf("共 %d 个服务、%d 个操作", len(meta.Services), len(rows)),
	})
}
