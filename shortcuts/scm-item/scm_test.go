package scm

import (
	"context"
	"reflect"
	"testing"
)

type fakeClient struct {
	getCalls   []string
	postCalls  []string
	postBody   map[string]any
	postBodies map[string]map[string]any
}

func (f *fakeClient) PostJSON(ctx context.Context, path string, body any) (any, int, error) {
	f.postCalls = append(f.postCalls, path)
	if f.postBodies == nil {
		f.postBodies = map[string]map[string]any{}
	}
	if m, ok := body.(map[string]any); ok {
		f.postBodies[path] = m
	}
	if path == pathStorageTask {
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
						"id":                     float64(456),
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
	case pathPddCarouselVideo:
		return map[string]any{"result": float64(1)}, 200, nil
	case pathPreCheckPrice:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"isLimit":        float64(0),
				"isDistribution": float64(0),
			},
		}, 200, nil
	case pathPddAuthorize:
		return map[string]any{
			"result": float64(1),
			"data": []any{
				map[string]any{
					"shopId":         "849217672",
					"authorizeValid": true,
					"tokenValid":     true,
				},
			},
		}, 200, nil
	case pathSecConfirm:
		return map[string]any{"result": float64(1), "data": []any{}}, 200, nil
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
						"taobaoId": float64(849217672),
						"title":    "zhengweihao",
						"source":   "pdd",
						"state":    float64(2),
						"deadline": "",
					},
				},
			},
		}, 200, nil
	case pathQueryTaskSpeed:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"finished": true,
				"result": map[string]any{
					"total":   float64(1),
					"success": float64(1),
					"failure": float64(0),
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
	case pathItemBaseDetail:
		return map[string]any{
			"result": float64(1),
			"data": map[string]any{
				"id":         float64(456),
				"baseItemId": "92ee1d081c7d000",
				"outerId":    "揭阳仓#M01-SCM-VPEgL",
				"title":      "测试商品",
				"companyId":  float64(30482),
				"skuList": []any{
					map[string]any{"skuId": "sku-1", "outerId": "sku-code-1"},
				},
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
	if containsPath(f.postCalls, pathStorageTask) {
		t.Fatalf("unexpected submit call: %#v", f.postCalls)
	}
	body := plan["publish_body"].(map[string]any)
	if !reflect.DeepEqual(body["shopIds"], []any{int64(123)}) {
		t.Fatalf("shopIds = %#v", body["shopIds"])
	}
	if body["taskType"] != defaultPublishType {
		t.Fatalf("taskType = %#v", body["taskType"])
	}
	if body["api_name"] != "taskScheduling_storageTask" {
		t.Fatalf("api_name = %#v", body["api_name"])
	}
	if !containsInterface(plan, interfaceStorageTask) {
		t.Fatalf("missing interface %s: %#v", interfaceStorageTask, plan["interfaces"])
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
	if !containsInterface(out, interfaceItemBasePage) {
		t.Fatalf("missing interface %s: %#v", interfaceItemBasePage, out["interfaces"])
	}
	if pathItemBasePage != "/item/base/page.json" {
		t.Fatalf("pathItemBasePage = %q", pathItemBasePage)
	}
	body := f.postBodies[pathItemBasePage]
	if body["api_name"] != apiNameItemBasePage {
		t.Fatalf("api_name = %#v", body["api_name"])
	}
	for _, key := range []string{"leafCategories", "cgSupplierIds", "platformItemIds", "companyNames", "shopNames"} {
		if _, ok := body[key].([]any); !ok {
			t.Fatalf("%s should be []any, got %#v", key, body[key])
		}
	}
}

func TestExecuteListProductsAllowsNoFilter(t *testing.T) {
	f := &fakeClient{}
	got, err := executeListProducts(context.Background(), f, listProductsOptions{
		PageNo:   1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	if !containsInterface(out, interfaceItemBasePage) {
		t.Fatalf("missing interface %s: %#v", interfaceItemBasePage, out["interfaces"])
	}
	body := f.postBodies[pathItemBasePage]
	if _, ok := body["outerIds"]; ok {
		t.Fatalf("outerIds should be omitted without style-code: %#v", body)
	}
	if _, ok := body["title"]; ok {
		t.Fatalf("title should be omitted without title filter: %#v", body)
	}
}

func TestExecuteUpdateTitlePlansWithoutSubmit(t *testing.T) {
	f := &fakeClient{}
	got, err := executeUpdateTitle(context.Background(), f, updateTitleOptions{
		StyleCode:  "揭阳仓#M01-SCM-VPEgL",
		Title:      "新商品名称",
		SkipAddERP: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	if out["dry_run"] != true {
		t.Fatalf("dry_run = %#v", out["dry_run"])
	}
	if containsPath(f.postCalls, pathQueryErpItems) || containsPath(f.postCalls, pathItemBaseEdit) {
		t.Fatalf("unexpected write calls: %#v", f.postCalls)
	}
	if !containsPath(f.postCalls, pathItemBasePage) {
		t.Fatalf("missing style-code lookup call: %#v", f.postCalls)
	}
	if !containsPath(f.getCalls, pathItemBaseDetail) {
		t.Fatalf("missing detail call: %#v", f.getCalls)
	}
	saveBody := out["save_body"].(map[string]any)
	item := saveBody["item"].(map[string]any)
	if item["title"] != "新商品名称" {
		t.Fatalf("title = %#v", item["title"])
	}
	if saveBody["checkOpenSync"] != false {
		t.Fatalf("checkOpenSync = %#v", saveBody["checkOpenSync"])
	}
	if saveBody["skipAddItemToErp"] != true {
		t.Fatalf("skipAddItemToErp = %#v", saveBody["skipAddItemToErp"])
	}
	if !containsInterface(out, interfaceItemBaseDetail) || !containsInterface(out, interfaceItemBaseEdit) {
		t.Fatalf("missing edit interfaces: %#v", out["interfaces"])
	}
}

func TestExecuteUpdateTitleSubmitsSyncThenEdit(t *testing.T) {
	f := &fakeClient{}
	got, err := executeUpdateTitle(context.Background(), f, updateTitleOptions{
		ID:         456,
		Title:      "新商品名称",
		Submit:     true,
		SkipAddERP: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	if out["dry_run"] != false {
		t.Fatalf("dry_run = %#v", out["dry_run"])
	}
	if !containsPath(f.postCalls, pathQueryErpItems) {
		t.Fatalf("missing sync check call: %#v", f.postCalls)
	}
	if !containsPath(f.postCalls, pathItemBaseEdit) {
		t.Fatalf("missing edit call: %#v", f.postCalls)
	}
	saveBody := f.postBodies[pathItemBaseEdit]
	item := saveBody["item"].(map[string]any)
	if item["title"] != "新商品名称" {
		t.Fatalf("saved title = %#v", item["title"])
	}
	if _, ok := f.postBodies[pathItemBasePage]; ok {
		t.Fatalf("id path should not query list: %#v", f.postBodies[pathItemBasePage])
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
	if !containsInterface(out, interfaceShopAll) {
		t.Fatalf("missing interface %s: %#v", interfaceShopAll, out["interfaces"])
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
	if !containsPath(f.postCalls, pathStorageTask) {
		t.Fatalf("missing submit call: %#v", f.postCalls)
	}
	if f.postBody == nil || f.postBody["taskType"] != defaultPublishType {
		t.Fatalf("submit body missing taskType: %#v", f.postBody)
	}
}

func TestExecutePublishPlatformFXGPlansWithoutPDDOnlyCalls(t *testing.T) {
	f := &fakeClient{}
	got, err := executePublishPlatform(context.Background(), f, publishPlatformOptions{
		Platform:  "fxg",
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
		ShopID:    123,
	})
	if err != nil {
		t.Fatal(err)
	}
	plan := got.(map[string]any)
	if plan["platform"] != "fxg" {
		t.Fatalf("platform = %#v", plan["platform"])
	}
	if containsPath(f.postCalls, pathPddCarouselVideo) || containsPath(f.postCalls, pathPddAuthorize) {
		t.Fatalf("fxg should not call PDD-only APIs: %#v", f.postCalls)
	}
	if !containsPath(f.postCalls, pathPreCheckPrice) {
		t.Fatalf("missing precheck call: %#v", f.postCalls)
	}
	if !containsPath(f.postCalls, pathSecConfirm) {
		t.Fatalf("missing sec confirm call: %#v", f.postCalls)
	}
	body := plan["publish_body"].(map[string]any)
	if body["taskType"] != defaultPublishType {
		t.Fatalf("taskType = %#v", body["taskType"])
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
	if !containsInterface(plan, interfaceStorageTask) {
		t.Fatalf("missing interface %s: %#v", interfaceStorageTask, plan["interfaces"])
	}
	if !containsInterface(plan, interfaceQueryTask) {
		t.Fatalf("missing interface %s: %#v", interfaceQueryTask, plan["interfaces"])
	}
	if !containsPath(f.postCalls, pathPublishLog) {
		t.Fatalf("missing log query call: %#v", f.postCalls)
	}
	if !containsPath(f.getCalls, pathPublishLogDetail) {
		t.Fatalf("missing log detail call: %#v", f.getCalls)
	}
}

func TestPDDPrimitivesExposeSingleInterfaces(t *testing.T) {
	f := &fakeClient{}
	got, err := executePDDPricePrecheck(context.Background(), f, pddPrimitiveOptions{
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
		ShopID:    123,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	if !containsInterface(out, interfacePreCheckPrice) {
		t.Fatalf("missing interface %s: %#v", interfacePreCheckPrice, out["interfaces"])
	}
	req := out["request"].(map[string]any)
	if req["api_name"] != "ltsTask_preCheckControllerPrice" {
		t.Fatalf("api_name = %#v", req["api_name"])
	}
	if !containsPath(f.postCalls, pathPreCheckPrice) {
		t.Fatalf("missing precheck call: %#v", f.postCalls)
	}
}

func TestPDDStorageTaskPrimitiveDefaultsToDryRun(t *testing.T) {
	f := &fakeClient{}
	got, err := executePDDStorageTask(context.Background(), f, pddPrimitiveOptions{
		StyleCode: "揭阳仓#M01-SCM-VPEgL",
		ShopID:    123,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	if out["dry_run"] != true {
		t.Fatalf("dry_run = %#v", out["dry_run"])
	}
	if containsPath(f.postCalls, pathStorageTask) {
		t.Fatalf("unexpected submit call: %#v", f.postCalls)
	}
	req := out["request"].(map[string]any)
	if req["taskType"] != defaultPublishType {
		t.Fatalf("taskType = %#v", req["taskType"])
	}
}

func TestPDDTaskSpeedPrimitive(t *testing.T) {
	f := &fakeClient{}
	got, err := executePDDTaskSpeed(context.Background(), f, pddPrimitiveOptions{
		BatchTaskID: "task-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	out := got.(map[string]any)
	if !containsInterface(out, interfaceQueryTask) {
		t.Fatalf("missing interface %s: %#v", interfaceQueryTask, out["interfaces"])
	}
	if !containsPath(f.getCalls, pathQueryTaskSpeed) {
		t.Fatalf("missing task speed call: %#v", f.getCalls)
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
	if !containsInterface(out, interfacePublishLogDetail) {
		t.Fatalf("missing interface %s: %#v", interfacePublishLogDetail, out["interfaces"])
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

func containsInterface(out map[string]any, want string) bool {
	items, ok := out["interfaces"].([]map[string]any)
	if !ok {
		return false
	}
	for _, item := range items {
		if item["name"] == want {
			return true
		}
	}
	return false
}
