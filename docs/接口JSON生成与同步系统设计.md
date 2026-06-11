# Kuaimai CLI Registry 生成与同步系统设计

  ## 1. 背景

  `kuaimai-cli` 需要通过标准化的 registry JSON 知道有哪些后台接口可以调用。

  这些接口 JSON 不应该由 `kuaimai-cli` 手写维护，也不应该让 CLI 直接理解 YAPI、OpenAPI、飞书/钉钉表格等上游来源。

  推荐方案是拆分为两个系统：

  - `registry-generator`：负责发现接口、拉取文档、评估完整性、提醒人员补齐、生成 registry JSON
  - `kuaimai-cli`：只负责消费 registry JSON，完成校验、同步、调用

  ---

  ## 2. 系统边界

  ### 2.1 registry-generator

  也可以先沿用当前项目里的 `kuaimaierp-cli-auto/internal-tools/api-onboard`。

  职责：

  - 扫描接口清单表
  - 拉取 YAPI 接口文档
  - 判断接口文档是否完整
  - 创建/更新工单
  - 通知接口负责人补充字段说明
  - Review 补齐后的接口
  - 生成标准 registry JSON
  - 发布 registry artifact 包

  ### 2.2 kuaimai-cli

  职责：

  - 从 registry artifact 拉取候选 JSON
  - 对候选 JSON 做 validate
  - 与本地 registry 做 diff
  - 人工确认后 apply
  - 通过 registry 提供能力：
    - `kuaimai schema`
    - `kuaimai capabilities`
    - `kuaimai web call`
    - 业务快捷命令 shortcut

  `kuaimai-cli` 不直接连接 YAPI、表格或通知系统。

  ---

  ## 3. 推荐系统架构

  ```text
  [接口清单表]
    记录接口、负责人、状态、优先级
          |
          v
  [registry-generator]
    拉 YAPI -> evaluate -> 工单/通知 -> review -> 生成 JSON
          |
          v
  [registry artifact]
    manifest.json + web-apis/*.json
          |
          v
  [kuaimai-cli]
    sync -> diff -> validate -> apply
          |
          v
  [本地执行]
    schema / capabilities / web call / shortcut

  ———

  ## 4. 接口清单表设计

  接口清单表用于记录哪些接口需要接入，以及当前处理状态。

  字段建议：

  apiId
  接口标题
  业务模块
  YAPI项目ID
  YAPI接口ID
  YAPI链接
  接口路径 path
  method
  负责人
  状态
  优先级
  最近YAPI更新时间
  最近拉取时间
  最近评估结果
  registry版本
  备注

  状态建议：

  待接入
  待补齐
  补齐中
  待Review
  已通过
  已产出
  已废弃

  ———

  ## 5. 接口发现与增量拉取策略

  不建议每次 CLI 初始化都全量拉取。

  推荐策略：

  首次初始化：
    扫描 YAPI project/category
    发现接口
    写入接口清单表

  日常运行：
    只处理状态为 待接入 / 待Review 的接口
    或处理 YAPI up_time 大于最近拉取时间的接口

  定期巡检：
    每天或每周低频全量扫描
    发现新增、删除、变更接口

  人工触发：
    支持输入单个 YAPI URL 或业务 URL
    立即拉取并评估

  判断是否需要重新拉取：

  YAPI up_time > 接口清单表.lastPulledYapiTime
  => 需要重新拉取、评估、生成候选 JSON

  ———

  ## 6. 文档补齐与在线编辑方案

  字段含义最终应该维护在 YAPI 上，而不是只修改生成出来的 registry JSON。

  原因：

  - YAPI 是上游文档源
  - registry JSON 是生成产物
  - 如果只改 JSON，下次重新生成会丢失
  - 其他系统也无法复用补齐后的文档

  推荐流程：

  registry-generator 拉取 YAPI
   -> evaluate 检查字段说明
   -> 不完整：创建工单并通知负责人
   -> 负责人在 YAPI 修改字段说明
   -> registry-generator review
   -> 通过后生成 registry JSON

  可选增强：

  在接口清单表/工单表中填写字段说明建议
   -> registry-generator 读取建议
   -> 调用 YAPI Open API 回写
   -> 再重新评估

  MVP 阶段建议：

  工单表只记录缺失项和状态
  字段说明直接在 YAPI 修改

  ———

  ## 7. Registry JSON 标准字段

  每个接口最终生成一个标准 JSON：

  web-apis/<apiId>.json

  示例：

  {
    "id": "item.stock.queryList",
    "title": "查询商品库存列表",
    "description": "查询 ERP 后台商品库存列表，返回商品、商家编码、库存、价格等字段。",
    "transport": "web",
    "method": "GET",
    "path": "/item/stock/queryList",
    "risk": "read",
    "stability": "stable",
    "auth": {
      "type": "header",
      "key": "accessToken"
    },
    "params": {
      "page": {
        "type": "number",
        "in": "query",
        "required": false,
        "description": "页码"
      },
      "limit": {
        "type": "number",
        "in": "query",
        "required": false,
        "description": "每页数量"
      }
    },
    "response": {
      "successPath": "suc",
      "listPath": "data.list",
      "primaryFields": [
        "sysItemId",
        "title",
        "outerId"
      ]
    },
    "source": {
      "type": "yapi",
      "yapiProject": 1220,
      "yapiId": 78991,
      "url": "https://yapi.raycloud.com/project/1220/interface/api/78991",
      "syncedAt": "2026-06-09T10:00:00+08:00"
    },
    "examples": [
      {
        "title": "通过通用 web call 调用",
        "command": "kuaimai web call item.stock.queryList --params '{\"page\":1,\"limit\":20}'"
      }
    ]
  }

  ———

  ## 8. Registry JSON 必填字段

  建议必填：

  id
  title
  description
  transport
  method
  path
  risk
  stability
  auth
  params
  response
  source
  examples

  字段说明：

  id:
    全局唯一接口 ID，例如 item.stock.queryList

  title:
    中文标题

  description:
    接口用途说明

  transport:
    当前固定为 web

  method:
    GET / POST / PUT / PATCH / DELETE

  path:
    ERP 后台接口路径

  risk:
    read / write / high-risk-write

  stability:
    stable / beta / deprecated

  auth:
    鉴权方式，当前一般是 header accessToken

  params:
    入参定义，没有参数也给 {}

  response:
    返回结构摘要

  source:
    来源信息，用于追踪来自 YAPI、OpenAPI、人工整理等

  examples:
    调用示例

  ———

  ## 9. JSON 校验责任

  需要两边都校验：

  registry-generator:
    生成 JSON 前校验
    不合规则不发布

  kuaimai-cli:
    sync/apply 前校验
    不合规则拒绝合并

  建议定义 JSON Schema：

  code/registry/schema/web-api.schema.json

  ———

  ## 10. Registry Artifact 包设计

  生成器最终发布的是一个 artifact 包，而不是让 CLI 直接连 YAPI。

  目录结构：

  registry-package/
    manifest.json
    web-apis/
      item.stock.queryList.json
      order.queryList.json

  manifest.json 示例：

  {
    "version": "2026.06.09.1",
    "generatedAt": "2026-06-09T10:00:00+08:00",
    "source": "registry-generator",
    "apis": [
      {
        "id": "item.stock.queryList",
        "path": "web-apis/item.stock.queryList.json",
        "method": "GET",
        "apiPath": "/item/stock/queryList",
        "stability": "stable",
        "risk": "read"
      }
    ]
  }

  artifact 可以发布到：

  Git 仓库
  内部对象存储/CDN
  内部 HTTP 服务
  飞书/钉钉云盘
  本地共享目录

  ———

  ## 11. kuaimai-cli 同步流程

  推荐命令：

  kuaimai registry sync
  kuaimai registry diff
  kuaimai registry validate
  kuaimai registry apply

  流程：

  1. kuaimai registry sync
     从 registry artifact 拉取候选 JSON

  2. kuaimai registry validate
     校验候选 JSON 是否符合 schema

  3. kuaimai registry diff
     对比候选 JSON 和本地 registry

  4. 人工 review

  5. kuaimai registry apply
     将候选 JSON 应用到本地 registry

  6. 本地 CLI 使用更新后的 registry

  本地目录建议：

  code/
    registry/
      web-apis/
        item.stock.queryList.json

    .registry-candidates/
      web-apis/
        item.stock.queryList.json

  ———

  ## 12. Diff 变更类型

  建议分类：

  safe-add:
    新增低风险读接口

  doc-only:
    只修改标题、描述、示例、字段说明

  param-change:
    入参发生变化

  response-change:
    返回结构发生变化

  path-change:
    path 或 method 发生变化

  risk-change:
    风险等级发生变化

  remove:
    上游删除接口

  处理建议：

  safe-add:
    可以快速通过

  doc-only:
    可以快速通过

  param-change:
    需要人工 review

  response-change:
    需要人工 review，必要时 smoke test

  path-change:
    必须人工确认

  risk-change:
    必须人工确认

  remove:
    不直接删除本地 registry，先标记 deprecated

  ———

  ## 13. kuaimai-cli 使用方式

  registry apply 后，CLI 自动暴露能力：

  kuaimai capabilities
  kuaimai schema item.stock.queryList
  kuaimai web call item.stock.queryList --params '{"page":1,"limit":20}'

  如果高频接口需要更友好的命令，再额外实现 shortcut：

  kuaimai item stock query-list --page 1 --limit 20

  shortcut 内部也必须读取 registry：

  registry.Get("item.stock.queryList")
   -> method/path
   -> transport.Request()

  不要在 shortcut 里重复手写 URL。

  ———

  ## 14. 部署阶段

  ### 阶段 1：本地半自动

  不需要单独部署服务。

  开发者本地运行 registry-generator
   -> 生成 output/web-apis/*.json
   -> kuaimai registry sync --source ../output

  适合验证流程。

  ### 阶段 2：CI/定时任务

  不需要常驻服务。

  定时任务 / 手动 CI
   -> 扫描接口清单表
   -> 拉 YAPI
   -> 生成 registry-package
   -> 发布到 Git 仓库或对象存储

  适合团队内部使用。

  ### 阶段 3：平台化服务

  接口数量多、协作人多之后再做。

  registry-generator 服务
   -> Web 页面管理接口清单
   -> webhook/定时扫描 YAPI
   -> 自动通知和 review
   -> 发布 registry HTTP endpoint

  ———

  ## 15. 推荐 MVP

  第一版建议只做这些：

  1. 建接口清单表
  2. 支持手动输入 YAPI URL 拉取接口
  3. 评估接口完整性
  4. 不完整时创建工单
  5. 人在 YAPI 上补字段
  6. review 通过后生成 registry JSON
  7. kuaimai-cli 从本地目录 sync JSON
  8. validate 后 apply
  9. schema / web call 可使用新接口

  暂时不做：

  实时服务
  自动监听群消息
  复杂 Web 管理后台
  自动上线写接口
  每次 CLI 启动自动拉取最新 registry

  ———

  ## 16. 核心原则

  kuaimai-cli 不直接依赖 YAPI、表格、通知系统
  registry-generator 负责生成标准 registry JSON
  registry artifact 是两个系统之间的稳定边界
  字段语义维护在 YAPI
  JSON 是生成产物
  CLI 正常执行只读本地已 apply 的 registry
  接口变更必须经过 validate、diff、review、apply
  写操作接口必须更加谨慎