package apicall

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/kuaimai-cli/kuaimai-cli/internal/client"
	"github.com/kuaimai-cli/kuaimai-cli/internal/core"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/kuaimai-cli/kuaimai-cli/shortcuts/common"
)

// Options configures registry-driven HTTP execution.
type Options struct {
	BaseURL    string
	ParamsJSON string
	DataJSON   string
	Runner     *common.Runner
}

// Execute runs a resolved registry API.
func Execute(ctx context.Context, resolved *registry.ResolvedAPI, opts Options) error {
	if opts.Runner == nil {
		return fmt.Errorf("Runner 不能为空")
	}
	if core.Ctx.DryRun && !resolved.Op.Write {
		return fmt.Errorf("接口 %s 为查询接口，不支持 --dry-run", resolved.APIID)
	}

	bodyJSON, err := resolved.ResolveBodyJSON(opts.ParamsJSON, opts.DataJSON)
	if err != nil {
		return err
	}

	method := strings.ToUpper(resolved.Op.Method)
	switch resolved.Op.ContentType {
	case registry.ContentTypeGetQuery:
		return runGetQuery(ctx, opts.Runner, opts.BaseURL, resolved.Op, method, bodyJSON)
	case registry.ContentTypePostForm, registry.ContentTypePostJSON:
		body, err := common.ParseBodyJSON(bodyJSON)
		if err != nil {
			return err
		}
		if err := resolved.Op.ValidateRequestBody(body); err != nil {
			return err
		}
		pageAll := resolved.Op.Pageable && core.Ctx.PageAll
		if core.Ctx.PageAll && !resolved.Op.Pageable {
			fmt.Fprintln(os.Stderr, "提示: 该接口 pageable=false，已忽略 --page-all")
		}
		return opts.Runner.ExecuteWrite(ctx, common.WriteOptions{
			Method:      method,
			Path:        resolved.Op.Path,
			Body:        body,
			FormEncoded: resolved.Op.FormEncoded(),
			BaseURL:     opts.BaseURL,
			PageAll:     &pageAll,
		})
	default:
		return fmt.Errorf("接口 %s 未知 contentType: %s", resolved.APIID, resolved.Op.ContentType)
	}
}

func runGetQuery(ctx context.Context, r *common.Runner, baseURL string, op registry.Operation, method, bodyJSON string) error {
	path := op.Path
	if strings.TrimSpace(bodyJSON) != "" && bodyJSON != "{}" {
		body, err := common.ParseBodyJSON(bodyJSON)
		if err != nil {
			return err
		}
		if err := op.ValidateRequestBody(body); err != nil {
			return err
		}
		q := url.Values{}
		for k, v := range body {
			q.Set(k, formatQueryValue(v))
		}
		if enc := q.Encode(); enc != "" {
			path = op.Path + "?" + enc
		}
	} else if op.RequestSchema != nil && len(op.RequestSchema.Required) > 0 {
		return fmt.Errorf("接口 %s 需要 --params 提供查询参数（见 kuaimai-cli schema %s）", op.Name, op.Name)
	}
	return r.ExecuteWithBase(ctx, baseURL, func(ctx context.Context, c *client.Client) (any, error) {
		data, _, err := c.Request(ctx, method, path, nil)
		if err != nil {
			return nil, err
		}
		return data, nil
	})
}

func formatQueryValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	default:
		return fmt.Sprint(v)
	}
}
