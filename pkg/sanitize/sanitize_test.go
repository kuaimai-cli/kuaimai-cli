package sanitize

import (
	"strings"
	"testing"
)

func TestMap_redactsToken(t *testing.T) {
	in := map[string]any{
		"title":       "test",
		"accessToken": "abcdefghijklmnop",
	}
	out := Map(in)
	if out["title"] != "test" {
		t.Fatalf("title = %v", out["title"])
	}
	got, _ := out["accessToken"].(string)
	if got == "abcdefghijklmnop" {
		t.Fatal("expected masked token")
	}
}

func TestJSONString_nested(t *testing.T) {
	raw := `{"user":{"token":"secret12345"},"pageNo":1}`
	got := JSONString(raw)
	if strings.Contains(got, "secret12345") {
		t.Fatalf("token leaked: %s", got)
	}
}
