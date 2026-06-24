## 8. 阶段栈重构实施（v8.0）

> **参考来源**：`/Users/time/Project/noname/card-ai-docs/` 目录下的无名杀 AI 参考文档。
> **原则**：借鉴 noname 的规则完整性，保留本项目的分层架构和声明式技能框架。
> **状态**：🟡 进行中

### 8.0 无名杀参考文档索引

| 编号 | 文件 | 内容 | 对应重构步骤 | 状态 |
|------|------|------|-------------|------|
| 01 | `01-event-lifecycle.md` | GameEvent 事件生命周期 | Step A | ✅ 已完成 |
| 02 | `02-trigger-system.md` | 技能触发-响应系统 | Step F（后续） | ⬜ 未开始 |
| 03 | `03-phase-flow.md` | 阶段流转 | Step B | ✅ 已完成 |
| 04 | `04-damage-system.md` | 伤害结算 | Step C | ✅已完成 |
| 05 | `05-judge-system.md` | 判定与改判 | Step D | ✅ 已完成（v8.1 重构） |
| 06 | `06-trick-card.md` | 锦囊牌系统 | Step D | 未检查 |
| 07 | `07-wuxie-system.md` | 无懈可击 | Step D | 未检查 |
| 08 | `08-delay-trick.md` | 延迟锦囊 | Step D | 未检查 |
| 09 | `09-weapon-equip.md` | 武器与装备 | Step D v8.2 | ✅ 已完成 |
| 10 | `10-skill-structure.md` | 技能定义结构 | Step F | ⬜ 未开始 |
| 11 | `11-card-distance.md` | 变牌系统与距离计算 | Step F（后续） | ⬜ 未开始 |

---

### Step A: 核心状态机重构 ✅ 已完成

> 参考 noname 01-event-lifecycle.md

**noname 设计要点**：
- 每个事件有固定生命周期钩子：Before → Begin → Content → End → After
- 事件栈追踪嵌套：子事件自动入栈/出栈
- 跳过机制：`skipList` + `checkSkipped()`
- 取消机制：`cancel()` 在 Begin 前走 Omitted
- 控制方法：`goto(N)` / `redo()` / `finish()` / `cancel()`

**新增文件**：
| 文件 | 说明 | 行数 |
|------|------|------|
| `engine/game_event.go` | GameEvent 核心状态机 | 400+ |
| `engine/phase_stack.go` | 阶段栈（已废弃，被 GameEvent 替代） | 保留 |

**修改文件**：

| 文件 | 改动 |
|------|------|
| `engine/game.go` | 新增 `eventManager *EventManager`、`RoundNumber int` |
| `engine/model.go` | 新增 `SkipPhases *SkipList`、`TurnedOver bool` |

**验收**：
```bash
grep "GameEventInstance\|EventPhase\|EventManager\|StartEvent\|checkEventSkipped\|FinishEvent\|CancelEvent\|runEventLoop\|PushChildEvent\|FinishCurrentPhaseEvent" engine/game_event.go
# 应有 15+ 个定义
```

---

### Step B: 阶段流转数据驱动 ✅ 已完成

> 参考 noname 03-phase-flow.md

**noname 设计要点**：
- `phaseList` 是数据数组，不是硬编码函数链
- 每个阶段是独立 GameEvent
- `phase()` 函数 step 0-13 完整回合时序
- `player.skip("phaseUse")` → `checkSkipped()` 自动跳过

**新增文件**：
| 文件 | 说明 | 行数 |
|------|------|------|
| `engine/phase_flow.go` | 阶段流转系统 | 290+ |

**修改文件**：

| 文件 | 改动 |
|------|------|
| `engine/turn.go` | 重写 `beginTurn`（step 0-13），新增 `tryAdvanceRound`、`triggerPhaseHook`、`RoundNumber` |
| `engine/phase_prepare.go` | `advanceToPlayPhase` 检查 SkipPlay；`advanceToDiscardPhase` 清理标记；判定阶段接入 GameEvent；乐生效用 `SkipToPhase` |
| `engine/skill_judge.go` | `completeJudgeResume` 用 `PopPhase` |

**验收**：
```bash
grep "phaseList\|PhaseDef\|startPhaseLoop\|runPhaseStep\|SkipToPhase\|IsTurnedOver\|TurnOver\|finishPhaseLoop\|tryAdvanceRound\|triggerPhaseHook" engine/phase_flow.go engine/turn.go
# 应有 10+ 个定义
```

---

### Step C: 伤害系统重构 ✅ 已完成

> 参考 noname 04-damage-system.md + 01-event-lifecycle.md

**noname 设计要点**：
- 伤害事件有完整生命周期：`damageBegin1~4 → damage → [濒死检测] → damageEnd`
- 扣血后**自动检查濒死**：`if (hp <= 0) player.dying(event)` — 不在外部手动调用

**新增文件**：
| 文件 | 说明 |
|------|------|
| `engine/damage_event.go` | `StartDamageEvent` — 伤害 GameEvent（完整生命周期：damageBegin→Content扣血→濒死→damageEnd） |

**noname 对应**：
| noname 步骤 | 实现 | 状态 |
|------------|------|------|
| step 0-3: damageBegin1~4 | `OnBefore` 回调（预留，后续 trigger 接入） | ✅ |
| step 4: 扣血 + damage 触发 | `Content`: `applyDamage` + `handleHPChange` | ✅ |
| step 5: 自动濒死 | `afterDamageApplied`（后续迁移到 GameEvent 子事件） | ✅ |
| step 6: damageSource | `OnAfter` 回调（预留） | ✅ |
| damageEnd（刚烈/反馈） | `OnEnd` 回调（预留） | ✅ |

**迁移清单**（10个文件，`applyDamageWithHook` + 手动 `afterDamageApplied` → `ApplyDamageAndCheckDeath` → `StartDamageEvent`）：
| 文件 | 状态 |
|------|------|
| `engine/phase_prepare.go`（闪电判定） | ✅ |
| `engine/card_tricks_ext.go`（火攻） | ✅ |
| `engine/skill_tianxiang.go`（天香/杀命中） | ✅ |
| `engine/skill_fankui.go`（闪电） | ✅ |
| `engine/response.go`（AOE 南蛮/万箭） | ✅ |
| `engine/skill_ganglie.go`（刚烈反伤） | ✅ |
| `engine/card_equipment.go`（铁索传导） | ✅ |
| `engine/skill_jiaxu.go`（乱武） | ✅ |
| `engine/skill_zhangjiao.go`（雷击） | ✅ |
| `engine/weapons.go`（贯石斧） | ✅ |

**验收**：
```bash
grep -r "applyDamageWithHook" engine/ --include="*.go" | grep -v "_test.go" | grep -v phase_hp_change.go
# 应输出 0 行（只剩定义处）
grep -rn "ApplyDamageAndCheckDeath" engine/ --include="*.go" | grep -v "_test.go"
# 应输出 10+ 行
```

---

### Step D: 判定系统修复 + v8.1 完整重构 ✅ 已完成

> 参考 noname 05-judge-system.md + 06-trick-card.md + 07-wuxie-system.md + 08-delay-trick.md

#### D 初版：Bug 修复

| # | 工作内容 | 文件 | 验收 | 状态 |
|---|---------|------|------|------|
| D1 | 乐不思蜀判定 `suit=="H"`（非 `isRedSuit`） | `phase_prepare.go` | `grep "isHeart" phase_prepare.go` | ✅ |
| D2 | rankLabel 14→A, 15→2 | `skill_fanjian.go` | `grep "case 14:" skill_fanjian.go` | ✅ |
| D3 | normalizeRank（14→1, 15→2） | `skill_fanjian.go` | `grep "func normalizeRank" skill_fanjian.go` | ✅ |
| D4 | isLightningStrike 用 normalizeRank | `judge.go` | `grep "normalizeRank" judge.go` | ✅ |
| D5 | advanceJudgeWuxiekQueue 传 EffectTarget | `response.go` | `grep "EffectTarget" response.go` | ✅ |
| D6 | 判定无懈窗口 AI 直接跳过 | `play.go` | `grep "isJudgeWuxiekMode" play.go` | ✅ |
| D7 | 无懈窗口结束→executeJudge | `phase_prepare.go` | `grep "executeJudge.*EffectTarget" phase_prepare.go` | ✅ |
| D8 | AI pickTrickTarget 只选敌人 | `ai.go` | `grep "enemiesOf" ai.go` | ✅ |
| D9 | 前端 judgeResultHandler 更新判定区 | `handlers.ts` | TypeScript 编译通过 | ✅ |
| D10 | 前端 play_phase_skip handler | `handlers.ts` | TypeScript 编译通过 | ✅ |
| D11 | 前端 draw_phase_skip handler | `handlers.ts` | TypeScript 编译通过 | ✅ |

#### D v8.1: 完整重构（2026-06-24）

> 严格按照 noname 05-judge-system.md 重构判定系统架构。

**noname 判定流程**：
```
取牌 → 亮出 → trigger("judge") 改判介入
→ 构建 result{card,name,number,suit,color,bool}
→ judge函数计算(>0成功/<0失败/=0无结果)
→ mod.judge 被动修改 → trigger("judgeFixing") → callback
```

**重构内容**：

| # | 工作内容 | 文件 | 状态 |
|---|---------|------|------|
| D12 | `JudgeResult` 完整结构（card/name/number/suit/color/bool/judge） | `skill/hooks.go` | ✅ |
| D13 | `JudgeFunc` 判定函数类型 | `skill/hooks.go` | ✅ |
| D14 | `ModJudgeCtx` + `HookModJudge` + `HookJudgeFixing` | `skill/hooks.go` | ✅ |
| D15 | `OnModJudge` 添加到 Decl/Handler | `skill/types.go` | ✅ |
| D16 | `runModJudgeHooks` + `runJudgeFixingHooks` | `engine/skill_hooks.go` | ✅ |
| D17 | `buildJudgeResult` 构建完整判定结果 | `engine/skill_judge.go` | ✅ |
| D18 | 8 个判定函数：Lebu/Bingliang/Shandian/Bagua/Tieqi/Ganglie/Luoshen/Leiji | `engine/skill_judge.go` | ✅ |
| D19 | `completeJudgeResume` 走完整流程（构建result→mod.judge→judgeFixing→callback） | `engine/skill_judge.go` | ✅ |
| D20 | `executeJudge` 接入完整判定流程 | `engine/phase_prepare.go` | ✅ |
| D21 | 删除旧函数 `resolveLebuJudge`/`resolveBingliangJudge`/`resolveShandianJudge` | `engine/phase_prepare.go` | ✅ |
| D22 | `startJudge` 签名改为接受 `JudgeFunc` | `engine/skill_judge.go` | ✅ |
| D23 | 所有调用方更新（zhangjiao/ganglie/tieqi/bagua/luoshen/ddz/ai/testhook） | 7 个文件 | ✅ |

**判定系统完整性对照 noname**：

| noname 要求 | 我们的实现 | 状态 |
|---|---|---|
| 取牌 → 亮出 → trigger("judge") 改判 | `flipJudgeCard` → `collectModifyJudgeSeats` → 队列询问 | ✅ |
| 构建 result{card,name,number,suit,color,bool} | `JudgeResult` 结构体 + `buildJudgeResult` | ✅ |
| judge 函数计算 (>0成功/<0失败/=0无结果) | 8 个 `JudgeFunc` 函数 | ✅ |
| mod.judge 被动修改 | `HookModJudge` + `OnModJudge` | ✅ |
| trigger("judgeFixing") 最终确认 | `runJudgeFixingHooks`（预留） | ✅ |
| callback 回调 | `completeJudgeResume` → resume 分发 | ✅ |
| 鬼才/鬼道改判队列 | 按座位顺序、每人一次、替换牌逻辑 | ✅ |

**验收**：
```bash
go build ./internal/game/yuzhousha/...
# 编译通过

# 验证判定函数
grep "func judgeFunc" engine/skill_judge.go
# judgeFuncLebu, judgeFuncBingliang, judgeFuncShandian,
# judgeFuncBagua, judgeFuncTieqi, judgeFuncGanglie,
# judgeFuncLuoshen, judgeFuncLeiji

# 验证 mod.judge 钩子
grep "HookModJudge" skill/hooks.go
grep "OnModJudge" skill/types.go
grep "runModJudgeHooks" engine/skill_hooks.go
```

---

### Step E: 回合完整时序 ⬜ 待测试

> 已实现，需冒烟测试验证

| # | 工作内容 | 状态 |
|---|---------|------|
| E1 | `go build` 编译通过 | ✅ |
| E2 | `./scripts/test.sh smoke -v` 冒烟测试 | ⬜ |
| E3 | `./scripts/test.sh yzs -v` 全量测试 | ⬜ |
| E4 | 手动测试：乐不思蜀判定生效跳过出牌 | ⬜ |
| E5 | 手动测试：闪电判定伤害+濒死 | ⬜ |
| E6 | 手动测试：2v2 全流程 | ⬜ |

---

### Step F: 技能触发系统重构 ⬜ 未开始

> **参考**：noname `02-trigger-system.md` + `10-skill-structure.md`
> **原则**：借鉴 noname 的四角色维度、三段结构、优先级系统，不抄运行时遍历
> **日期**：2026-06-24

---

#### 设计原则

| noname | 我们 | 借鉴？ |
|---|---|---|
| 运行时遍历所有玩家技能 | 声明式 DeclHook 注册表，编译时注册 | ❌ 不抄 |
| 字符串匹配触发时机 `trigger: {player:"damageEnd"}` | 类型安全的 `HookKind` 枚举 | ❌ 不抄 |
| `arrangeTrigger` 事件排序 | `runSkillHooks` 统一分发 | ❌ 不抄 |
| **四角色维度** player/source/target/global | 当前只有 player 维度 | ✅ 借鉴 |
| **三段结构** filter → cost → content | 当前无 filter/cost 分离 | ✅ 借鉴 |
| **优先级系统** priority + firstDo/lastDo | 当前无 | ✅ 借鉴 |
| **技能类型标记** forced/limited/equipSkill | 当前无 | ✅ 借鉴 |

---

#### Phase A: 框架增强（借鉴 noname 优点）

##### A1. 技能类型标记

在 `Decl` 中添加 `Tags` 字段，标记技能类型：

```
Decl{Tags: []SkillTag{TagForced, TagLimited, TagEquipSkill}}
```

| 标记 | 含义 | 来源 |
|---|---|---|
| `TagForced` | 锁定技：自动触发，不询问玩家 | noname `forced` |
| `TagLimited` | 限定技：一局只能发动一次 | noname `limited` |
| `TagAwaken` | 觉醒技：条件满足自动觉醒 | noname `awaken` |
| `TagLord` | 主公技 | noname `lord` |
| `TagEquipSkill` | 装备附带技能：卸下装备时自动移除 | noname `equipSkill` |

**文件**：`skill/types.go` — 新增 `SkillTag` 类型和常量，`Decl` 添加 `Tags []SkillTag`

##### A2. 四角色维度

当前 `runSkillHooks` 只查询事件主体玩家的技能（`playerSkillHandlers(ctx.Seat)`）。借鉴 noname 的四角色维度，增加 `HookRole` 参数：

```go
type HookRole int
const (
    RolePlayer HookRole = iota  // 事件主体（当前已有）
    RoleSource                   // 事件来源（noname: source）
    RoleTarget                   // 事件目标（noname: target）
    RoleGlobal                   // 全局监听（noname: global）
)
```

**`runSkillHooks` 改造**：
- `RolePlayer`：查询 `ctx.Seat` 的技能（当前行为）
- `RoleSource`：查询 `ctx.Source` 的技能
- `RoleTarget`：查询 `ctx.Target` 的技能
- `RoleGlobal`：遍历所有存活玩家的技能

**文件**：`skill/hooks.go` — 新增 `HookRole`，`HookCall` 添加 `Role` 字段；`engine/skill_hooks.go` — `runSkillHooks` 按 Role 分发

##### A3. 技能优先级

在 `Decl` 中添加 `Priority` 字段：

```
Decl{Priority: 5, FirstDo: true, LastDo: false}
```

- `Priority`：数字越大越先执行（默认 0）
- `FirstDo`：始终最先执行（如无懈可击，noname `firstDo`）
- `LastDo`：始终最后执行

`runSkillHooks` 收集技能后按 `Priority` 降序排序，`FirstDo` 排最前，`LastDo` 排最后。同优先级保持注册顺序（不抄 noname 的同优先级玩家选择）。

**文件**：`skill/types.go` — `Decl` 添加 `Priority`/`FirstDo`/`LastDo`；`engine/skill_hooks.go` — 添加排序逻辑

---

#### Phase B: Hook 补全

##### B1. 补齐缺失 HookKind

对照 noname 事件生命周期，缺失的 Hook：

| HookKind | noname 对应 | 触发场景 |
|---|---|---|
| `HookShaBegin` | `shaBegin` | 杀开始结算（仁王盾触发点） |
| `HookShaMiss` | `shaMiss` | 杀被闪抵消（青龙刀/贯石斧触发点） |
| `HookShaHit` | `shaHit` | 杀命中（麒麟弓触发点） |
| `HookUseCard` | `useCard` | 使用牌（集智触发点） |
| `HookUseCardToTarget` | `useCardToTarget` | 牌指定目标后（雌雄双股剑触发点） |
| `HookDamageBegin` | `damageBegin1~4` | 伤害开始（白银狮子触发点） |
| `HookDamageEnd` | `damageEnd` | 伤害结束（刚烈/反馈触发点） |
| `HookPhaseBegin` | `phaseBegin` | 阶段开始 |
| `HookPhaseEnd` | `phaseEnd` | 阶段结束 |

**文件**：`skill/hooks.go` — 新增 HookKind 常量；`skill/types.go` — Decl 添加对应回调

##### B2. 在引擎节点插入 HookCall

| 节点 | Hook | 文件 |
|---|---|---|
| `playShaWithCard` 创建 Pending 后 | `HookShaBegin` | `play.go` |
| `finishShanDodgeSuccess` 杀被闪后 | `HookShaMiss` | `skill_zhangjiao.go` |
| `finalizeDamageHit` 伤害命中后 | `HookShaHit` | `skill_tianxiang.go` |
| `playTrickWithCard` 锦囊使用后 | `HookUseCard` | `play.go` |
| `advanceShaBeforeTargetResponse` 指定目标后 | `HookUseCardToTarget` | `skill_pojun.go` |
| `damage_event.go` Content 扣血前 | `HookDamageBegin` | `damage_event.go` |
| `finalizeDamageHit` 伤害结算完后 | `HookDamageEnd` | `skill_tianxiang.go` |

---

#### Phase C: 验证与迁移

##### C1. 验证现有技能

确保在新框架下正确触发：
- 刚烈（`damageEnd` → `HookDamageEnd`）
- 反馈（`damageEnd` → `HookDamageEnd`）
- 鬼才（`global: "judge"` → 改判队列）
- 洛神（`phaseZhunbeiBegin` → `HookPhaseBegin`）
- 铁骑（`useCardToTarget` → `HookUseCardToTarget`）

##### C2. 改判队列迁移到 Hook

当前鬼才/鬼道硬编码在 `collectModifyJudgeSeats` 中。改为 `HookJudge`（global 角色），技能通过 Decl 注册，不再硬编码。

**文件**：`engine/skill_judge.go` — `collectModifyJudgeSeats` 改为查询 `HookJudge` global handlers

---

### 验收命令汇总

```bash
# 每步完成后必跑
cd backend && go build ./...

# Step E 完成后
./scripts/test.sh smoke -v
./scripts/test.sh yzs -v

# Step F 完成后（全量）
CARD_SIM=1 ./scripts/test.sh sim2v2 -run TestSim_2v2_SingleQuick -v
cd frontend && npm run build
```
