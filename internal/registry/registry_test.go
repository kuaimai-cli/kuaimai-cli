package registry

import "testing"

func TestLoadEmbeddedMetadata(t *testing.T) {
	meta, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if meta.Version == "" {
		t.Fatal("empty version")
	}
	if len(meta.ServiceNames()) < 2 {
		t.Fatalf("expected at least 2 services, got %d", len(meta.ServiceNames()))
	}
	svc, err := meta.FindService("item")
	if err != nil {
		t.Fatal(err)
	}
	op, err := svc.FindOperation("item-detail")
	if err != nil || op.Path != "/item/getItemDetail" {
		t.Fatalf("item-detail: %+v err=%v", op, err)
	}
	if op.ContentType != ContentTypeGetQuery {
		t.Fatalf("contentType: got %s", op.ContentType)
	}
	if op.Write {
		t.Fatal("item-detail should be read-only (write=false)")
	}
	if op.Pageable {
		t.Fatal("item-detail should not be pageable")
	}
	scm, err := meta.FindService("scm")
	if err != nil {
		t.Fatal(err)
	}
	if scm.BaseURL != "https://scm.superboss.cc/" {
		t.Fatalf("scm baseUrl: got %q", scm.BaseURL)
	}
	if scm.ResolveBaseURL("https://erp1.superboss.cc/") != "https://scm.superboss.cc" {
		t.Fatalf("scm ResolveBaseURL: got %q", scm.ResolveBaseURL("https://erp1.superboss.cc/"))
	}
	if item := meta.Services["item"]; item.ResolveBaseURL("https://erp1.superboss.cc/") != "https://erp1.superboss.cc" {
		t.Fatalf("item ResolveBaseURL fallback: got %q", item.ResolveBaseURL("https://erp1.superboss.cc/"))
	}
	if meta.OperationCount() < 100 {
		t.Fatalf("expected at least 100 operations, got %d", meta.OperationCount())
	}
}

func TestValidateRequestBody(t *testing.T) {
	meta, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	svc, _ := meta.FindService("item")
	op, _ := svc.FindOperation("item-detail")
	if err := op.ValidateRequestBody(map[string]any{}); err == nil {
		t.Fatal("expected missing sysItemId error")
	}
	if err := op.ValidateRequestBody(map[string]any{"sysItemId": 1}); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultBodyJSON(t *testing.T) {
	meta, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	svc, _ := meta.FindService("item")
	op, err := svc.FindOperation("item-query-list-v2")
	if err != nil {
		t.Fatal(err)
	}
	if !op.Pageable {
		t.Fatal("item-query-list-v2 should be pageable")
	}
	body := op.DefaultBodyJSON()
	if body == "{}" {
		t.Fatal("expected defaults for pageNo/pageSize")
	}
}
