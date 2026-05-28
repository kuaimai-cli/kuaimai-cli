package client

import "testing"

func TestFormPageHasMore_byTotal(t *testing.T) {
	data := map[string]any{
		"data": map[string]any{
			"total": float64(120),
		},
	}
	if !formPageHasMore(data, 1, 50, 50) {
		t.Fatal("expected more pages")
	}
	if formPageHasMore(data, 3, 50, 20) {
		t.Fatal("expected last page")
	}
}

func TestFormPageNo_defaults(t *testing.T) {
	if got := formPageNo(map[string]any{}); got != 1 {
		t.Fatalf("pageNo = %d", got)
	}
}
