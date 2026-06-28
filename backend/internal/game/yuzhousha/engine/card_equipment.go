package engine

import "fmt"

// 古锭刀、藤甲等装备牌效果。

const counterChained = "chained"

// removeEquipCard 从装备区移除牌（卸下装备时同时移除对应的装备技能ID）。
func (g *Game) removeEquipCard(seat int, zone string, events *[]GameEvent) Card {
	p := &g.Players[seat]
	var card *Card
	switch zone {
	case EquipWeapon:
		card = p.Weapon
		p.Weapon = nil
	case EquipArmor:
		card = p.Armor
		p.Armor = nil
	case EquipPlusHorse:
		card = p.PlusHorse
		p.PlusHorse = nil
	case EquipMinusHorse:
		card = p.MinusHorse
		p.MinusHorse = nil
	default:
		return Card{}
	}
	if card == nil {
		return Card{}
	}
	g.removeEquipSkill(seat, card.Kind) // TagEquipSkill: 移除装备技能
	g.SyncCounts()
	return *card
}

func (g *Game) hasVineArmor(seat int) bool {
	p := &g.Players[seat]
	return p.Armor != nil && p.Armor.Kind == CardArmorVine
}

func (g *Game) hasBaguaArmor(seat int) bool {
	p := &g.Players[seat]
	return p.Armor != nil && p.Armor.Kind == CardArmor
}

func (g *Game) vineBlocksTrick(target int, kind string) bool {
	if !g.hasVineArmor(target) {
		return false
	}
	// 藤甲：南蛮入侵、万箭齐发无效（锁定技）
	// 注意：普通杀无效的逻辑在 applyDamage 中处理，这里只处理锦囊牌
	switch kind {
	case CardNanMan, CardWanJian:
		return true
	default:
		return false
	}
}

func (g *Game) isChained(seat int) bool {
	return g.getSkillCounter(seat, counterChained) > 0
}

func (g *Game) setChained(seat int, chained bool) {
	old := g.isChained(seat)
	if chained {
		g.setSkillCounter(seat, counterChained, 1)
	} else {
		g.setSkillCounter(seat, counterChained, 0)
	}
	Logf("setChained: seat=%d(%s) old=%v new=%v", seat, g.Players[seat].Name, old, chained)
}

func (g *Game) toggleChained(seat int) {
	old := g.isChained(seat)
	g.setChained(seat, !old)
	Logf("toggleChained: seat=%d(%s) old=%v new=%v", seat, g.Players[seat].Name, old, !old)
}

func (g *Game) filterAoeQueue(queue []int, trickKind string) []int {
	if trickKind != CardNanMan && trickKind != CardWanJian {
		return queue
	}
	out := make([]int, 0, len(queue))
	for _, seat := range queue {
		if g.vineBlocksTrick(seat, trickKind) {
			continue
		}
		out = append(out, seat)
	}
	return out
}

// adjustDamageAmount 结算古锭刀、藤甲等加伤（青釭剑 ignoreArmor 时藤甲失效）。
// 统一从 card.DamageType 判断属性，未来新增雷属性加伤装备也在此处扩展。
func (g *Game) adjustDamageAmount(source, target, amount int, card Card, isFire, ignoreArmor bool) int {
	if amount <= 0 {
		amount = 1
	}
	// 古锭刀：杀且目标无手牌时+1
	if isSha(card.Kind) && g.hasWeaponKind(source, CardWeapon6) && len(g.Players[target].Hand) == 0 {
		amount++
	}
	// 藤甲：火焰伤害+1（青釭剑 ignoreArmor 时失效）
	if !ignoreArmor && g.hasVineArmor(target) && (isFire || card.DamageType == DamageTypeFire) {
		amount++
	}
	// TODO: 未来新增"雷电伤害+1"等装备在此扩展，例如：
	// if !ignoreArmor && g.hasThunderArmor(target) && card.DamageType == DamageTypeThunder {
	//     amount++
	// }
	return amount
}

// isSha 判断是否为任意种类的杀（含属性杀）。
func isSha(kind string) bool {
	return kind == CardSha || kind == CardShaFire || kind == CardShaThunder
}

// convertCardToKind 将牌的 Kind 统一转为目标类型（用于技能变牌）
func (g *Game) convertCardToKind(card Card, targetKind string) Card {
	switch targetKind {
	case CardSha:
		card.Kind = CardSha
		card.DamageType = DamageTypeNormal
		card.Name = "杀"
	case CardShan:
		card.Kind = CardShan
		card.Name = "闪"
	case CardTao:
		card.Kind = CardTao
		card.Name = "桃"
	case CardJiu:
		card.Kind = CardJiu
		card.Name = "酒"
	case CardGuoHe:
		card.Kind = CardGuoHe
		card.Name = "过河拆桥"
	case CardTanNang:
		card.Kind = CardTanNang
		card.Name = "顺手牵羊"
	case CardJueDou:
		card.Kind = CardJueDou
		card.Name = "决斗"
	case CardLeBu:
		card.Kind = CardLeBu
		card.Name = "乐不思蜀"
	case CardWuxiek:
		card.Kind = CardWuxiek
		card.Name = "无懈可击"
	case CardShaFire:
		card.Kind = CardShaFire
		card.DamageType = DamageTypeFire
		card.Name = "火杀"
	case CardShaThunder:
		card.Kind = CardShaThunder
		card.DamageType = DamageTypeThunder
		card.Name = "雷杀"
	}
	return card
}

// spreadChainedFireDamage 铁索连环AOE：完全类比南蛮入侵。
// 宣告 → 逐人扣血(无响应) → 完毕。
// 队列存 g.Pending.AoeQueue，濒死时 startDyingWindow 自动保存/恢复 Pending。
func (g *Game) spreadChainedFireDamage(source, primaryTarget, amount int, card Card, events *[]GameEvent) {
	Logf("spreadChainedFireDamage: source=%d(%s) primaryTarget=%d(%s) amount=%d card.Kind=%s card.DamageType=%s",
		source, g.Players[source].Name, primaryTarget, g.Players[primaryTarget].Name, amount, card.Kind, card.DamageType)
	if !g.isChained(primaryTarget) {
		Logf("spreadChainedFireDamage: primaryTarget not chained, skip")
		return
	}
	if card.DamageType != DamageTypeFire && card.DamageType != DamageTypeThunder {
		Logf("spreadChainedFireDamage: not elemental damage, skip")
		return
	}
	// 收集连环角色队列
	chainSeats := make([]int, 0)
	for seat := range g.Players {
		if seat == primaryTarget || !g.isChained(seat) || g.Players[seat].HP <= 0 {
			continue
		}
		chainSeats = append(chainSeats, seat)
	}
	Logf("spreadChainedFireDamage: chainSeats=%v", chainSeats)

	// 重置首要目标
	g.setChained(primaryTarget, false)

	// 宣告
	g.Message = fmt.Sprintf("【铁索连环】%s 受到属性伤害，传导开始", g.Players[primaryTarget].Name)
	*events = append(*events, GameEvent{
		Type:        "tiesuo_announce",
		PlayerIndex: source,
		TargetIndex: primaryTarget,
		Message:     g.Message,
	})

	if len(chainSeats) == 0 {
		*events = append(*events, GameEvent{
			Type:        "tiesuo_spread",
			PlayerIndex: source,
			Message:     "【铁索连环】传导完毕",
		})
		return
	}

	// 开始逐人：对第一个人扣血
	g.startTiesuoAoe(source, amount, card, chainSeats, events)
}

// startTiesuoAoe 铁索连环AOE：完全类比南蛮的 resolvePendingMiss。
// 对当前目标扣血 → 如果HP<=0则濒死（Pending自动保存）→ 未死亡则继续下一个。
// 队列和链式伤害值存 g.Pending，濒死时自动保存/恢复。
func (g *Game) startTiesuoAoe(source, amount int, card Card, remaining []int, events *[]GameEvent) {
	if len(remaining) == 0 {
		g.finishTiesuoAoe(source, events)
		return
	}
	seat := remaining[0]
	rest := remaining[1:]

	// 重置连环状态
	g.setChained(seat, false)

	// 链式加成（藤甲加伤等）
	dmg := g.adjustDamageAmount(source, seat, amount, card, card.DamageType == DamageTypeFire, false)
	// 白银狮子：每个传导目标独立检查（伤害 > 1 时锁定为 1）
	g.baiyinReduceDamage(seat, &dmg)
	Logf("startTiesuoAoe: seat=%d(%s) incoming=%d dmg=%d remaining=%v",
		seat, g.Players[seat].Name, amount, dmg, rest)

	// 宣告
	g.Message = fmt.Sprintf("【铁索连环】%s 受到 %d 点属性伤害", g.Players[seat].Name, dmg)
	*events = append(*events, GameEvent{
		Type:        "tiesuo_aoe",
		PlayerIndex: source,
		TargetIndex: seat,
		Damage:      dmg,
		Message:     g.Message,
	})

	// 保存队列到 g.Pending，濒死时 startDyingWindow 自动保存到 SavedPending
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  seat,
		EffectTarget: seat,
		Card:         card,
		Damage:       dmg,
		AoeQueue:     rest,
		ReturnIndex:  source,
		RequiredKind: "tiesuo",
	}

	// 构建带 AoeResume 的 DamageResume，扣血+自动濒死
	// 使用 ApplyDamageAndCheckDeathWithAoe 将 AOE 信息注入 DamageEvent，
	// 确保铁索传导中有人濒死时 AOE 链不因濒死流程而断裂。
	dyingResume := DamageResume{}
	g.setAoeResume(&dyingResume, source, dmg, card, rest, true)
	if g.ApplyDamageAndCheckDeathWithAoe(source, seat, dmg, card, dyingResume, events) {
		return
	}
	*events = append(*events, GameEvent{
		Type:        "trick_hit",
		PlayerIndex: source,
		TargetIndex: seat,
		Damage:      dmg,
		Message:     fmt.Sprintf("%s 受到【铁索连环】%d 点伤害", g.Players[seat].Name, dmg),
	})

	// 未死亡：类比万箭，走 continueAfterDamage 技能链
	g.clearPending()
	resume := DamageResume{}
	g.setAoeResume(&resume, source, dmg, card, rest, true)
	if g.continueAfterDamage(source, seat, dmg, card, resume, events) {
		return
	}
	g.continueTiesuoAoe(source, dmg, card, rest, events)
}

// continueTiesuoAoe 铁索连环AOE恢复：继续下一个人或完毕（完全类比 continueNanManAfterTarget）。
func (g *Game) continueTiesuoAoe(source, amount int, card Card, rest []int, events *[]GameEvent) {
	if len(rest) == 0 {
		g.finishTiesuoAoe(source, events)
		return
	}
	g.startTiesuoAoe(source, amount, card, rest, events)
}

// finishTiesuoAoe 铁索连环AOE完毕：恢复出牌。
func (g *Game) finishTiesuoAoe(source int, events *[]GameEvent) {
	Logf("finishTiesuoAoe: done")
	*events = append(*events, GameEvent{
		Type:        "tiesuo_spread",
		PlayerIndex: source,
		Message:     "【铁索连环】传导完毕",
	})
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
	g.resetTimer()
}
