# 宇宙杀结算 UI Fixtures

由后端 AI 自对弈在终局导出（`CARD_UI_FIXTURE=1`）。

```bash
# 生成（在 backend/）
CARD_SIM=1 CARD_UI_FIXTURE=1 CARD_UI_FIXTURE_ROUNDS=20 \
  go test -tags cardtest ./test/yuzhousha/... -run TestHarvestYzsSettlementFixtures -count=1 -v

# 校验（在 frontend/）
npm run test:settlement
```

或一键：`./scripts/ui-sim.sh`

JSON 格式：`{ "meta": { mode, seed, winner_team, label }, "state": <PublicView seat0> }`
