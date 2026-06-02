<!-- GitHub 创建 Release 时的默认说明（Web UI）。Tag 推送发版时由 GoReleaser 自动生成正文，本文件供手动补发/编辑时参考。 -->

## 概述

快麦 ERP 商品 CLI **{{VERSION}}** — 架构对标飞书 lark-cli。

**完整变更记录**：请同步更新并引用 [CHANGELOG.md](https://github.com/kuaimai-cli/kuaimai-cli/blob/main/CHANGELOG.md) 中对应版本小节。

---

## 本版亮点

<!-- 发版时填写：从 CHANGELOG Unreleased 挪入 -->

### Added

-

### Changed

-

### Fixed

-

---

## 能力快照（主线）

- **meta**：`internal/registry/meta_data.json` **v1.6.0**，`item` 域 **1157** 个 operation
- **shortcuts**：`item +list` / `count` / `get-detail` / `save` / `update-title`
- **分页**：`--page-all` · `--page-limit` · `--page-confirm`（`internal/pagination`）
- **Skill**：`kuaimai-shared` + `kuaimai-item` v2.0.0（8 个 `references/`）

---

## 安装与升级

### npm / npx（推荐）

```bash
npx @kuaimai-cli/cli@latest install
```

或全局安装：

```bash
npm install -g @kuaimai-cli/cli@{{VERSION}}
```

### Release 二进制

从 **Assets** 下载对应平台包：

| 平台 | 文件名 |
|------|--------|
| macOS Apple Silicon | `kuaimai-cli-{{VERSION}}-darwin-arm64.tar.gz` |
| macOS Intel | `kuaimai-cli-{{VERSION}}-darwin-amd64.tar.gz` |
| Linux amd64 | `kuaimai-cli-{{VERSION}}-linux-amd64.tar.gz` |
| Linux arm64 | `kuaimai-cli-{{VERSION}}-linux-arm64.tar.gz` |
| Windows | `kuaimai-cli-{{VERSION}}-windows-amd64.zip` |

解压后将 `kuaimai-cli` 放入 `PATH`，或使用 `checksums.txt` 校验后安装。

### 首次使用

```bash
kuaimai-cli config init
kuaimai-cli auth login "<accessToken>"
kuaimai-cli skill install
kuaimai-cli auth check --output json
kuaimai-cli doctor --output json
```

---

## 验收建议

```bash
kuaimai-cli --version
kuaimai-cli schema --output json | jq '.data.version'
kuaimai-cli item +list --body '{"pageNo":1,"pageSize":1}' --output json
```

写操作请先 `--dry-run`。

---

## 文档

- [README](https://github.com/kuaimai-cli/kuaimai-cli/blob/main/README.md)
- [文档索引](https://github.com/kuaimai-cli/kuaimai-cli/blob/main/docs/README.md)
- [Agent 安装指南](https://github.com/kuaimai-cli/kuaimai-cli/blob/main/docs/快麦%20CLI%20安装（Agent%20专用）.md)
- [开发发布流程](https://github.com/kuaimai-cli/kuaimai-cli/blob/main/docs/快麦%20CLI%20开发发布流程文档.md)

---

## 贡献者与维护

感谢所有参与本版本的同学。问题反馈请开 [GitHub Issues](https://github.com/kuaimai-cli/kuaimai-cli/issues)。
