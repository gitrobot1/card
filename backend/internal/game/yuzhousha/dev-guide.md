# 宇宙杀（Yuzhousha）开发指南

> 面向后续 AI / 人类开发者：说明六阶段架构重构后的**代码落点、扩展规范、验收命令**。  
> 测试 Agent 另见 [`.cursor/skills/card-test/SKILL.md`](../../../../.cursor/skills/card-test/SKILL.md)。

---

## 1. 架构总览

宇宙杀采用 **「模式层 → 引擎层 → 技能层 → 前端注册表」** 分层，避免在 `useYzsGame.ts` / `engine.go` 里堆 `switch`。

```text
backend/internal/game/yuzhousha/
├── engine/                 # 对局状态机、出牌、响应、AI
│   ├── mode/               # 阶段 1–2：模式元数据、选目标、敌友判定
│   ├── solo_setup.go       # NewSolo / NewSolo1v1 / NewSolo2v2
│   └── skill_register*.go  # 复杂主动技 / 主公技状态机
├── skill/                  # 阶段 5：Decl 声明式被动技、武将/皮肤/扩展包数据
│   ├── catalog_skills.go   # 可迁移的被动 Decl
│   ├── catalog_peek.go     # 观星类准备阶段技
│   └── data/               # heroes / skins / packs JSON
└── test/yuzhousha/         # cardtest 外部测试（禁止在 internal 写 _test 依赖 testhook 以外的内部细节）

frontend/src/composables/yuzhousha/
├── pending/                # 阶段 4：response_mode → UI 行为注册表
├── animation/              # 阶段 6：event.type → 动画回放注册表
├── useYzsGame.ts           # 薄编排层，委托 pending registry
├── useYzsHints.ts          # 提示优先走 pendingRegistry
└── useYzsAnimations.ts     # replayEvent 委托 animation registry
```

**黄金法则**

1. 新 `response_mode` → 只改 `pending/handlers.ts`，不要在 `useYzsGame` 加分支。
2. 新被动技 → 优先 `catalog_skills.go` 的 `Decl` hook；复杂状态机才进 `engine/skill_*.go`。
3. 新事件动画 → 只改 `animation/handlers.ts`。
4. 新模式 → 先 `engine/mode/registry.go` 注册 `Meta`，再补 targeting / layout。
5. 合码前：`npm run build` + `./scripts/test.sh smoke`（改 yzs 必跑）。

---

## 2. 六阶段改动说明

### 阶段 1 — 模式注册表（ModeSpec）

| 项 | 路径 |
|----|------|
| 模式 ID、人数、布局 key、武将池 | `backend/.../engine/mode/registry.go` |
| 敌友 / 默认目标 | `backend/.../engine/mode/spec.go` |
| 前端布局组件 | `frontend/src/components/yuzhousha/layouts/index.ts` |

**扩展新模式**

1. 在 `mode/registry.go` 的 `init()` 里 `Register(Meta{...})`。
2. 在 `spec.go` / `targeting.go` 补该模式的 `EnemiesOf` / `ValidPlayTargets` 规则。
3. 前端增加 `layout_key` 对应 Vue 布局。
4. `solo_setup.go` 增加 `setupSoloXxx` 并在 `NewSolo` 分支。

当前已注册：`1v1`（`solo_1v1`）、`2v2`（`cross_2v2`）、`3p_chain`（`triangle_3p`）、`3p_ddz`（`triangle_3p`）。

---

### 阶段 2 — 选目标走 ModeSpec

| 项 | 路径 |
|----|------|
| 后端合法目标 | `engine/mode/targeting.go` |
| 前端目标高亮 | `useYzsTargeting.ts` |

**禁止**在 `play.go` 或前端写死「2 人局只能选对面」。一律通过 `mode.Is2v2(ctx)` / `EnemiesOf` / `AlliesOf` 判断。

---

### 阶段 3 — 武将数据外置 + 皮肤/扩展包

| 项 | 路径 |
|----|------|
| 武将 JSON | `skill/data/heroes/*.json` |
| 皮肤 JSON | `skill/data/skins/*.json` |
| 扩展包清单 | `skill/data/packs/*.json` |
| 加载与校验 | `skill/load_heroes.go`, `skins.go`, `packs.go` |
| 选将分页 API | `engine/heroes_catalog.go` |
| 前端展示解析 | `resolveYzsHeroDisplay.ts` |

**新增武将**：编辑 JSON → 确保 `pack` 与 `packs/*.json` 一致 → 跑 `go test -tags cardtest ./internal/game/yuzhousha/skill/...`。

**新增皮肤**：`id` 格式 `hero_id:skin_key`，在 `skins/*.json` 注册。

---

### 阶段 4 — Pending UI 注册表

所有 `pending.response_mode` 的交互（能否出牌、提交技能、提示文案、进入/离开清理）集中在：

```text
frontend/src/composables/yuzhousha/pending/
├── types.ts      # PendingHandler 接口
├── context.ts    # 共享 ref / API
├── handlers.ts   # 各 mode 实现（约 24 个 handler）
└── registry.ts   # findPendingHandler、canSubmit*、onModeChange
```

`pendingRegistry.ts` 仅为 re-export，便于 import。

**新增 response_mode  checklist**

1. 在 `handlers.ts` 增加 `{ modes: ['your_mode'], match, ... }`。
2. 按需实现：`canPlayCard` / `submitPlay` / `submitSkill` / `submitAction` / `hint` / `onEnter` / `onModeLeave`。
3. 若需新 UI 状态，在 `context.ts` 增加 ref，**不要**回到 `useYzsGame` 写大 switch。
4. 在 `useYzsHints.ts` 仅当 registry 无 hint 时才写通用兜底。
5. 浏览器走一遍该响应窗；后端用 `scenario_test.go` 或 `skill_test.go` 断言 `Pending.ResponseMode`。

---

### 阶段 5 — 后端技能 Decl 瘦身

| 项 | 路径 |
|----|------|
| 声明式被动技 | `skill/catalog_skills.go` |
| 观星/洛神类 | `skill/catalog_peek.go` |
| Hook 调度 | `engine/skill_hooks.go`, `skill_decl_hooks.go` |
| 复杂技 | `engine/skill_register*.go` + 专用 `skill_*.go` |
| 开发说明 | `skill/doc.go` |

**Runtime 新接口**（2v2 友好）：`PlayerCount`, `TeamOf`, `EnemiesOf`, `AlliesOf`, `DrawSkillCards`, `IsSeatInDyingRescue`。

**Decl 常用 hook**

| Hook | 用途 |
|------|------|
| `DrawCountBonus` / `OnTurnEnd` / `OnHandEmpty` | 摸牌数、回合结束、空手 |
| `EffectiveSuit` / `BlocksTrickTarget` / `BlocksPeachUse` | 红颜、帷幕、完杀 |
| `DamageAsHPLoss` / `ExtraResponsesNeeded` | 绝情、无双 |
| `SkipsDiscardPhase` / `OnCardResolved` | 克己、激昂 |
| `OnDamageDealt` / `OnCardsDiscarded` / `OnJudgeResult` | 反馈链、连营、洛神 |

**决策树**

```text
新技能
├─ 仅改数值/被动判定 → catalog_skills.go
├─ 准备阶段看牌堆顶 → catalog_peek.go
├─ 多步 UI + pending → engine 开 response_mode + 阶段 4 补前端 handler
└─ 主公技 / 交牌状态机 → skill_register*.go
```

改技能后：`./scripts/test.sh smoke -v`，涉及伤害链再跑 `./scripts/test.sh yzs -run TestScenario -v`。

---

### 阶段 6 — 事件动画注册表

```text
frontend/src/composables/yuzhousha/animation/
├── types.ts      # EventReplayHandler
├── handlers.ts   # 各 event.type 的 replay()
└── registry.ts   # replayRegisteredEvent
```

`useYzsAnimations.ts` 中 `replayEvent` 先调 registry；**批量** draw/discard 仍可在 `applyState` 里处理。

**新增事件动画**

1. 在 `handlers.ts` 增加 `{ types: [...], match, replay }`。
2. `match` 要精确，避免与已有 handler 抢事件（注册顺序 = 优先级）。
3. `npm run build` 确认 TypeScript 无 unused import。

---

## 3. 2v2 全量测试

2v2 与 1v1 共用引擎，但 **敌友、顺时针默认目标、濒死救援、主公技** 等路径不同。

### 3.1 后端自动化（必跑）

在 `backend/` 目录：

```bash
# 2v2 冒烟（全武将开局 + 模式单测）— 无需 CARD_SIM
./scripts/test.sh 2v2 -v

# 2v2 AI 自对弈（需 CARD_SIM=1）
CARD_SIM=1 ./scripts/test.sh sim2v2 -v

# 仅随机四人阵容（默认 40 种子，可用 CARD_SIM_ROUNDS 调整）
CARD_SIM=1 CARD_SIM_ROUNDS=40 ./scripts/test.sh sim2v2 -run TestSim_2v2_RandomQuadsSeeded -v

# 仅全武将 0 号位矩阵（32 局，较慢）
CARD_SIM=1 ./scripts/test.sh sim2v2 -run TestSim_2v2_AllHeroesAsSeat0 -v
```

| 测试 | 文件 | 说明 |
|------|------|------|
| `TestSmoke_2v2_AllHeroesBootstrap` | `smoke_2v2_test.go` | 每位武将作 0 号位开局，敌友/牌数不变量 |
| `TestSim_2v2_SingleQuick` | `sim_2v2_test.go` | 固定四人快速 AI 局（跟 `yzs` 套件跑） |
| `TestSim_2v2_AllHeroesAsSeat0` | `sim_2v2_test.go` | 全武将 × 随机队友/敌人 AI 自对弈（需 `CARD_SIM=1`） |
| `TestSim_2v2_RandomQuadsSeeded` | `sim_2v2_test.go` | 种子 1..N 随机四人阵容（需 `CARD_SIM=1`） |
| mode 单测 | `engine/mode/*_test.go` | 敌友、选目标、registry |

创建固定四人盘：`engine.NewSolo2v2WithHeroes(id, [4]string{seat0, seat1, seat2, seat3})`。

| 测试 | 验证 |
|------|------|
| `TestDefaultEnemy2v2Clockwise` | 默认攻击目标为顺时针下一名存活敌人 |
| `TestEnemiesOf2v2` / `TestAlliesOf2v2` | 座位 0/2 为友，1/3 为敌 |
| `TestValidPlayTargets2v2Sha` | 杀不能打队友；濒死敌人仍可选 |
| `TestValidateHeroForMode("2v2", ...)` / `mode/registry_test` | 标准包武将可选 |

**2v2 改动的回归（与 1v1 共用逻辑）**

```bash
./scripts/test.sh smoke -v          # 全武将 × 牌型矩阵（1v1 盘）
./scripts/test.sh yzs -v            # 全部 cardtest
# 大改后
CARD_SIM=1 ./scripts/test.sh sim -v # 1v1 AI 自对弈
```

### 3.3 手动全量清单（2v2 人机）

启动：`backend/scripts/run.sh` + `frontend/scripts/dev.sh` → 宇宙杀 → **十字阵对战**。

| # | 场景 | 预期 |
|---|------|------|
| 1 | 选将 | 标准包武将可选；皮肤/名字正常 |
| 2 | 开局布局 | 你在下、队友在上、敌将在左/右；4 人各 4 体力 |
| 3 | 回合顺序 | 你 → 敌左 → 队友 → 敌右 循环 |
| 4 | 杀 / AOE | 不能指定队友；万箭/南蛮正确跳过已死 |
| 5 | 桃园 / 五谷 | 可给队友回血/分牌 |
| 6 | 濒死 | 队友可出桃；完杀/救援类技能在 2v2 生效 |
| 7 | 主公技（若选刘备等） | 激将可拉队友；1v1 禁用技在 2v2 可用 |
| 8 | 响应窗 UI | 反馈/突袭/观星/无懈等 pending 提示与按钮正常 |
| 9 | 事件动画 | 出牌飞牌、伤害光束、判定、弃牌堆动画无卡死 |
| 10 | 胜利条件 | 两名敌将均阵亡后结算 |

### 3.4 前端构建

```bash
cd frontend && npm run build
```

确认 `YzsLayout2v2.vue`、十字阵 seat 映射、`mode=2v2` 选将路由无 TS 错误。

Sim 失败日志：`backend/test/yuzhousha/sim_logs/`（`CARD_SIM_TRACE=1` 可附带事件）。

### 3.5 3p 模式测试（杀上保下 / 斗地主）

```bash
# 冒烟 + mode 单测 — 无需 CARD_SIM
./scripts/test.sh 3p_chain -v
./scripts/test.sh 3p_ddz -v

# AI 自对弈 — 需 CARD_SIM=1
CARD_SIM=1 ./scripts/test.sh sim3p_chain -v
CARD_SIM=1 ./scripts/test.sh sim3p_ddz -v

# 四模式随机一键（1v1 + 2v2 + 3p_chain + 3p_ddz）
CARD_SIM=1 ./scripts/test.sh simrandom -v
CARD_SIM_ROUNDS=100 ./scripts/test.sh simrandom -v
```

| 测试 | 文件 | 说明 |
|------|------|------|
| `TestSmoke_3pChain_AllHeroesBootstrap` | `smoke_3p_test.go` | 全武将链式开局 |
| `TestSmoke_3pDdz_AllHeroesBootstrap` | `smoke_3p_test.go` | 全武将斗地主开局 |
| `TestSmoke_3pDdz_LandlordPerks` | `smoke_ddz_test.go` | 地主多摸牌、双杀 |
| `TestSim_3pChain_*` | `sim_3p_chain_test.go` | 链式 AI 自对弈 |
| `TestSim_3pDdz_*` | `sim_3p_ddz_test.go` | 斗地主 AI 自对弈 |

固定三人盘：

- 链式：`engine.NewSolo3pChainWithHeroes(id, [3]string{0,1,2})`
- 斗地主：`engine.NewSolo3pDdzWithHeroes(id, [3]string{landlord, f1, f2})`

---

## 4. 合码前 Checklist

```text
[ ] 新 response_mode → pending/handlers.ts + 浏览器点测
[ ] 新被动技 → catalog_skills.go 或 engine 状态机 + skill/doc.go 能力矩阵
[ ] 新 event → animation/handlers.ts
[ ] 2v2 相关 targeting/team → mode 单测
[ ] 3p 链式/斗地主 → ./scripts/test.sh 3p_chain -v && ./scripts/test.sh 3p_ddz -v
[ ] cd backend && ./scripts/test.sh smoke -v
[ ] cd backend && ./scripts/test.sh 2v2 -v
[ ] 大改 2v2：CARD_SIM=1 ./scripts/test.sh sim2v2 -v
[ ] 大改 3p：CARD_SIM=1 ./scripts/test.sh sim3p_chain -run TestSim_3pChain_SingleQuick -v && CARD_SIM=1 ./scripts/test.sh sim3p_ddz -run TestSim_3pDdz_SingleQuick -v
[ ] cd frontend && npm run build
[ ] 未提交 config.yaml / .env / sim_logs/*.log
```

---

## 5. 常见错误

| 现象 | 排查 |
|------|------|
| pending 无按钮 | `findPendingHandler` 是否 match；`skillOnly` / `suppressPlaySubmit` |
| 2v2 能杀队友 | `mode/targeting.go` `ValidPlayTargets` |
| 技能双触发 | engine 与 catalog 重复 Register；查 `skill_register_wu.go` 是否已删旧注册 |
| sim stuck | `test/yuzhousha/sim_logs/*.log` → Pending 章节 |
| 动画重复/不播 | animation handler 顺序与 `match` 重叠 |

---

## 6. 相关文档

- 技能框架：`skill/doc.go`
- 引擎文件索引：`engine/doc.go`
- 测试 Agent：`.cursor/skills/card-test/SKILL.md`
- Sim 日志说明：`backend/test/yuzhousha/sim_logs/README.md`
