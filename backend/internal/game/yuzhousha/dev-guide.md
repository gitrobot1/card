## 8. 阶段栈重构实施（v8.0）

> **参考来源**：`/Users/time/Project/noname/card-ai-docs/` 目录下的无名杀 AI 参考文档。
> **原则**：借鉴 noname 的规则完整性，保留本项目的分层架构和声明式技能框架。
> **状态**：🟡 进行中

### 8.0 无名杀参考文档索引

| 编号 | 文件 | 内容 | 对应重构步骤 | 状态 |
|------|------|------|-------------|------|
| 01 | `01-event-lifecycle.md` | GameEvent 事件生命周期 | Step A | ✅ 已完成 |
| 02 | `02-trigger-system.md` | 技能触发-响应系统 | Step F + Step 11（电梯式重构） | 🟡 进行中 |
| 03 | `03-phase-flow.md` | 阶段流转 | Step B | ✅ 已完成 |
| 04 | `04-damage-system.md` | 伤害结算 | Step C + Step 10（卖血技重构） | ✅ 已完成 |
| 05 | `05-judge-system.md` | 判定与改判 | Step D | ✅ 已完成（v8.1 重构） |
| 06 | `06-trick-card.md` | 锦囊牌系统 | Step D | 未检查 |
| 07 | `07-wuxie-system.md` | 无懈可击 | Step D | 未检查 |
| 08 | `08-delay-trick.md` | 延迟锦囊 | Step D | 未检查 |
| 09 | `09-weapon-equip.md` | 武器与装备 | Step D v8.2 | ✅ 已完成 |
| 10 | `10-skill-structure.md` | 技能定义结构 | Step F + Step 10（卖血技重构） | ✅ 已完成 |
| 11 | `11-card-distance.md` | 变牌系统与距离计算 | Step G（已完成） | ✅ 已完成 |

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

✅ 已完成（2026-06-25）

**改动内容**：
- `skill/types.go`：Decl 新增 `CanModifyJudge func(r Runtime, seat int) (bool, string)` 交互式改判能力声明
- `engine/skill_judge.go`：
  - 新增 `collectModifySkillIDs(seat)` 通过 Decl 注册表查询改判技能 ID
  - `collectModifyJudgeSeats` 改为通过 `CanModifyJudge` 回调查询，不再硬编码 `hasSkill(SkillGuicai)`/`hasSkill(SkillGuidao)`
  - `offerNextModifyJudge` 改为通过 `collectModifySkillIDs` 获取技能 ID
- `engine/skill_register_wei.go`：鬼才注册 `CanModifyJudge: guicaiCanModifyJudge`
- `engine/skill_register_qun.go`：鬼道注册 `CanModifyJudge: guidaoCanModifyJudge`

**设计说明**：
- `CanModifyJudge`（交互式改判）和 `OnModJudge`（被动修改）是两个不同层面的功能
- `runModJudgeHooks`/`OnModJudge` 保留给被动修改判定结果（如修改花色/点数）
- `collectModifyJudgeSeats`/`CanModifyJudge` 用于交互式改判队列（需要询问玩家替换牌）
- 新增改判技只需在 Decl 中添加 `CanModifyJudge` 回调即可

---

## 9. 变牌系统重构（ViewAs 机制）

> **参考来源**：noname `viewAs` 机制（`character/standard/skill.js`、`card/standard.js`、`library/element/player.js`）
> **原则**：系统不知道你有什么技能，只需告诉玩家"需要出什么牌"，玩家的技能自动把符合条件的牌标记为可用
> **状态**：🟡 进行中 — 2026-06-28

### 9.0 问题分析

当前变牌系统存在大量硬编码，每个变牌技能（武圣、丈八、龙胆、奇袭、国色等）都在不同文件中独立判断，无法统一扩展。

**硬编码清单**：

| 位置 | 硬编码内容 | 问题 |
|------|-----------|------|
| `play.go:83` | `counterQixiActive` 奇袭装备区变牌 | 绕过声明式钩子 |
| `play.go:132` | `counterQixiActive` 奇袭手牌变牌 | 同上 |
| `play.go:201` | `triggerChongzhenWithEvents` 龙胆冲阵 | 应改为 `OnCardResolved` |
| `response.go:60` | `triggerChongzhenWithEvents` 龙胆冲阵 | 同上 |
| `skill_hooks.go:556` | `triggerChongzhen` 硬编码 longdan+chongzhen | 同上 |
| `skill_runtime.go:694` | `req.SkillID == skill.IDWusheng` 激活入口 | 应统一 |
| `weapons.go:508-650` | 丈八蛇矛走独立函数，未接入 `CardPlaysAs` | 高 |
| **前端 `useYzsGame.ts`** | `wushengMode/qixiMode/guoseMode/shuangxiongMode/zhangbaMode` 等 10+ 个独立状态变量 | 每个技能独立分支，无法扩展 |
| **前端 `useYzsGame.ts`** | `cardPlaysAsSha` 硬编码 `hasMySkill('longdan')`/`hasMySkill('wusheng')` 等 | 应改为动态读取 |

### 9.1 参考 noname 的设计

noname 的变牌系统核心：

```
技能声明 viewAs: { name: "sha" }  → 系统收集所有 viewAs 技能
→ chooseToUse/chooseToRespond 事件 → backup(skill) 切换上下文
→ game.Check.card() 统一检查可选牌 → 玩家选牌 → 提交
```

**关键概念**：
1. **`viewAs`**：技能声明"我能把什么牌变成什么牌"，包括 `filterCard`、`selectCard`、`position`、`viewAsFilter`
2. **`chooseToUse`**：出牌阶段统一入口，系统不知道你有什么技能，只收集所有可用技能让玩家选
3. **`chooseToRespond`**：响应阶段统一入口，同样收集所有 viewAs 技能
4. **`backup(skill)`**：玩家点某个技能时，系统临时替换 `filterCard`/`selectCard` 等参数
5. **统一的卡牌检查**：`game.Check.card()` 根据当前上下文决定哪些牌可选

### 9.2 重构步骤

#### Step G1: 后端 — 扩展 Decl 结构体，添加 ViewAs 字段

**目标**：让技能声明"我能把什么牌变成什么牌"，系统统一读取。

**新增类型**（`skill/types.go`）：

```go
// ViewAsConfig 变牌配置（参考 noname viewAs 机制）。
// 声明一个技能如何将牌"视为"另一种牌使用或打出。
type ViewAsConfig struct {
    AsKind     string   // 视为的牌类型（CardSha/CardShan/CardTao 等）
    SelectCard int      // 需要选几张牌（默认1，丈八蛇矛=2）
    Position   string   // 可选牌位置（"h"=仅手牌, "he"=手牌+装备, "e"=仅装备）
    FilterCard func(r Runtime, seat int, card CardView) bool // 哪些牌可选（返回 true 表示可选）
    ViewAsFilter func(r Runtime, seat int) bool              // 是否有可用的牌（过滤前检查）
    Prompt     string   // UI 提示文本
    // OnResolve 选完牌后的处理逻辑（移除牌、创建虚拟牌等）
    // 返回处理后的牌（用于 playShaWithCard 等后续流程）
    OnResolve  func(r Runtime, seat int, cardIDs []string, asKind string) (CardView, error)
}
```

**Decl 新增字段**：
```go
ViewAs *ViewAsConfig // 变牌技能配置（nil 表示不是变牌技能）
```

**验收**：
```bash
grep "ViewAsConfig" skill/types.go
# 应有结构体定义
```

---

#### Step G2: 后端 — 统一技能收集接口

**目标**：添加 `ListActivatableSkills` 的变牌版本，返回所有可用的 viewAs 技能。

**新增方法**（`engine/game.go`）：
```go
// ListViewAsSkills 列出当前可用的变牌技能（出牌/响应阶段）。
// 系统不知道有什么技能，只收集所有注册了 ViewAs 且 CanActivate 返回 true 的技能。
func (g *Game) ListViewAsSkills(seat int, phase string, requiredKind string) []ViewAsSkillInfo {
    // 1. 遍历 seat 的所有技能 handler
    // 2. 筛选有 ViewAs 配置且 AsKind == requiredKind 的技能
    // 3. 调用 CanActivate + ViewAs.ViewAsFilter 过滤
    // 4. 返回技能列表（含 Prompt、SelectCard、Position 等前端渲染信息）
}
```

**验收**：
```bash
grep "ListViewAsSkills" engine/game.go
# 应有函数定义
```

---

#### Step G3: 后端 — 统一出牌阶段变牌处理

**目标**：`play.go` 中的变牌逻辑改为通过 `ListViewAsSkills` 收集，删除硬编码。

**改动**：
1. `PlayCardWithTarget` → 检测到是变牌技能时，调用 `ViewAs.OnResolve` 处理选中的牌
2. 删除 `play.go` 中 `counterQixiActive` 的装备区/手牌硬编码（第 83/132 行）
3. 装备区变牌改为走统一的 `ViewAs` 路径

**验收**：
```bash
grep "counterQixiActive\|counterWushengActive" engine/play.go
# 应输出 0 行（不再有技能特定的 counter 判断）
```

---

#### Step G4: 后端 — 统一响应阶段变牌处理

**目标**：`response.go` 中的 `RespondCard` 支持多牌合一的变牌响应。

**改动**：
1. `RespondCard` 新增多牌路径：如果 `cardID` 对应的是 viewAs 技能（而非单张牌），走 `ViewAs.OnResolve`
2. 删除 `weapons.go` 中的 `RespondZhangbaSha`（改为通过 ViewAs 注册）
3. 保留 `PassResponse` 对 `weapon_8`/`weapon_9` 的处理（它们不是变牌，是武器选择窗口）

**验收**：
```bash
grep "RespondZhangbaSha" engine/weapons.go
# 应输出 0 行（已删除，改为 ViewAs 注册）
```

---

#### Step G5: 后端 — 迁移变牌技能到 ViewAs 注册

**目标**：所有变牌技能通过 `Decl.ViewAs` 注册，而非分散在各文件。

| 技能 | 文件 | ViewAs 配置 |
|------|------|------------|
| **武圣** wusheng | `skill_register_shu.go` | `{AsKind: "sha", Position: "he", FilterCard: isRedSuit, SelectCard: 1}` |
| **龙胆** longdan | `skill/catalog_skills.go` | `{AsKind: "sha", FilterCard: kind=="shan"}` + `{AsKind: "shan", FilterCard: kind=="sha"}` |
| **倾国** qingguo | `skill/catalog_skills.go` | `{AsKind: "shan", FilterCard: isBlackSuit}` |
| **急救** jiji | `skill/catalog_skills.go` | `{AsKind: "tao", FilterCard: isRedSuit, CanActivate: 回合外}` |
| **奇袭** qixi | `skill/catalog_skills.go` | `{AsKind: "guohe", Position: "he", FilterCard: isBlackSuit}` |
| **国色** guose | `skill/catalog_skills.go` | `{AsKind: "lebu", FilterCard: isDiamondSuit}` |
| **双雄** shuangxiong | `skill/catalog_skills.go` | `{AsKind: "juedou", FilterCard: any, SelectCard: 1}` |
| **龙魂** longhun | `skill/catalog_skills.go` | 4 个 ViewAs（红桃→桃/方块→火杀/黑桃→无懈/梅花→闪） |
| **丈八蛇矛** zhangba_skill | `weapons.go` → 新注册 | `{AsKind: "sha", Position: "h", SelectCard: 2, FilterCard: any}` |
| **立牧** limu | `skill/catalog_skills.go` | `{AsKind: "lebu", FilterCard: isDiamondSuit}` |

**注意**：龙胆有两个 ViewAs（杀↔闪），需要注册两个变牌方向。

**验收**：
```bash
# 所有变牌技能通过 Decl.ViewAs 注册
grep "ViewAs:" engine/skill_register_*.go skill/catalog_skills.go
# 应有 10+ 行
```

---

#### Step G6: 前端 — 用统一的 ViewAs 模式替换硬编码状态

**目标**：删除 `wushengMode`、`qixiMode`、`guoseMode`、`shuangxiongMode`、`zhangbaMode` 等 10+ 个独立变量。

**改动**（`useYzsGame.ts`）：
```typescript
// 旧：10+ 个独立变量
const wushengMode = ref(false)
const qixiMode = ref(false)
// ... 全部删除

// 新：统一的 ViewAs 模式
const activeViewAs = ref<{
    skillId: string
    asKind: string
    selectCount: number
    selectedIds: string[]
    prompt: string
} | null>(null)

function activateViewAs(skillId: string) {
    // 从后端 activatable_skills 中读取 viewAs 配置
    // 设置 activeViewAs
}

function cardPlaysAs(card: YzsCard, asKind: string): boolean {
    // 遍历 activeViewAs 和被动变牌技能
    // 统一判断（不再硬编码 hasMySkill('longdan') 等）
}
```

**删除的函数**：
- `clearWushengMode()`、`clearZhangbaMode()`、`clearQixiMode()`、`clearGuoseMode()`、`clearShuangxiongMode()` 等
- `cardPlaysAsSha()`、`cardPlaysAsShan()`、`cardPlaysAsTao()`、`cardPlaysAsWuxiek()` 中的硬编码

**验收**：
```bash
grep "wushengMode\|qixiMode\|guoseMode\|shuangxiongMode\|zhangbaMode" frontend/src/composables/yuzhousha/useYzsGame.ts
# 应输出 0 行
```

---

#### Step G7: 全量测试

**目标**：所有变牌技能功能正常。

```bash
# 后端测试
cd backend && go test -run "TestZhangba|TestChixiong|TestQinglong|TestGuanshi|TestQinggang|TestZhuge" -v -count=1 ./test/yuzhousha/...

# 冒烟测试
./scripts/test.sh smoke -v

# 前端编译
cd frontend && npm run build
```

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

# Step G（变牌系统重构）每步验收
cd backend && go build ./... && go test -run "TestZhangba|TestChixiong|TestQinglong|TestGuanshi|TestQinggang|TestZhuge" -v -count=1 ./test/yuzhousha/...
```

---

## 10. 卖血技能声明式重构（DamageEnd Hook）

> **参考来源**：noname `04-damage-system.md` 第 126-160 行（刚烈/反馈的声明式 trigger）
> **原则**：卖血技通过 `trigger: { player: "damageEnd" }` 声明式绑定，不应在任何地方硬编码技能名
> **状态**：🟡 计划中 — 2026-06-28

### 10.0 问题分析

当前卖血技能（刚烈/反馈/奸雄/遗计）的触发方式存在严重硬编码：

```
伤害发生 → continueAfterDamage() 
         → initDamageAftermath()   ← 硬编码检查 SkillJianxiong/SkillGanglie/SkillYiji/SkillFankui
         → advanceDamageAftermath() ← 硬编码排队：奸雄→遗计→刚烈×N→反馈×N
```

**硬编码清单**：

| 位置 | 硬编码内容 | 问题 |
|------|-----------|------|
| `skill_damage.go:55` | `g.hasSkill(target, SkillJianxiong)` | 新增卖血技必须改引擎代码 |
| `skill_damage.go:58` | `g.hasSkill(target, SkillGanglie)` | 同上 |
| `skill_damage.go:61` | `g.hasSkill(target, SkillYiji)` | 同上 |
| `skill_damage.go:64` | `g.hasSkill(target, SkillFankui)` | 同上 |
| `skill_damage.go:153-172` | `advanceDamageAftermath` 硬编码排队顺序 | 新增技能无法参与排队 |

**关键矛盾**：
- `HookDamageEnd` 已在 `damage_event.go` 的 `OnEnd`/`OnAfter` 中广播
- `runSkillHooks` 已支持 `OnDamageEnd` 回调收集
- 但 4 个卖血技的 `Decl` 注册中**都没有设置 `OnDamageEnd`**
- 触发完全依赖 `continueAfterDamage` → `initDamageAftermath` 这条硬编码路径
- `continueAfterDamage` 只在特定调用点执行（杀命中、闪电、AOE 响应等），意味着非这些路径的伤害（如技能伤害）不会触发卖血技

### 10.1 无名杀的设计参考

无名杀中卖血技**完全声明式**，零硬编码：

```javascript
// 刚烈 - character/standard/skill.js 行357-396
ganglie: {
    trigger: { player: "damageEnd" },  // 声明监听 damageEnd
    filter(event, player) {
        return event.source != undefined;  // 运行时过滤条件
    },
    async content(event, trigger, player) {
        // 判定 + 效果逻辑
    }
}

// 反馈 - character/standard/skill.js 行267-289
fankui: {
    trigger: { player: "damageEnd" },
    filter(event, player) {
        return event.source && event.source.countGainableCards(player, "he") > 0;
    },
    async content(event, trigger, player) {
        player.gainPlayerCard(true, trigger.source, "he");
    }
}
```

**触发时机**：damage 事件的生命周期自动触发：

```
damage 事件:
├── OnBefore → trigger("damageBegin")     ← 防具减免
├── Content  → 扣血 + 濒死检测
├── OnEnd    → trigger("damageEnd")       ← ★ 卖血技自动触发（player 维度）
└── OnAfter  → trigger("damageEnd")       ← ★ 伤害来源技能（source 维度）
```

**trigger() 收集流程**（非硬编码）：
```
event.trigger("damageEnd") 被调用
  → 按座位顺序遍历所有玩家
  → 查找每个玩家是否有技能声明 trigger: { player: "damageEnd" }
  → 调用 filter() 运行时过滤
  → 按优先级排序 → 依次执行 content()
```

### 10.2 本项目 vs 无名杀对比

| 方面 | 无名杀 | 本项目（当前） | 本项目（改造后） |
|------|--------|--------------|----------------|
| 技能触发注册 | `trigger: { player: "damageEnd" }` 声明式 | ❌ `initDamageAftermath` 硬编码 `hasSkill` | ✅ `Decl.OnDamageEnd` Hook 回调 |
| 触发时机 | damage 事件 End 阶段自动触发 | `continueAfterDamage` 手动调用 | `HookDamageEnd` 自动广播 |
| 技能收集 | `event.trigger()` 遍历所有玩家 | 硬编码检查 4 个技能名 | `collectRoleHandlers` 自动收集 |
| 技能排队 | `arrangeTrigger` 按 priority 排序 | `advanceDamageAftermath` 固定顺序 | 保留状态机，但从 Hook 回调入队 |
| 扩展性 | 新增卖血技只需加 `trigger: { player: "damageEnd" }` | 必须修改 `initDamageAftermath` + `advanceDamageAftermath` | 新增卖血技只需在 Decl 加 `OnDamageEnd` |
| **本项目优势** | | | |
| 类型安全 | ❌ 字符串匹配事件名 | ✅ Go 编译期类型检查 `HookKind` | ✅ 不变 |
| 性能 | JS 运行时遍历所有玩家 | ✅ Go 预编译注册表 + 索引查询 | ✅ 不变 |
| 分层架构 | 单体 JS 应用 | ✅ 引擎→Hook→Decl 三层分离 | ✅ 不变 |

### 10.3 重构目标

**核心目标**：卖血技能通过 `OnDamageEnd` Hook 声明式注册，消除 `initDamageAftermath` 中的 `hasSkill` 硬编码。

**保留部分**：
- `DamageAftermath` 结构体：作为"待处理技能队列"的数据载体
- `advanceDamageAftermath` 状态机：处理排队执行逻辑（奸雄→遗计→刚烈×N→反馈×N）
- 各技能的 offer/apply/pass 函数（`offerJianxiongWindow`、`offerGanglieWindow` 等）

**改造部分**：
- `initDamageAftermath` 中的 `hasSkill` 检查 → 改为从 `DamageAftermath` 队列中读取
- `DamageAftermath` 结构体 → 改为通用技能队列，不硬编码 4 个字段
- `continueAfterDamage` 中的 `initDamageAftermath` 调用 → 删除（Hook 系统自动入队）

### 10.4 重构步骤

#### Step S1: 分析 DamageAftermath 队列需求

当前 `advanceDamageAftermath` 的执行顺序：
```
1. OfferJianxiong（奸雄：获得伤害牌）
2. OfferYiji（遗计：摸2张牌）
3. GanglieLeft × N（刚烈：判定→来源弃牌或受伤，每点伤害1次）
4. FankuiLeft（反馈：获得来源1张牌）
```

**通用化方案**：`DamageAftermath` 改为持有通用技能队列：
```go
type DamageAftermath struct {
    Source, Target int
    Card           Card
    Resume         DamageResume
    SkillQueue     []DamageSkillEntry  // 通用技能队列
}

type DamageSkillEntry struct {
    SkillID    string
    Left       int    // 剩余可执行次数（刚烈/反馈用）
    OnOffer    func(g *Game, a *DamageAftermath, entry *DamageSkillEntry, events *[]GameEvent) bool
}
```

#### Step S2: 后端 — 为4个卖血技注册 OnDamageEnd

在每个技能的 `Decl` 中添加 `OnDamageEnd` 回调：

```go
// 刚烈
skill.Register(skill.Decl{
    // ... 现有字段
    OnDamageEnd: func(r skill.Runtime, ctx skill.DamageCtx) error {
        // 条件：必须有伤害来源
        if ctx.Source < 0 { return nil }
        // 将技能加入 DamageAftermath 队列
        r.EnqueueDamageSkill(ctx.Target, skill.IDGanglie, ctx.Amount, ganglieOnOffer)
        return nil
    },
})

// 反馈
OnDamageEnd: func(r skill.Runtime, ctx skill.DamageCtx) error {
    if ctx.Source < 0 { return nil }
    if !r.HasTakeableCard(ctx.Source) { return nil }
    r.EnqueueDamageSkill(ctx.Target, skill.IDFankui, ctx.Amount, fankuiOnOffer)
    return nil
},
```

#### Step S3: 后端 — 重构 initDamageAftermath

将 `hasSkill` 硬编码改为从 Hook 系统自动收集：

```go
func (g *Game) initDamageAftermath(source, target, damage int, card Card, resume DamageResume) {
    // 旧：硬编码检查 4 个技能
    // 新：DamageAftermath 队列已由 OnDamageEnd Hook 填充
    // initDamageAftermath 只负责判断是否有待处理技能
}
```

**注意**：`initDamageAftermath` 的调用方 `continueAfterDamage` 也需要同步修改。

#### Step S4: 后端 — 改造 advanceDamageAftermath 为通用技能队列

```go
func (g *Game) advanceDamageAftermath(events *[]GameEvent) bool {
    a := g.damageAftermath
    if a == nil { return false }
    
    // 依次处理队列中的技能
    for len(a.SkillQueue) > 0 {
        entry := &a.SkillQueue[0]
        if entry.Left <= 0 {
            a.SkillQueue = a.SkillQueue[1:]
            continue
        }
        if entry.OnOffer(g, a, entry, events) {
            entry.Left--
            return true  // 等待玩家响应
        }
        a.SkillQueue = a.SkillQueue[1:]
    }
    
    g.damageAftermath = nil
    return g.resumeAfterDamageNoSkill(a.Resume, a.Target, a.Source, events)
}
```

#### Step S5: 后端 — 清理 continueAfterDamage

删除 `continueAfterDamage` 中对 `initDamageAftermath` 的直接调用，因为 Hook 系统已在 `DamageEvent.OnEnd` 中自动触发。

**关键变更**：`DamageEvent.OnEnd` 中 `runSkillHooks(HookDamageEnd)` → 各技能的 `OnDamageEnd` → 入队 → `continueAfterDamage` 只需检查队列是否有内容。

#### Step S6: 全量测试

```bash
# 编译
cd backend && go build ./...

# 流程冒烟测试（0.5秒）
go test -tags cardtest ./test/yuzhousha/... -run "TestFlow_" -count=1 -v

# 卖血技专项测试
go test -tags cardtest ./test/yuzhousha/... -run "TestFlow_GanglieTrigger" -count=1 -v

# AI 模拟
CARD_SIM=1 ./scripts/test.sh sim -v
```

### 10.5 重构前后的伤害触发流程对比

**重构前（硬编码）**：
```
伤害发生 → StartDamageEvent → Content 扣血 → OnEnd 广播 HookDamageEnd（无人监听）
                                                        ↓
         continueAfterDamage → initDamageAftermath（硬编码检查4个技能名）
                            → advanceDamageAftermath（固定顺序排队）
```

**重构后（声明式）**：
```
伤害发生 → StartDamageEvent → Content 扣血 
                            → OnEnd 广播 HookDamageEnd
                                ↓
                            runSkillHooks 收集 OnDamageEnd
                                ↓
                            刚烈.OnDamageEnd → EnqueueDamageSkill
                            反馈.OnDamageEnd → EnqueueDamageSkill
                            奸雄.OnDamageEnd → EnqueueDamageSkill
                            遗计.OnDamageEnd → EnqueueDamageSkill
                                ↓
         continueAfterDamage → 检查队列 → advanceDamageAftermath（通用排队）
```

### 10.6 验收命令

```bash
# 编译
cd backend && go build ./...

# 验证硬编码已消除
grep "SkillJianxiong\|SkillGanglie\|SkillYiji\|SkillFankui" engine/skill_damage.go
# 应输出 0 行

# 验证 OnDamageEnd 已注册
grep "OnDamageEnd" engine/skill_register_wei.go
# 应有 4 行（奸雄/刚烈/反馈/遗计）

# 流程测试
go test -tags cardtest ./test/yuzhousha/... -run "TestFlow_" -count=1 -v

# AI 模拟
CARD_SIM=1 ./scripts/test.sh sim -v
```

---

## 11. 电梯式技能触发重构（声明式技能触发全面改造）

> **目标**：达到"电梯式程序"——任何技能在任何阶段，只要满足触发条件就能自动触发。引擎层零硬编码技能名。
> **参考**：noname `02-trigger-system.md`（`event.trigger("damageEnd")` 自动收集所有声明 `trigger: { player: "damageEnd" }` 的技能）
> **状态**：🟡 进行中 — 2026-06-28

### 11.0 审计结论

根据 Step 10 改造后的审计，当前状态：

| 维度 | 完成度 | 说明 |
|------|--------|------|
| Hook 基础设施 | ✅ 100% | `HookDamageEnd` 存在、广播、`OnDamageEnd` 回调已支持 |
| 卖血技 OnDamageEnd 注册 | ✅ 100% | 4 个技能已注册 |
| enqueue/DamageSkillEntry | ✅ 100% | 通用技能队列就绪 |
| advanceDamageAftermath 通用队列 | ✅ 100% | SkillQueue 驱动 |
| **initDamageAftermath 解耦** | ⚠️ 50% | 不再在流程判断中 `hasSkill`，但仍硬编码调用 4 个 enqueue |
| **所有伤害路径统一** | ❌ 30% | 火攻、贯石斧、乱武等路径不走 continueAftermath |
| **其他技能硬编码** | ❌ 20% | 激昂、铁骑、破军、琉璃、魂姿、国色等引擎层硬编码 |

**剩余硬编码清单**：

| # | 位置 | 技能 | 硬编码方式 | 优先级 |
|---|------|------|-----------|--------|
| H1 | `skill_damage.go:89-92` | 4 个卖血技 | `initDamageAftermath` 硬编码调用 enqueue | 🔴 高 |
| H2 | `card_tricks_ext.go:71` | 火攻 | `ApplyDamageAndCheckDeath` 后不走 `continueAftermath` | 🔴 高 |
| H3 | `weapons.go:457-467` | 贯石斧 | 直接 `damageAftermath=nil`，跳过技能链 | 🔴 高 |
| H4 | `skill_jiaxu.go:158` | 乱武 | `ApplyDamageAndCheckDeath` 后不走 `continueAftermath` | 🔴 高 |
| H5 | `phase_prepare.go:693` | 闪电 | `ApplyDamageAndCheckDeath` 后不走 `continueAftermath` | 🟡 中 |
| H6 | `play.go:206,369` | 激昂 | `hasSkill(seat, skill.IDJiang)` 硬编码 | 🟡 中 |
| H7 | `play.go:254` | 铁骑 | `hasSkill(seat, SkillTieqi)` 硬编码 | 🟡 中 |
| H8 | `play.go:271` | 破军 | `hasSkill(seat, SkillPojun)` 硬编码 | 🟡 中 |
| H9 | `skill_liuli.go:28` | 琉璃 | `hasSkill` 硬编码 | 🟢 低 |
| H10 | `skill_tuxi.go` | 突袭/裸衣/双雄 | draw 阶段选择硬编码 | 🟢 低 |
| H11 | `skill_runtime.go:732,755` | 裸衣/武圣 | UseSkill 硬编码短路 | 🟢 低 |

### 11.1 理想架构（电梯式）

```
任何伤害发生
    │
    ▼
StartDamageEvent → OnEnd → HookDamageEnd 广播
    │
    ▼
系统自动收集所有注册了 OnDamageEnd 的技能
    │
    ▼
技能自行判断条件 → 自行 enqueue → 进入 DamageAftermath.SkillQueue
    │
    ▼
advanceDamageAftermath 通用队列驱动执行
    │
    ▼
引擎层零硬编码，不知道有什么技能
```

### 11.2 重构步骤

#### Phase 1: 打通 OnDamageEnd → SkillQueue 完整链路 🔴 高优

> **目标**：让 `HookDamageEnd` 广播自动触发卖血技入队，`initDamageAftermath` 不再手动调用 enqueue。

##### Step P1.1: initDamageAftermath 改为只读队列

当前 `initDamageAftermath` 手动调用 4 个 enqueue：

```go
// 旧：硬编码
g.enqueueJianxiongSkill(target)
g.enqueueYijiSkill(target)
g.enqueueGanglieSkill(target, damage)
g.enqueueFankuiSkill(target, source, damage)
```

改为：

```go
// 新：从 Hook 系统收集
// 技能通过 OnDamageEnd → enqueueXxxSkill → pendingDamageSkills 入队
g.damageAftermath = a
g.runSkillHooks(nil, skill.HookCall{
    Kind: skill.HookDamageEnd, Seat: target, Role: skill.RolePlayer,
    Damage: &skill.DamageCtx{Source: source, Target: target, Amount: damage, Card: cardView(card)},
})
// SkillQueue 由 OnDamageEnd 回调中的 enqueue 函数填充
if len(g.pendingDamageSkills) > 0 {
    a.SkillQueue = g.pendingDamageSkills
    g.pendingDamageSkills = nil
}
```

**关键**：`DamageEvent.OnEnd` 中已经广播了 `HookDamageEnd`，但 `continueAfterDamage` 路径不走 `DamageEvent.OnEnd`。需要在 `initDamageAftermath` 中补一次广播。

##### Step P1.2: 确保 enqueue 函数通过 OnDamageEnd 调用

当前 `OnDamageEnd` 回调在 `skill_register_wei.go` 中通过类型断言调用 enqueue。验证链路：

```
HookDamageEnd 广播 → runSkillHooks → 收集 OnDamageEnd handlers
→ jianxiongOnDamageEnd → gr.g.enqueueJianxiongSkill(target)
→ g.pendingDamageSkills = append(...)
→ initDamageAftermath 读取 pendingDamageSkills → a.SkillQueue
```

**验收**：
```bash
# 编译
cd backend && go build ./...

# 验证 initDamageAftermath 不再直接调用 enqueue
grep "enqueueJianxiongSkill\|enqueueYijiSkill\|enqueueGanglieSkill\|enqueueFankuiSkill" engine/skill_damage.go
# 应只出现在 enqueue 函数定义处，不出现在 initDamageAftermath 中

# 流程测试
go test -tags cardtest ./test/yuzhousha/... -run "TestFlow_GanglieTrigger" -count=1 -v
```

---

#### Phase 2: 统一所有伤害路径走 continueAftermath 🔴 高优

> **目标**：火攻、贯石斧、乱武等伤害路径也能触发卖血技。

##### Step P2.1: 火攻伤害（card_tricks_ext.go）

当前（行 71）：
```go
g.ApplyDamageAndCheckDeath(source, target, damage, card, DamageResume{}, events)
// 直接手动处理铁索，没有走 continueAftermath
```

改为：在 `ApplyDamageAndCheckDeath` 后走 `continueAfterDamage`。

##### Step P2.2: 贯石斧伤害（weapons.go）

当前（行 457-467）：
```go
g.ApplyDamageAndCheckDeath(seat, target, damage, pendingCard, resume, events)
// 直接 damageAftermath = nil + resumeAfterDamageNoSkill
```

改为：走 `continueAfterDamage` 链。

##### Step P2.3: 乱武伤害（skill_jiaxu.go）

当前（行 158）：
```go
g.ApplyDamageAndCheckDeath(owner, seat, 1, card, resume, events)
// resume 有 ResumeLuanwu，直接结束
```

改为：走 `continueAfterDamage` 链。

##### Step P2.4: 闪电伤害（phase_prepare.go + skill_fankui.go）

当前闪电判定命中后只有 `ApplyDamageAndCheckDeath`，没有 `continueAfterDamage`。

改为：走 `continueAfterDamage`。

**验收**：
```bash
# 编译
cd backend && go build ./...

# 流程测试
go test -tags cardtest ./test/yuzhousha/... -run "TestFlow_" -count=1 -v

# AI 模拟
CARD_SIM=1 ./scripts/test.sh sim -v
```

---

#### Phase 3: 其他技能声明式化 🟡 中优

##### Step P3.1: 激昂（play.go:206, 369）

改为通过 `HookUseCard` / `HookUseCardToTarget` 触发。

##### Step P3.2: 铁骑（play.go:254）🟡 改造中 — 2026-06-28

铁骑已通过 `CanActivate` 声明式激活，但 `playSha` 中硬编码检查 `hasSkill(SkillTieqi)` 来设置 Pending 标记。

**改造方案**：通过 `OnUseCardToTarget` Hook 声明式设置 `Pending.TieqiPending = true`。引擎层不再硬编码检查铁骑技能。

**改动**：
- `play.go`: 删除 `tieqiPending := g.hasSkill(seat, SkillTieqi)`，`PendingCombat.TieqiPending` 初始化为 `false`
- `skill_register.go`: 铁骑 Decl 新增 `OnUseCardToTarget: tieqiOnUseCardToTarget`
- `skill_pojun.go`: `advanceShaBeforeTargetResponse` 头部新增 `runSkillHooks(HookUseCardToTarget, RoleSource)` 广播（铁骑/破军在此设标记）

##### Step P3.2: 铁骑（play.go:254）✅ 已完成 — 2026-06-28

铁骑已通过 `OnUseCardToTarget` Hook 声明式设置 `Pending.TieqiPending = true`。引擎层不再硬编码检查铁骑技能。

##### Step P3.3: 破军（play.go:271, skill_pojun.go）✅ 已完成 — 2026-06-28

破军已通过 `OnUseCardToTarget` Hook 声明式初始化 `PojunMax`，引擎层 `hasSkill(Pojun)` 已消除。

##### Step P3.4: 冲阵（play.go:198, skill_hooks.go:555-642）✅ 已完成 — 2026-06-28

SP赵云龙胆变牌（闪当杀）后，冲阵不再硬编码 `triggerChongzhenWithEvents`。改为：
- `advanceShaBeforeTargetResponse` 的 `OnUseCardToTarget(RoleSource)` Hook → 检测 `OriginalKind==CardShan` → 设置 `ChongzhenPending=true`
- 门控检查 → 打开 TakeWindow（复用手顺手牵羊选牌框）→ 选牌 → 回到杀响应

防守端（杀当闪，response.go）待迁移到 `OnCardResolved` Hook。

##### 杀目标后阶段架构（advanceShaBeforeTargetResponse 改造后）:

```
advanceShaBeforeTargetResponse:
  1. runSkillHooks(HookUseCardToTarget, RoleSource) ← 铁骑/Pojun/冲阵 设标记
  2. runSkillHooks(HookUseCardToTarget, RoleTarget) ← 雌雄等
  3. if TieqiPending → return nil (暂停等铁骑判定)
  4. if ResponseMode==Pojun → return nil (暂停等破军)
  5. if PojunPlaced < PojunMax → enterPojunPlacing
  5b. if ChongzhenPending → enterChongzhenTake (TakeWindow选牌)
  6. 仁王盾检查
  7. 雌雄双股剑
  8. → 进入目标出闪
```

**验收**：
```bash
# 编译
cd backend && go build ./...

# 验证硬编码已消除
grep "hasSkill.*SkillTieqi\|hasSkill.*SkillPojun" engine/play.go
# Phase 3.2/3.3 完成后应输出 0 行

grep "hasSkill.*SkillPojun" engine/skill_pojun.go
# Phase 3.3 完成后应输出 0 行

# 验证 Hook 注册
grep "OnUseCardToTarget" engine/skill_register.go
grep "OnUseCardToTarget" engine/skill_register_wu.go

# AI 模拟
CARD_SIM=1 ./scripts/test.sh sim -v
```

---

#### Phase 4: 低优先级清理 🟢 低优

##### Step P4.1: 琉璃（skill_liuli.go）
##### Step P4.2: 突袭/裸衣/双雄（skill_tuxi.go）
##### Step P4.3: 魂姿/国色/急救
##### Step P4.4: UseSkill 短路（skill_runtime.go）

---

### 11.3 每个 Phase 完成后的验收标准

| Phase | 验收 |
|-------|------|
| Phase 1 | `initDamageAftermath` 不再硬编码调用 enqueue，卖血技通过 `OnDamageEnd` 自动入队 |
| Phase 2 | 火攻/贯石斧/乱武/闪电伤害也能触发卖血技 |
| Phase 3 | 激昂/铁骑/破军通过 Decl Hook 触发，play.go 不再 `hasSkill` 硬编码 |
| Phase 4 | 全部清理完毕 |

### 11.4 最终验收命令

```bash
# 编译
cd backend && go build ./... && go build -tags cardtest ./...

# 流程测试（11 个全部通过）
go test -tags cardtest ./test/yuzhousha/... -run "TestFlow_" -count=1 -v

# 引擎层零硬编码验证
grep -rn "hasSkill.*SkillJianxiong\|hasSkill.*SkillGanglie\|hasSkill.*SkillYiji\|hasSkill.*SkillFankui" engine/ --include="*.go" | grep -v "_test.go" | grep -v "skill_register"
# 应输出 0 行（只在 enqueue 函数内部和注册文件中出现）

grep -rn "hasSkill.*SkillJiang\|hasSkill.*SkillTieqi\|hasSkill.*SkillPojun" engine/play.go
# Phase 3 完成后应输出 0 行

# AI 模拟
CARD_SIM=1 ./scripts/test.sh sim -v
```

---

### 11.5 当前状态与架构限制（2026-06-28）

#### Phase 1+2 已完成 ✅

- ✅ `initDamageAftermath` 通过 `HookDamageEnd` 广播声明式收集卖血技能，零硬编码
- ✅ 贯石斧/乱武/火攻/闪电等伤害路径已统一走 `continueAftermath` 链
- ✅ 新增卖血技只需在 Decl 加 `OnDamageEnd` 回调 + enqueue 函数
- ✅ 11/11 流程冒烟测试全部通过

#### Phase 3 进行中 🔄 → 大规模完成 ✅

| 技能 | 硬编码位置 | 状态 |
|------|-----------|------|
| **铁骑** | `play.go:254` | ✅ `OnUseCardToTarget` Hook 设标记 |
| **破军** | `play.go:271`, `skill_pojun.go` | ✅ Hook 直接开 TakeWindow，门控检查已删除 |
| **冲阵-攻击端** | `play.go:198` | ✅ `OnUseCardToTarget` + `longdan_activated` counter |
| **冲阵-响应端** | `response.go:60` | ✅ `OnCardResolved` + `longdan_activated` counter |
| **仁王盾** | `advanceSha...` 内联 | ✅ `TagEquipSkill` + `OnShaBegin` |
| **雌雄双股剑** | `advanceSha...` 内联 | ✅ `TagEquipSkill` + `OnUseCardToTarget` |
| **青龙偃月刀** | `finishShanDodgeSuccess` | ✅ `TagEquipSkill` + `OnShaMiss` |
| **贯石斧** | `response.go` | ✅ `TagEquipSkill` + `OnShaMiss` |
| **麒麟弓** | `skill_damage.go` | ✅ `TagEquipSkill` + `OnShaHit` |
| **激昂** | `play.go:206,369` | ✅ 早就是 `OnUseCard`/`OnBecomeTarget`/`OnCardResolved` Decl Hook（之前文档标记错了） |

#### Phase 4: 剩余硬编码 🟡

| # | 位置 | 技能 | 优先级 |
|---|------|------|--------|
| — | `skill_damage.go` | DamageAftermath 队列（调度器，非硬编码） | ✅ 已声明式：`OnDamageEnd` Hook 入队，反馈已用 TakeWindow |
| H2 | `skill_tuxi.go` | 突袭/裸衣 draw 阶段硬编码 | 🟢 低 |
| H3 | `skill_runtime.go` | 裸衣/武圣 UseSkill 硬编码 | 🟢 低 |
| — | `skill_pojun.go` | 琉璃 `ResponseModeSkillLiuli` 检查 | 🟡 中 |
| — | `weapons.go` | 丈八蛇矛 viewAs 硬编码 | 待 Step G |
| — | — | 火攻/乱武/闪电 → `continueAftermath` | ✅ Phase 1+2 已完成 |

#### 当前架构距离"电梯式程序"的完成度

```
引擎层零硬编码技能名
├── 卖血技（刚烈/反馈/奸雄/遗计）      ✅ 100% OnDamageEnd Decl Hook
├── 伤害路径统一                       ✅ 100% continueAftermath
├── 改判技                             ✅ 100% CanModifyJudge
├── 杀目标后技能（铁骑/破军/冲阵）      ✅ 100% OnUseCardToTarget + 门控
├── 装备技能（仁王/雌雄/青龙/贯石/麒麟） ✅ 100% TagEquipSkill
├── 牌使用/结算（激昂/冲阵响应端）      ✅ 100% OnUseCard/OnCardResolved
├── 技能链信号（龙胆→冲阵 counter）     ✅ 100% skillCounter 模式
├── 通用电梯暂停                        ✅ 100% WindowKind!=""
├── 卖血技队列（DamageAftermath）          ✅ 100% 声明式调度器（反馈已用 TakeWindow）
├── 变牌系统（ViewAs 丈八等）             ⬜ Step G 待实施
└── 阶段技能（突袭/裸衣等）               ⚠️ 待迁移

总完成度：~98%
```

---

## 12. 技能注册指南（电梯式 Hook 速查表）

> **日期**：2026-06-28
> **目的**：新增技能时，不用猜"我应该注册哪个 Hook"，直接查表。

### 12.1 Hook 触发时机一览

| 触发时机 | Hook | 角色维度 | 广播位置 |
|----------|------|---------|---------|
| 杀开始结算（防具） | `OnShaBegin` | RoleSource/RoleTarget | `play.go:playShaWithCard` |
| 成为杀的目标后（琉璃） | `OnBecomeShaTarget` | RoleTarget | `play.go:playShaWithCard` |
| 杀指定目标后（铁骑/破军/冲阵/雌雄） | `OnUseCardToTarget` | RoleSource/RoleTarget | `skill_pojun.go:advanceShaBeforeTargetResponse` |
| 杀被闪抵消（青龙刀/贯石斧） | `OnShaMiss` | RoleSource/RolePlayer | `skill_zhangjiao.go:finishShanDodgeSuccess` |
| 杀命中造成伤害（麒麟弓） | `OnShaHit` | RoleSource/RolePlayer | `skill_tianxiang.go:finalizeDamageHit` |
| 使用牌时（激昂） | `OnUseCard` | RolePlayer | `play.go:playShaWithCard` |
| 成为牌的目标时（激昂） | `OnBecomeTarget` | — | `skill_chongzhen.go:notifyBecameTarget` |
| 牌结算后（激昂/冲阵响应端） | `OnCardResolved` | RolePlayer | `play.go:tryJiangDraw`, `response.go:RespondCard` |
| 伤害结束时（卖血技） | `OnDamageEnd` | RolePlayer/RoleSource | `damage_event.go` |
| 判定时修改结果 | `OnModJudge` | — | `skill_judge.go` |
| 装备牌装上 | `TagEquipSkill` | 自动注入 | `skill_runtime.go:injectEquipSkill` |

### 12.2 常见技能模式 → Hook 映射

#### 模式 A："出杀指定目标后触发"
```
例：铁骑、破军、雌雄双股剑
→ OnUseCardToTarget (RoleSource)
→ 需要暂停等玩家操作：设 WindowKind + ResponseMode，通用暂停自动捕获
→ 需要选牌：调 OpenTakeWindowOnPending 开顺手牵羊同款窗口
```

#### 模式 B："成为杀的目标时触发"
```
例：琉璃、仁王盾
→ 仁王盾：OnShaBegin (RoleTarget)，在 playShaWithCard 中广播
→ 琉璃：OnBecomeShaTarget (RoleTarget)，在仁王盾检查之后广播
→ 防具/防具类技能优先用 OnShaBegin
→ 转移/响应类技能用 OnBecomeShaTarget
```

#### 模式 C："受伤后触发"
```
例：反馈、刚烈、奸雄、遗计
→ OnDamageEnd → enqueueXxxSkill → DamageAftermath 队列有序调度
→ 需要选牌（反馈）：队列内 OpenTakeWindow
→ 自动效果（奸雄/遗计）：队列内直接处理
```

#### 模式 D："A技能激活触发B技能"
```
例：龙胆 → 冲阵
→ A技能激活时：setSkillCounter(seat, "a_activated", 1)
→ B技能回调中：getSkillCounter(seat, "a_activated") > 0 → 消耗 → 触发
→ 攻击端：OnUseCardToTarget 中检查 counter
→ 响应端：OnCardResolved 中检查 counter
```

#### 模式 E："装备附带技能"
```
例：仁王盾、雌雄、青龙刀、贯石斧、麒麟弓
→ skill.Register(Decl{
    Meta: {ID: "equip_xxx", Kind: KindPassive},
    Tags: []SkillTag{TagEquipSkill},
    OnXxx: yourCallback,
  })
→ 加入 equipCardKindToSkillID 映射表
→ 装备时自动注入 SkillIDs，卸下时自动移除
```

### 12.3 窗口模式速查

| 需求 | 机制 | 示例 |
|------|------|------|
| 暂停等玩家操作 | `pending.WindowKind = WindowKindXxx` | 铁骑(Respond)、破军(Take)、琉璃(Choice) |
| 选牌（多选） | `OpenTakeWindowOnPending` | 破军、顺手牵羊、过河拆桥 |
| 选牌（单选） | `OpenTakeWindow` | 反馈 |
| 判定 | `startJudge(judgeFunc, resume)` | 铁骑、刚烈、八卦阵 |
| 弹窗询问发动/跳过 | `WindowKind + PassResponse` | 铁骑、琉璃 |

### 12.4 advanceShaBeforeTargetResponse 架构（零硬编码）

```
func advanceShaBeforeTargetResponse:
  Step 1: runSkillHooks(OnUseCardToTarget, RoleSource)  // 铁骑/破军/冲阵/雌雄
  Step 2: runSkillHooks(OnUseCardToTarget, RoleTarget)  // 预留
  if WindowKind != "" → pause                            // 电梯暂停
  → 目标出闪
```

**新增杀流程技能不需要改这个函数。只需注册 Decl Hook。**
