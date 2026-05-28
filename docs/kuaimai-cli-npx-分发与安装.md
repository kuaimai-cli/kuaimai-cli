# kuaimai-cli：对标飞书的 npx 一键安装与升级

本文说明如何实现、发布和使用 **飞书式** `npx @kuaimai-cli/cli@latest install` 分发能力：原理、仓库现状、发布步骤、用户命令，以及与飞书的差异。

> **读者**：维护者（首次发版 / 日常发版）、用户与 Agent 安装负责人。  
> **相关**：[Agent 安装指南](./kuaimai-cli-agent-installation-guide.md) · [内部分发](./kuaimai-cli-内部分发与使用.md) · [npm/ 目录说明](../npm/README.md) · [系统架构与飞书对标](./系统架构与飞书对标说明.md)

---

## 1. 目标与结论（先看这段）

| 问题 | 答案 |
|------|------|
| 飞书 `npx @larksuite/cli install` 本质是什么？ | **npm 公网 registry 上的薄壳包** + **按平台从 Release 下载的 Go 二进制** |
| kuaimai-cli 代码侧做完了吗？ | **是**。`npm/`、`install.js`、`release.yml`、GoReleaser 已具备 |
| 用户现在能直接 `npx` 吗？ | **取决于是否已发布**：`@kuaimai-cli/cli` 需出现在 registry，且 GitHub Release 有对应版本资产 |
| 和飞书是否 1:1？ | **安装形态接近**；鉴权、升级自动替换二进制、命令名等与飞书仍有差异（见 §8） |

**一句话**：不是「再写一个 npm 壳」，而是 **打 tag → GitHub Release → 发布到 npm registry**。

---

## 2. 名词

| 术语 | 含义 |
|------|------|
| **Registry** | 存放 npm 包的服务器。公网默认 `https://registry.npmjs.org`；内网可用公司私有 registry；国内常用 `registry.npmmirror.com` 作镜像加速 |
| **npx** | 临时下载并执行 registry 上某个 npm 包里的命令，无需先全局安装 |
| **薄壳（npm 包）** | 只有 Node 脚本，**不包含** Go 二进制；负责下载、转发子命令 |
| **Release** | GitHub 上按版本发布的压缩包（`kuaimai-cli-{version}-{os}-{arch}.tar.gz` 等） |

---

## 3. 原理：飞书与我们相同的部分

```text
用户执行:  npx @kuaimai-cli/cli@latest install
                │
                ▼
        npm 从 registry 拉取 @kuaimai-cli/cli（薄壳，几 KB～几十 KB）
                │
                ├─► 子命令 install → scripts/install-wizard.js
                │       · npm install -g @kuaimai-cli/cli（全局命令 kuaimai-cli）
                │       · skill install-all、config init、登录提示
                │
                └─► postinstall / 首次运行 → scripts/install.js
                        · 识别 darwin/linux/windows + amd64/arm64
                        · 从 GitHub Release（或镜像）下载 Go 二进制
                        · 校验 checksums.txt（有则校验）
                        · 写入 npm 包目录下的 bin/kuaimai-cli
```

**重要**：二进制 **不是打进 npm 包 tarball**，而是安装时 **按平台下载**。这与 `@larksuite/cli` 一致。

日常使用时：

```text
kuaimai-cli item +list ...
        │
        ▼
全局安装的 scripts/run.js → 执行已下载的 bin/kuaimai-cli
```

---

## 4. 仓库里已有什么

### 4.1 目录与职责

| 路径 | 职责 |
|------|------|
| [`npm/package.json`](../npm/package.json) | 包名 `@kuaimai-cli/cli`，`bin.kuaimai-cli` → `scripts/run.js` |
| [`npm/scripts/run.js`](../npm/scripts/run.js) | 入口：`install` 走向导；其它参数转发给 Go 二进制 |
| [`npm/scripts/install.js`](../npm/scripts/install.js) | 按平台下载 Release 资产，写入 `npm/bin/` |
| [`npm/scripts/install-wizard.js`](../npm/scripts/install-wizard.js) | 交互式一键安装（全局 npm、Skill、config） |
| [`npm/checksums.txt`](../npm/checksums.txt) | 发布时由 CI 写入，供 `install.js` 校验 |
| [`.goreleaser.yaml`](../.goreleaser.yaml) | 构建 6 平台二进制并打 Release |
| [`.github/workflows/release.yml`](../.github/workflows/release.yml) | tag 触发：GoReleaser → 同步 checksums → `npm publish` |

### 4.2 Release 产物命名（须与 install.js 一致）

模板（见 `.goreleaser.yaml`）：

```text
kuaimai-cli-{version}-{os}-{arch}.tar.gz   # macOS / Linux
kuaimai-cli-{version}-windows-{arch}.zip   # Windows
```

`install.js` 中仓库默认为 `kuaimai/kuaimai-cli`；若实际 GitHub 路径不同，须改 `REPO` 常量。

### 4.3 支持的平台

| OS | 架构 |
|----|------|
| darwin | amd64, arm64 |
| linux | amd64, arm64 |
| windows | amd64（无 windows/arm64） |

---

## 5. 用户侧：一键安装

### 5.1 推荐命令（对标飞书）

```bash
npx @kuaimai-cli/cli@latest install
```

前置：本机已安装 **Node.js 16+**（含 `npm` / `npx`），能访问 **registry** 与 **GitHub Release**（或已配置的镜像）。

### 5.2 安装向导会做什么

1. `npm install -g @kuaimai-cli/cli`（若尚未全局安装）
2. 下载当前 npm 包版本对应的 Go 二进制（`postinstall` / 首次运行）
3. `kuaimai-cli skill install-all`
4. `kuaimai-cli config init`
5. 提示执行 `kuaimai-cli auth login "<accessToken>"`（须从 ERP 浏览器获取）

非 TTY（如 Agent 子进程）时：尽力执行 1～4，并打印需用户手动完成的 auth 步骤。见 [Agent 安装指南](./kuaimai-cli-agent-installation-guide.md)。

### 5.3 安装后验证

```bash
kuaimai-cli --version
kuaimai-cli doctor --output json
kuaimai-cli auth check --output json
```

### 5.4 不用 npx 的替代方式

| 方式 | 场景 |
|------|------|
| [GitHub Releases](https://github.com/kuaimai/kuaimai-cli/releases) 下载二进制 | 无 Node、仅要 CLI |
| `go install github.com/kuaimai/kuaimai-cli@latest` | 开发者本机 |
| [内部分发压缩包](./kuaimai-cli-内部分发与使用.md) | 内网离线、`dist` + `skills` |

---

## 6. 用户侧：升级

### 6.1 检查是否有新版本

```bash
kuaimai-cli upgrade --output json
```

当前实现：**查询 GitHub 最新 Release**，对比本地版本，**不会自动替换**本机二进制。

### 6.2 实际升级方式（推荐）

与飞书类似，升级 = **重新安装新版本**：

```bash
npx @kuaimai-cli/cli@latest install
```

或：

```bash
npm install -g @kuaimai-cli/cli@latest
```

然后再次确认：

```bash
kuaimai-cli --version
kuaimai-cli upgrade
```

> **规划项**：`upgrade` 自动下载并替换本地二进制尚未实现；见 [系统架构与飞书对标](./系统架构与飞书对标说明.md) §8.2。

### 6.3 通过 npx 直接跑子命令

未全局安装时，也可：

```bash
npx @kuaimai-cli/cli@latest item +list --body '{"pageNo":1,"pageSize":10}' --output json
```

`run.js` 会在缺少二进制时尝试触发 `install.js` 下载。

---

## 7. 维护者：首次发布到 npm 公网 registry

按顺序执行，缺一步都会导致 `npx` 404 或下载失败。

### 7.1 准备 npm 账号与 scope

1. 注册 [npmjs.com](https://www.npmjs.com/) 账号。
2. 使用个人 scope **`@kuaimai-cli`**，确保有权限发布 `@kuaimai-cli/cli`。
3. 创建 **Automation Token**（用于 CI `npm publish`）。

### 7.2 配置 GitHub

1. 仓库建议为 **`kuaimai/kuaimai-cli`**（与 `release.yml` 中 `if: github.repository == 'kuaimai/kuaimai-cli'` 一致；否则需改 workflow）。
2. Settings → Secrets → Actions → 添加 **`NPM_TOKEN`**（上一步的 token）。

### 7.3 打 tag 触发 Release + npm publish

```bash
# 版本号与 npm/package.json、CHANGELOG 对齐
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions 将依次：

1. **GoReleaser**：编译多平台包，创建 GitHub Release，上传 `checksums.txt`
2. **npm-checksums**：把 `checksums.txt` 拷入 `npm/`
3. **npm-publish**：同步 `npm/package.json` 版本并 `npm publish --access public`

### 7.4 验证发布成功

```bash
npm view @kuaimai-cli/cli version
npm view @kuaimai-cli/cli bin

npx @kuaimai-cli/cli@latest --version
npx @kuaimai-cli/cli@latest install
```

### 7.5 手动发布（仅当 CI 不可用时）

须 **先有** 对应版本的 GitHub Release 与 `checksums.txt`：

```bash
cp /path/to/checksums.txt npm/checksums.txt
# 编辑 npm/package.json 的 version 与 tag 一致
cd npm && npm publish --access public
```

---

## 8. 维护者：日常发版

```text
1. 合并代码，更新 CHANGELOG
2. git tag vX.Y.Z && git push origin vX.Y.Z
3. 等待 Actions 完成 Release + npm
4. 通知用户：npx @kuaimai-cli/cli@latest install 或 npm install -g @kuaimai-cli/cli@latest
```

**版本对齐规则**：

- Git tag：`v0.2.0`
- GitHub Release 资产：`kuaimai-cli-0.2.0-...`
- npm 包版本：`0.2.0`（CI 从 tag 去掉 `v` 写入 `package.json`）

---

## 9. 内网 / 私有 registry（可选）

公网 `registry.npmjs.org` 不适合仅内网使用时：

| 方案 | 说明 |
|------|------|
| **私有 registry** | 将 `npm/` 同样 `npm publish` 到公司 Verdaccio 等；用户 `npm config set registry <内网 URL>` 后执行 `npx @kuaimai-cli/cli install` |
| **二进制镜像** | `install.js` 会读 `npm_config_registry`，非 npmjs 时尝试从 registry 的 `/-/binary/kuaimai-cli/v{version}/` 拉包；默认还尝试 `registry.npmmirror.com` |
| **压缩包分发** | 不经过 npm，见 [内部分发与使用](./kuaimai-cli-内部分发与使用.md) |

私有 registry **只解决「npm 薄壳从哪下」**；Go 二进制仍须来自 **Release 或同步到镜像的二进制路径**。

---

## 10. 本地调试 npm 薄壳

```bash
cd npm
npm link
kuaimai-cli --version
```

调试 `install` 向导：

```bash
node scripts/install-wizard.js
```

强制重新下载二进制（需已有 Release）：

```bash
KUAIMAI_CLI_RUN=true node scripts/install.js
```

---

## 11. 与飞书 lark-cli 的差异

| 项 | 飞书 `@larksuite/cli` | kuaimai `@kuaimai-cli/cli` |
|----|----------------------|-------------------------|
| 全局命令名 | `lark-cli` | `kuaimai-cli` |
| 二进制来源 | Release 下载 | 同左 |
| 一键安装命令 | `npx @larksuite/cli@latest install` | `npx @kuaimai-cli/cli@latest install` |
| 登录 | OAuth、`auth login --recommend` | 手动 `auth login "<accessToken>"` |
| `upgrade` | 生态较成熟 | 仅检查版本 + 提示重装 |
| 安装后 Skill | 飞书多域 Skill | `kuaimai-shared` + `kuaimai-item` |
| 公网 npm 状态 | 已发布 | **需维护者完成 §7 后可用** |

---

## 12. 常见问题

### `npm error 404 '@kuaimai-cli/cli'`

包尚未发布到当前 registry，或 scope/包名错误。维护者完成 §7；用户确认 `npm config get registry`。

### 下载二进制失败 / checksum 失败

- 确认 GitHub Release 存在且含 `kuaimai-cli-{version}-{platform}-{arch}` 文件。
- 确认 `npm/checksums.txt` 与 Release 一致（应由 CI 生成）。
- 内网检查是否需代理或改用镜像 / 压缩包分发。

### `npx` 很慢

npx 每次可能拉最新包；已安装可改用全局 `kuaimai-cli`。

### 全局命令找不到

确保 `npm prefix -g` 下的 `bin` 在 `PATH` 中；macOS/Linux 常见为 `export PATH="$(npm prefix -g)/bin:$PATH"`。

### 与「再做一个 npm 模板」的关系

仓库 **已有** 完整薄壳，**无需从零复制第三方模板**。未发布时缺的是 **registry + Release 运维**，不是缺代码。

---

## 13. 相关文档

| 文档 | 用途 |
|------|------|
| [Agent 安装指南](./kuaimai-cli-agent-installation-guide.md) | Agent / IDE 安装步骤 |
| [内部分发与使用](./kuaimai-cli-内部分发与使用.md) | 无 npm 时的 zip 分发 |
| [npm/README.md](../npm/README.md) | 发布命令速查 |
| [系统架构与飞书对标](./系统架构与飞书对标说明.md) | 架构与路线图 |

---

## 14. 速查卡片

**用户安装**

```bash
npx @kuaimai-cli/cli@latest install
kuaimai-cli auth login "<accessToken>"
kuaimai-cli auth check
```

**用户升级**

```bash
kuaimai-cli upgrade
npx @kuaimai-cli/cli@latest install
```

**维护者首发**

```bash
# NPM_TOKEN in GitHub Secrets
git tag v0.1.0 && git push origin v0.1.0
npm view @kuaimai-cli/cli version
```
