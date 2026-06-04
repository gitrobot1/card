package engine

import "github.com/time/card/backend/internal/game/yuzhousha/skill"

func (p *Player) judgeCardByKind(kind string) *Card {
	for i := range p.JudgeArea {
		if p.JudgeArea[i].Kind == kind {
			return &p.JudgeArea[i]
		}
	}
	return nil
}

func (p *Player) hasJudgeKind(kind string) bool {
	return p.judgeCardByKind(kind) != nil
}

func (g *Game) setJudgeCard(seat int, card Card) {
	p := &g.Players[seat]
	for i, c := range p.JudgeArea {
		if c.Kind == card.Kind {
			g.DiscardPile = append(g.DiscardPile, c)
			p.JudgeArea[i] = card
			return
		}
	}
	p.JudgeArea = append(p.JudgeArea, card)
}

func (g *Game) removeJudgeCard(seat int, cardID string) (Card, bool) {
	p := &g.Players[seat]
	for i, c := range p.JudgeArea {
		if cardID != "" && c.ID != cardID {
			continue
		}
		card := c
		p.JudgeArea = append(p.JudgeArea[:i], p.JudgeArea[i+1:]...)
		return card, true
	}
	return Card{}, false
}

func (g *Game) removeJudgeByKind(seat int, kind string) (Card, bool) {
	p := &g.Players[seat]
	for i, c := range p.JudgeArea {
		if c.Kind == kind {
			card := c
			p.JudgeArea = append(p.JudgeArea[:i], p.JudgeArea[i+1:]...)
			return card, true
		}
	}
	return Card{}, false
}

func (g *Game) takeJudgeCard(target int, cardID string) (Card, string, bool) {
	card, ok := g.removeJudgeCard(target, cardID)
	if !ok {
		return Card{}, "", false
	}
	if card.Kind == CardLeBu {
		g.Players[target].SkipPlay = false
	} else if card.Kind == CardBingLiang {
		g.Players[target].SkipDraw = false
	}
	return card, "判定区", true
}

func (g *Game) judgeAreaCount(target int) int {
	return len(g.Players[target].JudgeArea)
}

func isLightningStrike(suit string, rank int) bool {
	if suit != "S" {
		return false
	}
	if rank == 15 {
		return true
	}
	return rank >= 3 && rank <= 9
}

func trickStaysInJudge(kind string) bool {
	switch kind {
	case CardLeBu, CardBingLiang, CardShanDian:
		return true
	default:
		return false
	}
}

func (g *Game) canBingliangTarget(from, to int) bool {
	if g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookTrickIgnoresDistance, Seat: from, TrickKind: CardBingLiang,
	}).Bool {
		return true
	}
	return g.distanceBetween(from, to) <= 1
}
