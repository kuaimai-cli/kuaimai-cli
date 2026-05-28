package util

import (
	"os"
	"path/filepath"
)

// HomeDir returns the user home directory or empty on failure.
func HomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return dir
}

// ConfigDir returns ~/.kuaimai-cli
func ConfigDir() string {
	home := HomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".kuaimai-cli")
}
