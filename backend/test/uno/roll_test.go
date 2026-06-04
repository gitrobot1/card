package uno_test

import (
	"testing"

	uno "github.com/time/card/backend/internal/game/uno"
)

func TestRollForFirstUniqueWinner(t *testing.T) {
	g, err := uno.NewSoloGame("test", "玩家", 2)
	if err != nil {
		t.Fatal(err)
	}
	if g.Phase != uno.PhaseRollForFirst {
		t.Fatalf("expected roll phase, got %s", g.Phase)
	}
	g.SetRollRoundSum(0, 10)
	g.SetRollRoundSum(1, 7)
	g.SetRollRoundSum(2, 5)
	var events []uno.GameEvent
	if err := g.FinalizeRollRoundForTest(&events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != uno.PhasePlaying {
		t.Fatalf("expected playing, got %s", g.Phase)
	}
	if g.CurrentTurn != 0 {
		t.Fatalf("expected seat 0 first, got %d", g.CurrentTurn)
	}
	if len(g.Players[0].Hand) != uno.InitialHand {
		t.Fatal("expected deal after roll")
	}
	if len(events) != 1 || events[0].Type != uno.EventFirstPlayer {
		t.Fatalf("expected first_player event, got %v", events)
	}
}

func TestRollForFirstTie(t *testing.T) {
	g, _ := uno.NewSoloGame("test", "玩家", 2)
	g.SetRollRoundSum(0, 9)
	g.SetRollRoundSum(1, 9)
	g.SetRollRoundSum(2, 4)
	var events []uno.GameEvent
	if err := g.FinalizeRollRoundForTest(&events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != uno.PhaseRollForFirst {
		t.Fatal("expected still rolling")
	}
	if len(g.RollContenders()) != 2 {
		t.Fatalf("expected 2 contenders, got %v", g.RollContenders())
	}
	if g.RollRoundSum(0) != -1 || g.RollRoundSum(1) != -1 {
		t.Fatal("expected tied seats reset")
	}
	if len(events) != 1 || events[0].Type != uno.EventRollTie {
		t.Fatalf("expected roll_tie, got %v", events)
	}
}

func TestRollRoundCompletes(t *testing.T) {
	g, _ := uno.NewSoloGame("test", "玩家", 1)
	for g.Phase == uno.PhaseRollForFirst {
		var events []uno.GameEvent
		if err := g.RollRound(&events); err != nil {
			t.Fatal(err)
		}
		if len(events) == 0 {
			t.Fatal("expected events")
		}
		rollCount := 0
		for _, e := range events {
			if e.Type == uno.EventRollDice {
				rollCount++
			}
		}
		if rollCount != len(g.Players) && g.Phase == uno.PhaseRollForFirst {
			// tie round: only tied seats roll again
			tied := 0
			for _, e := range events {
				if e.Type == uno.EventRollDice {
					tied++
				}
			}
			if tied == 0 {
				t.Fatalf("expected roll events, got %v", events)
			}
		} else if g.Phase == uno.PhasePlaying && rollCount != len(g.Players) {
			t.Fatalf("expected %d roll events, got %d", len(g.Players), rollCount)
		}
	}
	if g.Phase != uno.PhasePlaying {
		t.Fatal("expected playing after all rolls")
	}
}
