package capabilitiescmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/spf13/cobra"
)

// Register adds capabilities listing command.
func Register(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "capabilities",
		Short: "列出 registry 中全部可调用接口",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCapabilities()
		},
	})
}

func runCapabilities() error {
	doc, err := registry.DocumentFromCache()
	if err != nil {
		meta, loadErr := registry.Load()
		if loadErr != nil {
			fmt.Fprintln(os.Stderr, loadErr.Error())
			return loadErr
		}
		return printFromMetadata(meta)
	}
	return printFromDocument(doc)
}

func printFromDocument(doc *registry.DocumentV2) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}

	domains := make([]map[string]any, 0)
	domainNames := make([]string, 0, len(doc.Domains))
	for name := range doc.Domains {
		domainNames = append(domainNames, name)
	}
	sort.Strings(domainNames)

	capabilities := make([]map[string]any, 0, len(doc.APIs))
	apiIDs := make([]string, 0, len(doc.APIs))
	for id := range doc.APIs {
		apiIDs = append(apiIDs, id)
	}
	sort.Strings(apiIDs)

	for _, apiID := range apiIDs {
		entry := doc.APIs[apiID]
		capabilities = append(capabilities, map[string]any{
			"apiId":       apiID,
			"domain":      entry.Domain,
			"title":       entry.Title,
			"method":      entry.Method,
			"path":        entry.Path,
			"baseUrl":     entry.BaseURL,
			"contentType": entry.ContentType,
			"write":       entry.Write,
			"pageable":    entry.Pageable,
			"risk":        entry.Risk,
			"stability":   entry.Stability,
		})
	}

	for _, name := range domainNames {
		idx := doc.Domains[name]
		domains = append(domains, map[string]any{
			"domain": name,
			"label":  idx.Label,
			"count":  idx.Count,
			"apiIds": idx.APIIDs,
		})
	}

	return f.Printer().Success(map[string]any{
		"version":       doc.Version,
		"generated_at":  doc.GeneratedAt,
		"registry_path": registry.CachePath(),
		"total":         len(capabilities),
		"domains":       domains,
		"capabilities":  capabilities,
		"hint":          fmt.Sprintf("共 %d 个接口；调用示例: kuaimai-cli web call <apiId> --params '{...}'", len(capabilities)),
	})
}

func printFromMetadata(meta *registry.Metadata) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	capabilities := make([]map[string]any, 0)
	for _, item := range meta.AllOperations() {
		op := item.Operation
		capabilities = append(capabilities, map[string]any{
			"apiId":       item.Service + "." + op.Name,
			"title":       op.Summary,
			"method":      op.Method,
			"path":        op.Path,
			"contentType": op.ContentType,
			"write":       op.Write,
			"pageable":    op.Pageable,
		})
	}
	return f.Printer().Success(map[string]any{
		"version":       meta.Version,
		"generated_at":  meta.GeneratedAt,
		"registry_path": registry.CachePath(),
		"total":         len(capabilities),
		"capabilities":  capabilities,
		"hint":          fmt.Sprintf("共 %d 个接口；调用示例: kuaimai-cli web call <apiId>", len(capabilities)),
	})
}
