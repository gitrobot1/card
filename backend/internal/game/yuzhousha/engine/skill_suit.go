package engine

import "github.com/time/card/backend/internal/game/yuzhousha/skill"

// effectiveSuit 技能修正后的花色（如【红颜】黑桃视为红桃）。
func (g *Game) effectiveSuit(seat int, suit string) string {
	return g.effectiveSuitViaHooks(seat, suit)
}

func (g *Game) isRedSuitFor(seat int, suit string) bool {
	return skill.IsRedSuit(g.effectiveSuit(seat, suit))
}

func (g *Game) hasRedHandCard(seat int) bool {
	for _, c := range g.Players[seat].Hand {
		if g.isRedSuitFor(seat, c.Suit) {
			return true
		}
	}
	return false
}

func (g *Game) isRedHandCard(seat int, card Card) bool {
	return g.isRedSuitFor(seat, card.Suit)
}
