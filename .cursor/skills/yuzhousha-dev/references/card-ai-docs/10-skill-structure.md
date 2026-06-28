# 10 - 技能实现完全指南

> **目的**: AI 在实现新技能时必读本文档。基于无名杀规则 + 本项目框架，提供精确的实现模板。
>
> **关键原则**: 
> - 本项目的框架已定，不要自己发明新的 HookKind 或 Runtime 方法
> - 所有技能 ID 必须在 `skill/ids.go` 中定义常量
> - 所有技能必须在 `engine/skill_register_*.go` 的 `init()` 中注册
> - 无名杀的规则逻辑作为参考，但用本项目的 API 实现

---

## 一、项目技能框架速览

### 1.1 文件位置和职责

| 层 | 文件 | 职责 | 是否可修改 |
|---|------|------|-----------|
| ID 定义 | `skill/ids.go` | 定义所有技能 ID 常量 | ✅ 新增技能时追加 |
| 类型定义 | `skill/types.go` | Decl 结构体、Runtime 接口 | ❌ 不要改（框架） |
| Hook 定义 | `skill/hooks.go` | HookKind 枚举、上下文结构体 | ❌ 不要改（除非框架升级） |
| 技能注册 | `engine/skill_register_*.go` | 技能注册 + 4 个入口函数 | ✅ 新增技能在此 |
| 技能逻辑 | `engine/skill_*.go` | 复杂技能的引擎实现 | ✅ 新增技能在此 |
| 前端 | `composables/yuzhousha/pending/` | 响应窗口 UI 交互 | ✅ 需要前端交互时修改 |

### 1.2 技能注册的四入口模式

每个技能注册时必须提供 **4 个函数**：

```go
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID:   skill.IDXxx,        // 在 skill/ids.go 中定义
        Name: "技能名",
        Kind: skill.KindPassive,  // passive / active / lord / awakening / limited
        Desc: "技能描述文字",
    },
    CanActivate: xxxCanActivate,  // 条件判断：现在是否可发动？
    Activate:    xxxActivate,     // 发动逻辑：执行技能效果
    AIPriority:  xxxAIPriority,   // AI 优先级：返回 0 表示不可发动
    AIActivate:  xxxAIActivate,   // AI 发动逻辑：AI 如何选择参数
})
```

---

## 二、五种经典技能模式（对照无名杀 + 本项目实现）

### 模式 1: 卖血技（受伤后触发）

**无名杀参考**: `ganglie` (character/standard/skill.js 行357-396)
**无名杀设计要点**: trigger: { player: "damageEnd" }，filter 检查 source 存在，content 执行判定+效果
**本项目已有实现**: 刚烈、反馈、奸雄、遗计

**实现模板**:

```go
// 1. 在 skill/ids.go 中添加 ID
const IDXxxSkill = "xxx_skill"

// 2. 在 engine/skill_register_wei.go (或其他王国文件) 中注册
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID: skill.IDXxxSkill, Name: "技能名", Kind: skill.KindPassive,
        Desc: "当你受到伤害后，你可以......",
    },
    CanActivate: xxxCanActivate,
    Activate:    xxxActivate,
    AIPriority:  xxxAIPriority,
    AIActivate:  xxxAIActivate,
})

// 3. 实现四个入口函数
func xxxCanActivate(r skill.Runtime, seat int) bool {
    // 使用 Runtime 的 Pending 方法检查当前是否在等待此技能
    return r.PendingXxxFor(seat)  // ← 这个方法需要在 Runtime 接口和引擎中实现
}

func xxxActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
    // 使用 Runtime 的 Activate 方法执行技能效果
    return r.ActivateXxx(seat, req.CardIDs, req.TargetIndex)
}

func xxxAIPriority(r skill.Runtime, seat int) int {
    if xxxCanActivate(r, seat) {
        return 88  // 优先级数字，参考现有技能
    }
    return 0  // 0 表示不可发动
}

func xxxAIActivate(r skill.Runtime, seat int) error {
    // AI 自动决策逻辑
    return xxxActivate(r, seat, skill.ActivateReq{})
}
```

**关键注意**: 卖血技需要在引擎层实现完整的 Pending 窗口流程：
1. 伤害结算后 → 创建 DamageAftermath
2. 检查技能条件 → 创建 PendingCombat (ResponseMode = "skill_xxx")
3. 玩家/AI 选择 → 执行效果 → advanceDamageAftermath

---

### 模式 2: 改判技（全局判定介入）

**无名杀参考**: `guicai` (character/standard/skill.js 行290-356)
**无名杀设计要点**: trigger: { global: "judge" }，替换 judging[0]
**本项目已有实现**: 鬼才、鬼道

**本项目的实现方式**:

改判技**不需要通过 Decl 的 CanActivate/Activate 单独注册**（鬼才/鬼道是硬编码在 `collectModifyJudgeSeats` 中的）。

如果需要新增改判技，修改 `engine/skill_judge.go` 中的 `collectModifyJudgeSeats`：

```go
func (g *Game) collectModifyJudgeSeats(startSeat int) []int {
    var seats []int
    for i := 0; i < len(g.Players); i++ {
        seat := (startSeat + i + 1) % len(g.Players)
        if g.Players[seat].HP <= 0 { continue }
        canModify := false
        // 鬼才：有手牌
        if g.hasSkill(seat, SkillGuicai) && len(g.Players[seat].Hand) > 0 {
            canModify = true
        }
        // 鬼道：有黑色手牌
        if g.hasSkill(seat, SkillGuidao) && g.hasBlackHandCard(seat) {
            canModify = true
        }
        // ★ 新增改判技在此添加
        if g.hasSkill(seat, SkillNewModifySkill) && /* 条件 */ {
            canModify = true
        }
        if canModify { seats = append(seats, seat) }
    }
    return seats
}
```

同时需要在 `skill_judge.go` 中添加对应的 ResponseMode 和 UI 处理。

---

### 模式 3: 主动技（出牌阶段主动使用）

**无名杀参考**: `qingnang` (character/standard/skill.js 行1612)
**无名杀设计要点**: enable: "phaseUse", filterCard/filterTarget, content 执行效果
**本项目已有实现**: 仁德、结姻、反间、国色

**实现模板**:

```go
// 注册（Kind = KindActive）
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID: skill.IDXxx, Name: "技能名", Kind: skill.KindActive,
        Desc: "出牌阶段，你可以......",
    },
    CanActivate: xxxCanActivate,
    Activate:    xxxActivate,
    AIPriority:  xxxAIPriority,
    AIActivate:  xxxAIActivate,
})

func xxxCanActivate(r skill.Runtime, seat int) bool {
    // 检查：是否在出牌阶段、条件是否满足
    if !r.HasSkill(seat, skill.IDXxx) { return false }
    if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay { return false }
    if r.CurrentTurn() != seat { return false }
    // 检查使用次数（如果需要限制）
    if r.SkillCounter(seat, "xxx_used") >= 1 { return false }
    // 检查其他条件...
    return true
}

func xxxActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
    // 执行技能效果
    target := req.TargetIndex
    cardIDs := req.CardIDs
    // ...
    return nil
}
```

**引擎层需要实现**:
- Runtime 接口中添加对应方法（如 `ActivateXxx(seat int, ...) error`）
- 在 `gameSkillRuntime` 中实现该方法
- 如果需要选牌/选目标，添加对应的 ResponseMode 和 PendingCombat

---

### 模式 4: 牌当牌 (viewAs) / 锁定技 (mod)

**无名杀参考**: `wusheng` (character/standard/skill.js 行907) / `qicai` (character/standard/skill.js 行1221)
**无名杀设计要点**: viewAs: { name: "sha" } / mod: { targetInRange }
**本项目已有实现**: 武圣/龙胆 (CardPlaysAs)、咆哮 (UnlimitedSha)、马术 (DistanceDelta)、奇才 (TrickIgnoresDistance)

**纯 Decl Hook 实现，不需要 CanActivate/Activate**:

```go
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID: skill.IDXxx, Name: "技能名", Kind: skill.KindPassive,
        Desc: "锁定技，......",
        Tags: []skill.SkillTag{skill.TagForced},  // ← 锁定技标记
    },
    // 直接使用 Decl 的 Hook 字段，不需要 CanActivate/Activate
    CardPlaysAs: func(r Runtime, seat int, cardKind, asKind, suit string) bool {
        // 牌当牌：如红色牌当杀
        if asKind == "sha" && (suit == "H" || suit == "D") { return true }
        return false
    },
    // 或
    UnlimitedSha: func(r Runtime, seat int) bool {
        return true  // 杀无次数限制
    },
    // 或
    DistanceDelta: func(r Runtime, from, to int) int {
        return -1  // 马术：攻击距离-1
    },
    // 或
    TrickIgnoresDistance: func(r Runtime, seat int, trickKind string) bool {
        return true  // 奇才：锦囊无距离限制
    },
})
```

**可用的 Decl Hook 字段** (不用 Runtime 接口，直接填函数):

| Hook 字段 | 触发时机 | 返回值含义 |
|-----------|---------|-----------|
| `CardPlaysAs` | 检查牌能否当作某牌使用 | bool: 能/不能 |
| `UnlimitedSha` | 检查杀是否有次数限制 | bool: true=无限制 |
| `DistanceDelta` | 计算距离修正 | int: 负数=减少距离 |
| `TrickIgnoresDistance` | 检查锦囊是否无视距离 | bool: true=无视 |
| `BlocksTarget` | 检查是否能成为目标 | bool: true=不能 |
| `BlocksTrickTarget` | 检查锦囊能否指定目标 | bool: true=不能 |
| `DrawCountBonus` | 摸牌阶段额外摸牌 | int: 额外张数 |
| `HandRetainLimit` | 手牌上限 | int: 上限值 |
| `SkipsDiscardPhase` | 是否跳过弃牌阶段 | bool: true=跳过 |
| `EffectiveSuit` | 花色视为 | string: 视为的花色 |
| `DamageAsHPLoss` | 伤害视为体力流失 | bool: true=视为 |
| `BlocksWuxiek` | 阻止无懈可击 | bool: true=阻止 |
| `BlocksPeachUse` | 阻止使用桃 | bool: true=阻止 |

---

### 模式 5: 装备技（equipSkill）

**无名杀参考**: `zhuge_skill` (card/standard.js 行2859)
**无名杀设计要点**: equipSkill: true，装备时 addSkillTrigger，卸下时 removeSkillTrigger
**本项目实现**: 武器技能直接写在 `weapons.go` 中，非 Decl 注册

**本项目武器技能实现方式**:

武器技能不需要通过 Decl 注册，直接在引擎层处理。参考 `weapons.go` 中的现有实现：

```go
// 诸葛连弩：杀无次数限制 → 通过 UnlimitedSha 在 Decl 中注册
// 贯石斧：杀被闪后可弃牌强制命中 → 在 weapons.go 中处理 shaMiss 响应
// 丈八蛇矛：2手牌当杀 → 通过 CardPlaysAs 在 Decl 中注册
```

如果需要新增装备附带技能：
1. 如果是简单被动效果 → 用 Decl Hook 注册，加 `TagEquipSkill` 标记
2. 如果是复杂交互 → 在对应的武器/防具处理函数中实现

---

## 三、实现新技能的完整流程

### Step 1: 确定技能模式

对照上面五种模式，确定你的技能属于哪种。

### Step 2: 添加技能 ID

在 `skill/ids.go` 中添加常量：
```go
const IDNewSkill = "new_skill"
```

### Step 3: 注册技能

在 `engine/skill_register_*.go` 的 `init()` 中注册：
```go
skill.Register(skill.Decl{...})
```

### Step 4: 实现引擎逻辑（如果需要）

如果技能需要 PendingCombat 窗口（如卖血技、主动选牌技），需要在引擎层：
1. 定义新的 ResponseMode 常量
2. 创建 PendingCombat 窗口
3. 实现效果处理函数
4. 在 Runtime 接口中添加方法
5. 在 gameSkillRuntime 中实现方法

### Step 5: 前端 UI（如果需要）

### Step 6: ★ 编写行为验证测试（必须！）

**这是最容易被跳过但最关键的一步。编译通过 ≠ 功能正确。**
**冒烟测试只验证不崩溃，AI 模拟只验证不卡死，都不验证效果是否正确。**

测试模板参考 `backend/test/yuzhousha/skill_test.go` 和 `scenario_test.go`。

**每种模式必须验证的场景**：

**卖血技（刚烈/反馈/奸雄/遗计）：**
- 正常受伤后触发（创建 PendingCombat）
- 判定结果正确（刚烈红桃无效、其他生效）
- 来源死亡时不触发
- 多技能同时触发时顺序正确

**主动技（青囊/仁德/结姻）：**
- 技能在正确时机可发动
- 发动后效果正确（回血/给牌/摸牌）
- 条件不满足时不可发动（如目标满血→青囊不可用）
- 使用次数限制正确

**牌当牌（武圣/龙胆/倾国）：**
- 符合条件的牌能当目标牌使用
- 不符合条件的牌不能当目标牌使用
- 发动后实际出牌流程正常

**锁定技（马术/奇才/空城/克己）：**
- 距离修正计算正确
- 目标封锁生效/不生效边界
- 手牌上限计算正确

**锦囊牌：**
- `useCard` switch 中有入口（grep 验证）
- `playTrickWithCard` 中有路由
- 无懈窗口正常弹出
- 效果正确执行
- 前端 handler 已注册（grep 验证）

**装备牌（武器/防具/马）：**
- 装备后技能生效
- 卸下后技能移除
- 同名装备不重复注册

### Step 5: 前端 UI（如果需要）

如果技能需要玩家选择/确认，在 `frontend/src/composables/yuzhousha/pending/handlers.ts` 中添加对应的 response_mode 处理器。

---

## 四、避免常见 Bug 的检查清单

### 4.1 状态污染

```go
// ❌ 错误：忘记清理 Pending
g.Pending = &PendingCombat{...}
// 技能结束后必须 g.Pending = nil

// ✅ 正确：确保所有路径都清理
if condition {
    g.Pending = nil  // ← 每个分支都要清理
    return xxx
}
g.Pending = nil
```

### 4.2 座位号越界

```go
// ❌ 错误：没有检查座位号
target := g.Pending.TargetIndex
p := &g.Players[target]  // 可能越界

// ✅ 正确：检查有效性
if target < 0 || target >= len(g.Players) {
    return ErrInvalidTarget
}
```

### 4.3 伤害来源死亡

无名杀的刚烈在 source 死亡后不应触发：

```go
// 检查伤害来源是否还存活
source := a.Source
if g.Players[source].HP <= 0 {
    // 来源已死亡，跳过
    g.advanceDamageAftermath(events)
    return nil
}
```

### 4.4 多技能同时触发顺序

参考无名杀的优先级系统：
- `FirstDo` 的技能最先执行（如无懈可击的 _wuxie）
- 同优先级按注册顺序
- `LastDo` 的技能最后执行

### 4.5 牌的来源检查

```go
// 获取牌时必须检查是否存在
idx, card, ok := g.findCard(seat, cardID)
if !ok {
    return ErrInvalidCard
}
```

### 4.6 改判与判定联动

改判技修改判定牌后，必须确保后续的 judge 函数使用的是**修改后的牌**，而不是原来的牌。

---

## 五、技能实现对照表

| 要实现的功能 | 无名杀做法 | 本项目做法 | 参考文件 |
|-------------|-----------|-----------|---------|
| 卖血摸牌（遗计/奸雄） | trigger: {player:"damageEnd"} | CanActivate 检查 Pending，Activate 执行效果 | skill_register_wei.go |
| 卖血反伤（刚烈） | trigger + judge + damage | PendingCombat 多阶段窗口 | skill_ganglie.go |
| 受伤拿牌（反馈） | trigger + gainPlayerCard | TakeWindow + FankuiTakeFrom | skill_fankui.go |
| 改判（鬼才） | trigger: {global:"judge"} + 替换 judging[0] | collectModifyJudgeSeats 硬编码 | skill_judge.go |
| 牌当杀（武圣） | enable + viewAs: {name:"sha"} | Decl.CardPlaysAs Hook | catalog_skills.go |
| 无限出杀（咆哮） | mod: {cardUsable: Infinity} | Decl.UnlimitedSha Hook | catalog_skills.go |
| 距离修正（马术） | mod: {globalFrom: -1} | Decl.DistanceDelta Hook | catalog_skills.go |
| 锦囊无视距离（奇才） | mod: {targetInRange} | Decl.TrickIgnoresDistance Hook | catalog_skills.go |
| 空城（不能成为目标） | mod: {targetInRange} | Decl.BlocksTarget Hook | catalog_skills.go |
| 额外摸牌（英姿） | mod: {phaseDraw} | Decl.DrawCountBonus Hook | catalog_skills.go |
| 跳过弃牌（克己） | mod + skip("phaseDiscard") | Decl.SkipsDiscardPhase Hook | catalog_skills.go |
| 准备阶段主动技（洛神） | trigger: {player:"phaseZhunbei"} | PreparePhase.Offer + CanActivate | skill_register_wei.go |
| 摸牌阶段主动技（裸衣/突袭） | trigger: {player:"phaseDrawBegin"} | CanActivate 检查 PendingDrawPhaseChoiceFor | skill_register_wei.go |
| 回合内增伤（裸衣） | mod: {damageBonus} | 引擎层 damageBonus 标记 + 伤害计算时检查 | skill_register_wei.go |
| 武器杀无次数限制（诸葛连弩） | equipSkill + mod.cardUsable | Decl.UnlimitedSha + TagEquipSkill | catalog_skills.go |
