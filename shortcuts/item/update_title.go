package item

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kuaimai/kuaimai-cli/internal/client"
	"github.com/kuaimai/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai/kuaimai-cli/shortcuts/common"
	"github.com/spf13/cobra"
)

func updateTitleCmd() *cobra.Command {
	var sysItemID int64
	var title string
	c := &cobra.Command{
		Use:   "update-title",
		Short: "修改商品标题（get-detail 合并后 save）",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateTitle(sysItemID, title)
		},
	}
	c.Flags().Int64Var(&sysItemID, "sys-item-id", 0, "系统商品 ID sysItemId（必填）")
	c.Flags().StringVar(&title, "title", "", "新标题（必填）")
	_ = c.MarkFlagRequired("sys-item-id")
	_ = c.MarkFlagRequired("title")
	return c
}

func runUpdateTitle(sysItemID int64, newTitle string) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	runner := common.NewRunner(f)

	q := url.Values{}
	q.Set("sysItemId", strconv.FormatInt(sysItemID, 10))
	detailPath := common.BuildPath(getItemDetailPath, q)

	var saveBody map[string]any
	err = runner.Execute(context.Background(), func(ctx context.Context, httpClient *client.Client) (any, error) {
		raw, _, err := httpClient.Request(ctx, "GET", detailPath, nil)
		if err != nil {
			return nil, err
		}
		item, err := ExtractDetailItem(raw)
		if err != nil {
			return nil, err
		}
		saveBody = PrepareSaveBody(item, newTitle)
		return map[string]any{
			"sysItemId": sysItemID,
			"title":     newTitle,
		}, nil
	})
	if err != nil {
		return err
	}

	return runner.ExecuteWrite(context.Background(), common.WriteOptions{
		Method: "POST",
		Path:   saveItemPath,
		Body:   saveBody,
	})
}
