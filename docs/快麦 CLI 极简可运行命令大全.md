# 快麦 CLI 极简可运行命令大全

> 开发 / 推送 / 发版 / 安装全拆分。命令直接复制即用。  
> 详细说明见 [开发发布流程](./快麦%20CLI%20开发发布流程文档.md)。

---

## 一、本地开发编译 & 自测（开发人员）

```bash
# 1. 编译（对齐 CI，推荐）
make build

# 2. 快速编译
go build -o kuaimai-cli .

# 3. 校验编译成功
./kuaimai-cli --version
./kuaimai-cli doctor --output json

# 4. 全量测试（单元 + E2E）
./scripts/fetch_meta/fetch_meta.sh
go test -mod=vendor ./...
go test -mod=vendor -v ./tests/cli_e2e/...

# 5. 业务自测
./kuaimai-cli auth check --output json
./kuaimai-cli item +list --body '{"title":"测试","pageNo":1,"pageSize":1}' --output json
./kuaimai-cli item update-title --sys-item-id xxx --title "测试" --dry-run --verbose
```

---

## 二、代码推送 GitHub（日常开发 / PR）

> 仓库：`https://github.com/kuaimai-cli/kuaimai-cli` · 默认分支：`main` · 推分支/PR **只触发 CI**，**不会发版**。

```bash
# ── 0. 从最新 main 拉功能分支（禁止直接在 main 开发）──
git checkout main
git pull origin main
git checkout -b feat/item-count          # 示例：feat/fix/docs/refactor(模块)-简述

# ── 1. 本地自测通过后提交（§一 make build + go test 全绿）──
git status                               # 确认改动文件
git add shortcuts/item/ docs/            # 按需指定路径；勿提交 __pycache__、.env
git commit -m "feat(item): 新增 count shortcut"

# ── 2. 推送到 origin 并建立 upstream ──
git push -u origin feat/item-count

# ── 3. 创建 PR 合并到 main（二选一）──
# 方式 A：GitHub 网页 → Compare & pull request
# 方式 B：CLI（在仓库根目录，需 gh auth login）
gh pr create --repo kuaimai-cli/kuaimai-cli \
  --base main \
  --head feat/item-count \
  --title "feat(item): 新增 count shortcut" \
  --body "## Summary\n- 新增 item count shortcut\n\n## Test plan\n- [ ] make build\n- [ ] go test -mod=vendor ./..."

# ── 4. 观测 CI（workflow: .github/workflows/ci.yml）──
# 步骤：fetch_meta → unit tests → E2E smoke
gh run list --repo kuaimai-cli/kuaimai-cli --workflow=ci.yml --limit 5
gh run watch                               # 跟踪当前仓库最近一次 run；Ctrl+C 退出

# 浏览器：<https://github.com/kuaimai-cli/kuaimai-cli/actions/workflows/ci.yml>

# ── 5. CI 全绿后合并 PR，本地同步 main ──
git checkout main
git pull origin main
git branch -d feat/item-count              # 可选：删除已合并的本地分支
```

---

## 三、版本发版 & NPM 发布（仅版本迭代）

> **铁律**：只有 `git push origin vX.Y.Z` 才会触发 `.github/workflows/release.yml`（GoReleaser + npm）。  
> **npm 仅官方仓库** `kuaimai-cli/kuaimai-cli` 会 publish（Fork 推 Tag 不会发 npm）。

### 3.1 发版前：查线上版本（禁止凭记忆打 Tag）

```bash
# GitHub 正式 Release（最权威；本地 tag 列表：v0.1.0 … v0.1.3）
gh release list --repo kuaimai-cli/kuaimai-cli --limit 5

# npm 用户实际安装版本（Tag 去掉 v，如 v0.1.3 → 0.1.3）
npm view @kuaimai-cli/cli version

# 本机已安装版本对比
kuaimai-cli --version

# 发版前须把 CHANGELOG.md 中 Unreleased 归入新版本并提交到 main
# 版本递增：小修复 v0.1.3→v0.1.4 · 新功能 v0.1.3→v0.2.0 · 破坏性 v1.0.0
```

### 3.2 打 Tag 触发 Release + npm

```bash
git checkout main
git pull origin main

# 再次兜底确认线上最新版（示例：若线上为 v0.1.3，小修复则打 v0.1.4）
gh release list --repo kuaimai-cli/kuaimai-cli --limit 3
npm view @kuaimai-cli/cli version

git tag v0.1.4                           # 替换为你确定的新版本号
git push origin v0.1.4                   # 推送 Tag → 触发 release.yml 三个 job
```

Release 流水线 job 顺序：`goreleaser`（5 平台包 + checksums.txt + GitHub Release）→ `npm-checksums` → `npm-publish`。

### 3.3 观测发布结果

```bash
gh run list --repo kuaimai-cli/kuaimai-cli --workflow=release.yml --limit 3
gh release view v0.1.4 --repo kuaimai-cli/kuaimai-cli
npm view @kuaimai-cli/cli version        # 应与 Tag 一致（无 v 前缀）

# 浏览器：<https://github.com/kuaimai-cli/kuaimai-cli/actions/workflows/release.yml>
# Release 页：<https://github.com/kuaimai-cli/kuaimai-cli/releases>
```

成功标志：三个 job 全绿；Release 含 5 平台包 + `checksums.txt`；npm 版本与 Tag 对齐。

### 3.4 发版失败：删 Tag 后升版本重发

```bash
# 同一 Tag 不可重复推送（GoReleaser 会跳过）；须删 Tag、修问题、换新版本号
git push origin --delete v0.1.4            # 删远端 Tag
git tag -d v0.1.4                          # 删本地 Tag

# 修正 CHANGELOG / 代码后，用更高版本重打，例如 v0.1.5
git tag v0.1.5
git push origin v0.1.5
```

---

## 四、开发人员本机更新 / 重置

```bash
# 1. 检测新版本
kuaimai-cli upgrade --output json

# 2. 重装最新版（须手动重装才更新）
npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install

# 3. 本地源码更新（未发版调试）
git pull origin main
make build
cp ./kuaimai-cli ~/bin/kuaimai-cli

# 4. 全局异常重置
rm -rf ~/.agents/skills/kuaimai-*
rm -rf ~/.cursor/skills/kuaimai-*
rm -rf ~/.claude/skills/kuaimai-*
rm -rf ~/.codex/skills/kuaimai-*
rm -rf ~/.windsurf/skills/kuaimai-*

npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install
kuaimai-cli doctor --output json
```

---

## 五、普通用户安装 & 使用

```bash
# 1. 全新安装
npx @kuaimai-cli/cli@latest install
# 或
npm install -g @kuaimai-cli/cli@latest

# 2. 版本 & 健康检测
kuaimai-cli --version
kuaimai-cli doctor --output json

# 3. 权限 & 技能
kuaimai-cli auth check --output json
kuaimai-cli skill install

# 4. 检测新版本
kuaimai-cli upgrade --output json
```
