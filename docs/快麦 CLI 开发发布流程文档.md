# 快麦 CLI 开发发布流程

> **读者**：研发、维护者、版本发布负责人。  
> **用途**：统一「本地开发 → 自测 → 合并主干 → 查线上版本 → 打 Tag 发版 → 观测 Release → 本机升级 / 故障重置」全流程，杜绝版本重复、漏发、漏更新、Skill 缓存残留等问题。  
> **相关**：[极简命令大全](./快麦%20CLI%20极简可运行命令大全.md) · [开发文档](./kuaimai-cli%20开发文档.md) · [meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md) · [验收测试](./kuaimai-cli%20验收测试.md) · [文档索引](./README.md)

---

## 核心铁律（全文最重要）

1. **普通代码 push / PR 合并只走 CI，不发版**
2. **必须先查线上最新版本，再打新 Tag，禁止凭记忆发版**
3. **只有 `git push origin vX.Y.Z` 才会触发 Release + npm 正式发包**

---

## 文档说明

本文档对齐当前 kuaimai-cli 工程现状：

| 机制 | 说明 |
|------|------|
| **日常推送** | 推分支 / 合并 `main` → 触发 **CI**（单元测试 + E2E），**不**发版 |
| **版本发布** | 仅推送 **`v*` Tag** → 触发 **Release**（GoReleaser + npm） |
| **用户安装** | `npx @kuaimai-cli/cli@latest install`（对标飞书 `npx @larksuite/cli install`） |
| **Skill** | `skill install` 从 GitHub `main` 经 Contents API 拉取整目录（`SKILL.md` + `references/`），复制到 **5 个 Agent 目录** |
| **升级** | `upgrade` 仅对比版本并提示；**须重装**才能替换本机二进制 |

```text
开发迭代 → 本地全量自检(§一) → 分支 PR + CI 绿(§二) → 核查线上版本(§三)
    → 打 Tag 推送发版(§四) → 观测 Release(§五) → 本机升级/重置(§六/§七)
```

---

## 一、本地开发 & 自测（开发阶段 100% 必做）

代码开发、联调完毕后，**先本地通过再提交**。在仓库根目录执行。

### 1.1 代码开发规范

- 所有功能、修复、优化必须**新建分支**开发，禁止直接在 `main` 提交
- Commit 规范：`feat/fix/docs/refactor(模块): 描述`（如 `feat(item): 新增 count shortcut`）

### 1.2 编译校验（对齐 CI）

```bash
# 推荐：与 CI 完全一致（自动 fetch meta）
make build

# 快速编译
go build -o kuaimai-cli .
```

**观测**：

```bash
./kuaimai-cli --version
./kuaimai-cli doctor --output json
```

**合格标准**：编译无报错、退出码 `0`；`doctor` 输出 `ready: true` 或明确的 `next` 步骤。

### 1.3 单元测试 + E2E 冒烟测试（和 CI 一致）

```bash
# meta 变更必须执行（CI 同步骤；本地改 meta 时用 generate_meta/filter_meta/normalize_meta）
./scripts/fetch_meta/fetch_meta.sh

# 分页逻辑变更时建议单独跑
go test -mod=vendor ./internal/pagination/...

# 全量单元测试
go test -mod=vendor ./...

# E2E 自动化测试
go test -mod=vendor -v ./tests/cli_e2e/...
```

依赖变更先执行：`make vendor`

**合格标准**：全部 `PASS`，无 `FAIL`、无 `panic`。

### 1.4 业务功能自测（按改动范围）

查询类：

```bash
./kuaimai-cli auth check --output json
./kuaimai-cli item +list \
  --body '{"title":"测试","pageNo":1,"pageSize":1}' \
  --output json
```

写操作必须带 `--dry-run`：

```bash
./kuaimai-cli item update-title --sys-item-id <id> --title "新标题" --dry-run --verbose
```

### 1.5 改动联动文件同步（关键）

只要改接口、字段、分页、业务逻辑、meta，必须同步更新：

- `shortcuts/item/`（或对应域）业务命令逻辑
- `internal/registry/meta_data.json`（见 [meta_data.json 定义规范](./kuaimai-cli%20meta_data.json%20定义规范.md)；可用 `scripts/generate_meta/` + `scripts/filter_meta/` + `scripts/normalize_meta/` 再生成）
- 分页逻辑变更：`internal/pagination/`
- Skill 文档：`skills/*/SKILL.md`、`references/`
- 相关配套文档（[AGENTS.md](../AGENTS.md)、[验收测试](./kuaimai-cli%20验收测试.md) 等）

完整清单见 [验收测试](./kuaimai-cli%20验收测试.md)。

---

## 二、代码提交、PR、合并主干

本地验证通过后提交，**建议走 PR 合并 `main`**，避免未测代码直接进入主干。

```bash
git status
git add <改动的文件>
git commit -m "feat(item): 本次变更说明"
git push -u origin <分支名>
```

在 GitHub 创建 Pull Request → 等待 **CI** 全部绿色 → 合并。

**合并 `main` 不会发版**，仅做代码集成。

### 观测 CI（普通 push / PR）

Workflow：`.github/workflows/ci.yml`  
地址：<https://github.com/kuaimai-cli/kuaimai-cli/actions>

| 步骤 | 内容 | 成功标志 |
|------|------|----------|
| Fetch meta | `./scripts/fetch_meta/fetch_meta.sh` | ✓ |
| Unit tests | `go test -mod=vendor ./...` | ✓ |
| E2E smoke | `tests/cli_e2e` | ✓ |

```bash
gh run list --workflow=ci.yml --limit 5
gh run watch
```

**要点**：仅推代码 **不会** 触发 Release / npm 发布；CI 红则先修再合，**CI 不绿禁止合并主干、禁止发版**。

---

## 三、版本发布前置：先查线上版本

需要让 **npm / GitHub Release 用户** 拿到新版本时再执行本节。

> **发版的最大禁忌：凭记忆打 Tag。** 发版前必须固定执行以下 3 条查询命令。

### 3.1 查询线上最新版本（绝对必须）

```bash
# 1. 查询 GitHub 线上正式 Release（最权威）
gh release list --limit 3

# 2. 查询 npm 线上安装版本（用户实际安装版本）
npm view @kuaimai-cli/cli version

# 3. 查看本机当前版本做对比
kuaimai-cli --version
```

### 3.2 版本号递增规则（唯一标准）

| 变更类型 | 示例（线上当前 `v0.1.2`） |
|----------|---------------------------|
| 小修复、微调 | 补丁版本 → `v0.1.3` |
| 新增功能、新增 shortcut、新增能力 | 次版本 → `v0.2.0` |
| 架构大改、破坏性变更 | 主版本 → `v1.0.0` |

**规则：只升不降、绝不重复、绝不回退。**

### 3.3 发版前置检查清单（必须全部满足）

- [ ] `main` 分支已合并、代码最新
- [ ] 最新 `main` CI 全绿
- [ ] [CHANGELOG.md](../CHANGELOG.md) 已更新版本与日期（将 `Unreleased` 归入 `vX.Y.Z`）
- [ ] 已核查线上版本，新版本号确定合法、不重复
- [ ] 浏览 [.goreleaser.yaml](../.goreleaser.yaml) 中 `release.header` / `footer` 是否需要随大版本更新能力快照
- [ ] 可选：对照 [.github/RELEASE_TEMPLATE.md](../.github/RELEASE_TEMPLATE.md) 在 GitHub Release 页补全「本版亮点」

---

## 四、正式打 Tag + 触发 Release 发版

### 4.1 标准发版命令（固定模板）

```bash
# 1. 切主干、更新最新代码
git checkout main
git pull origin main

# 2. 再次确认线上版本（兜底）
gh release list --limit 3
npm view @kuaimai-cli/cli version

# 3. 打新版 Tag（替换为你确定的新版本）
git tag v0.1.3

# 4. 推送 Tag 触发 Release + npm 发布流水线
git push origin v0.1.3
```

### 4.2 多端版本对齐规则

| 位置 | 示例 |
|------|------|
| Git tag | `v0.1.3` |
| GitHub Release | `v0.1.3` |
| npm `@kuaimai-cli/cli` | `0.1.3`（CI 从 tag 去掉 `v` 写入） |
| 二进制资产 | `kuaimai-cli-0.1.3-darwin-arm64.tar.gz` 等 |

---

## 五、观测 Release 发布结果

推送 Tag 后触发 `.github/workflows/release.yml`，共 **三个 job**，须全部绿灯：

| Job | 作用 | 成功标志 |
|-----|------|----------|
| **goreleaser** | 编译 **5 平台**包，创建 GitHub Release，上传 `checksums.txt` | Release 页有完整资产 |
| **npm-checksums** | 下载 `checksums.txt` 到 `npm/` | artifact 上传成功 |
| **npm-publish** | 同步 `npm/package.json` 并 `npm publish` | 见下方说明 |

> **npm 发布守卫**：`npm-publish` job 仅在官方仓库 `kuaimai-cli/kuaimai-cli` 执行（Fork 推 Tag 不会发 npm）。须配置 Secret `NPM_TOKEN`。

### 5.1 查看发布状态

```bash
gh run list --workflow=release.yml --limit 3
gh release view v0.1.3
npm view @kuaimai-cli/cli version
```

### 5.2 发布成功判定

**Release 资产应包含**（与 `npm/scripts/install.js` 一致；`.goreleaser.yaml` 忽略 `windows/arm64`，故为 **5 平台包 + checksums.txt**）：

```text
kuaimai-cli-0.1.3-darwin-arm64.tar.gz
kuaimai-cli-0.1.3-darwin-amd64.tar.gz
kuaimai-cli-0.1.3-linux-amd64.tar.gz
kuaimai-cli-0.1.3-linux-arm64.tar.gz
kuaimai-cli-0.1.3-windows-amd64.zip
checksums.txt
```

- 三个 job 全绿
- npm 版本号与发布 Tag 一致（去掉 `v`）

**Release 正文**：由 GoReleaser 根据 [.goreleaser.yaml](../.goreleaser.yaml) 的 `release.header` / `footer` 自动生成（含安装说明、能力快照、Full Changelog 链接）。维护者发版前请更新 [CHANGELOG.md](../CHANGELOG.md)；在 GitHub 网页手动编辑 Release 时可参考 [.github/RELEASE_TEMPLATE.md](../.github/RELEASE_TEMPLATE.md)。

### 5.3 发布失败处理

- **同 Tag 不可重复推送**（GoReleaser 会跳过）
- 失败需删除远端旧 Tag，修正问题后**升版本**重打 Tag

---

## 六、本机升级最新版本（发布后必做）

> **重要**：`upgrade` 仅检测版本，**不会**自动更新二进制，必须手动重装。

### 6.1 标准升级流程（推荐）

```bash
# 检测新版本
kuaimai-cli upgrade --output json

# 重装最新版
npx @kuaimai-cli/cli@latest install
# 或：npm install -g @kuaimai-cli/cli@latest

# 更新 Skill（技能变更必执行）
kuaimai-cli skill install

# 最终验收四连
kuaimai-cli --version
kuaimai-cli skill list --output json
kuaimai-cli doctor --output json
kuaimai-cli auth check --output json
```

**观测 `upgrade --output json`**：

- `data.current`：本机版本
- `data.latest`：GitHub 最新 Release
- `data.update_available`：为 `true` 时需执行上面的重装命令

### 6.2 开发者本地源码更新（未发版场景）

```bash
cd /path/to/kuaimai-cli
git pull origin main
make build
cp ./kuaimai-cli ~/bin/kuaimai-cli
kuaimai-cli --version
```

### 6.3 内网压缩包（不经 npm）

维护者在仓库根目录打包：

```bash
make dist
# 产物在 dist/ 目录，含 linux/darwin/windows 各架构二进制
```

接收方覆盖 PATH 中的二进制，例如：

```bash
cp dist/kuaimai-cli-darwin-arm64 ~/bin/kuaimai-cli
chmod +x ~/bin/kuaimai-cli
kuaimai-cli --version
```

---

## 七、本地异常干净重置（版本错乱 / Skill 缓存）

出现版本不生效、Skill 残留、命令行为与 `--version` 不一致时，执行全量清理后重装。

> **Skill 机制**：`skill install` 会把 `kuaimai-shared`、`kuaimai-item` **整目录**（含 `references/`）复制到 5 个 Agent 目录。只删其中一个目录会导致 Agent 仍加载旧 Skill。

### 7.1 全量清理旧缓存

```bash
# 清空所有 Agent 旧 Skill（必做）
rm -rf ~/.agents/skills/kuaimai-*
rm -rf ~/.cursor/skills/kuaimai-*
rm -rf ~/.claude/skills/kuaimai-*
rm -rf ~/.codex/skills/kuaimai-*
rm -rf ~/.windsurf/skills/kuaimai-*

# 可选：清空配置与审计日志（不会删除密钥链 token）
rm -rf ~/.kuaimai-cli
```

> **注意**：`accessToken` 存在**系统密钥链**，删 `~/.kuaimai-cli` **不会**清除登录态。若需重新初始化配置，再执行 `kuaimai-cli config init`；token 失效时再 `auth login`。

### 7.2 重装验收

```bash
npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install

kuaimai-cli doctor --output json
kuaimai-cli auth status --output json
kuaimai-cli auth check --output json
```

**重装后必须重启 AI Agent 会话**以加载新 Skill。

---

## 八、发布铁律（十条红线，禁止违反）

1. 代码 push / PR 合并**绝对不会发版**，只有 Tag 推送发版
2. 发版前**必须查线上版本**，禁止凭记忆写版本号
3. 版本号**只升不降、不重复、不回退**
4. **CI 不绿禁止合并主干、禁止发版**
5. 写功能必须自测 `--dry-run`
6. 改 meta / 接口 / 逻辑必须同步更新 Skill、文档、用例
7. **`upgrade` 只检查、不升级**，新版必须重装
8. Skill 更新须清理**全部五个** Agent 目录
9. 每次发版必须更新 [CHANGELOG.md](../CHANGELOG.md)
10. 发布后必须本机验收 `version` / `doctor` / `auth check`

---

## 九、一键复制：完整速查脚本

```bash
# 1. 开发自测全流程
make build && go test -mod=vendor ./... && go test -mod=vendor ./tests/cli_e2e/...

# 2. 提交推送（示例）
git push -u origin my-branch   # PR → CI 绿 → 合并 main

# 3. 发版前置核查
git checkout main && git pull origin main
gh release list --limit 3
npm view @kuaimai-cli/cli version

# 4. 打 tag 发布（替换为确定的新版本号）
git tag v0.1.3
git push origin v0.1.3

# 5. 观测发布
gh run list --workflow=release.yml --limit 3
npm view @kuaimai-cli/cli version

# 6. 本机升级验收
kuaimai-cli upgrade --output json
npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install
kuaimai-cli doctor --output json
```
