# kuaimai-cli 架构设计与分阶段开发白皮书（对标飞书 Lark-CLI）

> 配套：[文档索引](./README.md) · [meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md) · [系统架构说明](./系统架构与飞书对标说明.md) · [每阶段新增能力](./每阶段新增能力.md) · [验收测试](./kuaimai-cli%20验收测试.md)

---

## 一、项目概述

### 1.1 项目定位

kuaimai-cli 是快麦 **erp-items-core** 商品与 **erp-scm** 供应链业务专属、平台级私有化命令行工具，架构与交互对标飞书 **lark-cli**。

**核心目标**：架构一次性对齐飞书，能力分阶段渐进补齐；适配人工运维、Shell 脚本自动化、AI Agent。

**当前状态（框架已定稿 + item/scm 双域）**：

- 基建：config / auth / api / output / client / runner / **pagination**
- 业务：**商品标题查改闭环** — `item +list` → `item update-title`（或 `get-detail` → `save`）
- 业务：**供应链只读查询** — `service scm`（staff / logging / item-base / dsb 等，无 shortcuts）
- 元数据：`meta_data.json` **v1.7.0**（2 services / **1290** operations：item 1095 + scm 195；核心见 [定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md)）
- 平台：Skill（shared + item v2.0.0 + scm v1.0.0）· csv/ndjson · 连接池/熔断 · 脱敏 · 审计 · `--page-all` / **`--page-limit`** / **`--page-confirm`**
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
│   └── upgradecmd/         # 版本检查与一键升级（默认 npm 安装）
├── internal/
│   ├── config/             # template.go + Init/Get/Set
│   ├── auth/
│   ├── client/             # HTTP、重试、分页入口、连接池/熔断
│   ├── pagination/         # --page-all 阈值、交互续查、分片合并
│   ├── output/
│   ├── registry/           # meta_data.json（embed）
│   ├── cmdutil/
│   └── core/
├── shortcuts/
│   ├── common/             # runner, RunGET/RunPOST/RunPOSTForm…
│   └── item/               # erp-items-core 商品
├── scripts/fetch_meta/
├── skills/                 # kuaimai-shared、kuaimai-item、kuaimai-scm（SKILL.md + references/）
├── tests/cli_e2e/          # E2E 冒烟（mock HTTP）
└── .github/workflows/ci.yml
```

**铁规**：

1. `cmd` 无业务、无网络逻辑  
2. 业务仅在 `shortcuts/`  
3. `skills/` 无 Go 业务代码；域 Skill 主文件路由，细节放 `references/`  

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
kuaimai-cli service item|scm <operation> [--body '{}']

kuaimai-cli skill list | skill install [name...]
kuaimai-cli upgrade [--check-only] | doctor
kuaimai-cli completion bash|zsh|powershell
```

### 4.2 三级业务命令（对齐飞书）

| 层级 | 说明 | 示例 |
|------|------|------|
| 快捷 `+` | 高频列表 | `item +list` |
| 标准子命令 | 主力运维 | `item save`、`item get-detail` |
| 原始 `api` | 脚本兜底 | `api POST /item/saveItem` |

### 4.3 全局参数

`--output table|json|csv|ndjson` · `--dry-run` · `--verbose` · `--page-all` · `--page-limit` · `--page-confirm` · `--no-color`

### 4.4 config init 规范

**链路**：`config init` → `internal/config.Init()` → `template.go` → `~/.kuaimai-cli/config.yaml`

**默认 API**：`https://erp1.superboss.cc/`（单环境，路径不追加 `.json`）

**输出优先级**：`--output` > `cli.output` > `table`

### 4.5 元数据注册表（meta_data.json）

**唯一规范**：[kuaimai-cli meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md)

| 概念 | 说明 |
|------|------|
| 存储层 | `internal/registry/meta_data.json` 嵌入二进制，驱动 `service` / `schema` |
| 展示层 | `kuaimai-cli schema` 从 meta 读取并格式化输出（meta 存数据，schema 查数据） |
| service 名 | `item`（erp-items-core，用 config `api.url`）· `scm`（erp-scm，用 meta `baseUrl`） |
| operation 命名 | `模块-操作` 小写短横线，如 `stock-list`、`staff-query`（**非** shortcuts 子命令名） |
| contentType | `get_query` · `post_form` · `post_json` |
| write | 查询 `false`；写接口 `true`；`--dry-run` 仅对 `write:true` 生效 |
| pageable | 分页列表 `true`；配合 `--page-all` 自动翻页；`--page-limit` / `--page-confirm` 控制海量数据防护 |

**shortcuts 与 meta operation 对照**（双轨：Agent 优先左列）：

| shortcuts | meta operation | contentType | write | pageable |
|-----------|----------------|-------------|-------|----------|
| `item +list` / `item list` | `stock-list` | post_form | false | true |
| `item count` | `stock-count` | post_form | false | false |
| （无） | `item-query-list-v2` | post_form | false | true |
| `item get-detail` | `item-detail` | get_query | false | false |
| `item save` | `item-save` | post_json | true | false |
| `item update-title`（编排） | `item-update-title`（原子 save） | post_json | true | false |

其余 **1089** 个 item operation 与 **195** 个 scm operation 均通过 `service item|scm <operation>` 调用。

```bash
kuaimai-cli service item stock-list --body '{"title":"test","pageNo":1,"pageSize":50}'
kuaimai-cli service scm staff-query --body '{"pageNo":1,"pageSize":20}'
kuaimai-cli schema --output json
```

### 4.6 当前业务 shortcuts 索引（erp-items-core）

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

完整对照表见 [每阶段新增能力 · 附录](./每阶段新增能力.md#附录shortcuts-与-meta-operation-对照表当前)。

### 4.7 Skill 平台（对标飞书 lark-*）

| 项 | 说明 |
|----|------|
| 仓库源 | `skills/kuaimai-shared/`、`skills/kuaimai-item/`、`skills/kuaimai-scm/` |
| 域结构 | 主 `SKILL.md` 路由 + `references/` 工作流文档 |
| 安装 | `skill install` → GitHub Contents API 递归拉取整目录 → 5 个 Agent 目录（默认 3 个 Skill） |
| Agent 约定 | 域 Skill **CRITICAL** 读 shared；写操作先 Read `references/` |

**新增/变更 Skill 步骤**：

1. 更新 `skills/<domain>/SKILL.md`（选命令、Shortcuts 表、API Resources）
2. 复杂流程写入 `skills/<domain>/references/*.md`
3. 若涉及全局约定，同步 `skills/kuaimai-shared/SKILL.md`
4. 本地验证后 push，用户侧执行 `skill install` 或重装

**新增 shortcuts 步骤**：

1. 在 `shortcuts/item/`（或新域目录）实现，使用 `common.Run*`  
2. `cmd/root.go` 中 `Register(rootCmd)`  
3. 按 [meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md) 更新 `internal/registry/meta_data.json`（operation 用 `模块-操作` 命名，填齐 path/method/contentType/write/pageable/requestSchema/responseSchema）  
4. 更新 `skills/kuaimai-item/SKILL.md`、对应 `references/`（如有新 shortcut）、本文档、[验收测试](./kuaimai-cli%20验收测试.md)

---

## 五、分阶段开发落地指南

### 阶段一：基建骨架版（~60%，已完成）

配置 · 鉴权 · `api` · 输出信封 · 友好错误

### 阶段二：标准企业版（~80%，已完成）

- `shortcuts/common/runner` + `RunGET`/`RunPOST`/`RunPOSTForm`  
- `schema` / `service` / `meta_data.json`（当前 **v1.7.0**，item 1095 + scm 195 operations）
- 重试 · `--page-all` · **`--page-limit`** · **`--page-confirm`** · completion · dry-run · 彩色 table

### 阶段三：平台进阶版（已完成，item 域）

- **erp-items-core shortcuts**：list/count/get-detail/save  
- **商品标题查改**（系统当前主要业务验收项）  
- Skill v2.0.0 · csv/ndjson · 连接池/熔断 · 脱敏 · form 分页 + 海量数据防护 · 审计日志

### 阶段四：AI 终局版（进行中）

| 能力 | 状态 |
|------|------|
| E2E 冒烟 `tests/cli_e2e` | ✅ |
| `upgrade` / `doctor` | ✅（`upgrade` 默认 npm 升级；`--check-only` 仅检查） |
| 多账号 `auth list` / `auth use` / `--profile` | ✅ |
| `item update-title` | ✅ |
| `auth check` | ✅ |
| README / CHANGELOG / Dockerfile / CI | ✅ |
| GitHub Release · npm · Skill（飞书风格 + references） | ✅（阶段四初已具备） |
| 更多 erp 业务域 shortcuts | 规划中（scm 已通过 meta + Skill + service 接入；item 按场景加 shortcut） |

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

**已对齐**：分层 · 三级命令 · shortcuts 手写 · registry/service · 密钥链 · 结构化 stdout · Skill 目录（含 `references/`）· csv/ndjson  

**差异**：业务域体量（当前 2 meta 域 vs 飞书 18+）· scm 无 shortcuts · 鉴权（accessToken vs OAuth）· 更多 curated shortcuts 待扩展  

**专属优势**：贴合 erp-items-core 真实路径、商品标题场景端到端可脚本化、维护面小  

---

## 八、开发自查清单

- [ ] 目录符合第三节规范  
- [ ] `cmd` 无业务逻辑  
- [ ] 新接口同步 `shortcuts/item` + `meta_data.json`（符合[定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md)）+ Skill + 文档  
- [ ] `api.url` 为 `https://erp1.superboss.cc/`  
- [ ] 三级命令可用：`item +list` / `service item stock-list` / `api POST`  
- [ ] 标题流程：`+list` → `update-title --dry-run` → `update-title` 可跑通  
- [ ] 错误无堆栈  
- [ ] `go build` / `go test ./...` 通过  

详细验收见 [kuaimai-cli 验收测试.md](./kuaimai-cli%20验收测试.md)。

---

## 九、总结

kuaimai-cli 采用「**骨架一次搭建、能力分阶段补齐**」：阶段一/二交付飞书式基建；**阶段三交付 erp-items-core 商品域**（标题查改闭环）；**scm 域**通过 meta + `service scm` + `kuaimai-scm` Skill 接入。**框架已定稿**（meta v1.7.0 + Skill + CLI 分页防护）；后续 primarily 登记 meta 与 curated shortcuts。

日常联调见仓库根 [README.md](../README.md)；元数据见 [meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md)；架构见 [系统架构与飞书对标说明](./系统架构与飞书对标说明.md)；文档索引见 [docs/README.md](./README.md)。
