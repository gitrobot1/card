#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND="$ROOT/../backend"
ROUNDS="${CARD_UI_FIXTURE_ROUNDS:-50}"

echo "== UI Sim: harvest settlement fixtures (rounds/mode cap=$ROUNDS) =="
FIXTURE_DIR="$ROOT/test/fixtures/yzs/settlements"
rm -f "$FIXTURE_DIR"/*.json 2>/dev/null || true
cd "$BACKEND"
CARD_SIM=1 CARD_UI_FIXTURE=1 CARD_UI_FIXTURE_ROUNDS="$ROUNDS" \
  go test -tags cardtest ./test/yuzhousha/... -count=1 -run TestHarvestYzsSettlementFixtures -v

echo ""
echo "== UI Sim: validate frontend settlement display =="
cd "$ROOT"
npm run test:settlement

echo ""
echo "OK: ui-sim complete"
