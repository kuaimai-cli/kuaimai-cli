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
	if len(meta.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(meta.Services))
	}
	svc, err := meta.FindService("item")
	if err != nil {
		t.Fatal(err)
	}
	op, err := svc.FindOperation("list")
	if err != nil || op.Path != "/item/stock/queryList" {
		t.Fatalf("item list: %+v err=%v", op, err)
	}
}
