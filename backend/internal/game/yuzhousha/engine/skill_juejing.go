package engine

import (
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// SkillJuejing 绝境技能
// 锁定技，你的手牌上限+2；当你进入或脱离濒死状态时，你摸一张牌。
func init() {
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID:   skill.IDJuejing,
			Name: "绝境",
			Kind: skill.KindPassive,
			Desc: "锁定技，你的手牌上限+2；当你进入或脱离濒死状态时，你摸一张牌。",
		},
		HandRetainLimit: juejingHandRetainLimit,
		// 进入濒死状态时摸一张牌
		OnHPChanged: juejingOnHPChanged,
	})
}

// juejingHandRetainLimit 绝境：手牌上限+2
func juejingHandRetainLimit(r skill.Runtime, seat int) int {
	if !r.HasSkill(seat, skill.IDJuejing) {
		return 0
	}
	return 2
}

// juejingOnHPChanged 血量变化时检查是否进入或脱离濒死状态
func juejingOnHPChanged(r skill.Runtime, ctx skill.HPChangedCtx) error {
	// 只有拥有绝境技能的角色才触发
	if !r.HasSkill(ctx.Seat, skill.IDJuejing) {
		return nil
	}
	
	// 进入濒死状态（血量变为0或以下）
	if ctx.NewHP <= 0 && ctx.OldHP > 0 {
		return r.DrawSkillCards(ctx.Seat, skill.IDJuejing, 1, "")
	}
	
	// 脱离濒死状态（血量从0或以下变为大于0）
	if ctx.OldHP <= 0 && ctx.NewHP > 0 {
		return r.DrawSkillCards(ctx.Seat, skill.IDJuejing, 1, "")
	}
	
	return nil
}
