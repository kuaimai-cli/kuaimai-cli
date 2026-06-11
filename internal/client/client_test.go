package client

import "testing"

func TestIsWriteMethod(t *testing.T) {
	cases := map[string]bool{
		"GET":    false,
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}
	for method, want := range cases {
		if got := IsWriteMethod(method); got != want {
			t.Fatalf("%s: got %v want %v", method, got, want)
		}
	}
}

func TestSplitPathQuery(t *testing.T) {
	path, qp := splitPathQuery("/item/list?page=1&pageSize=20")
	if path != "/item/list" {
		t.Fatalf("path=%q", path)
	}
	if qp["page"] != "1" || qp["pageSize"] != "20" {
		t.Fatalf("query=%v", qp)
	}

	path, qp = splitPathQuery("item/list")
	if path != "/item/list" || qp != nil {
		t.Fatalf("path=%q query=%v", path, qp)
	}
}

func TestBuildTargetURL(t *testing.T) {
	got := buildTargetURL("https://erp1.superboss.cc", "/item/list", map[string]string{"page": "1"})
	want := "https://erp1.superboss.cc/item/list?page=1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
