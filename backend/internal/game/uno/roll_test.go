package uno

import "testing"

func TestRollForFirstUniqueWinner(t *testing.T) {
	g, err := NewSoloGame("test", "玩家", 2)
	if err != nil {
		t.Fatal(err)
	}
	if g.Phase != PhaseRollForFirst {
		t.Fatalf("expected roll phase, got %s", g.Phase)
	}
	g.rollRoundSums[0] = 10
	g.rollRoundSums[1] = 7
	g.rollRoundSums[2] = 5
	var events []GameEvent
	if err := g.finalizeRollRound(&events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != PhasePlaying {
		t.Fatalf("expected playing, got %s", g.Phase)
	}
	if g.CurrentTurn != 0 {
		t.Fatalf("expected seat 0 first, got %d", g.CurrentTurn)
	}
	if len(g.Players[0].Hand) != InitialHand {
		t.Fatal("expected deal after roll")
	}
	if len(events) != 1 || events[0].Type != EventFirstPlayer {
		t.Fatalf("expected first_player event, got %v", events)
	}
}

func TestRollForFirstTie(t *testing.T) {
	g, _ := NewSoloGame("test", "玩家", 2)
	g.rollRoundSums[0] = 9
	g.rollRoundSums[1] = 9
	g.rollRoundSums[2] = 4
	var events []GameEvent
	if err := g.finalizeRollRound(&events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != PhaseRollForFirst {
		t.Fatal("expected still rolling")
	}
	if len(g.rollContenders) != 2 {
		t.Fatalf("expected 2 contenders, got %v", g.rollContenders)
	}
	if g.rollRoundSums[0] != -1 || g.rollRoundSums[1] != -1 {
		t.Fatal("expected tied seats reset")
	}
	if len(events) != 1 || events[0].Type != EventRollTie {
		t.Fatalf("expected roll_tie, got %v", events)
	}
}

func TestRollRoundCompletes(t *testing.T) {
	g, _ := NewSoloGame("test", "玩家", 1)
	for g.Phase == PhaseRollForFirst {
		var events []GameEvent
		if err := g.RollRound(&events); err != nil {
			t.Fatal(err)
		}
		if len(events) == 0 {
			t.Fatal("expected events")
		}
		rollCount := 0
		for _, e := range events {
			if e.Type == EventRollDice {
				rollCount++
			}
		}
		if rollCount != len(g.Players) && g.Phase == PhaseRollForFirst {
			// tie round: only tied seats roll again
			tied := 0
			for _, e := range events {
				if e.Type == EventRollDice {
					tied++
				}
			}
			if tied == 0 {
				t.Fatalf("expected roll events, got %v", events)
			}
		} else if g.Phase == PhasePlaying && rollCount != len(g.Players) {
			t.Fatalf("expected %d roll events, got %d", len(g.Players), rollCount)
		}
	}
	if g.Phase != PhasePlaying {
		t.Fatal("expected playing after all rolls")
	}
}
