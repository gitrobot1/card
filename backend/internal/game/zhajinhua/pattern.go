package zhajinhua

import (
	"fmt"
	"sort"

	"github.com/time/card/backend/internal/game/card"
)

// HandPattern 三张牌分析结果
type HandPattern struct {
	Type       HandType    `json:"type"`
	TypeLabel  string      `json:"type_label"`
	Multiplier int         `json:"multiplier"`
	Weight     int         `json:"weight"`
	HighRanks  []int       `json:"high_ranks"`
	SuitWeight int         `json:"-"`
	Cards      []card.Card `json:"-"`
}

var typeLabels = map[HandType]string{
	Hand235:           "235",
	HandLeopard:       "豹子",
	HandStraightFlush: "顺金",
	HandFlush:         "金花",
	HandStraight:      "顺子",
	HandPair:          "对子",
	HandHighCard:      "单牌",
}

// AnalyzeHand 分析三张牌（必须恰好 3 张）
func AnalyzeHand(cards []card.Card) (*HandPattern, error) {
	if len(cards) != 3 {
		return nil, fmt.Errorf("zhajinhua hand must have 3 cards")
	}
	c := append([]card.Card(nil), cards...)
	sort.Slice(c, func(i, j int) bool {
		if c[i].Rank == c[j].Rank {
			return suitOrder(c[i].Suit) > suitOrder(c[j].Suit)
		}
		return c[i].Rank > c[j].Rank
	})

	ranks := []int{int(c[0].Rank), int(c[1].Rank), int(c[2].Rank)}
	suits := []card.Suit{c[0].Suit, c[1].Suit, c[2].Suit}
	suitW := maxSuit(suits)

	if is235(ranks, suits) {
		return &HandPattern{
			Type: Hand235, TypeLabel: typeLabels[Hand235], Multiplier: MultiplierFor(Hand235),
			Weight: 235, HighRanks: ranks, SuitWeight: suitW, Cards: c,
		}, nil
	}

	if ranks[0] == ranks[1] && ranks[1] == ranks[2] {
		return &HandPattern{
			Type: HandLeopard, TypeLabel: typeLabels[HandLeopard], Multiplier: MultiplierFor(HandLeopard),
			Weight: ranks[0], HighRanks: ranks, SuitWeight: suitW, Cards: c,
		}, nil
	}

	sameSuit := suits[0] == suits[1] && suits[1] == suits[2]
	strWeight, isStraight := straightWeight(ranks)

	if sameSuit && isStraight {
		return &HandPattern{
			Type: HandStraightFlush, TypeLabel: typeLabels[HandStraightFlush], Multiplier: MultiplierFor(HandStraightFlush),
			Weight: strWeight, HighRanks: ranks, SuitWeight: suitW, Cards: c,
		}, nil
	}
	if sameSuit {
		return &HandPattern{
			Type: HandFlush, TypeLabel: typeLabels[HandFlush], Multiplier: MultiplierFor(HandFlush),
			Weight: highCardWeight(ranks), HighRanks: ranks, SuitWeight: suitW, Cards: c,
		}, nil
	}
	if isStraight {
		return &HandPattern{
			Type: HandStraight, TypeLabel: typeLabels[HandStraight], Multiplier: MultiplierFor(HandStraight),
			Weight: strWeight, HighRanks: ranks, SuitWeight: suitW, Cards: c,
		}, nil
	}

	if ranks[0] == ranks[1] || ranks[1] == ranks[2] {
		pair, kicker := pairRanks(ranks)
		return &HandPattern{
			Type: HandPair, TypeLabel: typeLabels[HandPair], Multiplier: MultiplierFor(HandPair),
			Weight: pair*100 + kicker, HighRanks: ranks, SuitWeight: suitW, Cards: c,
		}, nil
	}

	return &HandPattern{
		Type: HandHighCard, TypeLabel: typeLabels[HandHighCard], Multiplier: MultiplierFor(HandHighCard),
		Weight: highCardWeight(ranks), HighRanks: ranks, SuitWeight: suitW, Cards: c,
	}, nil
}

// CompareHands 比较两手牌。a 赢返回 1，b 赢返回 -1，平返回 0
func CompareHands(a, b *HandPattern) int {
	if a == nil || b == nil {
		return 0
	}
	if a.Type == Hand235 && b.Type == HandLeopard {
		return 1
	}
	if b.Type == Hand235 && a.Type == HandLeopard {
		return -1
	}

	ta, tb := a.Type, b.Type
	if a.Type == Hand235 {
		ta = HandHighCard
	}
	if b.Type == Hand235 {
		tb = HandHighCard
	}

	ra, rb := HandTypeRank(ta), HandTypeRank(tb)
	if ra != rb {
		if ra > rb {
			return 1
		}
		return -1
	}
	if a.Weight != b.Weight {
		if a.Weight > b.Weight {
			return 1
		}
		return -1
	}
	if a.SuitWeight != b.SuitWeight {
		if a.SuitWeight > b.SuitWeight {
			return 1
		}
		return -1
	}
	return 0
}

func is235(ranks []int, suits []card.Suit) bool {
	if suits[0] == suits[1] || suits[1] == suits[2] || suits[0] == suits[2] {
		return false
	}
	has2, has3, has5 := false, false, false
	for _, r := range ranks {
		switch r {
		case int(card.Rank2):
			has2 = true
		case int(card.Rank3):
			has3 = true
		case 5:
			has5 = true
		}
	}
	return has2 && has3 && has5
}

func straightWeight(ranks []int) (int, bool) {
	norm := normalizeRanks(ranks)
	sort.Ints(norm)

	// A-2-3 最小顺子
	if norm[0] == 2 && norm[1] == 3 && norm[2] == 14 {
		return 3, true
	}
	if norm[2]-norm[0] == 2 && norm[1]-norm[0] == 1 {
		return norm[2], true
	}
	return 0, false
}

func normalizeRanks(ranks []int) []int {
	out := make([]int, len(ranks))
	for i, r := range ranks {
		if r == int(card.Rank2) {
			out[i] = 2
		} else {
			out[i] = r
		}
	}
	return out
}

func pairRanks(ranks []int) (pair, kicker int) {
	if ranks[0] == ranks[1] {
		return ranks[0], ranks[2]
	}
	return ranks[1], ranks[0]
}

func highCardWeight(ranks []int) int {
	return ranks[0]*10000 + ranks[1]*100 + ranks[2]
}

func suitOrder(s card.Suit) int {
	switch s {
	case card.SuitSpade:
		return 4
	case card.SuitHeart:
		return 3
	case card.SuitClub:
		return 2
	case card.SuitDiamond:
		return 1
	default:
		return 0
	}
}

func maxSuit(suits []card.Suit) int {
	m := 0
	for _, s := range suits {
		if v := suitOrder(s); v > m {
			m = v
		}
	}
	return m
}
