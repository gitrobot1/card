package uno

type GameEvent struct {
	Type        string `json:"type"`
	PlayerIndex int    `json:"player_index"`
	PlayerName  string `json:"player_name"`
	Card        *Card  `json:"card,omitempty"`
	Color       Color  `json:"color,omitempty"`
	Amount      int    `json:"amount,omitempty"`
	Dice1       int    `json:"dice1,omitempty"`
	Dice2       int    `json:"dice2,omitempty"`
	TiedSeats   []int  `json:"tied_seats,omitempty"`
	Message     string `json:"message,omitempty"`
}

const (
	EventDeal      = "deal"
	EventPlay      = "play"
	EventDraw      = "draw"
	EventPass      = "pass"
	EventColorPick = "color_pick"
	EventGameOver  = "game_over"
)

func appendEvent(events *[]GameEvent, event GameEvent) {
	if events == nil {
		return
	}
	*events = append(*events, event)
}

func filterEventsForSeat(events []GameEvent, seat int) []GameEvent {
	if len(events) == 0 {
		return events
	}
	out := make([]GameEvent, len(events))
	for i, e := range events {
		e2 := e
		if e.Type == EventDraw && e.PlayerIndex != seat {
			e2.Card = nil
		}
		out[i] = e2
	}
	return out
}
