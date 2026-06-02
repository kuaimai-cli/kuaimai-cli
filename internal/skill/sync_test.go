package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNeedsSync_missingInstall(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	// No mock for GitHub — only test local state path when release ref differs
	st := SyncState{CLIVersion: "0.1.0", ReleaseRef: "v0.0.1"}
	if err := os.MkdirAll(filepath.Join(dir, ".kuaimai-cli"), 0o700); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(skillSyncPath())
	_ = raw
	if err := saveSyncState(st); err != nil {
		t.Fatal(err)
	}
	for _, name := range DefaultSkillNames {
		ok, err := IsInstalled(name)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("expected %s not installed", name)
		}
	}
}

func TestSaveAndLoadSyncState(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	st := SyncState{CLIVersion: "0.1.6", ReleaseRef: "v0.1.6"}
	if err := saveSyncState(st); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadSyncState()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ReleaseRef != "v0.1.6" || loaded.CLIVersion != "0.1.6" {
		t.Fatalf("loaded = %+v", loaded)
	}
}
