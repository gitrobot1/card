package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) placeBingliang(source, target int, card Card, events *[]GameEvent) error {
	if !g.canBingliangTarget(source, target) {
		return ErrInvalidTarget
	}
	g.setJudgeCard(target, card)
	g.Players[target].SkipDraw = true
	g.Message = fmt.Sprintf("%s 被置入【兵粮寸断】", g.Players[target].Name)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: source,
		TargetIndex: target,
		Card:        &card,
		Message:     g.Message,
	})
	return nil
}

func (g *Game) placeShandian(source int, card Card, events *[]GameEvent) error {
	g.setJudgeCard(source, card)
	g.Message = fmt.Sprintf("%s 对自己使用【闪电】", g.Players[source].Name)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: source,
		TargetIndex: source,
		Card:        &card,
		Message:     g.Message,
	})
	return nil
}

func (g *Game) resolveWugu(source int, events *[]GameEvent) error {
	count := len(g.Players)
	if len(g.DrawPile) < count {
		count = len(g.DrawPile)
	}
	if count == 0 {
		g.Message = fmt.Sprintf("%s 使用【五谷丰登】，牌堆不足", g.Players[source].Name)
		return nil
	}
	revealed := make([]Card, 0, count)
	for i := 0; i < count; i++ {
		c := g.DrawPile[0]
		g.DrawPile = g.DrawPile[1:]
		revealed = append(revealed, c)
	}
	g.Message = fmt.Sprintf("%s 使用【五谷丰登】，亮出 %d 张牌", g.Players[source].Name, len(revealed))
	*events = append(*events, GameEvent{
		Type:        "wugu_reveal",
		PlayerIndex: source,
		Message:     g.Message,
		Amount:      len(revealed),
	})
	for i := range revealed {
		c := revealed[i]
		*events = append(*events, GameEvent{
			Type:        "wugu_show",
			PlayerIndex: source,
			Card:        &c,
			Message:     fmt.Sprintf("亮出 %s", c.Label),
		})
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   source,
		TargetIndex:   source,
		ReturnIndex:   source,
		Card:          Card{Kind: CardWuGu, Name: "五谷丰登"},
		ResponseMode:  ResponseModeWuguPick,
		RevealedCards: revealed,
		WuguPickSeat:  source,
	}
	g.Message = fmt.Sprintf("%s 请选择【五谷丰登】中的一张牌", g.Players[source].Name)
	g.resetTimer()
	return nil
}

func (g *Game) pickWuguCard(seat int, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeWuguPick {
		return ErrWrongPhase
	}
	if seat != g.Pending.WuguPickSeat {
		return ErrNotYourTurn
	}
	idx := -1
	for i, c := range g.Pending.RevealedCards {
		if c.ID == cardID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrInvalidCard
	}
	picked := g.Pending.RevealedCards[idx]
	g.Pending.RevealedCards = append(g.Pending.RevealedCards[:idx], g.Pending.RevealedCards[idx+1:]...)
	g.Players[seat].Hand = append(g.Players[seat].Hand, picked)
	g.syncCounts()
	*events = append(*events, GameEvent{
		Type:        "wugu_pick",
		PlayerIndex: seat,
		Card:        &picked,
		Message:     fmt.Sprintf("%s 获得 %s", g.Players[seat].Name, picked.Label),
	})
	return g.advanceWuguPick(events)
}

func (g *Game) advanceWuguPick(events *[]GameEvent) error {
	pending := g.Pending
	if pending == nil || pending.ResponseMode != ResponseModeWuguPick {
		return nil
	}
	if len(pending.RevealedCards) == 0 {
		return g.finishWugu(pending.SourceIndex, events)
	}
	next := g.nextWuguPicker(pending.WuguPickSeat, pending.SourceIndex)
	pending.WuguPickSeat = next
	g.Message = fmt.Sprintf("%s 请选择【五谷丰登】中的一张牌", g.Players[next].Name)
	g.resetTimer()
	return nil
}

func (g *Game) nextWuguPicker(current, source int) int {
	for i := 1; i <= len(g.Players); i++ {
		seat := (current + i) % len(g.Players)
		if seat != source || len(g.Players) == 1 {
			return seat
		}
	}
	return g.opponentOf(current)
}

func (g *Game) finishWugu(source int, events *[]GameEvent) error {
	if g.Pending != nil && len(g.Pending.RevealedCards) > 0 {
		g.DiscardPile = append(g.DiscardPile, g.Pending.RevealedCards...)
	}
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
	g.notifyInstantTrickUsed(source, CardWuGu, events)
	g.resetTimer()
	return nil
}

func (g *Game) processLightningAtTurnStart(seat int, events *[]GameEvent) bool {
	if !g.Players[seat].hasJudgeKind(CardShanDian) {
		return false
	}
	g.startWuxiekShandianWindow(seat, events)
	return true
}

func (g *Game) startWuxiekShandianWindow(seat int, events *[]GameEvent) {
	card := *g.Players[seat].judgeCardByKind(CardShanDian)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  g.opponentOf(seat),
		TargetIndex:  seat,
		ReturnIndex:  seat,
		EffectTarget: seat,
		Card:         card,
		ResponseMode: ResponseModeWuxiekShandian,
	}
	g.Message = fmt.Sprintf("%s 可对【闪电】使用【无懈可击】（判定前）", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: g.opponentOf(seat),
		TargetIndex: seat,
		Card:        &card,
		Message:     g.Message,
	})
}

func (g *Game) startWuxiekBingliangWindow(seat int, events *[]GameEvent) {
	card := *g.Players[seat].judgeCardByKind(CardBingLiang)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  g.opponentOf(seat),
		TargetIndex:  seat,
		ReturnIndex:  seat,
		EffectTarget: seat,
		Card:         card,
		ResponseMode: ResponseModeWuxiekBingliang,
	}
	g.Message = fmt.Sprintf("%s 可对【兵粮寸断】使用【无懈可击】（判定前）", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: g.opponentOf(seat),
		TargetIndex: seat,
		Card:        &card,
		Message:     g.Message,
	})
}

func (g *Game) applyBingliangSkipDraw(seat int, events *[]GameEvent) {
	p := &g.Players[seat]
	p.SkipDraw = false
	if card, ok := g.removeJudgeByKind(seat, CardBingLiang); ok {
		g.DiscardPile = append(g.DiscardPile, card)
	}
	*events = append(*events, GameEvent{
		Type:        "bingliang_skip",
		PlayerIndex: seat,
		Message:     fmt.Sprintf("%s 受到【兵粮寸断】，跳过摸牌", p.Name),
	})
}

func (g *Game) resolveShandianJudge(seat int, events *[]GameEvent) error {
	card, ok := g.removeJudgeByKind(seat, CardShanDian)
	if !ok {
		return nil
	}
	g.Pending = &PendingCombat{
		EffectTarget: seat,
		Card:         card,
	}
	if err := g.startJudge(seat, skill.JudgeShandian, guicaiResumeShandian, events); err != nil {
		g.setJudgeCard(seat, card)
		g.Pending = nil
		return err
	}
	return nil
}

func boolAmount(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (g *Game) continueTurnAfterJudge(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	return g.resumeBeginTurnAfterLightning(seat, events)
}

func (g *Game) resumeBeginTurnAfterLightning(seat int, events *[]GameEvent) error {
	if g.Players[seat].SkipDraw {
		if g.Players[seat].hasJudgeKind(CardBingLiang) {
			g.startWuxiekBingliangWindow(seat, events)
			return nil
		}
		g.applyBingliangSkipDraw(seat, events)
	} else if g.shouldOfferDrawPhaseChoice(seat) {
		g.offerDrawPhaseChoice(seat, events)
		return nil
	} else {
		g.drawCards(seat, g.drawCountFor(seat), events)
	}
	if g.IsFinished() {
		return nil
	}
	if g.Players[seat].SkipPlay {
		if g.Players[seat].hasJudgeKind(CardLeBu) {
			g.startWuxiekLebuJudgeWindow(seat, events)
			return nil
		}
		g.applyLebuSkipDirect(seat, events)
		return nil
	}
	g.TurnStep = StepPlay
	g.resetTimer()
	return nil
}
