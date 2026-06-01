package uno

import "testing"

func TestHumanPlayThenAIActs(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = ColorRed
	g.TopCard = Card{Color: ColorRed, Value: "5", Label: "5"}
	g.Players[0].Hand = []Card{
		{ID: "r7", Color: ColorRed, Value: "7", Label: "7"},
		{ID: "k", Color: ColorBlue, Value: "2", Label: "2"},
	}
	g.Players[1].Hand = []Card{
		{ID: "r1", Color: ColorRed, Value: "1", Label: "1"},
		{ID: "b3", Color: ColorBlue, Value: "3", Label: "3"},
	}
	var events []GameEvent
	if err := g.Play(0, "r7", "", &events); err != nil {
		t.Fatal(err)
	}
	before := len(events)
	RunAITurns(g, &events)
	if len(events) == before {
		t.Fatal("AI should produce events")
	}
	if g.Phase == PhasePlaying && g.CurrentTurn != 0 {
		t.Fatalf("expected turn back to human in 2p, got %d", g.CurrentTurn)
	}
}

func TestAIStackDrawThenPlay(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 1
	g.CurrentColor = ColorRed
	g.TopCard = Card{Color: ColorRed, Value: "5", Label: "5"}
	g.PendingDrawPenalty = 2
	g.Players[1].Hand = []Card{
		{ID: "r3", Color: ColorRed, Value: "3", Label: "3"},
		{ID: "g9", Color: ColorGreen, Value: "9", Label: "9"},
	}
	var events []GameEvent
	RunAITurns(g, &events)
	if g.CurrentTurn == 1 {
		t.Fatalf("AI should finish post-stack obligation, turn=%d", g.CurrentTurn)
	}
}

func TestAIDrawAdvancesTurn(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 1
	g.CurrentColor = ColorBlue
	g.TopCard = Card{Color: ColorRed, Value: "5", Label: "5"}
	g.Players[1].Hand = []Card{{ID: "x", Color: ColorGreen, Value: "8", Label: "8"}}
	var events []GameEvent
	RunAITurns(g, &events)
	if g.CurrentTurn != 0 {
		t.Fatalf("AI should draw and advance to human, turn=%d", g.CurrentTurn)
	}
	if len(events) == 0 || events[0].Type != EventDraw {
		t.Fatalf("expected draw event, got %v", events)
	}
}
