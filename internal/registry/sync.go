package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kuaimai-cli/kuaimai-cli/pkg/util"
)

const (
	DefaultRegistrySource = "http://open-cli.kuaimai.com/registry/registry.json"
	registryDirName       = "registry"
	registryFileName      = "registry.json"
)

// SyncResult describes a registry sync operation.
type SyncResult struct {
	Source      string `json:"source"`
	Path        string `json:"path"`
	Version     string `json:"version"`
	APICount    int    `json:"api_count"`
	Changed     bool   `json:"changed"`
	NotModified bool   `json:"not_modified,omitempty"`
}

type fetchResult struct {
	Raw         []byte
	ETag        string
	NotModified bool
}

// CachePath returns ~/.kuaimai-cli/registry/registry.json
func CachePath() string {
	return filepath.Join(util.ConfigDir(), registryDirName, registryFileName)
}

// MetaDataPath is kept for backward-compatible diagnostics; it now points to the cache file.
func MetaDataPath() string {
	return CachePath()
}

// Sync downloads registry JSON from source and writes it to the local cache.
func Sync(ctx context.Context, source string) (*SyncResult, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		source = DefaultRegistrySource
	}

	fetch, err := fetchRegistryConditional(ctx, source, "")
	if err != nil {
		return nil, err
	}
	if fetch.NotModified {
		st, _ := loadSyncState()
		version := ""
		if st != nil {
			version = st.Version
		}
		return &SyncResult{
			Source:      source,
			Path:        CachePath(),
			Version:     version,
			APICount:    cacheAPICount(),
			Changed:     false,
			NotModified: true,
		}, nil
	}

	var doc DocumentV2
	if err := parseRawDocument(fetch.Raw, &doc); err != nil {
		return nil, err
	}

	path := CachePath()
	changed, err := writeCache(path, fetch.Raw)
	if err != nil {
		return nil, err
	}
	_ = saveSyncState(&SyncState{
		Source:     source,
		Version:    doc.Version,
		ETag:       fetch.ETag,
		APICount:   len(doc.APIs),
		LastSyncAt: time.Now(),
	})

	return &SyncResult{
		Source:   source,
		Path:     path,
		Version:  doc.Version,
		APICount: len(doc.APIs),
		Changed:  changed,
	}, nil
}

func fetchRegistryConditional(ctx context.Context, source, etag string) (*fetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, err
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("拉取 registry 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &fetchResult{NotModified: true, ETag: etag}, nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 registry 响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body := strings.TrimSpace(string(raw))
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		return nil, fmt.Errorf("拉取 registry 失败: HTTP %d %s", resp.StatusCode, body)
	}
	newETag := strings.TrimSpace(resp.Header.Get("ETag"))
	if newETag == "" {
		newETag = strings.TrimSpace(resp.Header.Get("Etag"))
	}
	return &fetchResult{Raw: raw, ETag: newETag}, nil
}

func parseJSONDocument(raw []byte, doc *DocumentV2) error {
	if err := json.Unmarshal(raw, doc); err != nil {
		return fmt.Errorf("解析 registry JSON 失败: %w", err)
	}
	return nil
}

func parseRawDocument(raw []byte, doc *DocumentV2) error {
	if err := parseJSONDocument(raw, doc); err != nil {
		return err
	}
	if isEmptyDocumentV2(doc) {
		return fmt.Errorf("远程 registry 为空（apis 无数据）")
	}
	return validateDocumentV2(doc)
}

func cacheAPICount() int {
	_, _, count, ok := CacheInfo()
	if !ok {
		return 0
	}
	return count
}

// SyncIfNeeded pulls remote registry when cache is missing or remote ETag/version changed.
func SyncIfNeeded(ctx context.Context, source string) (*SyncResult, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		source = DefaultRegistrySource
	}

	st, _ := loadSyncState()
	if st != nil && st.Source != "" && st.Source != source {
		st = nil
	}

	if _, err := DocumentFromCache(); err != nil {
		return Sync(ctx, source)
	}

	etag := ""
	version := ""
	if st != nil {
		etag = strings.TrimSpace(st.ETag)
		version = strings.TrimSpace(st.Version)
	}

	fetch, err := fetchRegistryConditional(ctx, source, etag)
	if err != nil {
		return nil, err
	}
	if fetch.NotModified {
		return &SyncResult{
			Source:      source,
			Path:        CachePath(),
			Version:     version,
			APICount:    cacheAPICount(),
			Changed:     false,
			NotModified: true,
		}, nil
	}

	var doc DocumentV2
	if err := parseRawDocument(fetch.Raw, &doc); err != nil {
		return nil, err
	}

	path := CachePath()
	changed, err := writeCache(path, fetch.Raw)
	if err != nil {
		return nil, err
	}
	_ = saveSyncState(&SyncState{
		Source:     source,
		Version:    doc.Version,
		ETag:       fetch.ETag,
		APICount:   len(doc.APIs),
		LastSyncAt: time.Now(),
	})

	return &SyncResult{
		Source:   source,
		Path:     path,
		Version:  doc.Version,
		APICount: len(doc.APIs),
		Changed:  changed,
	}, nil
}

func writeCache(path string, raw []byte) (changed bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return false, fmt.Errorf("创建 registry 缓存目录失败: %w", err)
	}
	if existing, err := os.ReadFile(path); err == nil && string(existing) == string(raw) {
		return false, nil
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return false, fmt.Errorf("写入 registry 缓存失败: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return false, fmt.Errorf("更新 registry 缓存失败: %w", err)
	}
	return true, nil
}

func parseDocumentV2(raw []byte) (*DocumentV2, error) {
	var doc DocumentV2
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("解析 registry JSON 失败: %w", err)
	}
	if isEmptyDocumentV2(&doc) {
		return nil, fmt.Errorf("registry 为空，请先执行: kuaimai-cli registry sync")
	}
	if err := validateDocumentV2(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
