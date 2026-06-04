package engine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/time/card/backend/internal/game/card"
)

func cardName(kind string) string {
	switch kind {
	case CardSha:
		return "杀"
	case CardShan:
		return "闪"
	case CardTao:
		return "桃"
	case CardJiu:
		return "酒"
	case CardGuoHe:
		return "过河拆桥"
	case CardTanNang:
		return "顺手牵羊"
	case CardNanMan:
		return "南蛮入侵"
	case CardWanJian:
		return "万箭齐发"
	case CardJueDou:
		return "决斗"
	case CardLeBu:
		return "乐不思蜀"
	case CardBingLiang:
		return "兵粮寸断"
	case CardShanDian:
		return "闪电"
	case CardWuGu:
		return "五谷丰登"
	case CardTaoYuan:
		return "桃园结义"
	case CardWuZhong:
		return "无中生有"
	case CardWuxiek:
		return "无懈可击"
	case CardWeapon1:
		return "诸葛连弩"
	case CardWeapon2:
		return "青釭剑"
	case CardWeapon3:
		return "青龙偃月刀"
	case CardWeapon4:
		return "方天画戟"
	case CardWeapon5:
		return "麒麟弓"
	case CardArmor:
		return "八卦阵"
	case CardPlusHorse:
		return "+1马"
	case CardMinusHorse:
		return "-1马"
	default:
		return kind
	}
}

type deckSpec struct {
	kind  string
	count int
}

// NewBasicDeck 构建牌堆；张数可超过 52，花色点数从洗乱的扑克牌循环分配。
func NewBasicDeck() []Card {
	specs := []deckSpec{
		{CardSha, 10},
		{CardShan, 4},
		{CardWuxiek, 3},
		{CardTao, 4},
		{CardJiu, 3},
		{CardGuoHe, 2},
		{CardTanNang, 2},
		{CardWuZhong, 2},
		{CardNanMan, 2},
		{CardWanJian, 2},
		{CardJueDou, 2},
		{CardLeBu, 2},
		{CardBingLiang, 2},
		{CardShanDian, 1},
		{CardWuGu, 2},
		{CardTaoYuan, 2},
		{CardWeapon1, 1},
		{CardWeapon2, 1},
		{CardWeapon3, 1},
		{CardWeapon4, 1},
		{CardWeapon5, 1},
		{CardArmor, 3},
		{CardPlusHorse, 2},
		{CardMinusHorse, 2},
	}

	kinds := make([]string, 0, 64)
	for _, spec := range specs {
		for i := 0; i < spec.count; i++ {
			kinds = append(kinds, spec.kind)
		}
	}

	suitPool := card.ShuffleRandom(card.NewDeck52())
	deck := make([]Card, 0, len(kinds))
	for i, kind := range kinds {
		pc := suitPool[i%len(suitPool)]
		deck = append(deck, Card{
			ID:    fmt.Sprintf("%s-%d", kind, i+1),
			Kind:  kind,
			Suit:  string(pc.Suit),
			Rank:  int(pc.Rank),
			Label: pc.Label,
			Name:  cardName(kind),
		})
	}
	return deck
}

func shuffleDeck(deck []Card) []Card {
	out := append([]Card(nil), deck...)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func defaultCharacter(index int) Character {
	if index == 0 {
		return Character{ID: "pioneer", Name: "先锋", MaxHP: DefaultMaxHP}
	}
	return Character{ID: "guardian", Name: "卫士", MaxHP: DefaultMaxHP}
}
