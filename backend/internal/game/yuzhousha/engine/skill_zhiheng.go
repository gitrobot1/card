package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const counterZhihengUsed = "zhiheng_used_play"

func (g *Game) ActivateZhiheng(seat int, cardIDs []string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillZhiheng) || g.getSkillCounter(seat, counterZhihengUsed) > 0 {
		return ErrWrongPhase
	}
	if len(cardIDs) == 0 {
		return ErrInvalidCard
	}
	discarded := make([]Card, 0, len(cardIDs))
	for _, id := range cardIDs {
		idx, _, ok := g.findCard(seat, id)
		if !ok {
			return ErrInvalidCard
		}
		discarded = append(discarded, g.removeHandCard(seat, idx, events))
	}
	g.DiscardPile = append(g.DiscardPile, discarded...)
	g.SyncCounts()
	g.runCardsDiscardedHooks(seat, "cost", discarded, events)
	g.drawCards(seat, len(discarded), events)
	g.setSkillCounter(seat, counterZhihengUsed, 1)
	msg := fmt.Sprintf("%s 发动【制衡】，弃 %d 张摸 %d 张", g.Players[seat].Name, len(discarded), len(discarded))
	g.appendSkillEvent(events, skill.IDZhiheng, seat, seat, msg)
	g.Message = msg
	g.resetTimer()
	return nil
}
