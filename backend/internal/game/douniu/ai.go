package douniu

import "math/rand"

func RunAIActions(g *Game, events *[]GameEvent) {
	for !g.IsFinished() {
		acted := false
		if g.Phase == PhaseGrabBanker {
			for i := range g.Players {
				if !g.Players[i].IsAI || g.Players[i].GrabMult != GrabUnset {
					continue
				}
				mult := aiGrabMult(g, i)
				_ = g.GrabBanker(i, mult, events)
				acted = true
				break
			}
		} else if g.Phase == PhaseBetting {
			for i := range g.Players {
				if i == g.BankerIndex || !g.Players[i].IsAI || g.Players[i].BetMult != BetUnset {
					continue
				}
				mult := aiBetMult(g, i)
				_ = g.PlaceBet(i, mult, events)
				acted = true
				break
			}
		}
		if !acted {
			break
		}
	}
}

func aiGrabMult(g *Game, seat int) int {
	res := AnalyzeHand(g.Players[seat].Hand)
	rank := typeRank(res.Type)
	switch {
	case rank >= 8:
		return 4
	case rank >= 6:
		return 3
	case rank >= 4:
		return 2
	case rank >= 2:
		return 1
	default:
		if rand.Intn(100) < 25 {
			return 1
		}
		return 0
	}
}

func aiBetMult(g *Game, seat int) int {
	res := AnalyzeHand(g.Players[seat].Hand)
	rank := typeRank(res.Type)
	switch {
	case rank >= 8:
		return 5
	case rank >= 6:
		return 3
	case rank >= 4:
		return 2
	default:
		return 1
	}
}
