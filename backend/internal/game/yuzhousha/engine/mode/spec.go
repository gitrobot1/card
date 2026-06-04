package mode

import "fmt"

const (
	Solo1v1 = "1v1"
	Solo2v2 = "2v2"
)

// Context exposes read-only game state for mode rules without importing engine.
type Context interface {
	ModeID() string
	PlayerCount() int
	AliveHP(seat int) int
}

func Is2v2(ctx Context) bool {
	return ctx.ModeID() == Solo2v2 || (ctx.PlayerCount() == 4 && ctx.ModeID() != Solo3pChain && ctx.ModeID() != Solo3pDdz)
}

func TeamOf(ctx Context, seat int) int {
	if seat < 0 || seat >= ctx.PlayerCount() {
		return 0
	}
	if ls, ok := landlordSeat(ctx); ok {
		if seat == ls {
			return 0
		}
		return 1
	}
	if Is2v2(ctx) {
		return seat % 2
	}
	return seat
}

func IsEnemy(ctx Context, a, b int) bool {
	n := ctx.PlayerCount()
	if a < 0 || b < 0 || a >= n || b >= n || a == b {
		return false
	}
	if Is3pChain(ctx) {
		return b == MarkTarget(ctx, a)
	}
	if Is3pDdz(ctx) {
		return TeamOf(ctx, a) != TeamOf(ctx, b)
	}
	if !Is2v2(ctx) {
		return a != b
	}
	return TeamOf(ctx, a) != TeamOf(ctx, b)
}

func IsAlly(ctx Context, a, b int) bool {
	if a == b {
		return true
	}
	if Is3pChain(ctx) {
		return b == ProtectTarget(ctx, a)
	}
	if Is3pDdz(ctx) {
		return TeamOf(ctx, a) == TeamOf(ctx, b)
	}
	return !IsEnemy(ctx, a, b)
}

func TeammateOf(ctx Context, seat int) int {
	if Is3pDdz(ctx) {
		ls, ok := landlordSeat(ctx)
		if !ok {
			return Opponent1v1(seat)
		}
		if seat == ls {
			for i := 0; i < ctx.PlayerCount(); i++ {
				if i != seat && TeamOf(ctx, i) == TeamOf(ctx, seat) && ctx.AliveHP(i) > 0 {
					return i
				}
			}
			return (seat + 1) % ctx.PlayerCount()
		}
		return ls
	}
	if !Is2v2(ctx) {
		return Opponent1v1(seat)
	}
	return seat ^ 2
}

func Opponent1v1(seat int) int {
	return 1 - seat
}

// AlliesOf returns alive seats allied with seat (excludes self).
func AlliesOf(ctx Context, seat int) []int {
	out := make([]int, 0, ctx.PlayerCount())
	for i := 0; i < ctx.PlayerCount(); i++ {
		if i != seat && IsAlly(ctx, seat, i) && ctx.AliveHP(i) > 0 {
			out = append(out, i)
		}
	}
	return out
}

func EnemiesOf(ctx Context, seat int) []int {
	out := make([]int, 0, ctx.PlayerCount())
	for i := 0; i < ctx.PlayerCount(); i++ {
		if IsEnemy(ctx, seat, i) && ctx.AliveHP(i) > 0 {
			out = append(out, i)
		}
	}
	return out
}

// DefaultEnemy returns the primary enemy seat for single-target AI / legacy skills.
// In 2v2 it is the next alive enemy clockwise.
func DefaultEnemy(ctx Context, seat int) int {
	if Is3pChain(ctx) {
		mark := MarkTarget(ctx, seat)
		if ctx.AliveHP(mark) > 0 {
			return mark
		}
	}
	n := ctx.PlayerCount()
	if n <= 0 {
		return 0
	}
	if Is2v2(ctx) || Is3pDdz(ctx) {
		for i := 1; i < n; i++ {
			next := (seat + i) % n
			if IsEnemy(ctx, seat, next) && ctx.AliveHP(next) > 0 {
				return next
			}
		}
		return (seat + 1) % n
	}
	return Opponent1v1(seat)
}

func NextTurnSeat(ctx Context, seat int) int {
	n := ctx.PlayerCount()
	for i := 1; i <= n; i++ {
		next := (seat + i) % n
		if ctx.AliveHP(next) > 0 {
			return next
		}
	}
	return (seat + 1) % n
}

func SeatDistance(from, to, playerCount int) int {
	if playerCount <= 2 {
		return 1
	}
	diff := from - to
	if diff < 0 {
		diff = -diff
	}
	ring := diff
	if playerCount-ring < ring {
		ring = playerCount - ring
	}
	if ring < 1 {
		ring = 1
	}
	return ring
}

func AoeResponderQueue(ctx Context, source int) []int {
	n := ctx.PlayerCount()
	out := make([]int, 0, n-1)
	for i := 1; i < n; i++ {
		seat := (source + i) % n
		if ctx.AliveHP(seat) <= 0 {
			continue
		}
		if Is3pChain(ctx) && seat == ProtectTarget(ctx, source) {
			continue
		}
		out = append(out, seat)
	}
	return out
}

type TeamElimination interface {
	Context
	FinishTeamGame(winnerTeam int, events *[]TeamEvent)
}

type TeamEvent struct {
	Type        string
	PlayerIndex int
	Message     string
}

func CheckTeamElimination(g TeamElimination, events *[]TeamEvent) bool {
	if !Is2v2(g) && !Is3pDdz(g) {
		return false
	}
	for team := 0; team <= 1; team++ {
		alive := false
		for seat := 0; seat < g.PlayerCount(); seat++ {
			if TeamOf(g, seat) == team && g.AliveHP(seat) > 0 {
				alive = true
				break
			}
		}
		if !alive {
			g.FinishTeamGame(1-team, events)
			return true
		}
	}
	return false
}

func TeamWinLabel(humanTeam, winnerTeam int) string {
	if winnerTeam == humanTeam {
		return "己方"
	}
	return "敌方"
}

func FormatTeamWinMessage(humanTeam, winnerTeam int) string {
	return fmt.Sprintf("%s 获胜", TeamWinLabel(humanTeam, winnerTeam))
}
