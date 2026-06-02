# npmmirror 二进制镜像（对标飞书 lark-cli）

`npm install -g @kuaimai-cli/cli` 的 `postinstall` 会按顺序尝试：

1. **GitHub Release**（官方源）
2. **npmmirror** `/-/binary/kuaimai-cli/v<版本>/<资产文件名>`

与 [@larksuite/cli](https://www.npmjs.com/package/@larksuite/cli) 的 `install.js` 逻辑一致。

## 维护者：首次开通镜像同步

npmmirror **不会自动**镜像任意 GitHub Release，需在 [cnpm/cnpmcore](https://github.com/cnpm/cnpmcore) 注册仓库后，同步任务才会拉取资产。

向 `config/binaries.ts` 提交 PR，参考 `lark-cli` 条目：

```ts
  'kuaimai-cli': {
    category: 'kuaimai-cli',
    description: '快麦 ERP CLI（kuaimai-cli）',
    repo: 'kuaimai-cli/kuaimai-cli',
    distUrl: 'https://github.com/kuaimai-cli/kuaimai-cli/releases',
  },
```

PR 合并后，新版本 Release 会逐步同步到：

- `https://registry.npmmirror.com/-/binary/kuaimai-cli/v0.1.8/kuaimai-cli-0.1.8-darwin-amd64.tar.gz`
- CDN：`https://cdn.npmmirror.com/binaries/kuaimai-cli/...`

发版后可用下面命令自检（应返回 `302` 而非 `404`）：

```bash
curl -sI "https://registry.npmmirror.com/-/binary/kuaimai-cli/v0.1.8/kuaimai-cli-0.1.8-darwin-amd64.tar.gz" | head -3
```

## 用户：安装仍超时

在镜像未同步或 GitHub 仍不可达时：

```bash
# 代理
export https_proxy=http://127.0.0.1:7890
npm install -g @kuaimai-cli/cli@latest

# 或本地编译
cd kuaimai-cli && make build
cp ./kuaimai-cli "$(npm prefix -g)/bin/"
```

自定义 URL（内网 OSS 等）：

```bash
export KUAIMAI_CLI_DOWNLOAD_URL="https://内网地址/kuaimai-cli-0.1.8-darwin-amd64.tar.gz"
npm install -g @kuaimai-cli/cli@0.1.8
```

禁用镜像回退（仅 GitHub + 自定义 URL）：

```bash
export KUAIMAI_CLI_SKIP_MIRROR=1
npm install -g @kuaimai-cli/cli@latest
```
