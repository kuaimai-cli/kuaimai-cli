# 快麦 CLI 开发发布流程

本文只记录日常最常用的发布路径：本地编译检查、直接提交到 `main`、打 tag 发布新版本。

核心规则：

- 推送 `main` 只更新 GitHub 代码并触发 CI，不会发布 npm 新版本。
- 推送 `v*` tag 才会触发 `.github/workflows/release.yml`，自动创建 GitHub Release 并发布 npm。
- npm 包版本由 tag 自动写入。例如推送 `v0.2.5`，npm 发布版本就是 `0.2.5`，不需要手动修改 `npm/package.json`。
- 推送 GitHub 后查看代码：`https://github.com/kuaimai-cli/kuaimai-cli/commit/<commit-sha>`。
- 发布 npm 后查看自动化发布：`https://github.com/kuaimai-cli/kuaimai-cli/actions/workflows/release.yml`，进入最新的 tag run；最终 Release 地址是 `https://github.com/kuaimai-cli/kuaimai-cli/releases/tag/<tag>`，npm 地址是 `https://www.npmjs.com/package/@kuaimai-cli/cli/v/<version>`。
- 用户升级后需要重新安装 skill，命令见文档最后。

---

## 一、本地编译检查

每次准备提交前，先在项目根目录执行：

```bash
cd /Users/admin/Documents/project/kuaimai-cli

# 同步 skills 到 npm 包目录
node npm/scripts/sync-skills.js

# 编译检查
go build ./...

# 关键测试
go test ./shortcuts/erp-item ./shortcuts/scm-item ./cmd/...
go test ./internal/skill -run TestInstallDefaultsRemovesLegacyDefaultSkills
node --test npm/scripts/install-skills.test.js

# 检查 CLI 命令是否存在
go run . erp-item --help
go run . scm-item --help
```

如果只改了文档，可以不跑完整测试；如果改了命令、skill、npm 安装逻辑，建议完整执行上面的命令。

---

## 二、直接提交并推送到 main

确认本地编译检查通过后，直接提交到 `main`：

```bash
cd /Users/admin/Documents/project/kuaimai-cli

git status
git add -A
git commit -m "feat: 功能迭代更新"
git push origin main

# 查看刚推送到 GitHub 的 commit 链接
COMMIT_SHA="$(git rev-parse HEAD)"
echo "https://github.com/kuaimai-cli/kuaimai-cli/commit/${COMMIT_SHA}"

# 查看 main 分支 CI
echo "https://github.com/kuaimai-cli/kuaimai-cli/actions/workflows/ci.yml?query=branch%3Amain"
```

推送 `main` 后，GitHub 会跑 CI，但不会发布新版本。用户此时还拿不到 npm 新版本。

---

## 三、发布新版本

发布新版本必须打 tag。流程是：先查看 npm 线上版本和已有 tag，再递增一个新 tag。

```bash
cd /Users/admin/Documents/project/kuaimai-cli

# 确保 main 是最新代码
git checkout main
git pull origin main

# 查看 npm 线上版本
npm view @kuaimai-cli/cli version

# 查看已有 tag
git tag --list 'v*' --sort=-v:refname | head

# 递增一个新版本号，例如当前最新是 v0.2.5，则下一个可以是 v0.2.6
git tag v0.2.6
git push origin v0.2.6

# 查看自动化发布流水线
echo "https://github.com/kuaimai-cli/kuaimai-cli/actions/workflows/release.yml?query=branch%3Av0.2.6"

# 查看 GitHub Release
echo "https://github.com/kuaimai-cli/kuaimai-cli/releases/tag/v0.2.6"

# 查看 npm 包版本页
echo "https://www.npmjs.com/package/@kuaimai-cli/cli/v/0.2.6"
```

推送 tag 后，GitHub Actions 会自动执行 release 流水线。发布完成后检查 npm 版本：

```bash
npm view @kuaimai-cli/cli version
```

如果 npm 返回的新版本号等于刚才推送的 tag 去掉 `v` 后的版本号，说明发布成功。例如 tag 是 `v0.2.6`，npm 应该返回 `0.2.6`。

如果需要拿到本次自动化发布的精确 run 链接，可以在 tag 推送后执行：

```bash
gh run list \
  --workflow release.yml \
  --branch v0.2.6 \
  --limit 1 \
  --json url,status,conclusion,displayTitle
```

`url` 字段就是本次自动化发布链接。

---

## 四、发版规则说明

### npm 版本号

正常发版不需要手动修改 `npm/package.json` 的 `version`。

`release.yml` 会自动执行：

```bash
VERSION="${GITHUB_REF_NAME#v}"
pkg.version = VERSION
```

所以版本号以 tag 为准。

### skill 更新

默认 skill 安装会先删除旧默认 skill，再安装当前默认 skill：

```text
删除：kuaimai-shared、kuaimai-erp-item、kuaimai-scm-item、kuaimai-item、kuaimai-scm
安装：kuaimai-shared、kuaimai-erp-item、kuaimai-scm-item
```

这样可以避免旧版本用户升级后继续使用历史 `kuaimai-item` / `kuaimai-scm` 路由。

---

## 五、用户升级命令

新版本发布成功后，用户使用下面命令升级：

```bash
npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install --force
```

升级后建议重新打开 Agent 会话，确保读取到最新 skill。
