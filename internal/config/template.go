package config

// DefaultConfigTemplate 是 config init 写入的默认配置模板（静态字符串，无业务逻辑）。
const DefaultConfigTemplate = `# kuaimai-cli 本地默认配置（自动生成）
# 配置目录：~/.kuaimai-cli/config.yaml
# 由 config init 命令自动生成

api:
  url: "https://erp1.superboss.cc/"
  timeout: 30
  retry: 3
  pool_max_idle: 100
  pool_max_idle_per_host: 10
  circuit_threshold: 5
  circuit_cooldown_sec: 30

cli:
  output: "table"
  color: true

# 多账号 profile：auth login --profile <name>；切换：auth use <name>
auth:
  profile: default
  profiles:
    default: {}
`
