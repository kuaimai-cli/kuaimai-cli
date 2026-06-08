package servicecmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/kuaimai-cli/kuaimai-cli/internal/client"
	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/core"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/kuaimai-cli/kuaimai-cli/shortcuts/common"
	"github.com/spf13/cobra"
)

// Register adds metadata-driven service commands.
func Register(root *cobra.Command) {
	meta, err := registry.Load()
	if err != nil {
		return
	}
	cmd := &cobra.Command{
		Use:   "service",
		Short: "元数据驱动的 API 子命令",
	}
	for _, svcName := range meta.ServiceNames() {
		svc := meta.Services[svcName]
		cmd.AddCommand(serviceGroup(svcName, svc))
	}
	root.AddCommand(cmd)
}

func serviceGroup(svcName string, svc registry.Service) *cobra.Command {
	short := svc.Summary
	if short == "" {
		short = svc.Description
	}
	cmd := &cobra.Command{
		Use:   svcName,
		Short: short,
	}
	for _, opName := range svc.OperationNames() {
		op := svc.Operations[opName]
		cmd.AddCommand(operationCmd(svcName, svc, op))
	}
	return cmd
}

func operationCmd(svcName string, svc registry.Service, op registry.Operation) *cobra.Command {
	short := fmt.Sprintf("%s %s — %s", op.Method, op.Path, op.Summary)
	var bodyJSON string
	defaultBody := op.DefaultBodyJSON()
	c := &cobra.Command{
		Use:   op.Name,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(svcName, svc, op, bodyJSON)
		},
	}
	if op.NeedsBody() {
		c.Flags().StringVar(&bodyJSON, "body", defaultBody, "请求体 JSON（post_form 会转为表单；get_query 会转为 URL 查询参数）")
	}
	return c
}

func runOperation(_ string, svc registry.Service, op registry.Operation, bodyJSON string) error {
	if core.Ctx.DryRun && !op.Write {
		return fmt.Errorf("操作 %s 为查询接口，不支持 --dry-run（请使用 write:true 的写接口）", op.Name)
	}

	f, err := cmdutil.NewFactory()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	baseURL := svc.ResolveBaseURL(f.Config.APIURL())
	r := common.NewRunner(f)
	method := strings.ToUpper(op.Method)

	switch op.ContentType {
	case registry.ContentTypeGetQuery:
		return runGetQuery(r, baseURL, op, method, bodyJSON)
	case registry.ContentTypePostForm, registry.ContentTypePostJSON:
		body, err := common.ParseBodyJSON(bodyJSON)
		if err != nil {
			return err
		}
		if err := op.ValidateRequestBody(body); err != nil {
			return err
		}
		pageAll := op.Pageable && core.Ctx.PageAll
		if core.Ctx.PageAll && !op.Pageable {
			fmt.Fprintln(os.Stderr, "提示: 该操作 pageable=false，已忽略 --page-all")
		}
		return r.ExecuteWrite(context.Background(), common.WriteOptions{
			Method:      method,
			Path:        op.Path,
			Body:        body,
			FormEncoded: op.FormEncoded(),
			BaseURL:     baseURL,
			PageAll:     &pageAll,
		})
	default:
		return fmt.Errorf("未知 contentType: %s", op.ContentType)
	}
}

func runGetQuery(r *common.Runner, baseURL string, op registry.Operation, method, bodyJSON string) error {
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
		return fmt.Errorf("操作 %s 需要 --body 提供查询参数（见 kuaimai-cli schema）", op.Name)
	}
	return r.ExecuteWithBase(context.Background(), baseURL, func(ctx context.Context, c *client.Client) (any, error) {
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
