package common

import (
	"fmt"
	"net/url"
	"strings"
)

// BuildPath appends URL-encoded query parameters to path.
func BuildPath(path string, params url.Values) string {
	if len(params) == 0 {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + params.Encode()
}

// SetQuery sets a query parameter when value is non-empty.
func SetQuery(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

// QueryFromJSONMap flattens a JSON object map into URL query values (scalar values only).
func QueryFromJSONMap(m map[string]any) url.Values {
	q := url.Values{}
	for k, v := range m {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case string:
			if val != "" {
				q.Set(k, val)
			}
		case bool:
			q.Set(k, fmt.Sprintf("%t", val))
		case float64:
			q.Set(k, fmt.Sprintf("%v", val))
		default:
			q.Set(k, fmt.Sprint(v))
		}
	}
	return q
}

// BuildPathWithJSONQuery parses body JSON and appends scalar fields as query parameters.
func BuildPathWithJSONQuery(path, bodyJSON string) (string, error) {
	m, err := ParseBodyJSON(bodyJSON)
	if err != nil {
		return "", err
	}
	return BuildPath(path, QueryFromJSONMap(m)), nil
}
