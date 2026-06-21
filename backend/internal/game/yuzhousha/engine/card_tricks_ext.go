package engine

import "fmt"

const ResponseModeHuoGong = "huogong"

func (g *Game) playHuoGong(seat int, card Card, target int, events *[]GameEvent) error {
	if len(g.Players[target].Hand) == 0 {
		return ErrInvalidTarget
	}
	// 展示目标第一张手牌
	shown := g.Players[target].Hand[0]
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   seat,
		TargetIndex:   seat,
		ReturnIndex:   seat,
		EffectTarget:  target,
		Card:          card,
		ResponseMode:  ResponseModeHuoGong,
		RevealedCards: []Card{shown},
	}
	g.Message = fmt.Sprintf("【火攻】：%s 展示 %s（%s），%s 需弃置 %s 花色手牌",
		g.Players[target].Name, shown.Label, suitLabel(shown.Suit),
		g.Players[seat].Name, suitLabel(shown.Suit))
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "huogong_reveal",
		PlayerIndex: target,
		TargetIndex: seat,
		Card:        &shown,
		Message:     g.Message,
	})
	return nil
}

func (g *Game) respondHuoGongDiscard(seat int, cardID string, events *[]GameEvent) error {
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModeHuoGong {
		return ErrNoPendingCombat
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	if len(g.Pending.RevealedCards) == 0 {
		return ErrInvalidCard
	}
	requiredSuit := g.Pending.RevealedCards[0].Suit
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || cardObj.Suit != requiredSuit {
		return ErrInvalidCard
	}
	pending := *g.Pending
	source := pending.SourceIndex
	target := pending.EffectTarget
	card := pending.Card

	// 弃置同花色手牌
	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.runCardsDiscardedHooks(seat, "play", []Card{discarded}, events)
	// 显式发出 discard 事件，让前端播放弃牌飞向牌桌中央的动画
	*events = append(*events, GameEvent{
		Type:        "discard",
		PlayerIndex: seat,
		Card:        &discarded,
		Message:     fmt.Sprintf("%s 弃置 %s", g.Players[seat].Name, discarded.Label),
	})

	// 弃牌成功 → 对目标造成 1 点火焰伤害
	damage := g.adjustDamageAmount(source, target, 1, card, true, false)
	g.applyDamageWithHook(source, target, damage, card, events)
	*events = append(*events, GameEvent{
		Type:        "trick_hit",
		PlayerIndex: source,
		TargetIndex: target,
		Damage:      damage,
		Message:     g.damageMessage(&g.Players[target], card.Name, damage),
	})
	g.Pending = nil
	// 火攻造成的是火焰伤害
	fireCard := card
	fireCard.DamageType = DamageTypeFire
	// 类比南蛮：濒死前设置 Pending，濒死时自动保存，结束后恢复
	if g.isChained(target) {
		chainSeats := make([]int, 0)
		for seat := range g.Players {
			if seat == target || !g.isChained(seat) || g.Players[seat].HP <= 0 {
				continue
			}
			chainSeats = append(chainSeats, seat)
		}
		g.setChained(target, false)
		g.Pending = &PendingCombat{
			SourceIndex:  source,
			TargetIndex:  target,
			EffectTarget: target,
			Card:         fireCard,
			Damage:       damage,
			AoeQueue:     chainSeats,
			ReturnIndex:  source,
			RequiredKind: "tiesuo",
		}
	}
	if g.Players[target].HP <= 0 {
		if g.afterDamageApplied(source, target, damage, fireCard, DamageResume{}, events) {
			return nil
		}
		if g.IsFinished() {
			return nil
		}
	}
	// 未濒死：清理 Pending 并启动传导
	if g.Pending != nil && g.Pending.RequiredKind == "tiesuo" {
		g.Pending = nil
		g.spreadChainedFireDamage(source, target, damage, fireCard, events)
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModeDying {
			return nil
		}
	}
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 弃置同花色牌，【火攻】对 %s 造成 %d 点火焰伤害",
		g.Players[source].Name, g.Players[target].Name, damage)
	g.resetTimer()
	return nil
}

func (g *Game) resolveHuoGongFail(seat int, events *[]GameEvent) error {
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModeHuoGong {
		return ErrNoPendingCombat
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	pending := *g.Pending
	source := pending.SourceIndex
	target := pending.EffectTarget
	// 使用者放弃弃置同花色手牌 → 火攻无效，不造成伤害
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 未弃置同花色手牌，【火攻】对 %s 无效",
		g.Players[source].Name, g.Players[target].Name)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: source,
		TargetIndex: target,
		Message:     g.Message,
	})
	g.resetTimer()
	return nil
}

func (g *Game) playTieSuoRecast(seat int, card Card, events *[]GameEvent) error {
	g.DiscardPile = append(g.DiscardPile, card)
	g.runCardsDiscardedHooks(seat, "play", []Card{card}, events)
	g.drawCards(seat, 1, events)
	g.Message = fmt.Sprintf("%s 重铸【铁索连环】", g.Players[seat].Name)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: seat,
		TargetIndex: seat,
		Card:        &card,
		Message:     g.Message,
	})
	return nil
}

func (g *Game) resolveTieSuoChain(seat, target int, card Card, events *[]GameEvent) error {
	Logf("resolveTieSuoChain: source=%d(%s) target=%d(%s) before_chained=%v",
		seat, g.Players[seat].Name, target, g.Players[target].Name, g.isChained(target))
	g.toggleChained(target)
	state := "横置"
	if !g.isChained(target) {
		state = "重置"
	}
	Logf("resolveTieSuoChain: target=%d(%s) after_chained=%v state=%s",
		target, g.Players[target].Name, g.isChained(target), state)
	g.Message = fmt.Sprintf("%s 对 %s 使用【铁索连环】，%s", g.Players[seat].Name, g.Players[target].Name, state)
	*events = append(*events, GameEvent{
		Type:        "tiesuo_chain",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &card,
		Message:     g.Message,
	})
	return nil
}

// resolveTieSuoAOE 铁索连环AOE入口：类似南蛮入侵/桃园结义，逐人无懈窗口
func (g *Game) resolveTieSuoAOE(source int, spec PlayTarget, trick Card, events *[]GameEvent) error {
	target1 := spec.SeatIndex
	target2 := spec.SecondSeatIndex
	hasSecond := target2 >= 0 && target2 < len(g.Players) && target2 != target1

	// 构建目标队列（去重、保证source在其中时排在前面）
	var queue []int
	addIfNew := func(t int) {
		for _, q := range queue {
			if q == t {
				return
			}
		}
		queue = append(queue, t)
	}
	// 自己优先加入队列
	if target1 == source || (hasSecond && target2 == source) {
		addIfNew(source)
	}
	if target1 != source {
		addIfNew(target1)
	}
	if hasSecond && target2 != source {
		addIfNew(target2)
	}

	if len(queue) == 0 {
		return ErrInvalidTarget
	}

	Logf("resolveTieSuoAOE: source=%d(%s) queue=%v", source, g.Players[source].Name, queue)

	// 通知成为目标
	for _, t := range queue {
		g.notifyBecameTarget(t, source, trick, events)
	}

	// 宣告
	g.Message = fmt.Sprintf("%s 使用【铁索连环】", g.Players[source].Name)
	*events = append(*events, GameEvent{
		Type:        "tiesuo_announce",
		PlayerIndex: source,
		TargetIndex: source,
		Message:     g.Message,
	})

	// 开始第一个目标的无懈窗口
	rest := append([]int(nil), queue[1:]...)
	g.startTieSuoFor(source, queue[0], rest, trick, events)
	return nil
}

// startTieSuoFor 对单个目标发起铁索连环无懈窗口
// 铁索连环：对别人 → 类比决斗（排除source）；对自己 → 类比无中生有（排除source）
// 统一排除使用者(source)
func (g *Game) startTieSuoFor(source, target int, rest []int, trick Card, events *[]GameEvent) {
	// 构建无懈响应队列
	// 当 target != source：target 第一个，然后从 target 下家轮询，排除 source
	// 当 target == source：从 source 下家开始轮询，排除 source
	var queue []int
	if target != source {
		queue = append(queue, target)
	}
	allQueue := g.createResponseQueue((target + 1) % len(g.Players))
	for _, s := range allQueue {
		if s == target {
			continue
		}
		if s == source {
			continue
		}
		queue = append(queue, s)
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   source,
		TargetIndex:   -1,
		ReturnIndex:   source,
		EffectTarget:  target,
		Card:          trick,
		ResponseMode:  ResponseModeWuxiekTrick,
		AoeQueue:      rest,
		ResponseQueue: queue,
		ResponseIndex: 0,
		WuxiekChain:   nil,
	}
	g.advanceToNextWuxiekResponder(events)
	if g.Message == "" {
		if target == source {
			g.Message = fmt.Sprintf("【铁索连环】：%s 对自己使用，是否使用【无懈可击】？", g.Players[source].Name)
		} else {
			g.Message = fmt.Sprintf("【铁索连环】：是否对 %s 使用【无懈可击】？", g.Players[target].Name)
		}
	}
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: source,
		TargetIndex: target,
		Card:        &trick,
		Message:     g.Message,
	})
}

// continueTieSuoAfter 铁索连环：当前目标处理完毕，继续下一个
func (g *Game) continueTieSuoAfter(source int, trick Card, rest []int, events *[]GameEvent) {
	if len(rest) == 0 {
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = source
		g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
		g.resetTimer()
		return
	}
	next := rest[0]
	newRest := append([]int(nil), rest[1:]...)
	g.startTieSuoFor(source, next, newRest, trick, events)
}
