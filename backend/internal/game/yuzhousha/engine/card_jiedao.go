package engine

import (
	"fmt"
)

// ============================================================================
// 借刀杀人（JieDao）
// 参考 noname: card/standard.js jiedao
//
// 规则：
//   类型：非延时锦囊 (trick)，双目标
//   目标1（被借刀者）：必须有武器，且不能是使用者自己
//   目标2（出杀目标）：在被借刀者攻击范围内，且被借刀者能对其使用杀
//   结算：
//     无懈窗口 → 被借刀者选择：
//       a) 对目标2使用一张杀
//       b) 将武器牌交给使用者
// ============================================================================

// resolveJieDao 借刀杀人入口：校验双目标 → 进入无懈窗口。
func (g *Game) resolveJieDao(source int, targetSpec PlayTarget, trick Card, events *[]GameEvent) error {
	weaponHolder := targetSpec.SeatIndex
	shaTarget := targetSpec.SecondSeatIndex

	// 校验双目标合法性（必须两个目标都有效）
	if weaponHolder < 0 || weaponHolder >= len(g.Players) || weaponHolder == source {
		return ErrInvalidTarget
	}
	if shaTarget < 0 || shaTarget >= len(g.Players) || shaTarget == weaponHolder {
		return ErrInvalidTarget
	}
	// 被借刀者必须有武器
	if g.Players[weaponHolder].Weapon == nil {
		return ErrInvalidTarget
	}
	// 出杀目标必须在被借刀者攻击范围内
	if !g.inAttackRange(weaponHolder, shaTarget) {
		return ErrInvalidTarget
	}
	// 被借刀者必须能对目标使用杀（帷幕等技能阻止的目标不可选）
	if g.targetBlockedBySkill(shaTarget, CardSha) {
		return ErrInvalidTarget
	}

	g.notifyBecameTarget(weaponHolder, source, trick, events)

	// 进入无懈窗口（和过河拆桥等一样）
	responder := weaponHolder
	if !g.isEnemy(source, weaponHolder) {
		responder = g.opponentOf(source)
	}

	// 保存第二目标信息到 targetSpec（startWuxiekTrickWindow 不直接支持双目标，
	// 我们通过 Pending 的 JieDaoShaTarget 字段传递）
	return g.startJieDaoWuxiekWindow(source, responder, weaponHolder, shaTarget, trick, targetSpec, events)
}

// startJieDaoWuxiekWindow 借刀杀人的无懈窗口（保存第二目标信息）。
func (g *Game) startJieDaoWuxiekWindow(source, responder, weaponHolder, shaTarget int,
	trick Card, spec PlayTarget, events *[]GameEvent) error {

	n := len(g.Players)
	if source < 0 || source >= n || responder < 0 || responder >= n {
		return fmt.Errorf("jiedao: invalid params source=%d responder=%d", source, responder)
	}

	allQueue := g.createResponseQueue(responder)
	queue := make([]int, 0, len(allQueue))
	for _, s := range allQueue {
		if s != source {
			queue = append(queue, s)
		}
	}

	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:      source,
		TargetIndex:      -1,
		ReturnIndex:      source,
		EffectTarget:     weaponHolder,
		Card:             trick,
		ResponseMode:     ResponseModeWuxiekTrick,
		TargetZone:       spec.Zone,
		TargetCardID:     spec.CardID,
		ResponseQueue:    queue,
		ResponseIndex:    0,
		WuxiekChain:      nil,
		JieDaoShaTarget:  shaTarget, // 借刀杀人的出杀目标
	}

	g.advanceToNextWuxiekResponder(events)
	if g.Message == "" {
		weaponName := "武器"
		if g.Players[weaponHolder].Weapon != nil {
			weaponName = g.Players[weaponHolder].Weapon.Name
		}
		g.Message = fmt.Sprintf("【借刀杀人】：%s 需对 %s 出【杀】或交出 %s",
			g.Players[weaponHolder].Name, g.Players[shaTarget].Name, weaponName)
	}
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: source,
		TargetIndex: responder,
		Card:        &trick,
		Message:     g.Message,
	})
	return nil
}

// ============================================================================
// 借刀杀人结算（无懈通过后）
// ============================================================================

// resolveJieDaoEffect 无懈通过后，被借刀者选择：出杀 或 给武器。
func (g *Game) resolveJieDaoEffect(source, weaponHolder, shaTarget int, trick Card, events *[]GameEvent) error {
	// 被借刀者已死或武器已被拆 → 跳过
	if g.Players[weaponHolder].HP <= 0 || g.Players[weaponHolder].Weapon == nil {
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = source
		g.resetTimer()
		return nil
	}

	weaponName := g.Players[weaponHolder].Weapon.Name
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:     source,
		TargetIndex:     weaponHolder,
		ReturnIndex:     source,
		Card:            trick,
		ResponseMode:    ResponseModeJieDao,
		JieDaoShaTarget: shaTarget,
	}
	g.Message = fmt.Sprintf("%s 需对 %s 出【杀】或交出 %s",
		g.Players[weaponHolder].Name, g.Players[shaTarget].Name, weaponName)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "jiedao_choose",
		PlayerIndex: source,
		TargetIndex: weaponHolder,
		Message:     g.Message,
	})
	return nil
}

// ApplyJieDaoSha 被借刀者选择出杀。
// 注意：伤害来源是锦囊使用者（jiedaoSource），不是被借刀者。
// 参考 noname: respondTo: [player, card] → 伤害来源 = player（锦囊使用者）。
func (g *Game) ApplyJieDaoSha(seat int, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeJieDao || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	shaTarget := g.Pending.JieDaoShaTarget
	jiedaoSource := g.Pending.SourceIndex // 借刀杀人的使用者（伤害来源）

	// 查找杀牌
	idx, _, ok := g.findCard(seat, cardID)
	if !ok {
		return ErrInvalidCard
	}
	shaCard := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, shaCard)
	g.syncCounts()

	jiedaoMsg := fmt.Sprintf("%s 对 %s 使用【杀】（响应【借刀杀人】）",
		g.Players[seat].Name, g.Players[shaTarget].Name)
	*events = append(*events, GameEvent{
		Type:        "jiedao_sha",
		PlayerIndex: seat,
		TargetIndex: shaTarget,
		Card:        &shaCard,
		Message:     jiedaoMsg,
	})

	// 清除 Pending，进入杀响应流程
	g.Pending = nil

	// 触发杀响应（目标需出闪），伤害来源设置为锦囊使用者
	return g.startJieDaoShaResponse(jiedaoSource, seat, shaTarget, shaCard, jiedaoMsg, events)
}

// startJieDaoShaResponse 借刀杀人的杀响应：伤害来源是 jiedaoSource（锦囊使用者），不是出杀者。
// 参考 noname: respondTo: [player, card] → 伤害来源 = player。
func (g *Game) startJieDaoShaResponse(jiedaoSource, shaUser, shaTarget int, shaCard Card, message string, events *[]GameEvent) error {
	g.Phase = PhaseResponse
	g.appendWushuangMessage(shaUser, shaCard.Kind, &message)

	g.Pending = &PendingCombat{
		SourceIndex:     jiedaoSource, // 伤害来源 = 借刀杀人使用者（noname: respondTo: [player, card]）
		TargetIndex:     shaTarget,    // 被杀的目标
		ReturnIndex:     jiedaoSource,
		Card:            shaCard,
		RequiredKind:    CardSha,
		Damage:          1,
		// 无双响应数检查用出杀者的技能（shaUser 是被借刀者，无双是被借刀者的技能）
		ResponsesNeeded: g.wushuangResponsesNeeded(shaUser, shaCard.Kind),
	}
	g.Message = message
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "trick_response",
		PlayerIndex: jiedaoSource,
		TargetIndex: shaTarget,
		Card:        &shaCard,
		Message:     message,
	})
	return nil
}

// ApplyJieDaoGiveWeapon 被借刀者选择交出武器。
func (g *Game) ApplyJieDaoGiveWeapon(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeJieDao || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	source := g.Pending.SourceIndex

	weapon := g.Players[seat].Weapon
	if weapon == nil {
		return ErrInvalidCard
	}

	// 移除武器，加入使用者手牌
	g.Players[seat].Weapon = nil
	g.Players[source].Hand = append(g.Players[source].Hand, *weapon)
	g.syncCounts()

	msg := fmt.Sprintf("%s 将 %s 交给 %s（响应【借刀杀人】）",
		g.Players[seat].Name, weapon.Name, g.Players[source].Name)
	g.appendSkillEvent(events, "", seat, source, msg)
	*events = append(*events, GameEvent{
		Type:        "jiedao_give_weapon",
		PlayerIndex: seat,
		TargetIndex: source,
		Card:        weapon,
		Message:     msg,
	})

	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.resetTimer()
	return nil
}

// ============================================================================
// AI 辅助
// ============================================================================

// jieDaoAIChoose AI 决定借刀杀人的行为：出杀还是给武器。
func (g *Game) jieDaoAIChoose(seat int, events *[]GameEvent) {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeJieDao || g.Pending.TargetIndex != seat {
		return
	}
	// AI 简单策略：有杀就出杀，没杀就给武器
	for _, card := range g.Players[seat].Hand {
		if card.Kind == CardSha {
			_ = g.ApplyJieDaoSha(seat, card.ID, events)
			return
		}
	}
	// 没有杀，交出武器
	_ = g.ApplyJieDaoGiveWeapon(seat, events)
}

// ============================================================================
// 辅助函数
// ============================================================================

// inAttackRange 判断 from 是否在攻击范围内能打到 to。
func (g *Game) inAttackRange(from, to int) bool {
	dist := g.distanceBetween(from, to)
	// 默认攻击范围 1 + 武器加成
	attackRange := 1
	if g.Players[from].Weapon != nil {
		attackRange = weaponRange(g.Players[from].Weapon.Kind)
	}
	return dist <= attackRange
}

// libFilterTargetEnabled 模拟 noname 的 lib.filter.targetEnabled。
// 检查能否对目标使用杀（基本规则：不能对自己、目标存活）。
func libFilterTargetEnabled(card Card, source, target int, g *Game) bool {
	if source == target {
		return false
	}
	if g.Players[target].HP <= 0 {
		return false
	}
	// 检查技能阻止（帷幕等）
	if g.targetBlockedBySkill(target, card.Kind) {
		return false
	}
	return true
}
