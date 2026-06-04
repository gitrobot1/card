package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

const (
	Mode1v1     = mode.Solo1v1
	Mode2v2     = mode.Solo2v2
	Mode3pChain = mode.Solo3pChain
	Mode3pDdz   = mode.Solo3pDdz
)

func (g *Game) ModeID() string   { return g.Mode }
func (g *Game) PlayerCount() int { return len(g.Players) }
func (g *Game) AliveHP(seat int) int {
	if seat < 0 || seat >= len(g.Players) {
		return 0
	}
	return g.Players[seat].HP
}

func (g *Game) is2v2() bool              { return mode.Is2v2(g) }
func (g *Game) is3pChain() bool          { return mode.Is3pChain(g) }
func (g *Game) is3pDdz() bool            { return mode.Is3pDdz(g) }
func (g *Game) teamOf(seat int) int      { return mode.TeamOf(g, seat) }
func (g *Game) isEnemy(a, b int) bool    { return mode.IsEnemy(g, a, b) }
func (g *Game) isAlly(a, b int) bool     { return mode.IsAlly(g, a, b) }
func (g *Game) teammateOf(seat int) int { return mode.TeammateOf(g, seat) }
func (g *Game) opponentOf1v1(seat int) int { return mode.Opponent1v1(seat) }
func (g *Game) alliesOf(seat int) []int  { return mode.AlliesOf(g, seat) }
func (g *Game) enemiesOf(seat int) []int { return mode.EnemiesOf(g, seat) }

func (g *Game) opponentOf(seat int) int {
	return mode.DefaultEnemy(g, seat)
}

func (g *Game) firstEnemyWhere(seat int, pred func(int) bool) int {
	for _, e := range g.enemiesOf(seat) {
		if pred(e) {
			return e
		}
	}
	return g.opponentOf(seat)
}

func (g *Game) pickAITarget(seat int) int {
	return mode.PickAITarget(g.targetCtx(), seat, mode.TargetSha)
}

func (g *Game) nextTurnSeat(seat int) int {
	return mode.NextTurnSeat(g, seat)
}

func (g *Game) seatDistance(from, to int) int {
	return mode.SeatDistance(from, to, len(g.Players))
}

func (g *Game) aoeResponderQueue(source int) []int {
	return mode.AoeResponderQueue(g, source)
}

func (g *Game) finishTeamGame(winnerTeam int, events *[]GameEvent) {
	winnerSeat := 0
	for i := range g.Players {
		if g.teamOf(i) == winnerTeam {
			winnerSeat = i
			break
		}
	}
	g.Phase = PhaseFinished
	g.TurnStep = ""
	g.Pending = nil
	g.WinnerIndex = &winnerSeat
	g.WinnerTeam = &winnerTeam
	g.Message = mode.FormatTeamWinMessage(g.teamOf(g.HumanPlayer), winnerTeam)
	*events = append(*events, GameEvent{
		Type:        "game_over",
		PlayerIndex: winnerSeat,
		Message:     g.Message,
	})
}

func (g *Game) checkTeamElimination(events *[]GameEvent) bool {
	adapter := teamElimAdapter{g: g, events: events}
	return mode.CheckWinAfterDamage(g, adapter, &adapter.out)
}

type teamElimAdapter struct {
	g      *Game
	events *[]GameEvent
	out    []mode.TeamEvent
}

func (a teamElimAdapter) ModeID() string    { return a.g.ModeID() }
func (a teamElimAdapter) PlayerCount() int  { return a.g.PlayerCount() }
func (a teamElimAdapter) AliveHP(seat int) int { return a.g.AliveHP(seat) }

func (a teamElimAdapter) FinishTeamGame(winnerTeam int, events *[]mode.TeamEvent) {
	a.g.finishTeamGame(winnerTeam, a.events)
}

func (g *Game) finishChainGame(humanWon bool, message string, events *[]GameEvent) {
	g.Phase = PhaseFinished
	g.TurnStep = ""
	g.Pending = nil
	g.dyingContext = nil
	winner := g.HumanPlayer
	if !humanWon {
		for i := range g.Players {
			if i != g.HumanPlayer && g.AliveHP(i) > 0 {
				winner = i
				break
			}
		}
	}
	g.WinnerIndex = &winner
	g.Message = message
	*events = append(*events, GameEvent{
		Type:        "game_over",
		PlayerIndex: winner,
		Message:     message,
	})
}

func (g *Game) checkChainDeath(victim int, events *[]GameEvent) bool {
	if !g.is3pChain() {
		return false
	}
	human := g.HumanPlayer
	n := len(g.Players)
	mark := mode.UpperSeat(human, n)
	protect := mode.LowerSeat(human, n)
	switch victim {
	case human:
		g.finishChainGame(false, "你已阵亡，失败", events)
		return true
	case protect:
		g.finishChainGame(false, fmt.Sprintf("%s（你的下家）阵亡，失败", g.Players[victim].Name), events)
		return true
	case mark:
		g.finishChainGame(true, fmt.Sprintf("%s（你的上家）阵亡，胜利！", g.Players[victim].Name), events)
		return true
	default:
		return false
	}
}
