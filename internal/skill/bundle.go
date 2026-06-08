package skill

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const bundledSkillsEnv = "KUAIMAI_CLI_SKILLS_DIR"

// BundledSkillsRoot returns the directory containing bundled kuaimai-* skill folders
// (npm package skills/ or repo skills/ during development).
func BundledSkillsRoot() (string, bool) {
	if dir := strings.TrimSpace(os.Getenv(bundledSkillsEnv)); dir != "" {
		if validBundledSkillsRoot(dir) {
			return filepath.Clean(dir), true
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", false
	}
	exeDir := filepath.Dir(exe)
	for _, candidate := range []string{
		filepath.Join(exeDir, "..", "skills"), // npm: @kuaimai-cli/cli/bin/../skills
		filepath.Join(exeDir, "skills"),       // repo root binary next to skills/
	} {
		if validBundledSkillsRoot(candidate) {
			return filepath.Clean(candidate), true
		}
	}
	if root := findBundledSkillsRootWalk(exeDir); root != "" {
		return root, true
	}
	if cwd, err := os.Getwd(); err == nil {
		if root := findBundledSkillsRootWalk(cwd); root != "" {
			return root, true
		}
	}
	return "", false
}

func findBundledSkillsRootWalk(start string) string {
	dir := filepath.Clean(start)
	for i := 0; i < 10; i++ {
		candidate := filepath.Join(dir, "skills")
		if validBundledSkillsRoot(candidate) {
			return filepath.Clean(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func validBundledSkillsRoot(root string) bool {
	root = filepath.Clean(root)
	for _, name := range DefaultSkillNames {
		if _, err := os.Stat(filepath.Join(root, name, "SKILL.md")); err != nil {
			return false
		}
	}
	return true
}

func installFromBundled(bundledRoot, name string) (InstallResult, error) {
	files, err := readSkillTreeFromDir(bundledRoot, name)
	if err != nil {
		return InstallResult{}, err
	}
	roots, err := AgentSkillRoots()
	if err != nil {
		return InstallResult{}, err
	}
	var paths []string
	for _, root := range roots {
		dest, err := writeSkillTreeAtRoot(root, name, files)
		if err != nil {
			return InstallResult{Name: name, Paths: paths}, fmt.Errorf("安装到 %s 失败: %w", root, err)
		}
		paths = append(paths, dest)
	}
	return InstallResult{Name: name, Paths: paths}, nil
}

func readSkillTreeFromDir(bundledRoot, name string) ([]skillFile, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		return nil, fmt.Errorf("skill 名称不合法")
	}
	skillDir := filepath.Join(bundledRoot, name)
	info, err := os.Stat(skillDir)
	if err != nil {
		return nil, fmt.Errorf("读取 bundled skill %q 失败: %w", name, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("bundled skill %q 不是目录", name)
	}

	var files []skillFile
	err = filepath.WalkDir(skillDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(skillDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "" || strings.Contains(rel, "..") {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		if fi.Size() > maxSkillFileSize {
			return fmt.Errorf("文件过大: %s", rel)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取 %s 失败: %w", rel, err)
		}
		files = append(files, skillFile{relPath: rel, body: body})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("bundled skill %q 无文件", name)
	}
	hasSKILL := false
	for _, f := range files {
		if f.relPath == "SKILL.md" {
			hasSKILL = true
			break
		}
	}
	if !hasSKILL {
		return nil, fmt.Errorf("bundled skill %q 缺少 SKILL.md", name)
	}
	return files, nil
}
