package client

import "testing"

func TestMapToFormValues(t *testing.T) {
	q := MapToFormValues(map[string]any{
		"api_name":       "staff_query",
		"pageNo":         1,
		"pageSize":       20,
		"queryStaffName": "",
	})
	got := q.Encode()
	want := "api_name=staff_query&pageNo=1&pageSize=20&queryStaffName="
	if got != want {
		t.Fatalf("Encode() = %q, want %q", got, want)
	}
}

func TestMapToFormValuesBoolAndInt(t *testing.T) {
	q := MapToFormValues(map[string]any{
		"isAccurate": 0,
		"title":      "2026",
		"orderDesc":  false,
		"pageNo":     float64(1),
		"pageSize":   float64(50),
		"searchItems": "[]",
	})
	if q.Get("orderDesc") != "false" {
		t.Fatalf("orderDesc = %q", q.Get("orderDesc"))
	}
	if q.Get("title") != "2026" {
		t.Fatalf("title = %q", q.Get("title"))
	}
	if q.Get("pageNo") != "1" {
		t.Fatalf("pageNo = %q", q.Get("pageNo"))
	}
}

func TestResolveAPIURL(t *testing.T) {
	got := ResolveAPIURL("https://erp1.superboss.cc/", "/item/stock/queryList")
	want := "https://erp1.superboss.cc/item/stock/queryList"
	if got != want {
		t.Fatalf("ResolveAPIURL() = %q, want %q", got, want)
	}
}
