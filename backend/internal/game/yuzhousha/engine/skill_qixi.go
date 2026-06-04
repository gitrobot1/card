package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterQixiUsed        = "qixi_used_play"
	ResponseModeSkillQixi  = "skill_qixi"
)

func (g *Game) hasBlackHandCard(seat int) bool {
	for _, c := range g.Players[seat].Hand {
		if skill.IsBlackSuit(c.Suit) {
			return true
		}
	}
	return false
}

func (g *Game) opponentHasHandCard(seat int) bool {
	opp := g.firstEnemyWhere(seat, func(e int) bool { return len(g.Players[e].Hand) > 0 })
	return len(g.Players[opp].Hand) > 0
}

func (g *Game) ActivateQixi(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillQixi) || g.getSkillCounter(seat, counterQixiUsed) > 0 {
		return ErrWrongPhase
	}
	opp := g.firstEnemyWhere(seat, func(e int) bool { return len(g.Players[e].Hand) > 0 })
	if len(g.Players[opp].Hand) == 0 {
		return ErrInvalidTarget
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || !skill.IsBlackSuit(cardObj.Suit) {
		return ErrInvalidCard
	}
	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.syncCounts()
	g.runCardsDiscardedHooks(seat, "cost", []Card{discarded}, events)
	g.setSkillCounter(seat, counterQixiUsed, 1)

	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  opp,
		TargetIndex:  seat,
		ReturnIndex:  seat,
		ResponseMode: ResponseModeSkillQixi,
		SkillID:      skill.IDQixi,
	}
	msg := fmt.Sprintf("%s 发动【奇袭】，请选择获得 %s 的一张手牌", g.Players[seat].Name, g.Players[opp].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDQixi, seat, opp, msg)
	g.resetTimer()
	return nil
}

func (g *Game) QixiTakeFrom(seat int, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillQixi || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	source := g.Pending.SourceIndex
	spec := PlayTarget{SeatIndex: source, Zone: "hand", CardID: cardID}
	if cardID == "" && len(g.Players[source].Hand) > 0 {
		spec.CardID = g.Players[source].Hand[0].ID
	}
	card, label, ok := g.takeTargetCard(source, spec, events)
	if !ok {
		return ErrInvalidTarget
	}
	g.Players[seat].Hand = append(g.Players[seat].Hand, card)
	g.syncCounts()
	msg := fmt.Sprintf("%s 发动【奇袭】，获得 %s 的%s", g.Players[seat].Name, g.Players[source].Name, label)
	g.appendSkillEvent(events, skill.IDQixi, seat, source, msg)
	*events = append(*events, GameEvent{
		Type:        "qixi_take",
		PlayerIndex: seat,
		TargetIndex: source,
		Card:        &card,
		Message:     msg,
	})
	return g.finishQixi(seat, events)
}

func (g *Game) finishQixi(seat int, events *[]GameEvent) error {
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = seat
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[seat].Name)
	g.resetTimer()
	return nil
}

func (g *Game) aiPickHandTakeTarget(target int) (zone, cardID string) {
	if len(g.Players[target].Hand) > 0 {
		return "hand", g.Players[target].Hand[0].ID
	}
	return "", ""
}
