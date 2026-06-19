package engine

import "github.com/time/card/backend/internal/game/yuzhousha/skill"

const SkillChongzhen = skill.IDChongzhen

// notifyBecameTarget 某角色成为牌的目标后（如【激昂】）。
// 注意：冲阵不是在成为目标时触发，而是在发动龙胆时触发
func (g *Game) notifyBecameTarget(target, source int, card Card, events *[]GameEvent) {
	if target < 0 || target >= len(g.Players) {
		return
	}
	// 调用 OnBecomeTarget 钩子（如激昂）
	g.runBecomeTargetHooks(target, source, card, events)
}
