package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kuaimai/kuaimai-cli/internal/config"
	"github.com/zalando/go-keyring"
)

const (
	serviceName       = "kuaimai-cli"
	defaultProfile    = "default"
	HeaderAccessToken = "accessToken" // 快麦业务 API 鉴权请求头
)

// Store manages token in the OS keychain for one profile.
type Store struct {
	profile string
}

// NormalizeProfile returns a safe profile name.
func NormalizeProfile(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return defaultProfile
	}
	return name
}

// NewStore opens the system keyring for the active config profile.
func NewStore() (*Store, error) {
	cfg, err := config.New()
	if err != nil {
		return &Store{profile: defaultProfile}, nil
	}
	return &Store{profile: cfg.ActiveProfile()}, nil
}

// NewStoreForProfile uses an explicit profile name.
func NewStoreForProfile(profile string) (*Store, error) {
	return &Store{profile: NormalizeProfile(profile)}, nil
}

// Profile returns the keyring account name for this store.
func (s *Store) Profile() string {
	if s == nil || s.profile == "" {
		return defaultProfile
	}
	return s.profile
}

func (s *Store) accountName() string {
	return s.Profile()
}

// LoginWithToken saves the business accessToken for a profile and marks it active.
func LoginWithToken(profile, token string) error {
	profile = NormalizeProfile(profile)
	store := &Store{profile: profile}
	if err := store.SaveToken(token); err != nil {
		return err
	}
	cfg, err := config.New()
	if err != nil {
		return nil
	}
	if err := cfg.AddProfile(profile); err != nil {
		return err
	}
	if err := cfg.SetActiveProfile(profile); err != nil {
		return err
	}
	return nil
}

// SaveToken stores the access token securely in the system keychain.
func (s *Store) SaveToken(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("Token 不能为空")
	}
	if path := tokenFilePath(); path != "" {
		return saveTokenFile(path, s.accountName(), token)
	}
	if err := keyring.Set(serviceName, s.accountName(), token); err != nil {
		return fmt.Errorf("写入系统密钥链失败: %w", err)
	}
	return nil
}

// GetToken reads the token from keyring.
func (s *Store) GetToken() (string, error) {
	if path := tokenFilePath(); path != "" {
		return getTokenFile(path, s.accountName())
	}
	token, err := keyring.Get(serviceName, s.accountName())
	if err != nil {
		return "", err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("密钥链中 Token 为空")
	}
	return token, nil
}

// DeleteToken removes stored credentials for this profile.
func (s *Store) DeleteToken() error {
	if path := tokenFilePath(); path != "" {
		return deleteTokenFile(path, s.accountName())
	}
	err := keyring.Delete(serviceName, s.accountName())
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	return nil
}

// IsLoggedIn reports whether a token is available for this profile.
func (s *Store) IsLoggedIn() bool {
	_, err := s.GetToken()
	return err == nil
}

// TokenPreview returns a masked token for status display.
func TokenPreview(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// ListProfilesWithLoginStatus returns profile names and whether each has a token.
func ListProfilesWithLoginStatus(cfg *config.Manager) []map[string]any {
	names := cfg.ListProfiles()
	out := make([]map[string]any, 0, len(names))
	active := cfg.ActiveProfile()
	for _, name := range names {
		st, _ := NewStoreForProfile(name)
		loggedIn := st != nil && st.IsLoggedIn()
		entry := map[string]any{
			"name":      name,
			"logged_in": loggedIn,
			"active":    name == active,
			"api_url":   cfg.ProfileAPIURL(name),
		}
		if loggedIn {
			if tok, err := st.GetToken(); err == nil {
				entry["token_preview"] = TokenPreview(tok)
			}
		}
		out = append(out, entry)
	}
	return out
}
