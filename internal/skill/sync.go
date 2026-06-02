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

const (
	skillSyncFile          = "skill-sync.json"
	defaultSkillCheckInterval = 24 * time.Hour
)

// SyncState records the last skill sync relative to CLI / release.
type SyncState struct {
	CLIVersion           string    `json:"cli_version"`
	ReleaseRef           string    `json:"release_ref"`
	SyncedAt             time.Time `json:"synced_at"`
	LastReleaseCheckAt   time.Time `json:"last_release_check_at,omitempty"`
	CachedLatestRelease  string    `json:"cached_latest_release,omitempty"`
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

func writeSyncState(st SyncState, touchSyncedAt bool) error {
	dir := util.ConfigDir()
	if dir == "" {
		return fmt.Errorf("无法确定配置目录")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if touchSyncedAt {
		st.SyncedAt = time.Now()
	}
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(skillSyncPath(), raw, 0o600)
}

func saveSyncState(st SyncState) error {
	return writeSyncState(st, true)
}

func saveReleaseCheckCache(latest string) error {
	st, err := loadSyncState()
	if err != nil {
		st = SyncState{}
	}
	st.LastReleaseCheckAt = time.Now()
	st.CachedLatestRelease = latest
	return writeSyncState(st, false)
}

func latestReleaseRefCached(repo string) (string, error) {
	st, err := loadSyncState()
	if err != nil {
		st = SyncState{}
	}
	if st.CachedLatestRelease != "" && time.Since(st.LastReleaseCheckAt) < defaultSkillCheckInterval {
		return st.CachedLatestRelease, nil
	}
	ref, err := upgrade.LatestReleaseTag(repo)
	if err != nil {
		return "", err
	}
	if err := saveReleaseCheckCache(ref); err != nil {
		return ref, nil
	}
	return ref, nil
}

func defaultSkillsMissing() (bool, error) {
	for _, name := range DefaultSkillNames {
		ok, err := IsInstalled(name)
		if err != nil {
			return false, err
		}
		if !ok {
			return true, nil
		}
	}
	return false, nil
}

// NeedsSync reports whether skills should be refreshed for the latest GitHub release.
func NeedsSync(repo string) (bool, string, error) {
	repo = normalizeRepo(repo)
	if repo == "" {
		repo = defaultGitHubRepo
	}

	missing, err := defaultSkillsMissing()
	if err != nil {
		return false, "", err
	}
	if missing {
		ref, err := latestReleaseRefCached(repo)
		return true, ref, err
	}

	st, err := loadSyncState()
	if err != nil {
		ref, err2 := latestReleaseRefCached(repo)
		return true, ref, err2
	}
	if st.CLIVersion != build.Version {
		ref, err := latestReleaseRefCached(repo)
		return true, ref, err
	}

	ref, err := latestReleaseRefCached(repo)
	if err != nil {
		return false, "", err
	}
	if st.ReleaseRef != ref {
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
	st, _ := loadSyncState()
	st.CLIVersion = build.Version
	st.ReleaseRef = ref
	st.LastReleaseCheckAt = time.Now()
	st.CachedLatestRelease = ref
	_ = saveSyncState(st)
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
	st, _ := loadSyncState()
	st.CLIVersion = build.Version
	st.ReleaseRef = releaseRef
	st.LastReleaseCheckAt = time.Now()
	st.CachedLatestRelease = releaseRef
	return saveSyncState(st)
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

// MaybeAutoSync refreshes default skills in the background when missing, CLI
// version changed, or GitHub release advanced (24h release check cache).
func MaybeAutoSync(cmd *cobra.Command) {
	if ShouldSkipSkillAutoSync(cmd) {
		return
	}
	if build.Version == "" || build.Version == "dev" {
		return
	}
	go func() {
		updated, _, err := InstallIfStale(InstallOptions{Repo: defaultGitHubRepo})
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"[kuaimai-cli] Skill 自动同步失败: %v（可运行: kuaimai-cli skill install）\n",
				err,
			)
			return
		}
		if updated {
			fmt.Fprintf(os.Stderr, "[kuaimai-cli] Skills 已自动同步至最新 Release\n")
		}
	}()
}

// MaybeSyncOnCLIVersionChange is deprecated; use MaybeAutoSync.
func MaybeSyncOnCLIVersionChange(cmd *cobra.Command) {
	MaybeAutoSync(cmd)
}
