package webcmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kuaimai-cli/kuaimai-cli/internal/apicall"
	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/internal/registry"
	"github.com/kuaimai-cli/kuaimai-cli/shortcuts/common"
	"github.com/spf13/cobra"
)

// Register adds web call command group.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "按 registry apiId 调用 Web 接口",
		Long: `按远端 registry.json 中的 apiId 发起 HTTP 请求（对标飞书域命令调用）。

推荐流程：capabilities → schema <apiId> → web call <apiId>`,
	}
	cmd.AddCommand(callCmd())
	root.AddCommand(cmd)
}

func callCmd() *cobra.Command {
	var paramsJSON string
	var dataJSON string
	var bodyJSON string
	cmd := &cobra.Command{
		Use:   "call <apiId>",
		Short: "按 apiId 调用 registry 接口",
		Args:  cobra.ExactArgs(1),
		Example: `  kuaimai-cli web call api.luotao.test.get --params '{"keyword":"测试"}'
  kuaimai-cli web call api.luotao.test.post --data '{"title":"测试商品"}'
  kuaimai-cli web call scm.staff-query --body '{"pageNo":1,"pageSize":20}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebCall(args[0], paramsJSON, dataJSON, bodyJSON)
		},
	}
	cmd.Flags().StringVar(&paramsJSON, "params", "", "GET 查询参数 JSON（get_query）")
	cmd.Flags().StringVar(&dataJSON, "data", "", "POST 请求体 JSON（post_json / post_form）")
	cmd.Flags().StringVar(&bodyJSON, "body", "", "统一请求体 JSON（按 contentType 自动转为 params 或 data）")
	return cmd
}

func runWebCall(apiID, paramsJSON, dataJSON, bodyJSON string) error {
	meta, err := registry.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	resolved, err := meta.FindByAPIID(apiID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	paramsJSON, dataJSON, err = resolveCallBody(resolved, paramsJSON, dataJSON, bodyJSON)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	baseURL := resolved.Svc.ResolveBaseURL(f.Config.APIURL())
	r := common.NewRunner(f)
	return apicall.Execute(context.Background(), resolved, apicall.Options{
		BaseURL:    baseURL,
		ParamsJSON: paramsJSON,
		DataJSON:   dataJSON,
		Runner:     r,
	})
}

func resolveCallBody(resolved *registry.ResolvedAPI, paramsJSON, dataJSON, bodyJSON string) (string, string, error) {
	set := 0
	if strings.TrimSpace(paramsJSON) != "" {
		set++
	}
	if strings.TrimSpace(dataJSON) != "" {
		set++
	}
	if strings.TrimSpace(bodyJSON) != "" {
		set++
	}
	if set > 1 {
		return "", "", fmt.Errorf("请只使用 --params、--data、--body 其中之一")
	}
	if bodyJSON == "" {
		return paramsJSON, dataJSON, nil
	}
	switch resolved.Op.ContentType {
	case registry.ContentTypeGetQuery:
		return bodyJSON, "", nil
	default:
		return "", bodyJSON, nil
	}
}

