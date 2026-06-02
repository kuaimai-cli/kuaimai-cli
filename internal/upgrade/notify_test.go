package upgrade

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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
		LastResult: &CheckResult{
			Current:     "v0.1.0",
			Latest:      "v0.1.0",
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
