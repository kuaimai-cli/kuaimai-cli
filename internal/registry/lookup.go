package registry

import "fmt"

// ResolvedAPI is a registry entry ready for execution.
type ResolvedAPI struct {
	APIID   string
	Service string
	Svc     Service
	Op      Operation
}

// FindByAPIID locates an operation by global apiId (e.g. item.stock.queryList).
func (m *Metadata) FindByAPIID(apiID string) (*ResolvedAPI, error) {
	if m == nil {
		return nil, fmt.Errorf("registry 未加载")
	}
	svcName, opName := splitAPIID(apiID)
	svc, ok := m.Services[svcName]
	if !ok {
		return nil, fmt.Errorf("未找到接口 %q（无服务 %q）", apiID, svcName)
	}
	op, ok := svc.Operations[opName]
	if !ok {
		return nil, fmt.Errorf("未找到接口 %q", apiID)
	}
	return &ResolvedAPI{
		APIID:   apiID,
		Service: svcName,
		Svc:     svc,
		Op:      op,
	}, nil
}
