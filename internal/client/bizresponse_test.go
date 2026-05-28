package client

import "testing"

func TestParseBusinessError(t *testing.T) {
	if ParseBusinessError(map[string]any{"result": float64(1), "message": "ok"}) != nil {
		t.Fatal("result=1 should be success")
	}
	err := ParseBusinessError(map[string]any{
		"result":  float64(901),
		"message": "会话异常，请重新登录",
	})
	if err == nil || err.Result != 901 {
		t.Fatalf("expected 901 error, got %v", err)
	}
	if err.Message != "会话异常，请重新登录" {
		t.Fatalf("message = %q", err.Message)
	}
}

func TestEnrichBusinessError901(t *testing.T) {
	err := EnrichBusinessError(&BusinessError{Result: 901, Message: "会话异常"}, "abcd...wxyz")
	if err.TokenPreview != "abcd...wxyz" {
		t.Fatalf("token preview = %q", err.TokenPreview)
	}
	if err.Hint == "" {
		t.Fatal("expected enriched hint")
	}
}
