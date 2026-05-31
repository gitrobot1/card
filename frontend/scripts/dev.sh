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

npm install
exec npm run dev
