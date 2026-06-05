package mode

import "fmt"

const Solo3v3 = "3v3"

// 3v3 座位（顺时针）：0 暖主帅 · 1 冷前锋 · 2 冷主帅 · 3 冷前锋 · 4 暖前锋 · 5 暖前锋
// 暖色 team 0：0, 4, 5；冷色 team 1：1, 2, 3

func Is3v3(ctx Context) bool {
	return ctx.ModeID() == Solo3v3
}

func TeamOf3v3(seat int) int {
	switch seat {
	case 0, 4, 5:
		return 0
	case 1, 2, 3:
		return 1
	default:
		return seat % 2
	}
}

func CommanderSeat3v3(team int) int {
	if team == 0 {
		return 0
	}
	return 2
}

func IsCommander3v3(seat int) bool {
	return seat == 0 || seat == 2
}

func SeatRole3v3(seat int) string {
	if IsCommander3v3(seat) {
		return "commander"
	}
	return "forward"
}

func TeamLabel3v3(team int) string {
	if team == 0 {
		return "暖色"
	}
	return "冷色"
}

func FormatCommanderWinMessage(humanTeam, winnerTeam int) string {
	if winnerTeam == humanTeam {
		return "己方 获胜（敌方主帅阵亡）"
	}
	return "敌方 获胜（己方主帅阵亡）"
}

func EvaluateCommanderDeath(ctx Context, humanSeat, victim int) (finished bool, winnerTeam int, message string) {
	if !Is3v3(ctx) || !IsCommander3v3(victim) {
		return false, -1, ""
	}
	loserTeam := TeamOf3v3(victim)
	winnerTeam = 1 - loserTeam
	victimName := fmt.Sprintf("%d号位", victim)
	if humanSeat >= 0 && humanSeat < ctx.PlayerCount() {
		_ = victimName
	}
	return true, winnerTeam, FormatCommanderWinMessage(TeamOf3v3(humanSeat), winnerTeam)
}
