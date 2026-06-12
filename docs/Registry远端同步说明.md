# Registry 远端同步说明

> 对标 [飞书 lark-cli](https://github.com/larksuite/lark-cli) 的 OpenAPI registry 自动刷新机制。

## 数据流

```text
kuaimaierp-cli-auto  →  POST kuaimai-cli-open  →  registry.json
                                                      ↓ GET
kuaimai-cli（registry-backed 命令前自动 SyncIfNeeded）  →  ~/.kuaimai-cli/registry/registry.json
                                                      ↓
capabilities / schema / web call
```

远端地址：[http://open-cli.kuaimai.com/registry/registry.json](http://open-cli.kuaimai.com/registry/registry.json)

## 自动同步行为

1. Cobra 解析出具体命令后，在 `PersistentPreRunE` 执行 `bootstrapRegistry`
2. 仅 registry-backed 命令自动同步；`config`、`auth`、`doctor`、`skill`、`registry`、`upgrade`、`completion`、`help`、`version` 跳过
3. 读取 `config.yaml` 中 `registry.source` 与 `registry.auto_sync`
4. 本地无缓存 → 全量拉取
5. 本地有缓存 → 带 `If-None-Match` 条件请求（ETag 未变则 304，不重复写入）
6. version 变化时 stderr 提示：`registry 已更新: version=... apis=...`
7. 网络失败但本地有缓存 → 使用缓存并 stderr 警告
8. 同步后 `capabilities` / `schema` / `web call` 自动使用最新 registry

安装向导 `npx @kuaimai-cli/cli@latest install` 会在 `config init` 后主动执行一次 `registry sync`，把 registry 可用性问题提前暴露到安装阶段；失败不会中断安装，用户可稍后手动执行 `kuaimai-cli registry sync`。

## 配置

```yaml
registry:
  source: "http://open-cli.kuaimai.com/registry/registry.json"
  auto_sync: true
```

环境变量：

| 变量 | 作用 |
|------|------|
| `KUAIMAI_CLI_SKIP_REGISTRY_SYNC=1` | 禁用自动同步 |

## 命令

| 命令 | 说明 |
|------|------|
| `capabilities` | 列出全部 apiId |
| `schema [apiId]` | 接口自省 |
| `web call <apiId>` | 按 registry 调用（推荐，对标飞书域命令） |
| `registry sync` | 手动强制同步 |
| `registry watch --interval 30` | 轮询监听远端变化（开发/运维） |

## 本地开发

```bash
cd /Users/admin/Documents/project/kuaimai-cli
make build
./kuaimai-cli capabilities --output json --verbose
./kuaimai-cli registry watch --interval 15 --verbose
```
