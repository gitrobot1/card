---
name: card-ui-test
description: >-
  Card 项目前端 UI 自动化测试 Agent。ui-sim 大量随机 AI 自走导出结算快照并 Vitest 校验；
  或用 Browser MCP 点选测 pending/布局。用户说「测前端」「UI 测试」「ui-sim」「结算测试」
  「carduitest」「界面回归」「随机测前端」「identity_8 布局」或改 frontend 结算/pending 时使用。
disable-model-invocation: true
---

# Card 前端 UI 测试 Agent

> 用户 @card-ui-test 或说「测前端 / UI 测试 / ui-sim / 随机测结算」时，**你必须自己跑命令**（优先 ui-sim），Browser 抽检，按报告模板回复。
>
> **与 card-test 分工**：
>
> - **card-test** → 后端规则/状态机/AI sim（`./scripts/test.sh simrandom`）
> - **card-ui-test** → **前端结算展示是否正确**（`ui-sim`）+ 浏览器操作（Browser MCP）

---

## 核心：ui-sim（大量随机 AI 玩 → 校验前端结算）

与后端 `simrandom` 同思路，但终局把 **PublicView JSON** 交给前端 `validateSettlementState` 校验。

```bash
# 一键（推荐）
cd frontend && ./scripts/ui-sim.sh

# 或分步
cd backend
CARD_SIM=1 CARD_UI_FIXTURE=1 CARD_UI_FIXTURE_ROUNDS=20 ./scripts/test.sh uifixture -v
cd ../frontend && npm run test:settlement
```


| 变量                         | 默认            | 作用                         |
| -------------------------- | ------------- | -------------------------- |
| `CARD_SIM=1`               | uifixture 自动设 | 跑 AI 自对弈                   |
| `CARD_UI_FIXTURE=1`        | uifixture 自动设 | 终局写入 JSON                  |
| `CARD_UI_FIXTURE_ROUNDS=N` | `50`          | 每模式（identity/1v1/2v2）采样种子数 |


**产物**：`frontend/test/fixtures/yzs/settlements/*.json`  
**校验逻辑**：`frontend/src/composables/yuzhousha/settlementDisplay.ts`

每条 fixture 校验：

- `phase === finished`
- `message` 非空（= 页面 `centerHint` 结算文案）
- `winner_index` / `winner_team` 合法
- `events` 含 `game_over`
- **identity_5/8**：文案与 `winner_team` 一致（反贼/内奸/主公阵营）
- **1v1/2v2/3v3**：文案含「获胜」
- **identity_8**：`layout_key === octagon_8`

### 一键决策（用户说随机测前端 / 结算）

```
用户要测什么？
├─ 大量随机 / 结算对不对 / ui-sim      → ./scripts/ui-sim.sh（或加大 CARD_UI_FIXTURE_ROUNDS）
├─ 只校验已有 fixture                  → cd frontend && npm run test:settlement
├─ 改 pending / 响应窗                   → ui-smoke + Browser yzs-pending（scenarios-yuzhousha.md）
├─ 改布局 / identity_8                 → ui-sim + Browser yzs-layout-identity8
├─ 合码前（宇宙杀 UI）                 → ui-smoke.sh + ui-sim（ROUNDS=15）+ Browser smoke-global
└─ 全栈合码                            → card-test smoke + card-ui-test ui-sim
```

---

## 静态检查

```bash
cd frontend && ./scripts/ui-smoke.sh
```

---

## Browser MCP（抽检，非大量随机）

大量随机用 **ui-sim**；Browser 用于 pending 按钮、布局、动画等 **ui-sim 覆盖不到** 的交互。

```
browser_navigate → http://127.0.0.1:6677/
browser_lock → snapshot → click → unlock
```

场景清单：**[scenarios-yuzhousha.md](scenarios-yuzhousha.md)**

**Browser 抽检 ui-sim 产物（可选）**：

1. 从 `frontend/test/fixtures/yzs/settlements/` 挑 1 个 JSON，读 `state.message`
2. 开 identity_8 人机，快速结束或等 AI 打完
3. 对比页面 `.yzs__center-hint--result` 与 fixture 规则是否一致

---

## 前置条件


| 依赖            | 地址          | ui-sim 需要？               |
| ------------- | ----------- | ------------------------ |
| Go + cardtest | `backend/`  | **是**（纯 Go，不需 DB）        |
| MySQL/Redis   | 8088        | **否**（ui-sim 不启 HTTP 服务） |
| Node          | `frontend/` | **是**（vitest）            |
| Vite dev      | 6677        | 仅 Browser 测时需要           |


ui-sim **不需要**起后端 HTTP 或 MySQL——只在 Go 内存里 AI 自走并导出 JSON。

---

## 扩展：pending 窗口 fixture（后续）

ui-sim 当前只测 **终局结算**。pending 响应窗可复用同一模式：

1. 后端 `TestScenario`_* 导出 `phase=response` 的 PublicView
2. 前端 Vitest 测 `pendingCanSubmitPlay` / `centerHint`

---

## 安全

- 不起 Docker / MySQL / Redis（ui-sim 不需要；Browser 测时才可能要 run.sh，**问用户**）
- 不申请 `required_permissions: all`

---

## 报告模板

```markdown
## UI 测试报告
- **ui-sim**: `cd frontend && ./scripts/ui-sim.sh` → pass / fail
- **fixture 数**: N 个 JSON 通过 / M 失败
- **静态**: ui-smoke.sh → pass / fail
- **Browser**（如有）: 场景 ID → pass / fail
- **失败摘要**: fixture 名 → validateSettlementState 错误
- **与 card-test**: 是否还需 simrandom / smoke
```

