package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const ResponseModeSkillJianxiong = "skill_jianxiong"

func (g *Game) damageCardObtainable(card Card) bool {
	if card.ID == "" {
		return false
	}
	return g.cardInDiscard(card.ID)
}

func (g *Game) cardInDiscard(id string) bool {
	for _, c := range g.DiscardPile {
		if c.ID == id {
			return true
		}
	}
	return false
}

func (g *Game) takeCardFromDiscardByID(id string) (Card, bool) {
	for i := len(g.DiscardPile) - 1; i >= 0; i-- {
		if g.DiscardPile[i].ID == id {
			card := g.DiscardPile[i]
			g.DiscardPile = append(g.DiscardPile[:i], g.DiscardPile[i+1:]...)
			return card, true
		}
	}
	return Card{}, false
}

func (g *Game) offerJianxiongWindow(a *DamageAftermath, events *[]GameEvent) bool {
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		TargetIndex:  a.Target,
		SourceIndex:  a.Source,
		ReturnIndex:  a.Resume.ReturnIndex,
		ResponseMode: ResponseModeSkillJianxiong,
		Card:         a.Card,
	}
	g.Message = fmt.Sprintf("%s 可发动【奸雄】获得 %s", g.Players[a.Target].Name, a.Card.Name)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDJianxiong, a.Target, a.Source, g.Message)
	return true
}

func (g *Game) ApplyJianxiong(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillJianxiong || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	a := g.damageAftermath
	if a == nil || !a.OfferJianxiong {
		return ErrWrongPhase
	}
	card, ok := g.takeCardFromDiscardByID(g.Pending.Card.ID)
	if !ok {
		return ErrInvalidCard
	}
	g.Players[seat].Hand = append(g.Players[seat].Hand, card)
	g.SyncCounts()
	a.OfferJianxiong = false
	g.Pending = nil
	msg := fmt.Sprintf("%s 发动【奸雄】，获得 %s", g.Players[seat].Name, card.Name)
	g.appendSkillEvent(events, skill.IDJianxiong, seat, a.Source, msg)
	*events = append(*events, GameEvent{
		Type: "jianxiong_gain", PlayerIndex: seat, Card: &card, Message: msg,
	})
	return g.continueAfterJianxiong(events)
}

func (g *Game) PassJianxiong(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillJianxiong || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	a := g.damageAftermath
	if a != nil {
		a.OfferJianxiong = false
	}
	g.Pending = nil
	return g.continueAfterJianxiong(events)
}

func (g *Game) continueAfterJianxiong(events *[]GameEvent) error {
	if g.advanceDamageAftermath(events) {
		return nil
	}
	return nil
}
