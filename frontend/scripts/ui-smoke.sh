#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

NODE_VERSION="$(tr -d '[:space:]' < .nvmrc)"
NVM_BIN="$HOME/.nvm/versions/node/v${NODE_VERSION}/bin"

if [[ -d "$NVM_BIN" ]]; then
  export PATH="$NVM_BIN:$PATH"
elif ! command -v node >/dev/null 2>&1; then
  echo "Node v${NODE_VERSION} not found. Install with: nvm install ${NODE_VERSION}" >&2
  exit 1
fi

QUICK=0
if [[ "${1:-}" == "--quick" ]]; then
  QUICK=1
fi

echo "→ npm run build"
npm run build

if [[ "$QUICK" -eq 1 ]]; then
  echo "OK: frontend build passed (--quick, skip type-only checks note: build includes vue-tsc)"
  exit 0
fi

if curl -sf http://127.0.0.1:8088/health >/dev/null 2>&1; then
  echo "OK: backend health http://127.0.0.1:8088/health"
else
  echo "WARN: backend not reachable (Browser UI tests need ./backend/scripts/run.sh)"
fi

if curl -sf -o /dev/null http://127.0.0.1:6677/ 2>/dev/null; then
  echo "OK: frontend dev http://127.0.0.1:6677"
else
  echo "WARN: frontend dev not reachable (Browser tests need ./frontend/scripts/dev.sh)"
fi

echo "OK: ui-smoke static checks passed"
