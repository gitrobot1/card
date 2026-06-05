package engine

// 古锭刀、藤甲等装备牌效果。

const counterChained = "chained"

func (g *Game) hasVineArmor(seat int) bool {
	p := &g.Players[seat]
	return p.Armor != nil && p.Armor.Kind == CardArmorVine
}

func (g *Game) hasBaguaArmor(seat int) bool {
	p := &g.Players[seat]
	return p.Armor != nil && p.Armor.Kind == CardArmor
}

func (g *Game) vineBlocksTrick(target int, kind string) bool {
	if !g.hasVineArmor(target) {
		return false
	}
	switch kind {
	case CardGuoHe, CardTanNang, CardJueDou, CardNanMan, CardWanJian:
		return true
	default:
		return false
	}
}

func (g *Game) isChained(seat int) bool {
	return g.getSkillCounter(seat, counterChained) > 0
}

func (g *Game) setChained(seat int, chained bool) {
	if chained {
		g.setSkillCounter(seat, counterChained, 1)
	} else {
		g.setSkillCounter(seat, counterChained, 0)
	}
}

func (g *Game) toggleChained(seat int) {
	g.setChained(seat, !g.isChained(seat))
}

func (g *Game) filterAoeQueue(queue []int, trickKind string) []int {
	if trickKind != CardNanMan && trickKind != CardWanJian {
		return queue
	}
	out := make([]int, 0, len(queue))
	for _, seat := range queue {
		if g.vineBlocksTrick(seat, trickKind) {
			continue
		}
		out = append(out, seat)
	}
	return out
}

// adjustDamageAmount 结算古锭刀、藤甲等加伤（青釭剑 ignoreArmor 时藤甲失效）。
func (g *Game) adjustDamageAmount(source, target, amount int, card Card, isFire, ignoreArmor bool) int {
	if amount <= 0 {
		amount = 1
	}
	if card.Kind == CardSha && g.hasWeaponKind(source, CardWeapon6) && len(g.Players[target].Hand) == 0 {
		amount++
	}
	if !ignoreArmor && g.hasVineArmor(target) && (isFire || card.Kind == CardSha) {
		amount++
	}
	return amount
}

func (g *Game) spreadChainedFireDamage(source, primaryTarget, amount int, card Card, events *[]GameEvent) {
	if !g.isChained(primaryTarget) {
		return
	}
	for seat := range g.Players {
		if seat == primaryTarget || !g.isChained(seat) || g.Players[seat].HP <= 0 {
			continue
		}
		dmg := g.adjustDamageAmount(source, seat, amount, card, true, false)
		g.applyDamage(source, seat, dmg, card, events)
		*events = append(*events, GameEvent{
			Type:        "trick_hit",
			PlayerIndex: source,
			TargetIndex: seat,
			Damage:      dmg,
			Message:     g.damageMessage(&g.Players[seat], card.Name, dmg),
		})
		if g.Players[seat].HP <= 0 {
			_ = g.afterDamageApplied(source, seat, dmg, card, DamageResume{}, events)
		}
	}
}
