package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kuaimai-cli/kuaimai-cli/internal/config"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/kuaimai-cli/kuaimai-cli/pkg/logger"
	"github.com/spf13/cobra"
)

// bootstrapRegistry syncs remote registry before registry-backed commands run.
func bootstrapRegistry(cmd *cobra.Command) error {
	if registry.ShouldSkipAutoSync(cmd) {
		return nil
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}
	if !cfg.RegistryAutoSyncEnabled() {
		return nil
	}

	return syncRegistry(cfg.RegistrySource())
}

func syncRegistry(source string) error {
	result, err := registry.SyncIfNeeded(context.Background(), source)
	if err != nil {
		if _, cacheErr := registry.DocumentFromCache(); cacheErr == nil {
			fmt.Fprintf(os.Stderr,
				"[kuaimai-cli] registry 远端同步失败，使用本地缓存: %v\n",
				err,
			)
			return nil
		}
		return err
	}

	if result.Changed && !result.NotModified {
		fmt.Fprintf(os.Stderr,
			"[kuaimai-cli] registry 已更新: version=%s apis=%d\n",
			result.Version, result.APICount,
		)
	}
	if logger.Verbose() {
		logger.Debug("registry bootstrap: source=%s version=%s changed=%v",
			result.Source, result.Version, result.Changed)
	}
	return nil
}
