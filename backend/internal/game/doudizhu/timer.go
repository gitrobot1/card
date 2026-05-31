package doudizhu

import (
	"fmt"
	"time"
)

const TurnTimeout = 35 * time.Second

func (g *Game) resetTurnTimer() {
	g.TurnDeadline = time.Now().Add(TurnTimeout)
}

func (g *Game) secondsLeft() int {
	if g.TurnDeadline.IsZero() {
		return int(TurnTimeout.Seconds())
	}
	seconds := int(time.Until(g.TurnDeadline).Seconds())
	if seconds < 0 {
		return 0
	}
	if seconds > int(TurnTimeout.Seconds()) {
		return int(TurnTimeout.Seconds())
	}
	return seconds
}

func (g *Game) TurnDeadlineUnix() int64 {
	if g.TurnDeadline.IsZero() {
		return time.Now().Add(TurnTimeout).Unix()
	}
	return g.TurnDeadline.Unix()
}

func (g *Game) IsTurnExpired() bool {
	if g.TurnDeadline.IsZero() {
		return false
	}
	return time.Now().After(g.TurnDeadline)
}

func (g *Game) ApplyHumanTimeout(events *[]GameEvent) error {
	if !g.IsHumanTurn() || g.IsFinished() || !g.IsTurnExpired() {
		return nil
	}

	playerIndex := g.CurrentTurn
	if g.Phase == PhaseCalling {
		err := g.CallLandlord(playerIndex, false)
		if err != nil {
			return err
		}
		appendCallEvent(events, playerIndex, g.Players[playerIndex].Name, false)
		g.Message = fmt.Sprintf("%s 超时，默认不抢", g.Players[playerIndex].Name)
		return nil
	}

	if g.Phase == PhasePlaying {
		if g.LastPlay != nil && g.LastPlay.PlayerIndex != playerIndex {
			if err := g.Pass(playerIndex); err != nil {
				return err
			}
			appendPassEvent(events, playerIndex, g.Players[playerIndex].Name)
			g.Message = fmt.Sprintf("%s 超时，默认不出", g.Players[playerIndex].Name)
			return nil
		}
		cards := pickSmallestPattern(g.Players[playerIndex].Hand)
		record, err := g.Play(playerIndex, cardIDs(cards))
		if err != nil {
			return err
		}
		appendPlayEvent(events, record)
		g.Message = fmt.Sprintf("%s 超时，自动出牌", g.Players[playerIndex].Name)
	}
	return nil
}
