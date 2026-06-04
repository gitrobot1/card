package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterGuoseUsed        = "guose_used_play"
	counterGuoseShaBlocked  = "guose_sha_blocked"
)

func (g *Game) hasDiamondHandCard(seat int) bool {
	for _, c := range g.Players[seat].Hand {
		if c.Suit == "D" {
			return true
		}
	}
	return false
}

func (g *Game) ActivateGuose(seat, target int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillGuose) || g.getSkillCounter(seat, counterGuoseUsed) > 0 {
		return ErrWrongPhase
	}
	if target < 0 || target >= len(g.Players) || target == seat {
		return ErrInvalidTarget
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || cardObj.Suit != "D" {
		return ErrInvalidCard
	}
	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.syncCounts()
	g.runCardsDiscardedHooks(seat, "cost", []Card{discarded}, events)
	g.setSkillCounter(seat, counterGuoseUsed, 1)
	g.setSkillCounter(target, counterGuoseShaBlocked, 1)

	msg := fmt.Sprintf("%s 发动【国色】，%s 本回合不能使用【杀】", g.Players[seat].Name, g.Players[target].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDGuose, seat, target, msg)
	g.resetTimer()
	return nil
}
