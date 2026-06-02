package client

import (
	"context"
	"fmt"
	"os"

	"github.com/kuaimai-cli/kuaimai-cli/internal/core"
	"github.com/kuaimai-cli/kuaimai-cli/internal/pagination"
	"github.com/kuaimai-cli/kuaimai-cli/pkg/logger"
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

	res, err := pagination.CollectPages(core.PaginationSettings(), pageNo, pageSize,
		func(page, size int) ([]map[string]any, bool, int, error) {
			body := cloneFormBody(baseBody)
			body["pageNo"] = page
			body["pageSize"] = size

			data, _, err := c.PostForm(ctx, path, MapToFormValues(body))
			if err != nil {
				return nil, false, 0, err
			}
			items := extractItems(data)
			return items, formPageHasMore(data, page, size, len(items)), nestedTotalFromAny(data), nil
		})
	if err != nil {
		return nil, err
	}
	emitPageNotice(res)
	return res.Items, nil
}

func emitPageNotice(res pagination.Result) {
	if res.Notice == "" {
		return
	}
	fmt.Fprintln(os.Stderr, res.Notice)
	logger.Info("%s (reason=%s truncated=%v)", res.Notice, res.Reason, res.Truncated)
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
