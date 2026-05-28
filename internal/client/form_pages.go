package client

import (
	"context"
)

// PostFormAllPages POSTs form bodies with pageNo/pageSize until no more rows.
func (c *Client) PostFormAllPages(ctx context.Context, path string, baseBody map[string]any) (any, error) {
	if c.dryRun {
		data, _, err := c.PostForm(ctx, path, MapToFormValues(baseBody))
		return data, err
	}

	pageNo := formPageNo(baseBody)
	pageSize := formPageSize(baseBody)
	if pageSize <= 0 {
		pageSize = 50
	}

	allItems := make([]map[string]any, 0)
	for page := pageNo; page < pageNo+1000; page++ {
		body := cloneFormBody(baseBody)
		body["pageNo"] = page
		body["pageSize"] = pageSize

		data, _, err := c.PostForm(ctx, path, MapToFormValues(body))
		if err != nil {
			return nil, err
		}
		items := extractItems(data)
		allItems = append(allItems, items...)
		if !formPageHasMore(data, page, pageSize, len(items)) {
			break
		}
	}
	return allItems, nil
}

func cloneFormBody(base map[string]any) map[string]any {
	out := make(map[string]any, len(base))
	for k, v := range base {
		out[k] = v
	}
	return out
}

func formPageNo(body map[string]any) int {
	switch v := body["pageNo"].(type) {
	case float64:
		if v >= 1 {
			return int(v)
		}
	case int:
		if v >= 1 {
			return v
		}
	case int64:
		if v >= 1 {
			return int(v)
		}
	}
	return 1
}

func formPageSize(body map[string]any) int {
	switch v := body["pageSize"].(type) {
	case float64:
		if v > 0 {
			return int(v)
		}
	case int:
		if v > 0 {
			return v
		}
	case int64:
		if v > 0 {
			return int(v)
		}
	}
	return 50
}

func formPageHasMore(data any, pageNo, pageSize, itemCount int) bool {
	if itemCount == 0 {
		return false
	}
	if m, ok := data.(map[string]any); ok {
		if total := nestedTotal(m); total > 0 {
			return pageNo*pageSize < total
		}
	}
	return itemCount >= pageSize
}

func nestedTotal(m map[string]any) int {
	for _, key := range []string{"data", "result"} {
		if inner, ok := m[key].(map[string]any); ok {
			if n := numericInt(inner["total"]); n > 0 {
				return n
			}
		}
	}
	return numericInt(m["total"])
}

func numericInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}
