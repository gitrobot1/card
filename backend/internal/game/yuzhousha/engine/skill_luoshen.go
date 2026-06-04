package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) revealDeckTop(events *[]GameEvent, seat int) (Card, bool) {
	if len(g.DrawPile) == 0 {
		g.refillDrawPile()
	}
	if len(g.DrawPile) == 0 {
		return Card{}, false
	}
	card := g.DrawPile[0]
	g.DrawPile = g.DrawPile[1:]
	g.syncCounts()
	*events = append(*events, GameEvent{
		Type:        "judge_flip",
		PlayerIndex: seat,
		Card:        &card,
		Message:     fmt.Sprintf("判定牌 %s", card.Label),
	})
	return card, true
}

func (g *Game) StartLuoshen(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillLuoshen) || len(g.DrawPile) == 0 {
		return ErrWrongPhase
	}
	g.appendSkillEvent(events, skill.IDLuoshen, seat, seat, fmt.Sprintf("%s 发动【洛神】", g.Players[seat].Name))
	card, ok := g.revealDeckTop(events, seat)
	if !ok {
		return ErrInvalidCard
	}
	return g.afterJudgeFlip(seat, skill.JudgeLuoshen, guicaiResumeLuoshen, card, events)
}

func (g *Game) applyLuoshenJudgeResult(seat int, judgeCard Card, events *[]GameEvent) error {
	if isRedSuit(judgeCard.Suit) {
		g.DiscardPile = append(g.DiscardPile, judgeCard)
		msg := fmt.Sprintf("%s 【洛神】判定 %s 为红色，结束", g.Players[seat].Name, judgeCard.Label)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "luoshen_stop",
			PlayerIndex: seat,
			Card:        &judgeCard,
			SkillID:     skill.IDLuoshen,
			Message:     msg,
		})
		return g.continueAfterPrepare(seat, events)
	}
	g.Players[seat].Hand = append(g.Players[seat].Hand, judgeCard)
	g.syncCounts()
	msg := fmt.Sprintf("%s 【洛神】判定 %s 为黑色，获得该牌", g.Players[seat].Name, judgeCard.Label)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "luoshen_gain",
		PlayerIndex: seat,
		Card:        &judgeCard,
		SkillID:     skill.IDLuoshen,
		Message:     msg,
	})
	g.Phase = PhasePlaying
	g.TurnStep = StepPrepare
	g.resetTimer()
	return nil
}
