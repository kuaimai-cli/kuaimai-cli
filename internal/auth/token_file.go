package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// tokenFilePath returns the optional file store path (tests / headless CI).
func tokenFilePath() string {
	return strings.TrimSpace(os.Getenv("KUAIMAI_CLI_TOKEN_FILE"))
}

type tokenFileData map[string]string

func loadTokenFile(path string) (tokenFileData, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return tokenFileData{}, nil
		}
		return nil, err
	}
	var data tokenFileData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if data == nil {
		data = tokenFileData{}
	}
	return data, nil
}

func saveTokenFile(path, profile, token string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := loadTokenFile(path)
	if err != nil {
		return err
	}
	data[profile] = token
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func getTokenFile(path, profile string) (string, error) {
	data, err := loadTokenFile(path)
	if err != nil {
		return "", err
	}
	tok := strings.TrimSpace(data[profile])
	if tok == "" {
		return "", fmt.Errorf("profile %q 无 token", profile)
	}
	return tok, nil
}

func deleteTokenFile(path, profile string) error {
	data, err := loadTokenFile(path)
	if err != nil {
		return err
	}
	delete(data, profile)
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func useFileStore() bool {
	return tokenFilePath() != ""
}
