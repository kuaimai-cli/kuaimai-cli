package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDefaults(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	created, err := Init()
	if err != nil || !created {
		t.Fatalf("Init: created=%v err=%v", created, err)
	}
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if m.APIURL() != "https://erp1.superboss.cc/" {
		t.Fatalf("api url: %s", m.APIURL())
	}
	if got := m.ShortcutAPIURL("erp-item"); got != "https://erp1.superboss.cc/" {
		t.Fatalf("erp-item api url: %s", got)
	}
	if got := m.ShortcutAPIURL("scm-item"); got != "https://scm3.superboss.cc/" {
		t.Fatalf("scm-item api url: %s", got)
	}
}

func TestInitTemplate(t *testing.T) {
	if !containsAll(DefaultConfigTemplate,
		"api:",
		"url:",
		"shortcuts:",
		"erp-item:",
		"scm-item:",
		"timeout:",
		"retry:",
		"cli:",
		"output:",
		"color:",
	) {
		t.Fatal("template missing required keys")
	}
	if contains(DefaultConfigTemplate, "json_suffix") {
		t.Fatal("template must not contain json_suffix")
	}
}

func TestShortcutAPIURLCanBeOverridden(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Set("shortcuts.scm-item.api_url", "https://scm3.example.test/"); err != nil {
		t.Fatal(err)
	}
	if got := m.ShortcutAPIURL("scm-item"); got != "https://scm3.example.test/" {
		t.Fatalf("shortcut override = %q", got)
	}
}

func TestConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	want := filepath.Join(dir, ".kuaimai-cli", "config.yaml")
	if got := ConfigPath(); got != want {
		t.Fatalf("ConfigPath() = %q, want %q", got, want)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
