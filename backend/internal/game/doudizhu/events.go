package doudizhu

import "github.com/time/card/backend/internal/game/card"

type EventType string

const (
	EventPlay    EventType = "play"
	EventPass    EventType = "pass"
	EventCall    EventType = "call"
	EventTimeout EventType = "timeout"
	EventGameOver EventType = "game_over"
)

type GameEvent struct {
	Type        EventType   `json:"type"`
	PlayerIndex int         `json:"player_index"`
	PlayerName  string      `json:"player_name"`
	Cards       []card.Card `json:"cards,omitempty"`
	Call        *bool       `json:"call,omitempty"`
}

func appendPlayEvent(events *[]GameEvent, record *PlayRecord) {
	if record == nil {
		return
	}
	*events = append(*events, GameEvent{
		Type:        EventPlay,
		PlayerIndex: record.PlayerIndex,
		PlayerName:  record.PlayerName,
		Cards:       append([]card.Card(nil), record.Cards...),
	})
}

func appendPassEvent(events *[]GameEvent, playerIndex int, playerName string) {
	*events = append(*events, GameEvent{
		Type:        EventPass,
		PlayerIndex: playerIndex,
		PlayerName:  playerName,
	})
}

func appendCallEvent(events *[]GameEvent, playerIndex int, playerName string, call bool) {
	value := call
	*events = append(*events, GameEvent{
		Type:        EventCall,
		PlayerIndex: playerIndex,
		PlayerName:  playerName,
		Call:        &value,
	})
}

func AppendPlayEventPublic(events *[]GameEvent, record *PlayRecord) {
	appendPlayEvent(events, record)
}

func AppendPassEventPublic(events *[]GameEvent, playerIndex int, playerName string) {
	appendPassEvent(events, playerIndex, playerName)
}

func AppendCallEventPublic(events *[]GameEvent, playerIndex int, playerName string, call bool) {
	appendCallEvent(events, playerIndex, playerName, call)
}

func appendGameOverEvent(events *[]GameEvent, winner int, winnerName string) {
	*events = append(*events, GameEvent{
		Type:        EventGameOver,
		PlayerIndex: winner,
		PlayerName:  winnerName,
	})
}

func AppendGameOverEventPublic(events *[]GameEvent, winner int, winnerName string) {
	appendGameOverEvent(events, winner, winnerName)
}
