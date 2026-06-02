package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kuaimai-cli/kuaimai-cli/internal/build"
	"github.com/kuaimai-cli/kuaimai-cli/internal/upgrade"
	"github.com/kuaimai-cli/kuaimai-cli/pkg/util"
	"github.com/spf13/cobra"
)

const skillSyncFile = "skill-sync.json"

// SyncState records the last skill sync relative to CLI / release.
type SyncState struct {
	CLIVersion string    `json:"cli_version"`
	ReleaseRef string    `json:"release_ref"`
	SyncedAt   time.Time `json:"synced_at"`
}

func skillSyncPath() string {
	return filepath.Join(util.ConfigDir(), skillSyncFile)
}

func loadSyncState() (SyncState, error) {
	var st SyncState
	raw, err := os.ReadFile(skillSyncPath())
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return st, err
	}
	if err := json.Unmarshal(raw, &st); err != nil {
		return st, err
	}
	return st, nil
}

func saveSyncState(st SyncState) error {
	dir := util.ConfigDir()
	if dir == "" {
		return fmt.Errorf("无法确定配置目录")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	st.SyncedAt = time.Now()
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(skillSyncPath(), raw, 0o600)
}

// NeedsSync reports whether skills should be refreshed for the latest GitHub release.
func NeedsSync(repo string) (bool, string, error) {
	ref, err := upgrade.LatestReleaseTag(repo)
	if err != nil {
		return false, "", err
	}
	st, err := loadSyncState()
	if err != nil {
		return true, ref, nil
	}
	if st.ReleaseRef != ref {
		return true, ref, nil
	}
	for _, name := range DefaultSkillNames {
		ok, err := IsInstalled(name)
		if err != nil {
			return false, ref, err
		}
		if !ok {
			return true, ref, nil
		}
	}
	if st.CLIVersion != build.Version {
		return true, ref, nil
	}
	return false, ref, nil
}

// InstallIfStale installs default skills when missing or release ref changed.
func InstallIfStale(opts InstallOptions) (bool, []InstallResult, error) {
	repo := normalizeRepo(opts.Repo)
	if repo == "" {
		repo = defaultGitHubRepo
	}
	stale, ref, err := NeedsSync(repo)
	if err != nil {
		return false, nil, err
	}
	if !stale {
		return false, nil, nil
	}
	opts.Ref = ref
	results, err := InstallDefaults(opts)
	if err != nil {
		return false, results, err
	}
	_ = saveSyncState(SyncState{CLIVersion: build.Version, ReleaseRef: ref})
	return true, results, nil
}

// SyncDefaultsAfterUpgrade installs skills for the given release ref (tag name).
func SyncDefaultsAfterUpgrade(releaseRef string, opts InstallOptions) error {
	releaseRef = strings.TrimSpace(releaseRef)
	if releaseRef == "" {
		var err error
		releaseRef, err = upgrade.LatestReleaseTag(opts.Repo)
		if err != nil {
			return err
		}
	}
	opts.Ref = releaseRef
	if _, err := InstallDefaults(opts); err != nil {
		return err
	}
	return saveSyncState(SyncState{CLIVersion: build.Version, ReleaseRef: releaseRef})
}

// ShouldSkipSkillAutoSync skips background skill sync on selected commands.
func ShouldSkipSkillAutoSync(cmd *cobra.Command) bool {
	if strings.TrimSpace(os.Getenv("KUAIMAI_CLI_SKIP_SKILL_SYNC")) != "" {
		return true
	}
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "upgrade", "skill", "completion", "help":
			return true
		}
	}
	return false
}

// MaybeSyncOnCLIVersionChange refreshes skills once when the running CLI version changes.
func MaybeSyncOnCLIVersionChange(cmd *cobra.Command) {
	if ShouldSkipSkillAutoSync(cmd) {
		return
	}
	if build.Version == "" || build.Version == "dev" {
		return
	}
	st, err := loadSyncState()
	if err != nil || st.CLIVersion == build.Version {
		return
	}
	go func() {
		opts := InstallOptions{Repo: defaultGitHubRepo}
		ref, err := upgrade.LatestReleaseTag(opts.Repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[kuaimai-cli] Skill 自动同步跳过（查询 Release 失败）: %v\n", err)
			return
		}
		if err := SyncDefaultsAfterUpgrade(ref, opts); err != nil {
			fmt.Fprintf(os.Stderr, "[kuaimai-cli] Skill 自动同步失败: %v（可运行: kuaimai-cli skill install --if-stale）\n", err)
			return
		}
		fmt.Fprintf(os.Stderr, "[kuaimai-cli] CLI 版本已变更，Skills 已同步至 %s\n", ref)
	}()
}
