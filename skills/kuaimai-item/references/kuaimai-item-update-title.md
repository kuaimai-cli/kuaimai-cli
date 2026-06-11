# item update-title — 修改商品标题（推荐）

## 适用场景

用户要**改标题 / 改名**，且已知 `sysItemId`（或已通过 `+list` 定位）。

对应 apiId `item.item-update-title`（`post_json`，`write:true`）；shortcuts 含 get-detail → save **编排**，`web call` 仅为原子 save。

## 前置检查

```bash
kuaimai-cli auth status --output json
```

未登录不要继续。写操作**首次必须**带 `--dry-run --verbose`，用户确认后再去掉 `--dry-run`。

## 命令

```bash
kuaimai-cli item update-title \
  --sys-item-id <sysItemId> \
  --title "<新标题>" \
  --dry-run --verbose --output json --no-color
```

用户确认预览无误后，**同一命令**去掉 `--dry-run`（可保留 `--verbose`）再执行。

## 内部流程

CLI 自动完成：

1. `GET /item/getItemDetail?sysItemId=…` 拉全量详情
2. 合并新 `title` 及 save 所需字段（含 `suiteBridgeList` 等）
3. `POST JSON /item/saveItem` 提交

**dry-run 时第 ③ 步仅预览，第 ① 步仍会拉详情。**

## 端到端（只有标题、无 ID）

| 步骤 | 操作 |
|------|------|
| 1 | [`+list`](kuaimai-item-list.md) 按标题搜索，取 `sysItemId` |
| 2 | 本命令 `--dry-run` 预览 |
| 3 | 用户确认后去掉 `--dry-run` 提交 |

## 备选路径

复杂字段修改或 `update-title` 不满足时，见 [`kuaimai-item-save.md`](kuaimai-item-save.md)（get-detail + jq + save）。

## 禁止

- 不要只传 `{"sysItemId":...,"title":"..."}` 调 `save`
- 不要在未 `auth login` 时反复重试写操作

## 失败处理

优先读 `error.hint`；必要时 `--verbose` 查看请求预览。
