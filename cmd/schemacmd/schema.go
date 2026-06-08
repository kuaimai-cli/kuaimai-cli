package schemacmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/spf13/cobra"
)

// Register adds schema introspection command.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "API 元数据自省（meta_data.json）",
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
		svc := meta.Services[item.Service]
		row := map[string]any{
			"service":      item.Service,
			"operation":    op.Name,
			"summary":      op.Summary,
			"method":       op.Method,
			"path":         op.Path,
			"contentType":  op.ContentType,
			"write":        op.Write,
			"pageable":     op.Pageable,
			"shortcut":     shortcutForOperation(item.Service, op.Name),
		}
		if base := strings.TrimSpace(svc.BaseURL); base != "" {
			row["baseUrl"] = base
		}
		if op.RequestSchema != nil {
			row["requestSchema"] = op.RequestSchema
		}
		if op.ResponseSchema != nil {
			row["responseSchema"] = op.ResponseSchema
		}
		rows = append(rows, row)
	}
	return f.Printer().Success(map[string]any{
		"version":      meta.Version,
		"generated_at": meta.GeneratedAt,
		"meta_path":    registry.MetaDataPath(),
		"operations":   rows,
		"hint":         fmt.Sprintf("共 %d 个服务、%d 个操作；Agent 优先用 item shortcuts，service 为原子兜底", len(meta.ServiceNames()), len(rows)),
	})
}

// shortcutForOperation maps meta operation id to shortcut 子命令（若有）.
func shortcutForOperation(service, opName string) string {
	if service == "scm" {
		return "service scm " + opName
	}
	switch opName {
	case "stock-list":
		return "item +list / item list"
	case "stock-count":
		return "item count"
	case "item-detail":
		return "item get-detail"
	case "item-save":
		return "item save"
	case "item-update-title":
		return "item update-title"
	default:
		return ""
	}
}
