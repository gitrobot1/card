package engine

import "fmt"

func (g *Game) weaponKind(seat int) string {
	if seat < 0 || seat >= len(g.Players) || g.Players[seat].Weapon == nil {
		return ""
	}
	return g.Players[seat].Weapon.Kind
}

func (g *Game) hasWeaponKind(seat int, kind string) bool {
	return g.weaponKind(seat) == kind
}

func (g *Game) canUseSha(seat int) bool {
	if g.getSkillCounter(seat, counterGuoseShaBlocked) > 0 {
		return false
	}
	if g.skillUnlimitedSha(seat) {
		return true
	}
	p := &g.Players[seat]
	if !p.ShaUsedThisTurn {
		return true
	}
	if g.is3pDdz() && seat == g.LandlordSeat && !p.ShaExtraUsedThisTurn {
		return true
	}
	return false
}

func (g *Game) countShaInHand(seat int) int {
	n := 0
	for _, c := range g.Players[seat].Hand {
		if c.Kind == CardSha {
			n++
		}
	}
	return n
}

func (g *Game) targetHasHorse(seat int) bool {
	if seat < 0 || seat >= len(g.Players) {
		return false
	}
	p := &g.Players[seat]
	return p.PlusHorse != nil || p.MinusHorse != nil
}

func (g *Game) isGuanYuFollowPending() bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModeGuanYuFollow
}

func (g *Game) isQilinBowPending() bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModeQilinBow
}

func (g *Game) appendWeaponSkillEvent(events *[]GameEvent, source, target int, message string) {
	*events = append(*events, GameEvent{
		Type:        "weapon_skill",
		PlayerIndex: source,
		TargetIndex: target,
		Message:     message,
	})
}

func (g *Game) offerGuanYuFollowUp(source, target int, events *[]GameEvent) bool {
	if !g.hasWeaponKind(source, CardWeapon3) || g.countShaInHand(source) == 0 {
		return false
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  source,
		ReturnIndex:  source,
		EffectTarget: target,
		ResponseMode: ResponseModeGuanYuFollow,
		Card:         Card{Kind: CardSha, Name: "杀"},
	}
	g.Message = fmt.Sprintf("%s 可发动【青龙偃月刀】对 %s 再使用一张【杀】", g.Players[source].Name, g.Players[target].Name)
	g.appendWeaponSkillEvent(events, source, target, g.Message)
	g.resetTimer()
	return true
}

func (g *Game) finishGuanYuFollowUp(seat int, events *[]GameEvent) error {
	if !g.isGuanYuFollowPending() || seat != g.Pending.TargetIndex {
		return ErrWrongPhase
	}
	returnIndex := g.Pending.ReturnIndex
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = returnIndex
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[returnIndex].Name)
	g.resetTimer()
	return nil
}

func (g *Game) offerQilinBow(source, target int, returnIndex int, events *[]GameEvent) bool {
	if !g.hasWeaponKind(source, CardWeapon5) || !g.targetHasHorse(target) {
		return false
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  source,
		ReturnIndex:  returnIndex,
		EffectTarget: target,
		ResponseMode: ResponseModeQilinBow,
		Card:         Card{Kind: CardSha, Name: "杀"},
	}
	g.Message = fmt.Sprintf("%s 可发动【麒麟弓】弃置 %s 的坐骑", g.Players[source].Name, g.Players[target].Name)
	g.appendWeaponSkillEvent(events, source, target, g.Message)
	g.resetTimer()
	return true
}

func (g *Game) qilinDiscardHorse(seat int, zone string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isQilinBowPending() || seat != g.Pending.TargetIndex {
		return ErrWrongPhase
	}
	target := g.Pending.EffectTarget
	card, label, ok := g.takeTargetCard(target, PlayTarget{Zone: zone}, events)
	if !ok {
		return ErrInvalidTarget
	}
	g.DiscardPile = append(g.DiscardPile, card)
	returnIndex := g.Pending.ReturnIndex
	*events = append(*events, GameEvent{
		Type:        "qilin_discard",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &card,
		Message:     fmt.Sprintf("%s 发动【麒麟弓】弃置 %s 的%s", g.Players[seat].Name, g.Players[target].Name, label),
	})
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = returnIndex
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[returnIndex].Name)
	g.resetTimer()
	return nil
}

func (g *Game) finishQilinBow(seat int, events *[]GameEvent) error {
	if !g.isQilinBowPending() || seat != g.Pending.TargetIndex {
		return ErrWrongPhase
	}
	returnIndex := g.Pending.ReturnIndex
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = returnIndex
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[returnIndex].Name)
	g.resetTimer()
	return nil
}
