package engine

import (
	"fmt"
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
		g.syncCounts()
		return nil
	}
	// 初始化已选牌记录
	g.wuguPicked = make(map[int]bool)
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
	// 第一个选牌者（使用者自己），走标准无懈窗口
	g.startWuguPickFor(source, source, revealed, events)
	return nil
}

// startWuguPickFor 对 picker 开启选牌锦囊的无懈窗口，通过后进入选牌
// 五谷丰登本身不可无懈，但每个玩家的"单体选牌锦囊"可以被无懈抵消
func (g *Game) startWuguPickFor(source, picker int, revealed []Card, events *[]GameEvent) {
	if len(revealed) == 0 {
		g.finishWugu(source, events)
		return
	}
	revealedCopy := make([]Card, len(revealed))
	copy(revealedCopy, revealed)
	trick := Card{Kind: CardWuGu, Name: "五谷丰登"}
	spec := PlayTarget{SeatIndex: picker}
	// 复用标准无懈窗口：任何人可以对 picker 的"选牌锦囊"出无懈可击
	// responder 是 picker 的下家（五谷使用者 source 可以在反无懈链中出无懈）
	g.Phase = PhaseResponse
	allQueue := g.createResponseQueue((picker + 1) % len(g.Players))
	queue := make([]int, 0, len(allQueue))
	for _, s := range allQueue {
		// 排除 picker 自己（不能无懈自己对自己的选牌锦囊）
		if s != picker {
			queue = append(queue, s)
		}
	}
	g.Pending = &PendingCombat{
		SourceIndex:   source,
		TargetIndex:   -1,
		ReturnIndex:   source,
		EffectTarget:  picker,
		Card:          trick,
		ResponseMode:  ResponseModeWuxiekTrick,
		TargetZone:    spec.Zone,
		TargetCardID:  spec.CardID,
		ResponseQueue: queue,
		ResponseIndex: 0,
		WuxiekChain:   nil,
		RevealedCards: revealedCopy,
		WuguPickSeat:  picker,
	}
	g.advanceToNextWuxiekResponder(events)
	// 提示由 advanceToNextWuxiekResponder 中设置
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: source,
		TargetIndex: picker,
		Card:        &trick,
		Message:     g.Message,
	})
}

// wuguPickPass 无懈通过，让 picker 选牌
func (g *Game) wuguPickPass(picker int, revealed []Card, source int, events *[]GameEvent) {
	if len(revealed) == 0 {
		g.finishWugu(source, events)
		return
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   source,
		TargetIndex:   source,
		ReturnIndex:   source,
		Card:          Card{Kind: CardWuGu, Name: "五谷丰登"},
		ResponseMode:  ResponseModeWuguPick,
		RevealedCards: revealed,
		WuguPickSeat:  picker,
	}
	if picker == g.HumanPlayer {
		g.Message = "请选择【五谷丰登】中的一张牌"
	} else {
		g.Message = fmt.Sprintf("等待 %s 选牌...", g.Players[picker].Name)
	}
	FillPendingRoles(g.Pending)
	g.resetTimer()
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
	return g.pickWuguCardByIndex(seat, idx, events)
}

func (g *Game) pickWuguCardByIndex(seat, idx int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeWuguPick {
		return ErrWrongPhase
	}
	if idx < 0 || idx >= len(g.Pending.RevealedCards) {
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

func (g *Game) autoPickWuguCard(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeWuguPick {
		return ErrWrongPhase
	}
	if seat != g.Pending.WuguPickSeat {
		return ErrNotYourTurn
	}
	Logf("autoPickWuguCard: seat=%d, remaining=%d", seat, len(g.Pending.RevealedCards))
	if len(g.Pending.RevealedCards) == 0 {
		return g.finishWugu(g.Pending.SourceIndex, events)
	}
	return g.pickWuguCardByIndex(seat, 0, events)
}

func (g *Game) advanceWuguPick(events *[]GameEvent) error {
	pending := g.Pending
	if pending == nil || pending.ResponseMode != ResponseModeWuguPick {
		return nil
	}
	if len(pending.RevealedCards) == 0 {
		return g.finishWugu(pending.SourceIndex, events)
	}
	// 记录当前选牌者已选过
	g.wuguPicked[pending.WuguPickSeat] = true

	next := g.nextWuguPicker(pending.WuguPickSeat, pending.SourceIndex)
	if next == pending.SourceIndex {
		Logf("advanceWuguPick: back to source=%d, finishing wugu", pending.SourceIndex)
		return g.finishWugu(pending.SourceIndex, events)
	}
	Logf("advanceWuguPick: %d -> %d (enter wuxiek for next picker)", pending.WuguPickSeat, next)
	// 进入下一个人的选牌无懈窗口
	g.startWuguPickFor(pending.SourceIndex, next, pending.RevealedCards, events)
	return nil
}

func (g *Game) nextWuguPicker(current, source int) int {
	n := len(g.Players)
	for i := 1; i <= n; i++ {
		seat := (current + i) % n
		if g.Players[seat].HP > 0 && !g.wuguPicked[seat] {
			return seat
		}
	}
	// 所有人都选过/被跳过，返回 source 表示结束
	return source
}

func (g *Game) finishWugu(source int, events *[]GameEvent) error {
	if g.Pending != nil && len(g.Pending.RevealedCards) > 0 {
		g.DiscardPile = append(g.DiscardPile, g.Pending.RevealedCards...)
	}
	g.Pending = nil
	g.wuguPicked = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
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
	jc := g.Players[seat].judgeCardByKind(CardShanDian)
	if jc == nil {
		return
	}
	g.startJudgeWuxiekWindow(seat, *jc, events)
}

func (g *Game) startWuxiekBingliangWindow(seat int, events *[]GameEvent) {
	jc := g.Players[seat].judgeCardByKind(CardBingLiang)
	if jc == nil {
		return
	}
	g.startJudgeWuxiekWindow(seat, *jc, events)
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
