package upgrade

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kuaimai-cli/kuaimai-cli/internal/build"
	"github.com/spf13/cobra"
)

func TestShouldSkipUpdateCheck(t *testing.T) {
	root := &cobra.Command{Use: "kuaimai-cli"}
	upgradeCmd := &cobra.Command{Use: "upgrade"}
	doctorCmd := &cobra.Command{Use: "doctor"}
	itemCmd := &cobra.Command{Use: "item"}
	root.AddCommand(upgradeCmd, doctorCmd, itemCmd)

	if !ShouldSkipUpdateCheck(upgradeCmd) {
		t.Fatal("upgrade should skip update check")
	}
	if !ShouldSkipUpdateCheck(doctorCmd) {
		t.Fatal("doctor should skip update check")
	}
	if ShouldSkipUpdateCheck(itemCmd) {
		t.Fatal("item should not skip update check")
	}

	t.Setenv("KUAIMAI_CLI_SKIP_UPDATE_CHECK", "1")
	if !ShouldSkipUpdateCheck(itemCmd) {
		t.Fatal("env skip expected")
	}
	t.Setenv("KUAIMAI_CLI_SKIP_UPDATE_CHECK", "")
}

func TestCheckWithCacheUsesFile(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	cfgDir := filepath.Join(dir, ".kuaimai-cli")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	st := VersionCheckState{
		LastCheckAt: time.Now(),
		CLIVersion:  build.Version,
		LastResult: &CheckResult{
			Current:     build.Version,
			Latest:      build.Version,
			UpdateAvail: false,
		},
	}
	if err := saveVersionCheckState(st); err != nil {
		t.Fatal(err)
	}
	res, err := checkWithCache("")
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.UpdateAvail {
		t.Fatalf("cached result = %+v", res)
	}
}

func TestUpdateCheckCachePolicy(t *testing.T) {
	if defaultCheckInterval != time.Hour {
		t.Fatalf("defaultCheckInterval = %s, want 1h", defaultCheckInterval)
	}
	now := time.Now()
	current := build.Version
	st := VersionCheckState{
		LastCheckAt: now.Add(-2 * time.Hour),
		CLIVersion:  current,
		LastResult: &CheckResult{
			Current:     current,
			Latest:      "v9.9.9",
			UpdateAvail: true,
		},
	}
	if !shouldUseCachedVersionCheck(st, now) {
		t.Fatal("update-available cache should keep prompting until user upgrades")
	}

	st.LastResult.UpdateAvail = false
	if shouldUseCachedVersionCheck(st, now) {
		t.Fatal("no-update cache older than 1h should be refreshed")
	}

	st.LastCheckAt = now
	st.CLIVersion = "0.0.1"
	if shouldUseCachedVersionCheck(st, now) {
		t.Fatal("cache from a different CLI version should not be reused")
	}
}
