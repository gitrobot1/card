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

当前已注册：`1v1`（`solo_1v1`）、`2v2`（`cross_2v2`）、`3p_chain`（`triangle_3p`）、`3p_ddz`（`triangle_3p`）、`3v3`（`hex_3v3`）、`identity_5`（`pentagon_5`）、`identity_8`（`octagon_8`）。

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

# 七模式随机一键（1v1 + 2v2 + 3p_chain + 3p_ddz + 3v3 + identity_5 + identity_8）
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

### 3.6 3v3 模式测试

```bash
# 冒烟 + mode 单测 — 无需 CARD_SIM
./scripts/test.sh 3v3 -v

# AI 自对弈 — 需 CARD_SIM=1
CARD_SIM=1 ./scripts/test.sh sim3v3 -v
CARD_SIM=1 ./scripts/test.sh sim3v3 -run TestSim_3v3_RandomHexesSeeded/1 -v   # 单种子复现
```

| 测试 | 文件 | 说明 |
|------|------|------|
| `TestSmoke_3v3_AllHeroesBootstrap` | `smoke_3v3_test.go` | 全武将作暖主帅开局 |
| `TestSmoke_3v3_SingleQuick` | `smoke_3v3_test.go` | 团队/主帅座位不变量 |
| `TestSim_3v3_SingleQuick` | `sim_3v3_test.go` | 固定六人快速 AI 局 |
| `TestSim_3v3_AllHeroesAsSeat0` | `sim_3v3_test.go` | 全武将 × 随机五人 AI 自对弈 |
| `TestSim_3v3_RandomHexesSeeded` | `sim_3v3_test.go` | 种子 1..N 随机六人阵容 |
| mode 单测 | `engine/mode/3v3_test.go` | 敌友、主帅胜负、禁闪电 |

固定六人盘：`engine.NewSolo3v3WithHeroes(id, [6]string{seat0..seat5})`（0 暖主帅、2 冷主帅）。

### 3.7 identity_5 五人身份局

```bash
# 冒烟 + mode 单测 + 主公技 — 无需 CARD_SIM
./scripts/test.sh identity_5 -v

# AI 自对弈 — 需 CARD_SIM=1
CARD_SIM=1 ./scripts/test.sh simidentity -v
CARD_SIM=1 ./scripts/test.sh simidentity -run TestSim_Identity5_RandomPentasSeeded/1 -v   # 单种子复现
```

| 测试 | 文件 | 说明 |
|------|------|------|
| `TestSmoke_Identity5_*` | `smoke_identity_test.go` | 全武将主公开局、身份分配、随机洗牌 |
| `TestScenario_Identity_*` | `scenario_identity_test.go` | 主公阵亡、内奸独活、主内单挑等结算场景 |
| `TestIdentity5_LordSkills*` | `smoke_lord_skills_test.go` | 主公技在 identity_5 可用 |
| `TestSim_Identity5_*` | `sim_identity_test.go` | 五人 AI 自对弈 |
| mode 单测 | `engine/mode/identity_test.go` | 身份校验、胜负、敌友 |

固定五人盘：`engine.NewSoloIdentity5WithHeroes(id, [5]string{heroes}, [5]string{roles})`；人机 `NewSoloIdentity5(id, name, heroID)`（0 号位固定主公）。

### 3.8 identity_8 八人身份局

```bash
# 冒烟 + mode 单测 + 主公技 — 无需 CARD_SIM
./scripts/test.sh identity_8 -v

# AI 自对弈 — 需 CARD_SIM=1
CARD_SIM=1 ./scripts/test.sh simidentity8 -v
CARD_SIM=1 ./scripts/test.sh simidentity8 -run TestSim_Identity8_RandomOctasSeeded/1 -v   # 单种子复现
```

| 测试 | 文件 | 说明 |
|------|------|------|
| `TestSmoke_Identity8_*` | `smoke_identity8_test.go` | 全武将主公开局、2 忠 1 内 4 反、随机洗牌 |
| `TestScenario_Identity8_*` | `scenario_identity8_test.go` | 与 identity_5 对应的结算场景 |
| `TestIdentity8_LordSkills*` | `smoke_lord_skills_test.go` | 主公技在 identity_8 可用 |
| `TestSim_Identity8_*` | `sim_identity8_test.go` | 八人 AI 自对弈（步数上限 30000） |
| mode 单测 | `engine/mode/identity_test.go` | `ValidateIdentity8Roles`、`TestTeamOf_Identity8`、`TestValidPlayTargets_Identity8AnyOther`、`TestEvaluateIdentityWin`、`IsIdentityMode` |

固定八人盘：`engine.NewSoloIdentity8WithHeroes(id, [8]string{heroes}, [8]string{roles})`；人机 `NewSoloIdentity8(id, name, heroID)`（0 号位固定主公，余座随机 2 忠 1 内 4 反）。前端布局 `YzsLayoutIdentity8.vue`（`octagon_8`）。

---

## 4. 合码前 Checklist

```text
[ ] 新 response_mode → pending/handlers.ts + 浏览器点测
[ ] 新被动技 → catalog_skills.go 或 engine 状态机 + skill/doc.go 能力矩阵
[ ] 新 event → animation/handlers.ts
[ ] 2v2 相关 targeting/team → mode 单测
[ ] 3p 链式/斗地主 → ./scripts/test.sh 3p_chain -v && ./scripts/test.sh 3p_ddz -v
[ ] 3v3 团队/主帅 → ./scripts/test.sh 3v3 -v
[ ] identity_5 身份/胜负 → ./scripts/test.sh identity_5 -v
[ ] identity_8 身份/胜负 → ./scripts/test.sh identity_8 -v
[ ] cd backend && ./scripts/test.sh smoke -v
[ ] cd backend && ./scripts/test.sh 2v2 -v
[ ] 大改 2v2：CARD_SIM=1 ./scripts/test.sh sim2v2 -v
[ ] 大改 3p：CARD_SIM=1 ./scripts/test.sh sim3p_chain -run TestSim_3pChain_SingleQuick -v && CARD_SIM=1 ./scripts/test.sh sim3p_ddz -run TestSim_3pDdz_SingleQuick -v
[ ] 大改 3v3：CARD_SIM=1 ./scripts/test.sh sim3v3 -run TestSim_3v3_SingleQuick -v
[ ] 大改 identity_5：CARD_SIM=1 ./scripts/test.sh simidentity -run TestSim_Identity5_SingleQuick -v
[ ] 大改 identity_8：CARD_SIM=1 ./scripts/test.sh simidentity8 -run TestSim_Identity8_SingleQuick -v
[ ] cd frontend && npm run build
[ ] 改 pending/布局/yzs UI → @card-ui-test 或见 [`.cursor/skills/card-ui-test/SKILL.md`](../../../../.cursor/skills/card-ui-test/SKILL.md)；静态：`cd frontend && ./scripts/ui-smoke.sh`
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

## 6. AOE / 锦囊标准开发流程（南蛮入侵范本）

> AOE（南蛮入侵、万箭齐发、桃园结义、铁索连环）遵循统一的逐人处理模式。
> 每个目标经历：**宣告 → 无懈窗口 → 效果结算 → 玩家响应 → 扣血 → 技能链 → 濒死 → 恢复**。
> **任何阶段都可能被未知的新阶段（技能窗口、判定、濒死等）打断，打断后必须能正确恢复。**

### 6.1 流程总览

```
使用锦囊（出牌）
├─ 1. 宣告 (announce) → 构建目标队列 → 发事件 → 启动第一个目标
├─ 2. 逐人无懈窗口 (wuxiek_trick)  ← 未知阶段可插入
├─ 3. 效果结算 → 南蛮需出杀/万箭需出闪/桃园回复/铁索横置
├─ 4. 玩家响应 (response) → 出牌或跳过
├─ 5. 扣血 (damage) → adjustDamageAmount → applyDamageWithHook
├─ 6. 技能链 (aftermath) ← 刚烈/反馈/奸雄等插入
├─ 7. 濒死 (dying) ← 濒死插入
└─ 8. 恢复 AOE → 继续下一个目标或完毕
```

### 6.2 核心数据结构

**`g.Pending`** — 当前阶段挂起状态，关键字段：`ResponseMode`（阶段类型）、`AoeQueue`（剩余队列）、`SavedPending`（濒死前保存的 Pending）、`WuxiekChain`（无懈链）。

**`DamageResume.AoeResume`** — AOE 恢复信息，不存 `g.Pending` 中（避免被技能阶段覆盖）：

```go
AoeResume struct {
    Source int; Amount int; Card Card; Rest []int; Active bool; Tiesuo bool
}
```

### 6.3 阶段详解

#### 宣告 + 逐人无懈窗口

```go
// play.go — resolveNanMan
queue := g.filterAoeQueue(g.aoeResponderQueue(source), CardNanMan)
*events = append(*events, GameEvent{Type: "nanman_announce", ...})
g.startNanManJueDou(source, queue[0], queue[1:], events)
```

```go
// play.go — startNanManJueDou（逐人无懈）
g.Pending = &PendingCombat{
    ResponseMode: ResponseModeWuxiekTrick,
    EffectTarget: target,     // 当前目标
    AoeQueue:     rest,       // ★ 剩余队列
    ResponseQueue: [...],     // 无懈响应队列
}
g.advanceToNextWuxiekResponder(events) // 逐人询问
```

无懈链：奇数抵消（跳过当前目标 → continueXxxAfterTarget），偶数生效（进入效果结算）。

#### 效果结算 + 响应 + 扣血

```go
// finalizeWuxiekChain — 偶数链生效
g.Pending = &PendingCombat{
    RequiredKind: CardSha,     // 需出杀
    AoeQueue:     aoeQueue,    // ★ 剩余队列
}
```

```go
// response.go — resolvePendingMiss（玩家未出牌，扣血）
pending := *g.Pending  // ★ 先复制
g.applyDamageWithHook(...)
if HP<=0 → afterDamageApplied → 濒死（Pending 保存到 SavedPending）

// ★ 把队列存入 resume，不依赖 g.Pending
resume.AoeResume = {Rest: pending.AoeQueue, Active: true}
g.continueAfterDamage(...) → 技能链
```

#### 技能链 + 濒死 + 恢复

```go
// skill_damage.go — continueAfterDamage
if g.isChained(target) && 属性伤害 → resume.AoeResume = {Tiesuo: true, Rest: chainSeats}
if HP<=0 → afterDamageApplied → 濒死
g.initDamageAftermath(...) → advanceDamageAftermath → 刚烈/反馈等
// 技能链完毕 → resumeAfterDamageNoSkill → 检查 AoeResume 恢复 AOE
```

```go
// skill_dying.go — startDyingWindow
saved := g.Pending  // ★ 保存当前 Pending
g.Pending = &PendingCombat{SavedPending: saved, ResponseMode: ResponseModeDying}
// 濒死结束 → restorePendingAfterDying(saved) → 恢复 AOE
```

### 6.4 状态保存策略（黄金法则）

| 信息 | 存储位置 | 原因 |
|------|---------|------|
| AOE 队列 | `DamageResume.AoeResume` | 不被 `g.Pending` 覆盖 |
| 濒死前状态 | `Pending.SavedPending` | 濒死嵌套恢复 |
| 技能链状态 | `g.damageAftermath` | 独立于 Pending |

所有恢复点（`restorePendingAfterDying`、`resumeAfterDamageNoSkill`、`advanceDamageAftermath` 结尾）都要检查 AOE 恢复。

### 6.5 开发新 AOE Checklist

```text
后端:
[ ] 宣告 + 目标队列 + 发事件
[ ] startXxxFor（无懈窗口，AoeQueue: rest）
[ ] finalizeWuxiekChain 分支（奇数跳过/偶数生效）
[ ] resolvePendingMiss 处理（扣血→濒死→技能链→AoeResume 恢复）
[ ] continueXxxAfterTarget（继续下一个）
[ ] restorePendingAfterDying 分支

前端:
[ ] 宣告事件动画 / 效果事件动画 / 无懈窗口 UI

测试:
[ ] 基础流程 / 无懈抵消+生效 / 技能打断 / 濒死打断 / 濒死+技能 / 多目标
```

### 6.6 常见 Bug

| Bug | 原因 | 预防 |
|-----|------|------|
| AOE 中断后不恢复 | `g.Pending` 被覆盖 | 队列存 `AoeResume`，不依赖 `g.Pending` |
| 濒死后 AOE 丢失 | 未保存 Pending | `SavedPending` + `restorePendingAfterDying` 分支 |
| 技能后 AOE 丢失 | 技能覆盖 Pending | `resumeAfterDamageNoSkill` 检查 `AoeResume` |
| 铁索传导中断 | 同上 | 用 `AoeResume` 传递队列 |
| 多重濒死嵌套 | SavedPending 被覆盖 | 每次 `startDyingWindow` 都保存 |
| 前端动画顺序错乱 | 先设 state 后播动画 | 动画中更新状态，初始 state 保留旧值 |
| AOE 响应窗口 ActorSeat 错误 | `finalizeWuxiekChain` 设置 Pending 后未调 `FillPendingRoles` | 设置 Pending 后立即调 `FillPendingRoles(g.Pending)` |
| 濒死 resume 丢失 | `handleHPChange` 中自动触发濒死，使用空 `DamageResume` | 濒死统一由 `afterDamageApplied` 处理，`handleHPChange` 只负责通知血量变化 |

### 6.7 AI 拿牌规则

AI 在需要从目标身上获取牌时（顺手牵羊、过河拆桥、反馈、突袭、奇袭等），统一通过 `aiPickTakeTarget`（`skill_tuxi.go`）选择目标牌，优先级如下：

1. **手牌区**：优先随机拿手牌（`zone="hand", cardID=""` 表示由 TakeWindow 自动选择）
2. **装备区**：手牌区为空时，从武器/防具/+1马/-1马中**随机**选一个非空槽位
3. **判定区**：装备区也空时，从判定区**随机**选一张

```go
func aiPickTakeTarget(g *Game, target int) (zone, cardID string) {
    p := &g.Players[target]
    // 1. 手牌区
    if len(p.Hand) > 0 { return "hand", "" }
    // 2. 装备区（随机非空槽位）
    equips := 收集所有非空装备槽
    if len(equips) > 0 { return 随机选一个 }
    // 3. 判定区（随机）
    if len(p.JudgeArea) > 0 { return 随机选一张 }
    return "", ""
}
```

所有 TakeWindow 类 AI 操作（包括 `autoTakeWindowIfNeeded`）都使用此规则。新增需要从别人身上拿牌的技能时，直接调用 `aiPickTakeTarget` 即可。

### 6.7 关键文件索引

| 文件 | 关键函数 |
|------|---------|
| `play.go` | `resolveNanMan/WanJian/TaoYuan/TieSuoAOE`, `startXxxFor`, `continueXxxAfter`, `advanceToNextWuxiekResponder`, `finalizeWuxiekChain`, `restorePendingAfterDying` |
| `response.go` | `RespondWuxiek`, `PassResponse`, `resolvePendingMiss` |
| `skill_damage.go` | `DamageResume`（含 `AoeResume`）, `continueAfterDamage`, `advanceDamageAftermath`, `resumeAfterDamageNoSkill` |
| `skill_dying.go` | `startDyingWindow`, `playTaoForDying`, `resolveDyingSaved/Death` |
| `phase_hp_change.go` | `applyDamageWithHook`, `handleHPChange` |
| `card_equipment.go` | `startTiesuoAoe`, `continueTiesuoAoe`, `finishTiesuoAoe` |
| `card_tricks_ext.go` | `resolveTieSuoAOE`, `startTieSuoFor`, `continueTieSuoAfter` |
| `skill_tianxiang.go` | `finalizeDamageHit` |
| `model.go` | `PendingCombat` 结构体 |

### 6.8 阶段嵌套实战案例

> 以下是一个六人场万箭齐发的完整结算推演，展示了**宣告 → 无懈 → 出闪/不出闪 → 扣血 → 濒死 → 技能 → 恢复AOE** 的嵌套关系。
> 特别关注濒死和技能阶段如何插入 AOE 流程，以及状态如何保存和恢复。

> **验证状态**：此场景有对应的自动化测试 `TestScenario_WanJian_6pAoeWithDyingAndSkills`，位于 `test/yuzhousha/scenario_wanjian_6p_test.go`。
> 修复了以下 bug 后才跑通：`finishShanDodgeSuccess` 恢复 AOE、`FillPendingRoles` 缺失、`handleHPChange` 自动濒死导致 AoeResume 丢失、`PassResponse` 缺少刚烈 choice 处理。

#### 初始状态

| 座位 | 武将 | HP | 手牌 | 关键技能 |
|------|------|-----|------|------|
| 0 | 陆逊 | 3/3 | 万箭齐发 | 连营（失去最后一张手牌时摸1） |
| 1 | 张角 | 3/3 | 闪、黑桃2杀 | 雷击、鬼道 |
| 2 | 司马懿 | 1/3 | 桃×2 | 反馈（受伤拿来源1牌）、鬼才（手牌换判定） |
| 3 | 郭嘉 | 1/3 | 桃 | 遗计（受伤后摸2） |
| 4 | 夏侯惇 | 1/3 | 桃 | 刚烈（受伤后判定，非红桃则来源弃2或受1伤） |
| 5 | 张春华 | 3/3 | 无 | 绝情、伤逝（锁定技，手牌数 < 已损失体力时补牌） |

> 注：司马懿给2桃是因为鬼才阶段会消耗1桃换判定牌。

#### 完整事件序列

```
宣告
├── 事件1: 万箭齐发宣告
│   ├── 陆逊手牌 1→0，打出万箭
│   ├── 陆逊【连营】：摸1，手牌 0→1
│   └── 队列 [张角,司马懿,郭嘉,夏侯惇,张春华]

逐人处理 #1 — 张角 (HP 3/3)
├── 事件2-6: 无懈窗口 → 都跳过（0张，生效）
├── 事件7: 张角出闪，手牌 2→1（剩黑桃2杀）
├── 张角【雷击】判定劈陆逊 → 非黑桃不生效
└── 完毕，继续下一个

逐人处理 #2 — 司马懿 (HP 1/3)
├── 事件8-12: 无懈窗口 → 都跳过
├── 事件13: 司马懿无闪（手牌是桃），跳过 → 扣血
├── 事件14: 扣血 HP 1→0
│
├── ╔══ 濒死阶段 ══╗
│   ║ ★ SavedPending = [郭嘉,夏侯惇,张春华]（保存AOE队列）
│   ║ 事件15: 司马懿濒死（需1桃）
│   ║ 司马懿自己出桃，HP 0→1，脱离濒死
│   ╚══════════════╝
│
├── 事件16: 司马懿【反馈】→ 从陆逊拿1牌
│   ├── 陆逊手牌 1→0 → 陆逊【连营】→ 手牌 0→1
│   └── 司马懿手牌 0→1
│
├── ★ restorePendingAfterDying → AOE恢复 [郭嘉,夏侯惇,张春华]
└── 继续下一个

逐人处理 #3 — 郭嘉 (HP 1/3)
├── 事件17-21: 无懈窗口 → 都跳过
├── 事件22: 郭嘉无闪（手牌是桃），跳过 → 扣血
├── 事件23: 扣血 HP 1→0
│
├── ╔══ 濒死阶段 ══╗
│   ║ ★ SavedPending = [夏侯惇,张春华]
│   ║ 事件24: 郭嘉濒死（需1桃）
│   ║ 郭嘉自己出桃，HP 0→1，脱离濒死
│   ╚══════════════╝
│
├── 事件25: 郭嘉【遗计】→ 摸2牌，手牌 0→2
├── ★ restorePendingAfterDying → AOE恢复 [夏侯惇,张春华]
└── 继续下一个

逐人处理 #4 — 夏侯惇 (HP 1/3)
├── 事件26-30: 无懈窗口 → 都跳过
├── 事件31: 夏侯惇无闪（手牌是桃），跳过 → 扣血
├── 事件32: 扣血 HP 1→0
│
├── ╔══ 濒死阶段 ══╗
│   ║ ★ SavedPending = [张春华]
│   ║ 事件33: 夏侯惇濒死（需1桃）
│   ║ 夏侯惇自己出桃，HP 0→1，脱离濒死
│   ╚══════════════╝
│
├── 事件34: 夏侯惇【刚烈】→ 判定，非红桃
│   ├── 陆逊只有1手牌，不够弃2张 → 选择受伤
│   └── 事件35: 陆逊受刚烈1伤，HP 3→2
│       （连营不触发——陆逊还有手牌，不是"失去最后一张"）
│
├── ★ restorePendingAfterDying → AOE恢复 [张春华]
└── 继续下一个

逐人处理 #5 — 张春华 (HP 3/3)
├── 事件36-40: 无懈窗口 → 都跳过
├── 事件41: 张春华无手牌，跳过 → 扣血
├── 事件42: 扣血 HP 3→2（已损失体力=1）
│   （绝情不触发——绝情是她造成伤害时，这里是别人对她造成伤害）
├── 事件43: 张春华【伤逝】→ 手牌数 0 < 已损失体力 1
│   → 摸1牌，手牌 0→1
└── 完毕 → AOE队列空 → 万箭完毕，陆逊继续出牌
```

#### 最终状态（多局测试验证通过）

| 座位 | 武将 | HP | 手牌 | 实际 | 说明 |
|------|------|-----|------|------|------|
| 0 | 陆逊 | 2/3 | 1 | HP✓ | 刚烈反伤1血，连营摸牌 |
| 1 | 张角 | 3/3 | 1 | ✓ | 出闪剩黑桃2杀 |
| 2 | 司马懿 | 1/3 | 1 | ✓ | 鬼才用1桃，濒死用1桃，反馈拿牌 |
| 3 | 郭嘉 | 1/3 | 2 | ✓ | 濒死自救，遗计摸2 |
| 4 | 夏侯惇 | 1/3 | 0 | ✓ | 濒死自救，刚烈判定→陆逊受1伤 |
| 5 | 张春华 | 2/3 | 1 | ✓ | 扣1血，伤逝摸1 |

#### 阶段嵌套树

```
万箭齐发 AOE 队列 [张角, 司马懿, 郭嘉, 夏侯惇, 张春华]
│
├── 张角 → 出闪 → 雷击判定 → 完毕
│
├── 司马懿 → 无闪 → 扣血 HP=0
│   └── 【濒死】SavedPending=[郭嘉,夏侯惇,张春华]
│       ├── 出桃 → HP=1
│       ├── 【反馈】拿陆逊牌 → 陆逊连营
│       └── restorePendingAfterDying → 恢复
│
├── 郭嘉 → 无闪 → 扣血 HP=0
│   └── 【濒死】SavedPending=[夏侯惇,张春华]
│       ├── 出桃 → HP=1
│       ├── 【遗计】摸2牌
│       └── restorePendingAfterDying → 恢复
│
├── 夏侯惇 → 无闪 → 扣血 HP=0
│   └── 【濒死】SavedPending=[张春华]
│       ├── 出桃 → HP=1
│       ├── 【刚烈】判定 → 陆逊受1伤
│       └── restorePendingAfterDying → 恢复
│
└── 张春华 → 无闪 → 扣血 HP=2(损失1)
    └── 【伤逝】手牌0<损失1 → 摸1 → 完毕
```

#### 关键设计要点

1. **AOE 队列不存 `g.Pending`**：濒死和技能都会覆盖 `g.Pending`，队列必须通过 `SavedPending`（濒死恢复）和 `DamageResume.AoeResume`（技能链恢复）传递。

2. **濒死嵌套**：每次 `startDyingWindow` 都保存当前 `g.Pending` 到 `SavedPending`。濒死结束后 `restorePendingAfterDying` 根据 `SavedPending` 的类型选择正确恢复方式。

3. **`handleHPChange` 不触发濒死**：濒死应统一由 `afterDamageApplied` 处理，以便传递正确的 `DamageResume`（含 `AoeResume`）。`handleHPChange` 作为底层函数，只负责通知血量变化和触发钩子。

4. **设置 Pending 后立即调 `FillPendingRoles`**：`finalizeWuxiekChain` 中设置万箭/南蛮响应 Pending 后必须调用，否则 `ActorSeat` 默认为0，导致 AI 无法自动处理。

5. **PassResponse 需覆盖所有技能模式**：新增技能 choice 模式时，`PassResponse` 的 switch 中必须有对应分支，否则走 default → `resolvePendingMiss` 会错误地清空 Pending。

6. **被动技可随时触发**：连营在"失去最后一张手牌"时触发，可能在宣告阶段、反馈阶段等任意位置。

7. **技能描述要精确**：绝情是"她**造成**的伤害视为体力流失"，伤逝是**锁定技**"手牌数 < 已损失体力时补牌"。开发时必须严格区分"造成"和"受到"、区分"一次性触发"和"持续锁定条件"。

---

## 7. 相关文档

- 技能框架：`skill/doc.go`
- **交互窗口 / Pending 语义**：[`dev-interaction-window.md`](./dev-interaction-window.md)
- 测试 Skill：`.cursor/skills/card-test/SKILL.md`
- Sim 日志说明：`backend/test/yuzhousha/sim_logs/README.md`
