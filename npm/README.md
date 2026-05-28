# @kuaimai-cli/cli npm 包

对标飞书 `@larksuite/cli`：npm 包为 **薄壳**，`postinstall` / `run.js` 从 GitHub Release 下载 Go 二进制。

完整说明见 [docs/kuaimai-cli-npx-分发与安装.md](../docs/kuaimai-cli-npx-分发与安装.md)。

## 本地调试

```bash
cd npm
npm link
kuaimai-cli --version
```

## 发布

1. 打 tag：`git tag v0.1.0 && git push origin v0.1.0`
2. GitHub Actions 运行 GoReleaser，生成 Release 与 `checksums.txt`
3. 配置仓库 Secret `NPM_TOKEN` 后自动 `npm publish`

手动发布（需先有 Release 与 checksums）：

```bash
cp /path/to/checksums.txt npm/checksums.txt
cd npm && npm publish --access public
```

## 用户安装

```bash
npx @kuaimai-cli/cli@latest install
```
