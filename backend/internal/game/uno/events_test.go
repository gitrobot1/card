package uno

import "testing"

func TestFilterEventsForSeatHidesOpponentDrawCard(t *testing.T) {
	card := Card{ID: "c1", Color: ColorRed, Value: "7", Label: "7"}
	events := []GameEvent{
		{Type: EventDraw, PlayerIndex: 1, PlayerName: "电脑1", Card: &card, Amount: 1},
		{Type: EventDraw, PlayerIndex: 0, PlayerName: "我", Card: &card, Amount: 1},
	}
	filtered := filterEventsForSeat(events, 0)
	if filtered[0].Card != nil {
		t.Fatal("opponent draw card should be hidden")
	}
	if filtered[1].Card == nil {
		t.Fatal("own draw card should remain visible")
	}
}
