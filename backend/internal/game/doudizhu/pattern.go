package doudizhu

import (
	"errors"
	"sort"

	"github.com/time/card/backend/internal/game/card"
)

type PlayType string

const (
	PlaySingle            PlayType = "single"
	PlayPair              PlayType = "pair"
	PlayTriple            PlayType = "triple"
	PlayTripleOne         PlayType = "triple_one"
	PlayTriplePair        PlayType = "triple_pair"
	PlayStraight          PlayType = "straight"
	PlayPairStraight      PlayType = "pair_straight"
	PlayTripleStraight    PlayType = "triple_straight"
	PlayBomb              PlayType = "bomb"
	PlayRocket            PlayType = "rocket"
)

type HandPattern struct {
	Type   PlayType   `json:"type"`
	Cards  []card.Card `json:"cards"`
	Weight int        `json:"weight"`
	Length int        `json:"length"`
}

var (
	errPatternInvalid = errors.New("invalid card pattern")
)

func AnalyzePattern(cards []card.Card) (*HandPattern, error) {
	if len(cards) == 0 {
		return nil, errPatternInvalid
	}

	sorted := append([]card.Card(nil), cards...)
	card.SortByRank(sorted)
	counts := rankCounts(sorted)

	if len(sorted) == 2 && counts[card.RankSJ] == 1 && counts[card.RankBJ] == 1 {
		return &HandPattern{Type: PlayRocket, Cards: sorted, Weight: 1000, Length: 2}, nil
	}

	for rank, cnt := range counts {
		if cnt == 4 {
			if len(sorted) == 4 {
				return &HandPattern{Type: PlayBomb, Cards: sorted, Weight: int(rank), Length: 4}, nil
			}
		}
		_ = rank
	}

	switch len(sorted) {
	case 1:
		return &HandPattern{Type: PlaySingle, Cards: sorted, Weight: int(sorted[0].Rank), Length: 1}, nil
	case 2:
		if counts[sorted[0].Rank] == 2 {
			return &HandPattern{Type: PlayPair, Cards: sorted, Weight: int(sorted[0].Rank), Length: 2}, nil
		}
	case 3:
		if counts[sorted[0].Rank] == 3 {
			return &HandPattern{Type: PlayTriple, Cards: sorted, Weight: int(sorted[0].Rank), Length: 3}, nil
		}
	case 4:
		if triple, ok := pickTriple(counts); ok {
			return &HandPattern{Type: PlayTripleOne, Cards: sorted, Weight: int(triple), Length: 4}, nil
		}
	case 5:
		if triple, ok := pickTriple(counts); ok && len(counts) == 2 {
			return &HandPattern{Type: PlayTriplePair, Cards: sorted, Weight: int(triple), Length: 5}, nil
		}
		if straight, ok := parseStraight(counts, 1); ok {
			return &HandPattern{Type: PlayStraight, Cards: sorted, Weight: straight, Length: len(sorted)}, nil
		}
	}

	if straight, ok := parseStraight(counts, 1); ok && len(sorted) >= 5 {
		return &HandPattern{Type: PlayStraight, Cards: sorted, Weight: straight, Length: len(sorted)}, nil
	}
	if straight, ok := parseStraight(counts, 2); ok && len(sorted) >= 6 && len(sorted)%2 == 0 {
		return &HandPattern{Type: PlayPairStraight, Cards: sorted, Weight: straight, Length: len(sorted) / 2}, nil
	}
	if straight, ok := parseStraight(counts, 3); ok && len(sorted) >= 6 && len(sorted)%3 == 0 {
		return &HandPattern{Type: PlayTripleStraight, Cards: sorted, Weight: straight, Length: len(sorted) / 3}, nil
	}

	return nil, errPatternInvalid
}

func CanBeat(current, previous *HandPattern) bool {
	if current == nil || previous == nil {
		return false
	}
	if current.Type == PlayRocket {
		return true
	}
	if previous.Type == PlayRocket {
		return false
	}
	if current.Type == PlayBomb && previous.Type != PlayBomb {
		return true
	}
	if current.Type != previous.Type {
		return false
	}
	if current.Length != previous.Length {
		return false
	}
	return current.Weight > previous.Weight
}

func rankCounts(cards []card.Card) map[card.Rank]int {
	counts := make(map[card.Rank]int, len(cards))
	for _, c := range cards {
		counts[c.Rank]++
	}
	return counts
}

func pickTriple(counts map[card.Rank]int) (card.Rank, bool) {
	for rank, cnt := range counts {
		if cnt == 3 {
			return rank, true
		}
	}
	return 0, false
}

func parseStraight(counts map[card.Rank]int, width int) (int, bool) {
	ranks := make([]int, 0, len(counts))
	for rank, cnt := range counts {
		if rank >= card.Rank2 || rank >= card.RankSJ {
			return 0, false
		}
		if cnt != width {
			return 0, false
		}
		ranks = append(ranks, int(rank))
	}
	sort.Ints(ranks)
	if len(ranks) == 0 {
		return 0, false
	}
	for i := 1; i < len(ranks); i++ {
		if ranks[i] != ranks[i-1]+1 {
			return 0, false
		}
	}
	return ranks[len(ranks)-1], true
}
