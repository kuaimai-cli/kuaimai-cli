package upgrade

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kuaimai/kuaimai-cli/internal/build"
)

const defaultRepo = "kuaimai-cli/kuaimai-cli"

// ReleaseInfo is a subset of GitHub release JSON.
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
}

// CheckResult compares local version with GitHub latest release.
type CheckResult struct {
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	UpdateAvail bool   `json:"update_available"`
	ReleaseURL  string `json:"release_url,omitempty"`
	Hint        string `json:"hint,omitempty"`
}

// CheckLatest fetches the newest release tag from GitHub.
func CheckLatest(repo string) (*CheckResult, error) {
	if strings.TrimSpace(repo) == "" {
		repo = defaultRepo
	}
	current := strings.TrimPrefix(build.Version, "v")
	rel, err := fetchLatestRelease(repo)
	if err != nil {
		return nil, err
	}
	latest := strings.TrimPrefix(rel.TagName, "v")
	res := &CheckResult{
		Current:     build.Version,
		Latest:      rel.TagName,
		ReleaseURL:  rel.HTMLURL,
		UpdateAvail: versionLess(current, latest),
	}
	if res.UpdateAvail {
		res.Hint = fmt.Sprintf("可执行 npx @kuaimai-cli/cli@latest install 或从 %s 下载新版本", rel.HTMLURL)
	} else {
		res.Hint = "当前已是最新版本（或本地 dev 构建）"
	}
	return res, nil
}

func fetchLatestRelease(repo string) (*ReleaseInfo, error) {
	url := "https://api.github.com/repos/" + repo + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "kuaimai-cli")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("查询 GitHub Release 失败: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var rel ReleaseInfo
	if err := json.Unmarshal(raw, &rel); err != nil {
		return nil, fmt.Errorf("解析 Release 信息失败: %w", err)
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("未找到有效 Release 标签")
	}
	return &rel, nil
}

// versionLess reports whether a < b (semver-ish, numeric segments).
func versionLess(a, b string) bool {
	if a == b {
		return false
	}
	if a == "" || a == "dev" {
		return true
	}
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	for len(ap) < len(bp) {
		ap = append(ap, "0")
	}
	for len(bp) < len(ap) {
		bp = append(bp, "0")
	}
	for i := 0; i < len(ap); i++ {
		if ap[i] == bp[i] {
			continue
		}
		return ap[i] < bp[i]
	}
	return false
}
