快麦 CLI 极简可运行命令大全（100% 真实可落地）
全部命令无占位、可直接复制运行｜开发 / 推送 / PR / 发版 / 重置 / 用户安装 全流程闭环
详细规范见：开发发布流程

---

一、本地开发编译 & 自测（开发人员专用）

# 1. 对齐CI编译（推荐正式打包）

make build

# 2. 快速本地编译

go build -o kuaimai-cli .

# 3. 编译结果校验

./kuaimai-cli --version
./kuaimai-cli doctor --output json

# 4. 全量元数据+单元测试+E2E测试

./scripts/fetch_meta/fetch_meta.sh
go test -mod=vendor ./...
go test -mod=vendor -v ./tests/cli_e2e/...

# 5. 核心业务自测（可直接跑）

./kuaimai-cli auth check --output json
./kuaimai-cli item +list --body '{"title":"测试","pageNo":1,"pageSize":1}' --output json
./kuaimai-cli item update-title --sys-item-id 10001 --title "测试更新" --dry-run --verbose

---

二、代码提交 & 推送 GitHub（真实可运行）
核心规则：推 main 只跑 CI，不发版；仅 Tag 触发 Release & NPM
项目固定路径：/Users/admin/Documents/project/kuaimai-cli
2.1 日常快速推送 main（最常用）
cd /Users/admin/Documents/project/kuaimai-cli

git status
git diff --stat

# 提交所有改动、自动剔除python缓存

git add -A
git reset -- scripts/generate_meta/**pycache**/ scripts/filter_meta/**pycache**/

# 常规单行提交（直接改文字即可）

git commit -m "feat(item): 功能迭代更新"

# 推送主干

git push origin main

# 查看CI状态

gh run list --repo kuaimai-cli/kuaimai-cli --workflow=ci.yml --limit 3
gh run watch

2.2 标准分支 PR 流程（CodeReview 专用）
cd /Users/admin/Documents/project/kuaimai-cli

# 同步主干最新代码

git checkout main && git pull origin main

# 新建开发分支

git checkout -b feat/temp-feature

# 提交改动

git add -A
git reset -- scripts/generate_meta/**pycache**/ scripts/filter_meta/**pycache**/
git commit -m "feat: 迭代新功能"
git push -u origin feat/temp-feature

# 快速创建PR

gh pr create --repo kuaimai-cli/kuaimai-cli   
  --base main --head feat/temp-feature   
  --title "feat: 功能迭代更新"   
  --body "## Summary

- 本次功能迭代与优化

## Test plan

- make build
- go test -mod=vendor ./..."

# 监控CI

gh run list --repo kuaimai-cli/kuaimai-cli --workflow=ci.yml --limit 5
gh run watch

# PR合并后同步本地主干

git checkout main && git pull origin main

---

三、版本发版 & NPM 发布（唯一正式发版流程）
发版铁律：必须先查线上版本 → 再打新版Tag → 禁止重复/倒退
3.1 发版前置：查询线上真实版本（必跑）

# 查GitHub线上正式Release版本

gh release list --repo kuaimai-cli/kuaimai-cli --limit 5

# 查NPM用户实际安装版本

npm view @kuaimai-cli/cli version

# 对比本机版本

kuaimai-cli --version

3.2 正式打Tag发版（真实可运行模板）
示例：线上最新 v0.1.3，本次小修复升级为 v0.1.4
cd /Users/admin/Documents/project/kuaimai-cli
git checkout main && git pull origin main

# 兜底校验线上版本

gh release list --repo kuaimai-cli/kuaimai-cli --limit 3
npm view @kuaimai-cli/cli version

# 打新版标签 + 推送触发发布流水线

git tag v0.1.4
git push origin v0.1.4

3.3 观测 Release 发布结果

# 查看release流水线执行状态

gh run list --repo kuaimai-cli/kuaimai-cli --workflow=release.yml --limit 3

# 查看本次发布详情

gh release view v0.1.4 --repo kuaimai-cli/kuaimai-cli

# 校验NPM是否发布成功

npm view @kuaimai-cli/cli version

3.4 发版失败重置（Tag冲突/报错专用）

# 删除本地+远端错误Tag

git push origin --delete v0.1.4
git tag -d v0.1.4

# 修正代码/日志后，升版本重发示例

git tag v0.1.5
git push origin v0.1.5

---

四、开发人员本机更新 & 异常重置

# 1. 检测是否有新版本

kuaimai-cli upgrade --output json

# 2. 在线重装最新正式版

npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install

# 3. 本地源码更新（未发版调试用）

cd /Users/admin/Documents/project/kuaimai-cli
git pull origin main
make build
cp ./kuaimai-cli ~/bin/kuaimai-cli

# 4. 全局异常清理（版本错乱/缓存残留必杀）

rm -rf ~/.agents/skills/kuaimai-*
rm -rf ~/.cursor/skills/kuaimai-*
rm -rf ~/.claude/skills/kuaimai-*
rm -rf ~/.codex/skills/kuaimai-*
rm -rf ~/.windsurf/skills/kuaimai-*

# 清理后重装验收

npx @kuaimai-cli/cli@latest install
kuaimai-cli skill install
kuaimai-cli doctor --output json

---

五、普通用户：安装 & 日常使用命令

# 1. 一键安装最新版

npx @kuaimai-cli/cli@latest install

# 或全局安装

npm install -g @kuaimai-cli/cli@latest

# 2. 基础校验

kuaimai-cli --version
kuaimai-cli doctor --output json

# 3. 权限校验 & 技能同步

kuaimai-cli auth check --output json
kuaimai-cli skill install

# 4. 检测新版本

kuaimai-cli upgrade --output json

---

六、固定发版完整流水线（终极一键模板）
每次发版直接改版本号即可，全真实可用
cd /Users/admin/Documents/project/kuaimai-cli
git checkout main && git pull origin main
gh release list --repo kuaimai-cli/kuaimai-cli --limit 3
npm view @kuaimai-cli/cli version
git tag v0.1.4
git push origin v0.1.4
gh run list --repo kuaimai-cli/kuaimai-cli --workflow=release.yml --limit 3
npm view @kuaimai-cli/cli version