全部命令无占位、可直接复制运行｜开发 / 推送 / PR / 发版 / 重置 / 用户安装 全流程闭环
详细规范见：开发发布流程


# 1. 清理所有失败的旧Tag（v0.1.4/v0.1.5）
git push origin --delete v0.1.4
git push origin --delete v0.1.5
git tag -d v0.1.4
git tag -d v0.1.5

# 2. 同步最新代码，提交配置修改
git add -A
git reset -- scripts/generate_meta/__pycache__/ scripts/filter_meta/__pycache__/
git commit -m "fix: 修复goreleaser v2兼容报错，适配vendor编译"
git push origin main

# 3. 全新升级版本，正式发版
git checkout main
git pull origin main
npm view @kuaimai-cli/cli version
git tag v0.1.6
git push origin v0.1.6




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

二、完整分支开发流程（新建分支 → 推送分支 → 更新main → 合并分支 → 推送main）
适用场景：新建自定义分支开发、不直接改 main、开发完合并主干的标准团队流程
分支示例：开发分支名 = dev-test
2.1 完整一步不落标准流程（全可直接复制）
cd /Users/admin/Documents/project/kuaimai-cli

# ========== 第一步：从最新 main 创建【新开发分支 dev-test】 ==========

git checkout main
git pull origin main
git checkout -b dev-test

# ========== 第二步：开发完代码后，提交、推送 dev-test 分支 ==========

git status
git add -A
git reset -- scripts/generate_meta/**pycache**/ scripts/filter_meta/**pycache**/
git commit -m "feat: 开发分支功能迭代"

# 推送【新建的 dev-test 分支】到远端

git push -u origin dev-test

# ========== 第三步：开发完毕，更新本地 main 为最新远端代码 ==========

git checkout main
git pull origin main

# ========== 第四步：将 dev-test 合并到本地 main ==========

git merge dev-test

# ========== 第五步：推送最终合并后的 main 主干 ==========

git push origin main

2.2 每一步作用极简解释（彻底看懂）

- git checkout -b dev-test：新建分支并切换进去
- git push -u origin dev-test：推送新分支到 GitHub（远端创建该分支）
- git checkout main && git pull：保证主干是最新代码，防止冲突
- git merge dev-test：把开发分支代码合并进主干
- git push origin main：推送合并后的最终代码，触发 CI 检测
2.3 合并后可选：删除无用开发分支（干净规范）

# 删除本地分支

git branch -d dev-test

# 删除远端分支

git push origin --delete dev-test

---

三、代码直接推送 main（快速简单模式）
核心规则：推 main 只跑 CI，不发版；仅 Tag 触发 Release & NPM
项目固定路径：/Users/admin/Documents/project/kuaimai-cli
3.1 日常快速推送 main（最常用）
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

3.2 标准分支 PR 流程（CodeReview 专用）
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

四、核心闭环：代码推送 → 发版 → 用户可用（必走流程）
关键结论：仅推送代码到 GitHub main 分支，用户无法获取新版代码、无法更新。必须打 Tag 发版发布 NPM 包，用户执行安装命令才可使用最新功能。
4.1 第一步：正式打 Tag 发版（唯一触发 NPM 发布）

# 切换并同步最新主干代码

git checkout main
git pull origin main

# 【必做】查询线上当前最新 NPM 版本（避免版本重复/倒退）

npm view @kuaimai-cli/cli version

# 打新版标签（根据线上版本递增，示例：线上v0.1.3则打v0.1.4）

git tag v0.1.4

# 推送Tag，自动触发 release.yml 流水线（编译+打包+NPM发布）

git push origin v0.1.4

4.2 第二步：校验发布成功（无需gh工具）
等待1-3分钟流水线执行完毕，执行以下命令校验是否发布成功，版本号与你打的 Tag 一致即为发布完成。

# 查看NPM线上最新版本，匹配新版号即发布成功

npm view @kuaimai-cli/cli version

4.3 第三步：用户更新/安装新版 CLI（最终生效）
发版成功后，普通用户、测试、AI 客户端才可拉取到最新代码，执行以下命令完成更新安装。

# 一键安装/更新到最新版本

npx @kuaimai-cli/cli@latest install

4.4 完整闭环链路（牢记）
代码修改 & 推送main（仅更新代码、跑CI） → 打Tag发版（发布NPM包） → 校验版本成功 → 用户安装使用新版

---

四、版本发版 & NPM 发布（唯一正式发版流程）
发版铁律：必须先查线上版本 → 再打新版Tag → 禁止重复/倒退
4.1 发版前置：查询线上真实版本（必跑）

# 查GitHub线上正式Release版本

gh release list --repo kuaimai-cli/kuaimai-cli --limit 5

# 查NPM用户实际安装版本

npm view @kuaimai-cli/cli version

# 对比本机版本

kuaimai-cli --version

4.2 正式打Tag发版（真实可运行模板）
示例：线上最新 v0.1.3，本次小修复升级为 v0.1.4
cd /Users/admin/Documents/project/kuaimai-cli
git checkout main && git pull origin main

# 兜底校验线上版本

gh release list --repo kuaimai-cli/kuaimai-cli --limit 3
npm view @kuaimai-cli/cli version

# 打新版标签 + 推送触发发布流水线

git tag v0.1.4
git push origin v0.1.4

4.3 观测 Release 发布结果

# 查看release流水线执行状态

gh run list --repo kuaimai-cli/kuaimai-cli --workflow=release.yml --limit 3

# 查看本次发布详情

gh release view v0.1.4 --repo kuaimai-cli/kuaimai-cli

# 校验NPM是否发布成功

npm view @kuaimai-cli/cli version

4.4 发版失败重置（Tag冲突/报错专用）

# 删除本地+远端错误Tag

git push origin --delete v0.1.4
git tag -d v0.1.4

# 修正代码/日志后，升版本重发示例

git tag v0.1.5
git push origin v0.1.5

---

五、开发人员本机更新 & 异常重置

# 1. 一键升级（默认：检查 + npm 全局安装 + Skills 同步）

kuaimai-cli upgrade

# 仅检查、不安装

kuaimai-cli upgrade --check-only --output json

# 2. 等价手动路径（upgrade 内部即调用 npm install -g @latest）

npx @kuaimai-cli/cli@latest install
# 0.1.8+：向导对比全局 npm 版本，旧版会 npm install -g @kuaimai-cli/cli@<向导版本>，不再「有 0.1.0 就跳过」
kuaimai-cli skill install --if-stale

# 3. 本地源码更新（未发版调试用）

cd /Users/admin/Documents/project/kuaimai-cli
git pull origin main
make build
cp ./kuaimai-cli ~/bin/kuaimai-cli

# 4. 全局异常（版本仍卡在旧号）

npm install -g @kuaimai-cli/cli@latest
kuaimai-cli --version
kuaimai-cli skill install

# 强制重装向导（不删 ~/.kuaimai-cli 配置）
KUAIMAI_CLI_FORCE_INSTALL=1 npx @kuaimai-cli/cli@latest install

# 5. 仅 Skill 异常：覆盖重装（勿删 ~/.kuaimai-cli）
kuaimai-cli skill install
kuaimai-cli doctor --output json

# 6. 最后兜底：删各 Agent 下 kuaimai-* 后再装（勿删 ~/.kuaimai-cli）
# rm -rf ~/.agents/skills/kuaimai-* ~/.cursor/skills/kuaimai-* ...
# npx @kuaimai-cli/cli@latest install

---

六、普通用户：安装 & 日常使用命令

# 1. 一键安装最新版

npx @kuaimai-cli/cli@latest install

# 或全局安装

npm install -g @kuaimai-cli/cli@latest

# 2. 基础校验

kuaimai-cli --version
kuaimai-cli doctor --output json

# 3. 权限校验 & 技能同步

kuaimai-cli auth check --output json
kuaimai-cli skill install --if-stale

# 4. 升级（默认一键；任意命令后 stderr 也可能提示新版）

kuaimai-cli upgrade
# kuaimai-cli upgrade --check-only --output json

---

七、固定发版完整流水线（终极一键模板）
每次发版直接改版本号即可，全真实可用
cd /Users/admin/Documents/project/kuaimai-cli
git checkout main && git pull origin main
gh release list --repo kuaimai-cli/kuaimai-cli --limit 3
npm view @kuaimai-cli/cli version
git tag v0.1.4
git push origin v0.1.4
gh run list --repo kuaimai-cli/kuaimai-cli --workflow=release.yml --limit 3
npm view @kuaimai-cli/cli version