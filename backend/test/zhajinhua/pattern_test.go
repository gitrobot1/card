package zhajinhua_test

import (
	"testing"

	zhajinhua "github.com/time/card/backend/internal/game/zhajinhua"
	"github.com/time/card/backend/internal/game/card"
)

func c(suit card.Suit, rank card.Rank, label string) card.Card {
	return card.Card{ID: string(suit) + label, Suit: suit, Rank: rank, Label: label}
}

func TestAnalyzeLeopard(t *testing.T) {
	p, err := zhajinhua.AnalyzeHand([]card.Card{
		c(card.SuitSpade, 14, "A"),
		c(card.SuitHeart, 14, "A"),
		c(card.SuitClub, 14, "A"),
	})
	if err != nil || p.Type != zhajinhua.HandLeopard || p.Multiplier != 10 {
		t.Fatalf("expected leopard x10, got %+v err=%v", p, err)
	}
}

func Test235BeatsLeopard(t *testing.T) {
	p235, _ := zhajinhua.AnalyzeHand([]card.Card{
		c(card.SuitSpade, card.Rank2, "2"),
		c(card.SuitHeart, 3, "3"),
		c(card.SuitClub, 5, "5"),
	})
	pLeo, _ := zhajinhua.AnalyzeHand([]card.Card{
		c(card.SuitSpade, 14, "A"),
		c(card.SuitHeart, 14, "A"),
		c(card.SuitClub, 14, "A"),
	})
	if zhajinhua.CompareHands(p235, pLeo) != 1 {
		t.Fatal("235 should beat leopard")
	}
}

func TestStraightFlushBeatsFlush(t *testing.T) {
	sf, _ := zhajinhua.AnalyzeHand([]card.Card{
		c(card.SuitSpade, 14, "A"),
		c(card.SuitSpade, 13, "K"),
		c(card.SuitSpade, 12, "Q"),
	})
	fl, _ := zhajinhua.AnalyzeHand([]card.Card{
		c(card.SuitHeart, 14, "A"),
		c(card.SuitHeart, 11, "J"),
		c(card.SuitHeart, 9, "9"),
	})
	if zhajinhua.CompareHands(sf, fl) != 1 {
		t.Fatal("straight flush should beat flush")
	}
}

func TestPairCompare(t *testing.T) {
	pA, _ := zhajinhua.AnalyzeHand([]card.Card{
		c(card.SuitSpade, 13, "K"),
		c(card.SuitHeart, 13, "K"),
		c(card.SuitClub, 14, "A"),
	})
	pB, _ := zhajinhua.AnalyzeHand([]card.Card{
		c(card.SuitSpade, 12, "Q"),
		c(card.SuitHeart, 12, "Q"),
		c(card.SuitClub, 14, "A"),
	})
	if zhajinhua.CompareHands(pA, pB) != 1 {
		t.Fatal("KK should beat QQ")
	}
}

func TestMultipliers(t *testing.T) {
	if zhajinhua.MultiplierFor(zhajinhua.HandLeopard) != 10 || zhajinhua.MultiplierFor(zhajinhua.HandHighCard) != 1 {
		t.Fatal("unexpected multipliers")
	}
}
