package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
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
	if !g.hasSkill(seat, SkillGuose) {
		return ErrWrongPhase
	}
	if target < 0 || target >= len(g.Players) || target == seat {
		return ErrInvalidTarget
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || cardObj.Suit != "D" {
		return ErrInvalidCard
	}

	// 移除手牌
	played := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, played)
	g.syncCounts()
	g.runCardsDiscardedHooks(seat, "cost", []Card{played}, events)

	// 创建转化的乐不思蜀牌
	lebuCard := Card{
		ID:   played.ID,
		Kind: CardLeBu,
		Name: "乐不思蜀",
		Suit: played.Suit,
		Rank: played.Rank,
	}

	// 设置判定区的牌
	g.setJudgeCard(target, lebuCard)
	g.Players[target].SkipPlay = true

	msg := fmt.Sprintf("%s 发动【国色】，将 %s 当【乐不思蜀】对 %s 使用", g.Players[seat].Name, played.Label, g.Players[target].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDGuose, seat, target, msg)

	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &lebuCard,
		Message:     msg,
	})

	// 触发无懈可击响应窗口
	return g.startWuxiekGuoseWindow(seat, target, lebuCard, events)
}

// startWuxiekGuoseWindow 启动国色（乐不思蜀）的无懈可击响应窗口
func (g *Game) startWuxiekGuoseWindow(source, target int, lebu Card, events *[]GameEvent) error {
	g.Phase = PhaseResponse
	// 无懈可击响应从source的下家开始
	responder := g.nextTurnSeat(source)
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  responder,
		ReturnIndex:  source,
		EffectTarget: target,
		Card:         lebu,
		ResponseMode: ResponseModeWuxiekGuose,
	}
	g.Message = fmt.Sprintf("%s 可使用【无懈可击】抵消【国色】的【乐不思蜀】", g.Players[responder].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: source,
		TargetIndex: responder,
		Card:        &lebu,
		Message:     g.Message,
	})
	return nil
}
