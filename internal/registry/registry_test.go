package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func testRegistryPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "registry.json")
}

func TestLoadFromFile(t *testing.T) {
	meta, err := LoadFromFile(testRegistryPath(t))
	if err != nil {
		t.Fatal(err)
	}
	if meta.Version != "2026.06.11.99" {
		t.Fatalf("version: got %q", meta.Version)
	}
	if meta.OperationCount() != 2 {
		t.Fatalf("expected 2 operations, got %d", meta.OperationCount())
	}
	svc, err := meta.FindService("api")
	if err != nil {
		t.Fatal(err)
	}
	op, err := svc.FindOperation("luotao.test.get")
	if err != nil || op.Path != "/api/luotao/test/get" {
		t.Fatalf("luotao.test.get: %+v err=%v", op, err)
	}
	if op.ContentType != ContentTypeGetQuery {
		t.Fatalf("contentType: got %s", op.ContentType)
	}
	if op.BaseURL != "https://erp1.superboss.cc/" {
		t.Fatalf("baseUrl: got %q", op.BaseURL)
	}
	if op.Write {
		t.Fatal("get should be read-only")
	}
}

func TestValidateRequestBody(t *testing.T) {
	meta, err := LoadFromFile(testRegistryPath(t))
	if err != nil {
		t.Fatal(err)
	}
	svc, _ := meta.FindService("api")
	op, _ := svc.FindOperation("luotao.test.post")
	if err := op.ValidateRequestBody(map[string]any{}); err == nil {
		t.Fatal("expected missing title error")
	}
	if err := op.ValidateRequestBody(map[string]any{"title": "x"}); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultBodyJSON(t *testing.T) {
	meta, err := LoadFromFile(testRegistryPath(t))
	if err != nil {
		t.Fatal(err)
	}
	svc, _ := meta.FindService("api")
	op, err := svc.FindOperation("luotao.test.get")
	if err != nil {
		t.Fatal(err)
	}
	body := op.DefaultBodyJSON()
	if body == "{}" {
		t.Fatal("expected keyword default")
	}
}

func TestSplitAPIID(t *testing.T) {
	svc, op := splitAPIID("item.stock.queryList")
	if svc != "item" || op != "stock.queryList" {
		t.Fatalf("got %q / %q", svc, op)
	}
	svc, op = splitAPIID("api.luotao.test.get")
	if svc != "api" || op != "luotao.test.get" {
		t.Fatalf("got %q / %q", svc, op)
	}
}

func TestSyncWritesCache(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", dir)

	// Use local test fixture via file:// is not supported; mock by writing cache directly.
	raw, err := os.ReadFile(testRegistryPath(t))
	if err != nil {
		t.Fatal(err)
	}
	changed, err := writeCache(CachePath(), raw)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected first write to mark changed")
	}
	meta, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if meta.OperationCount() != 2 {
		t.Fatalf("cache load: got %d ops", meta.OperationCount())
	}
	_ = oldHome
}

func TestFindByAPIID(t *testing.T) {
	meta, err := LoadFromFile(testRegistryPath(t))
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := meta.FindByAPIID("api.luotao.test.get")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Op.Path != "/api/luotao/test/get" {
		t.Fatalf("path: got %q", resolved.Op.Path)
	}
	if got := resolved.Svc.ResolveOperationBaseURL(resolved.Op, "fallback"); got != "https://erp1.superboss.cc" {
		t.Fatalf("resolved baseUrl = %q", got)
	}
	post, err := meta.FindByAPIID("api.luotao.test.post")
	if err != nil {
		t.Fatal(err)
	}
	if got := post.Svc.ResolveOperationBaseURL(post.Op, "fallback"); got != "https://scm3.superboss.cc" {
		t.Fatalf("post baseUrl = %q", got)
	}
	if _, err := meta.FindByAPIID("missing.api"); err == nil {
		t.Fatal("expected error for missing api")
	}
}

func TestResolveBodyJSON(t *testing.T) {
	meta, err := LoadFromFile(testRegistryPath(t))
	if err != nil {
		t.Fatal(err)
	}
	getAPI, _ := meta.FindByAPIID("api.luotao.test.get")
	body, err := getAPI.ResolveBodyJSON(`{"keyword":"x"}`, "")
	if err != nil || body != `{"keyword":"x"}` {
		t.Fatalf("params: got %q err=%v", body, err)
	}
	postAPI, _ := meta.FindByAPIID("api.luotao.test.post")
	body, err = postAPI.ResolveBodyJSON("", `{"title":"y"}`)
	if err != nil || body != `{"title":"y"}` {
		t.Fatalf("data: got %q err=%v", body, err)
	}
}

func TestValidateDocumentV2RejectsEmptyAPIs(t *testing.T) {
	err := validateDocumentV2(&DocumentV2{
		SchemaVersion: "2.0",
		Version:       "2026.06.11.1",
		APIs:          map[string]APIEntry{},
	})
	if err == nil {
		t.Fatal("expected error for empty apis")
	}
}
