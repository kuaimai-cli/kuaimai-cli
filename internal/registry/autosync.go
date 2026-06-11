package registry

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// ShouldSkipAutoSync reports whether registry auto-sync should be skipped for this command.
func ShouldSkipAutoSync(cmd *cobra.Command) bool {
	if strings.TrimSpace(os.Getenv("KUAIMAI_CLI_SKIP_REGISTRY_SYNC")) != "" {
		return true
	}
	if cmd != nil && cmd.Root() != nil && cmd.Root().Flags().Changed("version") {
		return true
	}
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "registry", "upgrade", "completion", "help", "version":
			return true
		}
	}
	return false
}
