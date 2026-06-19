# 技能钩子使用示例

本文档展示如何使用新的技能钩子系统实现【遗计】、【反馈】、【奸雄】、【蛊惑】等技能。

## 钩子说明

### 1. OnDamageDealt（造成伤害后）
- **触发时机**：角色受到伤害后
- **使用场景**：【遗计】、【反馈】、【刚烈】等

### 2. OnHPLost（血量流失后）
- **触发时机**：非伤害导致的扣血（如【蛊惑】）
- **使用场景**：响应血量流失事件

### 3. OnHPChanged（血量变化后）
- **触发时机**：任何血量变化（伤害/流失/回复）
- **使用场景**：需要响应所有血量变化的技能

## 技能实现示例

### 【遗计】- 使用 OnDamageDealt

```go
// 在 skill/catalog_skills.go 中添加
{
    Meta: Meta{
        ID: IDYiji, Name: "遗计", Kind: KindPassive,
        Desc: "当你受到1点伤害后，你可以摸两张牌，然后可以将至多两张手牌交给其他角色。",
    },
    OnDamageDealt: func(r Runtime, ctx DamageCtx) error {
        if !r.HasSkill(ctx.Target, IDYiji) {
            return nil
        }
        // 受到伤害后摸牌
        if err := r.DrawCards(ctx.Target, 2); err != nil {
            return err
        }
        // 触发技能窗口，允许给出手牌
        return r.OfferSkillWindow(ctx.Target, IDYiji, "yiji_give", map[string]interface{}{
            "max_give": 2,
        })
    },
},
```

### 【反馈】- 使用 OnDamageDealt

```go
// 在 skill/catalog_skills.go 中添加
{
    Meta: Meta{
        ID: IDFankui, Name: "反馈", Kind: KindPassive,
        Desc: "当你受到伤害后，你可以获得伤害来源的一张牌。",
    },
    OnDamageDealt: func(r Runtime, ctx DamageCtx) error {
        if !r.HasSkill(ctx.Target, IDFankui) {
            return nil
        }
        if ctx.Source < 0 {
            return nil
        }
        // 获得伤害来源的一张牌
        return r.OfferTakeWindow(ctx.Target, ctx.Source, 1, "feedback_take")
    },
},
```

### 【奸雄】- 使用 OnDamageDealt

```go
// 在 skill/catalog_skills.go 中添加
{
    Meta: Meta{
        ID: IDJianxiong, Name: "奸雄", Kind: KindPassive,
        Desc: "当你受到伤害后，你可以获得造成此伤害的牌。",
    },
    OnDamageDealt: func(r Runtime, ctx DamageCtx) error {
        if !r.HasSkill(ctx.Target, IDJianxiong) {
            return nil
        }
        if ctx.Card.ID == "" {
            return nil
        }
        // 获得造成伤害的牌
        return r.OfferSkillWindow(ctx.Target, IDJianxiong, "jianxiong_take", map[string]interface{}{
            "card_id": ctx.Card.ID,
        })
    },
},
```

### 【蛊惑】- 使用 applyHPLossWithHook()

```go
// 在 engine/skill_guhuo.go 中实现
func (g *Game) applyGuhuoEffect(seat, target int, events *[]GameEvent) {
    // 使用 applyHPLossWithHook 触发血量流失钩子
    g.applyHPLossWithHook(target, 1, "skill", seat, "guhuo", events)
    
    // 后续处理...
}
```

### 【刚烈】- 使用 OnDamageDealt

```go
// 在 skill/catalog_skills.go 中添加
{
    Meta: Meta{
        ID: IDGanglie, Name: "刚烈", Kind: KindPassive,
        Desc: "当你受到1点伤害后，你可以进行判定，若判定牌不为♥，则伤害来源弃置两张牌。",
    },
    OnDamageDealt: func(r Runtime, ctx DamageCtx) error {
        if !r.HasSkill(ctx.Target, IDGanglie) {
            return nil
        }
        if ctx.Amount != 1 {
            return nil
        }
        // 触发判定窗口
        return r.OfferJudgeWindow(ctx.Target, "ganglie", ctx.Source)
    },
},
```

## 血量流失实现

### 使用 applyHPLossWithHook()

```go
// 在 engine/phase_hp_change.go 中定义
func (g *Game) applyHPLossWithHook(seat, amount int, reason string, source int, skillID string, events *[]GameEvent) {
    if amount <= 0 || seat < 0 || seat >= len(g.Players) {
        return
    }
    
    p := &g.Players[seat]
    oldHP := p.HP
    
    // 应用血量流失钩子
    g.runHPLostHooks(seat, amount, reason, source, events)
    
    // 处理血量变化
    if p.HP != oldHP {
        g.handleHPChange(HPChangeContext{
            Seat:    seat,
            OldHP:   oldHP,
            NewHP:   p.HP,
            Delta:   p.HP - oldHP,
            Reason:  "hp_loss",
            Source:  source,
            SkillID: skillID,
        }, events)
    }
}
```

## 测试示例

```go
// 在 engine/phase_hp_change_test.go 中添加
func TestYijiSkill(t *testing.T) {
    g := newTestGame()
    g.addPlayer("A", "yiji_hero")
    g.addPlayer("B", "normal_hero")
    
    // 模拟受到伤害
    g.applyDamageWithHook(1, 0, 1, Card{}, &[]GameEvent{})
    
    // 验证遗计触发
    assert.True(t, g.hasSkill(0, "yiji"))
    assert.Equal(t, 2, len(g.Players[0].Hand)) // 摸两张牌
}
```

## 迁移指南

### 从 DamageAftermath 迁移到钩子

**旧方式（DamageAftermath）**：
```go
// 在 engine/skill_yiji.go 中
func (g *Game) offerYijiWindow(a *DamageAftermath, events *[]GameEvent) bool {
    // 复杂的 aftermath 逻辑
}
```

**新方式（钩子）**：
```go
// 在 skill/catalog_skills.go 中
{
    Meta: Meta{ID: IDYiji, Name: "遗计"},
    OnDamageDealt: func(r Runtime, ctx DamageCtx) error {
        // 简洁的钩子逻辑
        return r.DrawCards(ctx.Target, 2)
    },
},
```

## 注意事项

1. **钩子执行顺序**：按技能注册顺序执行
2. **错误处理**：钩子返回 error 会中断后续钩子执行
3. **性能考虑**：避免在钩子中执行耗时操作
4. **兼容性**：旧的 DamageAftermath 机制仍然保留，建议逐步迁移

## 参考资料

- `skill/hooks.go` - 钩子类型定义
- `engine/skill_hooks.go` - 钩子调用实现
- `engine/phase_hp_change.go` - 血量变化处理
- `skill/catalog_skills.go` - 技能注册示例
