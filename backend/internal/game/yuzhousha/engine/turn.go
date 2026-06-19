package engine

import "fmt"

func (g *Game) drawCards(seat, count int, events *[]GameEvent) {
	p := &g.Players[seat]
	for i := 0; i < count; i++ {
		if len(g.DrawPile) == 0 {
			g.refillDrawPile()
		}
		if len(g.DrawPile) == 0 {
			break
		}
		c := g.DrawPile[0]
		g.DrawPile = g.DrawPile[1:]
		p.Hand = append(p.Hand, c)
		*events = append(*events, GameEvent{
			Type:        "draw",
			PlayerIndex: seat,
			Card:        &c,
			Message:     fmt.Sprintf("%s 摸牌", p.Name),
		})
	}
	g.syncCounts()
}

func (g *Game) refillDrawPile() {
	if len(g.DiscardPile) <= 1 {
		return
	}
	rest := append([]Card(nil), g.DiscardPile[:len(g.DiscardPile)-1]...)
	g.DrawPile = g.shuffleCards(rest)
	g.DiscardPile = g.DiscardPile[len(g.DiscardPile)-1:]
}

func (g *Game) beginTurn(events *[]GameEvent) {
	if events == nil {
		events = &[]GameEvent{}
	}
	seat := g.CurrentTurn
	
	// 重置回合状态
	g.Players[seat].ShaUsedThisTurn = false
	g.Players[seat].ShaExtraUsedThisTurn = false
	g.Players[seat].Drunk = false
	g.setSkillCounter(seat, counterShaInPlayPhase, 0)
	g.resetPlayPhaseSkillCounters(seat)
	
	// 1. 回合开始阶段
	g.beginStartPhase(seat, events)
}
func (g *Game) EndPlay(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase == PhaseResponse {
		return ErrPendingCombat
	}
	if g.Phase != PhasePlaying || g.CurrentTurn != seat {
		return ErrNotYourTurn
	}
	if g.TurnStep != StepPlay {
		return ErrWrongPhase
	}
	return g.finishPlayWithKejiOrDiscard(seat, events)
}

func (g *Game) DiscardCards(seat int, cardIDs []string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepDiscard || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	p := &g.Players[seat]
	cap := g.handRetainLimit(seat)
	need := len(p.Hand) - cap
	if need <= 0 {
		return ErrWrongPhase
	}
	if len(cardIDs) != need {
		return ErrInvalidDiscardCount
	}

	seen := make(map[string]struct{}, len(cardIDs))
	for _, id := range cardIDs {
		if id == "" {
			return ErrInvalidCard
		}
		if _, dup := seen[id]; dup {
			return ErrInvalidCard
		}
		seen[id] = struct{}{}
		if _, _, ok := g.findCard(seat, id); !ok {
			return ErrInvalidCard
		}
	}

	discarded := make([]Card, 0, len(cardIDs))
	for _, id := range cardIDs {
		idx, _, ok := g.findCard(seat, id)
		if !ok {
			return ErrInvalidCard
		}
		played := g.removeHandCard(seat, idx, events)
		g.DiscardPile = append(g.DiscardPile, played)
		discarded = append(discarded, played)
	}
	g.syncCounts()

	discardMsg := fmt.Sprintf("%s 弃牌", p.Name)
	for i := range discarded {
		*events = append(*events, GameEvent{
			Type:        "discard",
			PlayerIndex: seat,
			Card:        &discarded[i],
			Message:     discardMsg,
			Amount:      0,
		})
	}
	g.runCardsDiscardedHooks(seat, "discard_phase", discarded, events)
	g.Message = discardMsg
	return g.endTurn(events)
}

func (g *Game) autoDiscard(seat int, events *[]GameEvent) {
	p := &g.Players[seat]
	cap := g.handRetainLimit(seat)
	need := len(p.Hand) - cap
	if need <= 0 {
		return
	}
	discarded := make([]Card, 0, need)
	for len(p.Hand) > cap {
		c := p.Hand[len(p.Hand)-1]
		p.Hand = p.Hand[:len(p.Hand)-1]
		g.DiscardPile = append(g.DiscardPile, c)
		discarded = append(discarded, c)
	}
	g.syncCounts()
	discardMsg := fmt.Sprintf("%s 弃牌", p.Name)
	for i := range discarded {
		*events = append(*events, GameEvent{
			Type:        "discard",
			PlayerIndex: seat,
			Card:        &discarded[i],
			Message:     discardMsg,
			Amount:      0,
		})
	}
	if len(discarded) > 0 {
		g.Message = discardMsg
		g.runCardsDiscardedHooks(seat, "discard_phase", discarded, events)
	}
}

func (g *Game) endTurn(events *[]GameEvent) error {
	seat := g.CurrentTurn

	// 破军：回合结束后，获得「营」中的牌
	g.startPojunGainIfNeeded(seat, events)

	// 进入回合结束阶段
	return g.enterFinishPhase(seat, events)
}
