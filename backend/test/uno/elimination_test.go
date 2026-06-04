package uno_test

import (
	"testing"

	uno "github.com/time/card/backend/internal/game/uno"
)

func TestMultiPlayerEmptyHandEliminates(t *testing.T) {
	g := readyGame(t, 2) // 3 players
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.Players[0].Hand = []uno.Card{{ID: "c1", Color: uno.ColorRed, Value: "7", Label: "7"}}
	var events []uno.GameEvent
	if err := g.Play(0, "c1", "", &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != uno.PhasePlaying {
		t.Fatalf("expected still playing, got %s", g.Phase)
	}
	if !g.Players[0].Eliminated {
		t.Fatal("expected player 0 eliminated")
	}
	if g.Players[0].FinishRank != 1 {
		t.Fatalf("expected rank 1 for first to finish, got %d", g.Players[0].FinishRank)
	}
	if len(events) == 0 || events[len(events)-1].Type != uno.EventPlayerOut {
		t.Fatalf("expected player_out event, got %v", events)
	}
}

func TestTwoPlayerEmptyHandStillWins(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.Players[0].Hand = []uno.Card{{ID: "c1", Color: uno.ColorRed, Value: "7", Label: "7"}}
	var events []uno.GameEvent
	if err := g.Play(0, "c1", "", &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != uno.PhaseFinished {
		t.Fatal("expected finished in 2p")
	}
}

func TestVoteEndFinishesWhenAllActiveAgree(t *testing.T) {
	g := readyGame(t, 2)
	g.CanVoteToEnd = true
	g.Players[0].Hand = []uno.Card{{ID: "a", Color: uno.ColorRed, Value: "1", Label: "1"}}
	g.Players[1].Hand = []uno.Card{{ID: "b", Color: uno.ColorRed, Value: "2", Label: "2"}, {ID: "c", Color: uno.ColorRed, Value: "3", Label: "3"}}
	g.Players[2].Hand = []uno.Card{{ID: "d", Color: uno.ColorRed, Value: "4", Label: "4"}, {ID: "e", Color: uno.ColorRed, Value: "5", Label: "5"}, {ID: "f", Color: uno.ColorRed, Value: "6", Label: "6"}}
	g.SyncCountsForTest()
	var events []uno.GameEvent
	if err := g.VoteEnd(0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.VoteEnd(1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.VoteEnd(2, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != uno.PhaseFinished {
		t.Fatal("expected finished after all vote")
	}
	if g.Players[0].FinishRank != 1 {
		t.Fatalf("expected seat 0 rank 1 (fewest cards), got %d", g.Players[0].FinishRank)
	}
}

func TestLastStandingGetsWorstRank(t *testing.T) {
	g := readyGame(t, 2)
	g.Players[0].Eliminated = true
	g.Players[0].FinishRank = 1
	g.Players[1].Eliminated = true
	g.Players[1].FinishRank = 2
	g.EliminationNextRank = 3
	g.Players[2].Hand = []uno.Card{{ID: "x", Color: uno.ColorRed, Value: "1", Label: "1"}}
	g.SyncCountsForTest()
	var events []uno.GameEvent
	g.CheckAfterEliminationForTest(&events)
	if g.Phase != uno.PhaseFinished {
		t.Fatal("expected finished with one active")
	}
	if g.Players[2].FinishRank != 3 {
		t.Fatalf("expected last standing rank 3, got %d", g.Players[2].FinishRank)
	}
	if g.WinnerIndex == nil || *g.WinnerIndex != 0 {
		t.Fatalf("expected winner seat 0 (rank 1), got %v", g.WinnerIndex)
	}
}
