package registry

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

//go:embed meta_data.json
var metaFS embed.FS

// ContentType values for request encoding (see docs/kuaimai-cli meta_data.json 定义规范.md).
const (
	ContentTypeGetQuery  = "get_query"
	ContentTypePostForm  = "post_form"
	ContentTypePostJSON  = "post_json"
)

// PropertySchema describes one request/response field in meta_data.json.
// Nested objects use properties; arrays use items (see item-list-v2 responseSchema).
type PropertySchema struct {
	Type       string                    `json:"type,omitempty"`
	Desc       string                    `json:"desc,omitempty"`
	Default    any                       `json:"default,omitempty"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Items      *JSONSchema               `json:"items,omitempty"`
}

// JSONSchema is a lightweight schema object stored in meta_data.json.
type JSONSchema struct {
	Type       string                    `json:"type,omitempty"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// Operation describes a single API operation from metadata.
type Operation struct {
	Name           string      `json:"-"`
	Summary        string      `json:"summary"`
	Method         string      `json:"method"`
	Path           string      `json:"path"`
	ContentType    string      `json:"contentType"`
	Write          bool        `json:"write"`
	Pageable       bool        `json:"pageable,omitempty"`
	RequestSchema  *JSONSchema `json:"requestSchema,omitempty"`
	ResponseSchema *JSONSchema `json:"responseSchema,omitempty"`
}

// Service groups operations under a business scope.
type Service struct {
	Name        string               `json:"-"`
	Summary     string               `json:"summary"`
	Description string               `json:"description"`
	Operations  map[string]Operation `json:"operations"`
}

// Metadata is the full API registry document.
type Metadata struct {
	Version     string             `json:"version"`
	GeneratedAt string             `json:"generated_at,omitempty"`
	Services    map[string]Service `json:"services"`
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
	meta.normalize()
	return &meta, nil
}

func (m *Metadata) normalize() {
	if m.Services == nil {
		return
	}
	for svcName, svc := range m.Services {
		svc.Name = svcName
		if svc.Operations == nil {
			m.Services[svcName] = svc
			continue
		}
		for opName, op := range svc.Operations {
			op.Name = opName
			svc.Operations[opName] = op
		}
		m.Services[svcName] = svc
	}
}

// MetaDataPath returns the on-disk metadata file path.
func MetaDataPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "internal", "registry", "meta_data.json")
}

// ServiceNames returns sorted service keys.
func (m *Metadata) ServiceNames() []string {
	names := make([]string, 0, len(m.Services))
	for name := range m.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// FindService returns a service by name.
func (m *Metadata) FindService(name string) (*Service, error) {
	svc, ok := m.Services[name]
	if !ok {
		return nil, fmt.Errorf("未找到服务: %s", name)
	}
	return &svc, nil
}

// FindOperation locates an operation within a service.
func (s *Service) FindOperation(name string) (*Operation, error) {
	op, ok := s.Operations[name]
	if !ok {
		return nil, fmt.Errorf("未找到操作: %s", name)
	}
	return &op, nil
}

// OperationNames returns sorted operation keys for a service.
func (s *Service) OperationNames() []string {
	names := make([]string, 0, len(s.Operations))
	for name := range s.Operations {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// NeedsBody reports whether the operation accepts a JSON --body flag.
func (op Operation) NeedsBody() bool {
	switch op.ContentType {
	case ContentTypePostForm, ContentTypePostJSON:
		return true
	case ContentTypeGetQuery:
		return op.RequestSchema != nil && len(op.RequestSchema.Properties) > 0
	default:
		return false
	}
}

// FormEncoded reports post_form content type.
func (op Operation) FormEncoded() bool {
	return op.ContentType == ContentTypePostForm
}

// DefaultBodyJSON builds default --body from requestSchema property defaults.
func (op Operation) DefaultBodyJSON() string {
	if op.RequestSchema == nil || len(op.RequestSchema.Properties) == 0 {
		return "{}"
	}
	m := make(map[string]any, len(op.RequestSchema.Properties))
	for key, prop := range op.RequestSchema.Properties {
		if prop.Default != nil {
			m[key] = prop.Default
		}
	}
	if len(m) == 0 {
		return "{}"
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// ValidateRequestBody checks required fields from requestSchema (light validation).
func (op Operation) ValidateRequestBody(body map[string]any) error {
	if op.RequestSchema == nil || len(op.RequestSchema.Required) == 0 {
		return nil
	}
	for _, field := range op.RequestSchema.Required {
		v, ok := body[field]
		if !ok || v == nil {
			return fmt.Errorf("缺少必填参数 %s（见 schema %s）", field, op.Name)
		}
	}
	return nil
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
	for _, svcName := range m.ServiceNames() {
		svc := m.Services[svcName]
		for _, opName := range svc.OperationNames() {
			op := svc.Operations[opName]
			out = append(out, struct {
				Service   string
				Operation Operation
			}{Service: svcName, Operation: op})
		}
	}
	return out
}

// OperationCount returns total operations across services.
func (m *Metadata) OperationCount() int {
	n := 0
	for _, svcName := range m.ServiceNames() {
		n += len(m.Services[svcName].Operations)
	}
	return n
}
