package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const counterShaInPlayPhase = skill.CounterShaInPlayPhase

func (g *Game) markShaInPlayPhase(seat int) {
	if seat < 0 || seat >= len(g.Players) {
		return
	}
	if g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return
	}
	g.setSkillCounter(seat, counterShaInPlayPhase, 1)
}

func (g *Game) kejiSkipsDiscard(seat int) bool {
	return g.skipsDiscardViaHooks(seat)
}

func (g *Game) finishPlayWithKejiOrDiscard(seat int, events *[]GameEvent) error {
	p := &g.Players[seat]
	cap := g.handRetainLimit(seat)
	if len(p.Hand) <= cap {
		return g.endTurn(events)
	}
	if g.kejiSkipsDiscard(seat) {
		msg := fmt.Sprintf("%s 发动【克己】，跳过弃牌阶段", p.Name)
		g.Message = msg
		g.appendSkillEvent(events, skill.IDKeji, seat, seat, msg)
		*events = append(*events, GameEvent{
			Type:        "skill_keji",
			PlayerIndex: seat,
			SkillID:     skill.IDKeji,
			Message:     msg,
		})
		return g.endTurn(events)
	}
	need := len(p.Hand) - cap
	g.TurnStep = StepDiscard
	g.Message = fmt.Sprintf("%s 请一次选择 %d 张牌弃到牌桌", p.Name, need)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "discard_phase",
		PlayerIndex: seat,
		Amount:      need,
		Message:     g.Message,
	})
	return nil
}
