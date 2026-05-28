package sanitize

import (
	"encoding/json"
	"strings"
)

// SensitiveKeys are redacted in maps and JSON logs.
var SensitiveKeys = []string{
	"accessToken",
	"access_token",
	"token",
	"password",
	"secret",
	"authorization",
}

// Map returns a copy of m with sensitive string values masked.
func Map(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			out[k] = maskValue(v)
			continue
		}
		switch child := v.(type) {
		case map[string]any:
			out[k] = Map(child)
		case []any:
			out[k] = Slice(child)
		default:
			out[k] = v
		}
	}
	return out
}

// Slice sanitizes each element in a slice.
func Slice(items []any) []any {
	out := make([]any, len(items))
	for i, v := range items {
		switch child := v.(type) {
		case map[string]any:
			out[i] = Map(child)
		case []any:
			out[i] = Slice(child)
		default:
			out[i] = v
		}
	}
	return out
}

// JSONString parses raw JSON, redacts sensitive keys, and returns compact JSON.
func JSONString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return MaskString(raw)
	}
	sanitized := Value(v)
	b, err := json.Marshal(sanitized)
	if err != nil {
		return raw
	}
	return string(b)
}

// Value sanitizes an arbitrary decoded JSON value.
func Value(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return Map(t)
	case []any:
		return Slice(t)
	default:
		return v
	}
}

// MaskString hides most characters in a secret string.
func MaskString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

// Headers masks sensitive HTTP header values.
func Headers(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		if isSensitiveKey(k) {
			out[k] = MaskString(v)
			continue
		}
		out[k] = v
	}
	return out
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	for _, sk := range SensitiveKeys {
		if lower == strings.ToLower(sk) {
			return true
		}
	}
	return false
}

func maskValue(v any) string {
	switch t := v.(type) {
	case string:
		return MaskString(t)
	default:
		return "****"
	}
}
