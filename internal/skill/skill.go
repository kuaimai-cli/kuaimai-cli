package skill

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultGitHubRepo = "kuaimai-cli/kuaimai-cli"
	defaultGitRef     = "main"
	httpTimeout       = 30 * time.Second
	maxSkillFileSize  = 1 << 20 // 1 MiB per file
)

// DefaultGitHubRepo is the repository used by skill install.
func DefaultGitHubRepo() string { return defaultGitHubRepo }

// DefaultSkillNames are bundled skills in the GitHub repo skills/ directory.
var DefaultSkillNames = []string{"kuaimai-shared", "kuaimai-erp-item", "kuaimai-scm-item"}

// LegacyDefaultSkillNames are old default skill directories that should be
// removed when refreshing defaults, so agents do not read stale routing rules.
var LegacyDefaultSkillNames = []string{"kuaimai-item", "kuaimai-scm"}

// Entry describes an installed SKILL.md.
type Entry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Source  string `json:"source"`
	Preview string `json:"preview,omitempty"`
}

// InstallResult describes one installed skill.
type InstallResult struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
}

// RootStatus describes whether one agent skill root has all default skills.
type RootStatus struct {
	Root    string          `json:"root"`
	OK      bool            `json:"ok"`
	Skills  map[string]bool `json:"skills"`
	Refs    map[string]bool `json:"references"`
	Missing []string        `json:"missing,omitempty"`
}

// InstallOptions configures skill install (bundled npm/repo skills first, GitHub fallback).
type InstallOptions struct {
	Repo string
	Ref  string
}

type githubContentEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

type skillFile struct {
	relPath string
	body    []byte
}

// AgentSkillRoots returns global skill directories for mainstream AI agents.
func AgentSkillRoots() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法确定用户目录: %w", err)
	}
	rel := []string{
		filepath.Join(".agents", "skills"),
		filepath.Join(".cursor", "skills"),
		filepath.Join(".codex", "skills"),
		filepath.Join(".claude", "skills"),
		filepath.Join(".windsurf", "skills"),
	}
	var roots []string
	for _, r := range rel {
		roots = append(roots, filepath.Join(home, r))
	}
	return roots, nil
}

// List discovers skills installed under agent skill roots (~/.agents/skills, etc.).
func List() ([]Entry, error) {
	roots, err := AgentSkillRoots()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var out []Entry
	for _, root := range roots {
		entries, err := listInRoot(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if _, ok := seen[e.Name]; ok {
				continue
			}
			seen[e.Name] = struct{}{}
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func listInRoot(root string) ([]Entry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		skillPath := filepath.Join(root, ent.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err != nil {
			continue
		}
		preview, _ := readPreview(skillPath, 3)
		out = append(out, Entry{
			Name:    ent.Name(),
			Path:    skillPath,
			Source:  root,
			Preview: preview,
		})
	}
	return out, nil
}

func readPreview(path string, maxLines int) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(raw), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

// IsInstalled reports whether name exists under any agent skill root.
func IsInstalled(name string) (bool, error) {
	roots, err := AgentSkillRoots()
	if err != nil {
		return false, err
	}
	for _, root := range roots {
		if _, err := os.Stat(filepath.Join(root, name, "SKILL.md")); err == nil {
			return true, nil
		}
	}
	return false, nil
}

// HasReferences reports whether name has a references/ directory under any agent skill root.
func HasReferences(name string) (bool, error) {
	roots, err := AgentSkillRoots()
	if err != nil {
		return false, err
	}
	for _, root := range roots {
		refDir := filepath.Join(root, name, "references")
		info, err := os.Stat(refDir)
		if err == nil && info.IsDir() {
			return true, nil
		}
	}
	return false, nil
}

// DefaultRootStatuses reports default skill readiness for every supported agent root.
func DefaultRootStatuses() ([]RootStatus, error) {
	return RootStatuses(DefaultSkillNames)
}

// RootStatuses reports skill and references readiness for every supported agent root.
func RootStatuses(names []string) ([]RootStatus, error) {
	roots, err := AgentSkillRoots()
	if err != nil {
		return nil, err
	}
	var out []RootStatus
	for _, root := range roots {
		st := RootStatus{
			Root:   root,
			OK:     true,
			Skills: make(map[string]bool, len(names)),
			Refs:   make(map[string]bool, len(names)),
		}
		for _, name := range names {
			skillOK := fileExists(filepath.Join(root, name, "SKILL.md"))
			refOK := dirExists(filepath.Join(root, name, "references"))
			st.Skills[name] = skillOK
			st.Refs[name] = refOK
			if !skillOK {
				st.OK = false
				st.Missing = append(st.Missing, name+"/SKILL.md")
			}
			if !refOK {
				st.OK = false
				st.Missing = append(st.Missing, name+"/references")
			}
		}
		out = append(out, st)
	}
	return out, nil
}

// DefaultsInstalledInAllRoots reports whether each supported agent root has every default skill.
func DefaultsInstalledInAllRoots() (bool, error) {
	statuses, err := DefaultRootStatuses()
	if err != nil {
		return false, err
	}
	for _, st := range statuses {
		if !st.OK {
			return false, nil
		}
	}
	return true, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Install installs the full skill directory into all agent skill roots.
// Prefers bundled skills shipped with the npm package or repo; falls back to GitHub.
func Install(name string, opts InstallOptions) (InstallResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return InstallResult{}, fmt.Errorf("skill 名称不能为空")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		return InstallResult{}, fmt.Errorf("skill 名称不合法")
	}
	if root, ok := BundledSkillsRoot(); ok {
		if _, err := os.Stat(filepath.Join(root, name, "SKILL.md")); err == nil {
			return installFromBundled(root, name)
		}
	}
	repo := normalizeRepo(opts.Repo)
	if repo == "" {
		repo = defaultGitHubRepo
	}
	ref := strings.TrimSpace(opts.Ref)
	if ref == "" {
		ref = defaultGitRef
	}
	files, err := fetchSkillTreeFromGitHub(name, repo, ref)
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

// InstallDefaults installs DefaultSkillNames (bundled first, GitHub fallback).
func InstallDefaults(opts InstallOptions) ([]InstallResult, error) {
	if err := RemoveDefaultSkillDirs(); err != nil {
		return nil, err
	}
	var out []InstallResult
	for _, name := range DefaultSkillNames {
		res, err := Install(name, opts)
		if err != nil {
			return out, err
		}
		out = append(out, res)
	}
	return out, nil
}

// RemoveDefaultSkillDirs deletes current and legacy default skill dirs from all
// agent roots before a default refresh. Non-kuaimai user skills are untouched.
func RemoveDefaultSkillDirs() error {
	roots, err := AgentSkillRoots()
	if err != nil {
		return err
	}
	names := append([]string{}, DefaultSkillNames...)
	names = append(names, LegacyDefaultSkillNames...)
	for _, root := range roots {
		for _, name := range names {
			if err := os.RemoveAll(filepath.Join(root, name)); err != nil {
				return fmt.Errorf("删除旧 Skill %s 失败: %w", name, err)
			}
		}
	}
	return nil
}

func fetchSkillTreeFromGitHub(name, repo, ref string) ([]skillFile, error) {
	prefix := "skills/" + name + "/"
	entries, err := listGitHubContents(repo, ref, "skills/"+name)
	if err != nil {
		// Fallback: install SKILL.md only when Contents API is unavailable.
		body, fetchErr := fetchSkillMDFromGitHub(name, repo, ref)
		if fetchErr != nil {
			return nil, fmt.Errorf("列举 skill 目录失败: %w（回退 SKILL.md 也失败: %v）", err, fetchErr)
		}
		fmt.Fprintf(os.Stderr, "[skill] 警告: 无法列举完整 skill 目录，仅安装 SKILL.md: %v\n", err)
		return []skillFile{{relPath: "SKILL.md", body: body}}, nil
	}
	files, err := collectGitHubSkillFiles(repo, ref, prefix, entries)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("skill %q 在仓库中无文件", name)
	}
	hasSKILL := false
	for _, f := range files {
		if f.relPath == "SKILL.md" {
			hasSKILL = true
			break
		}
	}
	if !hasSKILL {
		return nil, fmt.Errorf("skill %q 缺少 SKILL.md", name)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].relPath < files[j].relPath })
	return files, nil
}

func collectGitHubSkillFiles(repo, ref, prefix string, entries []githubContentEntry) ([]skillFile, error) {
	var files []skillFile
	for _, ent := range entries {
		rel, ok := skillRelPath(prefix, ent.Path)
		if !ok {
			continue
		}
		switch ent.Type {
		case "file":
			if ent.DownloadURL == "" {
				return nil, fmt.Errorf("缺少 download_url: %s", ent.Path)
			}
			body, err := fetchURL(ent.DownloadURL)
			if err != nil {
				return nil, fmt.Errorf("下载 %s 失败: %w", ent.Path, err)
			}
			files = append(files, skillFile{relPath: rel, body: body})
		case "dir":
			sub, err := listGitHubContents(repo, ref, ent.Path)
			if err != nil {
				return nil, err
			}
			subFiles, err := collectGitHubSkillFiles(repo, ref, prefix, sub)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		}
	}
	return files, nil
}

func skillRelPath(prefix, fullPath string) (string, bool) {
	if !strings.HasPrefix(fullPath, prefix) {
		return "", false
	}
	rel := strings.TrimPrefix(fullPath, prefix)
	if rel == "" || strings.Contains(rel, "..") {
		return "", false
	}
	return rel, true
}

func listGitHubContents(repo, ref, path string) ([]githubContentEntry, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", repo, path, ref)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("GitHub API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var entries []githubContentEntry
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&entries); err != nil {
		return nil, fmt.Errorf("解析 GitHub API 响应失败: %w", err)
	}
	return entries, nil
}

func fetchSkillMDFromGitHub(name, repo, ref string) ([]byte, error) {
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/skills/%s/SKILL.md", repo, ref, name)
	return fetchURL(rawURL)
}

func normalizeRepo(repo string) string {
	repo = strings.TrimSpace(repo)
	repo = strings.TrimPrefix(repo, "github:")
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimSuffix(repo, "/")
	repo = strings.TrimSuffix(repo, ".git")
	if repo == "" || strings.Count(repo, "/") != 1 {
		return ""
	}
	return repo
}

func fetchURL(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSkillFileSize))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("下载内容为空")
	}
	return body, nil
}

func writeSkillTreeAtRoot(root, name string, files []skillFile) (string, error) {
	skillDir := filepath.Join(root, name)
	if err := os.RemoveAll(skillDir); err != nil {
		return "", err
	}
	var skillMDPath string
	for _, f := range files {
		destPath := filepath.Join(skillDir, filepath.FromSlash(f.relPath))
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(destPath, f.body, 0o644); err != nil {
			return "", err
		}
		if f.relPath == "SKILL.md" {
			skillMDPath = destPath
		}
	}
	if skillMDPath == "" {
		return "", fmt.Errorf("未写入 SKILL.md")
	}
	return skillMDPath, nil
}

// writeSkillBytesAtRoot writes a single SKILL.md (used in tests).
func writeSkillBytesAtRoot(root, name string, body []byte) (string, error) {
	return writeSkillTreeAtRoot(root, name, []skillFile{{relPath: "SKILL.md", body: body}})
}
