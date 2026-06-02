package douniu

const (
	MinPlayers = 2
	MaxPlayers = 8

	PhaseGrabBanker = "grab_banker"
	PhaseBetting    = "betting"
	PhaseFinished   = "finished"

	DefaultChips = 2000
	DefaultAnte  = 10
	TurnTimeoutSec = 20

	GrabUnset = -1
	BetUnset  = 0

	MaxGrabMult = 4
)

var BetOptions = []int{1, 2, 3, 5}
