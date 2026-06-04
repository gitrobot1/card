package doudizhu_test

import (
	"testing"

	doudizhu "github.com/time/card/backend/internal/game/doudizhu"
	"github.com/time/card/backend/internal/game/card"
)

func TestDebugOneCardDealStartsPlaying(t *testing.T) {
	if !doudizhu.DebugDealOneCardEnabled() {
		t.Skip("doudizhu.DebugDealOneCardEnabled() disabled")
	}

	g := doudizhu.NewGame("t", "me")
	if g.Phase != doudizhu.PhasePlaying {
		t.Fatalf("expected playing phase, got %s", g.Phase)
	}
	for i := 0; i < 3; i++ {
		if len(g.Players[i].Hand) != 1 {
			t.Fatalf("player %d expected 1 card, got %d", i, len(g.Players[i].Hand))
		}
	}
	if len(g.BottomCards) != 0 {
		t.Fatalf("expected no bottom cards, got %d", len(g.BottomCards))
	}
	if g.CurrentTurn != 0 {
		t.Fatalf("expected human to lead, got turn %d", g.CurrentTurn)
	}
}

func TestPlayLastPairWins(t *testing.T) {
	g := doudizhu.NewGame("t", "me")
	g.Phase = doudizhu.PhasePlaying
	g.CurrentTurn = 0
	g.LeaderIndex = 0
	g.LastPlay = nil
	g.PassCount = 0
	g.Players[0].Hand = []card.Card{
		{ID: "C_12", Suit: card.SuitClub, Rank: card.Rank(12), Label: "Q"},
		{ID: "S_12", Suit: card.SuitSpade, Rank: card.Rank(12), Label: "Q"},
	}

	_, err := g.Play(0, []string{"C_12", "S_12"})
	if err != nil {
		t.Fatalf("play failed: %v", err)
	}
	if g.Phase != doudizhu.PhaseFinished {
		t.Fatalf("expected finished phase, got %s", g.Phase)
	}
	if g.WinnerIndex == nil || *g.WinnerIndex != 0 {
		t.Fatalf("expected human winner 0, got %v", g.WinnerIndex)
	}
	if len(g.Players[0].Hand) != 0 {
		t.Fatalf("expected empty hand, got %d", len(g.Players[0].Hand))
	}
}

func TestFindByIDsRejectsDuplicateIDs(t *testing.T) {
	hand := []card.Card{
		{ID: "C_12", Suit: card.SuitClub, Rank: card.Rank(12), Label: "Q"},
	}
	_, err := card.FindByIDs(hand, []string{"C_12", "C_12"})
	if err == nil {
		t.Fatal("expected duplicate id error")
	}
}

func TestPlayDuplicateCardIDRejected(t *testing.T) {
	g := doudizhu.NewGame("t", "me")
	g.Phase = doudizhu.PhasePlaying
	g.CurrentTurn = 0
	g.LastPlay = nil
	g.Players[0].Hand = []card.Card{
		{ID: "C_12", Suit: card.SuitClub, Rank: card.Rank(12), Label: "Q"},
		{ID: "S_12", Suit: card.SuitSpade, Rank: card.Rank(12), Label: "Q"},
	}

	_, err := g.Play(0, []string{"C_12", "C_12"})
	if err == nil {
		t.Fatal("expected duplicate id play to fail")
	}
	if g.Phase == doudizhu.PhaseFinished {
		t.Fatal("duplicate id play should not finish game")
	}
	if len(g.Players[0].Hand) != 2 {
		t.Fatalf("expected 2 cards left, got %d", len(g.Players[0].Hand))
	}
}
