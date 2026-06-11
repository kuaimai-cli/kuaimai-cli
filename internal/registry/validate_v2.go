package registry

import (
	"fmt"
	"strings"
)

const schemaVersionV2 = "2.0"

func validateDocumentV2(doc *DocumentV2) error {
	if doc.SchemaVersion != schemaVersionV2 {
		return fmt.Errorf("schemaVersion 必须为 %s，当前为 %q", schemaVersionV2, doc.SchemaVersion)
	}
	if strings.TrimSpace(doc.Version) == "" {
		return fmt.Errorf("version 不能为空")
	}
	if len(doc.APIs) == 0 {
		return fmt.Errorf("apis 不能为空")
	}
	for apiID, entry := range doc.APIs {
		if strings.TrimSpace(entry.Path) == "" {
			return fmt.Errorf("apis.%s.path 不能为空", apiID)
		}
		if strings.TrimSpace(entry.Method) == "" {
			return fmt.Errorf("apis.%s.method 不能为空", apiID)
		}
		if strings.TrimSpace(entry.ContentType) == "" {
			return fmt.Errorf("apis.%s.contentType 不能为空", apiID)
		}
	}
	return nil
}

func isEmptyDocumentV2(doc *DocumentV2) bool {
	return doc == nil || len(doc.APIs) == 0
}
