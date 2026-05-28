package common

import (
	"net/url"
	"testing"
)

func TestBuildPath(t *testing.T) {
	q := url.Values{}
	q.Set("pageNo", "1")
	q.Set("source", "fxg")
	got := BuildPath("/shop/info", q)
	want := "/shop/info?pageNo=1&source=fxg"
	if got != want && got != "/shop/info?source=fxg&pageNo=1" {
		t.Fatalf("BuildPath = %q, want %q or swapped order", got, want)
	}
}

func TestBuildPathWithJSONQuery(t *testing.T) {
	got, err := BuildPathWithJSONQuery("/item/stock/queryList", `{"pageNo":1,"pageSize":20}`)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/item/stock/queryList?pageNo=1&pageSize=20" && got != "/item/stock/queryList?pageSize=20&pageNo=1" {
		t.Fatalf("BuildPathWithJSONQuery = %q", got)
	}
}
