#!/usr/bin/env bash
# 阶段二：从模板生成 internal/registry/meta_data.json
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SRC="${ROOT}/internal/registry/meta_data.json"
DEST="${ROOT}/internal/registry/meta_data.json"

if [[ ! -f "$SRC" ]]; then
  echo "[fetch_meta] error: missing meta_data.json template at $SRC" >&2
  exit 1
fi

# 若配置了 OpenAPI URL，可在此拉取并转换；当前使用内置元数据
if [[ -n "${KUAIMAI_OPENAPI_URL:-}" ]]; then
  echo "[fetch_meta] fetching OpenAPI from ${KUAIMAI_OPENAPI_URL}"
  tmp="$(mktemp)"
  if curl -fsSL "${KUAIMAI_OPENAPI_URL}" -o "$tmp"; then
    cp "$tmp" "$DEST"
    rm -f "$tmp"
    echo "[fetch_meta] updated from OpenAPI URL"
    exit 0
  fi
  rm -f "$tmp"
  echo "[fetch_meta] OpenAPI fetch failed, using embedded metadata" >&2
fi

echo "[fetch_meta] using embedded metadata at ${DEST}"
