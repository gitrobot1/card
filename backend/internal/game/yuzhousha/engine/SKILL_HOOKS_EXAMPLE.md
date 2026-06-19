# 技能钩子使用示例文档

本文档展示如何使用新的血量变化钩子（`OnHPLost`、`OnHPChanged`）实现技能。

## 1. 监听伤害结算后（`OnDamageDealt`）

### 示例：【奸雄】- 曹操

```go
// skill_jianxiong.go
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID:   skill.IDJianxiong,
        Name: "奸雄",
        Kind: skill.KindPassive,
        Desc: "当你受到伤害后，你可以获得造成此伤害的牌。",
    },
    OnDamageDealt: func(r skill.Runtime, ctx skill.DamageCtx) error {
        // 受到伤害后，可以获得伤害来源的牌
        if ctx.Target == r.Seat() && ctx.CardKind != "" {
            return r.OfferSkillWindow(skill.IDJianxiong)
        }
        return nil
    },
})
```

## 2. 监听血量流失后（`OnHPLost`）

### 示例：【蛊惑】- 张春华（假设实现）

```go
// skill_zhangchunhua.go
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID:   "guhuo",
        Name: "蛊惑",
        Kind: skill.KindPassive,
        Desc: "当你造成血量流失时，可以观看目标的手牌。",
    },
    OnHPLost: func(r skill.Runtime, ctx skill.HPLostCtx) error {
        // 当目标血量流失时触发
        if ctx.Source == r.Seat() && ctx.Reason == "skill" {
            return r.OfferSkillWindow("guhuo")
        }
        return nil
    },
})
```

## 3. 监听血量变化后（`OnHPChanged`）

### 示例：【遗计】- 郭嘉（改进版）

```go
// skill_yiji.go
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID:   skill.IDYiji,
        Name: "遗计",
        Kind: skill.KindPassive,
        Desc: "当你受到1点伤害后，你可以摸两张牌，然后可以将至多两张手牌交给其他角色。",
    },
    OnHPChanged: func(r skill.Runtime, ctx skill.HPChangedCtx) error {
        // 当受到1点伤害后触发
        if ctx.Seat == r.Seat() && ctx.Delta == -1 && ctx.Reason == "damage" {
            return r.OfferSkillWindow(skill.IDYiji)
        }
        return nil
    },
})
```

### 示例：【反馈】- 司马懿（改进版）

```go
// skill_fankui.go
skill.Register(skill.Decl{
    Meta: skill.Meta{
        ID:   skill.IDFankui,
        Name: "反馈",
        Kind: skill.KindPassive,
        Desc: "当你受到1点伤害后，你可以获得伤害来源的一张牌。",
    },
    OnHPChanged: func(r skill.Runtime, ctx skill.HPChangedCtx) error {
        // 当受到1点伤害后触发
        if ctx.Seat == r.Seat() && ctx.Delta < 0 && ctx.Reason == "damage" {
            if r.HasTakeableCard(ctx.Source) {
                return r.OfferSkillWindow(skill.IDFankui)
            }
        }
        return nil
    },
})
```

## 4. 完整的技能实现流程

### 4.1 【遗计】完整实现

```go
// skill_yiji.go

func init() {
    skill.Register(skill.Decl{
        Meta: skill.Meta{
            ID:   skill.IDYiji,
            Name: "遗计",
            Kind: skill.KindPassive,
            Desc: "当你受到1点伤害后，你可以摸两张牌，然后可以将至多两张手牌交给其他角色。",
        },
        OnHPChanged: func(r skill.Runtime, ctx skill.HPChangedCtx) error {
            // 触发条件：受到1点伤害
            if ctx.Seat == r.Seat() && ctx.Delta == -1 && ctx.Reason == "damage" {
                return r.OfferSkillWindow(skill.IDYiji)
            }
            return nil
        },
    })
}

// 在 engine 层处理遗计窗口
func (g *Game) ApplyYiji(seat int, events *[]GameEvent) error {
    // 1. 摸两张牌
    g.drawCards(seat, 2, events)
    
    // 2. 打开交牌窗口
    g.Pending = &PendingCombat{
        ResponseMode: ResponseModeSkillYijiGive,
        SourceIndex:  seat,
        TargetIndex: seat,
        YijiGiveRemaining: 2,
    }
    g.Message = fmt.Sprintf("%s 可将至多2张手牌交给其他角色", g.Players[seat].Name)
    g.resetTimer()
    
    return nil
}

func (g *Game) YijiGiveCards(seat, target int, cardIDs []string, events *[]GameEvent) error {
    // 将手牌交给目标
    // ...
    return nil
}
```

## 5. 血量变化时机对比

### 5.1 伤害流程

```
applyDamageWithHook(source, target, amount, card, events)
    ↓
applyDamage(source, target, amount, card, events)  // 扣血
    ↓
runDamageDealtHooks(ctx, events)                   // 触发 OnDamageDealt（如【奸雄】）
    ↓
handleHPChange(ctx, events)                        // 处理血量变化
    ├─ 记录 hp_changed 事件
    ├─ runHPChangedHooks(...)                     // 触发 OnHPChanged（如【遗计】、【反馈】）
    └─ 检查濒死状态 → startDyingWindow(...)
```

### 5.2 血量流失流程

```
applyHPLossWithHook(seat, amount, reason, source, skillID, events)
    ↓
runHPLostHooks(seat, amount, reason, source, events)  // 扣血 + 触发 OnHPLost
    ↓
handleHPChange(ctx, events)                            // 处理血量变化
    ├─ 记录 hp_changed 事件
    ├─ runHPChangedHooks(...)                         // 触发 OnHPChanged
    └─ 检查濒死状态 → startDyingWindow(...)
```

### 5.3 血量回复流程

```
applyHealWithHook(seat, amount, reason, source, skillID, events)
    ↓
applyHeal(seat, amount, reason, source, skillID, events)  // 回复血量
    ↓
handleHPChange(ctx, events)                                // 处理血量变化
    └─ 记录 hp_changed 事件 + 触发 OnHPChanged
```

## 6. 最佳实践

### 6.1 选择合适的钩子

- **`OnDamageDealt`**：需要获得伤害牌、修改伤害值时使用（如【奸雄】、【刚烈】）
- **`OnHPLost`**：需要响应血量流失（非伤害扣血）时使用（如【蛊惑】）
- **`OnHPChanged`**：需要在血量变化后触发技能时使用（如【遗计】、【反馈】）

### 6.2 触发条件判断

```go
OnHPChanged: func(r skill.Runtime, ctx skill.HPChangedCtx) error {
    // 1. 检查是否是自己
    if ctx.Seat != r.Seat() {
        return nil
    }
    
    // 2. 检查变化量
    if ctx.Delta >= 0 {
        return nil // 只关心扣血
    }
    
    // 3. 检查原因
    if ctx.Reason != "damage" {
        return nil // 只关心伤害
    }
    
    // 4. 检查具体条件（如伤害值）
    if -ctx.Delta != 1 {
        return nil // 只关心1点伤害
    }
    
    // 5. 检查额外条件（如是否有可获得的牌）
    if !r.HasTakeableCard(ctx.Source) {
        return nil
    }
    
    // 6. 触发技能窗口
    return r.OfferSkillWindow(skillID)
}
```

### 6.3 事件广播

所有血量变化都会自动广播 `hp_changed` 事件：

```json
{
    "type": "hp_changed",
    "player_index": 0,
    "target_index": 1,
    "damage": 1,
    "heal": 0,
    "message": "甲 血量变化：4 → 3 (damage)",
    "skill_id": ""
}
```

## 7. 测试示例

```go
// phase_hp_change_test.go

func TestYijiTrigger(t *testing.T) {
    g, err := NewSolo1v1("yiji-test", "甲", "guo_jia", "liu_bei")
    if err != nil {
        t.Fatal(err)
    }

    seat := 0 // 郭嘉
    events := []GameEvent{}

    // 造成1点伤害
    g.applyDamageWithHook(1, seat, 1, Card{Kind: CardSha, Name: "杀"}, &events)

    // 检查是否触发了遗计窗口
    found := false
    for _, e := range events {
        if e.Type == "skill_yiji" && e.PlayerIndex == seat {
            found = true
            break
        }
    }
    if !found {
        t.Error("yiji skill window not offered")
    }
}
```

## 8. 常见问题

### Q1: 为什么【遗计】不触发？

**A**: 检查以下条件：
1. 是否使用了 `applyDamageWithHook()` 而不是 `applyDamage()`
2. `OnHPChanged` 中的触发条件是否正确
3. 伤害值是否是1点（`ctx.Delta == -1`）

### Q2: 【刚烈】造成的伤害会触发【遗计】吗？

**A**: 会！因为【刚烈】使用了 `applyDamageWithHook()`，会触发 `OnHPChanged` 钩子。

### Q3: 如何在技能中获得伤害牌？

**A**: 使用 `OnDamageDealt` 钩子，此时 `ctx.Card` 包含伤害牌信息。

## 9. 总结

✅ **推荐使用的新函数**：
- `applyDamageWithHook()` - 造成伤害
- `applyHPLossWithHook()` - 血量流失
- `applyHealWithHook()` - 血量回复

✅ **推荐的技能钩子**：
- `OnDamageDealt` - 伤害结算后
- `OnHPLost` - 血量流失后
- `OnHPChanged` - 血量变化后（最通用）

✅ **自动处理的逻辑**：
- 濒死状态检查
- 事件广播（`hp_changed`）
- 技能窗口触发

通过这套系统，所有血量相关的技能都有了清晰的触发时机！🎉
