# 血量变化时机实现文档

## 概述

本文档描述了宇宙杀游戏引擎中血量变化时机的实现，包括：
- 伤害结算后时机
- 血量流失时机（非伤害扣血）
- 血量变化统一处理

## 设计目标

1. **清晰的时机区分**：区分伤害、血量流失、血量回复三种情况
2. **统一的钩子系统**：所有血量变化都触发统一的 `OnHPChanged` 钩子
3. **技能触发支持**：为【遗计】、【刚烈】等技能提供清晰的触发时机
4. **向后兼容**：保留原有的 `applyDamage` 函数，新增带钩子的版本

## 实现架构

### 1. 技能钩子定义（skill 包）

#### 新增钩子类型

```go
// skill/hooks.go

// HPLostCtx 血量流失后广播（非伤害导致的扣血，如【蛊惑】、【刚烈】等）
type HPLostCtx struct {
    Seat   int
    Amount int
    Reason string // skill | card_effect
    Source int    // 伤害来源（若有）
}

// HPChangedCtx 血量变化后广播（伤害/流失/回复）
type HPChangedCtx struct {
    Seat    int
    OldHP   int
    NewHP   int
    Delta   int    // 变化量：正数=回复，负数=扣血
    Reason  string // damage | hp_loss | heal | skill
    Source  int    // 来源（若有）
    SkillID string // 技能ID（若是技能导致）
}
```

#### 新增 HookKind 常量

```go
// skill/hooks.go
const (
    HookHPLost    HookKind = "hp_lost"     // 血量流失后
    HookHPChanged HookKind = "hp_changed"  // 血量变化后
)
```

#### 新增 Decl 方法

```go
// skill/types.go
type Decl struct {
    // ... 其他字段 ...
    OnHPLost    func(r Runtime, ctx HPLostCtx) error    // 血量流失后（非伤害）
    OnHPChanged func(r Runtime, ctx HPChangedCtx) error // 血量变化后（伤害/流失/回复）
}
```

### 2. 引擎层实现（engine 包）

#### 核心函数

```go
// engine/phase_hp_change.go

// handleHPChange 处理血量变化后的统一逻辑
func (g *Game) handleHPChange(ctx HPChangeContext, events *[]GameEvent)

// applyDamageWithHook 应用伤害并触发钩子（推荐使用的伤害函数）
func (g *Game) applyDamageWithHook(source, target, amount int, damageCard Card, events *[]GameEvent) int

// applyHPLossWithHook 应用血量流失（非伤害扣血）并触发钩子
func (g *Game) applyHPLossWithHook(seat, amount int, reason string, source int, skillID string, events *[]GameEvent)

// applyHealWithHook 应用血量回复并触发钩子
func (g *Game) applyHealWithHook(seat, amount int, reason string, source int, skillID string, events *[]GameEvent)
```

#### 钩子触发函数

```go
// engine/skill_hooks.go

// runHPLostHooks 广播血量流失事件（非伤害扣血）
func (g *Game) runHPLostHooks(seat, amount int, reason string, source int, events *[]GameEvent)

// runHPChangedHooks 广播血量变化事件
func (g *Game) runHPChangedHooks(seat, oldHP, newHP, delta int, reason string, source int, skillID string, events *[]GameEvent)

// applyHeal 统一回复血量并广播 HPChanged
func (g *Game) applyHeal(seat, amount int, reason string, source int, skillID string, events *[]GameEvent)
```

### 3. 血量变化流程

#### 伤害流程

```
applyDamageWithHook(source, target, amount, card, events)
    ↓
applyDamage(source, target, amount, card, events)  // 扣血
    ↓
runDamageDealtHooks(ctx, events)                   // 触发 OnDamageDealt 钩子
    ↓
handleHPChange(ctx, events)                        // 处理血量变化
    ├─ 记录 hp_changed 事件
    ├─ runHPChangedHooks(...)                     // 触发 OnHPChanged 钩子
    └─ 检查濒死状态 → startDyingWindow(...)
```

#### 血量流失流程

```
applyHPLossWithHook(seat, amount, reason, source, skillID, events)
    ↓
runHPLostHooks(seat, amount, reason, source, events)  // 扣血 + 触发 OnHPLost
    ↓
handleHPChange(ctx, events)                            // 处理血量变化
    ├─ 记录 hp_changed 事件
    ├─ runHPChangedHooks(...)                         // 触发 OnHPChanged 钩子
    └─ 检查濒死状态 → startDyingWindow(...)
```

#### 血量回复流程

```
applyHealWithHook(seat, amount, reason, source, skillID, events)
    ↓
applyHeal(seat, amount, reason, source, skillID, events)  // 回复血量
    ↓
handleHPChange(ctx, events)                                // 处理血量变化
    ├─ 记录 hp_changed 事件
    └─ runHPChangedHooks(...)                             // 触发 OnHPChanged 钩子
```

## 使用示例

### 1. 技能监听血量变化

```go
// skill_yiji.go - 【遗计】技能实现
func YijiDecl() skill.Decl {
    return skill.Decl{
        Meta: skill.Meta{
            ID:   "yiji",
            Name: "遗计",
        },
        OnHPChanged: func(r skill.Runtime, ctx skill.HPChangedCtx) error {
            // 当血量变化时检查是否触发遗计
            if ctx.Seat == r.Seat() && ctx.Delta < 0 && ctx.Reason == "damage" {
                // 触发遗计技能
                return r.OfferSkillWindow("yiji")
            }
            return nil
        },
    }
}
```

### 2. 技能监听血量流失

```go
// skill_ganglie.go - 【刚烈】技能实现（假设刚烈改为血量流失）
func GanglieDecl() skill.Decl {
    return skill.Decl{
        Meta: skill.Meta{
            ID:   "ganglie",
            Name: "刚烈",
        },
        OnHPLost: func(r skill.Runtime, ctx skill.HPLostCtx) error {
            // 当血量流失时检查是否触发刚烈
            if ctx.Seat == r.Seat() && ctx.Reason == "skill" {
                // 触发刚烈技能
                return r.OfferSkillWindow("ganglie")
            }
            return nil
        },
    }
}
```

### 3. 调用伤害函数

```go
// 使用新的带钩子的伤害函数
g.applyDamageWithHook(source, target, 1, Card{Kind: CardSha, Name: "杀"}, &events)

// 使用血量流失函数
g.applyHPLossWithHook(target, 1, "skill", source, "ganglie", &events)

// 使用血量回复函数
g.applyHealWithHook(target, 1, "skill", source, "tao", &events)
```

## 测试覆盖

### 测试文件：phase_hp_change_test.go

- **TestHPChangeHooks**：测试伤害导致的血量变化钩子
- **TestHPLossHooks**：测试血量流失钩子
- **TestHealHooks**：测试血量回复钩子
- **TestDyingAfterHPChange**：测试血量变化后濒死触发
- **TestDamageVsHPLoss**：测试伤害和血量流失的区别
- **TestMultipleHPChanges**：测试连续血量变化

### 运行测试

```bash
cd backend && go test ./internal/game/yuzhousha/engine -run "TestHP|TestHeal|TestDying|TestDamageVs|TestMultiple" -v
```

## 兼容性说明

### 向后兼容

- 保留原有的 `applyDamage` 函数，现有代码无需修改
- 新增 `applyDamageWithHook` 函数，推荐新代码使用
- 原有的 `continueAfterDamage` 逻辑不受影响

### 迁移建议

1. **新代码**：直接使用 `applyDamageWithHook`、`applyHPLossWithHook`、`applyHealWithHook`
2. **旧代码**：可以逐步迁移到新函数，获得更清晰的时机控制
3. **技能实现**：使用 `OnHPChanged` 和 `OnHPLost` 钩子监听血量变化

## 未来扩展

### 可能的扩展方向

1. **血量变化阶段**：如果需要更复杂的交互，可以实现独立的血量变化阶段（类似出牌阶段）
2. **优先级系统**：为血量变化钩子添加优先级，控制触发顺序
3. **取消机制**：允许某些技能取消血量变化（如【涅槃】）
4. **连锁反应**：支持血量变化触发其他血量变化（如【反馈】导致【刚烈】）

## 文件清单

### 新增文件

- `engine/phase_hp_change.go` - 血量变化处理逻辑
- `engine/phase_hp_change_test.go` - 测试文件
- `engine/HP_CHANGE_IMPLEMENTATION.md` - 本文档

### 修改文件

- `skill/types.go` - 添加 `OnHPLost` 和 `OnHPChanged` 钩子定义
- `skill/hooks.go` - 添加 `HookHPLost`、`HookHPChanged` 常量和相关上下文类型
- `engine/skill_hooks.go` - 实现新的钩子调用逻辑
- `engine/constants.go` - 添加 `PhaseHPChange` 常量（预留）

## 总结

本实现提供了清晰、统一的血量变化时机系统，支持：
- ✅ 伤害结算后时机（OnDamageDealt + OnHPChanged）
- ✅ 血量流失时机（OnHPLost + OnHPChanged）
- ✅ 血量回复时机（OnHPChanged）
- ✅ 濒死状态自动检查
- ✅ 完整的事件广播
- ✅ 向后兼容

通过这套系统，所有涉及血量变化的逻辑都有了清晰的触发时机，便于技能实现和调试。
