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
