package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterFanjianUsed         = "fanjian_used_play"
	ResponseModeSkillFanjianSuit = "skill_fanjian_suit"
)

func (g *Game) ActivateFanjian(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillFanjian) || g.getSkillCounter(seat, counterFanjianUsed) > 0 {
		return ErrWrongPhase
	}
	target := g.opponentOf(seat)
	idx, _, ok := g.findCard(seat, cardID)
	if !ok {
		return ErrInvalidCard
	}
	given := g.removeHandCard(seat, idx, events)
	g.syncCounts()
	g.setSkillCounter(seat, counterFanjianUsed, 1)

	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  seat,
		TargetIndex:  target,
		ReturnIndex:  seat,
		EffectTarget: target,
		Card:         given,
		ResponseMode: ResponseModeSkillFanjianSuit,
		SkillID:      skill.IDFanjian,
	}
	msg := fmt.Sprintf("%s 发动【反间】，请 %s 选择一种花色", g.Players[seat].Name, g.Players[target].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDFanjian, seat, target, msg)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "skill_fanjian_offer",
		PlayerIndex: seat,
		TargetIndex: target,
		Message:     msg,
	})
	return nil
}

func (g *Game) ResolveFanjianSuit(seat int, chosenSuit string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillFanjianSuit {
		return ErrNoPendingCombat
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	if !isValidSuitChoice(chosenSuit) {
		return ErrInvalidTarget
	}

	pending := *g.Pending
	g.Pending = nil
	card := pending.Card
	g.DiscardPile = append(g.DiscardPile, card)
	g.syncCounts()

	match := card.Suit == chosenSuit
	revealMsg := fmt.Sprintf("%s 选择【%s】，展示 %s", g.Players[seat].Name, suitLabel(chosenSuit), card.Label)
	*events = append(*events, GameEvent{
		Type:        "skill_fanjian_reveal",
		PlayerIndex: pending.SourceIndex,
		TargetIndex: seat,
		Card:        &card,
		Message:     revealMsg,
	})

	if match {
		damageMsg := fmt.Sprintf("%s 猜中花色，受到 1 点伤害", g.Players[seat].Name)
		g.Message = damageMsg
		*events = append(*events, GameEvent{
			Type:        "skill_fanjian_hit",
			PlayerIndex: pending.SourceIndex,
			TargetIndex: seat,
			Card:        &card,
			Damage:      1,
			SkillID:     skill.IDFanjian,
			Message:     damageMsg,
		})
		resume := DamageResume{
			Mode:        damageResumeFanjian,
			ReturnIndex: pending.ReturnIndex,
		}
		if err := g.finalizeDamageHit(pending.SourceIndex, seat, 1, card, resume, events); err != nil {
			return err
		}
		return nil
	}

	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = pending.ReturnIndex
	g.Message = fmt.Sprintf("%s 未猜中，【反间】结束", g.Players[seat].Name)
	g.resetTimer()
	return nil
}

func isValidSuitChoice(s string) bool {
	switch s {
	case "H", "D", "S", "C":
		return true
	default:
		return false
	}
}

func suitLabel(s string) string {
	switch s {
	case "H":
		return "红桃"
	case "D":
		return "方块"
	case "S":
		return "黑桃"
	case "C":
		return "梅花"
	default:
		return s
	}
}

func (g *Game) aiPickFanjianSuit() string {
	// 均匀猜花色；AI 无完美信息。
	suits := []string{"H", "D", "S", "C"}
	return suits[len(g.DiscardPile)%len(suits)]
}

const damageResumeFanjian = "fanjian"

func (g *Game) resumeAfterFanjianDamage(resume DamageResume, events *[]GameEvent) {
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = resume.ReturnIndex
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[resume.ReturnIndex].Name)
	g.resetTimer()
}
