package douniu

import (
	"testing"

	"github.com/time/card/backend/internal/game/card"
)

func c(id string, suit card.Suit, rank card.Rank, label string) card.Card {
	return card.Card{ID: id, Suit: suit, Rank: rank, Label: label}
}

func TestAnalyzeHandNiuNine(t *testing.T) {
	h := []card.Card{
		c("S_11", card.SuitSpade, 11, "J"),
		c("H_8", card.SuitHeart, 8, "8"),
		c("D_2", card.SuitDiamond, 15, "2"),
		c("C_13", card.SuitClub, 13, "K"),
		c("S_9", card.SuitSpade, 9, "9"),
	}
	res := AnalyzeHand(h)
	if res.Type != HandNiu9 {
		t.Fatalf("expected niu_9, got %s (%s)", res.Type, res.Label)
	}
	if res.Multiplier != 2 {
		t.Fatalf("expected mult 2, got %d", res.Multiplier)
	}
}

func TestAnalyzeHandFiveFlower(t *testing.T) {
	h := []card.Card{
		c("S_11", card.SuitSpade, 11, "J"),
		c("H_12", card.SuitHeart, 12, "Q"),
		c("D_13", card.SuitDiamond, 13, "K"),
		c("C_11", card.SuitClub, 11, "J"),
		c("S_12", card.SuitSpade, 12, "Q"),
	}
	res := AnalyzeHand(h)
	if res.Type != HandFiveFlower {
		t.Fatalf("expected five_flower, got %s", res.Type)
	}
}

func TestAnalyzeHandFiveSmall(t *testing.T) {
	h := []card.Card{
		c("S_14", card.SuitSpade, 14, "A"),
		c("H_14", card.SuitHeart, 14, "A"),
		c("D_15", card.SuitDiamond, 15, "2"),
		c("C_15", card.SuitClub, 15, "2"),
		c("S_3", card.SuitSpade, 3, "3"),
	}
	res := AnalyzeHand(h)
	if res.Type != HandFiveSmall {
		t.Fatalf("expected five_small, got %s sum check", res.Type)
	}
}

func TestHandLayoutNiuNine(t *testing.T) {
	h := []card.Card{
		c("S_11", card.SuitSpade, 11, "J"),
		c("H_8", card.SuitHeart, 8, "8"),
		c("D_2", card.SuitDiamond, 15, "2"),
		c("C_13", card.SuitClub, 13, "K"),
		c("S_9", card.SuitSpade, 9, "9"),
	}
	res := AnalyzeHand(h)
	layout := res.BuildLayout(h)
	if len(layout.HeadIDs) != 3 || len(layout.NiuIDs) != 2 {
		t.Fatalf("expected 3 head + 2 niu ids, got %d + %d", len(layout.HeadIDs), len(layout.NiuIDs))
	}
	niuSet := map[string]bool{layout.NiuIDs[0]: true, layout.NiuIDs[1]: true}
	if !niuSet["C_13"] || !niuSet["S_9"] {
		t.Fatalf("expected K and 9 in niu group, got %v", layout.NiuIDs)
	}
}

func TestCompareHands(t *testing.T) {
	flower := AnalyzeHand([]card.Card{
		c("S_11", card.SuitSpade, 11, "J"),
		c("H_12", card.SuitHeart, 12, "Q"),
		c("D_13", card.SuitDiamond, 13, "K"),
		c("C_11", card.SuitClub, 11, "J"),
		c("S_12", card.SuitSpade, 12, "Q"),
	})
	niu9 := AnalyzeHand([]card.Card{
		c("S_11", card.SuitSpade, 11, "J"),
		c("H_8", card.SuitHeart, 8, "8"),
		c("D_2", card.SuitDiamond, 15, "2"),
		c("C_13", card.SuitClub, 13, "K"),
		c("S_9", card.SuitSpade, 9, "9"),
	})
	if CompareHands(flower, niu9) <= 0 {
		t.Fatal("five flower should beat niu nine")
	}
}

func TestCarryChipsRematch(t *testing.T) {
	g, err := NewSoloGame("t1", "玩家", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Chips = 1850
	g.Players[1].Chips = 2150
	g.Phase = PhaseFinished

	chips := CarryChipsFromGame(g, []string{"玩家", "电脑1"})
	if len(chips) != 2 || chips[0] != 1850 || chips[1] != 2150 {
		t.Fatalf("unexpected carry chips: %v", chips)
	}

	next, err := NewSoloGame("t2", "玩家", 1, chips)
	if err != nil {
		t.Fatal(err)
	}
	if next.Players[0].Chips != 1850 || next.Players[1].Chips != 2150 {
		t.Fatalf("rematch chips not preserved: %d, %d", next.Players[0].Chips, next.Players[1].Chips)
	}
}

func TestNewSoloGameDeal(t *testing.T) {
	g, err := NewSoloGame("t1", "玩家", 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(g.Players))
	}
	for _, p := range g.Players {
		if len(p.Hand) != 5 {
			t.Fatalf("expected 5 cards, got %d", len(p.Hand))
		}
	}
}
