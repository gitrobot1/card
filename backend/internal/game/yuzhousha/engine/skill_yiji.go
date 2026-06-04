package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	ResponseModeSkillYijiOffer = "skill_yiji_offer"
	ResponseModeSkillYijiGive  = "skill_yiji_give"
	yijiDrawCount              = 2
	yijiGiveLimit              = 2
)

func (g *Game) offerYijiWindow(a *DamageAftermath, events *[]GameEvent) bool {
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		TargetIndex:  a.Target,
		SourceIndex:  a.Source,
		ReturnIndex:  a.Resume.ReturnIndex,
		ResponseMode: ResponseModeSkillYijiOffer,
		SkillID:      skill.IDYiji,
	}
	g.Message = fmt.Sprintf("%s 可发动【遗计】，摸 %d 张牌后可将至多 %d 张手牌交给其他角色", g.Players[a.Target].Name, yijiDrawCount, yijiGiveLimit)
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDYiji, a.Target, a.Source, g.Message)
	return true
}

func (g *Game) ApplyYiji(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillYijiOffer || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	a := g.damageAftermath
	if a == nil || !a.OfferYiji {
		return ErrWrongPhase
	}
	a.OfferYiji = false
	g.drawCards(seat, yijiDrawCount, events)
	msg := fmt.Sprintf("%s 发动【遗计】，摸了 %d 张牌", g.Players[seat].Name, yijiDrawCount)
	g.appendSkillEvent(events, skill.IDYiji, seat, a.Source, msg)

	g.Pending.ResponseMode = ResponseModeSkillYijiGive
	g.Pending.YijiGiveRemaining = yijiGiveLimit
	g.Pending.EffectTarget = g.opponentOf(seat)
	g.Message = fmt.Sprintf("%s 可将至多 %d 张手牌交给其他角色", g.Players[seat].Name, yijiGiveLimit)
	g.resetTimer()
	return nil
}

func (g *Game) YijiGiveCards(seat, target int, cardIDs []string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillYijiGive || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	if len(cardIDs) == 0 {
		return g.finishYijiGive(seat, events)
	}
	if target < 0 || target >= len(g.Players) || target == seat {
		return ErrInvalidTarget
	}
	remaining := g.Pending.YijiGiveRemaining
	if remaining <= 0 || len(cardIDs) > remaining {
		return ErrInvalidCard
	}
	given := make([]Card, 0, len(cardIDs))
	for _, id := range cardIDs {
		idx, _, ok := g.findCard(seat, id)
		if !ok {
			return ErrInvalidCard
		}
		given = append(given, g.removeHandCard(seat, idx, events))
	}
	g.Players[target].Hand = append(g.Players[target].Hand, given...)
	g.syncCounts()
	g.Pending.YijiGiveRemaining -= len(given)
	msg := fmt.Sprintf("%s 发动【遗计】，交给 %s %d 张牌", g.Players[seat].Name, g.Players[target].Name, len(given))
	g.appendSkillEvent(events, skill.IDYiji, seat, target, msg)
	for i := range given {
		c := given[i]
		*events = append(*events, GameEvent{
			Type:        "skill_give_card",
			PlayerIndex: seat,
			TargetIndex: target,
			Card:        &c,
			SkillID:     skill.IDYiji,
			Message:     fmt.Sprintf("给出 %s", c.Label),
		})
	}
	if g.Pending.YijiGiveRemaining > 0 && len(g.Players[seat].Hand) > 0 {
		g.Message = fmt.Sprintf("%s 还可给出 %d 张牌", g.Players[seat].Name, g.Pending.YijiGiveRemaining)
		g.resetTimer()
		return nil
	}
	return g.finishYijiGive(seat, events)
}

func (g *Game) PassYijiOffer(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillYijiOffer || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	a := g.damageAftermath
	if a != nil {
		a.OfferYiji = false
	}
	g.Pending = nil
	return g.continueAfterYiji(events)
}

func (g *Game) PassYijiGive(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillYijiGive || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	return g.finishYijiGive(seat, events)
}

func (g *Game) finishYijiGive(seat int, events *[]GameEvent) error {
	if g.Pending != nil {
		g.Pending.YijiGiveRemaining = 0
	}
	g.Pending = nil
	return g.continueAfterYiji(events)
}

func (g *Game) continueAfterYiji(events *[]GameEvent) error {
	if g.advanceDamageAftermath(events) {
		return nil
	}
	return nil
}
