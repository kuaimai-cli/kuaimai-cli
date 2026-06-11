package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kuaimai-cli/kuaimai-cli/internal/config"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/kuaimai-cli/kuaimai-cli/pkg/logger"
)

// bootstrapRegistry syncs remote registry before command dispatch (lark-cli style).
func bootstrapRegistry(args []string) error {
	if shouldSkipRegistryBootstrap(args) {
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

func shouldSkipRegistryBootstrap(args []string) bool {
	if strings.TrimSpace(os.Getenv("KUAIMAI_CLI_SKIP_REGISTRY_SYNC")) != "" {
		return true
	}
	skip := map[string]struct{}{
		"registry":   {},
		"upgrade":    {},
		"completion": {},
		"help":       {},
		"-h":         {},
		"--help":     {},
		"version":    {},
		"config":     {},
		"auth":       {},
	}
	for _, a := range args[1:] {
		if strings.HasPrefix(a, "-") {
			if a == "--version" {
				return true
			}
			continue
		}
		if _, ok := skip[a]; ok {
			return true
		}
	}
	return false
}
