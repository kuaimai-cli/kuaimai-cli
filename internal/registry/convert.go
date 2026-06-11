package registry

import (
	"fmt"
	"strings"
)

func documentV2ToMetadata(doc *DocumentV2) (*Metadata, error) {
	if doc == nil {
		return nil, fmt.Errorf("registry 文档为空")
	}
	if err := validateDocumentV2(doc); err != nil {
		return nil, err
	}

	services := make(map[string]Service)
	for apiID, entry := range doc.APIs {
		if strings.TrimSpace(entry.ID) == "" {
			entry.ID = apiID
		}
		if entry.ID != apiID {
			return nil, fmt.Errorf("apis.%s.id 与 key 不一致: %s", apiID, entry.ID)
		}
		svcName, opName := splitAPIID(apiID)
		svc, ok := services[svcName]
		if !ok {
			svc = Service{
				Summary:    serviceSummary(svcName),
				Operations: make(map[string]Operation),
			}
		}
		if strings.TrimSpace(entry.BaseURL) != "" && strings.TrimSpace(svc.BaseURL) == "" {
			svc.BaseURL = strings.TrimSpace(entry.BaseURL)
		}
		svc.Operations[opName] = apiEntryToOperation(opName, entry)
		services[svcName] = svc
	}

	meta := &Metadata{
		Version:     doc.Version,
		GeneratedAt: doc.GeneratedAt,
		Services:    services,
	}
	meta.normalize()
	return meta, nil
}

func splitAPIID(id string) (service, operation string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "web", ""
	}
	if i := strings.Index(id, "."); i > 0 {
		return id[:i], id[i+1:]
	}
	return "web", id
}

func serviceSummary(name string) string {
	switch name {
	case "item":
		return "ERP 商品域接口"
	case "scm":
		return "供应链域接口"
	case "api":
		return "Registry 测试/扩展接口"
	default:
		return name + " 域接口"
	}
}

func apiEntryToOperation(name string, entry APIEntry) Operation {
	summary := strings.TrimSpace(entry.Title)
	if summary == "" {
		summary = strings.TrimSpace(entry.Description)
	}
	return Operation{
		Name:           name,
		Summary:        summary,
		Method:         strings.ToUpper(strings.TrimSpace(entry.Method)),
		Path:           entry.Path,
		ContentType:    entry.ContentType,
		Write:          entry.Write,
		Pageable:       entry.Pageable,
		RequestSchema:  entry.RequestSchema,
		ResponseSchema: entry.ResponseSchema,
	}
}
