package item

import (
	"net/url"
	"strconv"

	"github.com/kuaimai/kuaimai-cli/shortcuts/common"
	"github.com/spf13/cobra"
)

const (
	stockBasePath     = "/item/stock"
	queryListPath     = stockBasePath + "/queryList"
	// QueryCountPath is POST /item/stock/queryCount (used by auth check probe).
	QueryCountPath = stockBasePath + "/queryCount"
	queryCountPath = QueryCountPath
	getItemDetailPath = "/item/getItemDetail"
	saveItemPath      = "/item/saveItem"

	defaultListAPIName = "item_stock_queryList"

	// 与浏览器 POST /item/stock/queryList 表单字段对齐（ARCHIVE_V2 商品库存页）
	defaultStockListBody = `{"isAccurate":0,"tagQueryType":0,"suiteQueryType":0,"skuLevelStockWarnDiff":false,"isAccurateSku":0,"suiteSearchType":1,"searchItems":"[]","orderColumn":"","orderDesc":false,"shipperItemFlag":0,"pageType":"ITEM_STOCK","subPageType":"ARCHIVE_V2","pageNo":1,"pageSize":50,"api_name":"item_stock_queryList"}`
)

// Register attaches erp-items-core item commands to root.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "item",
		Short: "商品命令（erp-items-core /item）",
	}
	cmd.AddCommand(listCmd(false))
	cmd.AddCommand(listCmd(true))
	cmd.AddCommand(countCmd())
	cmd.AddCommand(getDetailCmd())
	cmd.AddCommand(saveCmd())
	cmd.AddCommand(updateTitleCmd())
	root.AddCommand(cmd)
}

func listCmd(shortcut bool) *cobra.Command {
	use := "list"
	short := "商品库存列表（POST /item/stock/queryList，form 表单）"
	if shortcut {
		use = "+list"
		short = "商品库存列表（快捷命令）"
	}
	var bodyJSON string
	c := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := common.ParseBodyJSON(bodyJSON)
			if err != nil {
				return err
			}
			ApplyStockListDefaults(body)
			return common.RunPOSTFormListMap(queryListPath, body)
		},
	}
	c.Flags().StringVar(&bodyJSON, "body", defaultStockListBody, "筛选与分页 JSON（QueryItemStockListRequest，与浏览器 form 字段一致）")
	return c
}

func countCmd() *cobra.Command {
	var bodyJSON string
	c := &cobra.Command{
		Use:   "count",
		Short: "商品库存总数（POST /item/stock/queryCount，form 表单）",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := common.ParseBodyJSON(bodyJSON)
			if err != nil {
				return err
			}
			ApplyStockListDefaults(body)
			return common.RunPOSTFormMap(queryCountPath, body)
		},
	}
	c.Flags().StringVar(&bodyJSON, "body", defaultStockListBody, "筛选条件 JSON（与 list 相同字段）")
	return c
}

// ApplyStockListDefaults fills list/count form defaults (ARCHIVE_V2).
func ApplyStockListDefaults(body map[string]any) {
	if body == nil {
		return
	}
	if _, ok := body["api_name"]; !ok {
		body["api_name"] = defaultListAPIName
	}
	if _, ok := body["pageType"]; !ok {
		body["pageType"] = "ITEM_STOCK"
	}
	if _, ok := body["subPageType"]; !ok {
		body["subPageType"] = "ARCHIVE_V2"
	}
}

func getDetailCmd() *cobra.Command {
	var sysItemID int64
	var apiName string
	c := &cobra.Command{
		Use:   "get-detail",
		Short: "商品详情（GET /item/getItemDetail）",
		RunE: func(cmd *cobra.Command, args []string) error {
			q := url.Values{}
			q.Set("sysItemId", strconv.FormatInt(sysItemID, 10))
			common.SetQuery(q, "api_name", apiName)
			return common.RunGET(common.BuildPath(getItemDetailPath, q))
		},
	}
	c.Flags().Int64Var(&sysItemID, "sys-item-id", 0, "系统商品 ID sysItemId（必填）")
	c.Flags().StringVar(&apiName, "api-name", "", "接口名 api_name（可选）")
	_ = c.MarkFlagRequired("sys-item-id")
	return c
}

func saveCmd() *cobra.Command {
	var bodyJSON string
	var syncUpdateHSCode bool
	var apiName string
	c := &cobra.Command{
		Use:   "save",
		Short: "保存/修改商品（POST /item/saveItem）",
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := common.ParseBodyJSON(bodyJSON)
			if err != nil {
				return err
			}
			q := url.Values{}
			if syncUpdateHSCode {
				q.Set("syncUpdateHSCode", "true")
			}
			common.SetQuery(q, "api_name", apiName)
			path := common.BuildPath(saveItemPath, q)
			return common.RunPOST(path, body)
		},
	}
	c.Flags().StringVar(&bodyJSON, "body", "{}", "商品 JSON（SysItemModel）")
	c.Flags().BoolVar(&syncUpdateHSCode, "sync-hs-code", false, "是否同步更新海关编码 syncUpdateHSCode")
	c.Flags().StringVar(&apiName, "api-name", "", "接口名 api_name（可选）")
	return c
}
