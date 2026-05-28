package item

import "testing"

func TestPrepareSaveBody(t *testing.T) {
	item := map[string]any{
		"sysItemId": float64(123),
		"title":     "旧标题",
		"itemSuiteBridgeList": []any{
			map[string]any{"id": 1},
		},
	}
	body := PrepareSaveBody(item, "新标题")
	if body["title"] != "新标题" {
		t.Fatalf("title = %v", body["title"])
	}
	if _, ok := body["itemSuiteBridgeList"]; ok {
		t.Fatal("itemSuiteBridgeList should be removed")
	}
	if body["suiteBridgeList"] == nil {
		t.Fatal("suiteBridgeList should be set")
	}
}
