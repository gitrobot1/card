package douniu

import "time"

func (g *Game) resetPhaseTimer() {
	g.TurnDeadline = time.Now().Add(TurnTimeoutSec * time.Second)
	g.TurnDeadlineUnix = g.TurnDeadline.Unix()
}

func (g *Game) IsPhaseExpired() bool {
	if g.TurnDeadline.IsZero() || g.IsFinished() {
		return false
	}
	return time.Now().After(g.TurnDeadline)
}

func (g *Game) ApplyHumanTimeout(events *[]GameEvent) error {
	if !g.IsHumanPending() || !g.IsPhaseExpired() {
		return nil
	}
	seat := g.HumanPlayer
	switch g.Phase {
	case PhaseGrabBanker:
		return g.GrabBanker(seat, 0, events)
	case PhaseBetting:
		return g.PlaceBet(seat, 1, events)
	}
	return nil
}
