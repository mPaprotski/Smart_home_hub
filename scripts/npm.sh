#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
if [[ -x "$ROOT/.tools/node_modules/node/bin/node" ]]; then
  export PATH="$ROOT/.tools/node_modules/node/bin:$PATH"
fi
export NPM_CONFIG_CACHE="$ROOT/.cache/npm"
exec npm --prefix "$ROOT/frontend" "$@"
