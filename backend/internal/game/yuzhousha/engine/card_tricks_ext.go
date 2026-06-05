package engine

import "fmt"

const ResponseModeHuoGong = "huogong"

func (g *Game) playHuoGong(seat int, card Card, target int, events *[]GameEvent) error {
	if len(g.Players[target].Hand) == 0 {
		return ErrInvalidTarget
	}
	shown := g.Players[target].Hand[0]
	g.notifyBecameTarget(target, seat, card, events)
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
	g.Message = fmt.Sprintf("%s 对 %s 使用【火攻】，展示 %s，%s 需弃置同花色手牌",
		g.Players[seat].Name, g.Players[target].Name, shown.Label, g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "trick_response",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &card,
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
	source := g.Pending.SourceIndex
	target := g.Pending.EffectTarget
	g.Pending = nil

	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.runCardsDiscardedHooks(seat, "play", []Card{discarded}, events)
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 弃置同花色牌，【火攻】对 %s 无效", g.Players[seat].Name, g.Players[target].Name)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: source,
		TargetIndex: target,
		Message:     g.Message,
	})
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
	g.Pending = nil
	source := pending.SourceIndex
	target := pending.EffectTarget
	card := pending.Card
	damage := g.adjustDamageAmount(source, target, 1, card, true, false)
	g.applyDamage(source, target, damage, card, events)
	*events = append(*events, GameEvent{
		Type:        "trick_hit",
		PlayerIndex: source,
		TargetIndex: target,
		Damage:      damage,
		Message:     g.damageMessage(&g.Players[target], card.Name, damage),
	})
	if g.Players[target].HP <= 0 {
		if g.afterDamageApplied(source, target, damage, card, DamageResume{}, events) {
			return nil
		}
	}
	g.spreadChainedFireDamage(source, target, damage, card, events)
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
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
	g.toggleChained(target)
	state := "横置"
	if !g.isChained(target) {
		state = "重置"
	}
	g.Message = fmt.Sprintf("%s 对 %s 使用【铁索连环】，%s", g.Players[seat].Name, g.Players[target].Name, state)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &card,
		Message:     g.Message,
	})
	return nil
}
