package douniu

import (
	"fmt"
	"sort"

	"github.com/time/card/backend/internal/game/card"
)

const (
	HandFiveSmall  = "five_small"
	HandBomb       = "bomb"
	HandFiveFlower = "five_flower"
	HandNiuNiu     = "niu_niu"
	HandNiu9       = "niu_9"
	HandNiu8       = "niu_8"
	HandNiu7       = "niu_7"
	HandNiu6       = "niu_6"
	HandNiu5       = "niu_5"
	HandNiu4       = "niu_4"
	HandNiu3       = "niu_3"
	HandNiu2       = "niu_2"
	HandNiu1       = "niu_1"
	HandNone       = "none"
)

var HandLabels = map[string]string{
	HandFiveSmall:  "五小牛",
	HandBomb:       "炸弹牛",
	HandFiveFlower: "五花牛",
	HandNiuNiu:     "牛牛",
	HandNiu9:       "牛九",
	HandNiu8:       "牛八",
	HandNiu7:       "牛七",
	HandNiu6:       "牛六",
	HandNiu5:       "牛五",
	HandNiu4:       "牛四",
	HandNiu3:       "牛三",
	HandNiu2:       "牛二",
	HandNiu1:       "牛一",
	HandNone:       "没牛",
}

type HandResult struct {
	Type         string
	Label        string
	NiuPoint     int
	Multiplier   int
	MaxRank      card.Rank
	HeadIndices  []int
	NiuIndices   []int
}

type HandLayout struct {
	HeadIDs []string `json:"head_ids"`
	NiuIDs  []string `json:"niu_ids,omitempty"`
}

func allHandIndices(n int) []int {
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	return idx
}

func sortIndicesByRank(hand []card.Card, indices []int) []int {
	sorted := append([]int(nil), indices...)
	sort.Slice(sorted, func(i, j int) bool {
		return rankOrder(hand[sorted[i]].Rank) > rankOrder(hand[sorted[j]].Rank)
	})
	return sorted
}

func indicesToIDs(hand []card.Card, indices []int) []string {
	ids := make([]string, len(indices))
	for i, idx := range indices {
		ids[i] = hand[idx].ID
	}
	return ids
}

func (r HandResult) BuildLayout(hand []card.Card) HandLayout {
	if len(hand) != 5 {
		return HandLayout{}
	}
	if len(r.NiuIndices) == 0 {
		sorted := sortIndicesByRank(hand, allHandIndices(len(hand)))
		return HandLayout{HeadIDs: indicesToIDs(hand, sorted)}
	}
	head := sortIndicesByRank(hand, r.HeadIndices)
	niu := sortIndicesByRank(hand, r.NiuIndices)
	return HandLayout{
		HeadIDs: indicesToIDs(hand, head),
		NiuIDs:  indicesToIDs(hand, niu),
	}
}

func LayoutForHand(hand []card.Card) HandLayout {
	return AnalyzeHand(hand).BuildLayout(hand)
}

func niuPointValue(c card.Card) int {
	switch {
	case c.Rank >= 11 && c.Rank <= 13:
		return 10
	case c.Rank == 10:
		return 10
	case c.Rank == 14:
		return 1
	case c.Rank == 15:
		return 2
	case c.Rank >= 3 && c.Rank <= 9:
		return int(c.Rank)
	default:
		return 0
	}
}

func smallNiuPoint(c card.Card) int {
	switch c.Rank {
	case 14:
		return 1
	case 15:
		return 2
	default:
		return int(c.Rank)
	}
}

func isFiveSmall(hand []card.Card) bool {
	if len(hand) != 5 {
		return false
	}
	sum := 0
	for _, c := range hand {
		if c.Rank != 14 && c.Rank != 15 && (c.Rank < 3 || c.Rank > 5) {
			return false
		}
		sum += smallNiuPoint(c)
	}
	return sum <= 10
}

func isBomb(hand []card.Card) bool {
	counts := map[card.Rank]int{}
	for _, c := range hand {
		counts[c.Rank]++
		if counts[c.Rank] >= 4 {
			return true
		}
	}
	return false
}

func isFiveFlower(hand []card.Card) bool {
	for _, c := range hand {
		if c.Rank < 11 || c.Rank > 13 {
			return false
		}
	}
	return true
}

func maxCardRank(hand []card.Card) card.Rank {
	if len(hand) == 0 {
		return 0
	}
	sorted := append([]card.Card(nil), hand...)
	sort.Slice(sorted, func(i, j int) bool {
		return compareRank(sorted[i].Rank, sorted[j].Rank) > 0
	})
	return sorted[0].Rank
}

func compareRank(a, b card.Rank) int {
	return rankOrder(a) - rankOrder(b)
}

func rankOrder(r card.Rank) int {
	switch r {
	case 15:
		return 14
	case 13:
		return 13
	case 12:
		return 12
	case 11:
		return 11
	case 10:
		return 10
	case 14:
		return 1
	default:
		return int(r)
	}
}

func AnalyzeHand(hand []card.Card) HandResult {
	if len(hand) != 5 {
		return HandResult{Type: HandNone, Label: HandLabels[HandNone], Multiplier: 1}
	}
	maxRank := maxCardRank(hand)

	if isFiveSmall(hand) {
		return HandResult{Type: HandFiveSmall, Label: HandLabels[HandFiveSmall], NiuPoint: 10, Multiplier: 6, MaxRank: maxRank, HeadIndices: allHandIndices(len(hand))}
	}
	if isBomb(hand) {
		return HandResult{Type: HandBomb, Label: HandLabels[HandBomb], NiuPoint: 10, Multiplier: 5, MaxRank: maxRank, HeadIndices: allHandIndices(len(hand))}
	}
	if isFiveFlower(hand) {
		return HandResult{Type: HandFiveFlower, Label: HandLabels[HandFiveFlower], NiuPoint: 10, Multiplier: 4, MaxRank: maxRank, HeadIndices: allHandIndices(len(hand))}
	}

	points := make([]int, 5)
	for i, c := range hand {
		points[i] = niuPointValue(c)
	}

	bestNiu := -1
	var bestTriple [3]int
	found := false
	for a := 0; a < 3; a++ {
		for b := a + 1; b < 4; b++ {
			for c := b + 1; c < 5; c++ {
				sum3 := points[a] + points[b] + points[c]
				if sum3%10 != 0 {
					continue
				}
				rest := 0
				for i := 0; i < 5; i++ {
					if i != a && i != b && i != c {
						rest += points[i]
					}
				}
				niu := rest % 10
				if niu > bestNiu {
					bestNiu = niu
					bestTriple = [3]int{a, b, c}
					found = true
				}
			}
		}
	}

	if !found {
		return HandResult{
			Type:        HandNone,
			Label:       HandLabels[HandNone],
			NiuPoint:    0,
			Multiplier:  1,
			MaxRank:     maxRank,
			HeadIndices: allHandIndices(len(hand)),
		}
	}

	niuIndices := make([]int, 0, 2)
	for i := 0; i < 5; i++ {
		if i != bestTriple[0] && i != bestTriple[1] && i != bestTriple[2] {
			niuIndices = append(niuIndices, i)
		}
	}

	handType, mult := niuTypeAndMultiplier(bestNiu)
	return HandResult{
		Type:        handType,
		Label:       HandLabels[handType],
		NiuPoint:    bestNiu,
		Multiplier:  mult,
		MaxRank:     maxRank,
		HeadIndices: []int{bestTriple[0], bestTriple[1], bestTriple[2]},
		NiuIndices:  niuIndices,
	}
}

func niuTypeAndMultiplier(niu int) (string, int) {
	switch niu {
	case 0:
		return HandNiuNiu, 3
	case 9:
		return HandNiu9, 2
	case 8:
		return HandNiu8, 2
	case 7:
		return HandNiu7, 2
	case 6:
		return HandNiu6, 1
	case 5:
		return HandNiu5, 1
	case 4:
		return HandNiu4, 1
	case 3:
		return HandNiu3, 1
	case 2:
		return HandNiu2, 1
	case 1:
		return HandNiu1, 1
	default:
		return HandNone, 1
	}
}

func typeRank(t string) int {
	switch t {
	case HandFiveSmall:
		return 9
	case HandBomb:
		return 8
	case HandFiveFlower:
		return 7
	case HandNiuNiu:
		return 6
	case HandNiu9:
		return 5
	case HandNiu8:
		return 4
	case HandNiu7:
		return 3
	case HandNiu6:
		return 2
	case HandNiu5:
		return 1
	case HandNiu4:
		return 1
	case HandNiu3:
		return 1
	case HandNiu2:
		return 1
	case HandNiu1:
		return 1
	default:
		return 0
	}
}

// CompareHands returns >0 if a beats b, <0 if b beats a, 0 tie.
func CompareHands(a, b HandResult) int {
	if tr := typeRank(a.Type) - typeRank(b.Type); tr != 0 {
		return tr
	}
	if a.NiuPoint != b.NiuPoint {
		return a.NiuPoint - b.NiuPoint
	}
	return compareRank(a.MaxRank, b.MaxRank)
}

func HandMultipliersTable() map[string]int {
	return map[string]int{
		HandFiveSmall:  6,
		HandBomb:       5,
		HandFiveFlower: 4,
		HandNiuNiu:     3,
		HandNiu9:       2,
		HandNiu8:       2,
		HandNiu7:       2,
		HandNiu6:       1,
		HandNiu5:       1,
		HandNiu4:       1,
		HandNiu3:       1,
		HandNiu2:       1,
		HandNiu1:       1,
		HandNone:       1,
	}
}

func validGrab(mult int) bool {
	return mult >= 0 && mult <= MaxGrabMult
}

func validBet(mult int) bool {
	for _, v := range BetOptions {
		if v == mult {
			return true
		}
	}
	return false
}

func formatSettle(playerName string, amount int, win bool) string {
	if win {
		return fmt.Sprintf("%s 赢 %d", playerName, amount)
	}
	return fmt.Sprintf("%s 输 %d", playerName, amount)
}

func formatGrabMessage(name string, mult int) string {
	if mult == 0 {
		return fmt.Sprintf("%s 不抢", name)
	}
	return fmt.Sprintf("%s 抢庄 ×%d", name, mult)
}
