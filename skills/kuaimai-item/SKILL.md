---
name: kuaimai-item
version: 1.0.0
description: "快麦 ERP 商品（erp-items-core）：按标题搜索列表、统计数量、查详情、改标题。用户提到商品、SKU、标题、货号、列表、有多少、详情、改名时使用。"
metadata:
  requires:
    bins: ["kuaimai-cli"]
---

# kuaimai-item

快麦 CLI **商品域**。处理商品相关自然语言时：

1. **先读** [`kuaimai-shared`](../kuaimai-shared/SKILL.md)（鉴权、输出信封、CLI 路径、安全）
2. **在终端执行** 下文命令，禁止手写 HTTP / curl

## 意图路由

| 用户说法（关键词） | 命令 | 结果字段 |
|-------------------|------|----------|
| 有多少 / 几个 / 总数 / 统计 + 标题 | `item count` + `title` | `data.data.total` |
| 列出 / 搜索 / 查找 / 有哪些 + 标题 | `item +list` + `title` | `data` 列表 |
| 某商品详情 / sysItemId | `item get-detail` | `data[0]` |
| 改标题 / 改名 | `item update-title`（或 `get-detail` → jq → `save`） | 先 `--dry-run` |

只有标题没有 ID 时：先 `+list` 取 `sysItemId`，再 `get-detail` 或 `save`。

## 前置检查

```bash
kuaimai-cli auth status --output json
```

未登录则按 `kuaimai-shared` 引导用户 `auth login`，不要继续调 item 接口。

写操作（`save`）**首次必须** `--dry-run --verbose`，用户确认后再去掉 `--dry-run`。

## 命令模板

将 `<关键词>`、`<sysItemId>`、`<新标题>` 替换为用户提供的值；示例 ID 勿照抄。

### 按标题统计数量

```bash
kuaimai-cli item count \
  --body '{"title":"<关键词>"}' \
  --output json --no-color
```

用中文回答 **`data.data.total`**。不要用 `+list` 再人工数条数。

### 按标题搜索列表

```bash
kuaimai-cli item +list \
  --body '{"title":"<关键词>","pageNo":1,"pageSize":50}' \
  --output json --no-color
```

全量：加 `--page-all`（自动识别 `pageNo`/`pageSize`）。

`list` / `count` 的 `--body` 为 JSON，CLI 会转为 `application/x-www-form-urlencoded`。

### 商品详情

参数为 **sysItemId**（系统商品长整型 ID），不是货号、商家编码等短编号。

```bash
kuaimai-cli item get-detail \
  --sys-item-id <sysItemId> \
  --output json --no-color
```

### 修改标题（推荐 update-title）

```bash
kuaimai-cli item update-title \
  --sys-item-id <sysItemId> \
  --title "<新标题>" \
  --dry-run --verbose --output json --no-color
```

确认后去掉 `--dry-run` 再执行。CLI 内部会 `get-detail` 合并全量 body 后 `save`。

### 修改标题（手动 jq + save）

**禁止**只传 `{"sysItemId":...,"title":"..."}` 调用 `save`，会失败。必须 **全量 body**：

```bash
kuaimai-cli item save \
  --body "$(kuaimai-cli item get-detail --sys-item-id <sysItemId> --output json | jq -c '.data[0] | .title = "<新标题>" | .suiteBridgeList = .itemSuiteBridgeList | del(.itemSuiteBridgeList)')" \
  --dry-run --verbose --output json --no-color
```

用户确认预览无误后，**同一命令**去掉 `--dry-run --verbose` 再执行。

## 对话示例

| 用户说法 | 执行要点 |
|----------|----------|
| 有多少个标题带「春季」的商品？ | `count` + `"title":"春季"` → 报 `total` |
| 查标题包含 test 的商品 | `+list` + `"title":"test"` |
| 商品 5444595188086784 的详情 | `get-detail --sys-item-id 5444595188086784` |
| 把这个商品标题改成「新名称」 | `update-title`（先 dry-run） |

## 禁止

- 不要手写 URL 或 curl 调 ERP；只用 `kuaimai-cli item` 子命令
- 按标题问数量必须用 **`item count`** 且 body 含 `title`
- `item count` **未带** `title` 时是其它默认筛选下的总数，≠「标题含某词」的数量
- 不要在未 `auth login` 时反复重试写操作

## 子命令速查

| 命令 | Method | 说明 |
|------|--------|------|
| `item +list` | POST form | 列表；`--body` 含 `title`、`pageNo`、`pageSize` |
| `item count` | POST form | 总数；`--body` 含 `title` |
| `item get-detail` | GET | `--sys-item-id` |
| `item save` | POST JSON | 全量 body；写前先 `--dry-run` |
| `item update-title` | GET+POST | `--sys-item-id` + `--title`；写前先 `--dry-run` |

进一步参数：`kuaimai-cli item --help` · `kuaimai-cli service item list --help` · `kuaimai-cli schema`
