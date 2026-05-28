package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// MapToFormValues converts a flat map to application/x-www-form-urlencoded fields.
func MapToFormValues(m map[string]any) url.Values {
	q := url.Values{}
	for k, v := range m {
		if v == nil {
			continue
		}
		q.Set(k, formFieldValue(v))
	}
	return q
}

func formFieldValue(v any) string {
	switch val := v.(type) {
	case bool:
		return strconv.FormatBool(val)
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case []any:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(b)
	case map[string]any:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(b)
	default:
		return fmt.Sprint(v)
	}
}

// ResolveAPIURL builds the full request URL from base URL and path.
func ResolveAPIURL(baseURL, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	} else if path[0] != '/' {
		path = "/" + path
	}
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + path
}
