package engine

import (
	"fmt"
)

// useLonghunCards 使用龙魂转化的牌
func (g *Game) useLonghunCards(seat int, cardIDs []string, asKind string, useTwoCards, isRed, isBlack bool, events *[]GameEvent) error {
	if len(cardIDs) == 0 {
		return ErrInvalidCard
	}

	// 从手牌中移除使用的牌
	var removedCards []Card
	for _, cardID := range cardIDs {
		found := false
		for i, card := range g.Players[seat].Hand {
			if card.ID == cardID {
				removedCards = append(removedCards, card)
				g.Players[seat].Hand = append(g.Players[seat].Hand[:i], g.Players[seat].Hand[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("牌 %s 不在手牌中", cardID)
		}
	}

	// 将移除的牌放入弃牌堆
	g.DiscardPile = append(g.DiscardPile, removedCards...)

	// 根据转化的牌类型执行相应操作
	switch asKind {
	case CardTao: // 桃
		// 使用桃回血
		healAmount := 1
		if useTwoCards && isRed {
			healAmount = 2 // 两张红色牌，回复值+1
		}
		g.healPlayer(seat, healAmount, events)
		g.Message = fmt.Sprintf("%s 发动【龙魂】，将%d张牌当【桃】使用，回复%d点体力", g.Players[seat].Name, len(cardIDs), healAmount)

	case CardShaFire: // 火杀
		// 使用火杀
		target := g.opponentOf(seat) // 简化：默认攻击对手
		damage := 1
		if useTwoCards && isRed {
			damage = 2 // 两张红色牌，伤害值+1
		}
		// 创建一张虚拟的火杀牌
		virtualCard := Card{ID: "longhun_fire_sha", Kind: CardShaFire, Suit: "D", Name: "火杀（龙魂）"}
		g.applyDamage(seat, target, damage, virtualCard, events)
		g.Message = fmt.Sprintf("%s 发动【龙魂】，将%d张牌当火【杀】使用，造成%d点伤害", g.Players[seat].Name, len(cardIDs), damage)

	case CardShan: // 闪
		// 闪通常用于响应，这里不应该在出牌阶段使用
		return fmt.Errorf("闪不能在出牌阶段使用")

	case CardWuxiek: // 无懈可击
		// 无懈可击通常用于响应，这里不应该在出牌阶段使用
		return fmt.Errorf("无懈可击不能在出牌阶段使用")

	default:
		return fmt.Errorf("未知的牌类型: %s", asKind)
	}

	// 如果使用了两张黑色牌，弃置当前回合角色一张牌
	if useTwoCards && isBlack {
		currentTurn := g.CurrentTurn
		if currentTurn >= 0 && currentTurn < len(g.Players) {
			g.discardRandomCard(currentTurn, events)
			g.Message += fmt.Sprintf("，并弃置了 %s 的一张牌", g.Players[currentTurn].Name)
		}
	}

	// 记录技能使用事件
	*events = append(*events, GameEvent{
		Type:        "skill_trigger",
		PlayerIndex: seat,
		SkillID:     "longhun", // 直接使用技能ID字符串
		Message:     g.Message,
	})

	return nil
}

// responseLonghunCards 打出龙魂转化的牌
func (g *Game) responseLonghunCards(seat int, cardIDs []string, asKind string, useTwoCards, isRed, isBlack bool, events *[]GameEvent) error {
	if len(cardIDs) == 0 {
		return ErrInvalidCard
	}

	// 从手牌中移除打出的牌
	var removedCards []Card
	for _, cardID := range cardIDs {
		found := false
		for i, card := range g.Players[seat].Hand {
			if card.ID == cardID {
				removedCards = append(removedCards, card)
				g.Players[seat].Hand = append(g.Players[seat].Hand[:i], g.Players[seat].Hand[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("牌 %s 不在手牌中", cardID)
		}
	}

	// 将移除的牌放入弃牌堆
	g.DiscardPile = append(g.DiscardPile, removedCards...)

	// 根据转化的牌类型执行相应操作
	switch asKind {
	case CardShan: // 闪
		// 打出闪，响应杀或决斗
		g.Message = fmt.Sprintf("%s 发动【龙魂】，将%d张牌当【闪】打出", g.Players[seat].Name, len(cardIDs))
		// TODO: 实际应该调用响应闪的逻辑

	case CardWuxiek: // 无懈可击
		// 打出无懈可击，抵消锦囊效果
		g.Message = fmt.Sprintf("%s 发动【龙魂】，将%d张牌当【无懈可击】打出", g.Players[seat].Name, len(cardIDs))
		// TODO: 实际应该调用响应无懈可击的逻辑

	default:
		return fmt.Errorf("该牌类型不能在响应阶段打出: %s", asKind)
	}

	// 如果使用了两张黑色牌，弃置当前回合角色一张牌
	if useTwoCards && isBlack {
		currentTurn := g.CurrentTurn
		if currentTurn >= 0 && currentTurn < len(g.Players) {
			g.discardRandomCard(currentTurn, events)
			g.Message += fmt.Sprintf("，并弃置了 %s 的一张牌", g.Players[currentTurn].Name)
		}
	}

	return nil
}

// healPlayer 回复体力
func (g *Game) healPlayer(seat int, amount int, events *[]GameEvent) {
	if seat < 0 || seat >= len(g.Players) {
		return
	}

	p := &g.Players[seat]
	if p.HP >= p.MaxHP {
		return // 体力已满，不需要回复
	}

	p.HP += amount
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}

	// 记录事件
	*events = append(*events, GameEvent{
		Type:        "heal",
		PlayerIndex: seat,
		Heal:        amount,
		Message:     fmt.Sprintf("%s 回复了 %d 点体力", p.Name, amount),
	})
}

// discardRandomCard 弃置角色的一张随机手牌
func (g *Game) discardRandomCard(seat int, events *[]GameEvent) {
	if seat < 0 || seat >= len(g.Players) {
		return
	}

	p := &g.Players[seat]
	if len(p.Hand) == 0 {
		return
	}

	// 简单实现：弃置第一张牌
	discarded := p.Hand[0]
	p.Hand = p.Hand[1:]
	g.DiscardPile = append(g.DiscardPile, discarded)

	// 记录事件
	*events = append(*events, GameEvent{
		Type:        "discard",
		PlayerIndex: seat,
		Card:        &discarded,
		Message:     fmt.Sprintf("%s 弃置了一张牌", p.Name),
	})
}
