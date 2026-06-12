package scm

import (
	"context"
	"reflect"
	"testing"
)

type fakeClient struct {
	getCalls  []string
	postCalls []string
	postBody  map[string]any
}

func (f *fakeClient) PostJSON(ctx context.Context, path string, body any) (any, int, error) {
	f.postCalls = append(f.postCalls, path)
	if path == pathPddBatchPublish {
		if m, ok := body.(map[string]any); ok {
			f.postBody = m
		}
		return map[string]any{"result": float64(1), "data": "task-1"}, 200, nil
	}
	switch path {
	case pathItemBasePage:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"records": []any{
					map[string]any{
						"outerId":                "揭阳仓#M01-SCM-VPEgL",
						"baseItemId":             "92ee1d081c7d000",
						"title":                  "测试商品",
						"companyId":              float64(30482),
						"isDistribution":         float64(0),
						"canPublishPlatformList": []any{"pdd", "fxg"},
					},
				},
			},
		}, 200, nil
	case pathPddSaveTempConf:
		return map[string]any{"result": float64(1)}, 200, nil
	case pathPddQueryDetail:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"batchDetailList": []any{
					map[string]any{
						"baseItemId":       "92ee1d081c7d000",
						"informationLack":  float64(1),
						"name":             "测试商品",
						"shopBrandConf":    []any{map[string]any{"shopId": float64(123), "vvalue": "品牌"}},
						"skuDetailList":    []any{map[string]any{"skuId": "sku-1", "skuSpecifications": []any{"drop"}}},
						"batchFissionConf": nil,
					},
				},
			},
		}, 200, nil
	case pathPublishLog:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"total": float64(1),
				"records": []any{
					map[string]any{
						"id":             "log-1",
						"operationLogId": "op-1",
						"platformType":   "pdd",
						"shopId":         float64(123),
						"shopName":       "zhengweihao",
					},
				},
			},
		}, 200, nil
	default:
		return nil, 0, nil
	}
}

func (f *fakeClient) GetQuery(ctx context.Context, path string, params map[string]any) (any, int, error) {
	f.getCalls = append(f.getCalls, path)
	switch path {
	case pathShopAll:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"list": []any{
					map[string]any{
						"id":       float64(123),
						"shopId":   float64(123),
						"title":    "zhengweihao",
						"source":   "pdd",
						"state":    float64(2),
						"deadline": "",
					},
				},
			},
		}, 200, nil
	case pathPublishLogDetail:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"records": []any{
					map[string]any{
						"operationLogId": "op-1",
						"outerId":        "揭阳仓#M01-SCM-VPEgL",
						"status":         float64(4),
						"errorMessage":   "类目属性缺失",
					},
				},
			},
		}, 200, nil
	case pathPublishLogByID:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"id":     "op-1",
				"status": float64(4),
			},
		}, 200, nil
	default:
		return nil, 0, nil
	}
}

func TestExecutePublishPDDPlansWithoutSubmit(t *testing.T) {
	f := &fakeClient{}
	got, err := executePublishPDD(context.Background(), f, publishPDDOptions{
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
		Shop:      "zhengweihao",
	})
	if err != nil {
		t.Fatal(err)
	}
	plan := got.(map[string]any)
	if plan["dry_run"] != true {
		t.Fatalf("dry_run = %v, want true", plan["dry_run"])
	}
	if containsPath(f.postCalls, pathPddBatchPublish) {
		t.Fatalf("unexpected submit call: %#v", f.postCalls)
	}
	body := plan["publish_body"].(map[string]any)
	if !reflect.DeepEqual(body["shopIds"], []any{int64(123)}) {
		t.Fatalf("shopIds = %#v", body["shopIds"])
	}
	details := body["batchItemDetailList"].([]map[string]any)
	skus := details[0]["skuDetailList"].([]map[string]any)
	if _, ok := skus[0]["skuSpecifications"]; ok {
		t.Fatal("skuSpecifications should be removed for pdd publish")
	}
	if _, ok := details[0]["shopBrandConf"].(string); !ok {
		t.Fatalf("shopBrandConf should be serialized, got %#v", details[0]["shopBrandConf"])
	}
}

func TestExecuteListProductsReturnsSCMItems(t *testing.T) {
	f := &fakeClient{}
	got, err := executeListProducts(context.Background(), f, listProductsOptions{
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	records := out["records"].([]map[string]any)
	if len(records) != 1 {
		t.Fatalf("records len = %d", len(records))
	}
	platforms := records[0]["canPublishPlatform"].([]string)
	if !containsString(platforms, "pdd") {
		t.Fatalf("canPublishPlatform = %#v", platforms)
	}
	if !containsPath(f.postCalls, pathItemBasePage) {
		t.Fatalf("missing item query call: %#v", f.postCalls)
	}
}

func TestExecuteShopsReturnsAvailability(t *testing.T) {
	f := &fakeClient{}
	got, err := executeShops(context.Background(), f, shopsOptions{
		Platform:  "pdd",
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	shops := out["shops"].([]map[string]any)
	if len(shops) != 1 {
		t.Fatalf("shops len = %d", len(shops))
	}
	if shops[0]["can_publish"] != true {
		t.Fatalf("can_publish = %#v", shops[0]["can_publish"])
	}
	if !containsPath(f.getCalls, pathShopAll) {
		t.Fatalf("missing shop query call: %#v", f.getCalls)
	}
}

func TestExecutePublishPDDSubmitsWhenRequested(t *testing.T) {
	f := &fakeClient{}
	_, err := executePublishPDD(context.Background(), f, publishPDDOptions{
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
		ShopID:    123,
		Submit:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !containsPath(f.postCalls, pathPddBatchPublish) {
		t.Fatalf("missing submit call: %#v", f.postCalls)
	}
	if f.postBody == nil || f.postBody["flowNumber"] == "" {
		t.Fatalf("submit body missing flowNumber: %#v", f.postBody)
	}
}

func TestExecutePublishPDDCanCheckLogAfterSubmit(t *testing.T) {
	f := &fakeClient{}
	got, err := executePublishPDD(context.Background(), f, publishPDDOptions{
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
		ShopID:    123,
		Submit:    true,
		CheckLog:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	plan := got.(map[string]any)
	if plan["publish_log"] == nil {
		t.Fatalf("missing publish_log: %#v", plan)
	}
	if !containsPath(f.postCalls, pathPublishLog) {
		t.Fatalf("missing log query call: %#v", f.postCalls)
	}
	if !containsPath(f.getCalls, pathPublishLogDetail) {
		t.Fatalf("missing log detail call: %#v", f.getCalls)
	}
}

func TestExecutePublishLogReturnsFailureReason(t *testing.T) {
	f := &fakeClient{}
	got, err := executePublishLog(context.Background(), f, publishLogOptions{
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
		Shop:      "zhengweihao",
		Detail:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	details := out["details"].([]map[string]any)
	summary := details[0]["summary"].(map[string]any)
	failures := summary["failures"].([]map[string]any)
	if failures[0]["errorMessage"] != "类目属性缺失" {
		t.Fatalf("errorMessage = %#v", failures[0]["errorMessage"])
	}
}

func containsPath(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
