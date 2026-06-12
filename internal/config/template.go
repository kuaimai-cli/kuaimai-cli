package config

// DefaultConfigTemplate 是 config init 写入的默认配置模板（静态字符串，无业务逻辑）。
const DefaultConfigTemplate = `# kuaimai-cli 本地默认配置（自动生成）
# 配置目录：~/.kuaimai-cli/config.yaml
# 由 config init 命令自动生成

api:
  url: "https://erp1.superboss.cc/"
  gateway_url: "https://open-cli.kuaimai.com"
  timeout: 60
  retry: 3
  pool_max_idle: 100
  pool_max_idle_per_host: 10
  circuit_threshold: 5
  circuit_cooldown_sec: 30

# curated shortcuts 的业务目标域名；实际 HTTP 仍统一经 api.gateway_url 转发
shortcuts:
  erp-item:
    api_url: "https://erp1.superboss.cc/"
  scm-item:
    api_url: "https://scm3.superboss.cc/"

cli:
  output: "table"
  color: true

# Registry 远程源（对标飞书 CLI OpenAPI registry 刷新）
registry:
  source: "http://open-cli.kuaimai.com/registry/registry.json"
  auto_sync: true   # 每次命令前自动检查远端 version/ETag 并更新本地缓存

# 多账号 profile：auth login --profile <name>；切换：auth use <name>
auth:
  profile: default
  profiles:
    default: {}
`
