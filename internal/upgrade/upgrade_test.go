package upgrade

import "testing"

func TestVersionLess(t *testing.T) {
	cases := []struct {
		a, b   string
		less   bool
	}{
		{"0.1.0", "0.2.0", true},
		{"0.2.0", "0.1.0", false},
		{"1.0.0", "1.0.0", false},
		{"dev", "1.0.0", true},
	}
	for _, tc := range cases {
		got := versionLess(tc.a, tc.b)
		if got != tc.less {
			t.Fatalf("versionLess(%q,%q)=%v want %v", tc.a, tc.b, got, tc.less)
		}
	}
}
