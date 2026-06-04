package engine

import "github.com/time/card/backend/internal/game/yuzhousha/skill"

func cardView(c Card) skill.CardView {
	return skill.CardView{
		ID: c.ID, Kind: c.Kind, Suit: c.Suit, Label: c.Label, Name: c.Name, Rank: c.Rank,
	}
}

// runSkillHooks 统一技能 hook 入口；新增时机只扩展 skill.HookKind 与 switch 分支。
func (g *Game) runSkillHooks(events *[]GameEvent, call skill.HookCall) skill.HookResult {
	rt := g.skillRuntime(events)
	switch call.Kind {
	case skill.HookTargetBlocked:
		for _, h := range g.playerSkillHandlers(call.Target) {
			if h.BlocksTarget(rt, call.Target, call.CardKind) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookDistanceDelta:
		sum := 0
		for _, h := range g.playerSkillHandlers(call.From) {
			sum += h.DistanceDelta(rt, call.From, call.To)
		}
		return skill.HookResult{Int: sum}

	case skill.HookTrickIgnoresDistance:
		for _, h := range g.playerSkillHandlers(call.Seat) {
			if h.TrickIgnoresDistance(rt, call.Seat, call.TrickKind) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookInstantTrickUsed:
		if trickStaysInJudge(call.TrickKind) {
			return skill.HookResult{}
		}
		for _, h := range g.playerSkillHandlers(call.Seat) {
			if err := h.OnInstantTrickUsed(rt, call.Seat, call.TrickKind); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookCardPlaysAs:
		for _, h := range g.playerSkillHandlers(call.Seat) {
			if h.CardPlaysAs(rt, call.Seat, call.CardKind, call.AsKind, call.Suit) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookUnlimitedSha:
		for _, h := range g.playerSkillHandlers(call.Seat) {
			if h.UnlimitedSha(rt, call.Seat) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookDamageDealt:
		if call.Damage == nil {
			return skill.HookResult{}
		}
		ctx := *call.Damage
		for _, h := range g.playerSkillHandlers(ctx.Target) {
			if err := h.OnDamageDealt(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookJudgeResult:
		if call.Judge == nil {
			return skill.HookResult{}
		}
		ctx := *call.Judge
		for _, h := range g.playerSkillHandlers(ctx.Seat) {
			if err := h.OnJudgeResult(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookCardsDiscarded:
		if call.Discarded == nil {
			return skill.HookResult{}
		}
		ctx := *call.Discarded
		for _, h := range g.playerSkillHandlers(ctx.Seat) {
			if err := h.OnCardsDiscarded(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookEquipLost:
		if call.EquipLost == nil {
			return skill.HookResult{}
		}
		ctx := *call.EquipLost
		for _, h := range g.playerSkillHandlers(ctx.Seat) {
			if err := h.OnEquipLost(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	default:
		return skill.HookResult{}
	}
}

func (g *Game) targetBlockedBySkill(target int, cardKind string) bool {
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookTargetBlocked, Target: target, CardKind: cardKind,
	}).Bool
}

func (g *Game) skillDistanceDelta(from, to int) int {
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookDistanceDelta, From: from, To: to,
	}).Int
}

func (g *Game) trickIgnoresDistance(seat int, trickKind string) bool {
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookTrickIgnoresDistance, Seat: seat, TrickKind: trickKind,
	}).Bool
}

func (g *Game) notifyInstantTrickUsed(seat int, trickKind string, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookInstantTrickUsed, Seat: seat, TrickKind: trickKind,
	})
}

func (g *Game) cardPlaysAsViaHooks(seat int, card Card, asKind string) bool {
	if card.Kind == asKind {
		return true
	}
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookCardPlaysAs, Seat: seat,
		CardKind: card.Kind, AsKind: asKind, Suit: card.Suit,
	}).Bool
}

func (g *Game) skillUnlimitedShaViaHooks(seat int) bool {
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookUnlimitedSha, Seat: seat,
	}).Bool
}

func (g *Game) runDamageDealtHooks(ctx skill.DamageCtx, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookDamageDealt, Damage: &ctx,
	})
}

func (g *Game) runJudgeResultHooks(ctx skill.JudgeCtx, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookJudgeResult, Judge: &ctx,
	})
}

func (g *Game) runCardsDiscardedHooks(seat int, reason string, cards []Card, events *[]GameEvent) {
	if len(cards) == 0 {
		return
	}
	views := make([]skill.CardView, len(cards))
	for i, c := range cards {
		views[i] = cardView(c)
	}
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookCardsDiscarded,
		Seat: seat,
		Discarded: &skill.CardsDiscardedCtx{
			Seat: seat, Reason: reason, Cards: views,
		},
	})
}

func (g *Game) runCardResolvedHooks(seat int, card Card, events *[]GameEvent) {
	rt := g.skillRuntime(events)
	ctx := skill.CardResolvedCtx{Seat: seat, Card: cardView(card)}
	for _, h := range g.playerSkillHandlers(seat) {
		if err := h.OnCardResolved(rt, ctx); err != nil {
			return
		}
	}
}

func (g *Game) effectiveSuitViaHooks(seat int, suit string) string {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(seat) {
		suit = h.EffectiveSuit(rt, seat, suit)
	}
	return suit
}

func (g *Game) trickBlockedViaHooks(target int, card Card) bool {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(target) {
		if h.BlocksTrickTarget(rt, target, card.Kind, card.Suit) {
			return true
		}
	}
	return false
}

func (g *Game) peachBlockedViaHooks(userSeat int) bool {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(g.CurrentTurn) {
		if h.BlocksPeachUse(rt, userSeat) {
			return true
		}
	}
	return false
}

func (g *Game) damageAsHPLossViaHooks(source int) bool {
	if source < 0 || source >= len(g.Players) {
		return false
	}
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(source) {
		if h.DamageAsHPLoss(rt, source) {
			return true
		}
	}
	return false
}

func (g *Game) extraResponsesNeededViaHooks(source int, cardKind string) int {
	if source < 0 || source >= len(g.Players) {
		return 0
	}
	rt := g.skillRuntime(nil)
	extra := 0
	for _, h := range g.playerSkillHandlers(source) {
		extra += h.ExtraResponsesNeeded(rt, source, cardKind)
	}
	return extra
}

func (g *Game) skipsDiscardViaHooks(seat int) bool {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(seat) {
		if h.SkipsDiscardPhase(rt, seat) {
			return true
		}
	}
	return false
}

func (g *Game) notifyEquipLost(seat int, card Card, reason string, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookEquipLost,
		Seat: seat,
		EquipLost: &skill.EquipLostCtx{
			Seat: seat, Reason: reason, Card: cardView(card),
		},
	})
}

func isEquipZone(zone string) bool {
	return zone == EquipWeapon || zone == EquipArmor || zone == EquipPlusHorse || zone == EquipMinusHorse
}

// applyDamage 统一扣血并广播 OnDamageDealt。
func (g *Game) applyDamage(source, target, amount int, damageCard Card, events *[]GameEvent) int {
	if amount <= 0 || target < 0 || target >= len(g.Players) {
		return 0
	}
	p := &g.Players[target]
	if p.HP <= 0 {
		return 0
	}
	p.HP -= amount
	if p.HP < 0 {
		p.HP = 0
	}
	if !g.isJueqingHarm(source) {
		ctx := skill.DamageCtx{
			Source: source, Target: target, Amount: amount,
			CardKind: damageCard.Kind, CardName: damageCard.Name,
		}
		if damageCard.ID != "" {
			ctx.Card = cardView(damageCard)
		}
		g.runDamageDealtHooks(ctx, events)
	}
	return amount
}
