# Changelog

本文件记录 kuaimai-cli 各版本的主要变更。

## Unreleased

### Added

- **meta_data.json v1.6.0**：`item` 域 **1157** 个 operation（erp-items-core `/item` Controller 全量注册）
- **`internal/pagination`**：`--page-all` 海量数据防护（500/1000 条阈值、交互 `[y/N]`、分片合并）
- 全局参数 **`--page-limit`**、**`--page-confirm`**（`prompt` | `yes` | `no`）
- **`service item item-query-list-v2`** 等 meta 驱动命令（无 shortcut 的接口走 service）
- Skill **kuaimai-item v2.0.0**：架构分层说明 + 3 份新 references（meta-execution、service、query-list-v2）
- `auth check`：探测 accessToken 与 API 连通性
- `auth list` / `auth use` / `auth login --profile`：多账号 profile
- `item update-title`：get-detail 合并后 save，简化改标题
- `kuaimai-cli upgrade`：对比 GitHub Release 版本
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
