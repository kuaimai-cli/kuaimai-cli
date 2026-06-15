# Changelog

本文件记录 kuaimai-cli 各版本的主要变更。

## 0.2.8

### Added

- **SCM 商品编辑 shortcut**：新增 `scm-item update-title`，支持按 `--style-code` 或 `--id` 读取 SCM 商品详情，只修改 `title` 并按前端保存格式提交 `/item/base/queryErpItems` 与 `/item/base/edit`；默认 dry-run，显式 `--submit` 才保存。
- **SCM 商品接口文档**：将原 PDD 铺货文档重命名为 `docs/SCM商品相关接口.md`，全量记录 SCM 商品查询、编辑、铺货、日志相关接口。

### Changed

- **Skill 路由约束**：强化 ERP 商品档案与 SCM 商品的意图区分，普通“商品 / 款式编码 / outerId / 标题修改”默认走 SCM，只有明确“商品档案 / sysItemId”才走 ERP 商品档案。
- **版本提示缓存**：升级提示从 24 小时缓存调整为 1 小时缓存；已发现新版本时持续提示到用户升级。

## 0.2.6

### Fixed

- **SCM shortcut 路径修正**：`scm-item +list`、铺货、铺货日志等 SCM Web 接口统一使用 `.json` 业务接口路径，避免命中 SCM 前端页面并返回 HTML。
- **SCM 查询参数补齐**：`scm-item +list` 补齐浏览器请求中的空数组筛选字段与 `api_name=item_base_page`，更贴近 DevTools 实际请求。
- **Registry 单接口域名路由**：`web call` 按单个 `apis[apiId].baseUrl` 路由，避免同一 service 前缀下 ERP/SCM 接口混用目标域。

### Changed

- **Shortcut API 配置**：新增 `shortcuts.erp-item.api_url` 与 `shortcuts.scm-item.api_url` 默认配置，ERP/SCM curated shortcuts 不再依赖单一 `api.url`。
- **诊断提示**：接口返回 HTML 时提示检查 `targetHost/path` 与 SCM `.json` 后缀。

## 0.2.5

### Changed

- **SCM 可铺货商品前置查询**：新增 `kuaimai-cli scm-item +list` 按款式编码/标题查询 SCM 可铺货商品，新增 `kuaimai-cli scm-item shops` 查询指定平台可铺货店铺并输出 `can_publish` / `disabled_reason`。
- **Registry 自动同步**：从进程启动前的 `os.Args` 预判断改为 Cobra 解析后的 `PersistentPreRunE` 判断；`config/auth/doctor/skill/registry/upgrade/completion/help/version` 跳过 registry 同步，避免基础命令被远端 registry 可用性阻断。
- **npm 安装向导**：`config init` 后主动执行一次 `registry sync`，将 registry 可用性问题提前到安装阶段暴露；同步失败不阻断安装。
- **Skill 同步**：默认 Skill 完整性从“任一目录存在”升级为 5 个 Agent 目录全量检查（`.agents/.cursor/.codex/.claude/.windsurf`），缺任一目录或 `references/` 会触发重装；命令结束后的自动同步改为同步完成后再退出。
- **Skill 重装策略**：默认 Skill 安装/自动同步会先删除旧默认目录（含历史 `kuaimai-item`、`kuaimai-scm`），再安装 `kuaimai-shared`、`kuaimai-erp-item`、`kuaimai-scm-item`，避免 Agent 读取旧路由。
- **doctor**：新增 `skill_roots` 输出，逐目录展示默认 Skill 与 `references/` 是否齐全。
- **SCM PDD 铺货 shortcut**：新增 `kuaimai-cli scm-item publish-pdd`，按款式编码定位任意 SCM 商品、按店铺名/ID 定位任意 PDD 店铺，执行临时配置与平台资料完整性校验；默认停在最终提交前，只有 `--submit` 才调用 `/pdd/batchPublishItem`，可加 `--check-log` 查询最近铺货日志与失败原因。
- **SCM 铺货日志 shortcut**：新增 `kuaimai-cli scm-item publish-log`，支持按款式编码、店铺名/ID、时间范围查询 `/logging/publishLog`，并用 `--detail` 拉取 `/logging/publishLogDetail` 的单品状态和 `errorMessage`。

## 0.2.3

### Added

- **Registry 远端同步**：`registry sync` / `registry watch`；命令前 `SyncIfNeeded` 自动拉取 `registry.source`（默认 `http://open-cli.kuaimai.com/registry/registry.json`）；新增 `capabilities`、`web call`
- **API 网关转发**：业务 HTTP 统一经 `api.gateway_url`（默认 `https://open-cli.kuaimai.com`）的 `POST /api/forward` 转发至 ERP/SCM；配置项 `api.gateway_url`；文档 [API网关转发说明](./docs/API网关转发说明.md)
- **`web call --body`**：统一请求体参数（按 `contentType` 自动转为 `--params` 或 `--data`），兼容原 `service` 调用习惯
- **`schema [apiId]`**：优先输出 registry v2 字段（`domain`、`title`、`transport` 等）

### Removed

- **`service` 命令**：registry 接口统一走 `web call <apiId>`（对标飞书域命令，避免与 `service` 子命令重复）

### Changed

- **`api.timeout` 默认 60s**（与网关上游超时一致）
- **`kuaimai-shared` v1.2.0**：补充网关配置与 429 排错说明
- **Registry 消费路径**：`capabilities` → `schema` → `web call`；文档与 Skill 全面移除 `service item|scm` 写法
- **`bootstrapRegistry`**：同步后不再动态注册 `service` 子命令树
- **Skill 对标飞书瘦身**：`kuaimai-shared` v1.1 增加 registry 发现流程；`kuaimai-erp-item` v3 / `kuaimai-scm-item` v2 移除硬编码 meta 表，仅保留意图路由、shortcuts 与 references；`kuaimai-*-service.md` 重命名为 `kuaimai-*-web-call.md`

### Added (prior)

- **meta_data.json v1.7.0**：新增 `scm` 域 **195** 个 operation（erp-scm staff/logging/item/dsb）；`item` 域 **1095** 个 operation；scm 使用 meta `baseUrl`（`https://scm.superboss.cc/`）
- **Skill kuaimai-scm-item v1.0.0**：供应链域路由 + **7** 个 `references/`（domain-routing、meta-execution、service、staff、logging、item-base、dsb）
- **`web call scm.<operation>`**：meta 驱动，自动请求 scm 域名；与 item 共用分页/dry-run/Schema 管线
- **scm meta 生成脚本**：`scripts/generate_meta/generate_scm_meta.py`、`scripts/normalize_meta/normalize_scm_meta.py`
- **`doctor`**：新增 `skill_kuaimai_scm_item` 检查（含 `references/`）
- **`skill install`**：默认安装 `kuaimai-shared` + `kuaimai-erp-item` + `kuaimai-scm-item`

### Changed

- **文档全量对齐 v1.7.0**：README、`docs/` 索引、架构说明、Agent 选型、安装指南、验收测试、每阶段能力
- **`AGENTS.md`**：补充 scm 域路由与 `web call` 示例
- **Skill 自动同步**：任意命令结束后后台执行 `install --if-stale` 逻辑（未安装 / Release 更新 / CLI 版本变化，24h Release 查询缓存）；`skill install` 无参数时默认等同 `--if-stale`，强制重装用 `--force`

### Fixed

- **npm `install.js` 双源下载（对标 @larksuite/cli）**：GitHub Release 失败后自动尝试 `registry.npmmirror.com/-/binary/kuaimai-cli/...`；支持 `KUAIMAI_CLI_SKIP_MIRROR`、`KUAIMAI_CLI_DOWNLOAD_URL`
- 维护说明见 [npmmirror 二进制镜像](./docs/npmmirror-二进制镜像.md)（须在 cnpmcore 注册后镜像才可用）

- **npm 安装向导**：全局已存在旧版 `@kuaimai-cli/cli` 时不再一律跳过；对比向导包版本后自动 `npm install -g @kuaimai-cli/cli@<版本>`，并刷新各安装目录下 Go 二进制；CLI 升级后强制重跑 `skill install`
- 环境变量 `KUAIMAI_CLI_FORCE_INSTALL=1` 可强制重装全局包与 Skills

### Added

- **版本感知与一键升级（对标飞书 lark-cli）**：
  - 任意命令结束后 stderr 提示新版本（24h 缓存，`~/.kuaimai-cli/version-check.json`）
  - `kuaimai-cli upgrade` 默认执行 `npm install -g @kuaimai-cli/cli@latest` 并同步 Skills；`--check-only` 仅检查
  - `skill install --if-stale`：仅在 Release/CLI 变化或未安装时更新
  - CLI 版本变更后后台自动同步 Skills（`~/.kuaimai-cli/skill-sync.json`）
  - 环境变量：`KUAIMAI_CLI_SKIP_UPDATE_CHECK`、`KUAIMAI_CLI_SKIP_SKILL_SYNC`

- **meta_data.json v1.6.0**：`item` 域 **1157** 个 operation（erp-items-core `/item` Controller 全量注册）
- **`internal/pagination`**：`--page-all` 海量数据防护（500/1000 条阈值、交互 `[y/N]`、分片合并）
- 全局参数 **`--page-limit`**、**`--page-confirm`**（`prompt` | `yes` | `no`）
- **`web call item.item-query-list-v2`** 等 meta 驱动命令（无 shortcut 的接口走 service）
- Skill **kuaimai-erp-item v2.0.0**：架构分层说明 + 3 份新 references（meta-execution、service、query-list-v2）
- `auth check`：探测 accessToken 与 API 连通性
- `auth list` / `auth use` / `auth login --profile`：多账号 profile
- `erp-item update-title`：get-detail 合并后 save，简化改标题
- `kuaimai-cli upgrade`：对比 GitHub Release；**后续版本**默认 npm 一键升级（见 Unreleased）
- `kuaimai-cli doctor`：安装自检
- `tests/cli_e2e`：冒烟 E2E（mock HTTP）
- 根目录 `README.md`、CI workflow
- `Dockerfile`（可选容器分发）
- meta 维护脚本：`scripts/generate_meta/`、`scripts/filter_meta/`、`scripts/normalize_meta/`

### Changed

- **文档全量对齐代码**：`docs/` 索引、架构说明、开发白皮书、验收测试、Agent 选型流程、meta 定义规范
- **发布与分发**：`npm/README.md`、`.goreleaser.yaml` Release 说明模板、`.github/RELEASE_TEMPLATE.md`；归档附带 `CHANGELOG.md`
- `PostFormAllPages` / `RequestAllPages` 统一走 `internal/pagination`（硬上限 1000 页保留）
- Skill 飞书风格对齐：`kuaimai-shared`、`kuaimai-erp-item` 重写；`kuaimai-erp-item/references/` 扩展至 **8** 份
- `skill install`：GitHub Contents API 递归安装整目录；API 失败时回退仅 `SKILL.md`
- `doctor`：检测 `kuaimai-erp-item` 是否含 `references/` 目录
- `web call`：`contentType` 路由、`requestSchema` required 轻校验、Schema 默认值
- `config init` 模板增加 `auth.profile` / `auth.profiles`
- `AGENTS.md`：补充分页参数与 service 兜底说明

### Framework

- **meta + Skill + CLI 基础能力闭环**：后续新增接口 primarily 登记 meta +（可选）shortcut/Skill，无需改底层框架

## 0.1.0

- 阶段一～三：config / auth / api / item 域（list、count、get-detail、save）
- Skill：kuaimai-erp-item、kuaimai-shared
- npm `@kuaimai-cli/cli` 与 GitHub Release 分发
