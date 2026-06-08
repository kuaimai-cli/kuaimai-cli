kuaimai-cli meta_data.json 定义规范（V1.0 最终定稿）

**适用范围**：erp-items-core 商品域（`item`）、erp-scm 供应链域（`scm`）

**对标标准**：飞书 lark-cli openapi 规范

**文件路径**：`internal/registry/meta_data.json`

**核心定位**：kuaimai-cli 底层原子接口注册表，CLI 执行、参数校验、schema 自省、全量分页拉取的唯一数据来源

---

## 一、整体文件结构

### 通用模板

```json
{
  "version": "1.7.0",
  "services": {
    "item": {
      "summary": "快麦ERP商品服务",
      "description": "商品管理、库存、查询、编辑、上下架",
      "operations": { }
    },
    "scm": {
      "summary": "快麦ERP供应链服务",
      "description": "供应链管理：员工、商品中心、操作日志、平台铺货配置",
      "baseUrl": "https://scm.superboss.cc/",
      "operations": { }
    }
  }
}
```

### 顶层字段说明


| **字段**        | **说明**                            |
| ------------- | --------------------------------- |
| `version`     | 元数据文件版本号，用于版本兼容、格式校验              |
| `services`    | 业务服务集合，按 ERP 业务域分组管理所有接口          |
| `item`        | 商品业务域，对应 **erp-items-core** |
| `scm`         | 供应链业务域，对应 **erp-scm**；须含 `baseUrl` |
| `baseUrl`     | （scm 等独立域名）CLI 请求根地址；item 域用 config `api.url` |
| `summary`     | 业务域简短描述                           |
| `description` | 业务域详细功能说明                         |
| `operations`  | 当前业务域下所有接口集合，单个 key 对应一条后端原子接口    |


---

## 二、基础接口字段定义规则

### 1. service 名称

| service | 后端工程 | API 根地址 |
|---------|----------|------------|
| `item`  | erp-items-core | config `api.url`（默认 erp1.superboss.cc） |
| `scm`   | erp-scm        | meta `baseUrl`（`https://scm.superboss.cc/`） |

### 2. operation 接口标识命名规则（解决后续接口重复问题）

**旧规则问题**：单纯`list/count/save` 会随着接口增多、多模块复用导致 key 重复，命令匹配冲突。

**最终规范**：采用 `业务模块-操作语义` 组合命名，全局唯一、永不重复，小写短横线分隔、见名知意。


| **后端 Controller 方法名** | **新版唯一 meta 接口标识** | **说明**    |
| --------------------- | ------------------ | --------- |
| queryList             | stock-list         | 库存模块-列表查询 |
| queryCount            | stock-count        | 库存模块-数量统计 |
| getItemDetail         | item-detail        | 商品模块-详情查询 |
| saveItem              | item-save          | 商品模块-保存修改 |
| updateTitle           | item-update-title  | 商品模块-标题修改 |


### 3. path 接口请求路径

- 规则：完整复刻后端 `@RequestMapping` 路由地址
- 示例：后端 `@RequestMapping("/item/stock/queryList")` → 配置 `"path": "/item/stock/queryList"`

### 4. method 请求方式

- 与后端接口请求方式完全一致
- 可选值：`GET`、`POST`

### 5. contentType 内容类型（最终固定枚举）

作用：决定 CLI 如何组装请求参数、请求头与请求体格式


| **后端接口类型**    | **固定取值**    | **说明**                                   |
| ------------- | ----------- | ---------------------------------------- |
| GET 请求        | `get_query` | 参数拼接在 URL 后面                             |
| POST 表单提交     | `post_form` | `application/x-www-form-urlencoded` 表单传参 |
| POST JSON 请求体 | `post_json` | `@RequestBody`接收 JSON 参数                 |


### 6. write 读写标记

作用：控制 CLI `--dry-run` 模拟预览功能开关


|              |         |
| ------------ | ------- |
| 查询类接口        | `false` |
| 新增/修改/删除 写接口 | `true`  |


### 7. pageable 分页标记（核心逻辑详解）

**字段作用**：控制 CLI `--page-all` 全量拉取功能开关


|                |         |
| -------------- | ------- |
| 分页列表接口         | `true`  |
| 详情/保存/统计 非列表接口 | `false` |


**--page-all 执行逻辑**：

- 开启后 CLI 自动循环翻页查询，无需人工分页
- 单次默认查询条数：**50 条**（由 requestSchema 中 pageSize 默认值控制）
- 自动迭代 pageNo，直到后端返回数据为空、达到硬上限（**1000 页**）、或触发防护规则

**海量数据防护（CLI 全局 flag，与 Skill 对齐）**：

| flag | 说明 |
|------|------|
| `--page-all` | 开启自动翻页（仅 `pageable:true`） |
| `--page-limit N` | 最多拉取 N 条（0=不限条数） |
| `--page-confirm` | `prompt`（默认，500/1000 阈值交互 `[y/N]`）\| `yes` \| `no` |

实现：`internal/pagination` → `PostFormAllPages` / `RequestAllPages`；分片合并 `ChunkSize=500`。

### 8. summary 接口描述

- 填写**中文**功能简述，用于 CLI 列表展示、AI 识别接口能力
- 优先来源：`@ApiOperation(value=...)` → 方法 Javadoc → 禁止类头注释（Created by / Copyright）
- 示例：`"summary": "铺货日志分页查询"`、`"summary": "一键应用品牌配置"`

### 9. requestSchema 字段 desc 规范

- `desc` 须为可读中文说明，**禁止**仅重复字段名（如 `"desc": "shopType"`）
- 优先来源：`@ApiModelProperty` → 字段 Javadoc → `scripts/normalize_meta/normalize_scm_meta.py` 中 `FIELD_DESC_MAP`
- 分页字段须含 `default`：`pageNo` 默认 1，`pageSize` 默认 50
- 分页接口 `required` 须含 `pageNo`、`pageSize`；日志类接口须含 `startTime`、`endTime`（若存在）

---

## 三、Schema 结构定义 & 核心逻辑答疑

### 1. requestSchema 请求结构

数据源：优先 `erp-items-core` 入参 DTO，缺失则取 `erp-core`

作用：定义接口入参字段、类型、说明、默认值、必填项，用于 CLI 参数校验、AI 自动拼装参数

### 2. responseSchema 响应结构

数据源：后端统一 `BaseResponse` + 业务出参实体

作用：定义接口返回数据结构，用于结果解析、字段展示

### 3. 关键答疑：meta 已包含 Schema，为何还要单独 Schema 查询能力？

- **meta_data.json**：**存储层**，永久存放所有接口的入参、出参结构、接口元信息，是唯一数据源
- **schema 命令**：**展示/自省层**，执行 `kuaimai-cli schema` 时，从 meta_data.json 中读取 schema 结构并格式化展示

**一句话总结**：meta 存数据，schema 查数据、看数据。

### 4. CLI 完整执行链路（核心业务流程）

1. 程序启动：加载`meta_data.json` 全量接口配置并缓存
2. 用户执行 service 命令：匹配对应 operation 接口
3. 读取 meta 中 path / method / contentType / write / pageable 规则
4. 读取 requestSchema 校验入参合法性
5. 根据 contentType 自动组装 GET/表单/JSON 请求体
6. 若开启 `--page-all` 且 pageable=true，经 `internal/pagination` 全量翻页（含阈值与 `--page-limit`）
7. 请求后端接口，根据 responseSchema 解析返回结果

---

## 四、最终规范示例（单接口最简可直接使用）

文件路径：`@kuaimai-cli/internal/registry/meta_data.json`

```json
{
  "version": "1.2.0",
  "services": {
    "item": {
      "summary": "快麦ERP商品服务",
      "description": "商品管理、库存、查询、编辑、上下架",
      "operations": {
        "stock-list": {
          "summary": "查询商品库存列表",
          "method": "POST",
          "path": "/item/stock/queryList",
          "contentType": "post_form",
          "write": false,
          "pageable": true,
          "requestSchema": {
            "type": "object",
            "properties": {
              "title": {
                "type": "string",
                "desc": "商品标题"
              },
              "categoryId": {
                "type": "number",
                "desc": "分类ID"
              },
              "pageNo": {
                "type": "number",
                "desc": "页码",
                "default": 1
              },
              "pageSize": {
                "type": "number",
                "desc": "每页条数",
                "default": 50
              }
            },
            "required": [
              "pageNo",
              "pageSize"
            ]
          },
          "responseSchema": {
            "type": "object",
            "properties": {
              "ok": {
                "type": "boolean"
              },
              "data": {
                "type": "array"
              },
              "total": {
                "type": "number"
              }
            }
          }
        }
      }
    }
  }
}

```

---

## 五、核心问题最终总结

**1. 接口重名问题**：已彻底解决，采用 `模块-操作` 命名，全局唯一，支持后续无限新增接口不冲突

**2. pageable 分页逻辑**：true 开启全量自动翻页；配合 `--page-limit` / `--page-confirm` 做海量数据防护

**3. Schema 与 meta 关系**：meta 是数据存储载体，schema 是 CLI 自省查询能力，数据同源不冲突

**4. 最新 contentType 规范**：统一使用 get_query / post_form / post_json 枚举值

---

## 六、与代码实现对齐说明（2026-06-08）

| 决策项 | 采用方案 |
|--------|----------|
| JSON 结构 | `services` / `operations` 均为**对象 map**（与本规范 §一 一致） |
| 当前版本 | **v1.7.0**；`generated_at` 可选（当前 2026-06-08） |
| 登记范围 | `item` **1095** 个 operation（erp-items-core）；`scm` **195** 个 operation（erp-scm staff/logging/item/dsb） |
| scm 生成 | `python3 scripts/generate_meta/generate_scm_meta.py` → 自动调用 `normalize_scm_meta.py` |
| scm baseUrl | `service scm` 使用 meta `baseUrl`，**不**读 config `api.url` |
| operation 命名 | meta + `service` 使用 `stock-list` 等；**shortcuts** 仍为 `item +list` / `get-detail` 等（见 `schema` 输出 `shortcut` 字段） |
| `write` | **业务读写**：查询 `false`，写接口 `true`；`--dry-run` 仅对 `write:true` 生效 |
| `pageable` | 仅 `pageable:true` 且用户传 `--page-all` 时全量翻页；item **60** + scm **41** 个为 true |
| 分页防护 | `--page-limit` · `--page-confirm`（500/1000 阈值）；硬上限 1000 页 |
| Schema | `requestSchema` / `responseSchema` 已写入 meta；`service` 做 **required 轻校验**；`schema` 命令输出完整结构 |
| 双轨 | item：Agent **优先 shortcuts** + `service item` 兜底；scm：**仅** `service scm`（无 shortcuts） |

`operation` 命名：`模块-路径` 小写短横线；CLI **核心 6 个**（5 shortcut + 1 service-only）：`stock-list`、`stock-count`、`item-query-list-v2`、`item-detail`、`item-save`、`item-update-title`。请求 Schema 优先 erp-items-core 入参类，缺失从 erp-core 补；保存商品为 `SysItemModel`。

**meta 生成流水线**（维护者）：

```bash
# 从 erp-items-core 生成 → 过滤 /item Controller → 规范化 → 写入 internal/registry/meta_data.json
python3 scripts/generate_meta/generate_item_meta.py
python3 scripts/filter_meta/filter_by_item_controller.py
python3 scripts/normalize_meta/normalize_meta.py   # 若存在
```

`service` 示例：

```bash
kuaimai-cli service item stock-list --body '{"title":"test","pageNo":1,"pageSize":50}'
kuaimai-cli service item item-query-list-v2 --body '{"title":"test","pageNo":1,"pageSize":50}'
kuaimai-cli service item item-detail --body '{"sysItemId":123}'
kuaimai-cli service scm staff-query --body '{"pageNo":1,"pageSize":20}'
kuaimai-cli service scm logging-publish-log \
  --body '{"pageNo":1,"pageSize":20,"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59"}'
kuaimai-cli schema --output json
kuaimai-cli item +list --body '{"title":"test"}' --page-all --page-limit 200 --page-confirm yes
```