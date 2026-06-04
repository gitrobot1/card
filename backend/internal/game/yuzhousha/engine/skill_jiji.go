package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) canUseJijiHeal(seat int, card Card) bool {
	if g.wanshaBlocksPeachUse(seat) {
		return false
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn == seat {
		return false
	}
	if g.Pending != nil {
		return false
	}
	p := &g.Players[seat]
	if p.HP >= p.MaxHP || !g.hasSkill(seat, SkillJiji) {
		return false
	}
	if card.Kind == CardTao {
		return false
	}
	return isRedSuit(card.Suit)
}

func (g *Game) playJijiHeal(seat int, cardID string, events *[]GameEvent) error {
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || !g.canUseJijiHeal(seat, cardObj) {
		return ErrInvalidCard
	}
	p := &g.Players[seat]
	played := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, played)
	p.HP++
	msg := fmt.Sprintf("%s 发动【急救】，将 %s 当【桃】使用，体力 %d/%d", p.Name, played.Label, p.HP, p.MaxHP)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDJiji, seat, seat, msg)
	*events = append(*events, GameEvent{
		Type:        "skill_jiji",
		PlayerIndex: seat,
		SkillID:     skill.IDJiji,
		Card:        &played,
		Heal:        1,
		Message:     msg,
	})
	return nil
}

func (g *Game) tryAIJijiHeal(events *[]GameEvent) bool {
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.Pending != nil {
		return false
	}
	for seat := range g.Players {
		if seat == g.CurrentTurn || !g.Players[seat].IsAI {
			continue
		}
		p := &g.Players[seat]
		if p.HP >= p.MaxHP || !g.hasSkill(seat, SkillJiji) {
			continue
		}
		for _, c := range p.Hand {
			if g.canUseJijiHeal(seat, c) {
				_ = g.playJijiHeal(seat, c.ID, events)
				return true
			}
		}
	}
	return false
}
