# kuaimai-cli 本地使用实操指南

下文命令均在项目根目录执行，二进制名为 `./kuaimai-cli`。勿把真实 token 提交到 Git。

**Agent 安装**（无需 clone 仓库）：见 [Agent 安装指南](./kuaimai-cli-agent-installation-guide.md)。

---

## 零、用户安装（非开发者）

### npm 一键安装（对标飞书）

```bash
npx @kuaimai-cli/cli@latest install
```

### 或下载 Release 二进制

从 [GitHub Releases](https://github.com/kuaimai-cli/kuaimai-cli/releases) 下载 `kuaimai-cli-{version}-{os}-{arch}.tar.gz`，解压后加入 `PATH`。

### 安装 Skills

```bash
kuaimai-cli skill install-all
kuaimai-cli skill list --output json
```

---

## 一、编译（开发者）

```bash
make build
# 或
go build -o kuaimai-cli .

./kuaimai-cli --help
./kuaimai-cli item --help
```

---

## 二、初始化配置

配置文件：`~/.kuaimai-cli/config.yaml`（token **不**写在这里）。

```bash
./kuaimai-cli config init
./kuaimai-cli config get
```

已存在配置时，`config init` 不会覆盖。

按需修改：

```bash
./kuaimai-cli config set api.url "https://erp1.superboss.cc/"
./kuaimai-cli config set cli.output json
```

---

## 三、登录 Token

```bash
./kuaimai-cli auth login "<你的accessToken>"
./kuaimai-cli auth status --output json
./kuaimai-cli auth check --output json    # 探测 token 与 API 是否可用
./kuaimai-cli auth logout
```

登录后，每次请求会自动在请求头带上 `accessToken`（来自密钥链，无需手动设置）。

### 多账号（可选）

```bash
./kuaimai-cli auth login "<tokenA>" --profile prod
./kuaimai-cli auth login "<tokenB>" --profile test
./kuaimai-cli auth list --output json
./kuaimai-cli auth use prod
```

### 安装自检

```bash
./kuaimai-cli doctor --output json
./kuaimai-cli upgrade --output json    # 查询是否有新版本（需网络）
```

---

## 四、商品操作

### 4.1 按标题查列表

```bash
./kuaimai-cli item +list \
  --body '{"title":"2026","pageNo":1,"pageSize":50}' \
  --output json
```

### 4.2 查总数

```bash
./kuaimai-cli item count \
  --body '{"title":"2026"}' \
  --output json
```

### 4.3 商品详情

```bash
./kuaimai-cli item get-detail \
  --sys-item-id 5444595188086784 \
  --output json
```

### 4.4 修改商品标题（推荐 update-title）

```bash
./kuaimai-cli item update-title \
  --sys-item-id 5444595188086784 \
  --title "被cli修改后的商品" \
  --output json

# 试跑（不提交）
./kuaimai-cli item update-title \
  --sys-item-id 5444595188086784 \
  --title "被cli修改后的商品" \
  --dry-run --verbose --output json
```

CLI 内部会 `get-detail` 合并全量字段后 `save`，无需手写 `jq`。

### 4.5 修改商品标题（get-detail + jq + save）

需改多个字段或自定义合并逻辑时，可手动 `jq`（`sys-item-id` 换成你的商品 ID）：

```bash
./kuaimai-cli item save \
  --body "$(./kuaimai-cli item get-detail \
    --sys-item-id 5444595188086784 \
    --output json | jq -c '.data[0] | .title = "被cli修改后的商品" | .suiteBridgeList = .itemSuiteBridgeList | del(.itemSuiteBridgeList)')" \
  --output json
```

试跑（不提交）：

```bash
./kuaimai-cli item save \
  --body "$(./kuaimai-cli item get-detail \
    --sys-item-id 5444595188086784 \
    --output json | jq -c '.data[0] | .title = "被cli修改后的商品" | .suiteBridgeList = .itemSuiteBridgeList | del(.itemSuiteBridgeList)')" \
  --dry-run --verbose --output json
```

### 4.6 常用参数（可选）

```bash
--verbose      # 调试日志（stderr）
--page-all     # 列表自动翻页
--output csv   # 导出 CSV
--output ndjson
```

---

## 五、一条龙示例：改标题

**场景**：把标题里含有 `2026` 的商品，改名为 `被cli修改后的商品`。

**步骤**：编译 → 配置 → 登录 → 列表查出 `sysItemId` → 拉详情并保存。

```bash
go build -o kuaimai-cli .
./kuaimai-cli config init
./kuaimai-cli auth login "<accessToken>"

# 1. 按标题「2026」查列表，从结果里记下要改的商品 sysItemId
./kuaimai-cli item +list \
  --body '{"title":"2026","pageNo":1,"pageSize":20}' \
  --output json

# 2. 将下面 sys-item-id 换成上一步的 ID，保存为新标题「被cli修改后的商品」
./kuaimai-cli item update-title \
  --sys-item-id 5444595188086784 \
  --title "被cli修改后的商品" \
  --output json
```
