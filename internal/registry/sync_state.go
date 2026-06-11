package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/kuaimai-cli/kuaimai-cli/pkg/util"
)

const syncStateFileName = "sync-state.json"

// SyncState tracks local registry sync metadata for conditional remote checks.
type SyncState struct {
	Source     string    `json:"source"`
	Version    string    `json:"version"`
	ETag       string    `json:"etag,omitempty"`
	APICount   int       `json:"api_count"`
	LastSyncAt time.Time `json:"last_sync_at"`
}

func syncStatePath() string {
	return filepath.Join(util.ConfigDir(), registryDirName, syncStateFileName)
}

func loadSyncState() (*SyncState, error) {
	raw, err := os.ReadFile(syncStatePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var st SyncState
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func saveSyncState(st *SyncState) error {
	if st == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(syncStatePath()), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(syncStatePath(), raw, 0o600)
}
