package engine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/time/card/backend/internal/game/card"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func cardName(kind string) string {
	switch kind {
	case CardSha:
		return "杀"
	case CardShaFire:
		return "火杀"
	case CardShaThunder:
		return "雷杀"
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
	case CardWeapon6:
		return "古锭刀"
	case CardWeapon7:
		return "朱雀羽扇"
	case CardWeapon8:
		return "雌雄双股剑"
	case CardWeapon9:
		return "贯石斧"
	case CardWeapon10:
		return "丈八蛇矛"
	case CardArmor:
		return "八卦阵"
	case CardArmorVine:
		return "藤甲"
	case CardArmorRenwang:
		return "仁王盾"
	case CardArmorBaiyin:
		return "白银狮子"
	case CardHuoGong:
		return "火攻"
	case CardTieSuo:
		return "铁索连环"
	case CardJieDao:
		return "借刀杀人"
	case CardPlusHorse:
		return "+1马"
	case CardMinusHorse:
		return "-1马"
	default:
		return kind
	}
}

// NewBasicDeck 构建 legacy 牌堆（64 张）；与 DeckProfileFor 默认配置一致。
func NewBasicDeck() []Card {
	return NewDeckForMode("")
}

// NewDeckForMode 按模式构建未洗牌的牌堆。
func NewDeckForMode(modeID string) []Card {
	return buildDeckFromProfile(mode.DeckProfileFor(modeID))
}

// trickScope 返回锦囊的作用域。
func trickScope(kind string) string {
	switch kind {
	case CardNanMan, CardWanJian:
		return TrickScopeAoe
	case CardGuoHe, CardTanNang, CardJueDou, CardLeBu, CardBingLiang, CardShanDian,
		CardWuGu, CardTaoYuan, CardWuZhong, CardWuxiek, CardHuoGong, CardTieSuo, CardJieDao:
		return TrickScopeSingle
	default:
		return ""
	}
}

// damageType 返回杀牌的伤害类型。
func damageType(kind string) string {
	switch kind {
	case CardShaFire:
		return DamageTypeFire
	case CardShaThunder:
		return DamageTypeThunder
	case CardSha:
		return DamageTypeNormal
	default:
		return ""
	}
}

func buildDeckFromProfile(profile mode.DeckProfile) []Card {
	specs := profile.Specs
	kinds := make([]string, 0, profile.TotalCards())
	for _, spec := range specs {
		for i := 0; i < spec.Count; i++ {
			kinds = append(kinds, spec.Kind)
		}
	}

	suitPool := card.ShuffleRandom(card.NewDeck52())
	deck := make([]Card, 0, len(kinds))
	for i, kind := range kinds {
		pc := suitPool[i%len(suitPool)]
		deck = append(deck, Card{
			ID:         fmt.Sprintf("%s-%d", kind, i+1),
			Kind:       kind,
			Suit:       string(pc.Suit),
			Rank:       int(pc.Rank),
			Label:      pc.Label,
			Name:       cardName(kind),
			TrickScope: trickScope(kind),
			DamageType: damageType(kind),
		})
	}
	return deck
}

func (g *Game) shuffleCards(deck []Card) []Card {
	out := append([]Card(nil), deck...)
	var r *rand.Rand
	if g != nil && g.testRand != nil {
		r = g.testRand
	} else {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
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
