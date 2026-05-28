package auth

import "testing"

func TestLoginWithTokenEmpty(t *testing.T) {
	if err := LoginWithToken("default", "  "); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestTokenPreview(t *testing.T) {
	got := TokenPreview("ed47c0b0c5e56008507126daa931d82c")
	if got == "" || got == "ed47c0b0c5e56008507126daa931d82c" {
		t.Fatalf("expected masked preview, got %q", got)
	}
}
