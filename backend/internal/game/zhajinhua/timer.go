package zhajinhua

import "time"

const TurnTimeout = 35 * time.Second

func (g *Game) resetTurnTimer() {
	g.TurnDeadline = time.Now().Add(TurnTimeout)
}

func (g *Game) TurnDeadlineUnix() int64 {
	if g.TurnDeadline.IsZero() {
		return time.Now().Add(TurnTimeout).Unix()
	}
	return g.TurnDeadline.Unix()
}

func (g *Game) IsTurnExpired() bool {
	if g.TurnDeadline.IsZero() || g.Phase != PhaseBetting {
		return false
	}
	return time.Now().After(g.TurnDeadline)
}

func (g *Game) IsHumanTurn() bool {
	if g.Phase != PhaseBetting {
		return false
	}
	if g.CurrentTurn < 0 || g.CurrentTurn >= len(g.Players) {
		return false
	}
	return !g.Players[g.CurrentTurn].IsAI
}
