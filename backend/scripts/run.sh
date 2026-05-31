#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

GO_VERSION="$(tr -d '[:space:]' < .go-version)"
GVM_GO="$HOME/.gvm/gos/${GO_VERSION}/bin/go"

if [[ -x "$GVM_GO" ]]; then
  GO="$GVM_GO"
elif command -v go >/dev/null 2>&1; then
  GO="go"
else
  echo "Go ${GO_VERSION} not found. Install with: gvm install ${GO_VERSION}" >&2
  exit 1
fi

export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
exec "$GO" run ./cmd/server -config ./config/config.yaml
