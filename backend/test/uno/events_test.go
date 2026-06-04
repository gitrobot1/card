package uno_test

import (
	"testing"

	uno "github.com/time/card/backend/internal/game/uno"
)

func TestFilterEventsForSeatHidesOpponentDrawCard(t *testing.T) {
	card := uno.Card{ID: "c1", Color: uno.ColorRed, Value: "7", Label: "7"}
	events := []uno.GameEvent{
		{Type: uno.EventDraw, PlayerIndex: 1, PlayerName: "电脑1", Card: &card, Amount: 1},
		{Type: uno.EventDraw, PlayerIndex: 0, PlayerName: "我", Card: &card, Amount: 1},
	}
	filtered := uno.FilterEventsForSeat(events, 0)
	if filtered[0].Card != nil {
		t.Fatal("opponent draw card should be hidden")
	}
	if filtered[1].Card == nil {
		t.Fatal("own draw card should remain visible")
	}
}
