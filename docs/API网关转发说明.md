# API 网关转发说明

> kuaimai-cli 业务 API 统一经 `kuaimai-cli-open` 网关转发，不再直连 ERP/SCM 后端。

## 数据流

```text
kuaimai-cli（item / web call / api / auth check）
  │  POST {api.gateway_url}/api/forward
  │  Header: accessToken（不变）
  │  Body: targetHost + method + path + body/queryParams/contentType
  ▼
kuaimai-cli-open（open-cli.kuaimai.com）
  │  1. 校验 targetHost ∈ *.superboss.cc
  │  2. 限流 key = host + 规范化 path + accessToken（100 次/分钟）
  │  3. 透传请求头，转发至真实后端
  │  4. 透传响应 status + body
  ▼
erp1.superboss.cc / scm.superboss.cc / …
```

**不经网关的请求**（保持直连）：

- `registry sync` → `registry.source`
- `skill install` → GitHub
- `upgrade` → 版本检查

## CLI 配置

```yaml
api:
  url: "https://erp1.superboss.cc/"          # 逻辑目标域名（body 中的 targetHost）
  gateway_url: "https://open-cli.kuaimai.com" # 网关根地址
  timeout: 60                                 # 秒，与网关上游超时一致
```

| 项 | 说明 |
|----|------|
| `api.url` | ERP 默认后端；item shortcuts、`web call item.*` 的 `targetHost` |
| `api.gateway_url` | 所有业务 HTTP 的实际请求地址 |
| scm 域 | `web call scm.*` 的 `targetHost` 来自 registry `baseUrl`，仍经同一网关 |

## 网关接口

```
POST /api/forward
Header: accessToken: <用户 token>
Content-Type: application/json
```

请求体：

```json
{
  "targetHost": "https://erp1.superboss.cc",
  "method": "POST",
  "path": "/item/stock/queryList",
  "contentType": "application/json",
  "body": "{\"pageNo\":1,\"pageSize\":10}",
  "queryParams": {}
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `targetHost` | 是 | 目标域名，http/https 均可 |
| `method` | 是 | GET / POST / … |
| `path` | 是 | 接口路径，不含 query |
| `body` | 否 | 原始请求体字符串 |
| `contentType` | 否 | 上游 Content-Type |
| `queryParams` | 否 | URL 查询参数 map |

**成功**：原样返回上游 HTTP 状态码与 JSON body。  
**网关错误**（400/401/403/429/502）：返回 `{"result":<code>,"message":"..."}`。

## 限流

| 项 | 值 |
|----|-----|
| 算法 | 固定窗口（内存 Map，单节点） |
| 额度 | 100 次/分钟（可配置 `kuaimai.forward.rate-limit-per-minute`） |
| key | `host + path + accessToken`（host 忽略 http/https；path 小写、去尾斜杠；不含 query） |

CLI 收到 **429** 时不重试，提示：`请求过于频繁，请稍后重试`。

## 网关配置（kuaimai-cli-open）

```yaml
kuaimai:
  forward:
    enabled: true
    rate-limit-per-minute: 100
    allowed-host-suffix: superboss.cc
    timeout-seconds: 60
```

## 部署顺序

1. 部署 **kuaimai-cli-open**（含 `/api/forward`）
2. 发布 **kuaimai-cli** 新版（业务请求走网关）

## 本地验证

```bash
# 1. 启动 open 服务
cd kuaimai-cli-open && mvn -pl kuaimai-cli-open-web -am spring-boot:run -Dspring-boot.run.profiles=dev

# 2. 网关探针（缺 token 应 401）
curl -s -o /dev/null -w "%{http_code}" -X POST http://127.0.0.1:8080/api/forward \
  -H 'Content-Type: application/json' \
  -d '{"targetHost":"https://erp1.superboss.cc","method":"GET","path":"/"}'

# 3. CLI dry-run 预览网关路径
cd kuaimai-cli && make build
./kuaimai-cli item +list --body '{"pageNo":1,"pageSize":1}' --dry-run --verbose
```
