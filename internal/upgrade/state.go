package upgrade

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/kuaimai-cli/kuaimai-cli/internal/build"
	"github.com/kuaimai-cli/kuaimai-cli/pkg/util"
)

const versionCheckFile = "version-check.json"

// VersionCheckState caches the last remote version lookup.
type VersionCheckState struct {
	LastCheckAt  time.Time    `json:"last_check_at"`
	CLIVersion   string       `json:"cli_version"`
	LastResult   *CheckResult `json:"last_result,omitempty"`
}

func versionCheckPath() string {
	return filepath.Join(util.ConfigDir(), versionCheckFile)
}

func loadVersionCheckState() (VersionCheckState, error) {
	var st VersionCheckState
	path := versionCheckPath()
	raw, err := os.ReadFile(path)
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

func saveVersionCheckState(st VersionCheckState) error {
	dir := util.ConfigDir()
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	st.CLIVersion = build.Version
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(versionCheckPath(), raw, 0o600)
}
