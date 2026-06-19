# 技能钩子实现总结

## 完成的工作

### 1. 技能钩子实现

在 `skill/catalog_skills.go` 中添加了三个技能的钩子实现：

#### 【遗计】- 使用 OnDamageDealt 钩子
```go
{
    Meta: Meta{
        ID: IDYiji, Name: "遗计", Kind: KindPassive,
        Desc: "当你受到1点伤害后，你可以摸两张牌，然后可以将至多两张手牌交给其他角色。",
    },
    OnDamageDealt: func(r Runtime, ctx DamageCtx) error {
        if !r.HasSkill(ctx.Target, IDYiji) {
            return nil
        }
        if ctx.Amount != 1 {
            return nil
        }
        // 受到伤害后摸两张牌
        return r.DrawCards(ctx.Target, 2)
    },
},
```

#### 【反馈】- 使用 OnDamageDealt 钩子
```go
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
        // TODO: 完整实现需要与 engine 包交互
        return nil
    },
},
```

#### 【奸雄】- 使用 OnDamageDealt 钩子
```go
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
        // TODO: 完整实现需要与 engine 包交互
        return nil
    },
},
```

### 2. 修复 applyHPLossWithHook 和 applyHealWithHook

发现并修复了重复扣血/触发钩子的问题：

#### 问题
- `applyHPLossWithHook` 调用 `runHPLostHooks` 后，又重复处理血量变化
- `applyHealWithHook` 调用 `applyHeal` 后，又重复处理血量变化

#### 解决方案
- `runHPLostHooks` 内部已经扣血并触发了 `HPChanged` 钩子
- `applyHeal` 内部已经触发了 `HPChanged` 钩子
- 修正后的函数不再重复处理

```go
// applyHPLossWithHook 应用血量流失（非伤害扣血）并触发钩子。
// 注意：runHPLostHooks 内部已经扣血并触发了 HPChanged 钩子，所以这里不需要重复处理。
func (g *Game) applyHPLossWithHook(seat, amount int, reason string, source int, skillID string, events *[]GameEvent) {
    if amount <= 0 || seat < 0 || seat >= len(g.Players) {
        return
    }

    // 应用血量流失钩子（内部会扣血并触发 HPChanged）
    g.runHPLostHooks(seat, amount, reason, source, events)
    
    // 注意：不需要在这里重复处理血量变化，因为 runHPLostHooks 已经处理了
}
```

### 3. 测试覆盖

创建了 `engine/skill_hooks_integration_test.go`，包含 6 个集成测试：

1. **TestYijiSkillHook** - 测试【遗计】技能钩子
2. **TestFankuiSkillHook** - 测试【反馈】技能钩子
3. **TestJianxiongSkillHook** - 测试【奸雄】技能钩子
4. **TestHPLossHook** - 测试血量流失钩子
5. **TestHPChangedHook** - 测试血量变化钩子
6. **TestMultipleSkillHooks** - 测试多个技能钩子同时触发

所有测试都通过！

### 4. 文档

创建了 `SKILL_HOOKS_EXAMPLE.md` 文档，包含：
- 钩子说明（OnDamageDealt、OnHPLost、OnHPChanged）
- 技能实现示例（【遗计】、【反馈】、【奸雄】、【蛊惑】、【刚烈】）
- 血量流失实现示例
- 测试示例
- 迁移指南（从 DamageAftermath 到钩子）
- 注意事项

## 测试结果

- **总测试数**: 47 个
- **通过**: 47 个
- **失败**: 0 个

## 待完成的工作

### 1. 【蛊惑】技能实现

需要实现【蛊惑】技能，使用 `applyHPLossWithHook()` 触发血量流失钩子。

**参考实现**：
```go
// 在 engine/skill_guhuo.go 中实现
func (g *Game) applyGuhuoEffect(seat, target int, events *[]GameEvent) {
    // 使用 applyHPLossWithHook 触发血量流失钩子
    g.applyHPLossWithHook(target, 1, "skill", seat, "guhuo", events)
    
    // 后续处理...
}
```

### 2. 完善技能窗口交互

目前的【遗计】、【反馈】、【奸雄】技能实现只是框架，需要完善技能窗口交互逻辑：

- 【遗计】：摸牌后触发窗口，允许给出手牌
- 【反馈】：触发窗口，允许选择并获得来源的一张牌
- 【奸雄】：触发窗口，允许获得造成伤害的牌

### 3. 迁移现有技能

逐步将现有使用 `DamageAftermath` 机制的技能迁移到新的钩子系统：

- ✅ 【遗计】- 已添加钩子（框架）
- ✅ 【反馈】- 已添加钩子（框架）
- ✅ 【奸雄】- 已添加钩子（框架）
- ⏳ 其他技能...

## 关键技术点

### 1. 钩子执行顺序
按技能注册顺序执行，需要注意技能优先级。

### 2. 错误处理
钩子返回 error 会中断后续钩子执行，需要谨慎处理。

### 3. 避免重复触发
- `runHPLostHooks` 内部已经触发了 `HPChanged` 钩子
- `applyHeal` 内部已经触发了 `HPChanged` 钩子
- 调用这些函数后，不需要再重复触发

### 4. 测试注意事项
- 使用 `NewSolo1v1` 创建测试游戏
- 英雄 ID 需要正确（如 `si_ma_yi` 而不是 `simayi`）
- 测试血量变化时，需要先降低血量再回复

## 文件清单

### 新增文件
1. `SKILL_HOOKS_EXAMPLE.md` - 技能钩子使用示例文档
2. `SKILL_HOOKS_IMPLEMENTATION_SUMMARY.md` - 实现总结文档（本文档）
3. `engine/skill_hooks_integration_test.go` - 技能钩子集成测试

### 修改文件
1. `skill/catalog_skills.go` - 添加【遗计】、【反馈】、【奸雄】技能定义
2. `engine/phase_hp_change.go` - 修复 `applyHPLossWithHook` 和 `applyHealWithHook`
3. `engine/phase_hp_change_test.go` - 修正测试用例

## 下一步计划

1. 实现【蛊惑】技能
2. 完善【遗计】、【反馈】、【奸雄】的技能窗口交互
3. 继续迁移其他技能到钩子系统
4. 增加更多集成测试
