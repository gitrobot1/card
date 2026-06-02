package douniu

type GameEvent struct {
	Type         string `json:"type"`
	PlayerIndex  int    `json:"player_index"`
	PlayerName   string `json:"player_name"`
	TargetIndex  int    `json:"target_index,omitempty"`
	TargetName   string `json:"target_name,omitempty"`
	Amount       int    `json:"amount,omitempty"`
	GrabMult     int    `json:"grab_mult"`
	BetMult      int    `json:"bet_mult"`
	HandType     string `json:"hand_type,omitempty"`
	HandLabel    string `json:"hand_label,omitempty"`
	Multiplier   int    `json:"multiplier,omitempty"`
	Message      string `json:"message,omitempty"`
}
