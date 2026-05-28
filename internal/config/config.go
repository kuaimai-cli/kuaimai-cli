package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kuaimai/kuaimai-cli/pkg/util"
	"github.com/spf13/viper"
)

const (
	fileName         = "config.yaml"
	keyAPIURL     = "api.url"
	keyAPITimeout = "api.timeout"
	keyAPIRetry              = "api.retry"
	keyAPIPoolMaxIdle        = "api.pool_max_idle"
	keyAPIPoolMaxIdlePerHost = "api.pool_max_idle_per_host"
	keyAPICircuitThreshold   = "api.circuit_threshold"
	keyAPICircuitCooldownSec = "api.circuit_cooldown_sec"
	keyCLIOutput             = "cli.output"
	keyCLIColor              = "cli.color"
	keyAuthProfile           = "auth.profile"

	defaultAPIURL = "https://erp1.superboss.cc/"
	defaultTimeoutSec = 30
	defaultCLIOutput  = "table"
	defaultAPIRetry   = 3
)

// Manager wraps viper for kuaimai-cli settings.
type Manager struct {
	v *viper.Viper
}

// New creates a config manager (does not require file to exist).
func New() (*Manager, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	dir := util.ConfigDir()
	if dir == "" {
		return nil, fmt.Errorf("无法确定配置目录")
	}
	v.AddConfigPath(dir)
	v.SetDefault(keyAPIURL, defaultAPIURL)
	v.SetDefault(keyAPITimeout, defaultTimeoutSec)
	v.SetDefault(keyCLIOutput, defaultCLIOutput)
	v.SetDefault(keyAPIRetry, defaultAPIRetry)
	v.SetDefault(keyAPIPoolMaxIdle, 100)
	v.SetDefault(keyAPIPoolMaxIdlePerHost, 10)
	v.SetDefault(keyAPICircuitThreshold, 5)
	v.SetDefault(keyAPICircuitCooldownSec, 30)
	v.SetDefault(keyCLIColor, true)
	v.SetDefault(keyAuthProfile, "default")
	_ = v.ReadInConfig()
	m := &Manager{v: v}
	_ = m.ensureProfileRecord("default")
	return m, nil
}

// ConfigPath returns the expected config file path.
func ConfigPath() string {
	return filepath.Join(util.ConfigDir(), fileName)
}

// Init creates ~/.kuaimai-cli/ and writes DefaultConfigTemplate when config.yaml is absent.
func Init() (created bool, err error) {
	dir := util.ConfigDir()
	if dir == "" {
		return false, fmt.Errorf("无法确定配置目录")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return false, fmt.Errorf("创建配置目录失败: %w", err)
	}
	path := ConfigPath()
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	if err := os.WriteFile(path, []byte(DefaultConfigTemplate), 0o600); err != nil {
		return false, fmt.Errorf("写入默认配置失败: %w", err)
	}
	return true, nil
}

// Get returns a config value by dotted key.
func (m *Manager) Get(key string) (any, error) {
	if m.v.IsSet(key) {
		return m.v.Get(key), nil
	}
	if m.v.InConfig(key) {
		return m.v.Get(key), nil
	}
	def := m.v.Get(key)
	if def != nil {
		return def, nil
	}
	return nil, fmt.Errorf("配置项不存在: %s", key)
}

// Set sets a config value and persists the full config file.
func (m *Manager) Set(key, value string) error {
	m.v.Set(key, value)
	return m.Save()
}

// Save writes config to disk.
func (m *Manager) Save() error {
	dir := util.ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return m.v.WriteConfigAs(ConfigPath())
}

// APIURL returns the configured API base URL (erp1.superboss.cc).
func (m *Manager) APIURL() string {
	return m.v.GetString(keyAPIURL)
}

// CLIOutput returns the default CLI output format from config (table/json).
func (m *Manager) CLIOutput() string {
	return strings.TrimSpace(m.v.GetString(keyCLIOutput))
}

// CLIColorEnabled reports whether colored terminal output is enabled.
func (m *Manager) CLIColorEnabled() bool {
	return m.v.GetBool(keyCLIColor)
}

// APIRetry returns configured retry count for HTTP requests.
func (m *Manager) APIRetry() int {
	raw := m.v.Get(keyAPIRetry)
	switch v := raw.(type) {
	case int:
		if v >= 0 {
			return v
		}
	case int64:
		if v >= 0 {
			return int(v)
		}
	case float64:
		if v >= 0 {
			return int(v)
		}
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n >= 0 {
			return n
		}
	}
	return defaultAPIRetry
}

// APIPoolMaxIdle returns api.pool_max_idle.
func (m *Manager) APIPoolMaxIdle() int {
	return intFromConfig(m.v.Get(keyAPIPoolMaxIdle), 100)
}

// APIPoolMaxIdlePerHost returns api.pool_max_idle_per_host.
func (m *Manager) APIPoolMaxIdlePerHost() int {
	return intFromConfig(m.v.Get(keyAPIPoolMaxIdlePerHost), 10)
}

// APICircuitThreshold returns api.circuit_threshold.
func (m *Manager) APICircuitThreshold() int {
	return intFromConfig(m.v.Get(keyAPICircuitThreshold), 5)
}

// APICircuitCooldown returns api.circuit_cooldown_sec as duration.
func (m *Manager) APICircuitCooldown() time.Duration {
	return time.Duration(intFromConfig(m.v.Get(keyAPICircuitCooldownSec), 30)) * time.Second
}

func intFromConfig(raw any, def int) int {
	switch v := raw.(type) {
	case int:
		if v > 0 {
			return v
		}
	case int64:
		if v > 0 {
			return int(v)
		}
	case float64:
		if v > 0 {
			return int(v)
		}
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// APITimeout returns request timeout from api.timeout (seconds int or duration string).
func (m *Manager) APITimeout() time.Duration {
	raw := m.v.Get(keyAPITimeout)
	switch v := raw.(type) {
	case int:
		return time.Duration(v) * time.Second
	case int64:
		return time.Duration(v) * time.Second
	case float64:
		return time.Duration(v) * time.Second
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			break
		}
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if sec, err := strconv.Atoi(v); err == nil {
			return time.Duration(sec) * time.Second
		}
	}
	return defaultTimeoutSec * time.Second
}

// AllSettings returns all keys for display.
func (m *Manager) AllSettings() map[string]any {
	return m.v.AllSettings()
}

// Raw returns the underlying viper (for advanced use in tests).
func (m *Manager) Raw() *viper.Viper {
	return m.v
}

// ActiveProfile returns the current auth profile name.
func (m *Manager) ActiveProfile() string {
	p := strings.TrimSpace(m.v.GetString(keyAuthProfile))
	if p == "" {
		return "default"
	}
	return p
}

// SetActiveProfile switches the active profile and persists config.
func (m *Manager) SetActiveProfile(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("profile 名称不能为空")
	}
	if err := m.ensureProfileRecord(name); err != nil {
		return err
	}
	m.v.Set(keyAuthProfile, name)
	return m.Save()
}

// AddProfile registers a profile in config (idempotent).
func (m *Manager) AddProfile(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("profile 名称不能为空")
	}
	if err := m.ensureProfileRecord(name); err != nil {
		return err
	}
	return m.Save()
}

func (m *Manager) ensureProfileRecord(name string) error {
	key := "auth.profiles." + name
	if !m.v.IsSet(key) {
		m.v.Set(key, map[string]any{})
	}
	return nil
}

// ListProfiles returns known profile names from config.
func (m *Manager) ListProfiles() []string {
	raw, ok := m.v.Get("auth.profiles").(map[string]any)
	if !ok || len(raw) == 0 {
		return []string{"default"}
	}
	names := make([]string, 0, len(raw))
	for k := range raw {
		names = append(names, k)
	}
	sort.Strings(names)
	active := m.ActiveProfile()
	found := false
	for _, n := range names {
		if n == active {
			found = true
			break
		}
	}
	if !found {
		names = append(names, active)
		sort.Strings(names)
	}
	return names
}

// ProfileAPIURL returns api.url for a profile, or the global api.url when unset.
func (m *Manager) ProfileAPIURL(profile string) string {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		profile = m.ActiveProfile()
	}
	key := "auth.profiles." + profile + ".api_url"
	u := strings.TrimSpace(m.v.GetString(key))
	if u != "" {
		return u
	}
	return m.APIURL()
}

// SetProfileAPIURL sets a per-profile API base URL override.
func (m *Manager) SetProfileAPIURL(profile, apiURL string) error {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return fmt.Errorf("profile 名称不能为空")
	}
	if err := m.ensureProfileRecord(profile); err != nil {
		return err
	}
	m.v.Set("auth.profiles."+profile+".api_url", strings.TrimSpace(apiURL))
	return m.Save()
}

