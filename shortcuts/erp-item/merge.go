package item

import "fmt"

// ExtractDetailItem picks the first item map from a get-detail API payload.
func ExtractDetailItem(body any) (map[string]any, error) {
	switch v := body.(type) {
	case []map[string]any:
		if len(v) == 0 {
			return nil, fmt.Errorf("商品详情为空")
		}
		return v[0], nil
	case []any:
		if len(v) == 0 {
			return nil, fmt.Errorf("商品详情为空")
		}
		if m, ok := v[0].(map[string]any); ok {
			return m, nil
		}
	case map[string]any:
		if data, ok := v["data"]; ok {
			return ExtractDetailItem(data)
		}
		return v, nil
	}
	return nil, fmt.Errorf("无法解析商品详情")
}

// PrepareSaveBody clones detail and applies title + suiteBridgeList field rename.
func PrepareSaveBody(item map[string]any, newTitle string) map[string]any {
	body := make(map[string]any, len(item)+1)
	for k, val := range item {
		body[k] = val
	}
	body["title"] = newTitle
	if bridge, ok := body["itemSuiteBridgeList"]; ok {
		body["suiteBridgeList"] = bridge
		delete(body, "itemSuiteBridgeList")
	}
	return body
}
