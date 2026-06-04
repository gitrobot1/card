package mode

import "fmt"

const Solo3pChain = "3p_chain"

func Is3pChain(ctx Context) bool {
	return ctx.ModeID() == Solo3pChain
}

// UpperSeat 上家：出牌顺序中的前一位（逆时针）。
func UpperSeat(seat, playerCount int) int {
	if playerCount <= 0 {
		return 0
	}
	return (seat - 1 + playerCount) % playerCount
}

// LowerSeat 下家：出牌顺序中的后一位（顺时针）。
func LowerSeat(seat, playerCount int) int {
	if playerCount <= 0 {
		return 0
	}
	return (seat + 1) % playerCount
}

// MarkTarget 需击杀的上家。
func MarkTarget(ctx Context, seat int) int {
	return UpperSeat(seat, ctx.PlayerCount())
}

// ProtectTarget 需保住的下家。
func ProtectTarget(ctx Context, seat int) int {
	return LowerSeat(seat, ctx.PlayerCount())
}

// HumanChainOutcome after victim dies in 3p chain (solo: human at humanSeat).
type HumanChainOutcome int

const (
	ChainContinue HumanChainOutcome = iota
	ChainHumanWin
	ChainHumanLose
)

func EvaluateHumanChainDeath(ctx Context, humanSeat, victim int) (HumanChainOutcome, string) {
	if !Is3pChain(ctx) || humanSeat < 0 || humanSeat >= ctx.PlayerCount() {
		return ChainContinue, ""
	}
	n := ctx.PlayerCount()
	mark := UpperSeat(humanSeat, n)
	protect := LowerSeat(humanSeat, n)
	switch victim {
	case humanSeat:
		return ChainHumanLose, "你已阵亡，失败"
	case protect:
		return ChainHumanLose, fmt.Sprintf("你的下家（%d号位）阵亡，失败", protect)
	case mark:
		return ChainHumanWin, fmt.Sprintf("你的上家（%d号位）阵亡，胜利！", mark)
	default:
		return ChainContinue, ""
	}
}
