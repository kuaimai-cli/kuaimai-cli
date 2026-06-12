package scm

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/kuaimai-cli/kuaimai-cli/internal/client"
	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/shortcuts/common"
	"github.com/spf13/cobra"
)

const (
	scmBaseURL = "https://scm3.superboss.cc/"

	pathItemBasePage      = "/item/base/page"
	pathShopAll           = "/shop/allShop"
	pathPddSaveTempConf   = "/pdd/saveBatchTempConf"
	pathPddQueryDetail    = "/pdd/queryBatchDetail"
	pathPddBatchPublish   = "/pdd/batchPublishItem"
	pathPublishLog        = "/logging/publishLog"
	pathPublishLogDetail  = "/logging/publishLogDetail"
	pathPublishLogByID    = "/logging/publishLogById"
	defaultPddShelfState  = 1
	defaultPddPublishType = "PUBLISH_ITEM"
	defaultLogPageSize    = 10
)

type httpClient interface {
	PostJSON(context.Context, string, any) (any, int, error)
	GetQuery(context.Context, string, map[string]any) (any, int, error)
}

type publishPDDOptions struct {
	StyleCode  string
	Shop       string
	ShopID     int64
	ShelfState int
	Submit     bool
	CheckLog   bool
	LogStart   string
	LogEnd     string
	LogPage    int
	LogSize    int
	Detail     bool
}

type listProductsOptions struct {
	StyleCode string
	Title     string
	PageNo    int
	PageSize  int
}

type shopsOptions struct {
	Platform   string
	StyleCode  string
	BaseItemID string
	Shop       string
	ShopID     int64
}

type publishLogOptions struct {
	StyleCode string
	Shop      string
	ShopID    int64
	StartTime string
	EndTime   string
	PageNo    int
	PageSize  int
	Detail    bool
}

// Register attaches SCM item shortcuts to root.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "scm-item",
		Short: "SCM 可铺货商品命令（erp-scm，可上货/铺货到店铺）",
	}
	cmd.AddCommand(listProductsCmd())
	cmd.AddCommand(shopsCmd())
	cmd.AddCommand(publishPDDCmd())
	cmd.AddCommand(publishLogCmd())
	root.AddCommand(cmd)
}

func listProductsCmd() *cobra.Command {
	var opts listProductsOptions
	c := &cobra.Command{
		Use:   "+list",
		Short: "查询 SCM 可铺货商品列表（可按款式编码精确定位）",
		Example: `  kuaimai-cli scm-item +list --style-code '<款式编码>' --output json
  kuaimai-cli scm-item +list --title '<标题关键词>' --page-size 20 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListProducts(context.Background(), opts)
		},
	}
	c.Flags().StringVar(&opts.StyleCode, "style-code", "", "款式编码 outerId（精确查询）")
	c.Flags().StringVar(&opts.Title, "title", "", "标题关键词（简单查询；复杂条件请用 capabilities/schema/web call）")
	c.Flags().IntVar(&opts.PageNo, "page", 1, "页码")
	c.Flags().IntVar(&opts.PageSize, "page-size", 10, "每页条数")
	return c
}

func shopsCmd() *cobra.Command {
	var opts shopsOptions
	c := &cobra.Command{
		Use:   "shops",
		Short: "查询 SCM 商品可铺货店铺及不可铺货原因",
		Example: `  kuaimai-cli scm-item shops --platform pdd --style-code '<款式编码>' --output json
  kuaimai-cli scm-item shops --platform pdd --style-code '<款式编码>' --shop '<店铺名>' --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShops(context.Background(), opts)
		},
	}
	c.Flags().StringVar(&opts.Platform, "platform", "pdd", "平台 source，例如 pdd")
	c.Flags().StringVar(&opts.StyleCode, "style-code", "", "款式编码 outerId（与 --base-item-id 二选一）")
	c.Flags().StringVar(&opts.BaseItemID, "base-item-id", "", "SCM 商品 baseItemId/itemId（与 --style-code 二选一）")
	c.Flags().StringVar(&opts.Shop, "shop", "", "店铺简称/名称过滤（可选）")
	c.Flags().Int64Var(&opts.ShopID, "shop-id", 0, "店铺 ID 过滤（可选）")
	return c
}

func publishPDDCmd() *cobra.Command {
	var opts publishPDDOptions
	c := &cobra.Command{
		Use:   "publish-pdd",
		Short: "将 SCM 商品铺货到拼多多店铺（默认只预检，不提交）",
		Example: `  kuaimai-cli scm-item publish-pdd --style-code '<款式编码>' --shop '<拼多多店铺名>' --output json
  kuaimai-cli scm-item publish-pdd --style-code '<款式编码>' --shop '<拼多多店铺名>' --submit --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPublishPDD(context.Background(), opts)
		},
	}
	c.Flags().StringVar(&opts.StyleCode, "style-code", "", "款式编码 outerId（必填）")
	c.Flags().StringVar(&opts.Shop, "shop", "", "拼多多店铺简称/名称（与 --shop-id 二选一）")
	c.Flags().Int64Var(&opts.ShopID, "shop-id", 0, "拼多多店铺 ID（与 --shop 二选一）")
	c.Flags().IntVar(&opts.ShelfState, "shelf-state", defaultPddShelfState, "上架设置 shelfState：1 立即上架，2 放入草稿箱")
	c.Flags().BoolVar(&opts.Submit, "submit", false, "通过校验后实际提交 /pdd/batchPublishItem")
	c.Flags().BoolVar(&opts.CheckLog, "check-log", false, "提交后查询最近铺货日志与失败原因")
	c.Flags().StringVar(&opts.LogStart, "log-start-time", "", "日志查询开始时间 yyyy-MM-dd HH:mm:ss（默认近 30 天）")
	c.Flags().StringVar(&opts.LogEnd, "log-end-time", "", "日志查询结束时间 yyyy-MM-dd HH:mm:ss（默认当前时间）")
	c.Flags().IntVar(&opts.LogPage, "log-page", 1, "日志查询页码")
	c.Flags().IntVar(&opts.LogSize, "log-page-size", defaultLogPageSize, "日志查询每页条数")
	_ = c.MarkFlagRequired("style-code")
	return c
}

func publishLogCmd() *cobra.Command {
	var opts publishLogOptions
	c := &cobra.Command{
		Use:   "publish-log",
		Short: "查询 SCM 铺货日志，可按款式编码和店铺筛选失败原因",
		Example: `  kuaimai-cli scm-item publish-log --style-code '<款式编码>' --shop '<拼多多店铺名>' --detail --output json
  kuaimai-cli scm-item publish-log --style-code '<款式编码>' --shop-id 123456 --start-time '2026-06-01 00:00:00' --end-time '2026-06-12 23:59:59' --detail --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPublishLog(context.Background(), opts)
		},
	}
	c.Flags().StringVar(&opts.StyleCode, "style-code", "", "款式编码 outerId（可选；提供后会在日志明细中过滤）")
	c.Flags().StringVar(&opts.Shop, "shop", "", "店铺简称/名称（可选）")
	c.Flags().Int64Var(&opts.ShopID, "shop-id", 0, "店铺 ID（可选）")
	c.Flags().StringVar(&opts.StartTime, "start-time", "", "开始时间 yyyy-MM-dd HH:mm:ss（默认近 30 天）")
	c.Flags().StringVar(&opts.EndTime, "end-time", "", "结束时间 yyyy-MM-dd HH:mm:ss（默认当前时间）")
	c.Flags().IntVar(&opts.PageNo, "page", 1, "页码")
	c.Flags().IntVar(&opts.PageSize, "page-size", defaultLogPageSize, "每页条数")
	c.Flags().BoolVar(&opts.Detail, "detail", false, "拉取日志明细，返回单品状态与失败原因 errorMessage")
	return c
}

func runPublishPDD(ctx context.Context, opts publishPDDOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, scmBaseURL, func(ctx context.Context, c *client.Client) (any, error) {
		return executePublishPDD(ctx, c, opts)
	})
}

func runListProducts(ctx context.Context, opts listProductsOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, scmBaseURL, func(ctx context.Context, c *client.Client) (any, error) {
		return executeListProducts(ctx, c, opts)
	})
}

func runShops(ctx context.Context, opts shopsOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, scmBaseURL, func(ctx context.Context, c *client.Client) (any, error) {
		return executeShops(ctx, c, opts)
	})
}

func runPublishLog(ctx context.Context, opts publishLogOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, scmBaseURL, func(ctx context.Context, c *client.Client) (any, error) {
		return executePublishLog(ctx, c, opts)
	})
}

func executeListProducts(ctx context.Context, c httpClient, opts listProductsOptions) (any, error) {
	if strings.TrimSpace(opts.StyleCode) == "" && strings.TrimSpace(opts.Title) == "" {
		return nil, fmt.Errorf("--style-code 与 --title 至少提供一个；复杂条件请用 capabilities/schema/web call")
	}
	if opts.PageNo <= 0 {
		opts.PageNo = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 10
	}
	body := map[string]any{
		"pageNo":   opts.PageNo,
		"pageSize": opts.PageSize,
	}
	if styleCode := strings.TrimSpace(opts.StyleCode); styleCode != "" {
		body["outerIds"] = []any{styleCode}
		body["outerIdBlur"] = 0
	}
	if title := strings.TrimSpace(opts.Title); title != "" {
		body["title"] = title
	}
	raw, _, err := c.PostJSON(ctx, pathItemBasePage, body)
	if err != nil {
		return nil, fmt.Errorf("查询 SCM 可铺货商品失败: %w", err)
	}
	data := dataMap(raw)
	records := mapList(data["records"])
	summaries := make([]map[string]any, 0, len(records))
	for _, record := range records {
		summaries = append(summaries, productSummary(record))
	}
	return map[string]any{
		"filters": map[string]any{
			"style_code": strings.TrimSpace(opts.StyleCode),
			"title":      strings.TrimSpace(opts.Title),
		},
		"page": map[string]any{
			"pageNo":   opts.PageNo,
			"pageSize": opts.PageSize,
			"total":    data["total"],
		},
		"records":   summaries,
		"raw_count": len(records),
		"endpoints": []string{
			pathItemBasePage,
		},
		"next": "确认商品 canPublishPlatform 包含目标平台后，可用 scm-item shops 查询可铺货店铺",
	}, nil
}

func executeShops(ctx context.Context, c httpClient, opts shopsOptions) (any, error) {
	platform := strings.TrimSpace(opts.Platform)
	if platform == "" {
		platform = "pdd"
	}
	if strings.TrimSpace(opts.StyleCode) == "" && strings.TrimSpace(opts.BaseItemID) == "" {
		return nil, fmt.Errorf("--style-code 与 --base-item-id 至少提供一个")
	}
	baseItemID := strings.TrimSpace(opts.BaseItemID)
	product := map[string]any{}
	if baseItemID == "" {
		found, err := findProductByStyleCode(ctx, c, opts.StyleCode)
		if err != nil {
			return nil, err
		}
		product = found
		baseItemID = stringField(found, "itemId")
		if baseItemID == "" {
			baseItemID = stringField(found, "baseItemId")
		}
		if baseItemID == "" {
			return nil, fmt.Errorf("商品 %s 缺少 baseItemId/itemId，无法查询店铺", opts.StyleCode)
		}
	}
	shops, err := queryShops(ctx, c, platform, baseItemID)
	if err != nil {
		return nil, err
	}
	filtered := filterShops(shops, opts)
	summaries := make([]map[string]any, 0, len(filtered))
	for _, shop := range filtered {
		summaries = append(summaries, shopAvailabilitySummary(shop))
	}
	return map[string]any{
		"platform":     platform,
		"style_code":   strings.TrimSpace(opts.StyleCode),
		"base_item_id": baseItemID,
		"product":      productSummary(product),
		"filters": map[string]any{
			"shop":    strings.TrimSpace(opts.Shop),
			"shop_id": opts.ShopID,
		},
		"shops": summaries,
		"count": len(summaries),
		"endpoints": []string{
			pathItemBasePage,
			pathShopAll,
		},
		"next": "选择 can_publish=true 的店铺后，可用 scm-item publish-pdd --shop-id 提交铺货预检",
	}, nil
}

func executePublishPDD(ctx context.Context, c httpClient, opts publishPDDOptions) (any, error) {
	if strings.TrimSpace(opts.StyleCode) == "" {
		return nil, fmt.Errorf("--style-code 不能为空")
	}
	if strings.TrimSpace(opts.Shop) == "" && opts.ShopID == 0 {
		return nil, fmt.Errorf("--shop 与 --shop-id 至少提供一个")
	}
	if opts.ShelfState == 0 {
		opts.ShelfState = defaultPddShelfState
	}

	product, err := findProductByStyleCode(ctx, c, opts.StyleCode)
	if err != nil {
		return nil, err
	}
	baseItemID := stringField(product, "itemId")
	if baseItemID == "" {
		baseItemID = stringField(product, "baseItemId")
	}
	if baseItemID == "" {
		return nil, fmt.Errorf("商品 %s 缺少 baseItemId/itemId，无法铺货", opts.StyleCode)
	}
	if !containsString(stringListField(product, "canPublishPlatformList"), "pdd") {
		return nil, fmt.Errorf("商品 %s 未包含 pdd 平台资料，无法铺货", opts.StyleCode)
	}

	shop, err := findPDDShop(ctx, c, baseItemID, opts)
	if err != nil {
		return nil, err
	}
	shopID, err := int64Field(shop, "id", "shopId")
	if err != nil {
		return nil, fmt.Errorf("店铺缺少 id/shopId: %w", err)
	}
	if disabledReason := shopDisabledReason(shop); disabledReason != "" {
		return nil, fmt.Errorf("店铺 %s 当前不可铺货：%s", displayShopName(shop), disabledReason)
	}

	flowNumber := generateFlowNumber(16)
	baseItemIDs := []any{baseItemID}
	shopIDs := []any{shopID}
	saveConfBody := map[string]any{
		"shopType":   "pdd",
		"shelfState": opts.ShelfState,
		"flowNumber": flowNumber,
	}
	if _, _, err := c.PostJSON(ctx, pathPddSaveTempConf, saveConfBody); err != nil {
		return nil, fmt.Errorf("保存 PDD 批量铺货临时配置失败: %w", err)
	}

	detailRaw, _, err := c.PostJSON(ctx, pathPddQueryDetail, map[string]any{
		"flowNumber":  flowNumber,
		"baseItemIds": baseItemIDs,
		"shopIds":     shopIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("查询 PDD 批量铺货详情失败: %w", err)
	}
	detailData := dataMap(detailRaw)
	batchList := mapList(detailData["batchDetailList"])
	completed, incomplete := splitCompleted(batchList)
	if len(completed) == 0 {
		return nil, fmt.Errorf("商品 %s 的 PDD 平台资料未完善，无法铺货", opts.StyleCode)
	}
	publishDetails := formatPDDPublishDetails(completed)
	publishBody := map[string]any{
		"shopIds":             shopIDs,
		"flowNumber":          flowNumber,
		"companyId":           valueOf(detailData, product, "companyId"),
		"isDistribution":      intFieldDefault(product, "isDistribution", 0),
		"batchItemDetailList": publishDetails,
	}

	plan := map[string]any{
		"platform":         "pdd",
		"submit":           opts.Submit,
		"style_code":       opts.StyleCode,
		"product":          productSummary(product),
		"shop":             shopSummary(shop),
		"flow_number":      flowNumber,
		"completed_count":  len(completed),
		"incomplete_count": len(incomplete),
		"endpoints": []string{
			pathItemBasePage,
			pathShopAll,
			pathPddSaveTempConf,
			pathPddQueryDetail,
			pathPddBatchPublish,
		},
		"publish_body": publishBody,
	}
	if !opts.Submit {
		plan["dry_run"] = true
		plan["next"] = "校验通过；如确认实际铺货，请加 --submit"
		return plan, nil
	}

	res, _, err := c.PostJSON(ctx, pathPddBatchPublish, publishBody)
	if err != nil {
		return nil, fmt.Errorf("提交 PDD 铺货任务失败: %w", err)
	}
	plan["dry_run"] = false
	plan["publish_result"] = res
	if opts.CheckLog {
		logs, err := queryPublishLogs(ctx, c, publishLogOptions{
			StyleCode: opts.StyleCode,
			Shop:      displayShopName(shop),
			ShopID:    shopID,
			StartTime: opts.LogStart,
			EndTime:   opts.LogEnd,
			PageNo:    opts.LogPage,
			PageSize:  opts.LogSize,
			Detail:    true,
		})
		if err != nil {
			plan["log_error"] = err.Error()
		} else {
			plan["publish_log"] = logs
		}
	}
	return plan, nil
}

func executePublishLog(ctx context.Context, c httpClient, opts publishLogOptions) (any, error) {
	if strings.TrimSpace(opts.StyleCode) == "" && strings.TrimSpace(opts.Shop) == "" && opts.ShopID == 0 {
		return nil, fmt.Errorf("--style-code、--shop、--shop-id 至少提供一个，避免日志范围过宽")
	}
	return queryPublishLogs(ctx, c, opts)
}

func queryPublishLogs(ctx context.Context, c httpClient, opts publishLogOptions) (map[string]any, error) {
	if opts.PageNo <= 0 {
		opts.PageNo = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = defaultLogPageSize
	}
	startTime, endTime := normalizeLogRange(opts.StartTime, opts.EndTime)
	raw, _, err := c.PostJSON(ctx, pathPublishLog, map[string]any{
		"operateType": "",
		"startTime":   startTime,
		"endTime":     endTime,
		"hasRetry":    0,
		"pageNo":      opts.PageNo,
		"pageSize":    opts.PageSize,
		"api_name":    "logging_publishLog",
	})
	if err != nil {
		return nil, fmt.Errorf("查询铺货日志失败: %w", err)
	}
	data := dataMap(raw)
	records := extractRecords(data)
	matches := filterPublishLogRecords(records, opts)
	out := map[string]any{
		"platform":   "pdd",
		"style_code": strings.TrimSpace(opts.StyleCode),
		"shop":       strings.TrimSpace(opts.Shop),
		"shop_id":    opts.ShopID,
		"time_range": map[string]any{
			"startTime": startTime,
			"endTime":   endTime,
		},
		"page": map[string]any{
			"pageNo":   opts.PageNo,
			"pageSize": opts.PageSize,
			"total":    data["total"],
		},
		"records": matches,
		"endpoints": []string{
			pathPublishLog,
		},
	}
	if opts.Detail {
		details := make([]map[string]any, 0, len(matches))
		for _, record := range matches {
			detail, err := fetchPublishLogDetail(ctx, c, record, opts)
			if err != nil {
				detail = map[string]any{
					"id":    stringField(record, "id"),
					"error": err.Error(),
				}
			}
			details = append(details, detail)
		}
		out["details"] = details
		out["endpoints"] = []string{
			pathPublishLog,
			pathPublishLogDetail,
			pathPublishLogByID,
		}
	}
	return out, nil
}

func fetchPublishLogDetail(ctx context.Context, c httpClient, record map[string]any, opts publishLogOptions) (map[string]any, error) {
	id := stringField(record, "id", "publishLogId")
	if id == "" {
		return nil, fmt.Errorf("日志记录缺少 id")
	}
	raw, _, err := c.GetQuery(ctx, pathPublishLogDetail, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("查询日志明细失败: %w", err)
	}
	rows := extractRecords(dataMap(raw))
	if len(rows) == 0 {
		rows = mapList(dataMap(raw)["list"])
	}
	filtered := filterPublishLogDetailRows(rows, opts)
	header := map[string]any{}
	if operationID := firstStringFieldFrom([]string{"operationLogId"}, append(filtered, record)...); operationID != "" {
		if byID, _, err := c.GetQuery(ctx, pathPublishLogByID, map[string]any{"id": operationID}); err == nil {
			header = dataMap(byID)
		}
	}
	return map[string]any{
		"id":      id,
		"record":  record,
		"header":  header,
		"rows":    filtered,
		"summary": summarizeLogRows(filtered),
	}, nil
}

func findProductByStyleCode(ctx context.Context, c httpClient, styleCode string) (map[string]any, error) {
	raw, _, err := c.PostJSON(ctx, pathItemBasePage, map[string]any{
		"pageNo":      1,
		"pageSize":    10,
		"outerIds":    []any{styleCode},
		"outerIdBlur": 0,
	})
	if err != nil {
		return nil, fmt.Errorf("查询 SCM 商品失败: %w", err)
	}
	records := mapList(dataMap(raw)["records"])
	matches := make([]map[string]any, 0, len(records))
	for _, record := range records {
		if stringField(record, "outerId") == styleCode {
			matches = append(matches, record)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("未找到款式编码 %s 对应的 SCM 商品", styleCode)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("款式编码 %s 匹配到 %d 个商品，请先保证唯一", styleCode, len(matches))
	}
	return matches[0], nil
}

func queryShops(ctx context.Context, c httpClient, platform, baseItemID string) ([]map[string]any, error) {
	raw, _, err := c.GetQuery(ctx, pathShopAll, map[string]any{
		"source":     platform,
		"subSource":  "",
		"baseItemId": baseItemID,
		"channel":    "",
	})
	if err != nil {
		return nil, fmt.Errorf("查询 %s 授权店铺失败: %w", platform, err)
	}
	shops := mapList(dataMap(raw)["list"])
	if len(shops) == 0 {
		return nil, fmt.Errorf("没有可用的 %s 授权店铺", platform)
	}
	return shops, nil
}

func findPDDShop(ctx context.Context, c httpClient, baseItemID string, opts publishPDDOptions) (map[string]any, error) {
	shops, err := queryShops(ctx, c, "pdd", baseItemID)
	if err != nil {
		return nil, err
	}
	if opts.ShopID != 0 {
		for _, shop := range shops {
			id, _ := int64Field(shop, "id", "shopId")
			if id == opts.ShopID {
				return shop, nil
			}
		}
		return nil, fmt.Errorf("未找到 shopId=%d 的 PDD 店铺", opts.ShopID)
	}

	keyword := strings.TrimSpace(opts.Shop)
	exact := make([]map[string]any, 0, 1)
	fuzzy := make([]map[string]any, 0, 1)
	for _, shop := range shops {
		shortTitle := stringField(shop, "shortTitle")
		title := stringField(shop, "title")
		if shortTitle == keyword || title == keyword {
			exact = append(exact, shop)
		} else if strings.Contains(shortTitle, keyword) || strings.Contains(title, keyword) {
			fuzzy = append(fuzzy, shop)
		}
	}
	if len(exact) == 1 {
		return exact[0], nil
	}
	if len(exact) > 1 {
		return nil, fmt.Errorf("店铺 %s 精确匹配到 %d 个结果，请改用 --shop-id", keyword, len(exact))
	}
	if len(fuzzy) == 1 {
		return fuzzy[0], nil
	}
	if len(fuzzy) > 1 {
		return nil, fmt.Errorf("店铺 %s 模糊匹配到 %d 个结果，请改用 --shop-id", keyword, len(fuzzy))
	}
	return nil, fmt.Errorf("未找到 PDD 店铺 %s", keyword)
}

func filterShops(shops []map[string]any, opts shopsOptions) []map[string]any {
	out := make([]map[string]any, 0, len(shops))
	keyword := strings.TrimSpace(opts.Shop)
	for _, shop := range shops {
		if opts.ShopID != 0 {
			id, _ := int64Field(shop, "id", "shopId")
			if id != opts.ShopID {
				continue
			}
		}
		if keyword != "" {
			shortTitle := stringField(shop, "shortTitle")
			title := stringField(shop, "title")
			shopName := stringField(shop, "shopName")
			if shortTitle != keyword && title != keyword && shopName != keyword &&
				!strings.Contains(shortTitle, keyword) &&
				!strings.Contains(title, keyword) &&
				!strings.Contains(shopName, keyword) {
				continue
			}
		}
		out = append(out, shop)
	}
	return out
}

func splitCompleted(items []map[string]any) ([]map[string]any, []map[string]any) {
	completed := make([]map[string]any, 0, len(items))
	incomplete := make([]map[string]any, 0)
	for _, item := range items {
		if intFieldDefault(item, "informationLack", 0) == 1 {
			completed = append(completed, item)
		} else {
			incomplete = append(incomplete, item)
		}
	}
	return completed, incomplete
}

func normalizeLogRange(start, end string) (string, string) {
	const layout = "2006-01-02 15:04:05"
	now := time.Now()
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if end == "" {
		end = now.Format(layout)
	}
	if start == "" {
		start = now.AddDate(0, 0, -30).Format(layout)
	}
	return start, end
}

func extractRecords(data map[string]any) []map[string]any {
	for _, key := range []string{"records", "list", "rows", "data"} {
		if records := mapList(data[key]); len(records) > 0 {
			return records
		}
	}
	return nil
}

func filterPublishLogRecords(records []map[string]any, opts publishLogOptions) []map[string]any {
	out := make([]map[string]any, 0, len(records))
	for _, record := range records {
		if !recordLooksLikePDD(record) {
			continue
		}
		if opts.ShopID != 0 {
			id, err := int64Field(record, "shopId", "shopID", "shop_id")
			if err == nil && id != opts.ShopID {
				continue
			}
		}
		if shop := strings.TrimSpace(opts.Shop); shop != "" && !recordMatchesShop(record, shop) {
			continue
		}
		if style := strings.TrimSpace(opts.StyleCode); style != "" {
			// The list API may not always include item rows; keep likely task records and filter strictly in detail.
			if outer := stringField(record, "outerId", "styleCode"); outer != "" && outer != style {
				continue
			}
		}
		out = append(out, record)
	}
	return out
}

func filterPublishLogDetailRows(rows []map[string]any, opts publishLogOptions) []map[string]any {
	style := strings.TrimSpace(opts.StyleCode)
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if style != "" && stringField(row, "outerId", "styleCode") != style {
			continue
		}
		out = append(out, row)
	}
	return out
}

func recordLooksLikePDD(record map[string]any) bool {
	seenPlatform := false
	for _, key := range []string{"platformType", "platform", "source", "shopType"} {
		value := stringField(record, key)
		if value == "" {
			continue
		}
		seenPlatform = true
		if strings.EqualFold(value, "pdd") {
			return true
		}
	}
	// Older log rows may omit platform in the list but still contain PDD shop fields.
	if stringField(record, "pddGoodsId", "pddProductId") != "" {
		return true
	}
	return !seenPlatform
}

func recordMatchesShop(record map[string]any, shop string) bool {
	fields := []string{
		"shopName",
		"shopTitle",
		"shortTitle",
		"title",
		"storeName",
	}
	for _, key := range fields {
		value := stringField(record, key)
		if value == shop || strings.Contains(value, shop) {
			return true
		}
	}
	return false
}

func summarizeLogRows(rows []map[string]any) map[string]any {
	summary := map[string]any{
		"success":   0,
		"failed":    0,
		"executing": 0,
		"queued":    0,
		"other":     0,
		"failures":  []map[string]any{},
	}
	failures := make([]map[string]any, 0)
	for _, row := range rows {
		status := intFieldDefault(row, "status", 999)
		switch {
		case status == 1:
			summary["success"] = summary["success"].(int) + 1
		case status == 0 || status == 4:
			summary["failed"] = summary["failed"].(int) + 1
			failures = append(failures, map[string]any{
				"outerId":      stringField(row, "outerId", "styleCode"),
				"status":       row["status"],
				"errorMessage": stringField(row, "errorMessage", "failReason", "reason"),
			})
		case status == -1 || status == 3:
			summary["executing"] = summary["executing"].(int) + 1
		case status == -3 || status == -2:
			summary["queued"] = summary["queued"].(int) + 1
		default:
			summary["other"] = summary["other"].(int) + 1
		}
	}
	summary["failures"] = failures
	return summary
}

func formatPDDPublishDetails(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		cp := cloneMap(item)
		if v, ok := cp["shopBrandConf"]; ok && v != nil {
			switch vv := v.(type) {
			case string:
				cp["shopBrandConf"] = vv
			default:
				b, err := json.Marshal(vv)
				if err == nil {
					cp["shopBrandConf"] = string(b)
				}
			}
		} else {
			cp["shopBrandConf"] = nil
		}
		if skuList := mapList(cp["skuDetailList"]); len(skuList) > 0 {
			next := make([]map[string]any, 0, len(skuList))
			for _, sku := range skuList {
				s := cloneMap(sku)
				delete(s, "skuSpecifications")
				next = append(next, s)
			}
			cp["skuDetailList"] = next
		}
		out = append(out, cp)
	}
	return out
}

func shopDisabledReason(shop map[string]any) string {
	if s := stringField(shop, "abnormalState"); s != "" {
		return "店铺异常: " + s
	}
	if b, ok := boolField(shop, "uncheck"); ok && b {
		return "平台发布上限限制"
	}
	if b, ok := boolField(shop, "xyIsPro"); ok && !b {
		return "店铺状态不满足发布要求"
	}
	if b, ok := boolField(shop, "isCheckFreightTemplate"); ok && !b {
		return "未配置运费模板"
	}
	state := intFieldDefault(shop, "state", 0)
	if state != 0 && state != 2 && state != 3 {
		return fmt.Sprintf("授权状态不可用 state=%d", state)
	}
	if deadline := stringField(shop, "deadline"); deadline != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", deadline, time.Local); err == nil && time.Now().After(t) {
			return "授权已过期"
		}
	}
	return ""
}

func dataMap(raw any) map[string]any {
	if m, ok := raw.(map[string]any); ok {
		if data, ok := m["data"].(map[string]any); ok {
			return data
		}
		return m
	}
	return map[string]any{}
}

func mapList(v any) []map[string]any {
	switch list := v.(type) {
	case []map[string]any:
		return list
	case []any:
		out := make([]map[string]any, 0, len(list))
		for _, item := range list {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func stringField(m map[string]any, keys ...string) string {
	for _, key := range keys {
		switch v := m[key].(type) {
		case string:
			return v
		case fmt.Stringer:
			return v.String()
		case float64:
			if v == float64(int64(v)) {
				return strconv.FormatInt(int64(v), 10)
			}
		case int:
			return strconv.Itoa(v)
		case int64:
			return strconv.FormatInt(v, 10)
		}
	}
	return ""
}

func int64Field(m map[string]any, keys ...string) (int64, error) {
	for _, key := range keys {
		switch v := m[key].(type) {
		case float64:
			return int64(v), nil
		case int:
			return int64(v), nil
		case int64:
			return v, nil
		case json.Number:
			return v.Int64()
		case string:
			if v == "" {
				continue
			}
			return strconv.ParseInt(v, 10, 64)
		}
	}
	return 0, errors.New("字段不存在")
}

func intFieldDefault(m map[string]any, key string, def int) int {
	n, err := int64Field(m, key)
	if err != nil {
		return def
	}
	return int(n)
}

func boolField(m map[string]any, key string) (bool, bool) {
	v, ok := m[key]
	if !ok {
		return false, false
	}
	switch b := v.(type) {
	case bool:
		return b, true
	case string:
		parsed, err := strconv.ParseBool(b)
		return parsed, err == nil
	default:
		return false, false
	}
}

func stringListField(m map[string]any, key string) []string {
	switch v := m[key].(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		if v == "" {
			return nil
		}
		var arr []string
		if json.Unmarshal([]byte(v), &arr) == nil {
			return arr
		}
		return strings.Split(v, ",")
	default:
		return nil
	}
}

func containsString(list []string, want string) bool {
	for _, item := range list {
		if strings.TrimSpace(item) == want {
			return true
		}
	}
	return false
}

func valueOf(primary, fallback map[string]any, key string) any {
	if v, ok := primary[key]; ok && v != nil {
		return v
	}
	return fallback[key]
}

func firstStringFieldFrom(keys []string, maps ...map[string]any) string {
	for _, m := range maps {
		if m == nil {
			continue
		}
		if v := stringField(m, keys...); v != "" {
			return v
		}
	}
	return ""
}

func productSummary(product map[string]any) map[string]any {
	return map[string]any{
		"baseItemId":         stringField(product, "baseItemId"),
		"itemId":             stringField(product, "itemId"),
		"outerId":            stringField(product, "outerId"),
		"title":              stringField(product, "title"),
		"canPublishPlatform": stringListField(product, "canPublishPlatformList"),
		"publishShopList":    product["publishShopList"],
	}
}

func shopSummary(shop map[string]any) map[string]any {
	id, _ := int64Field(shop, "id", "shopId")
	return map[string]any{
		"id":         id,
		"shortTitle": stringField(shop, "shortTitle"),
		"title":      stringField(shop, "title"),
		"name":       displayShopName(shop),
		"source":     stringField(shop, "source"),
		"state":      shop["state"],
	}
}

func shopAvailabilitySummary(shop map[string]any) map[string]any {
	out := shopSummary(shop)
	reason := shopDisabledReason(shop)
	out["can_publish"] = reason == ""
	out["disabled_reason"] = reason
	out["deadline"] = stringField(shop, "deadline")
	out["abnormalState"] = shop["abnormalState"]
	out["uncheck"] = shop["uncheck"]
	out["xyIsPro"] = shop["xyIsPro"]
	out["isCheckFreightTemplate"] = shop["isCheckFreightTemplate"]
	return out
}

func displayShopName(shop map[string]any) string {
	if s := stringField(shop, "shortTitle"); s != "" {
		return s
	}
	if s := stringField(shop, "title"); s != "" {
		return s
	}
	if s := stringField(shop, "shopName"); s != "" {
		return s
	}
	return stringField(shop, "id", "shopId")
}

func generateFlowNumber(length int) string {
	if length <= 0 {
		length = 16
	}
	now := strconv.FormatInt(time.Now().UnixMilli(), 10)
	var b strings.Builder
	for i := 0; i < length; i++ {
		if i < len(now) {
			b.WriteByte(now[i])
			continue
		}
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			b.WriteByte('0')
			continue
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	return b.String()
}
