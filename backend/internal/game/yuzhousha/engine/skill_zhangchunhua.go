package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) finishJueqingDeath(source, target int, events *[]GameEvent) bool {
	if target < 0 || target >= len(g.Players) || g.Players[target].HP > 0 {
		return false
	}
	victim := g.Players[target].Name
	msg := fmt.Sprintf("%s 因【绝情】失去体力至 0，阵亡", victim)
	*events = append(*events, GameEvent{
		Type:        "jueqing_death",
		PlayerIndex: target,
		TargetIndex: source,
		SkillID:     skill.IDJueqing,
		Message:     msg,
	})
	if g.checkTeamElimination(events) {
		return true
	}
	if g.checkChainDeath(target, events) {
		return true
	}
	if g.is2v2() {
		g.Phase = PhasePlaying
		g.Message = fmt.Sprintf("%s 阵亡，对局继续", victim)
		g.resetTimer()
		return true
	}
	g.finishGame(source, events)
	return true
}
