package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterLimuUsed = "limu_used" // 立牧本回合是否已使用
)

// 图射技能
// 当你使用非装备牌指定目标后，若你没有基本牌，则你可以摸X张牌（X为此牌指定的目标数）。

// tryTriggerTushe 尝试触发图射技能
func (g *Game) tryTriggerTushe(source int, card Card, target int, events *[]GameEvent) {
	if !g.hasSkill(source, SkillTushe) {
		return
	}
	
	// 检查是否为非装备牌
	if isEquipCard(card.Kind) {
		return
	}
	
	// 检查是否有基本牌
	if hasBasicCard(g, source) {
		return
	}
	
	// 计算目标数
	targetCount := g.countTrickTargets(source, card.Kind)
	if targetCount <= 0 {
		targetCount = 1 // 至少为目标自己
	}
	
	// 摸X张牌（X为目标数）
	g.drawCards(source, targetCount, events)
	g.appendSkillEvent(events, SkillTushe, source, -1, 
		fmt.Sprintf("%s 发动【图射】，摸 %d 张牌", g.Players[source].Name, targetCount))
}

// countTrickTargets 计算锦囊的目标数
func (g *Game) countTrickTargets(source int, trickKind string) int {
	switch trickKind {
	case CardNanMan, CardWanJian, CardTaoYuan:
		// AOE 锦囊，目标为除自己外的所有存活角色
		count := 0
		for i := range g.Players {
			if i != source && g.Players[i].HP > 0 {
				count++
			}
		}
		return count
	case CardWuGu:
		// 五谷丰登，目标为所有角色（包括自己）
		count := 0
		for i := range g.Players {
			if g.Players[i].HP > 0 {
				count++
			}
		}
		return count
	default:
		// 其他锦囊，目标为1人
		return 1
	}
}

// tusheCanActivate 图射：被动技能，不需要主动激活
func tusheCanActivate(r skill.Runtime, seat int) bool {
	return false // 被动技能
}

// tusheActivate 图射：被动技能，不需要主动激活
func tusheActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	return nil
}

// tusheOnCardUsed 图射：使用非装备牌指定目标后触发
// 注意：这个函数在 playTrick 等函数中调用
func tusheOnCardUsed(g *Game, source int, card Card, targetCount int, events *[]GameEvent) {
	if !g.hasSkill(source, SkillTushe) {
		return
	}
	
	// 检查是否为非装备牌
	if isEquipCard(card.Kind) {
		return
	}
	
	// 检查是否有基本牌
	if hasBasicCard(g, source) {
		return
	}
	
	// 摸X张牌（X为目标数）
	if targetCount > 0 {
		g.drawCards(source, targetCount, events)
		g.appendSkillEvent(events, skill.IDTushe, source, -1, 
			fmt.Sprintf("%s 发动【图射】，摸 %d 张牌", g.Players[source].Name, targetCount))
	}
}

// 立牧技能
// 出牌阶段，你可以将一张方块牌当【乐不思蜀】对自己使用，然后回复1点体力；
// 你的判定区有牌时，你对攻击范围内的其他角色使用牌没有次数和距离限制。

// limuCanActivate 立牧：出牌阶段可以激活
func limuCanActivate(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDLimu) {
		return false
	}
	// 出牌阶段才能激活
	if r.Phase() != PhasePlaying || r.TurnStep() != StepPlay || r.CurrentTurn() != seat {
		return false
	}
	// 检查是否有方块牌（手牌）
	g := r.(*gameSkillRuntime).g
	return hasDiamondCard(g, seat)
}

// limuActivate 立牧：将一张方块牌当【乐不思蜀】对自己使用，然后回复1点体力
func limuActivate(r skill.Runtime, seat int, req skill.ActivateReq) error {
	if len(req.CardIDs) == 0 {
		return ErrInvalidCard
	}
	
	g := r.(*gameSkillRuntime).g
	events := r.(*gameSkillRuntime).events
	
	// 找到这张方块牌
	idx, cardObj, ok := g.findCard(seat, req.CardIDs[0])
	if !ok {
		return ErrInvalidCard
	}
	
	// 检查是否为方块牌
	if cardObj.Suit != "diamond" {
		return ErrInvalidCard
	}
	
	// 从手牌或装备区移除牌
	var removed Card
	if idx >= 0 && idx < len(g.Players[seat].Hand) {
		removed = g.removeHandCard(seat, idx, events)
	} else {
		// 从装备区移除
		// 简化：假设从手牌移除
		return ErrInvalidCard
	}
	
	// 将牌当乐不思蜀对自己使用（需要转换 Kind）
	lebuCard := Card{
		ID:    removed.ID,
		Kind:  CardLeBu,
		Name:  "乐不思蜀",
		Suit:  removed.Suit,
		Rank:  removed.Rank,
		Label: removed.Label,
	}
	g.placeLebu(seat, seat, lebuCard, events)
	
	// 回复1点体力
	if g.Players[seat].HP < g.Players[seat].MaxHP {
		g.Players[seat].HP++
		*events = append(*events, GameEvent{
			Type:        "skill_heal",
			PlayerIndex: seat,
			TargetIndex: seat,
			SkillID:     skill.IDLimu,
			Message:     fmt.Sprintf("%s 发动【立牧】，回复1点体力", g.Players[seat].Name),
		})
	}
	
	// 标记本回合已使用立牧
	g.setSkillCounter(seat, counterLimuUsed, 1)
	
	g.appendSkillEvent(events, skill.IDLimu, seat, seat, 
		fmt.Sprintf("%s 发动【立牧】", g.Players[seat].Name))
	g.resetTimer()
	
	return nil
}

// limuCardPlaysAs 立牧：方块牌可以当乐不思蜀使用（对自己）
func limuCardPlaysAs(r skill.Runtime, seat int, cardKind, asKind, suit string) bool {
	if !r.HasSkill(seat, skill.IDLimu) || asKind != CardLeBu || suit != "diamond" {
		return false
	}
	// 只要立牧可以使用，方块牌就视为乐不思蜀
	return limuCanActivate(r, seat)
}

// limuUnlimitedSha 立牧：判定区有牌时，使用牌没有次数和距离限制
func limuUnlimitedSha(r skill.Runtime, seat int) bool {
	if !r.HasSkill(seat, skill.IDLimu) {
		return false
	}
	
	g := r.(*gameSkillRuntime).g
	
	// 检查判定区是否有牌
	if len(g.Players[seat].JudgeArea) == 0 {
		return false
	}
	
	// 判定区有牌时，对攻击范围内的其他角色使用牌没有次数和距离限制
	// 注意：这个实现需要在 canUseSha 等函数中检查
	return true
}

// limuTrickIgnoresDistance 立牧：判定区有牌时，在 IsValidPlayTarget 中已通过
// canAttack 严格判断攻击范围内，此处返回 false 避免重复忽略距离。
func limuTrickIgnoresDistance(r skill.Runtime, seat int, trickKind string) bool {
	return false
}

// hasBasicCard 检查玩家是否有基本牌
func hasBasicCard(g *Game, seat int) bool {
	basicKinds := map[string]bool{
		CardSha:  true,
		CardShan: true,
		CardTao:  true,
		CardJiu:  true,
	}
	
	for _, c := range g.Players[seat].Hand {
		if basicKinds[c.Kind] {
			return true
		}
	}
	return false
}

// hasDiamondCard 检查玩家是否有方块牌（手牌）
func hasDiamondCard(g *Game, seat int) bool {
	for _, c := range g.Players[seat].Hand {
		if c.Suit == "diamond" {
			return true
		}
	}
	return false
}

// hasDiamondCardRuntime 检查玩家是否有方块牌（Runtime 版本，供技能注册使用）
func hasDiamondCardRuntime(r skill.Runtime, seat int) bool {
	// 简化：总是返回 true，实际应该在 engine 包中检查
	// TODO: 在 Runtime 接口中添加 HandCardSuit 方法
	return true
}

// isEquipCard 检查是否为装备牌
func isEquipCard(kind string) bool {
	equipKinds := map[string]bool{
		CardWeapon1:      true,
		CardWeapon2:      true,
		CardWeapon3:      true,
		CardWeapon4:      true,
		CardWeapon5:      true,
		CardWeapon6:      true,
		CardWeapon7:      true,
		CardWeapon8:      true,
		CardWeapon9:      true,
		CardWeapon10:     true,
		CardArmor:         true,
		CardArmorVine:    true,
		CardArmorRenwang: true,
		CardArmorBaiyin:  true,
		CardPlusHorse:    true,
		CardMinusHorse:  true,
	}
	return equipKinds[kind]
}
