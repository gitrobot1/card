package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// HPChangeContext 血量变化后的技能触发上下文。
type HPChangeContext struct {
	Seat    int
	OldHP   int
	NewHP   int
	Delta   int    // 变化量：正数=回复，负数=扣血
	Reason  string // damage | hp_loss | heal | skill
	Source  int    // 来源座位号
	SkillID string // 技能ID（若是技能导致）
	Damage  int    // 伤害值（若是伤害导致）
}

// handleHPChange 处理血量变化后的统一逻辑。
// 这是血量变化后的核心处理函数，会触发相应的技能钩子。
func (g *Game) handleHPChange(ctx HPChangeContext, events *[]GameEvent) {
	if ctx.Delta == 0 {
		return
	}

	seat := ctx.Seat
	p := &g.Players[seat]

	// 记录血量变化事件
	*events = append(*events, GameEvent{
		Type:        "hp_changed",
		PlayerIndex: seat,
		TargetIndex: ctx.Source,
		Damage:      ctx.Damage,
		Heal:        ctx.Delta,
		Message:     fmt.Sprintf("%s 血量变化：%d → %d (%s)", p.Name, ctx.OldHP, ctx.NewHP, ctx.Reason),
		SkillID:     ctx.SkillID,
	})

	// 触发血量变化钩子
	g.runHPChangedHooks(seat, ctx.OldHP, ctx.NewHP, ctx.Delta, ctx.Reason, ctx.Source, ctx.SkillID, events)

	// 参考 noname damage.content step 5: if (hp<=0) player.dying(event)
	// 濒死统一由 ApplyDamageAndCheckDeath 在扣血后自动处理。
	// handleHPChange 作为底层函数，只负责通知血量变化和触发钩子。
}

// applyDamageWithHook 应用伤害并触发钩子（推荐使用的伤害函数）。
// 与 applyDamage 的区别：这个函数会额外处理血量流失和血量变化钩子。
// 重构后：扣血后自动检查濒死（参考 noname damage.content → if hp<=0 → dying()）。
func (g *Game) applyDamageWithHook(source, target, amount int, damageCard Card, events *[]GameEvent) int {
	if amount <= 0 || target < 0 || target >= len(g.Players) {
		return 0
	}

	p := &g.Players[target]
	oldHP := p.HP

	// 应用伤害
	actualDamage := g.applyDamage(source, target, amount, damageCard, events)

	// 处理血量变化
	if p.HP != oldHP {
		g.handleHPChange(HPChangeContext{
			Seat:    target,
			OldHP:   oldHP,
			NewHP:   p.HP,
			Delta:   p.HP - oldHP,
			Reason:  "damage",
			Source:  source,
			Damage:  actualDamage,
			SkillID: "",
		}, events)
	}

	return actualDamage
}

// ApplyDamageAndCheckDeath 应用伤害并自动检查濒死/死亡。
// 内部使用 StartDamageEvent（完整 GameEvent 生命周期）。
// 返回 true 表示进入了濒死流程。
func (g *Game) ApplyDamageAndCheckDeath(source, target, amount int, damageCard Card, resume DamageResume, events *[]GameEvent) bool {
	return g.applyDamageAndCheckDeathImpl(source, target, amount, damageCard, resume, events)
}

// applyHPLossWithHook 应用血量流失（非伤害扣血）并触发钩子。
// 用于【蛊惑】、【刚烈】等导致血量流失的技能。
// 注意：runHPLostHooks 内部已经扣血并触发了 HPChanged 钩子，所以这里不需要重复处理。
func (g *Game) applyHPLossWithHook(seat, amount int, reason string, source int, skillID string, events *[]GameEvent) {
	if amount <= 0 || seat < 0 || seat >= len(g.Players) {
		return
	}

	// 应用血量流失钩子（内部会扣血并触发 HPChanged）
	g.runHPLostHooks(seat, amount, reason, source, events)
	
	// 注意：不需要在这里重复处理血量变化，因为 runHPLostHooks 已经处理了
}

// applyHealWithHook 应用血量回复并触发钩子。
// 用于【桃】、【酒】、技能回复等。
// 注意：applyHeal 内部已经触发了 HPChanged 钩子，所以这里不需要重复处理。
func (g *Game) applyHealWithHook(seat, amount int, reason string, source int, skillID string, events *[]GameEvent) {
	if amount <= 0 || seat < 0 || seat >= len(g.Players) {
		return
	}

	// 应用回复（内部会触发 HPChanged）
	g.applyHeal(seat, amount, reason, source, skillID, events)
}

// IsHPLoss 判断是否为血量流失（非伤害扣血）。
// 血量流失不会被【麒麟弓】、【反馈】等技能响应。
func (g *Game) IsHPLoss(reason string) bool {
	return reason == "hp_loss" || reason == "skill"
}

// GetHPChangeReason 获取血量变化原因的描述。
func (g *Game) GetHPChangeReason(ctx HPChangeContext) string {
	switch ctx.Reason {
	case "damage":
		return fmt.Sprintf("受到 %d 点伤害", ctx.Damage)
	case "hp_loss":
		return "血量流失"
	case "heal":
		return fmt.Sprintf("回复 %d 点血量", ctx.Delta)
	case "skill":
		skillName := ctx.SkillID
		if h, ok := skill.Lookup(ctx.SkillID); ok {
			skillName = h.Meta().Name
		}
		return fmt.Sprintf("技能【%s】效果", skillName)
	default:
		return "未知原因"
	}
}
