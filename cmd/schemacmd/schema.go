package schemacmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/spf13/cobra"
)

// Register adds schema introspection command.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "schema [apiId]",
		Short: "API 元数据自省（registry.json）",
		Long: `查看远端 registry.json 同步到本地的接口定义（对标飞书 lark-cli schema）。

  kuaimai-cli schema                    # 列出全部 apiId
  kuaimai-cli schema api.luotao.test.get  # 单接口 requestSchema / responseSchema`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runSchemaOne(args[0])
			}
			return runSchemaAll()
		},
	}
	root.AddCommand(cmd)
}

func runSchemaAll() error {
	doc, docErr := registry.DocumentFromCache()
	if docErr == nil && doc != nil && len(doc.APIs) > 0 {
		return printSchemaFromDocument(doc)
	}

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
		apiID := item.Service + "." + item.Operation.Name
		rows = append(rows, apiRow(meta, apiID, item.Service, item.Operation))
	}
	return f.Printer().Success(map[string]any{
		"version":       meta.Version,
		"generated_at":  meta.GeneratedAt,
		"registry_path": registry.CachePath(),
		"apis":          rows,
		"total":         len(rows),
		"hint":          fmt.Sprintf("共 %d 个接口；单接口: kuaimai-cli schema <apiId>；调用: kuaimai-cli web call <apiId>", len(rows)),
	})
}

func printSchemaFromDocument(doc *registry.DocumentV2) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	apiIDs := make([]string, 0, len(doc.APIs))
	for id := range doc.APIs {
		apiIDs = append(apiIDs, id)
	}
	sort.Strings(apiIDs)

	rows := make([]map[string]any, 0, len(apiIDs))
	for _, apiID := range apiIDs {
		entry := doc.APIs[apiID]
		row := map[string]any{
			"apiId":       apiID,
			"title":       entry.Title,
			"domain":      entry.Domain,
			"method":      entry.Method,
			"path":        entry.Path,
			"contentType": entry.ContentType,
			"write":       entry.Write,
			"pageable":    entry.Pageable,
			"transport":   entry.Transport,
			"stability":   entry.Stability,
			"shortcut":    shortcutForAPIID(apiID),
		}
		if base := strings.TrimSpace(entry.BaseURL); base != "" {
			row["baseUrl"] = base
		}
		rows = append(rows, row)
	}
	return f.Printer().Success(map[string]any{
		"schema_version": doc.SchemaVersion,
		"version":        doc.Version,
		"generated_at":   doc.GeneratedAt,
		"registry_path":  registry.CachePath(),
		"apis":           rows,
		"total":          len(rows),
		"hint":           fmt.Sprintf("共 %d 个接口；调用: kuaimai-cli web call <apiId>", len(rows)),
	})
}

func runSchemaOne(apiID string) error {
	meta, err := registry.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	resolved, err := meta.FindByAPIID(apiID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	row := apiRow(meta, apiID, resolved.Service, resolved.Op)
	if doc, docErr := registry.DocumentFromCache(); docErr == nil {
		if entry, ok := doc.APIs[apiID]; ok {
			row["title"] = entry.Title
			row["domain"] = entry.Domain
			row["description"] = entry.Description
			row["transport"] = entry.Transport
			row["stability"] = entry.Stability
			row["risk"] = entry.Risk
			if len(entry.Examples) > 0 {
				row["examples"] = entry.Examples
			}
		}
	}
	return f.Printer().Success(map[string]any{
		"version":       meta.Version,
		"generated_at":  meta.GeneratedAt,
		"registry_path": registry.CachePath(),
		"api":           row,
		"hint":          fmt.Sprintf("调用: kuaimai-cli web call %s", apiID),
	})
}

func apiRow(meta *registry.Metadata, apiID, service string, op registry.Operation) map[string]any {
	svc := meta.Services[service]
	row := map[string]any{
		"apiId":       apiID,
		"summary":     op.Summary,
		"method":      op.Method,
		"path":        op.Path,
		"contentType": op.ContentType,
		"write":       op.Write,
		"pageable":    op.Pageable,
		"shortcut":    shortcutForAPIID(apiID),
	}
	if base := strings.TrimSpace(svc.ResolveOperationBaseURL(op, "")); base != "" {
		row["baseUrl"] = base
	}
	if op.RequestSchema != nil {
		row["requestSchema"] = op.RequestSchema
	}
	if op.ResponseSchema != nil {
		row["responseSchema"] = op.ResponseSchema
	}
	return row
}

func shortcutForAPIID(apiID string) string {
	_, opName, _ := strings.Cut(apiID, ".")
	switch opName {
	case "stock-list":
		return "erp-item +list / erp-item list"
	case "stock-count":
		return "erp-item count"
	case "item-detail":
		return "erp-item get-detail"
	case "item-save":
		return "erp-item save"
	case "item-update-title":
		return "erp-item update-title"
	default:
		return "web call " + apiID
	}
}
