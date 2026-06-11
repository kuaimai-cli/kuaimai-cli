package registry

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestShouldSkipAutoSyncRegistrySubcommand(t *testing.T) {
	root := &cobra.Command{Use: "kuaimai-cli"}
	reg := &cobra.Command{Use: "registry"}
	sync := &cobra.Command{Use: "sync"}
	reg.AddCommand(sync)
	root.AddCommand(reg)

	if !ShouldSkipAutoSync(sync) {
		t.Fatal("expected skip for registry sync")
	}
}

func TestShouldSkipAutoSyncSchema(t *testing.T) {
	root := &cobra.Command{Use: "kuaimai-cli"}
	schema := &cobra.Command{Use: "schema"}
	root.AddCommand(schema)

	if ShouldSkipAutoSync(schema) {
		t.Fatal("schema should not skip autosync")
	}
}
