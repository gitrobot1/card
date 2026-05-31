package zhajinhua

// HandType 炸金花牌型（与斗地主 pattern 完全独立）
type HandType string

const (
	Hand235           HandType = "235"            // 豹子杀手：花色不同的 235
	HandLeopard       HandType = "leopard"        // 豹子
	HandStraightFlush HandType = "straight_flush" // 顺金
	HandFlush         HandType = "flush"          // 金花
	HandStraight      HandType = "straight"       // 顺子
	HandPair          HandType = "pair"           // 对子
	HandHighCard      HandType = "high_card"      // 单牌
)

// HandMultipliers 获胜牌型倍率（结算时：底池有效部分 × 倍率）
var HandMultipliers = map[HandType]int{
	Hand235:           12,
	HandLeopard:       10,
	HandStraightFlush: 6,
	HandFlush:         4,
	HandStraight:      3,
	HandPair:          2,
	HandHighCard:      1,
}

func MultiplierFor(t HandType) int {
	if m, ok := HandMultipliers[t]; ok {
		return m
	}
	return 1
}

func HandTypeRank(t HandType) int {
	switch t {
	case Hand235:
		return 0 // 235 仅在比豹子时特殊，常规排序按散牌
	case HandLeopard:
		return 6
	case HandStraightFlush:
		return 5
	case HandFlush:
		return 4
	case HandStraight:
		return 3
	case HandPair:
		return 2
	default:
		return 1
	}
}
