package uno

import (
	"fmt"
	"sort"
)

const (
	EventPlayerOut = "player_out"
	EventVoteEnd   = "vote_end"
)

var (
	ErrCannotVoteEnd = fmt.Errorf("cannot vote to end now")
	ErrAlreadyVoted  = fmt.Errorf("already voted to end")
)

func (g *Game) initEliminationTracking() {
	g.EliminationNextRank = 1
	g.EndVotes = nil
	g.updateEndVoteAvailability()
}

func (g *Game) isActive(seat int) bool {
	if seat < 0 || seat >= len(g.Players) {
		return false
	}
	return !g.Players[seat].Eliminated
}

func (g *Game) activeCount() int {
	n := 0
	for _, p := range g.Players {
		if !p.Eliminated {
			n++
		}
	}
	return n
}

func (g *Game) activeSeats() []int {
	var out []int
	for _, p := range g.Players {
		if !p.Eliminated {
			out = append(out, p.Index)
		}
	}
	return out
}

func (g *Game) finishedCount() int {
	n := 0
	for _, p := range g.Players {
		if p.Eliminated && p.FinishRank > 0 {
			n++
		}
	}
	return n
}

func (g *Game) nextEliminationRank() int {
	rank := g.EliminationNextRank
	g.EliminationNextRank++
	return rank
}

func (g *Game) hasVotedEnd(seat int) bool {
	for _, s := range g.EndVotes {
		if s == seat {
			return true
		}
	}
	return false
}

func (g *Game) updateEndVoteAvailability() {
	g.CanVoteToEnd = g.Phase == PhasePlaying && g.activeCount() <= 3 && g.activeCount() > 1
}

func (g *Game) eliminatePlayer(seat int, events *[]GameEvent) {
	if g.Players[seat].Eliminated {
		return
	}
	g.Players[seat].Eliminated = true
	rank := g.nextEliminationRank()
	g.Players[seat].FinishRank = rank
	p := g.Players[seat]
	appendEvent(events, GameEvent{
		Type:        EventPlayerOut,
		PlayerIndex: seat,
		PlayerName:  p.Name,
		Amount:      rank,
		Message:     fmt.Sprintf("%s 出完牌（第 %d 名）", p.Name, rank),
	})
}

func (g *Game) skipToActiveTurn() {
	if g.isActive(g.CurrentTurn) {
		return
	}
	g.CurrentTurn = g.nextSeat(g.CurrentTurn)
}

func (g *Game) checkAfterElimination(events *[]GameEvent) {
	g.updateEndVoteAvailability()
	active := g.activeCount()
	if active == 1 {
		for i := range g.Players {
			if !g.Players[i].Eliminated {
				g.Players[i].FinishRank = len(g.Players)
				g.finishMultiPlayerGame(events)
				return
			}
		}
	}
	g.resetTurnTimer()
	if g.hasDrawStack() {
		g.setStackTurnMessage()
	} else if g.isActive(g.CurrentTurn) {
		g.Message = fmt.Sprintf("%s 出牌", g.Players[g.CurrentTurn].Name)
	}
}

func (g *Game) handleEmptyHand(seat int, card Card, events *[]GameEvent) {
	if len(g.Players) <= 2 {
		g.finishWinner(seat, events)
		return
	}
	g.eliminatePlayer(seat, events)
	g.resolveTurnAfterPlay(card, events)
	g.skipToActiveTurn()
	g.checkAfterElimination(events)
}

func (g *Game) VoteEnd(seat int, events *[]GameEvent) error {
	if g.Phase != PhasePlaying {
		return ErrGameOver
	}
	if !g.CanVoteToEnd {
		return ErrCannotVoteEnd
	}
	if !g.isActive(seat) {
		return ErrNotYourTurn
	}
	if g.hasVotedEnd(seat) {
		return ErrAlreadyVoted
	}
	g.EndVotes = append(g.EndVotes, seat)
	p := g.Players[seat]
	appendEvent(events, GameEvent{
		Type:        EventVoteEnd,
		PlayerIndex: seat,
		PlayerName:  p.Name,
		Message:     fmt.Sprintf("%s 同意结束", p.Name),
	})
	if g.allActiveVotedEnd() {
		g.finishByVote(events)
	}
	return nil
}

func (g *Game) allActiveVotedEnd() bool {
	for _, p := range g.Players {
		if !p.Eliminated && !g.hasVotedEnd(p.Index) {
			return false
		}
	}
	return g.activeCount() > 0
}

func (g *Game) finishByVote(events *[]GameEvent) {
	active := g.activeSeats()
	sort.Slice(active, func(i, j int) bool {
		return len(g.Players[active[i]].Hand) < len(g.Players[active[j]].Hand)
	})
	base := g.finishedCount()
	for i, seat := range active {
		g.Players[seat].FinishRank = base + i + 1
	}
	g.finishMultiPlayerGame(events)
}

func (g *Game) finishMultiPlayerGame(events *[]GameEvent) {
	g.Placements = g.buildPlacements()
	g.Phase = PhaseFinished
	winner := g.placementWinner()
	g.WinnerIndex = &winner
	msg := fmt.Sprintf("%s 获胜！", g.Players[winner].Name)
	appendEvent(events, GameEvent{
		Type:        EventGameOver,
		PlayerIndex: winner,
		PlayerName:  g.Players[winner].Name,
		Message:     msg,
	})
	g.Message = msg
}

func (g *Game) placementWinner() int {
	for _, p := range g.Players {
		if p.FinishRank == 1 {
			return p.Index
		}
	}
	if len(g.Placements) > 0 {
		return g.Placements[0]
	}
	return 0
}

func (g *Game) buildPlacements() []int {
	type ranked struct {
		seat int
		rank int
	}
	var rows []ranked
	for _, p := range g.Players {
		if p.FinishRank > 0 {
			rows = append(rows, ranked{seat: p.Index, rank: p.FinishRank})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].rank < rows[j].rank
	})
	out := make([]int, len(rows))
	for i, r := range rows {
		out[i] = r.seat
	}
	return out
}

func RunAIVoteEnd(g *Game, events *[]GameEvent) {
	if !g.CanVoteToEnd || g.Phase != PhasePlaying {
		return
	}
	for _, p := range g.Players {
		if p.IsAI && !p.Eliminated && !g.hasVotedEnd(p.Index) {
			_ = g.VoteEnd(p.Index, events)
		}
	}
}
