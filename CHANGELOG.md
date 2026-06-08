# Changelog

本文件记录 kuaimai-cli 各版本的主要变更。

## Unreleased

### Added

- **meta_data.json v1.7.0**：新增 `scm` 域 **195** 个 operation（erp-scm staff/logging/item/dsb）；`item` 域 **1095** 个 operation；scm 使用 meta `baseUrl`（`https://scm.superboss.cc/`）
- **Skill kuaimai-scm v1.0.0**：供应链域路由 + **7** 个 `references/`（domain-routing、meta-execution、service、staff、logging、item-base、dsb）
- **`service scm <operation>`**：meta 驱动，自动请求 scm 域名；与 item 共用分页/dry-run/Schema 管线
- **scm meta 生成脚本**：`scripts/generate_meta/generate_scm_meta.py`、`scripts/normalize_meta/normalize_scm_meta.py`
- **`doctor`**：新增 `skill_kuaimai_scm` 检查（含 `references/`）
- **`skill install`**：默认安装 `kuaimai-shared` + `kuaimai-item` + `kuaimai-scm`

### Changed

- **文档全量对齐 v1.7.0**：README、`docs/` 索引、架构说明、Agent 选型、安装指南、验收测试、每阶段能力
- **`AGENTS.md`**：补充 scm 域路由与 `service scm` 示例
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
- **`service item item-query-list-v2`** 等 meta 驱动命令（无 shortcut 的接口走 service）
- Skill **kuaimai-item v2.0.0**：架构分层说明 + 3 份新 references（meta-execution、service、query-list-v2）
- `auth check`：探测 accessToken 与 API 连通性
- `auth list` / `auth use` / `auth login --profile`：多账号 profile
- `item update-title`：get-detail 合并后 save，简化改标题
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
- Skill 飞书风格对齐：`kuaimai-shared`、`kuaimai-item` 重写；`kuaimai-item/references/` 扩展至 **8** 份
- `skill install`：GitHub Contents API 递归安装整目录；API 失败时回退仅 `SKILL.md`
- `doctor`：检测 `kuaimai-item` 是否含 `references/` 目录
- `service` 层：`contentType` 路由、`requestSchema` required 轻校验、Schema 默认值
- `config init` 模板增加 `auth.profile` / `auth.profiles`
- `AGENTS.md`：补充分页参数与 service 兜底说明

### Framework

- **meta + Skill + CLI 基础能力闭环**：后续新增接口 primarily 登记 meta +（可选）shortcut/Skill，无需改底层框架

## 0.1.0

- 阶段一～三：config / auth / api / item 域（list、count、get-detail、save）
- Skill：kuaimai-item、kuaimai-shared
- npm `@kuaimai-cli/cli` 与 GitHub Release 分发
