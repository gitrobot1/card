package card

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Suit string

const (
	SuitSpade   Suit = "S"
	SuitHeart   Suit = "H"
	SuitClub    Suit = "C"
	SuitDiamond Suit = "D"
	SuitJoker   Suit = "J"
)

// Rank 3-15 => 3..2, 16 small joker, 17 big joker.
type Rank int

const (
	Rank3 Rank = 3
	Rank2 Rank = 15
	RankSJ Rank = 16
	RankBJ Rank = 17
)

type Card struct {
	ID    string `json:"id"`
	Suit  Suit   `json:"suit"`
	Rank  Rank   `json:"rank"`
	Label string `json:"label"`
}

func NewDeck54() []Card {
	labels := map[Rank]string{
		3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9", 10: "10",
		11: "J", 12: "Q", 13: "K", 14: "A", 15: "2",
		RankSJ: "小王", RankBJ: "大王",
	}
	suits := []Suit{SuitSpade, SuitHeart, SuitClub, SuitDiamond}

	deck := make([]Card, 0, 54)
	for rank := Rank3; rank <= Rank2; rank++ {
		for _, suit := range suits {
			deck = append(deck, Card{
				ID:    fmt.Sprintf("%s_%d", suit, rank),
				Suit:  suit,
				Rank:  rank,
				Label: labels[rank],
			})
		}
	}
	deck = append(deck,
		Card{ID: "J_16", Suit: SuitJoker, Rank: RankSJ, Label: labels[RankSJ]},
		Card{ID: "J_17", Suit: SuitJoker, Rank: RankBJ, Label: labels[RankBJ]},
	)
	return deck
}

func Shuffle(deck []Card, seed int64) []Card {
	copied := append([]Card(nil), deck...)
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(copied), func(i, j int) {
		copied[i], copied[j] = copied[j], copied[i]
	})
	return copied
}

func ShuffleRandom(deck []Card) []Card {
	return Shuffle(deck, time.Now().UnixNano())
}

func SortByRank(cards []Card) {
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Rank == cards[j].Rank {
			return cards[i].Suit < cards[j].Suit
		}
		return cards[i].Rank < cards[j].Rank
	})
}

func ContainsAll(hand []Card, play []Card) bool {
	used := make(map[string]int, len(hand))
	for _, c := range hand {
		used[c.ID]++
	}
	for _, c := range play {
		if used[c.ID] == 0 {
			return false
		}
		used[c.ID]--
	}
	return true
}

func RemoveCards(hand []Card, play []Card) []Card {
	remove := make(map[string]int, len(play))
	for _, c := range play {
		remove[c.ID]++
	}
	result := make([]Card, 0, len(hand))
	for _, c := range hand {
		if remove[c.ID] > 0 {
			remove[c.ID]--
			continue
		}
		result = append(result, c)
	}
	return result
}

func FindByIDs(hand []Card, ids []string) ([]Card, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no cards selected")
	}

	index := make(map[string]Card, len(hand))
	available := make(map[string]int, len(hand))
	for _, c := range hand {
		index[c.ID] = c
		available[c.ID]++
	}

	used := make(map[string]int, len(ids))
	result := make([]Card, 0, len(ids))
	for _, id := range ids {
		card, ok := index[id]
		if !ok {
			return nil, fmt.Errorf("card %s not in hand", id)
		}
		if used[id] >= available[id] {
			return nil, fmt.Errorf("duplicate card %s", id)
		}
		used[id]++
		result = append(result, card)
	}
	return result, nil
}
