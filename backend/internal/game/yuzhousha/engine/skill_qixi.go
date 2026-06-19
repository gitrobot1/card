package engine

import (
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterQixiActive      = "qixi_active"
	ResponseModeSkillQixi  = "skill_qixi"
)

// qixiCanActivate 奇袭：出牌阶段，你可以将一张黑色的牌当过河拆桥打出。
func qixiCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDQixi) {
		return false
	}
	// 出牌阶段才能激活
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	// 如果已经激活，可以取消激活
	if r.SkillCounter(seat, counterQixiActive) > 0 {
		return true
	}
	// 检查是否有黑色牌（手牌或装备区）
	return hasBlackCard(r, seat)
}

// qixiActivate 激活奇袭
func qixiActivate(r skill.Runtime, seat int, _ skill.ActivateReq) error {
	return r.ToggleQixi(seat)
}

// qixiCardPlaysAs 奇袭：黑色牌视为过河拆桥
func qixiCardPlaysAs(r skill.Runtime, seat int, _, asKind, suit string) bool {
	if !r.HasSkill(seat, skill.IDQixi) || asKind != CardGuoHe || !skill.IsBlackSuit(suit) {
		return false
	}
	// 只要奇袭处于激活状态，黑色牌就视为过河拆桥
	return r.SkillCounter(seat, counterQixiActive) > 0
}

// qixiAIPriority AI 优先级
func qixiAIPriority(r skill.Runtime, seat int) int {
	if !qixiCanActivate(r, seat) || r.SkillCounter(seat, counterQixiActive) > 0 {
		return 0
	}
	// 如果有过河拆桥的目标，优先激活奇袭
	if r.HandPlaysAs(seat, CardGuoHe) {
		return 70
	}
	return 0
}

// qixiAIActivate AI 激活奇袭
func qixiAIActivate(r skill.Runtime, seat int) error {
	return r.ToggleQixi(seat)
}

// hasBlackCard 检查是否有黑色牌（手牌或装备区）
func hasBlackCard(r skill.Runtime, seat int) bool {
	return r.HasBlackCard(seat)
}

// hasBlackHandCard 检查是否有黑色手牌
func (g *Game) hasBlackHandCard(seat int) bool {
	for _, c := range g.Players[seat].Hand {
		if skill.IsBlackSuit(c.Suit) {
			return true
		}
	}
	return false
}
