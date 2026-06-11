package registry

import (
	"fmt"
	"strings"
)

// ResolveBodyJSON picks request body/query JSON from flags and schema defaults.
func (r *ResolvedAPI) ResolveBodyJSON(paramsJSON, dataJSON string) (string, error) {
	switch r.Op.ContentType {
	case ContentTypeGetQuery:
		if strings.TrimSpace(paramsJSON) != "" {
			return paramsJSON, nil
		}
		if strings.TrimSpace(dataJSON) != "" {
			return dataJSON, nil
		}
		return r.Op.DefaultBodyJSON(), nil
	case ContentTypePostForm, ContentTypePostJSON:
		if strings.TrimSpace(dataJSON) != "" {
			return dataJSON, nil
		}
		if strings.TrimSpace(paramsJSON) != "" {
			return paramsJSON, nil
		}
		return r.Op.DefaultBodyJSON(), nil
	default:
		return "", fmt.Errorf("接口 %s 未知 contentType: %s", r.APIID, r.Op.ContentType)
	}
}
