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
	Mode3v3       = mode.Solo3v3
	ModeIdentity5 = mode.SoloIdentity5
	ModeIdentity8 = mode.SoloIdentity8
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
func (g *Game) is3v3() bool              { return mode.Is3v3(g) }
func (g *Game) isIdentity() bool         { return mode.IsIdentity(g) }
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

func (g *Game) checkCommanderDeath(victim int, events *[]GameEvent) bool {
	if !g.is3v3() || !mode.IsCommander3v3(victim) {
		return false
	}
	finished, winnerTeam, msg := mode.EvaluateCommanderDeath(g, g.HumanPlayer, victim)
	if !finished {
		return false
	}
	g.finishTeamGame(winnerTeam, events)
	if msg != "" {
		g.Message = msg
		if len(*events) > 0 {
			(*events)[len(*events)-1].Message = msg
		}
	}
	return true
}

func (g *Game) IdentityLordSeat() int { return g.LordSeat }

func (g *Game) IdentityOf(seat int) string {
	if seat < 0 || seat >= len(g.Identities) {
		return ""
	}
	return g.Identities[seat]
}

func (g *Game) IdentityRevealed(seat int) bool {
	if seat < 0 || seat >= len(g.RoleRevealed) {
		return false
	}
	return g.RoleRevealed[seat]
}

func (g *Game) revealIdentity(victim int, events *[]GameEvent) {
	if victim < 0 || victim >= len(g.RoleRevealed) || g.IdentityRevealed(victim) {
		return
	}
	g.RoleRevealed[victim] = true
	role := g.IdentityOf(victim)
	*events = append(*events, GameEvent{
		Type:        "identity_revealed",
		PlayerIndex: victim,
		Message:     fmt.Sprintf("%s 身份揭示：%s", g.Players[victim].Name, mode.RoleLabel(role)),
	})
}

func (g *Game) checkIdentityDeath(victim, killer int, events *[]GameEvent) bool {
	if !g.isIdentity() {
		return false
	}
	g.revealIdentity(victim, events)
	finished, winnerTeam, msg := mode.EvaluateIdentityWin(g, victim, killer)
	if !finished {
		return false
	}
	g.finishTeamGame(winnerTeam, events)
	if msg != "" {
		g.Message = msg
		if len(*events) > 0 {
			(*events)[len(*events)-1].Message = msg
		}
	}
	return true
}
