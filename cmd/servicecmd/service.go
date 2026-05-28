package servicecmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kuaimai/kuaimai-cli/internal/client"
	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/internal/core"
	"github.com/kuaimai/kuaimai-cli/internal/registry"
	"github.com/kuaimai/kuaimai-cli/shortcuts/common"
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
	for _, svc := range meta.Services {
		cmd.AddCommand(serviceGroup(svc))
	}
	root.AddCommand(cmd)
}

func serviceGroup(svc registry.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   svc.Name,
		Short: svc.Description,
	}
	for _, op := range svc.Operations {
		cmd.AddCommand(operationCmd(op))
	}
	return cmd
}

func operationCmd(op registry.Operation) *cobra.Command {
	short := fmt.Sprintf("%s %s — %s", op.Method, op.Path, op.Description)
	var bodyJSON string
	c := &cobra.Command{
		Use:   op.Name,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOperation(op, bodyJSON)
		},
	}
	if op.Write {
		c.Flags().StringVar(&bodyJSON, "body", "{}", "请求体 JSON")
	}
	return c
}

func runOperation(op registry.Operation, bodyJSON string) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	r := common.NewRunner(f)
	method := strings.ToUpper(op.Method)

	if op.Write || client.IsWriteMethod(method) {
		body, err := common.ParseBodyJSON(bodyJSON)
		if err != nil {
			return err
		}
		return r.ExecuteWrite(context.Background(), common.WriteOptions{
			Method:      method,
			Path:        op.Path,
			Body:        body,
			FormEncoded: op.FormEncoded,
		})
	}

	if op.Paginated || core.Ctx.PageAll {
		return r.ExecuteList(context.Background(), common.ListOptions{
			Method: method,
			Path:   op.Path,
		})
	}
	return r.Execute(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
		data, _, err := c.Request(ctx, method, op.Path, nil)
		if err != nil {
			return nil, err
		}
		return common.NormalizeList(data), nil
	})
}
