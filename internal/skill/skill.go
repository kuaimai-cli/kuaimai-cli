package skill

import (
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
	defaultGitHubRepo = "kuaimai/kuaimai-cli"
	defaultGitRef     = "main"
	httpTimeout       = 30 * time.Second
)

// DefaultGitHubRepo is the repository used by skill install-all.
func DefaultGitHubRepo() string { return defaultGitHubRepo }

// DefaultSkillNames are bundled skills shipped in the repository skills/ directory.
var DefaultSkillNames = []string{"kuaimai-shared", "kuaimai-item"}

// Entry describes an installed SKILL.md.
type Entry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Source  string `json:"source"`
	Preview string `json:"preview,omitempty"`
}

// AgentSkillsDir returns ~/.agents/skills (cross-agent skill convention).
func AgentSkillsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法确定用户目录: %w", err)
	}
	return filepath.Join(home, ".agents", "skills"), nil
}

// List discovers skills under ./skills and ~/.agents/skills.
func List() ([]Entry, error) {
	roots := searchRoots()
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

func searchRoots() []string {
	var roots []string
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, filepath.Join(cwd, "skills"))
	}
	if dir, err := AgentSkillsDir(); err == nil {
		roots = append(roots, dir)
	}
	return roots
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

// Add installs a SKILL.md into ~/.agents/skills/<name>/SKILL.md from a local file path.
func Add(name, fromPath string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("skill 名称不能为空")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		return "", fmt.Errorf("skill 名称不合法")
	}
	fromPath = strings.TrimSpace(fromPath)
	if fromPath == "" {
		return "", fmt.Errorf("请指定 SKILL.md 源文件路径")
	}
	src, err := os.Open(fromPath)
	if err != nil {
		return "", fmt.Errorf("读取源文件失败: %w", err)
	}
	defer src.Close()
	return writeSkill(name, src, fromPath)
}

// AddFromURL downloads SKILL.md from url and installs it.
func AddFromURL(name, rawURL string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("skill 名称不能为空")
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("请指定 URL")
	}
	body, err := fetchURL(rawURL)
	if err != nil {
		return "", err
	}
	return writeSkillBytes(name, body, rawURL)
}

// AddFromGitHub installs SKILL.md from raw.githubusercontent.com/{repo}/{ref}/skills/{name}/SKILL.md.
func AddFromGitHub(name, repo, ref string) (string, error) {
	repo = normalizeRepo(repo)
	if repo == "" {
		return "", fmt.Errorf("GitHub 仓库不合法")
	}
	if ref == "" {
		ref = defaultGitRef
	}
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/skills/%s/SKILL.md", repo, ref, name)
	return AddFromURL(name, rawURL)
}

// InstallResult describes one installed skill.
type InstallResult struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// InstallAllFromDir copies every skills/*/SKILL.md from dir into ~/.agents/skills/.
func InstallAllFromDir(dir string) ([]InstallResult, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("请指定 skills 目录")
	}
	names, err := discoverSkillNames(dir)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("目录 %s 下未找到 Skill", dir)
	}
	var out []InstallResult
	for _, name := range names {
		srcPath := filepath.Join(dir, name, "SKILL.md")
		dest, err := Add(name, srcPath)
		if err != nil {
			return out, fmt.Errorf("安装 %s 失败: %w", name, err)
		}
		out = append(out, InstallResult{Name: name, Path: dest})
	}
	return out, nil
}

// InstallOptions configures a single-skill install.
type InstallOptions struct {
	FromDir string
	Repo    string
	Ref     string
}

// Install installs one skill into ~/.agents/skills/<name>/ from a local skills dir or GitHub.
func Install(name string, opts InstallOptions) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("skill 名称不能为空")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		return "", fmt.Errorf("skill 名称不合法")
	}
	if dir := strings.TrimSpace(opts.FromDir); dir != "" {
		srcPath := filepath.Join(dir, name, "SKILL.md")
		if _, err := os.Stat(srcPath); err != nil {
			return "", fmt.Errorf("未找到 %s: %w", srcPath, err)
		}
		return Add(name, srcPath)
	}
	repo := normalizeRepo(opts.Repo)
	if repo == "" {
		repo = defaultGitHubRepo
	}
	ref := strings.TrimSpace(opts.Ref)
	if ref == "" {
		ref = defaultGitRef
	}
	return AddFromGitHub(name, repo, ref)
}

// InstallAllFromGitHub installs default skills from a GitHub repository.
func InstallAllFromGitHub(repo, ref string) ([]InstallResult, error) {
	repo = normalizeRepo(repo)
	if repo == "" {
		repo = defaultGitHubRepo
	}
	var out []InstallResult
	for _, name := range DefaultSkillNames {
		dest, err := AddFromGitHub(name, repo, ref)
		if err != nil {
			return out, fmt.Errorf("安装 %s 失败: %w", name, err)
		}
		out = append(out, InstallResult{Name: name, Path: dest})
	}
	return out, nil
}

func discoverSkillNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}
	var names []string
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, ent.Name(), "SKILL.md")); err == nil {
			names = append(names, ent.Name())
		}
	}
	sort.Strings(names)
	return names, nil
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
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("下载内容为空")
	}
	return body, nil
}

func writeSkill(name string, src io.Reader, source string) (string, error) {
	dir, err := AgentSkillsDir()
	if err != nil {
		return "", err
	}
	destDir := filepath.Join(dir, name)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	destPath := filepath.Join(destDir, "SKILL.md")
	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer dest.Close()
	if _, err := io.Copy(dest, src); err != nil {
		return "", err
	}
	_ = source
	return destPath, nil
}

func writeSkillBytes(name string, body []byte, source string) (string, error) {
	return writeSkill(name, strings.NewReader(string(body)), source)
}
