package uno

import "math/rand"

func RunAITurns(g *Game, events *[]GameEvent) {
	for g.Phase == PhasePlaying && !g.IsHumanTurn() && !g.IsFinished() {
		seat := g.CurrentTurn

		if g.hasDrawStack() {
			playable := g.PlayableCards(seat)
			if stackCard := pickAIStackCard(g, playable); stackCard != nil {
				color := g.CurrentColor
				if IsWildCard(*stackCard) {
					color = pickAIColor(g, seat)
				}
				if err := g.Play(seat, stackCard.ID, color, events); err == nil {
					continue
				}
			}
			_ = g.Draw(seat, events)
			continue
		}

		if g.MustPlayAfterStack {
			playable := g.PlayableCards(seat)
			if len(playable) > 0 {
				card := pickAICard(g, seat, playable)
				color := g.CurrentColor
				if IsWildCard(card) {
					color = pickAIColor(g, seat)
				}
				if err := g.Play(seat, card.ID, color, events); err == nil {
					continue
				}
			}
			if err := g.Draw(seat, events); err != nil {
				break
			}
			continue
		}

		playable := g.PlayableCards(seat)
		if len(playable) > 0 {
			card := pickAICard(g, seat, playable)
			color := g.CurrentColor
			if IsWildCard(card) {
				color = pickAIColor(g, seat)
			}
			if err := g.Play(seat, card.ID, color, events); err == nil {
				continue
			}
		}

		if err := g.Draw(seat, events); err != nil {
			break
		}
	}
}

func pickAIColor(g *Game, seat int) Color {
	counts := map[Color]int{}
	for _, c := range g.Players[seat].Hand {
		if IsWildCard(c) {
			continue
		}
		counts[c.Color]++
	}
	best := ColorRed
	bestN := -1
	for _, color := range PlayColors {
		if counts[color] > bestN {
			bestN = counts[color]
			best = color
		}
	}
	if bestN <= 0 {
		return PlayColors[rand.Intn(len(PlayColors))]
	}
	return best
}

func pickAIStackCard(g *Game, playable []Card) *Card {
	var draw2, wild4 []Card
	for _, c := range playable {
		switch Value(c.Value) {
		case ValueDraw2:
			draw2 = append(draw2, c)
		case ValueWild4:
			wild4 = append(wild4, c)
		}
	}
	if !g.DrawStackWild4Only && len(draw2) > 0 {
		c := draw2[rand.Intn(len(draw2))]
		return &c
	}
	if len(wild4) > 0 {
		c := wild4[rand.Intn(len(wild4))]
		return &c
	}
	return nil
}

func pickAICard(g *Game, seat int, playable []Card) Card {
	var normal, wilds []Card
	for _, c := range playable {
		if IsWildCard(c) {
			wilds = append(wilds, c)
		} else {
			normal = append(normal, c)
		}
	}
	if len(normal) > 0 {
		return normal[rand.Intn(len(normal))]
	}
	return wilds[rand.Intn(len(wilds))]
}
