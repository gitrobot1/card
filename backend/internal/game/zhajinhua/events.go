package zhajinhua

const (
	MinPlayers = 2
	MaxPlayers = 8

	PhaseBetting  = "betting"
	PhaseFinished = "finished"

	DefaultChips  = 2000
	DefaultAnte   = 10
	DefaultMinRaise = 10
	CompareCost   = 10
)

type GameEvent struct {
	Type         string `json:"type"`
	PlayerIndex  int    `json:"player_index"`
	PlayerName   string `json:"player_name"`
	TargetIndex  int    `json:"target_index,omitempty"`
	TargetName   string `json:"target_name,omitempty"`
	Amount       int    `json:"amount,omitempty"`
	HandType     string `json:"hand_type,omitempty"`
	HandLabel    string `json:"hand_label,omitempty"`
	Multiplier   int    `json:"multiplier,omitempty"`
	Message      string `json:"message,omitempty"`
}
