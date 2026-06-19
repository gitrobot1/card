package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// SkillLonghun 龙魂技能
// 你可以将至多两张花色相同的牌按下列规则使用或打出：
// 红桃当【桃】；方块当火【杀】；梅花当【闪】；黑桃当【无懈可击】。
// 若你以此法使用了两张红色牌，则此牌回复值或伤害值+1。
// 若你以此法使用了两张黑色牌，则你弃置当前回合角色一张牌。
func init() {
	skill.Register(skill.Decl{
		Meta: skill.Meta{
			ID:   skill.IDLonghun,
			Name: "龙魂",
			Kind: skill.KindActive,
			Desc: "你可以将至多两张花色相同的牌按下列规则使用或打出：红桃当【桃】；方块当火【杀】；梅花当【闪】；黑桃当【无懈可击】。若你以此法使用了两张红色牌，则此牌回复值或伤害值+1。若你以此法使用了两张黑色牌，则你弃置当前回合角色一张牌。",
		},
		CanActivate: longhunCanActivate,
		Activate:    longhunActivate,
		CardPlaysAs: longhunCardPlaysAs,
		AIPriority:  longhunAIPriority,
		AIActivate:  longhunAIActivate,
	})
}

// longhunCanActivate 龙魂可以在出牌阶段使用牌，或在响应阶段打出牌
func longhunCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDLonghun) {
		return false
	}

	// 出牌阶段可以使用龙魂转化的牌
	if r.Phase() == PhasePlaying && r.TurnStep() == StepPlay && r.CurrentTurn() == seat {
		return len(getLonghunPlayableCards(r, seat)) > 0
	}

	// 响应阶段可以打出龙魂转化的牌
	if r.Phase() == PhaseResponse && r.PendingTargetSeat() == seat {
		return len(getLonghunResponseCards(r, seat)) > 0
	}

	return false
}

// longhunActivate 龙魂发动：选择牌并转化为对应类型
func longhunActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return ErrInvalidCard
	}

	// 检查牌的花色是否相同
	cards := getCardsByIDs(r, seat, req.CardIDs)
	if !sameSuit(cards) {
		return fmt.Errorf("牌的花色必须相同")
	}

	// 根据花色确定转化的牌类型
	suit := cards[0].Suit
	asKind := longhunSuitToKind(suit)
	if asKind == "" {
		return fmt.Errorf("无法转化的牌花色")
	}

	// 检查是否使用了两张牌
	useTwoCards := len(cards) == 2

	// 如果是红色牌（红桃、方块），使用两张时增强效果
	isRed := suit == "H" || suit == "D"
	isBlack := suit == "S" || suit == "C"

	// 执行转化
	if r.Phase() == PhasePlaying {
		// 使用牌
		return r.UseLonghunCards(seat, req.CardIDs, asKind, useTwoCards, isRed, isBlack)
	} else if r.Phase() == PhaseResponse {
		// 打出牌
		return r.ResponseLonghunCards(seat, req.CardIDs, asKind, useTwoCards, isRed, isBlack)
	}

	return ErrWrongPhase
}

// longhunCardPlaysAs 龙魂：牌可以当其他牌使用
func longhunCardPlaysAs(r skill.Runtime, seat int, cardKind, asKind, suit string) bool {
	if !r.HasSkill(seat, skill.IDLonghun) {
		return false
	}

	// 检查是否可以转化为目标类型
	return longhunSuitToKind(suit) == asKind
}

// longhunAIPriority AI优先级
func longhunAIPriority(r skill.Runtime, seat int) int {
	if !longhunCanActivate(r, seat) {
		return 0
	}

	// 响应阶段优先级高
	if r.Phase() == PhaseResponse {
		return 85
	}

	// 出牌阶段根据情况判断
	hp, _ := r.PlayerHP(seat)
	if hp <= 2 {
		return 75 // 低血量时优先使用桃
	}

	return 60
}

// longhunAIActivate AI自动发动
func longhunAIActivate(r skill.Runtime, seat int) error {
	// 简单的AI逻辑：优先使用桃回血，其次使用闪防御
	cards := getLonghunPlayableCards(r, seat)
	if len(cards) == 0 {
		return nil
	}

	// 选择第一张可用的牌
	cardIDs := []string{cards[0].ID}
	return longhunActivate(r, seat, skill.ActivateReq{CardIDs: cardIDs})
}

// getLonghunPlayableCards 获取可以发动龙魂的牌（出牌阶段）
func getLonghunPlayableCards(r skill.Runtime, seat int) []skill.CardView {
	var result []skill.CardView

	// 获取手牌
	hand := r.PlayerHandCards(seat)
	for _, card := range hand {
		asKind := longhunSuitToKind(card.Suit)
		if asKind != "" {
			result = append(result, card)
		}
	}

	return result
}

// getLonghunResponseCards 获取可以发动龙魂的牌（响应阶段）
func getLonghunResponseCards(r skill.Runtime, seat int) []skill.CardView {
	var result []skill.CardView

	// 获取手牌
	hand := r.PlayerHandCards(seat)
	requiredKind := r.PendingRequiredKind()

	for _, card := range hand {
		asKind := longhunSuitToKind(card.Suit)
		if asKind == requiredKind {
			result = append(result, card)
		}
	}

	return result
}

// longhunSuitToKind 根据花色返回对应的牌类型
func longhunSuitToKind(suit string) string {
	switch suit {
	case "H": // 红桃 -> 桃
		return CardTao
	case "D": // 方块 -> 火杀
		return CardShaFire
	case "C": // 梅花 -> 闪
		return CardShan
	case "S": // 黑桃 -> 无懈可击
		return CardWuxiek
	default:
		return ""
	}
}

// sameSuit 检查多张牌是否花色相同
func sameSuit(cards []skill.CardView) bool {
	if len(cards) == 0 {
		return false
	}

	suit := cards[0].Suit
	for _, card := range cards[1:] {
		if card.Suit != suit {
			return false
		}
	}

	return true
}

// getCardsByIDs 根据ID获取牌对象
func getCardsByIDs(r skill.Runtime, seat int, cardIDs []string) []skill.CardView {
	hand := r.PlayerHandCards(seat)
	idSet := make(map[string]bool)
	for _, id := range cardIDs {
		idSet[id] = true
	}

	var result []skill.CardView
	for _, card := range hand {
		if idSet[card.ID] {
			result = append(result, card)
		}
	}

	return result
}
