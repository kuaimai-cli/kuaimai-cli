# 快麦 CLI 开发发布流程

本文面向第一次发布的人，按“本地检查 → 提交 main → 打 tag → GitHub Actions 发布 → 用户升级验证”的顺序执行。

核心规则：

- **推送 main 只触发 CI，不发布新版。**
- **推送 `v*` tag 才触发 GitHub Release 与 npm 发布。**
- **npm 包版本由 release workflow 自动从 tag 写入，不需要手动改 `npm/package.json`。**
- **默认 Skills 安装/自动同步会删除旧默认 Skill，再安装当前默认 Skill。**

---

## 一、新手发版最短路径

以下流程用于功能已开发完成、准备发布一个正式版本时执行。

```bash
cd /Users/admin/Documents/project/kuaimai-cli

# 1. 同步 skills 到 npm 包
node npm/scripts/sync-skills.js

# 2. 编译
go build ./...

# 3. 跑不依赖本地监听端口的关键测试
go test ./shortcuts/erp-item ./shortcuts/scm-item ./cmd/...
go test ./internal/skill -run TestInstallDefaultsRemovesLegacyDefaultSkills
node --test npm/scripts/install-skills.test.js

# 4. 检查命令是否存在
go run . erp-item --help
go run . scm-item --help

# 5. 整理提交
git status
git add -A
git commit -m "feat: add scm item publishing shortcuts"
git push origin main

# 6. 发布新版本：先确认最新 tag，再打下一个版本号
git tag --list 'v*' --sort=-v:refname | head
git tag v0.2.5
git push origin v0.2.5

# 7. 查看发布流水线和 npm 版本
gh run list --repo kuaimai-cli/kuaimai-cli --workflow=release.yml --limit 3
npm view @kuaimai-cli/cli version
```

本地 `go test ./...` 也应该在 CI 中通过；如果本地沙箱禁止 `httptest` 监听端口，可能出现 `failed to listen on a port`，此时以 GitHub Actions CI 为准。

---

## 二、本地开发编译 & 自测

```bash
cd /Users/admin/Documents/project/kuaimai-cli

# 1. 对齐 CI 编译（推荐）
make build

# 2. 快速本地编译
go build -mod=vendor -o kuaimai-cli .

# 3. 编译校验
./kuaimai-cli --version
./kuaimai-cli -h
./kuaimai-cli doctor --output json

# 4. 单元测试 + E2E
go test -mod=vendor ./...
go test -mod=vendor -v ./tests/cli_e2e/...

# 5. Registry 链路自测（mock，E2E 已覆盖；本地可手动）
./kuaimai-cli registry sync --output json
./kuaimai-cli capabilities --output json
./kuaimai-cli schema api.luotao.test.get --output json

# 6. Skill 同步到 npm 包（发版前必做）
node npm/scripts/sync-skills.js
```

### 业务自测（需有效 token + 网络）

```bash
./kuaimai-cli auth check --output json
./kuaimai-cli erp-item +list --body '{"title":"测试","pageNo":1,"pageSize":1}' --output json
./kuaimai-cli erp-item update-title --sys-item-id 10001 --title "测试更新" --dry-run --verbose
./kuaimai-cli web call api.luotao.test.get --params '{"keyword":"测试"}' --output json
```

### 本地开发注意

- 使用 `./kuaimai-cli`，避免与全局旧版混用
- registry 默认从 `http://open-cli.kuaimai.com/registry/registry.json` 同步
- 修改 `skills/` 后执行 `node npm/scripts/sync-skills.js` 再发 npm 包
- `service` 命令已移除；registry 接口统一 `web call <apiId>`

---

## 三、完整分支开发流程

分支示例：`dev-test`

```bash
cd /Users/admin/Documents/project/kuaimai-cli

git checkout main && git pull origin main
git checkout -b dev-test

# 开发、自测通过后提交
git add -A
git reset -- scripts/generate_meta/__pycache__/ scripts/filter_meta/__pycache__/
git commit -m "feat: 开发分支功能迭代"
git push -u origin dev-test

# 合并回 main
git checkout main && git pull origin main
git merge dev-test
git push origin main

# 可选：删除开发分支
git branch -d dev-test
git push origin --delete dev-test
```

---

## 四、代码直接推送 main

```bash
cd /Users/admin/Documents/project/kuaimai-cli

git add -A
git reset -- scripts/generate_meta/__pycache__/ scripts/filter_meta/__pycache__/
git commit -m "feat: 功能迭代更新"
git push origin main

gh run list --repo kuaimai-cli/kuaimai-cli --workflow=ci.yml --limit 3
gh run watch
```

**规则**：推 main 只跑 CI，不发版；**仅 Tag 触发 Release & NPM**。

### 标准 PR 流程

```bash
git checkout main && git pull origin main
git checkout -b feat/temp-feature
git add -A
git commit -m "feat: 迭代新功能"
git push -u origin feat/temp-feature

gh pr create --repo kuaimai-cli/kuaimai-cli \
  --base main --head feat/temp-feature \
  --title "feat: 功能迭代更新" \
  --body "$(cat <<'EOF'
## Summary
- 本次功能迭代与优化

## Test plan
- make build
- go test -mod=vendor ./...
- ./kuaimai-cli capabilities --output json
EOF
)"
```

---

## 五、发版 → 用户可用（必走）

**关键**：仅推送 main，用户无法获取新版；必须打 Tag 触发 `release.yml` 发布 NPM。

```bash
cd /Users/admin/Documents/project/kuaimai-cli
git checkout main && git pull origin main

# 查线上版本和已有 Tag，递增打 Tag
npm view @kuaimai-cli/cli version
git tag --list 'v*' --sort=-v:refname | head
git tag v下一版本号
git push origin v下一版本号

# 等待 1～3 分钟后校验
npm view @kuaimai-cli/cli version
gh run list --repo kuaimai-cli/kuaimai-cli --workflow=release.yml --limit 3
```

发版前检查清单：

- [ ] `go test -mod=vendor ./...` 通过
- [ ] `node npm/scripts/sync-skills.js` 已执行（`npm/skills/` 与仓库 `skills/` 一致）
- [ ] `CHANGELOG.md` 已更新
- [ ] Tag 版本号高于 npm 线上版本

### npm 版本号规则

`release.yml` 会在 `npm-publish` job 中自动把 tag 写入 `npm/package.json`：

```bash
VERSION="${GITHUB_REF_NAME#v}"
pkg.version = VERSION
```

因此正常发版时 **不需要手动修改 `npm/package.json` 的 version**；以 tag 为准，例如推送 `v0.2.5` 时，npm 发布包版本会自动变成 `0.2.5`。  
只有手动在本地执行 `npm publish` 时，才需要先手动同步 `npm/package.json` 版本。正常发布必须走 `git push origin v下一版本号`，不要直接在本地 `npm publish`。

### Skill 重装规则

默认 Skill 安装/自动同步会先删除旧默认 Skill 目录，再安装当前默认 Skills：

```text
删除：kuaimai-shared、kuaimai-erp-item、kuaimai-scm-item、kuaimai-item、kuaimai-scm
安装：kuaimai-shared、kuaimai-erp-item、kuaimai-scm-item
```

这可以避免旧版本用户升级后，Agent 继续读取历史 `kuaimai-item` / `kuaimai-scm` 路由。用户升级后建议执行：

```bash
kuaimai-cli skill install --force
kuaimai-cli doctor --output json
```

用户更新：

```bash
npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install --force
```

### 发版失败重置 Tag

```bash
git push origin --delete v错误版本号
git tag -d v错误版本号
# 修正后升版本重发
git tag v新版本号
git push origin v新版本号
```

---

## 六、开发人员本机更新

```bash
kuaimai-cli upgrade
npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install --force

# 本地源码调试
cd /Users/admin/Documents/project/kuaimai-cli
git pull origin main
make build
./kuaimai-cli doctor --output json
```

**不要**删除 `~/.kuaimai-cli/`（会丢失 `config.yaml` 与 token）。

---

## 七、普通用户安装

```bash
npx @kuaimai-cli/cli@latest install
kuaimai-cli config init
kuaimai-cli auth login "<accessToken>"
kuaimai-cli doctor --output json
kuaimai-cli registry sync --output json
```

---

## 八、闭环链路（牢记）

```text
代码修改 & 推送 main（CI）
    → 打 Tag 发版（GitHub Release + npm）
    → node npm/scripts/sync-skills.js（发版流水线内或手动）
    → 用户 npx install + skill install
    → Agent Read kuaimai-shared → capabilities → schema → web call
```

配套文档：[Agent 安装指南](./快麦%20CLI%20安装（Agent%20专用）.md) · [系统架构说明](./系统架构与飞书对标说明.md) · [Registry 同步](./Registry远端同步说明.md)
