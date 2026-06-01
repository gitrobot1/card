package uno

import (
	"fmt"
	"math/rand"
)

const (
	EventRollDice    = "roll_dice"
	EventRollTie     = "roll_tie"
	EventFirstPlayer = "first_player"
)

func (g *Game) initRollForFirst() {
	g.Phase = PhaseRollForFirst
	g.rollContenders = make([]int, len(g.Players))
	for i := range g.Players {
		g.rollContenders[i] = i
	}
	g.rollRoundSums = make(map[int]int, len(g.Players))
	for _, seat := range g.rollContenders {
		g.rollRoundSums[seat] = -1
	}
	g.CurrentTurn = -1
	g.Message = "掷骰定先手"
}

// RollRound 所有待定玩家同时掷 2d6，一轮结束后判定先手或平局重掷。
func (g *Game) RollRound(events *[]GameEvent) error {
	if g.Phase != PhaseRollForFirst {
		return ErrWrongPhase
	}
	for _, seat := range g.rollContenders {
		d1 := rand.Intn(6) + 1
		d2 := rand.Intn(6) + 1
		sum := d1 + d2
		g.rollRoundSums[seat] = sum
		appendEvent(events, GameEvent{
			Type:        EventRollDice,
			PlayerIndex: seat,
			PlayerName:  g.Players[seat].Name,
			Dice1:       d1,
			Dice2:       d2,
			Amount:      sum,
			Message:     fmt.Sprintf("%s 掷出 %d+%d=%d", g.Players[seat].Name, d1, d2, sum),
		})
	}
	return g.finalizeRollRound(events)
}

func (g *Game) finalizeRollRound(events *[]GameEvent) error {
	maxSum := -1
	var winners []int
	for _, seat := range g.rollContenders {
		sum := g.rollRoundSums[seat]
		if sum > maxSum {
			maxSum = sum
			winners = []int{seat}
		} else if sum == maxSum {
			winners = append(winners, seat)
		}
	}
	if len(winners) == 1 {
		g.CurrentTurn = winners[0]
		g.deal()
		g.Phase = PhasePlaying
		g.resetTurnTimer()
		g.Message = fmt.Sprintf("%s 先手出牌", g.Players[g.CurrentTurn].Name)
		appendEvent(events, GameEvent{
			Type:        EventFirstPlayer,
			PlayerIndex: winners[0],
			PlayerName:  g.Players[winners[0]].Name,
			Message:     g.Message,
		})
		return nil
	}
	g.rollContenders = winners
	for _, seat := range winners {
		g.rollRoundSums[seat] = -1
	}
	appendEvent(events, GameEvent{
		Type:      EventRollTie,
		Amount:    maxSum,
		TiedSeats: append([]int(nil), winners...),
		Message:   fmt.Sprintf("平局 %d 点，以上玩家重掷", maxSum),
	})
	return nil
}

// SkipRollForFirst 测试或跳过时直接指定先手并发牌。
func (g *Game) SkipRollForFirst(seat int) {
	if g.Phase != PhaseRollForFirst {
		return
	}
	if seat < 0 || seat >= len(g.Players) {
		seat = 0
	}
	g.CurrentTurn = seat
	g.deal()
	g.Phase = PhasePlaying
	g.resetTurnTimer()
	g.Message = fmt.Sprintf("%s 出牌", g.Players[g.CurrentTurn].Name)
}
