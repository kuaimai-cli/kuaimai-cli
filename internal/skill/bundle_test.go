package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBundledSkillsRoot_fromEnv(t *testing.T) {
	dir := t.TempDir()
	bundled := filepath.Join(dir, "skills")
	for _, name := range DefaultSkillNames {
		skillDir := filepath.Join(bundled, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	old := os.Getenv("KUAIMAI_CLI_SKILLS_DIR")
	t.Setenv("KUAIMAI_CLI_SKILLS_DIR", bundled)

	got, ok := BundledSkillsRoot()
	if !ok {
		t.Fatal("expected bundled root")
	}
	if got != bundled {
		t.Fatalf("got %q, want %q", got, bundled)
	}
	_ = old
}

func TestInstallFromBundled_writesReferences(t *testing.T) {
	dir := t.TempDir()
	bundled := filepath.Join(dir, "skills")
	itemDir := filepath.Join(bundled, "kuaimai-erp-item", "references")
	if err := os.MkdirAll(itemDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundled, "kuaimai-erp-item", "SKILL.md"), []byte("# item\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(itemDir, "kuaimai-erp-item-list.md"), []byte("# list\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"kuaimai-shared", "kuaimai-scm-item"} {
		skillDir := filepath.Join(bundled, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", filepath.Join(dir, "home"))
	t.Setenv("KUAIMAI_CLI_SKILLS_DIR", bundled)

	res, err := installFromBundled(bundled, "kuaimai-erp-item")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Paths) == 0 {
		t.Fatal("expected install paths")
	}
	ok, err := HasReferences("kuaimai-erp-item")
	if err != nil || !ok {
		t.Fatalf("HasReferences = %v, err = %v", ok, err)
	}
	_ = oldHome
}

func TestInstallDefaultsRemovesLegacyDefaultSkills(t *testing.T) {
	dir := t.TempDir()
	bundled := filepath.Join(dir, "skills")
	for _, name := range DefaultSkillNames {
		skillDir := filepath.Join(bundled, name, "references")
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(bundled, name, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("HOME", filepath.Join(dir, "home"))
	t.Setenv("KUAIMAI_CLI_SKILLS_DIR", bundled)
	roots, err := AgentSkillRoots()
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range roots {
		for _, name := range LegacyDefaultSkillNames {
			legacyDir := filepath.Join(root, name)
			if err := os.MkdirAll(legacyDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(legacyDir, "SKILL.md"), []byte("# legacy\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	if _, err := InstallDefaults(InstallOptions{}); err != nil {
		t.Fatal(err)
	}
	for _, root := range roots {
		for _, name := range LegacyDefaultSkillNames {
			if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
				t.Fatalf("legacy skill still exists: %s", filepath.Join(root, name))
			}
		}
		for _, name := range DefaultSkillNames {
			if _, err := os.Stat(filepath.Join(root, name, "SKILL.md")); err != nil {
				t.Fatalf("default skill missing: %s: %v", name, err)
			}
		}
	}
}

func TestInstall_prefersBundledOverGitHub(t *testing.T) {
	dir := t.TempDir()
	bundled := filepath.Join(dir, "skills")
	for _, name := range DefaultSkillNames {
		skillDir := filepath.Join(bundled, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := []byte("# bundled-" + name + "\n")
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", filepath.Join(dir, "home"))
	t.Setenv("KUAIMAI_CLI_SKILLS_DIR", bundled)

	res, err := Install("kuaimai-shared", InstallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Paths) == 0 {
		t.Fatal("expected paths")
	}
	raw, err := os.ReadFile(res.Paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "# bundled-kuaimai-shared\n" {
		t.Fatalf("unexpected content: %q", string(raw))
	}
	_ = oldHome
}

func TestNeedsSync_usesBundledWithoutGitHub(t *testing.T) {
	dir := t.TempDir()
	bundled := filepath.Join(dir, "skills")
	for _, name := range DefaultSkillNames {
		skillDir := filepath.Join(bundled, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", filepath.Join(dir, "home"))
	t.Setenv("KUAIMAI_CLI_SKILLS_DIR", bundled)

	stale, ref, err := NeedsSync("")
	if err != nil {
		t.Fatal(err)
	}
	if !stale {
		t.Fatal("expected stale when skills not installed to agent dirs")
	}
	if ref == "" {
		t.Fatal("expected sync ref")
	}
	_ = oldHome
}
