package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const ResponseModeSkillLiuli = "skill_liuli"

func (g *Game) liuliRedirectTarget(victim int) int {
	for i := range g.Players {
		if i == victim || g.Players[i].HP <= 0 {
			continue
		}
		if !g.canAttack(victim, i) {
			continue
		}
		if g.runSkillHooks(nil, skill.HookCall{Kind: skill.HookTargetBlocked, Target: i, CardKind: CardSha}).Bool {
			continue
		}
		return i
	}
	return -1
}

func (g *Game) canOfferLiuli(victim int) bool {
	if !g.hasSkill(victim, SkillLiuli) || len(g.Players[victim].Hand) == 0 {
		return false
	}
	return g.liuliRedirectTarget(victim) >= 0
}

func (g *Game) offerLiuliWindow(victim int, events *[]GameEvent) bool {
	if g.Pending == nil || g.Pending.Card.Kind != CardSha || g.Pending.TargetIndex != victim {
		return false
	}
	if !g.canOfferLiuli(victim) {
		return false
	}
	redirect := g.liuliRedirectTarget(victim)
	g.Pending.ResponseMode = ResponseModeSkillLiuli
	g.Pending.EffectTarget = redirect
	g.Pending.SkillID = skill.IDLiuli
	msg := fmt.Sprintf("%s 可发动【流离】，弃一张牌将【杀】转移给 %s", g.Players[victim].Name, g.Players[redirect].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDLiuli, victim, redirect, msg)
	g.resetTimer()
	return true
}

func (g *Game) ApplyLiuli(seat int, cardID string, redirect int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillLiuli || g.Pending.TargetIndex != seat {
		return ErrNoPendingCombat
	}
	if redirect < 0 {
		redirect = g.Pending.EffectTarget
	}
	if redirect < 0 || redirect >= len(g.Players) || redirect == seat {
		return ErrInvalidTarget
	}
	if !g.canAttack(seat, redirect) {
		return ErrInvalidTarget
	}
	if g.runSkillHooks(nil, skill.HookCall{Kind: skill.HookTargetBlocked, Target: redirect, CardKind: CardSha}).Bool {
		return ErrInvalidTarget
	}
	idx, _, ok := g.findCard(seat, cardID)
	if !ok {
		return ErrInvalidCard
	}

	source := g.Pending.SourceIndex
	damage := g.Pending.Damage
	ignoreArmor := g.Pending.IgnoreArmor
	tieqiPending := g.Pending.TieqiPending
	card := g.Pending.Card
	returnIndex := g.Pending.ReturnIndex

	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.syncCounts()
	g.runCardsDiscardedHooks(seat, "cost", []Card{discarded}, events)

	msg := fmt.Sprintf("%s 发动【流离】，弃 %s，【杀】目标改为 %s", g.Players[seat].Name, discarded.Label, g.Players[redirect].Name)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "skill_liuli",
		PlayerIndex: seat,
		TargetIndex: redirect,
		Card:        &discarded,
		SkillID:     skill.IDLiuli,
		Message:     msg,
	})

	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  redirect,
		ReturnIndex:  returnIndex,
		Card:         card,
		RequiredKind: CardShan,
		Damage:       damage,
		IgnoreArmor:  ignoreArmor,
		TieqiPending: tieqiPending,
	}
	g.initPojunOnShaPending(source, redirect, g.Pending)
	g.Message = fmt.Sprintf("%s 对 %s 使用【杀】，等待出闪", g.Players[source].Name, g.Players[redirect].Name)
	g.resetTimer()
	return g.advanceShaBeforeTargetResponse(events)
}

func (g *Game) PassLiuli(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillLiuli || g.Pending.TargetIndex != seat {
		return ErrNoPendingCombat
	}
	g.Pending.ResponseMode = ""
	g.Pending.SkillID = ""
	msg := fmt.Sprintf("%s 未发动【流离】", g.Players[seat].Name)
	g.Message = msg
	g.resetTimer()
	return g.advanceShaBeforeTargetResponse(events)
}
