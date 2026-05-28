# kuaimai-cli 架构设计与分阶段开发白皮书（对标飞书 Lark-CLI）

> 配套：[文档索引](./README.md) · [本地使用实操指南](./kuaimai-cli%20本地使用实操指南.md) · [系统架构说明](./系统架构与飞书对标说明.md) · [每阶段新增能力](./每阶段新增能力.md) · [验收测试](./kuaimai-cli%20验收测试.md)

---

## 一、项目概述

### 1.1 项目定位

kuaimai-cli 是快麦 **erp-items-core** 商品业务专属、平台级私有化命令行工具，架构与交互对标飞书 **lark-cli**。

**核心目标**：架构一次性对齐飞书，能力分阶段渐进补齐；适配人工运维、Shell 脚本自动化、AI Agent。

**当前状态（阶段三 item 域 + 阶段四部分）**：

- 基建：config / auth / api / output / client / runner  
- 业务：**商品标题查改闭环** — `item +list` → `item update-title`（或 `get-detail` → `save`）  
- 元数据：`meta_data.json` **v1.2.0**（1 service / 4 operations）  
- 平台：Skill · csv/ndjson · 连接池/熔断 · 脱敏 · 审计 · `--page-all`  
- 阶段四：`auth check` / 多 profile · `doctor` / `upgrade` · E2E · 根 README/CHANGELOG/CI

### 1.2 核心设计原则

- **分层解耦**：`cmd` → `internal` → `shortcuts`，禁止跨层乱调用  
- **基建先行**：配置、鉴权、HTTP、输出先于业务  
- **架构终局思维**：目录按终局搭建，后续叠加能力与业务域  
- **体验统一**：三级命令、全局参数、stdout/stderr 分离对标飞书  

### 1.3 终局能力目标（阶段四）

E2E · 自动升级 · 多账号 · 更多 erp 业务域 · NPM/Docker · AGENTS.md 终版

---

## 二、技术栈规范

| 模块 | 技术 | 作用 |
|------|------|------|
| 语言 | Go 1.22+ | 单二进制、跨平台 |
| 命令 | Cobra | 子命令、帮助、补全 |
| 配置 | Viper | `~/.kuaimai-cli/config.yaml` |
| 凭证 | go-keyring | `accessToken` 密钥链 |
| HTTP | net/http 封装 | 重试、超时、form/JSON、连接池/熔断 |
| 输出 | go-pretty + 自研 | table / json / csv / ndjson |

---

## 三、项目目录架构（终局，一次性搭建）

```text
kuaimai-cli/
├── cmd/                    # 仅注册与参数
│   ├── root.go             # 注册各子命令
│   ├── authcmd/            # login、check、list、use…
│   ├── doctorcmd/          # 安装自检
│   └── upgradecmd/         # 版本查询
├── internal/
│   ├── config/             # template.go + Init/Get/Set
│   ├── auth/
│   ├── client/             # HTTP、重试、--page-all、连接池/熔断
│   ├── output/
│   ├── registry/           # meta_data.json（embed）
│   ├── cmdutil/
│   └── core/
├── shortcuts/
│   ├── common/             # runner, RunGET/RunPOST/RunPOSTForm…
│   └── item/               # erp-items-core 商品
├── scripts/fetch_meta/
├── skills/                 # kuaimai-item 等 SKILL.md
├── tests/cli_e2e/          # E2E 冒烟（mock HTTP）
└── .github/workflows/ci.yml
```

**铁规**：

1. `cmd` 无业务、无网络逻辑  
2. 业务仅在 `shortcuts/`  
3. `skills/` 无 Go 业务代码  

---

## 四、统一命令体系

### 4.1 全局基础命令

```bash
kuaimai-cli --help
kuaimai-cli --verbose

kuaimai-cli config init | get | set
kuaimai-cli auth login <accessToken> [--profile name] | logout | status | check | list | use <profile>

kuaimai-cli api GET|POST|PUT|DELETE <path> [--body '{}']

kuaimai-cli schema
kuaimai-cli service item <operation> [--body '{}']

kuaimai-cli skill list | skill install <name> | skill install-all
kuaimai-cli upgrade | doctor
kuaimai-cli completion bash|zsh|powershell
```

### 4.2 三级业务命令（对齐飞书）

| 层级 | 说明 | 示例 |
|------|------|------|
| 快捷 `+` | 高频列表 | `item +list` |
| 标准子命令 | 主力运维 | `item save`、`item get-detail` |
| 原始 `api` | 脚本兜底 | `api POST /item/saveItem` |

### 4.3 全局参数

`--output table|json|csv|ndjson` · `--dry-run` · `--verbose` · `--page-all` · `--no-color`

### 4.4 config init 规范

**链路**：`config init` → `internal/config.Init()` → `template.go` → `~/.kuaimai-cli/config.yaml`

**默认 API**：`https://erp1.superboss.cc/`（单环境，路径不追加 `.json`）

**输出优先级**：`--output` > `cli.output` > `table`

详见 [本地使用实操指南 · 配置管理](./kuaimai-cli%20本地使用实操指南.md#二配置管理首次必做)。

### 4.5 当前业务 shortcuts 索引（erp-items-core）

**注册位置**：`cmd/root.go` → `shortcuts/item.Register`

| 命令组 | 子命令 | HTTP | 后端路径 |
|--------|--------|------|----------|
| `item` | `list`/`+list` | POST form | `/item/stock/queryList` |
| `item` | `count` | POST form | `/item/stock/queryCount` |
| `item` | `get-detail` | GET | `/item/getItemDetail` |
| `item` | `save` | POST JSON | `/item/saveItem` |
| `item` | `update-title` | GET+POST | `getItemDetail` + `saveItem` |

**商品标题能力映射**：

- 查：`list`/`+list` 的 `--body.title`（配合 `isAccurate` 模糊/精确）  
- 改：`save` 的 `--body` 中 `sysItemId` + `title`  

完整对照表见 [每阶段新增能力 · 附录](./每阶段新增能力.md#附录shortcuts-与-api-对照表当前)。

**新增 shortcuts 步骤**：

1. 在 `shortcuts/item/`（或新域目录）实现，使用 `common.Run*`  
2. `cmd/root.go` 中 `Register(rootCmd)`  
3. 更新 `internal/registry/meta_data.json`  
4. 更新 `skills/kuaimai-item/SKILL.md` 与本文档、[验收测试](./kuaimai-cli%20验收测试.md)

---

## 五、分阶段开发落地指南

### 阶段一：基建骨架版（~60%，已完成）

配置 · 鉴权 · `api` · 输出信封 · 友好错误

### 阶段二：标准企业版（~80%，已完成）

- `shortcuts/common/runner` + `RunGET`/`RunPOST`/`RunPOSTForm`  
- `schema` / `service` / `meta_data.json`（随业务演进，当前 v1.2.0）  
- 重试 · `--page-all` · completion · dry-run · 彩色 table  

### 阶段三：平台进阶版（已完成，item 域）

- **erp-items-core shortcuts**：list/count/get-detail/save  
- **商品标题查改**（系统当前主要业务验收项）  
- Skill · csv/ndjson · 连接池/熔断 · 脱敏 · form 分页 `--page-all` · 审计日志  

### 阶段四：AI 终局版（进行中）

| 能力 | 状态 |
|------|------|
| E2E 冒烟 `tests/cli_e2e` | ✅ |
| `upgrade` / `doctor` | ✅ |
| 多账号 `auth list` / `auth use` / `--profile` | ✅ |
| `item update-title` | ✅ |
| `auth check` | ✅ |
| README / CHANGELOG / Dockerfile / CI | ✅ |
| GitHub Release · npm · Skill | ✅（阶段四初已具备） |
| 更多 erp 业务域 shortcuts | 规划 |

---

## 六、本地调试与编译

```bash
go build -o kuaimai-cli
make build

./kuaimai-cli --verbose item +list \
  --body '{"title":"测试","pageNo":1,"pageSize":10}' --output json

./kuaimai-cli item save \
  --body '{"sysItemId":123,"title":"新标题"}' \
  --dry-run --verbose
```

跨平台见 `Makefile` / `dist/`。

---

## 七、与飞书 CLI 差异（迭代依据）

**已对齐**：分层 · 三级命令 · shortcuts 手写 · registry/service · 密钥链 · 结构化 stdout · Skill 目录 · csv/ndjson  

**差异**：业务域体量（当前 1 域 vs 飞书 18+）· 鉴权（accessToken vs OAuth）· E2E/升级待补齐  

**专属优势**：贴合 erp-items-core 真实路径、商品标题场景端到端可脚本化、维护面小  

---

## 八、开发自查清单

- [ ] 目录符合第三节规范  
- [ ] `cmd` 无业务逻辑  
- [ ] 新接口同步 `shortcuts/item` + `meta_data.json` + Skill + 文档  
- [ ] `api.url` 为 `https://erp1.superboss.cc/`  
- [ ] 三级命令可用：`item +list` / `service item list` / `api POST`  
- [ ] 标题流程：`+list` → `update-title --dry-run` → `update-title` 可跑通  
- [ ] 错误无堆栈  
- [ ] `go build` / `go test ./...` 通过  

详细验收见 [kuaimai-cli 验收测试.md](./kuaimai-cli%20验收测试.md)。

---

## 九、总结

kuaimai-cli 采用「**骨架一次搭建、能力分阶段补齐**」：阶段一/二交付飞书式基建；**阶段三交付 erp-items-core 商品域，并以标题查改为首要业务场景**。日常联调与命令示例以 [本地使用实操指南](./kuaimai-cli%20本地使用实操指南.md) 为准；架构与对标见 [系统架构与飞书对标说明](./系统架构与飞书对标说明.md)。
