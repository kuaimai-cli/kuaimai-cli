package scm

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kuaimai-cli/kuaimai-cli/internal/client"
	"github.com/kuaimai-cli/kuaimai-cli/internal/cmdutil"
	"github.com/kuaimai-cli/kuaimai-cli/shortcuts/common"
	"github.com/spf13/cobra"
)

const (
	shortcutName = "scm-item"

	pathItemBasePage       = "/item/base/page.json"
	pathItemBaseDetail     = "/item/base/detail.json"
	pathItemBaseEdit       = "/item/base/edit.json"
	pathQueryErpItems      = "/item/base/queryErpItems.json"
	pathShopAll            = "/shop/allShop.json"
	pathPddCarouselVideo   = "/pdd/getCarouselVideoInfo.json"
	pathPreCheckPrice      = "/ltsTask/preCheckControllerPrice.json"
	pathPddAuthorize       = "/pdd/authorize/authStatus.json"
	pathSecConfirm         = "/item/base/secConfirmationItemV2.json"
	pathStorageTask        = "/taskScheduling/storageTask.json"
	pathQueryTaskSpeed     = "/taskScheduling/queryTaskSpeed.json"
	pathPublishLog         = "/logging/publishLog.json"
	pathPublishLogDetail   = "/logging/publishLogDetail.json"
	pathPublishLogByID     = "/logging/publishLogById.json"
	defaultShelfState      = 1
	defaultPublishType     = "PUBLISH_ITEM"
	defaultPublishEntrance = 0
	defaultLogPageSize     = 10
	apiNameItemBasePage    = "item_base_page"

	interfaceItemBasePage     = "item.base.page"
	interfaceItemBaseDetail   = "item.base.detail"
	interfaceItemBaseEdit     = "item.base.edit"
	interfaceQueryErpItems    = "item.base.queryErpItems"
	interfaceShopAll          = "shop.allShop"
	interfacePddCarouselVideo = "pdd.getCarouselVideoInfo"
	interfacePreCheckPrice    = "ltsTask.preCheckControllerPrice"
	interfacePddAuthorize     = "pdd.authorize.authStatus"
	interfaceSecConfirm       = "item.base.secConfirmationItemV2"
	interfaceStorageTask      = "taskScheduling.storageTask"
	interfaceQueryTask        = "taskScheduling.queryTaskSpeed"
	interfacePublishLog       = "logging.publishLog"
	interfacePublishLogDetail = "logging.publishLogDetail"
	interfacePublishLogByID   = "logging.publishLogById"
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

type publishPlatformOptions struct {
	Platform   string
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
	Platform  string
	StyleCode string
	Shop      string
	ShopID    int64
	StartTime string
	EndTime   string
	PageNo    int
	PageSize  int
	Detail    bool
}

type updateTitleOptions struct {
	ID         int64
	StyleCode  string
	Title      string
	Submit     bool
	CheckSync  bool
	SkipAddERP bool
}

type pddPrimitiveOptions struct {
	StyleCode      string
	BaseItemID     string
	Shop           string
	ShopID         int64
	PlatformShopID int64
	BatchTaskID    string
	Submit         bool
}

type platformPrimitiveOptions struct {
	Platform    string
	StyleCode   string
	BaseItemID  string
	Shop        string
	ShopID      int64
	Submit      bool
	BatchTaskID string
}

type publishTarget struct {
	Product        map[string]any
	Shop           map[string]any
	BaseItemID     string
	ShopID         int64
	PlatformShopID int64
	StyleCode      string
	Platform       string
}

type platformConfig struct {
	Key                string
	Label              string
	ShopSource         string
	ShopSubSource      string
	SecConfirmPlatform string
	NeedsPricePrecheck bool
	NeedsSecConfirm    bool
	NeedsPDDVideo      bool
	NeedsPDDAuthorize  bool
	SubmitPath         string
}

// Register attaches SCM item shortcuts to root.
func Register(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "scm-item",
		Short: "SCM 可铺货商品命令（erp-scm，可上货/铺货到店铺）",
	}
	cmd.AddCommand(listProductsCmd())
	cmd.AddCommand(updateTitleCmd())
	cmd.AddCommand(shopsCmd())
	cmd.AddCommand(publishPlatformCmd())
	cmd.AddCommand(publishPDDCmd())
	cmd.AddCommand(platformPrimitiveCmds()...)
	cmd.AddCommand(pddPrimitiveCmds()...)
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

func updateTitleCmd() *cobra.Command {
	var opts updateTitleOptions
	c := &cobra.Command{
		Use:   "update-title",
		Short: "修改 SCM 商品名称（先读取详情，默认只生成保存请求体）",
		Example: `  kuaimai-cli scm-item update-title --style-code '<款式编码>' --title '<新商品名称>' --output json
  kuaimai-cli scm-item update-title --id 123456 --title '<新商品名称>' --submit --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateTitle(context.Background(), opts)
		},
	}
	c.Flags().Int64Var(&opts.ID, "id", 0, "SCM 商品自增 id（与 --style-code 二选一）")
	c.Flags().StringVar(&opts.StyleCode, "style-code", "", "款式编码 outerId（与 --id 二选一，会先从列表唯一定位 id）")
	c.Flags().StringVar(&opts.Title, "title", "", "新商品名称（必填）")
	c.Flags().BoolVar(&opts.Submit, "submit", false, "实际调用 /item/base/queryErpItems 与 /item/base/edit 保存修改")
	c.Flags().BoolVar(&opts.CheckSync, "check-open-sync", false, "保存体 checkOpenSync，默认 false")
	c.Flags().BoolVar(&opts.SkipAddERP, "skip-add-item-to-erp", true, "保存体 skipAddItemToErp，默认 true")
	_ = c.MarkFlagRequired("title")
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

func publishPlatformCmd() *cobra.Command {
	var opts publishPlatformOptions
	c := &cobra.Command{
		Use:   "publish",
		Short: "将 SCM 商品铺货到指定平台店铺（默认只预检，不提交）",
		Example: `  kuaimai-cli scm-item publish --platform tb --style-code '<款式编码>' --shop '<淘宝店铺名>' --output json
  kuaimai-cli scm-item publish --platform fxg --style-code '<款式编码>' --shop-id 123 --submit --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPublishPlatform(context.Background(), opts)
		},
	}
	addPublishPlatformFlags(c, &opts)
	_ = c.MarkFlagRequired("platform")
	_ = c.MarkFlagRequired("style-code")
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
	c.Flags().IntVar(&opts.ShelfState, "shelf-state", defaultShelfState, "上架设置 shelfState：1 立即上架，2 放入草稿箱")
	c.Flags().BoolVar(&opts.Submit, "submit", false, "通过校验后实际提交 PDD 铺货任务")
	c.Flags().BoolVar(&opts.CheckLog, "check-log", false, "提交后查询最近铺货日志与失败原因")
	c.Flags().StringVar(&opts.LogStart, "log-start-time", "", "日志查询开始时间 yyyy-MM-dd HH:mm:ss（默认近 30 天）")
	c.Flags().StringVar(&opts.LogEnd, "log-end-time", "", "日志查询结束时间 yyyy-MM-dd HH:mm:ss（默认当前时间）")
	c.Flags().IntVar(&opts.LogPage, "log-page", 1, "日志查询页码")
	c.Flags().IntVar(&opts.LogSize, "log-page-size", defaultLogPageSize, "日志查询每页条数")
	_ = c.MarkFlagRequired("style-code")
	return c
}

func addPublishPlatformFlags(c *cobra.Command, opts *publishPlatformOptions) {
	c.Flags().StringVar(&opts.Platform, "platform", "", "平台 key，例如 pdd、tb、fxg、kuaishou、jd、1688、tm、yz、wxsph、wd、xhs、xy、pddtemu、shein、fxg_gx")
	c.Flags().StringVar(&opts.StyleCode, "style-code", "", "款式编码 outerId（必填）")
	c.Flags().StringVar(&opts.Shop, "shop", "", "店铺简称/名称（与 --shop-id 二选一）")
	c.Flags().Int64Var(&opts.ShopID, "shop-id", 0, "店铺 ID（与 --shop 二选一）")
	c.Flags().IntVar(&opts.ShelfState, "shelf-state", defaultShelfState, "上架设置 shelfState：1 立即上架，2 放入草稿箱")
	c.Flags().BoolVar(&opts.Submit, "submit", false, "通过校验后实际提交铺货任务")
	c.Flags().BoolVar(&opts.CheckLog, "check-log", false, "提交后查询最近铺货日志与失败原因")
	c.Flags().StringVar(&opts.LogStart, "log-start-time", "", "日志查询开始时间 yyyy-MM-dd HH:mm:ss（默认近 30 天）")
	c.Flags().StringVar(&opts.LogEnd, "log-end-time", "", "日志查询结束时间 yyyy-MM-dd HH:mm:ss（默认当前时间）")
	c.Flags().IntVar(&opts.LogPage, "log-page", 1, "日志查询页码")
	c.Flags().IntVar(&opts.LogSize, "log-page-size", defaultLogPageSize, "日志查询每页条数")
}

func platformPrimitiveCmds() []*cobra.Command {
	return []*cobra.Command{
		platformPricePrecheckCmd(),
		platformSecConfirmCmd(),
		platformStorageTaskCmd(),
		platformTaskSpeedCmd(),
	}
}

func platformPricePrecheckCmd() *cobra.Command {
	var opts platformPrimitiveOptions
	c := &cobra.Command{
		Use:     "platform-price-precheck",
		Short:   "通用底层接口：平台铺货控价预检",
		Example: `  kuaimai-cli scm-item platform-price-precheck --platform tb --style-code '<款式编码>' --shop-id 123 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlatformPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePlatformPricePrecheck(ctx, c, opts)
			})
		},
	}
	addPlatformTargetFlags(c, &opts)
	return c
}

func platformSecConfirmCmd() *cobra.Command {
	var opts platformPrimitiveOptions
	c := &cobra.Command{
		Use:     "platform-sec-confirm",
		Short:   "通用底层接口：平台铺货前二次确认",
		Example: `  kuaimai-cli scm-item platform-sec-confirm --platform fxg --style-code '<款式编码>' --shop-id 123 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlatformPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePlatformSecConfirm(ctx, c, opts)
			})
		},
	}
	addPlatformTargetFlags(c, &opts)
	return c
}

func platformStorageTaskCmd() *cobra.Command {
	var opts platformPrimitiveOptions
	c := &cobra.Command{
		Use:   "platform-storage-task",
		Short: "通用底层接口：提交平台铺货任务（默认只输出 body）",
		Example: `  kuaimai-cli scm-item platform-storage-task --platform kuaishou --style-code '<款式编码>' --shop-id 123 --output json
  kuaimai-cli scm-item platform-storage-task --platform kuaishou --style-code '<款式编码>' --shop-id 123 --submit --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlatformPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePlatformStorageTask(ctx, c, opts)
			})
		},
	}
	addPlatformTargetFlags(c, &opts)
	c.Flags().BoolVar(&opts.Submit, "submit", false, "实际调用 taskScheduling.storageTask 提交铺货任务")
	return c
}

func platformTaskSpeedCmd() *cobra.Command {
	var opts platformPrimitiveOptions
	c := &cobra.Command{
		Use:     "platform-task-speed",
		Short:   "通用底层接口：查询平台铺货任务进度",
		Example: `  kuaimai-cli scm-item platform-task-speed --batch-task-id '<batchTaskId>' --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlatformPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePlatformTaskSpeed(ctx, c, opts)
			})
		},
	}
	c.Flags().StringVar(&opts.BatchTaskID, "batch-task-id", "", "taskScheduling.storageTask 返回的批次任务 ID")
	_ = c.MarkFlagRequired("batch-task-id")
	return c
}

func addPlatformTargetFlags(c *cobra.Command, opts *platformPrimitiveOptions) {
	c.Flags().StringVar(&opts.Platform, "platform", "", "平台 key，例如 pdd、tb、fxg、kuaishou、fxg_gx")
	c.Flags().StringVar(&opts.StyleCode, "style-code", "", "款式编码 outerId（与 --base-item-id 二选一）")
	c.Flags().StringVar(&opts.BaseItemID, "base-item-id", "", "SCM 商品 baseItemId/itemId（与 --style-code 二选一）")
	c.Flags().StringVar(&opts.Shop, "shop", "", "店铺简称/名称（与 --shop-id 二选一）")
	c.Flags().Int64Var(&opts.ShopID, "shop-id", 0, "店铺内部 ID（shop/allShop 返回的 id）")
	_ = c.MarkFlagRequired("platform")
}

func pddPrimitiveCmds() []*cobra.Command {
	return []*cobra.Command{
		pddVideoInfoCmd(),
		pddPricePrecheckCmd(),
		pddAuthStatusCmd(),
		pddSecConfirmCmd(),
		pddStorageTaskCmd(),
		pddTaskSpeedCmd(),
	}
}

func pddVideoInfoCmd() *cobra.Command {
	var opts pddPrimitiveOptions
	c := &cobra.Command{
		Use:   "pdd-video-info",
		Short: "PDD 底层接口：查询轮播视频信息",
		Example: `  kuaimai-cli scm-item pdd-video-info --style-code '<款式编码>' --output json
  kuaimai-cli scm-item pdd-video-info --base-item-id '<baseItemId>' --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPDDPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePDDVideoInfo(ctx, c, opts)
			})
		},
	}
	addPDDItemFlags(c, &opts)
	return c
}

func pddPricePrecheckCmd() *cobra.Command {
	var opts pddPrimitiveOptions
	c := &cobra.Command{
		Use:     "pdd-price-precheck",
		Short:   "PDD 底层接口：控价预检",
		Example: `  kuaimai-cli scm-item pdd-price-precheck --style-code '<款式编码>' --shop-id 123456 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPDDPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePDDPricePrecheck(ctx, c, opts)
			})
		},
	}
	addPDDTargetFlags(c, &opts)
	return c
}

func pddAuthStatusCmd() *cobra.Command {
	var opts pddPrimitiveOptions
	c := &cobra.Command{
		Use:   "pdd-auth-status",
		Short: "PDD 底层接口：授权状态校验",
		Example: `  kuaimai-cli scm-item pdd-auth-status --platform-shop-id 849217672 --output json
  kuaimai-cli scm-item pdd-auth-status --style-code '<款式编码>' --shop-id 123456 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPDDPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePDDAuthStatus(ctx, c, opts)
			})
		},
	}
	c.Flags().Int64Var(&opts.PlatformShopID, "platform-shop-id", 0, "PDD 平台店铺 ID（shop/allShop 返回的 taobaoId）")
	addPDDTargetFlags(c, &opts)
	return c
}

func pddSecConfirmCmd() *cobra.Command {
	var opts pddPrimitiveOptions
	c := &cobra.Command{
		Use:     "pdd-sec-confirm",
		Short:   "PDD 底层接口：发布前二次确认",
		Example: `  kuaimai-cli scm-item pdd-sec-confirm --style-code '<款式编码>' --shop-id 123456 --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPDDPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePDDSecConfirm(ctx, c, opts)
			})
		},
	}
	addPDDTargetFlags(c, &opts)
	return c
}

func pddStorageTaskCmd() *cobra.Command {
	var opts pddPrimitiveOptions
	c := &cobra.Command{
		Use:   "pdd-storage-task",
		Short: "PDD 底层接口：提交铺货任务（默认只输出 body）",
		Example: `  kuaimai-cli scm-item pdd-storage-task --style-code '<款式编码>' --shop-id 123456 --output json
  kuaimai-cli scm-item pdd-storage-task --style-code '<款式编码>' --shop-id 123456 --submit --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPDDPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePDDStorageTask(ctx, c, opts)
			})
		},
	}
	addPDDTargetFlags(c, &opts)
	c.Flags().BoolVar(&opts.Submit, "submit", false, "实际调用 taskScheduling.storageTask 提交铺货任务")
	return c
}

func pddTaskSpeedCmd() *cobra.Command {
	var opts pddPrimitiveOptions
	c := &cobra.Command{
		Use:     "pdd-task-speed",
		Short:   "PDD 底层接口：查询铺货任务进度",
		Example: `  kuaimai-cli scm-item pdd-task-speed --batch-task-id '<batchTaskId>' --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPDDPrimitive(context.Background(), func(ctx context.Context, c *client.Client) (any, error) {
				return executePDDTaskSpeed(ctx, c, opts)
			})
		},
	}
	c.Flags().StringVar(&opts.BatchTaskID, "batch-task-id", "", "taskScheduling.storageTask 返回的批次任务 ID")
	_ = c.MarkFlagRequired("batch-task-id")
	return c
}

func addPDDItemFlags(c *cobra.Command, opts *pddPrimitiveOptions) {
	c.Flags().StringVar(&opts.StyleCode, "style-code", "", "款式编码 outerId（与 --base-item-id 二选一）")
	c.Flags().StringVar(&opts.BaseItemID, "base-item-id", "", "SCM 商品 baseItemId/itemId（与 --style-code 二选一）")
}

func addPDDTargetFlags(c *cobra.Command, opts *pddPrimitiveOptions) {
	addPDDItemFlags(c, opts)
	c.Flags().StringVar(&opts.Shop, "shop", "", "PDD 店铺简称/名称（与 --shop-id 二选一）")
	c.Flags().Int64Var(&opts.ShopID, "shop-id", 0, "PDD 店铺内部 ID（shop/allShop 返回的 id）")
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
	return runPublishPlatform(ctx, publishPlatformOptions{
		Platform:   "pdd",
		StyleCode:  opts.StyleCode,
		Shop:       opts.Shop,
		ShopID:     opts.ShopID,
		ShelfState: opts.ShelfState,
		Submit:     opts.Submit,
		CheckLog:   opts.CheckLog,
		LogStart:   opts.LogStart,
		LogEnd:     opts.LogEnd,
		LogPage:    opts.LogPage,
		LogSize:    opts.LogSize,
		Detail:     opts.Detail,
	})
}

func runPublishPlatform(ctx context.Context, opts publishPlatformOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, f.Config.ShortcutAPIURL(shortcutName), func(ctx context.Context, c *client.Client) (any, error) {
		return executePublishPlatform(ctx, c, opts)
	})
}

func runListProducts(ctx context.Context, opts listProductsOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, f.Config.ShortcutAPIURL(shortcutName), func(ctx context.Context, c *client.Client) (any, error) {
		return executeListProducts(ctx, c, opts)
	})
}

func runShops(ctx context.Context, opts shopsOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, f.Config.ShortcutAPIURL(shortcutName), func(ctx context.Context, c *client.Client) (any, error) {
		return executeShops(ctx, c, opts)
	})
}

func runUpdateTitle(ctx context.Context, opts updateTitleOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, f.Config.ShortcutAPIURL(shortcutName), func(ctx context.Context, c *client.Client) (any, error) {
		return executeUpdateTitle(ctx, c, opts)
	})
}

func runPublishLog(ctx context.Context, opts publishLogOptions) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, f.Config.ShortcutAPIURL(shortcutName), func(ctx context.Context, c *client.Client) (any, error) {
		return executePublishLog(ctx, c, opts)
	})
}

func runPDDPrimitive(ctx context.Context, fn func(context.Context, *client.Client) (any, error)) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, f.Config.ShortcutAPIURL(shortcutName), fn)
}

func runPlatformPrimitive(ctx context.Context, fn func(context.Context, *client.Client) (any, error)) error {
	f, err := cmdutil.NewFactory()
	if err != nil {
		return err
	}
	r := common.NewRunner(f)
	return r.ExecuteWithBase(ctx, f.Config.ShortcutAPIURL(shortcutName), fn)
}

func executeListProducts(ctx context.Context, c httpClient, opts listProductsOptions) (any, error) {
	if opts.PageNo <= 0 {
		opts.PageNo = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 10
	}
	body := map[string]any{
		"pageNo":          opts.PageNo,
		"pageSize":        opts.PageSize,
		"leafCategories":  []any{},
		"cgSupplierIds":   []any{},
		"platformItemIds": []any{},
		"companyNames":    []any{},
		"shopNames":       []any{},
		"api_name":        apiNameItemBasePage,
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
		"interfaces": []map[string]any{
			interfaceInfo(interfaceItemBasePage, "POST", pathItemBasePage),
		},
		"endpoints": []string{
			pathItemBasePage,
		},
		"next": "确认商品 canPublishPlatform 包含目标平台后，可用 scm-item shops 查询可铺货店铺",
	}, nil
}

func executeUpdateTitle(ctx context.Context, c httpClient, opts updateTitleOptions) (any, error) {
	title := strings.TrimSpace(opts.Title)
	if title == "" {
		return nil, fmt.Errorf("--title 不能为空")
	}
	if opts.ID == 0 && strings.TrimSpace(opts.StyleCode) == "" {
		return nil, fmt.Errorf("--id 与 --style-code 至少提供一个")
	}
	id := opts.ID
	product := map[string]any{}
	if id == 0 {
		found, err := findProductByStyleCode(ctx, c, opts.StyleCode)
		if err != nil {
			return nil, err
		}
		product = found
		var parseErr error
		id, parseErr = int64Field(found, "id")
		if parseErr != nil || id == 0 {
			return nil, fmt.Errorf("商品 %s 缺少列表 id，无法调用 /item/base/detail", opts.StyleCode)
		}
	}

	detail, rawDetail, err := fetchBaseItemDetail(ctx, c, id)
	if err != nil {
		return nil, err
	}
	oldTitle := stringField(detail, "title")
	item := prepareSCMEditItem(detail, title)
	saveBody := scmEditBody(item, opts)
	queryBody := cloneMap(item)
	queryBody["api_name"] = "item_base_queryErpItems"

	interfaces := []map[string]any{
		interfaceInfo(interfaceItemBaseDetail, "GET", pathItemBaseDetail),
		interfaceInfo(interfaceQueryErpItems, "POST", pathQueryErpItems),
		interfaceInfo(interfaceItemBaseEdit, "POST", pathItemBaseEdit),
	}
	endpoints := []string{pathItemBaseDetail, pathQueryErpItems, pathItemBaseEdit}
	out := map[string]any{
		"id":         id,
		"style_code": strings.TrimSpace(opts.StyleCode),
		"old_title":  oldTitle,
		"new_title":  title,
		"submit":     opts.Submit,
		"dry_run":    !opts.Submit,
		"product":    productSummary(product),
		"detail_raw": rawDetail,
		"query_body": queryBody,
		"save_body":  saveBody,
		"interfaces": interfaces,
		"endpoints":  endpoints,
	}
	if !opts.Submit {
		out["next"] = "确认 save_body 只变更 title 后，加 --submit 实际保存"
		return out, nil
	}

	queryRaw, _, err := c.PostJSON(ctx, pathQueryErpItems, queryBody)
	if err != nil {
		return nil, fmt.Errorf("SCM 保存前 ERP 同步校验失败: %w", err)
	}
	saveRaw, _, err := c.PostJSON(ctx, pathItemBaseEdit, saveBody)
	if err != nil {
		return nil, fmt.Errorf("保存 SCM 商品名称失败: %w", err)
	}
	out["dry_run"] = false
	out["sync_check"] = queryRaw
	out["save_result"] = saveRaw
	return out, nil
}

func executePDDVideoInfo(ctx context.Context, c httpClient, opts pddPrimitiveOptions) (any, error) {
	baseItemID, product, err := resolveBaseItem(ctx, c, opts.StyleCode, opts.BaseItemID)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"baseItemIds": []any{baseItemID},
		"api_name":    "pdd_getCarouselVideoInfo",
	}
	raw, _, err := c.PostJSON(ctx, pathPddCarouselVideo, body)
	if err != nil {
		return nil, fmt.Errorf("查询 PDD 轮播视频信息失败: %w", err)
	}
	return primitiveResult(interfacePddCarouselVideo, "POST", pathPddCarouselVideo, body, raw, map[string]any{
		"style_code":   strings.TrimSpace(opts.StyleCode),
		"base_item_id": baseItemID,
		"product":      productSummary(product),
	}), nil
}

func executePDDPricePrecheck(ctx context.Context, c httpClient, opts pddPrimitiveOptions) (any, error) {
	target, err := resolvePublishTarget(ctx, c, platformPrimitiveOptions{
		Platform:   "pdd",
		StyleCode:  opts.StyleCode,
		BaseItemID: opts.BaseItemID,
		Shop:       opts.Shop,
		ShopID:     opts.ShopID,
	})
	if err != nil {
		return nil, err
	}
	body := pricePrecheckBody(target.Platform, target.BaseItemID, target.ShopID)
	raw, _, err := c.PostJSON(ctx, pathPreCheckPrice, body)
	if err != nil {
		return nil, fmt.Errorf("校验 PDD 控价失败: %w", err)
	}
	return primitiveResult(interfacePreCheckPrice, "POST", pathPreCheckPrice, body, raw, target.Meta()), nil
}

func executePDDAuthStatus(ctx context.Context, c httpClient, opts pddPrimitiveOptions) (any, error) {
	meta := map[string]any{}
	platformShopID := opts.PlatformShopID
	if platformShopID == 0 {
		target, err := resolvePublishTarget(ctx, c, platformPrimitiveOptions{
			Platform:   "pdd",
			StyleCode:  opts.StyleCode,
			BaseItemID: opts.BaseItemID,
			Shop:       opts.Shop,
			ShopID:     opts.ShopID,
		})
		if err != nil {
			return nil, err
		}
		platformShopID = target.PlatformShopID
		meta = target.Meta()
	} else {
		meta["platform_shop_id"] = platformShopID
	}
	body := map[string]any{
		"shopIds":  []any{platformShopID},
		"api_name": "pdd_authorize_authStatus",
	}
	raw, _, err := c.PostJSON(ctx, pathPddAuthorize, body)
	if err != nil {
		return nil, fmt.Errorf("校验 PDD 授权状态失败: %w", err)
	}
	return primitiveResult(interfacePddAuthorize, "POST", pathPddAuthorize, body, raw, meta), nil
}

func executePDDSecConfirm(ctx context.Context, c httpClient, opts pddPrimitiveOptions) (any, error) {
	target, err := resolvePublishTarget(ctx, c, platformPrimitiveOptions{
		Platform:   "pdd",
		StyleCode:  opts.StyleCode,
		BaseItemID: opts.BaseItemID,
		Shop:       opts.Shop,
		ShopID:     opts.ShopID,
	})
	if err != nil {
		return nil, err
	}
	body := secConfirmBody(target)
	raw, _, err := c.PostJSON(ctx, pathSecConfirm, body)
	if err != nil {
		return nil, fmt.Errorf("PDD 铺货二次确认失败: %w", err)
	}
	return primitiveResult(interfaceSecConfirm, "POST", pathSecConfirm, body, raw, target.Meta()), nil
}

func executePDDStorageTask(ctx context.Context, c httpClient, opts pddPrimitiveOptions) (any, error) {
	target, err := resolvePublishTarget(ctx, c, platformPrimitiveOptions{
		Platform:   "pdd",
		StyleCode:  opts.StyleCode,
		BaseItemID: opts.BaseItemID,
		Shop:       opts.Shop,
		ShopID:     opts.ShopID,
	})
	if err != nil {
		return nil, err
	}
	body := storageTaskBody(target)
	out := primitiveResult(interfaceStorageTask, "POST", pathStorageTask, body, nil, target.Meta())
	out["submit"] = opts.Submit
	if !opts.Submit {
		out["dry_run"] = true
		out["next"] = "如确认实际铺货，请加 --submit"
		return out, nil
	}
	raw, _, err := c.PostJSON(ctx, pathStorageTask, body)
	if err != nil {
		return nil, fmt.Errorf("提交 PDD 铺货任务失败: %w", err)
	}
	out["dry_run"] = false
	out["response"] = raw
	out["batch_task_id"] = responseDataString(raw)
	return out, nil
}

func executePDDTaskSpeed(ctx context.Context, c httpClient, opts pddPrimitiveOptions) (any, error) {
	batchTaskID := strings.TrimSpace(opts.BatchTaskID)
	if batchTaskID == "" {
		return nil, fmt.Errorf("--batch-task-id 不能为空")
	}
	params := map[string]any{
		"taskTypeEnum": defaultPublishType,
		"batchTaskId":  batchTaskID,
		"api_name":     "taskScheduling_queryTaskSpeed",
	}
	raw, _, err := c.GetQuery(ctx, pathQueryTaskSpeed, params)
	if err != nil {
		return nil, fmt.Errorf("查询 PDD 铺货任务进度失败: %w", err)
	}
	return primitiveResult(interfaceQueryTask, "GET", pathQueryTaskSpeed, params, raw, map[string]any{
		"batch_task_id": batchTaskID,
	}), nil
}

func executePlatformPricePrecheck(ctx context.Context, c httpClient, opts platformPrimitiveOptions) (any, error) {
	target, err := resolvePublishTarget(ctx, c, opts)
	if err != nil {
		return nil, err
	}
	body := pricePrecheckBody(target.Platform, target.BaseItemID, target.ShopID)
	raw, _, err := c.PostJSON(ctx, pathPreCheckPrice, body)
	if err != nil {
		return nil, fmt.Errorf("校验 %s 控价失败: %w", target.Platform, err)
	}
	return primitiveResult(interfacePreCheckPrice, "POST", pathPreCheckPrice, body, raw, target.Meta()), nil
}

func executePlatformSecConfirm(ctx context.Context, c httpClient, opts platformPrimitiveOptions) (any, error) {
	target, err := resolvePublishTarget(ctx, c, opts)
	if err != nil {
		return nil, err
	}
	body := secConfirmBody(target)
	raw, _, err := c.PostJSON(ctx, pathSecConfirm, body)
	if err != nil {
		return nil, fmt.Errorf("%s 铺货二次确认失败: %w", target.Platform, err)
	}
	return primitiveResult(interfaceSecConfirm, "POST", pathSecConfirm, body, raw, target.Meta()), nil
}

func executePlatformStorageTask(ctx context.Context, c httpClient, opts platformPrimitiveOptions) (any, error) {
	target, err := resolvePublishTarget(ctx, c, opts)
	if err != nil {
		return nil, err
	}
	body := storageTaskBody(target)
	out := primitiveResult(interfaceStorageTask, "POST", pathStorageTask, body, nil, target.Meta())
	out["submit"] = opts.Submit
	if !opts.Submit {
		out["dry_run"] = true
		out["next"] = "如确认实际铺货，请加 --submit"
		return out, nil
	}
	raw, _, err := c.PostJSON(ctx, pathStorageTask, body)
	if err != nil {
		return nil, fmt.Errorf("提交 %s 铺货任务失败: %w", target.Platform, err)
	}
	out["dry_run"] = false
	out["response"] = raw
	out["batch_task_id"] = responseDataString(raw)
	return out, nil
}

func executePlatformTaskSpeed(ctx context.Context, c httpClient, opts platformPrimitiveOptions) (any, error) {
	batchTaskID := strings.TrimSpace(opts.BatchTaskID)
	if batchTaskID == "" {
		return nil, fmt.Errorf("--batch-task-id 不能为空")
	}
	params := map[string]any{
		"taskTypeEnum": defaultPublishType,
		"batchTaskId":  batchTaskID,
		"api_name":     "taskScheduling_queryTaskSpeed",
	}
	raw, _, err := c.GetQuery(ctx, pathQueryTaskSpeed, params)
	if err != nil {
		return nil, fmt.Errorf("查询铺货任务进度失败: %w", err)
	}
	return primitiveResult(interfaceQueryTask, "GET", pathQueryTaskSpeed, params, raw, map[string]any{
		"batch_task_id": batchTaskID,
	}), nil
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
		"interfaces": []map[string]any{
			interfaceInfo(interfaceItemBasePage, "POST", pathItemBasePage),
			interfaceInfo(interfaceShopAll, "GET", pathShopAll),
		},
		"endpoints": []string{
			pathItemBasePage,
			pathShopAll,
		},
		"next": "选择 can_publish=true 的店铺后，可用 scm-item publish-pdd --shop-id 提交铺货预检",
	}, nil
}

func executePublishPDD(ctx context.Context, c httpClient, opts publishPDDOptions) (any, error) {
	return executePublishPlatform(ctx, c, publishPlatformOptions{
		Platform:   "pdd",
		StyleCode:  opts.StyleCode,
		Shop:       opts.Shop,
		ShopID:     opts.ShopID,
		ShelfState: opts.ShelfState,
		Submit:     opts.Submit,
		CheckLog:   opts.CheckLog,
		LogStart:   opts.LogStart,
		LogEnd:     opts.LogEnd,
		LogPage:    opts.LogPage,
		LogSize:    opts.LogSize,
		Detail:     opts.Detail,
	})
}

func executePublishPlatform(ctx context.Context, c httpClient, opts publishPlatformOptions) (any, error) {
	cfg, err := getPlatformConfig(opts.Platform)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(opts.StyleCode) == "" {
		return nil, fmt.Errorf("--style-code 不能为空")
	}
	if strings.TrimSpace(opts.Shop) == "" && opts.ShopID == 0 {
		return nil, fmt.Errorf("--shop 与 --shop-id 至少提供一个")
	}
	if opts.ShelfState == 0 {
		opts.ShelfState = defaultShelfState
	}

	target, err := resolvePublishTarget(ctx, c, platformPrimitiveOptions{
		Platform:  cfg.Key,
		StyleCode: opts.StyleCode,
		Shop:      opts.Shop,
		ShopID:    opts.ShopID,
	})
	if err != nil {
		return nil, err
	}
	canPublishKey := cfg.Key
	if cfg.Key == "1688" {
		canPublishKey = "1688"
	}
	if !containsString(stringListField(target.Product, "canPublishPlatformList"), canPublishKey) {
		return nil, fmt.Errorf("商品 %s 未包含 %s 平台资料，无法铺货", opts.StyleCode, cfg.Key)
	}
	interfaces := []map[string]any{
		interfaceInfo(interfaceItemBasePage, "POST", pathItemBasePage),
		interfaceInfo(interfaceShopAll, "GET", pathShopAll),
	}
	endpoints := []string{pathItemBasePage, pathShopAll}
	var videoRaw any
	if cfg.NeedsPDDVideo {
		videoRaw, _, err = c.PostJSON(ctx, pathPddCarouselVideo, pddVideoInfoBody(target.BaseItemID))
		if err != nil {
			return nil, fmt.Errorf("查询 PDD 轮播视频信息失败: %w", err)
		}
		interfaces = append(interfaces, interfaceInfo(interfacePddCarouselVideo, "POST", pathPddCarouselVideo))
		endpoints = append(endpoints, pathPddCarouselVideo)
	}
	var priceCheckRaw any
	if cfg.NeedsPricePrecheck {
		priceCheckRaw, _, err = c.PostJSON(ctx, pathPreCheckPrice, pricePrecheckBody(cfg.Key, target.BaseItemID, target.ShopID))
		if err != nil {
			return nil, fmt.Errorf("校验 %s 控价失败: %w", cfg.Key, err)
		}
		interfaces = append(interfaces, interfaceInfo(interfacePreCheckPrice, "POST", pathPreCheckPrice))
		endpoints = append(endpoints, pathPreCheckPrice)
	}
	var authRaw any
	if cfg.NeedsPDDAuthorize {
		authRaw, _, err = c.PostJSON(ctx, pathPddAuthorize, map[string]any{
			"shopIds":  []any{target.PlatformShopID},
			"api_name": "pdd_authorize_authStatus",
		})
		if err != nil {
			return nil, fmt.Errorf("校验 PDD 授权状态失败: %w", err)
		}
		if invalid := invalidPDDAuth(dataList(authRaw)); len(invalid) > 0 {
			return nil, fmt.Errorf("PDD 店铺授权不可用: %s", strings.Join(invalid, "；"))
		}
		interfaces = append(interfaces, interfaceInfo(interfacePddAuthorize, "POST", pathPddAuthorize))
		endpoints = append(endpoints, pathPddAuthorize)
	}
	var secConfirmRaw any
	if cfg.NeedsSecConfirm {
		secConfirmRaw, _, err = c.PostJSON(ctx, pathSecConfirm, secConfirmBody(target))
		if err != nil {
			return nil, fmt.Errorf("%s 铺货二次确认失败: %w", cfg.Key, err)
		}
		if rows := dataList(secConfirmRaw); secConfirmHasBlockingRows(rows) {
			return nil, fmt.Errorf("%s 铺货二次确认未通过: %s", cfg.Key, compactJSON(rows))
		}
		interfaces = append(interfaces, interfaceInfo(interfaceSecConfirm, "POST", pathSecConfirm))
		endpoints = append(endpoints, pathSecConfirm)
	}
	publishBody := storageTaskBody(target)
	interfaces = append(interfaces, interfaceInfo(interfaceStorageTask, "POST", pathStorageTask))
	endpoints = append(endpoints, pathStorageTask)

	plan := map[string]any{
		"platform":      cfg.Key,
		"platform_name": cfg.Label,
		"submit":        opts.Submit,
		"style_code":    opts.StyleCode,
		"product":       productSummary(target.Product),
		"shop":          shopSummary(target.Shop),
		"publish_body":  publishBody,
		"interfaces":    interfaces,
		"endpoints":     endpoints,
	}
	if videoRaw != nil {
		plan["video_info"] = dataList(videoRaw)
	}
	if priceCheckRaw != nil {
		plan["precheck"] = dataMap(priceCheckRaw)
	}
	if authRaw != nil {
		plan["auth_status"] = dataList(authRaw)
	}
	if secConfirmRaw != nil {
		plan["sec_confirm"] = dataList(secConfirmRaw)
	}
	if !opts.Submit {
		plan["dry_run"] = true
		plan["next"] = "校验通过；如确认实际铺货，请加 --submit"
		return plan, nil
	}

	res, _, err := c.PostJSON(ctx, pathStorageTask, publishBody)
	if err != nil {
		return nil, fmt.Errorf("提交 %s 铺货任务失败: %w", cfg.Key, err)
	}
	batchTaskID := responseDataString(res)
	plan["dry_run"] = false
	plan["publish_result"] = res
	plan["batch_task_id"] = batchTaskID
	if batchTaskID != "" {
		progress, err := queryTaskSpeed(ctx, c, batchTaskID, 6)
		if err != nil {
			plan["progress_error"] = err.Error()
		} else {
			plan["progress"] = progress
		}
		plan["interfaces"] = append(plan["interfaces"].([]map[string]any), interfaceInfo(interfaceQueryTask, "GET", pathQueryTaskSpeed))
		plan["endpoints"] = append(plan["endpoints"].([]string), pathQueryTaskSpeed)
	}
	if opts.CheckLog {
		logs, err := queryPublishLogs(ctx, c, publishLogOptions{
			Platform:  cfg.Key,
			StyleCode: opts.StyleCode,
			Shop:      displayShopName(target.Shop),
			ShopID:    target.ShopID,
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
		"platform":   defaultString(strings.TrimSpace(opts.Platform), "pdd"),
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
		"interfaces": []map[string]any{
			interfaceInfo(interfacePublishLog, "POST", pathPublishLog),
		},
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
		out["interfaces"] = []map[string]any{
			interfaceInfo(interfacePublishLog, "POST", pathPublishLog),
			interfaceInfo(interfacePublishLogDetail, "GET", pathPublishLogDetail),
			interfaceInfo(interfacePublishLogByID, "GET", pathPublishLogByID),
		}
		out["endpoints"] = []string{
			pathPublishLog,
			pathPublishLogDetail,
			pathPublishLogByID,
		}
	}
	return out, nil
}

func queryTaskSpeed(ctx context.Context, c httpClient, batchTaskID string, maxPolls int) ([]map[string]any, error) {
	if maxPolls <= 0 {
		maxPolls = 1
	}
	out := make([]map[string]any, 0, maxPolls)
	for i := 0; i < maxPolls; i++ {
		raw, _, err := c.GetQuery(ctx, pathQueryTaskSpeed, map[string]any{
			"taskTypeEnum": defaultPublishType,
			"batchTaskId":  batchTaskID,
			"api_name":     "taskScheduling_queryTaskSpeed",
		})
		if err != nil {
			return out, fmt.Errorf("查询 PDD 铺货任务进度失败: %w", err)
		}
		data := dataMap(raw)
		out = append(out, data)
		if b, ok := boolField(data, "finished"); ok && b {
			break
		}
	}
	return out, nil
}

func fetchBaseItemDetail(ctx context.Context, c httpClient, id int64) (map[string]any, any, error) {
	if id == 0 {
		return nil, nil, fmt.Errorf("SCM 商品 id 不能为空")
	}
	q := url.Values{}
	q.Set("id", strconv.FormatInt(id, 10))
	q.Set("api_name", "item_base_detail")
	raw, _, err := c.GetQuery(ctx, pathItemBaseDetail, map[string]any{
		"id":       id,
		"api_name": "item_base_detail",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("查询 SCM 商品详情失败: %w", err)
	}
	item := dataMap(raw)
	if len(item) == 0 {
		return nil, raw, fmt.Errorf("SCM 商品详情为空: %s", q.Encode())
	}
	return item, raw, nil
}

func prepareSCMEditItem(detail map[string]any, newTitle string) map[string]any {
	item := cloneMap(detail)
	item["title"] = newTitle
	item["api_name"] = "item_base_edit"
	return item
}

func scmEditBody(item map[string]any, opts updateTitleOptions) map[string]any {
	return map[string]any{
		"item":             item,
		"checkOpenSync":    opts.CheckSync,
		"skipAddItemToErp": opts.SkipAddERP,
		"api_name":         "item_base_edit",
	}
}

func resolveBaseItem(ctx context.Context, c httpClient, styleCode, inputBaseItemID string) (string, map[string]any, error) {
	baseItemID := strings.TrimSpace(inputBaseItemID)
	if baseItemID != "" {
		return baseItemID, map[string]any{}, nil
	}
	styleCode = strings.TrimSpace(styleCode)
	if styleCode == "" {
		return "", nil, fmt.Errorf("--style-code 与 --base-item-id 至少提供一个")
	}
	product, err := findProductByStyleCode(ctx, c, styleCode)
	if err != nil {
		return "", nil, err
	}
	baseItemID = stringField(product, "itemId")
	if baseItemID == "" {
		baseItemID = stringField(product, "baseItemId")
	}
	if baseItemID == "" {
		return "", nil, fmt.Errorf("商品 %s 缺少 baseItemId/itemId", styleCode)
	}
	return baseItemID, product, nil
}

func resolvePublishTarget(ctx context.Context, c httpClient, opts platformPrimitiveOptions) (publishTarget, error) {
	cfg, err := getPlatformConfig(opts.Platform)
	if err != nil {
		return publishTarget{}, err
	}
	baseItemID, product, err := resolveBaseItem(ctx, c, opts.StyleCode, opts.BaseItemID)
	if err != nil {
		return publishTarget{}, err
	}
	if strings.TrimSpace(opts.Shop) == "" && opts.ShopID == 0 {
		return publishTarget{}, fmt.Errorf("--shop 与 --shop-id 至少提供一个")
	}
	shop, err := findPlatformShop(ctx, c, cfg, baseItemID, opts.Shop, opts.ShopID)
	if err != nil {
		return publishTarget{}, err
	}
	shopID, err := int64Field(shop, "id", "shopId")
	if err != nil {
		return publishTarget{}, fmt.Errorf("店铺缺少 id/shopId: %w", err)
	}
	if disabledReason := shopDisabledReason(shop); disabledReason != "" {
		return publishTarget{}, fmt.Errorf("店铺 %s 当前不可铺货：%s", displayShopName(shop), disabledReason)
	}
	platformShopID := int64(0)
	if cfg.NeedsPDDAuthorize {
		platformShopID, err = int64Field(shop, "taobaoId", "platformShopId")
		if err != nil {
			return publishTarget{}, fmt.Errorf("店铺缺少 taobaoId/platformShopId，无法校验 PDD 授权: %w", err)
		}
	}
	return publishTarget{
		Product:        product,
		Shop:           shop,
		BaseItemID:     baseItemID,
		ShopID:         shopID,
		PlatformShopID: platformShopID,
		StyleCode:      strings.TrimSpace(opts.StyleCode),
		Platform:       cfg.Key,
	}, nil
}

func (t publishTarget) Meta() map[string]any {
	return map[string]any{
		"platform":         t.Platform,
		"style_code":       t.StyleCode,
		"base_item_id":     t.BaseItemID,
		"shop_id":          t.ShopID,
		"platform_shop_id": t.PlatformShopID,
		"product":          productSummary(t.Product),
		"shop":             shopSummary(t.Shop),
	}
}

func pddVideoInfoBody(baseItemID string) map[string]any {
	return map[string]any{
		"baseItemIds": []any{baseItemID},
		"api_name":    "pdd_getCarouselVideoInfo",
	}
}

func pricePrecheckBody(platform, baseItemID string, shopID int64) map[string]any {
	return map[string]any{
		"itemId":       baseItemID,
		"shopIds":      []any{shopID},
		"platformType": platform,
		"api_name":     "ltsTask_preCheckControllerPrice",
	}
}

func secConfirmBody(target publishTarget) map[string]any {
	platform := target.Platform
	if platform == "1688" {
		platform = "a1688"
	}
	return map[string]any{
		"platformType": platform,
		"itemShopList": []map[string]any{
			{
				"baseItemId": target.BaseItemID,
				"shops": []map[string]any{
					{
						"shopId":  target.ShopID,
						"channel": "",
					},
				},
				"outId": stringField(target.Product, "outerId"),
			},
		},
		"api_name": "item_base_secConfirmationItemV2",
	}
}

func storageTaskBody(target publishTarget) map[string]any {
	return map[string]any{
		"itemIds":                []any{target.BaseItemID},
		"shopIds":                []any{target.ShopID},
		"taskEntrance":           defaultPublishEntrance,
		"taskType":               defaultPublishType,
		"channel":                "",
		"batchOperationEntrance": 0,
		"isDistribution":         intFieldDefault(target.Product, "isDistribution", 0),
		"api_name":               "taskScheduling_storageTask",
	}
}

func primitiveResult(name, method, path string, request, response any, meta map[string]any) map[string]any {
	out := map[string]any{
		"interface": interfaceInfo(name, method, path),
		"interfaces": []map[string]any{
			interfaceInfo(name, method, path),
		},
		"endpoint": path,
		"endpoints": []string{
			path,
		},
		"request":  request,
		"response": response,
	}
	for k, v := range meta {
		out[k] = v
	}
	return out
}

func interfaceInfo(name, method, path string) map[string]any {
	return map[string]any{
		"name":   name,
		"method": method,
		"path":   path,
	}
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
		"pageNo":          1,
		"pageSize":        10,
		"leafCategories":  []any{},
		"outerIds":        []any{styleCode},
		"outerIdBlur":     0,
		"cgSupplierIds":   []any{},
		"platformItemIds": []any{},
		"companyNames":    []any{},
		"shopNames":       []any{},
		"api_name":        apiNameItemBasePage,
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
	cfg, err := getPlatformConfig(platform)
	if err != nil {
		return nil, err
	}
	source := cfg.ShopSource
	subSource := cfg.ShopSubSource
	raw, _, err := c.GetQuery(ctx, pathShopAll, map[string]any{
		"source":     source,
		"subSource":  subSource,
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

func findPlatformShop(ctx context.Context, c httpClient, cfg platformConfig, baseItemID, shopName string, shopID int64) (map[string]any, error) {
	shops, err := queryShops(ctx, c, cfg.Key, baseItemID)
	if err != nil {
		return nil, err
	}
	if shopID != 0 {
		for _, shop := range shops {
			id, _ := int64Field(shop, "id", "shopId")
			if id == shopID {
				return shop, nil
			}
		}
		return nil, fmt.Errorf("未找到 shopId=%d 的 %s 店铺", shopID, cfg.Key)
	}

	keyword := strings.TrimSpace(shopName)
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
	return nil, fmt.Errorf("未找到 %s 店铺 %s", cfg.Key, keyword)
}

func getPlatformConfig(platform string) (platformConfig, error) {
	key := strings.TrimSpace(platform)
	if key == "" {
		return platformConfig{}, fmt.Errorf("--platform 不能为空")
	}
	cfgs := map[string]platformConfig{
		"pdd":      {Key: "pdd", Label: "拼多多", ShopSource: "pdd", NeedsPricePrecheck: true, NeedsSecConfirm: true, NeedsPDDVideo: true, NeedsPDDAuthorize: true},
		"tb":       {Key: "tb", Label: "淘宝", ShopSource: "tb", NeedsPricePrecheck: true, NeedsSecConfirm: true},
		"fxg":      {Key: "fxg", Label: "抖音", ShopSource: "fxg", NeedsPricePrecheck: true, NeedsSecConfirm: true},
		"kuaishou": {Key: "kuaishou", Label: "快手", ShopSource: "kuaishou", NeedsPricePrecheck: true},
		"jd":       {Key: "jd", Label: "京东", ShopSource: "jd", NeedsPricePrecheck: true, NeedsSecConfirm: true},
		"1688":     {Key: "1688", Label: "阿里巴巴", ShopSource: "1688", SecConfirmPlatform: "a1688", NeedsPricePrecheck: true, NeedsSecConfirm: true},
		"tm":       {Key: "tm", Label: "天猫", ShopSource: "tm", NeedsPricePrecheck: true},
		"tjb":      {Key: "tjb", Label: "淘特", ShopSource: "tb", ShopSubSource: "tjb", NeedsPricePrecheck: true},
		"yz":       {Key: "yz", Label: "有赞", ShopSource: "yz", NeedsPricePrecheck: true},
		"wxsph":    {Key: "wxsph", Label: "微信小店（视频号）", ShopSource: "wxsph", NeedsPricePrecheck: true},
		"wxxd":     {Key: "wxxd", Label: "微信小店", ShopSource: "wxxd", NeedsPricePrecheck: true},
		"wd":       {Key: "wd", Label: "微店", ShopSource: "wd", NeedsPricePrecheck: true},
		"xhs":      {Key: "xhs", Label: "小红书", ShopSource: "xhs", NeedsPricePrecheck: true},
		"xy":       {Key: "xy", Label: "闲鱼", ShopSource: "xy", NeedsPricePrecheck: true},
		"fxg_gx":   {Key: "fxg_gx", Label: "抖音供销平台", ShopSource: "fxg_gx", NeedsPricePrecheck: true},
		"yyjk":     {Key: "yyjk", Label: "美团闪购平台", ShopSource: "yyjk", NeedsPricePrecheck: true},
		"pddtemu":  {Key: "pddtemu", Label: "Temu", ShopSource: "pddtemu", NeedsPricePrecheck: true},
		"shein":    {Key: "shein", Label: "Shein", ShopSource: "shein", NeedsPricePrecheck: true},
		"ktt":      {Key: "ktt", Label: "快团团", ShopSource: "ktt"},
		"jdms":     {Key: "jdms", Label: "京东到家", ShopSource: "jdms"},
	}
	cfg, ok := cfgs[key]
	if !ok {
		return platformConfig{}, fmt.Errorf("暂不支持平台 %s", key)
	}
	if cfg.SecConfirmPlatform == "" {
		cfg.SecConfirmPlatform = cfg.Key
	}
	return cfg, nil
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
		if !recordLooksLikePlatform(record, opts.Platform) {
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

func recordLooksLikePlatform(record map[string]any, platform string) bool {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		platform = "pdd"
	}
	seenPlatform := false
	for _, key := range []string{"platformType", "platform", "source", "shopType"} {
		value := stringField(record, key)
		if value == "" {
			continue
		}
		seenPlatform = true
		if strings.EqualFold(value, platform) {
			return true
		}
	}
	// Older log rows may omit platform in the list but still contain PDD shop fields.
	if platform == "pdd" && stringField(record, "pddGoodsId", "pddProductId") != "" {
		return true
	}
	return !seenPlatform
}

func secConfirmHasBlockingRows(rows []map[string]any) bool {
	for _, row := range rows {
		if b, ok := boolField(row, "finished"); ok && !b {
			return true
		}
		if failMsgs := anyList(row["failMsg"]); len(failMsgs) > 0 {
			return true
		}
		if items := mapList(row["itemShopList"]); len(items) > 0 {
			for _, item := range items {
				if b, ok := boolField(item, "checkResult"); ok && !b {
					return true
				}
			}
		}
		if capacity := mapList(row["capacityList"]); len(capacity) > 0 {
			return true
		}
	}
	return false
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

func dataList(raw any) []map[string]any {
	if m, ok := raw.(map[string]any); ok {
		return mapList(m["data"])
	}
	return mapList(raw)
}

func asMap(raw any) map[string]any {
	if m, ok := raw.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func responseDataString(raw any) string {
	if value := stringField(dataMap(raw), "data"); value != "" {
		return value
	}
	return stringField(asMap(raw), "data")
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

func anyList(v any) []any {
	switch list := v.(type) {
	case []any:
		return list
	case []string:
		out := make([]any, 0, len(list))
		for _, item := range list {
			out = append(out, item)
		}
		return out
	case []map[string]any:
		out := make([]any, 0, len(list))
		for _, item := range list {
			out = append(out, item)
		}
		return out
	default:
		return nil
	}
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func invalidPDDAuth(rows []map[string]any) []string {
	out := make([]string, 0)
	for _, row := range rows {
		authorizeValid, hasAuthorizeValid := boolField(row, "authorizeValid")
		tokenValid, hasTokenValid := boolField(row, "tokenValid")
		if (!hasAuthorizeValid || authorizeValid) && (!hasTokenValid || tokenValid) {
			continue
		}
		shopID := stringField(row, "shopId")
		if shopID == "" {
			shopID = "unknown"
		}
		out = append(out, fmt.Sprintf("shopId=%s authorizeValid=%v tokenValid=%v", shopID, authorizeValid, tokenValid))
	}
	return out
}

func compactJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
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
