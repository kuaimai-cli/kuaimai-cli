package skill

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestInstall_requiresName(t *testing.T) {
	_, err := Install("", InstallOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestList_skipsMissingRoots(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	agentsRoot := filepath.Join(dir, ".agents", "skills", "demo")
	_ = os.MkdirAll(agentsRoot, 0o700)
	_ = os.WriteFile(filepath.Join(agentsRoot, "SKILL.md"), []byte("# demo\n"), 0o600)

	entries, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "demo" {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestInstall_writesMultipleRoots(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	body := []byte("# test-skill\n")
	roots, err := AgentSkillRoots()
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range roots {
		if _, err := writeSkillBytesAtRoot(root, "test-skill", body); err != nil {
			t.Fatal(err)
		}
	}
	ok, err := IsInstalled("test-skill")
	if err != nil || !ok {
		t.Fatalf("IsInstalled = %v, err = %v", ok, err)
	}
}

func TestWriteSkillTreeAtRoot(t *testing.T) {
	dir := t.TempDir()
	files := []skillFile{
		{relPath: "SKILL.md", body: []byte("# item\n")},
		{relPath: "references/kuaimai-item-list.md", body: []byte("# list\n")},
	}
	skillMD, err := writeSkillTreeAtRoot(dir, "kuaimai-item", files)
	if err != nil {
		t.Fatal(err)
	}
	if skillMD != filepath.Join(dir, "kuaimai-item", "SKILL.md") {
		t.Fatalf("skillMD = %q", skillMD)
	}
	refPath := filepath.Join(dir, "kuaimai-item", "references", "kuaimai-item-list.md")
	if _, err := os.Stat(refPath); err != nil {
		t.Fatalf("reference missing: %v", err)
	}
}

func TestHasReferences(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	refDir := filepath.Join(dir, ".agents", "skills", "kuaimai-item", "references")
	_ = os.MkdirAll(refDir, 0o700)
	ok, err := HasReferences("kuaimai-item")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected references to exist")
	}
}

func TestCollectGitHubSkillFiles(t *testing.T) {
	const skillMD = "# test-skill\n"
	const refMD = "# reference\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/raw/SKILL.md":
			_, _ = w.Write([]byte(skillMD))
		case "/raw/demo.md":
			_, _ = w.Write([]byte(refMD))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	entries := []githubContentEntry{
		{Name: "SKILL.md", Path: "skills/test-skill/SKILL.md", Type: "file", DownloadURL: srv.URL + "/raw/SKILL.md"},
		{Name: "demo.md", Path: "skills/test-skill/references/demo.md", Type: "file", DownloadURL: srv.URL + "/raw/demo.md"},
	}
	files, err := collectGitHubSkillFiles("acme/repo", "main", "skills/test-skill/", entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("files = %d, want 2", len(files))
	}

	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	installFiles := []skillFile{
		{relPath: "SKILL.md", body: []byte(skillMD)},
		{relPath: "references/demo.md", body: []byte(refMD)},
	}
	roots, err := AgentSkillRoots()
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range roots {
		if _, err := writeSkillTreeAtRoot(root, "test-skill", installFiles); err != nil {
			t.Fatal(err)
		}
	}
	hasRef, err := HasReferences("test-skill")
	if err != nil || !hasRef {
		t.Fatalf("HasReferences = %v, err = %v", hasRef, err)
	}
}

func TestSkillRelPath(t *testing.T) {
	rel, ok := skillRelPath("skills/foo/", "skills/foo/SKILL.md")
	if !ok || rel != "SKILL.md" {
		t.Fatalf("rel=%q ok=%v", rel, ok)
	}
	_, ok = skillRelPath("skills/foo/", "skills/foo/../bar")
	if ok {
		t.Fatal("expected false for path traversal")
	}
}

func TestNormalizeRepo(t *testing.T) {
	cases := map[string]string{
		"kuaimai-cli/kuaimai-cli":                    "kuaimai-cli/kuaimai-cli",
		"github:kuaimai-cli/kuaimai-cli":             "kuaimai-cli/kuaimai-cli",
		"https://github.com/kuaimai-cli/kuaimai-cli": "kuaimai-cli/kuaimai-cli",
		"invalid": "",
	}
	for in, want := range cases {
		if got := normalizeRepo(in); got != want {
			t.Fatalf("normalizeRepo(%q) = %q, want %q", in, got, want)
		}
	}
}
