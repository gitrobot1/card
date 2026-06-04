package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterYinghunUsed           = "yinghun_used_prepare"
	ResponseModeSkillYinghun     = "skill_yinghun"
	ResponseModeSkillYinghunDiscard = "skill_yinghun_discard"
	YinghunOptionDrawBoth        = "draw_both"
	YinghunOptionDrawTwoDiscard  = "draw_two_discard"
)

func (g *Game) ActivateYinghun(seat int, target int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillYinghun) || g.getSkillCounter(seat, counterYinghunUsed) > 0 {
		return ErrWrongPhase
	}
	if target < 0 {
		target = g.opponentOf(seat)
	}
	if target == seat || target < 0 || target >= len(g.Players) {
		return ErrInvalidTarget
	}
	g.setSkillCounter(seat, counterYinghunUsed, 1)

	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  seat,
		TargetIndex:  target,
		ReturnIndex:  seat,
		EffectTarget: seat,
		ResponseMode: ResponseModeSkillYinghun,
		SkillID:      skill.IDYinghun,
	}
	msg := fmt.Sprintf("%s 发动【英魂】，请 %s 选择一项", g.Players[seat].Name, g.Players[target].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDYinghun, seat, target, msg)
	g.resetTimer()
	return nil
}

func (g *Game) ResolveYinghunChoice(target int, option string, events *[]GameEvent) error {
	return g.resolveYinghunChoice(target, option, "", events)
}

func (g *Game) YinghunDiscard(target int, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillYinghunDiscard || g.Pending.TargetIndex != target {
		return ErrWrongPhase
	}
	source := g.Pending.SourceIndex
	pending := *g.Pending
	g.Pending = nil

	idx, _, ok := g.findCard(target, cardID)
	if !ok {
		return ErrInvalidCard
	}
	discarded := g.removeHandCard(target, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.syncCounts()
	g.runCardsDiscardedHooks(target, "yinghun", []Card{discarded}, events)

	g.drawCards(source, 2, events)
	msg := fmt.Sprintf("%s 选择令 %s 摸两张牌并弃置手牌", g.Players[target].Name, g.Players[source].Name)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "skill_yinghun",
		PlayerIndex: source,
		TargetIndex: target,
		SkillID:     skill.IDYinghun,
		Message:     msg,
	})
	return g.finishYinghun(source, pending.ReturnIndex, events)
}

func (g *Game) resolveYinghunChoice(target int, option, discardCardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillYinghun {
		return ErrNoPendingCombat
	}
	if target != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	source := g.Pending.SourceIndex
	returnIndex := g.Pending.ReturnIndex
	g.Pending = nil

	switch option {
	case YinghunOptionDrawBoth:
		g.drawCards(target, 1, events)
		g.drawCards(source, 1, events)
		msg := fmt.Sprintf("%s 选择双方各摸一张牌", g.Players[target].Name)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "skill_yinghun",
			PlayerIndex: source,
			TargetIndex: target,
			SkillID:     skill.IDYinghun,
			Message:     msg,
		})
		return g.finishYinghun(source, returnIndex, events)
	case YinghunOptionDrawTwoDiscard:
		if len(g.Players[target].Hand) > 0 {
			if discardCardID == "" {
				g.Phase = PhaseResponse
				g.Pending = &PendingCombat{
					SourceIndex:  source,
					TargetIndex:  target,
					ReturnIndex:  returnIndex,
					ResponseMode: ResponseModeSkillYinghunDiscard,
					SkillID:      skill.IDYinghun,
				}
				g.Message = fmt.Sprintf("%s 请选择一张手牌弃置（【英魂】）", g.Players[target].Name)
				g.resetTimer()
				return nil
			}
			idx, _, ok := g.findCard(target, discardCardID)
			if !ok {
				return ErrInvalidCard
			}
			discarded := g.removeHandCard(target, idx, events)
			g.DiscardPile = append(g.DiscardPile, discarded)
			g.syncCounts()
			g.runCardsDiscardedHooks(target, "yinghun", []Card{discarded}, events)
		}
		g.drawCards(source, 2, events)
		msg := fmt.Sprintf("%s 选择令 %s 摸两张牌", g.Players[target].Name, g.Players[source].Name)
		if discardCardID != "" {
			msg += "并弃置一张手牌"
		}
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "skill_yinghun",
			PlayerIndex: source,
			TargetIndex: target,
			SkillID:     skill.IDYinghun,
			Message:     msg,
		})
		return g.finishYinghun(source, returnIndex, events)
	default:
		return ErrInvalidTarget
	}
}

func (g *Game) finishYinghun(source, returnIndex int, events *[]GameEvent) error {
	_ = source
	g.Phase = PhasePlaying
	g.TurnStep = StepPrepare
	g.CurrentTurn = returnIndex
	return g.continueAfterPrepare(returnIndex, events)
}

func (g *Game) aiPickYinghunOption(target, source int) string {
	_ = source
	if len(g.Players[target].Hand) == 0 {
		return YinghunOptionDrawTwoDiscard
	}
	return YinghunOptionDrawBoth
}
