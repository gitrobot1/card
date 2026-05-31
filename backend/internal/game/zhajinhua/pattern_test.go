package zhajinhua

import (
	"testing"

	"github.com/time/card/backend/internal/game/card"
)

func c(suit card.Suit, rank card.Rank, label string) card.Card {
	return card.Card{ID: string(suit) + label, Suit: suit, Rank: rank, Label: label}
}

func TestAnalyzeLeopard(t *testing.T) {
	p, err := AnalyzeHand([]card.Card{
		c(card.SuitSpade, 14, "A"),
		c(card.SuitHeart, 14, "A"),
		c(card.SuitClub, 14, "A"),
	})
	if err != nil || p.Type != HandLeopard || p.Multiplier != 10 {
		t.Fatalf("expected leopard x10, got %+v err=%v", p, err)
	}
}

func Test235BeatsLeopard(t *testing.T) {
	p235, _ := AnalyzeHand([]card.Card{
		c(card.SuitSpade, card.Rank2, "2"),
		c(card.SuitHeart, 3, "3"),
		c(card.SuitClub, 5, "5"),
	})
	pLeo, _ := AnalyzeHand([]card.Card{
		c(card.SuitSpade, 14, "A"),
		c(card.SuitHeart, 14, "A"),
		c(card.SuitClub, 14, "A"),
	})
	if CompareHands(p235, pLeo) != 1 {
		t.Fatal("235 should beat leopard")
	}
}

func TestStraightFlushBeatsFlush(t *testing.T) {
	sf, _ := AnalyzeHand([]card.Card{
		c(card.SuitSpade, 14, "A"),
		c(card.SuitSpade, 13, "K"),
		c(card.SuitSpade, 12, "Q"),
	})
	fl, _ := AnalyzeHand([]card.Card{
		c(card.SuitHeart, 14, "A"),
		c(card.SuitHeart, 11, "J"),
		c(card.SuitHeart, 9, "9"),
	})
	if CompareHands(sf, fl) != 1 {
		t.Fatal("straight flush should beat flush")
	}
}

func TestPairCompare(t *testing.T) {
	pA, _ := AnalyzeHand([]card.Card{
		c(card.SuitSpade, 13, "K"),
		c(card.SuitHeart, 13, "K"),
		c(card.SuitClub, 14, "A"),
	})
	pB, _ := AnalyzeHand([]card.Card{
		c(card.SuitSpade, 12, "Q"),
		c(card.SuitHeart, 12, "Q"),
		c(card.SuitClub, 14, "A"),
	})
	if CompareHands(pA, pB) != 1 {
		t.Fatal("KK should beat QQ")
	}
}

func TestMultipliers(t *testing.T) {
	if MultiplierFor(HandLeopard) != 10 || MultiplierFor(HandHighCard) != 1 {
		t.Fatal("unexpected multipliers")
	}
}
