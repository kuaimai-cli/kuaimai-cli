package skill

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAdd_requiresName(t *testing.T) {
	_, err := Add("", "x.md")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestList_skipsMissingRoots(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })
	old, _ := os.Getwd()
	_ = os.Chdir(dir)
	t.Cleanup(func() { _ = os.Chdir(old) })
	_ = os.MkdirAll(filepath.Join(dir, "skills", "demo"), 0o700)
	_ = os.WriteFile(filepath.Join(dir, "skills", "demo", "SKILL.md"), []byte("# demo\n"), 0o600)
	entries, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "demo" {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestAddFromURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("# test-skill\n"))
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	dest, err := AddFromURL("test-skill", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, ".agents", "skills", "test-skill", "SKILL.md")
	if dest != want {
		t.Fatalf("dest = %q, want %q", dest, want)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatal(err)
	}
}

func TestInstall_fromDir(t *testing.T) {
	dir := t.TempDir()
	skillsRoot := filepath.Join(dir, "skills-src")
	_ = os.MkdirAll(filepath.Join(skillsRoot, "b"), 0o700)
	_ = os.WriteFile(filepath.Join(skillsRoot, "b", "SKILL.md"), []byte("# b\n"), 0o600)

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	dest, err := Install("b", InstallOptions{FromDir: skillsRoot})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatal(err)
	}
}

func TestInstall_requiresName(t *testing.T) {
	_, err := Install("", InstallOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstallAllFromDir(t *testing.T) {
	dir := t.TempDir()
	skillsRoot := filepath.Join(dir, "skills-src")
	_ = os.MkdirAll(filepath.Join(skillsRoot, "a"), 0o700)
	_ = os.WriteFile(filepath.Join(skillsRoot, "a", "SKILL.md"), []byte("# a\n"), 0o600)

	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	results, err := InstallAllFromDir(skillsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "a" {
		t.Fatalf("results = %+v", results)
	}
}

func TestNormalizeRepo(t *testing.T) {
	cases := map[string]string{
		"kuaimai/kuaimai-cli":                    "kuaimai/kuaimai-cli",
		"github:kuaimai/kuaimai-cli":             "kuaimai/kuaimai-cli",
		"https://github.com/kuaimai/kuaimai-cli": "kuaimai/kuaimai-cli",
		"invalid":                                "",
	}
	for in, want := range cases {
		if got := normalizeRepo(in); got != want {
			t.Fatalf("normalizeRepo(%q) = %q, want %q", in, got, want)
		}
	}
}
