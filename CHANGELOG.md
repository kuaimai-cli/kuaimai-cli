# Changelog

本文件记录 kuaimai-cli 各版本的主要变更。

## Unreleased

### Added

- `auth check`：探测 accessToken 与 API 连通性
- `auth list` / `auth use` / `auth login --profile`：多账号 profile
- `item update-title`：get-detail 合并后 save，简化改标题
- `kuaimai-cli upgrade`：对比 GitHub Release 版本
- `kuaimai-cli doctor`：安装自检
- `tests/cli_e2e`：冒烟 E2E（mock HTTP）
- 根目录 `README.md`、CI workflow
- `Dockerfile`（可选容器分发）

### Changed

- `config init` 模板增加 `auth.profile` / `auth.profiles`

## 0.1.0

- 阶段一～三：config / auth / api / item 域（list、count、get-detail、save）
- Skill：kuaimai-item、kuaimai-shared
- npm `@kuaimai/cli` 与 GitHub Release 分发
