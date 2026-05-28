package registry

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed meta_data.json
var metaFS embed.FS

// Operation describes a single API operation from metadata.
type Operation struct {
	Name        string `json:"name"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Paginated   bool   `json:"paginated,omitempty"`
	Write       bool   `json:"write,omitempty"`
	FormEncoded bool   `json:"form_encoded,omitempty"`
}

// Service groups operations under a business scope.
type Service struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Operations  []Operation `json:"operations"`
}

// Metadata is the full API registry document.
type Metadata struct {
	Version     string    `json:"version"`
	GeneratedAt string    `json:"generated_at"`
	Services    []Service `json:"services"`
}

// Load reads meta_data.json from disk or falls back to embedded copy.
func Load() (*Metadata, error) {
	path := MetaDataPath()
	raw, err := metaFS.ReadFile("meta_data.json")
	if err != nil {
		return nil, fmt.Errorf("读取内置 API 元数据失败: %w", err)
	}
	if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
		raw = b
	}
	var meta Metadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, fmt.Errorf("解析 API 元数据失败: %w", err)
	}
	return &meta, nil
}

// MetaDataPath returns the on-disk metadata file path.
func MetaDataPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "internal", "registry", "meta_data.json")
}

// FindService returns a service by name.
func (m *Metadata) FindService(name string) (*Service, error) {
	for i := range m.Services {
		if m.Services[i].Name == name {
			return &m.Services[i], nil
		}
	}
	return nil, fmt.Errorf("未找到服务: %s", name)
}

// FindOperation locates an operation within a service.
func (s *Service) FindOperation(name string) (*Operation, error) {
	for i := range s.Operations {
		if s.Operations[i].Name == name {
			return &s.Operations[i], nil
		}
	}
	return nil, fmt.Errorf("未找到操作: %s", name)
}

// AllOperations flattens service/operation pairs for schema display.
func (m *Metadata) AllOperations() []struct {
	Service   string
	Operation Operation
} {
	var out []struct {
		Service   string
		Operation Operation
	}
	for _, svc := range m.Services {
		for _, op := range svc.Operations {
			out = append(out, struct {
				Service   string
				Operation Operation
			}{Service: svc.Name, Operation: op})
		}
	}
	return out
}
