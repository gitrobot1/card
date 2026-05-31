package zhajinhua

import (
	"math/rand"
)

func RunAITurns(g *Game, events *[]GameEvent) {
	for g.Phase == PhaseBetting && !g.IsHumanTurn() && !g.IsFinished() {
		idx := g.CurrentTurn
		p := &g.Players[idx]
		if p.Folded {
			g.advanceTurn(events)
			continue
		}

		if !p.Looked && rand.Intn(100) < 80 {
			_ = g.Look(idx, events)
		}

		pat, err := AnalyzeHand(p.Hand)
		if err != nil {
			_ = g.Fold(idx, events)
			continue
		}

		strength := HandTypeRank(pat.Type)
		if strength <= 1 && rand.Intn(100) < 55 {
			_ = g.Fold(idx, events)
			continue
		}

		active := g.activeIndices()
		if strength >= 4 && len(active) == 2 {
			other := active[0]
			if other == idx {
				other = active[1]
			}
			if g.Players[other].Looked && rand.Intn(100) < 40 {
				_ = g.Compare(idx, other, events)
				continue
			}
		}

		if strength >= 5 && g.CurrentBet < g.BaseAnte*4 && rand.Intn(100) < 35 {
			target := g.CurrentBet + g.MinRaise
			if target <= p.BetRound+p.Chips {
				_ = g.Raise(idx, target, events)
				continue
			}
		}

		if g.callCost(idx) > p.Chips/3 && strength < 3 {
			_ = g.Fold(idx, events)
			continue
		}

		_ = g.Follow(idx, events)
	}
}
