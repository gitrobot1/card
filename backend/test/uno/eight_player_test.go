package uno_test

import (
	"fmt"
	"testing"
	"time"

	uno "github.com/time/card/backend/internal/game/uno"
)

func TestEightPlayerReverseAIChain(t *testing.T) {
	g := readyGame(t, 7)
	g.CurrentTurn = 0
	g.Direction = 1
	g.CurrentColor = uno.ColorBlue
	g.TopCard = uno.Card{Color: uno.ColorBlue, Value: "3", Label: "3"}
	for i := range g.Players {
		g.Players[i].Hand = []uno.Card{
			{ID: fmtCard(i, "a"), Color: uno.ColorBlue, Value: "1", Label: "1"},
			{ID: fmtCard(i, "b"), Color: uno.ColorRed, Value: "2", Label: "2"},
		}
	}
	g.Players[0].Hand = append(g.Players[0].Hand, uno.Card{
		ID: "rev0", Color: uno.ColorBlue, Value: string(uno.ValueReverse), Label: "反转",
	})
	var events []uno.GameEvent
	done := make(chan error, 1)
	go func() {
		if err := g.Play(0, "rev0", "", &events); err != nil {
			done <- err
			return
		}
		uno.RunAITurns(g, &events)
		done <- nil
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("uno.RunAITurns hung after reverse in 8p game")
	}
	if g.Phase == uno.PhasePlaying && g.CurrentTurn != 0 {
		t.Fatalf("expected turn back to human, got %d", g.CurrentTurn)
	}
	if len(events) == 0 {
		t.Fatal("expected AI events")
	}
}

func fmtCard(seat int, suffix string) string {
	return fmt.Sprintf("s%d-%s", seat, suffix)
}
