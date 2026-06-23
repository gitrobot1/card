package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) PlayCard(seat int, cardID string, targetIndex int, events *[]GameEvent) error {
	return g.PlayCardWithTarget(seat, cardID, PlayTarget{SeatIndex: targetIndex}, events)
}

func (g *Game) PlayCardWithTarget(seat int, cardID string, target PlayTarget, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase == PhaseResponse && g.Pending != nil {
		switch g.Pending.ResponseMode {
		case ResponseModeGuanYuFollow:
			if seat != g.Pending.TargetIndex {
				return ErrNotYourTurn
			}
			_, cardObj, ok := g.findCard(seat, cardID)
			if !ok || !g.cardPlaysAs(seat, cardObj, CardSha) {
				return ErrInvalidCard
			}
			return g.playSha(seat, cardID, g.Pending.EffectTarget, events)
		case ResponseModeQilinBow:
			if seat != g.Pending.TargetIndex {
				return ErrNotYourTurn
			}
			if target.Zone != EquipPlusHorse && target.Zone != EquipMinusHorse {
				return ErrInvalidTarget
			}
			return g.qilinDiscardHorse(seat, target.Zone, events)
		case ResponseModeWuguPick:
			if seat != g.Pending.WuguPickSeat {
				return ErrNotYourTurn
			}
			return g.pickWuguCard(seat, cardID, events)
		case ResponseModeSkillJijiang:
			if seat != g.Pending.TargetIndex {
				return ErrNotYourTurn
			}
			return g.respondJijiangSha(seat, cardID, events)
		case ResponseModeSkillLuanwu:
			if seat != g.Pending.TargetIndex {
				return ErrNotYourTurn
			}
			t := g.Pending.EffectTarget
			if t < 0 {
				t = seat
			}
			return g.playLuanwuSha(seat, cardID, t, events)
		case ResponseModeHuoGong:
			if seat != g.Pending.TargetIndex {
				return ErrNotYourTurn
			}
			return g.respondHuoGongDiscard(seat, cardID, events)
		default:
			return ErrPendingCombat
		}
	}
	if g.Phase == PhaseResponse {
		return ErrPendingCombat
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		if _, cardObj, ok := g.findCard(seat, cardID); ok && g.canUseJijiHeal(seat, cardObj) {
			return g.playJijiHeal(seat, cardID, events)
		}
		return ErrNotYourTurn
	}

	// 查找牌（手牌或装备区）
	zone, idx, cardObj, ok := g.findCardInHandOrEquip(seat, cardID)
	if !ok {
		return ErrInvalidCard
	}

	// 通用：如果是装备牌被技能变牌（cardPlaysAs 返回 true），先移除装备，再按变牌逻辑处理
	if isEquipKind(cardObj.Kind) {
		// 奇袭：装备区黑色牌视为过河拆桥
		if g.getSkillCounter(seat, counterQixiActive) > 0 && skill.IsBlackSuit(cardObj.Suit) {
			discarded := g.removeEquipCard(seat, zone, events)
			g.DiscardPile = append(g.DiscardPile, discarded)
			g.notifyEquipLost(seat, discarded, "skill", events)
			g.appendSkillEvent(events, skill.IDQixi, seat, -1, fmt.Sprintf("%s 使用【奇袭】", g.Players[seat].Name))
			return g.playTrickAsGuoHe(seat, target, events)
		}
		// 变牌为杀（武圣/龙胆等）
		if g.cardPlaysAs(seat, cardObj, CardSha) && g.canUseSha(seat) && g.isValidPlayTarget(seat, target.SeatIndex, CardSha) {
			discarded := g.removeEquipCard(seat, zone, events)
			g.notifyEquipLost(seat, discarded, "skill", events)
			return g.playShaWithCard(seat, discarded, target.SeatIndex, events)
		}
		// 变牌为桃（急救等）
		if g.cardPlaysAs(seat, cardObj, CardTao) {
			discarded := g.removeEquipCard(seat, zone, events)
			g.notifyEquipLost(seat, discarded, "skill", events)
			discarded = g.convertCardToKind(discarded, CardTao)
			return g.playTaoWithCard(seat, discarded, events)
		}
		// 变牌为闪（龙胆/倾国等）——出牌阶段一般不出闪，但预留
		if g.cardPlaysAs(seat, cardObj, CardShan) {
			// 出牌阶段不出闪
		}
		// 变牌为锦囊（国色/立牧等）
		if g.cardPlaysAs(seat, cardObj, CardLeBu) || g.cardPlaysAs(seat, cardObj, CardGuoHe) || g.cardPlaysAs(seat, cardObj, CardTanNang) || g.cardPlaysAs(seat, cardObj, CardJueDou) {
			discarded := g.removeEquipCard(seat, zone, events)
			g.notifyEquipLost(seat, discarded, "skill", events)
			// 确定目标锦囊类型并转换
			trickKind := ""
			if g.cardPlaysAs(seat, cardObj, CardLeBu) {
				trickKind = CardLeBu
			} else if g.cardPlaysAs(seat, cardObj, CardGuoHe) {
				trickKind = CardGuoHe
			} else if g.cardPlaysAs(seat, cardObj, CardTanNang) {
				trickKind = CardTanNang
			} else if g.cardPlaysAs(seat, cardObj, CardJueDou) {
				trickKind = CardJueDou
			}
			if trickKind != "" {
				discarded = g.convertCardToKind(discarded, trickKind)
			}
			return g.playTrickWithCard(seat, discarded, target, events)
		}
		// 没有变牌，正常装备
		return g.playEquip(seat, cardID, events)
	}

	// 手牌：奇袭
	if g.getSkillCounter(seat, counterQixiActive) > 0 && skill.IsBlackSuit(cardObj.Suit) {
		discarded := g.removeHandCard(seat, idx, events)
		g.DiscardPile = append(g.DiscardPile, discarded)
		g.runCardsDiscardedHooks(seat, "cost", []Card{discarded}, events)
		g.appendSkillEvent(events, skill.IDQixi, seat, -1, fmt.Sprintf("%s 使用【奇袭】", g.Players[seat].Name))
		return g.playTrickAsGuoHe(seat, target, events)
	}

	if cardObj.Kind != CardSha && g.cardPlaysAs(seat, cardObj, CardSha) {
		if g.canUseSha(seat) && g.isValidPlayTarget(seat, target.SeatIndex, CardSha) {
			return g.playSha(seat, cardID, target.SeatIndex, events)
		}
	}

	switch cardObj.Kind {
	case CardSha, CardShaFire, CardShaThunder:
		return g.playSha(seat, cardID, target.SeatIndex, events)
	case CardShan:
		return ErrInvalidCard
	case CardTao:
		return g.playTao(seat, cardID, events)
	case CardJiu:
		return g.playJiu(seat, cardID, events)
	case CardGuoHe, CardTanNang, CardNanMan, CardWanJian, CardJueDou, CardLeBu, CardBingLiang, CardShanDian, CardWuGu, CardTaoYuan, CardWuZhong, CardHuoGong, CardTieSuo:
		return g.playTrick(seat, cardID, target, events)
	default:
		return ErrInvalidCard
	}
}

// isEquipKind 判断是否为装备牌
func isEquipKind(kind string) bool {
	switch kind {
	case CardWeapon1, CardWeapon2, CardWeapon3, CardWeapon4, CardWeapon5, CardWeapon6, CardWeapon7, CardWeapon8, CardWeapon9, CardArmor, CardArmorVine, CardPlusHorse, CardMinusHorse:
		return true
	}
	return false
}

func (g *Game) playSha(seat int, cardID string, targetIndex int, events *[]GameEvent) error {
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || !g.cardPlaysAs(seat, cardObj, CardSha) {
		return ErrInvalidCard
	}
	played := g.removeHandCard(seat, idx, events)
	return g.playShaWithCard(seat, played, targetIndex, events)
}

// playShaWithCard 用已移除的牌打出杀（支持装备牌变牌）
func (g *Game) playShaWithCard(seat int, played Card, targetIndex int, events *[]GameEvent) error {
	guanYuFollow := g.isGuanYuFollowPending() && seat == g.Pending.TargetIndex
	if !guanYuFollow && !g.canUseSha(seat) {
		return ErrAlreadyActed
	}
	if guanYuFollow && targetIndex != g.Pending.EffectTarget {
		return ErrInvalidTarget
	}
	if guanYuFollow {
		if !g.isEnemy(seat, targetIndex) {
			return ErrInvalidTarget
		}
		if g.runSkillHooks(nil, skill.HookCall{Kind: skill.HookTargetBlocked, Target: targetIndex, CardKind: CardSha}).Bool {
			return ErrInvalidTarget
		}
	} else if !g.isValidPlayTarget(seat, targetIndex, CardSha) {
		return ErrInvalidTarget
	}

	// 检查是否因龙胆而将闪当杀使用，如果是则触发冲阵
	g.triggerChongzhenWithEvents(seat, played, CardSha, events)

	// 如果牌本身不是杀（通过技能变牌，如武圣/龙胆），统一转为普通杀
	if !isSha(played.Kind) {
		played = g.convertCardToKind(played, CardSha)
	}

	// 激昂：使用红色杀时摸一张牌
	if g.hasSkill(seat, skill.IDJiang) && skill.IsJiangCard(CardSha, played.Suit) {
		_ = g.drawSkillCards(seat, skill.IDJiang, 1, "", events)
	}

	// 朱雀羽扇：将普通杀转为火杀
	hasZhuQue := g.hasWeaponKind(seat, CardWeapon7)
	if hasZhuQue && played.DamageType == DamageTypeNormal {
		played.DamageType = DamageTypeFire
		played.Name = "火杀"
	}

	g.DiscardPile = append(g.DiscardPile, played)
	if !guanYuFollow {
		g.runCardsDiscardedHooks(seat, "play", []Card{played}, events)
	}
	if guanYuFollow {
		g.Pending = nil
	} else if !g.skillUnlimitedSha(seat) {
		if !g.Players[seat].ShaUsedThisTurn {
			g.Players[seat].ShaUsedThisTurn = true
		} else {
			g.Players[seat].ShaExtraUsedThisTurn = true
		}
	}
	g.markShaInPlayPhase(seat)
	damage := g.shaBaseDamage(seat)

	ignoreArmor := g.hasWeaponKind(seat, CardWeapon2)
	msg := fmt.Sprintf("%s 对 %s 使用【杀】，等待出闪", g.Players[seat].Name, g.Players[targetIndex].Name)
	if g.hasWeaponKind(seat, CardWeapon4) && len(g.Players[seat].Hand) == 0 {
		msg = fmt.Sprintf("%s 对 %s 使用【杀】（【方天画戟】最后一张手牌）", g.Players[seat].Name, g.Players[targetIndex].Name)
	}
	if ignoreArmor {
		msg += "（【青釭剑】无视防具）"
	}
	if hasZhuQue && played.DamageType == DamageTypeFire {
		msg = fmt.Sprintf("%s 对 %s 使用【火杀】（【朱雀羽扇】）", g.Players[seat].Name, g.Players[targetIndex].Name)
	} else if played.DamageType == DamageTypeFire {
		msg = fmt.Sprintf("%s 对 %s 使用【火杀】，等待出闪", g.Players[seat].Name, g.Players[targetIndex].Name)
	} else if played.DamageType == DamageTypeThunder {
		msg = fmt.Sprintf("%s 对 %s 使用【雷杀】，等待出闪", g.Players[seat].Name, g.Players[targetIndex].Name)
	}
	if g.getSkillCounter(seat, counterLuoyiActive) > 0 {
		msg += "（【裸衣】+1）"
	}
	g.appendWushuangMessage(seat, CardSha, &msg)

	g.Phase = PhaseResponse
	tieqiPending := g.hasSkill(seat, SkillTieqi)
	g.Pending = &PendingCombat{
		SourceIndex:     seat,
		TargetIndex:     targetIndex,
		ReturnIndex:     seat,
		Card:            played,
		RequiredKind:    CardShan,
		Damage:          damage,
		IgnoreArmor:     ignoreArmor,
		TieqiPending:    tieqiPending,
		ResponsesNeeded: g.wushuangResponsesNeeded(seat, CardSha),
		ActorSeat:       targetIndex,
		SubjectSeat:     targetIndex,
	}
	g.initPojunOnShaPending(seat, targetIndex, g.Pending)
	
	// 破军技能：如果源有破军且目标有可拿的牌，直接打开选牌窗口
	if g.hasSkill(seat, SkillPojun) && g.hasTakeableCard(targetIndex) {
		g.Message = msg
		g.resetTimer()
		*events = append(*events, GameEvent{
			Type:        "play_sha",
			PlayerIndex: seat,
			TargetIndex: targetIndex,
			Card:        &played,
			Message:     g.Message,
		})
		g.notifyBecameTarget(targetIndex, seat, played, events)
		return g.enterPojunPlacing(events)
	}
	
	g.Message = msg
	g.resetTimer()

	*events = append(*events, GameEvent{
		Type:        "play_sha",
		PlayerIndex: seat,
		TargetIndex: targetIndex,
		Card:        &played,
		Message:     g.Message,
	})
	g.notifyBecameTarget(targetIndex, seat, played, events)
	if g.canOfferLiuli(targetIndex) {
		g.offerLiuliWindow(targetIndex, events)
	} else {
		_ = g.advanceShaBeforeTargetResponse(events)
	}
	return nil
}

func (g *Game) playTrick(seat int, cardID string, targetSpec PlayTarget, events *[]GameEvent) error {
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok {
		return ErrInvalidCard
	}
	// 铁索连环重铸（前端用 targetZone='recast' 标记）
	if cardObj.Kind == CardTieSuo && targetSpec.Zone == "recast" {
		played := g.removeHandCard(seat, idx, events)
		g.runCardsDiscardedHooks(seat, "play", []Card{played}, events)
		*events = append(*events, GameEvent{
			Type:        "play_trick",
			PlayerIndex: seat,
			TargetIndex: seat,
			Card:        &played,
			Message:     fmt.Sprintf("%s 重铸【铁索连环】", g.Players[seat].Name),
		})
		return g.playTieSuoRecast(seat, played, events)
	}
	played := g.removeHandCard(seat, idx, events)
	return g.playTrickWithCard(seat, played, targetSpec, events)
}

// playTrickWithCard 用已移除的牌当锦囊使用（支持装备牌变牌，如国色）
func (g *Game) playTrickWithCard(seat int, played Card, targetSpec PlayTarget, events *[]GameEvent) error {
	if g.is3v3() && played.Kind == CardShanDian {
		return ErrInvalidCard
	}
	if g.isIdentity() && played.Kind == CardShanDian {
		return ErrInvalidCard
	}

	target := targetSpec.SeatIndex
	if trickNeedsOpponentTarget(played.Kind) {
		if target < 0 || target >= len(g.Players) || !g.isValidPlayTarget(seat, target, played.Kind) {
			return ErrInvalidTarget
		}
	} else if target < 0 || target >= len(g.Players) {
		target = g.opponentOf(seat)
	}
	if (played.Kind == CardGuoHe || played.Kind == CardTanNang) && !g.hasTakeableCard(target) {
		return ErrInvalidTarget
	}
	if played.Kind == CardBingLiang && !g.canBingliangTarget(seat, target) {
		return ErrInvalidTarget
	}
	if played.Kind == CardHuoGong && len(g.Players[target].Hand) == 0 {
		return ErrInvalidTarget
	}
	// 延时锦囊：目标判定区已有同名牌则不能使用
	if trickStaysInJudge(played.Kind) && g.Players[target].hasJudgeKind(played.Kind) {
		return ErrInvalidTarget
	}

	// 激昂：使用【决斗】时摸一张牌
	if g.hasSkill(seat, skill.IDJiang) && played.Kind == CardJueDou {
		_ = g.drawSkillCards(seat, skill.IDJiang, 1, "", events)
	}

	effectTarget := targetSpec.SeatIndex
	if effectTarget < 0 || effectTarget >= len(g.Players) {
		effectTarget = target
	}
	if g.targetBlockedByTrick(effectTarget, played) {
		return ErrInvalidTarget
	}
	if played.Kind == CardJueDou && g.runSkillHooks(nil, skill.HookCall{Kind: skill.HookTargetBlocked, Target: effectTarget, CardKind: CardJueDou}).Bool {
		return ErrInvalidTarget
	}

	if !trickStaysInJudge(played.Kind) {
		g.DiscardPile = append(g.DiscardPile, played)
		g.runCardsDiscardedHooks(seat, "play", []Card{played}, events)
	}
	*events = append(*events, GameEvent{
		Type:        "play_trick",
		PlayerIndex: seat,
		TargetIndex: targetSpec.SeatIndex,
		Card:        &played,
		Message:     fmt.Sprintf("%s 使用【%s】", g.Players[seat].Name, played.Name),
	})

	// 图射：使用非装备牌指定目标后，若没有基本牌，摸X张牌
	g.tryTriggerTushe(seat, played, targetSpec.SeatIndex, events)

	Logf("playTrickWithCard: seat=%d(%s) kind=%s", seat, g.Players[seat].Name, played.Kind)

	// 非延时锦囊打出时触发集智（延时锦囊乐/兵/闪电不触发）
	if !trickStaysInJudge(played.Kind) {
		g.notifyInstantTrickUsed(seat, played.Kind, events)
	}

	switch played.Kind {
	case CardLeBu:
		g.notifyBecameTarget(target, seat, played, events)
		return g.placeLebu(seat, target, played, events)
	case CardBingLiang:
		g.notifyBecameTarget(target, seat, played, events)
		return g.placeBingliang(seat, target, played, events)
	case CardShanDian:
		return g.placeShandian(seat, played, events)
	case CardWuGu:
		return g.resolveWugu(seat, events)
	case CardTaoYuan:
		return g.resolveTaoYuan(seat, events)
	case CardGuoHe, CardTanNang, CardJueDou, CardWuZhong, CardHuoGong:
		effectTarget := targetSpec.SeatIndex
		if effectTarget < 0 || effectTarget >= len(g.Players) {
			effectTarget = target
		}
		g.notifyBecameTarget(effectTarget, seat, played, events)
		responder := effectTarget
		if !g.isEnemy(seat, effectTarget) {
			responder = g.opponentOf(seat)
		}
		return g.startWuxiekTrickWindow(seat, responder, effectTarget, played, targetSpec, events)
	case CardNanMan:
		return g.resolveNanMan(seat, events)
	case CardWanJian:
		return g.resolveWanJian(seat, events)
	case CardTieSuo:
		// 铁索连环：类似南蛮入侵/桃园结义，逐人无懈窗口
		return g.resolveTieSuoAOE(seat, targetSpec, played, events)
	default:
		return ErrInvalidCard
	}
}

func (g *Game) placeLebu(source, target int, lebu Card, events *[]GameEvent) error {
	g.setJudgeCard(target, lebu)
	g.Players[target].SkipPlay = true
	g.Message = fmt.Sprintf("%s 被置入【乐不思蜀】", g.Players[target].Name)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: source,
		TargetIndex: target,
		Message:     g.Message,
	})
	return nil
}

func (g *Game) startWuxiekTrickWindow(source, responder, effectTarget int, trick Card, spec PlayTarget, events *[]GameEvent) error {
	n := len(g.Players)
	if source < 0 || source >= n {
		return fmt.Errorf("wuxiek: invalid source=%d", source)
	}
	if responder < 0 || responder >= n {
		responder = source
	}
	if effectTarget < 0 || effectTarget >= n {
		effectTarget = source
	}
	// 构建响应队列：从 responder 开始，轮询所有存活玩家，排除锦囊使用者自己
	allQueue := g.createResponseQueue(responder)
	queue := make([]int, 0, len(allQueue))
	for _, s := range allQueue {
		if s != source {
			queue = append(queue, s)
		}
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   source,
		TargetIndex:   -1,
		ReturnIndex:   source,
		EffectTarget:  effectTarget,
		Card:          trick,
		ResponseMode:  ResponseModeWuxiekTrick,
		TargetZone:    spec.Zone,
		TargetCardID:  spec.CardID,
		ResponseQueue: queue,
		ResponseIndex: 0,
		WuxiekChain:   nil,
	}
	g.advanceToNextWuxiekResponder(events)
	// 提示由 advanceToNextWuxiekResponder -> setWuxiekMessage 统一设置
	if g.Message == "" {
		g.Message = fmt.Sprintf("【%s】：是否使用【无懈可击】？", trick.Name)
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

// advanceToNextWuxiekResponder 推进到队列中的下一个响应者（每次只处理当前队列头的人）
// 如果是AI就处理并返回，如果是人类就停下等待。由 RunAIActionStep 外层循环驱动多步。
func (g *Game) advanceToNextWuxiekResponder(events *[]GameEvent) {
	if g.Pending == nil {
		return
	}
	if len(g.Pending.ResponseQueue) == 0 {
		g.finalizeWuxiekChain(events)
		return
	}
	idx := g.Pending.ResponseIndex
	if idx >= len(g.Pending.ResponseQueue) {
		idx = 0
	}
	nextSeat := g.Pending.ResponseQueue[idx]
	g.Pending.ActorSeat = nextSeat
	g.Pending.SubjectSeat = nextSeat

	// 如果是人类玩家，停下等待人类操作
	if !g.Players[nextSeat].IsAI {
		g.setWuxiekMessage()
		return
	}
	// 判定阶段无懈窗口：AI 不应主动出无懈抵消延时锦囊（延时锦囊对自己不利）
	// 反无懈窗口（WuxiekChain 非空）：AI 也不主动出反无懈
	if g.isJudgeWuxiekMode(g.Pending.ResponseMode) {
		_ = g.advanceJudgeWuxiekQueue(nextSeat, events)
		return
	}
	// AI 自动决定：该出无懈且有牌才出，否则跳过
	if shouldAIWuxiekTrick(g, nextSeat, g.Pending) {
		for _, card := range g.Players[nextSeat].Hand {
			if card.Kind == CardWuxiek {
				_ = g.RespondWuxiek(nextSeat, card.ID, events)
				return
			}
		}
	}
	// AI 没有无懈（或不该出），跳过
	g.advanceWuxiekQueueAfterPass(nextSeat, events)
}

// autoAIWuxiekRespond AI 自动决定是否出无懈可击
func (g *Game) autoAIWuxiekRespond(seat int, events *[]GameEvent) {
	// 判断该不该出无懈（目标是自己或队友）
	if !shouldAIWuxiekTrick(g, seat, g.Pending) {
		g.advanceWuxiekQueueAfterPass(seat, events)
		return
	}
	// 检查 AI 是否有无懈可击（且不在链中）
	inChain := false
	for _, entry := range g.Pending.WuxiekChain {
		if entry.Seat == seat {
			inChain = true
			break
		}
	}
	if !inChain {
		for _, card := range g.Players[seat].Hand {
			if card.Kind == CardWuxiek {
				_ = g.RespondWuxiek(seat, card.ID, events)
				return
			}
		}
	}
	// 没有无懈可击（或已在链中），跳过
	g.advanceWuxiekQueueAfterPass(seat, events)
}

func (g *Game) startWuxiekLebuJudgeWindow(seat int, events *[]GameEvent) {
	jc := g.Players[seat].judgeCardByKind(CardLeBu)
	if jc == nil {
		return
	}
	g.startJudgeWuxiekWindow(seat, *jc, events)
}

func (g *Game) continueTrickAfterWuxiekPass(events *[]GameEvent) error {
	pending := *g.Pending
	g.Pending = nil
	source := pending.SourceIndex
	target := pending.EffectTarget
	if target < 0 || target >= len(g.Players) {
		target = g.opponentOf(source)
	}
	spec := PlayTarget{SeatIndex: target, Zone: pending.TargetZone, CardID: pending.TargetCardID}

	g.Phase = PhasePlaying
	g.TurnStep = StepPlay

	var err error
	switch pending.Card.Kind {
	case CardGuoHe:
		err = g.resolveGuoHe(source, target, spec, events)
	case CardTanNang:
		err = g.resolveTanNang(source, target, spec, events)
	case CardJueDou:
		err = g.startCardResponse(source, target, pending.Card, CardSha, fmt.Sprintf("%s 对 %s 发起【决斗】，%s 需出杀", g.Players[source].Name, g.Players[target].Name, g.Players[target].Name), events)
	case CardWuZhong:
		g.Message = fmt.Sprintf("%s 使用【无中生有】，摸两张牌", g.Players[source].Name)
		*events = append(*events, GameEvent{Type: "trick_effect", PlayerIndex: source, TargetIndex: source, Message: g.Message})
		g.drawCards(source, 2, events)
	default:
		return ErrInvalidCard
	}
	if err != nil {
		return err
	}
	g.resetTimer()
	return nil
}

// playTrickAsGuoHe 执行过河拆桥效果（用于奇袭）
func (g *Game) playTrickAsGuoHe(seat int, target PlayTarget, events *[]GameEvent) error {
	// 创建一张虚拟的过河拆桥牌
	fakeCard := Card{
		ID:   "qixi_guobe",
		Kind:  CardGuoHe,
		Name:  "过河拆桥",
		Label: "过河拆桥",
	}
	
	targetSeat := target.SeatIndex
	if targetSeat < 0 || targetSeat >= len(g.Players) {
		targetSeat = g.opponentOf(seat)
	}
	
	if !g.isValidPlayTarget(seat, targetSeat, CardGuoHe) {
		return ErrInvalidTarget
	}
	
	if !g.hasTakeableCard(targetSeat) {
		return ErrInvalidTarget
	}
	
	g.notifyBecameTarget(targetSeat, seat, fakeCard, events)
	return g.startWuxiekTrickWindow(seat, targetSeat, targetSeat, fakeCard, target, events)
}

func (g *Game) applyLebuSkipDirect(seat int, events *[]GameEvent) {
	p := &g.Players[seat]
	p.SkipPlay = false
	if card, ok := g.removeJudgeByKind(seat, CardLeBu); ok {
		g.DiscardPile = append(g.DiscardPile, card)
	}
	*events = append(*events, GameEvent{
		Type:        "lebu_skip",
		PlayerIndex: seat,
		Message:     fmt.Sprintf("%s 受到【乐不思蜀】，跳过出牌阶段", p.Name),
	})
	_ = g.endTurn(events)
}

func (g *Game) applyLebuSkip(seat int, events *[]GameEvent) error {
	if g.offerDdzTrickCancelWindow(seat, ddzResumeLebu, events) {
		return nil
	}
	g.Pending = nil
	g.Phase = PhasePlaying
	g.applyLebuSkipDirect(seat, events)
	return nil
}

// cancelTrickWithWuxiek 有人出了无懈可击后，将其加入无懈链，继续问下一个人
func (g *Game) cancelTrickWithWuxiek(pending PendingCombat, events *[]GameEvent) error {
	wuxiekSeat := pending.TargetIndex // RespondWuxiek 中已记录打出无懈可击的人
	g.Pending = nil

	switch pending.ResponseMode {
	case ResponseModeWuxiekGuose:
		target := pending.EffectTarget
		g.removeJudgeByKind(target, CardLeBu)
		g.Players[target].SkipPlay = false
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = pending.SourceIndex
		g.Message = fmt.Sprintf("【国色】的【乐不思蜀】被【无懈可击】抵消")
		*events = append(*events, GameEvent{
			Type: "trick_cancelled", PlayerIndex: pending.SourceIndex, TargetIndex: target, Message: g.Message,
		})
		g.resetTimer()
	case ResponseModeWuxiekLebu:
		p := &g.Players[wuxiekSeat]
		p.SkipPlay = false
		g.removeJudgeByKind(wuxiekSeat, CardLeBu)
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = wuxiekSeat
		g.Message = fmt.Sprintf("【乐不思蜀】被【无懈可击】抵消，%s 可正常出牌", p.Name)
		g.resetTimer()
	case ResponseModeWuxiekBingliang:
		p := &g.Players[wuxiekSeat]
		p.SkipDraw = false
		g.removeJudgeByKind(wuxiekSeat, CardBingLiang)
		g.Phase = PhasePlaying
		g.TurnStep = StepDraw
		g.CurrentTurn = wuxiekSeat
		g.Message = fmt.Sprintf("【兵粮寸断】被【无懈可击】抵消，%s 正常摸牌", p.Name)
		g.drawCards(wuxiekSeat, g.drawCountFor(wuxiekSeat), events)
		if g.IsFinished() {
			return nil
		}
		if p.SkipPlay {
			if p.hasJudgeKind(CardLeBu) {
				g.startWuxiekLebuJudgeWindow(wuxiekSeat, events)
				return nil
			}
			g.applyLebuSkipDirect(wuxiekSeat, events)
			return nil
		}
		g.TurnStep = StepPlay
		g.resetTimer()
	case ResponseModeWuxiekShandian:
		g.removeJudgeByKind(wuxiekSeat, CardShanDian)
		g.Phase = PhasePlaying
		g.TurnStep = StepDraw
		g.CurrentTurn = wuxiekSeat
		g.Message = fmt.Sprintf("【闪电】被【无懈可击】抵消")
		*events = append(*events, GameEvent{
			Type: "trick_cancelled", PlayerIndex: pending.SourceIndex, TargetIndex: pending.TargetIndex,
			Card: &pending.Card, Message: g.Message,
		})
	default:
		// 锦囊被无懈可击抵消：记录到链中，继续问下一个
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = pending.SourceIndex
		g.Message = fmt.Sprintf("【%s】被【无懈可击】抵消", pending.Card.Name)
		*events = append(*events, GameEvent{
			Type: "trick_cancelled", PlayerIndex: pending.SourceIndex, TargetIndex: pending.TargetIndex,
			Card: &pending.Card, Message: g.Message,
		})
		if pending.Card.Kind == CardJueDou {
			g.tryJiangDraw(pending.SourceIndex, pending.Card, events)
		}
		g.resetTimer()
	}
	return nil
}

// setWuxiekMessage 根据当前无懈窗口状态设置提示消息
func (g *Game) setWuxiekMessage() {
	if g.Pending == nil {
		return
	}
	actor := g.Pending.ActorSeat
	if actor < 0 || actor >= len(g.Players) {
		return
	}
	actorName := g.Players[actor].Name
	chain := g.Pending.WuxiekChain
	trickName := g.Pending.Card.Name

	// 确定被无懈的目标描述
	targetDesc := ""
	sourceName := g.Players[g.Pending.SourceIndex].Name
	switch g.Pending.Card.Kind {
	case CardWuGu:
		pickerName := ""
		if g.Pending.WuguPickSeat >= 0 && g.Pending.WuguPickSeat < len(g.Players) {
			pickerName = g.Players[g.Pending.WuguPickSeat].Name
		}
		if len(chain) == 0 {
			targetDesc = fmt.Sprintf("%s 的选牌", pickerName)
		}
	case CardTaoYuan:
		effectName := ""
		if g.Pending.EffectTarget >= 0 && g.Pending.EffectTarget < len(g.Players) {
			effectName = g.Players[g.Pending.EffectTarget].Name
		}
		if len(chain) == 0 {
			targetDesc = fmt.Sprintf("%s 回复体力", effectName)
		}
	case CardNanMan:
		effectName := ""
		if g.Pending.EffectTarget >= 0 && g.Pending.EffectTarget < len(g.Players) {
			effectName = g.Players[g.Pending.EffectTarget].Name
		}
		if len(chain) == 0 {
			targetDesc = fmt.Sprintf("%s 受到的【南蛮入侵】", effectName)
		}
	case CardWanJian:
		effectName := ""
		if g.Pending.EffectTarget >= 0 && g.Pending.EffectTarget < len(g.Players) {
			effectName = g.Players[g.Pending.EffectTarget].Name
		}
		if len(chain) == 0 {
			targetDesc = fmt.Sprintf("%s 受到的【万箭齐发】", effectName)
		}
	default:
		// 普通锦囊：显示为 "XX 的【锦囊名】"
		if len(chain) == 0 {
			targetDesc = fmt.Sprintf("%s 的【%s】", sourceName, trickName)
		}
	}

	if actor == g.HumanPlayer {
		if len(chain) > 0 {
			lastName := g.Players[chain[len(chain)-1].Seat].Name
			g.Message = fmt.Sprintf("是否对 %s 的【无懈可击】使用【无懈可击】？", lastName)
		} else {
			g.Message = fmt.Sprintf("是否对【%s】使用【无懈可击】？", targetDesc)
		}
	} else {
		if len(chain) > 0 {
			lastName := g.Players[chain[len(chain)-1].Seat].Name
			g.Message = fmt.Sprintf("等待 %s 响应 %s 的【无懈可击】...", actorName, lastName)
		} else {
			g.Message = fmt.Sprintf("等待 %s 决定是否出【无懈可击】...", actorName)
		}
	}
}

// advanceWuxiekQueueAfterPass 当前响应者跳过，推进到下一个
func (g *Game) advanceWuxiekQueueAfterPass(seat int, events *[]GameEvent) {
	if g.Pending == nil || len(g.Pending.ResponseQueue) == 0 {
		g.finalizeWuxiekChain(events)
		return
	}
	// 记录队列起始座位
	queueStart := g.Pending.ResponseQueue[0]
	g.Pending.ResponseIndex++
	if g.Pending.ResponseIndex >= len(g.Pending.ResponseQueue) {
		g.Pending.ResponseIndex = 0
	}
	// 如果队列只有一个人，或已回到队列起点，终止
	if len(g.Pending.ResponseQueue) == 1 || g.Pending.ResponseQueue[g.Pending.ResponseIndex] == queueStart {
		g.finalizeWuxiekChain(events)
		return
	}
	g.advanceToNextWuxiekResponder(events)
}

// nextAliveSeat 返回指定座位的下一个存活玩家
func (g *Game) nextAliveSeat(seat int) int {
	n := len(g.Players)
	for i := 1; i <= n; i++ {
		next := (seat + i) % n
		if g.Players[next].HP > 0 {
			return next
		}
	}
	return seat
}

// finalizeWuxiekChain 统计无懈可击链，决定锦囊是否生效
func (g *Game) finalizeWuxiekChain(events *[]GameEvent) {
	if g.Pending == nil {
		return
	}
	chainLen := len(g.Pending.WuxiekChain)
	source := g.Pending.SourceIndex
	effectTarget := g.Pending.EffectTarget
	trick := g.Pending.Card
	spec := PlayTarget{SeatIndex: effectTarget, Zone: g.Pending.TargetZone, CardID: g.Pending.TargetCardID}

	// 保存 AOE 队列信息和五谷状态（在清空 Pending 之前）
	aoeQueue := g.Pending.AoeQueue
	isWanJian := trick.Kind == CardWanJian
	isNanMan := trick.Kind == CardNanMan
	isTaoYuan := trick.Kind == CardTaoYuan
	isWuGu := trick.Kind == CardWuGu
	isTieSuo := trick.Kind == CardTieSuo
	wuguPicker := g.Pending.WuguPickSeat
	wuguRevealed := append([]Card(nil), g.Pending.RevealedCards...)
	wuguRevealedAll := append([]Card(nil), g.Pending.WuguRevealedAll...)

	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source

	if chainLen%2 == 1 {
		// 奇数张无懈可击 → 锦囊被抵消（或万箭/南蛮/桃园/五谷玩家跳过）
		if isWanJian {
			g.Message = fmt.Sprintf("【无懈可击】阻止了 %s 受到的【万箭齐发】", g.Players[effectTarget].Name)
			*events = append(*events, GameEvent{
				Type:        "trick_cancelled",
				PlayerIndex: effectTarget,
				TargetIndex: effectTarget,
				Card:        &trick,
				Message:     g.Message,
			})
			g.continueWanJianAfterTarget(source, aoeQueue, events)
			return
		}
		if isNanMan {
			g.Message = fmt.Sprintf("【无懈可击】阻止了 %s 受到的【南蛮入侵】", g.Players[effectTarget].Name)
			*events = append(*events, GameEvent{
				Type:        "trick_cancelled",
				PlayerIndex: effectTarget,
				TargetIndex: effectTarget,
				Card:        &trick,
				Message:     g.Message,
			})
			g.continueNanManAfterTarget(source, aoeQueue, events)
			return
		}
		if isTaoYuan {
			g.Message = fmt.Sprintf("【无懈可击】阻止了 %s 回复体力", g.Players[effectTarget].Name)
			*events = append(*events, GameEvent{
				Type:        "trick_cancelled",
				PlayerIndex: effectTarget,
				TargetIndex: effectTarget,
				Card:        &trick,
				Message:     g.Message,
			})
			g.continueTaoYuanAfterTarget(source, aoeQueue, events)
			return
		}
		if isWuGu {
			g.Message = fmt.Sprintf("【无懈可击】阻止了 %s 选牌", g.Players[wuguPicker].Name)
			// 标记当前 picker 已被跳过
			g.wuguPicked[wuguPicker] = true
			next := g.nextWuguPicker(wuguPicker, source)
			if next == source {
				g.finishWugu(source, events)
				return
			}
			g.startWuguPickForWithAll(source, next, wuguRevealed, wuguRevealedAll, events)
			return
		}
		if isTieSuo {
			g.Message = fmt.Sprintf("【无懈可击】阻止了 %s 受到【铁索连环】", g.Players[effectTarget].Name)
			*events = append(*events, GameEvent{
				Type:        "trick_cancelled",
				PlayerIndex: effectTarget,
				TargetIndex: effectTarget,
				Card:        &trick,
				Message:     g.Message,
			})
			g.continueTieSuoAfter(source, trick, aoeQueue, events)
			return
		}
		g.Message = fmt.Sprintf("【%s】被【无懈可击】抵消", trick.Name)
		g.resetTimer()
		return
	}
	// 偶数张（含0）→ 锦囊生效（或万箭/南蛮/桃园/五谷玩家继续）
	if isWanJian {
		card := Card{Kind: CardWanJian, Name: "万箭齐发"}
		msg := fmt.Sprintf("【万箭齐发】：%s 需出【闪】", g.Players[effectTarget].Name)
		g.Phase = PhaseResponse
		g.Pending = &PendingCombat{
			SourceIndex:  source,
			TargetIndex:  effectTarget,
			ReturnIndex:  source,
			Card:         card,
			RequiredKind: CardShan,
			Damage:       1,
			AoeQueue:     aoeQueue,
		}
		FillPendingRoles(g.Pending)
		g.Message = msg
		g.resetTimer()
		*events = append(*events, GameEvent{
			Type:        "trick_response",
			PlayerIndex: source,
			TargetIndex: effectTarget,
			Message:     msg,
		})
		return
	}
	if isNanMan {
		// 无懈通过：虚拟电脑对 target 使用决斗，target 需出杀，不出杀则扣血
		// 用 startCardResponse 进入出杀阶段，rest 保存剩余队列，通过后继续下一个
		card := Card{Kind: CardNanMan, Name: "南蛮入侵"}
		msg := fmt.Sprintf("【南蛮入侵】：%s 需出【杀】", g.Players[effectTarget].Name)
		g.Phase = PhaseResponse
		g.Pending = &PendingCombat{
			SourceIndex:  source,
			TargetIndex:  effectTarget,
			ReturnIndex:  source,
			Card:         card,
			RequiredKind: CardSha,
			Damage:       1,
			AoeQueue:     aoeQueue,
		}
		FillPendingRoles(g.Pending)
		g.Message = msg
		g.resetTimer()
		*events = append(*events, GameEvent{
			Type:        "trick_response",
			PlayerIndex: source,
			TargetIndex: effectTarget,
			Message:     msg,
		})
		return
	}
	if isTaoYuan {
		if g.Players[effectTarget].HP < g.Players[effectTarget].MaxHP {
			g.Players[effectTarget].HP++
			*events = append(*events, GameEvent{
				Type:        "trick_heal",
				PlayerIndex: source,
				TargetIndex: effectTarget,
				Heal:        1,
				Message:     fmt.Sprintf("%s 回复 1 点体力", g.Players[effectTarget].Name),
			})
		}
		g.continueTaoYuanAfterTarget(source, aoeQueue, events)
		return
	}
	if isWuGu {
		// 无懈通过（偶数链），picker 直接选牌
		g.wuguPickPass(wuguPicker, wuguRevealed, wuguRevealedAll, source, events)
		return
	}
	if isTieSuo {
		// 无懈通过 → 对当前目标执行横置/重置，然后继续下一个
		g.resolveTieSuoChain(source, effectTarget, trick, events)
		for i, pl := range g.Players {
			Logf("finalizeWuxiekChain TieSuo: Player[%d]=%s chained=%v", i, pl.Name, g.isChained(i))
		}
		g.continueTieSuoAfter(source, trick, aoeQueue, events)
		return
	}
	n := len(g.Players)
	if source < 0 || source >= n {
		g.resetTimer()
		return
	}
	if effectTarget < 0 || effectTarget >= n {
		effectTarget = source
	}
	g.continueTrickAfterWuxiekPassDirect(source, effectTarget, trick, spec, events)
}

// continueTrickAfterWuxiekPassDirect 锦囊生效（无懈窗口结束）
func (g *Game) continueTrickAfterWuxiekPassDirect(source, target int, trick Card, spec PlayTarget, events *[]GameEvent) {
	switch trick.Kind {
	case CardGuoHe:
		_ = g.resolveGuoHe(source, target, spec, events)
	case CardTanNang:
		_ = g.resolveTanNang(source, target, spec, events)
	case CardJueDou:
		_ = g.startCardResponse(source, target, trick, CardSha, fmt.Sprintf("%s 对 %s 发起【决斗】，%s 需出杀", g.Players[source].Name, g.Players[target].Name, g.Players[target].Name), events)
	case CardHuoGong:
		_ = g.playHuoGong(source, trick, target, events)
	case CardTieSuo:
		// 铁索连环通常在 finalizeWuxiekChain 中直接处理（保留双目标信息）
		// 此分支仅作为兜底：对单目标执行横置/重置
		g.resolveTieSuoChain(source, target, trick, events)
	case CardWuZhong:
		g.Message = fmt.Sprintf("%s 使用【无中生有】，摸两张牌", g.Players[source].Name)
		*events = append(*events, GameEvent{Type: "trick_effect", PlayerIndex: source, TargetIndex: source, Message: g.Message})
		g.drawCards(source, 2, events)
	}
	g.resetTimer()
}

func (g *Game) cancelAoeSelfWithWuxiek(pending PendingCombat, events *[]GameEvent) error {
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = pending.ReturnIndex
	g.Message = fmt.Sprintf("%s 用【无懈可击】抵消了【%s】对自己的效果", g.Players[pending.TargetIndex].Name, pending.Card.Name)
	*events = append(*events, GameEvent{
		Type:        "trick_cancelled",
		PlayerIndex: pending.TargetIndex,
		TargetIndex: pending.TargetIndex,
		Card:        &pending.Card,
		Message:     g.Message,
	})
	g.resetTimer()
	return nil
}

func (g *Game) playJiu(seat int, cardID string, events *[]GameEvent) error {
	if g.Players[seat].Drunk {
		return ErrAlreadyActed
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || cardObj.Kind != CardJiu {
		return ErrInvalidCard
	}
	played := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, played)
	g.Players[seat].Drunk = true
	g.Message = fmt.Sprintf("%s 使用【酒】，本回合下一张杀伤害+1", g.Players[seat].Name)
	*events = append(*events, GameEvent{
		Type:        "play_jiu",
		PlayerIndex: seat,
		TargetIndex: seat,
		Card:        &played,
		Message:     g.Message,
	})
	return nil
}

func (g *Game) playEquip(seat int, cardID string, events *[]GameEvent) error {
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || equipSlot(cardObj.Kind) == "" {
		return ErrInvalidCard
	}
	played := g.removeHandCard(seat, idx, events)
	old := g.setEquipment(seat, played)
	if old != nil {
		g.DiscardPile = append(g.DiscardPile, *old)
		g.notifyEquipLost(seat, *old, "replace", events)
	}
	g.Message = fmt.Sprintf("%s 装备【%s】", g.Players[seat].Name, played.Name)
	*events = append(*events, GameEvent{
		Type:        "equip",
		PlayerIndex: seat,
		TargetIndex: seat,
		Card:        &played,
		Message:     g.Message,
	})
	return nil
}

func equipSlot(kind string) string {
	switch kind {
	case CardWeapon1, CardWeapon2, CardWeapon3, CardWeapon4, CardWeapon5, CardWeapon6, CardWeapon7, CardWeapon8, CardWeapon9:
		return EquipWeapon
	case CardArmor, CardArmorVine:
		return EquipArmor
	case CardPlusHorse:
		return EquipPlusHorse
	case CardMinusHorse:
		return EquipMinusHorse
	default:
		return ""
	}
}

func (g *Game) setEquipment(seat int, card Card) *Card {
	p := &g.Players[seat]
	switch equipSlot(card.Kind) {
	case EquipWeapon:
		old := p.Weapon
		p.Weapon = &card
		return old
	case EquipArmor:
		old := p.Armor
		p.Armor = &card
		return old
	case EquipPlusHorse:
		old := p.PlusHorse
		p.PlusHorse = &card
		return old
	case EquipMinusHorse:
		old := p.MinusHorse
		p.MinusHorse = &card
		return old
	default:
		return nil
	}
}

func trickNeedsOpponentTarget(kind string) bool {
	switch kind {
	case CardGuoHe, CardTanNang, CardJueDou, CardLeBu, CardBingLiang, CardHuoGong, CardTieSuo:
		return true
	default:
		return false
	}
}

func trickNeedsSelfTarget(kind string) bool {
	switch kind {
	case CardShanDian, CardWuGu, CardTieSuo:
		return true
	default:
		return false
	}
}

func (g *Game) hasTakeableCard(target int) bool {
	p := &g.Players[target]
	return len(p.Hand) > 0 ||
		p.Weapon != nil ||
		p.Armor != nil ||
		p.PlusHorse != nil ||
		p.MinusHorse != nil ||
		len(p.JudgeArea) > 0
}

func (g *Game) takeTargetCard(target int, spec PlayTarget, events *[]GameEvent) (Card, string, bool) {
	p := &g.Players[target]
	zone := spec.Zone
	var card Card
	var label string
	var ok bool
	switch zone {
	case "", "hand":
		if len(p.Hand) == 0 {
			return Card{}, "", false
		}
		card = p.Hand[0]
		p.Hand = p.Hand[1:]
		g.syncCounts()
		g.runHandEmptyHooks(target, events)
		return card, "手牌", true
	case EquipWeapon:
		card, label, ok = takeCardPtr(&p.Weapon, spec.CardID, "武器")
	case EquipArmor:
		card, label, ok = takeCardPtr(&p.Armor, spec.CardID, "八卦阵")
	case EquipPlusHorse:
		card, label, ok = takeCardPtr(&p.PlusHorse, spec.CardID, "+1马")
	case EquipMinusHorse:
		card, label, ok = takeCardPtr(&p.MinusHorse, spec.CardID, "-1马")
	case "judge":
		return g.takeJudgeCard(target, spec.CardID)
	default:
		return Card{}, "", false
	}
	if ok && isEquipZone(zone) {
		g.notifyEquipLost(target, card, "taken", events)
	}
	return card, label, ok
}

func takeCardPtr(slot **Card, cardID string, label string) (Card, string, bool) {
	if *slot == nil {
		return Card{}, "", false
	}
	if cardID != "" && (*slot).ID != cardID {
		return Card{}, "", false
	}
	card := **slot
	*slot = nil
	return card, label, true
}

func (g *Game) resolveGuoHe(seat, target int, spec PlayTarget, events *[]GameEvent) error {
	// 无懈通过后，打开选牌窗口，让玩家选择要拆掉的牌
	msg := fmt.Sprintf("%s 对 %s 使用【过河拆桥】，请选择要拆掉的牌", g.Players[seat].Name, g.Players[target].Name)
	err := g.OpenTakeWindow(TakeWindowConfig{
		SkillID:          "", // 不是技能，是锦囊牌
		ResponseMode:     ResponseModeGuoHe,
		ActorSeat:        seat,
		SubjectSeat:      target,
		OriginSeat:       seat,
		MaxTake:          1,
		Destination:      TakeDestination{Zone: ZoneVoid, Seat: -1}, // 弃置到虚无（直接进弃牌堆）
		Message:          msg,
		EventType:        "guohe_discard",
		PassClosesWindow: true,
		OnEachTake:       guoheOnEachTake,
		OnComplete:       guoheTakeComplete,
	}, events)
	if err != nil {
		return err
	}
	return nil
}

func guoheOnEachTake(g *Game, card Card, label string, events *[]GameEvent) error {
	// 将牌放入弃牌堆
	g.DiscardPile = append(g.DiscardPile, card)
	source := g.Pending.ActorSeat
	target := g.Pending.SubjectSeat
	msg := fmt.Sprintf("%s 拆掉 %s 的%s", g.Players[source].Name, g.Players[target].Name, label)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: source,
		TargetIndex: target,
		Card:        &card,
		Message:     msg,
	})
	return nil
}

func guoheTakeComplete(g *Game, events *[]GameEvent) error {
	// 选牌完成，回到出牌阶段
	// 先保存 ActorSeat（出牌者），因为 Pending 会被清空
	source := -1
	if g.Pending != nil {
		source = g.Pending.ActorSeat
	}
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	if source < 0 {
		// 如果 ActorSeat 无效，尝试从事件中获取
		for i := len(*events) - 1; i >= 0; i-- {
			if (*events)[i].Type == "play_trick" && (*events)[i].Card != nil && (*events)[i].Card.Kind == CardGuoHe {
				source = (*events)[i].PlayerIndex
				break
			}
		}
	}
	g.CurrentTurn = source
	g.resetTimer()
	return nil
}

func (g *Game) resolveTanNang(seat, target int, spec PlayTarget, events *[]GameEvent) error {
	// 无懈通过后，打开选牌窗口，让玩家选择要获得的牌
	msg := fmt.Sprintf("%s 对 %s 使用【顺手牵羊】，请选择要获得的牌", g.Players[seat].Name, g.Players[target].Name)
	err := g.OpenTakeWindow(TakeWindowConfig{
		SkillID:          "", // 不是技能，是锦囊牌
		ResponseMode:     ResponseModeTanNang,
		ActorSeat:        seat,
		SubjectSeat:      target,
		OriginSeat:       seat,
		MaxTake:          1,
		Destination:      TakeDestination{Zone: ZoneHand, Seat: seat}, // 获得的牌放入手牌
		Message:          msg,
		EventType:        "tannang_take",
		PassClosesWindow: true,
		OnEachTake:       tannangOnEachTake,
		OnComplete:       tannangTakeComplete,
	}, events)
	if err != nil {
		return err
	}
	return nil
}

func tannangOnEachTake(g *Game, card Card, label string, events *[]GameEvent) error {
	// 牌已经通过 TakeOne 函数放入目标区域（手牌）
	source := g.Pending.ActorSeat
	target := g.Pending.SubjectSeat
	msg := fmt.Sprintf("%s 获得 %s 的%s", g.Players[source].Name, g.Players[target].Name, label)
	g.Message = msg
	g.syncCounts()
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: source,
		TargetIndex: target,
		Card:        &card,
		Message:     msg,
		Amount:      1,
	})
	return nil
}

func tannangTakeComplete(g *Game, events *[]GameEvent) error {
	// 选牌完成，回到出牌阶段
	// 先保存 ActorSeat（出牌者），因为 Pending 会被清空
	source := -1
	if g.Pending != nil {
		source = g.Pending.ActorSeat
	}
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	if source < 0 {
		// 从事件中获取来源座位号
		for i := len(*events) - 1; i >= 0; i-- {
			if (*events)[i].Type == "play_trick" && (*events)[i].Card != nil && (*events)[i].Card.Kind == CardTanNang {
				source = (*events)[i].PlayerIndex
				break
			}
		}
	}
	g.CurrentTurn = source
	g.resetTimer()
	return nil
}

func (g *Game) resolveTaoYuan(seat int, events *[]GameEvent) error {
	g.skipWuxiekSeats = nil // 新群体锦囊，清空跳过标记
	// 构建需要回复的玩家队列（从使用者开始，跳过满血玩家）
	queue := make([]int, 0, len(g.Players))
	n := len(g.Players)
	for i := 0; i < n; i++ {
		s := (seat + i) % n
		if g.Players[s].HP > 0 && g.Players[s].HP < g.Players[s].MaxHP {
			queue = append(queue, s)
		}
	}
	if len(queue) == 0 {
		g.Message = fmt.Sprintf("%s 使用【桃园结义】，无人需要回复", g.Players[seat].Name)
		return nil
	}
	// 宣告：发事件让前端显示，和五谷亮牌一样
	g.Message = fmt.Sprintf("%s 使用【桃园结义】，依次回复体力", g.Players[seat].Name)
	*events = append(*events, GameEvent{
		Type:        "taoyuan_announce",
		PlayerIndex: seat,
		TargetIndex: seat,
		Message:     g.Message,
	})
	// 直接开始第一个人的回复（带无懈可击窗口），和五谷 startWuguPickFor 一样
	g.startTaoYuanHeal(seat, queue[0], queue[1:], events)
	return nil
}

// resolveNanMan 南蛮入侵：宣告 → 逐人无懈窗口 → 无懈通过则进入决斗
func (g *Game) resolveNanMan(source int, events *[]GameEvent) error {
	g.skipWuxiekSeats = nil // 新群体锦囊，清空跳过标记
	// 构建受影响玩家队列（从使用者下家开始，过滤藤甲）
	allQueue := g.aoeResponderQueue(source)
	queue := g.filterAoeQueue(allQueue, CardNanMan)
	if len(queue) == 0 {
		g.Message = fmt.Sprintf("%s 使用【南蛮入侵】，无人受影响", g.Players[source].Name)
		return nil
	}
	// 宣告：发事件让前端显示
	g.Message = fmt.Sprintf("%s 使用【南蛮入侵】", g.Players[source].Name)
	*events = append(*events, GameEvent{
		Type:        "nanman_announce",
		PlayerIndex: source,
		TargetIndex: source,
		Message:     g.Message,
	})
	// 直接开始第一个人的无懈窗口（和五谷/桃园一样）
	g.startNanManJueDou(source, queue[0], queue[1:], events)
	return nil
}

// filterSkipWuxiek 从队列中移除 skipWuxiekSeats 标记的座位
func (g *Game) filterSkipWuxiek(queue []int) []int {
	if len(g.skipWuxiekSeats) == 0 {
		return queue
	}
	out := make([]int, 0, len(queue))
	for _, s := range queue {
		if !g.skipWuxiekSeats[s] {
			out = append(out, s)
		}
	}
	return out
}

// startNanManJueDou 对单个目标发起南蛮决斗无懈窗口
// 虚拟电脑对 target 使用决斗，target 第一个被询问是否无懈，然后是其他人
func (g *Game) startNanManJueDou(source, target int, rest []int, events *[]GameEvent) {
	trick := Card{Kind: CardNanMan, Name: "南蛮入侵"}
	// 构建无懈响应队列：target 第一个，然后从 target 下家开始轮询
	allQueue := g.createResponseQueue((target + 1) % len(g.Players))
	queue := make([]int, 0, len(allQueue)+1)
	queue = append(queue, target) // target 第一个
	for _, s := range allQueue {
		if s != target {
			queue = append(queue, s)
		}
	}
	queue = g.filterSkipWuxiek(queue)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   source,
		TargetIndex:   -1,
		ReturnIndex:   source,
		EffectTarget:  target,
		Card:          trick,
		ResponseMode:  ResponseModeWuxiekTrick,
		AoeQueue:      rest,
		ResponseQueue: queue,
		ResponseIndex: 0,
		WuxiekChain:   nil,
	}
	g.advanceToNextWuxiekResponder(events)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: source,
		TargetIndex: target,
		Card:        &trick,
		Message:     g.Message,
	})
}

// continueNanManAfterTarget 南蛮入侵：当前目标处理完毕，继续下一个
func (g *Game) continueNanManAfterTarget(source int, rest []int, events *[]GameEvent) {
	if len(rest) == 0 {
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = source
		g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
		g.resetTimer()
		return
	}
	next := rest[0]
	newRest := append([]int(nil), rest[1:]...)
	g.startNanManJueDou(source, next, newRest, events)
}

// resolveWanJian 万箭齐发：宣告 → 逐人无懈窗口 → 无懈通过则需出闪
func (g *Game) resolveWanJian(source int, events *[]GameEvent) error {
	g.skipWuxiekSeats = nil // 新群体锦囊，清空跳过标记
	allQueue := g.aoeResponderQueue(source)
	queue := g.filterAoeQueue(allQueue, CardWanJian)
	if len(queue) == 0 {
		g.Message = fmt.Sprintf("%s 使用【万箭齐发】，无人受影响", g.Players[source].Name)
		return nil
	}
	g.Message = fmt.Sprintf("%s 使用【万箭齐发】", g.Players[source].Name)
	*events = append(*events, GameEvent{
		Type:        "wanjian_announce",
		PlayerIndex: source,
		TargetIndex: source,
		Message:     g.Message,
	})
	g.startWanJianShan(source, queue[0], queue[1:], events)
	return nil
}

func (g *Game) startWanJianShan(source, target int, rest []int, events *[]GameEvent) {
	trick := Card{Kind: CardWanJian, Name: "万箭齐发"}
	allQueue := g.createResponseQueue((target + 1) % len(g.Players))
	queue := make([]int, 0, len(allQueue)+1)
	queue = append(queue, target)
	for _, s := range allQueue {
		if s != target {
			queue = append(queue, s)
		}
	}
	queue = g.filterSkipWuxiek(queue)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   source,
		TargetIndex:   -1,
		ReturnIndex:   source,
		EffectTarget:  target,
		Card:          trick,
		ResponseMode:  ResponseModeWuxiekTrick,
		AoeQueue:      rest,
		ResponseQueue: queue,
		ResponseIndex: 0,
		WuxiekChain:   nil,
	}
	g.advanceToNextWuxiekResponder(events)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: source,
		TargetIndex: target,
		Card:        &trick,
		Message:     g.Message,
	})
}

func (g *Game) continueWanJianAfterTarget(source int, rest []int, events *[]GameEvent) {
	if len(rest) == 0 {
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = source
		g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
		g.resetTimer()
		return
	}
	next := rest[0]
	newRest := append([]int(nil), rest[1:]...)
	g.startWanJianShan(source, next, newRest, events)
}

// restorePendingAfterDying 濒死结算结束后恢复之前保存的 Pending
func (g *Game) restorePendingAfterDying(saved *PendingCombat, events *[]GameEvent) bool {
	if saved == nil {
		return false
	}
	source := saved.SourceIndex
	queue := saved.AoeQueue
	switch {
	case saved.RequiredKind == "tiesuo":
		// 铁索连环AOE恢复：濒死结束后继续下一个人
		// 链式伤害值 = saved.Damage（上一个人的最终伤害值）
		Logf("restorePendingAfterDying: tiesuo aoe resume, source=%d amount=%d rest=%v", source, saved.Damage, queue)
		g.continueTiesuoAoe(source, saved.Damage, saved.Card, queue, events)
		return true
	case saved.Card.Kind == CardNanMan:
		g.continueNanManAfterTarget(source, queue, events)
		return true
	case saved.Card.Kind == CardWanJian:
		g.continueWanJianAfterTarget(source, queue, events)
		return true
	case saved.ResponseMode == ResponseModeWuguPick:
		// 五谷丰登选牌被濒死中断，恢复选牌流程
		g.Pending = saved
		g.Phase = PhaseResponse
		g.resetTimer()
		return true
	default:
		// 恢复通用 Pending
		g.Pending = saved
		g.Phase = PhaseResponse
		g.resetTimer()
		return true
	}
}

func (g *Game) startTaoYuanHeal(source, target int, rest []int, events *[]GameEvent) {
	trick := Card{Kind: CardTaoYuan, Name: "桃园结义"}
	// 构建无懈响应队列：从 target 的下一个玩家开始，轮询所有存活玩家，排除 target（回复者本人）
	allQueue := g.createResponseQueue((target + 1) % len(g.Players))
	queue := make([]int, 0, len(allQueue))
	for _, s := range allQueue {
		if s != target {
			queue = append(queue, s)
		}
	}
	queue = g.filterSkipWuxiek(queue)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   source,
		TargetIndex:   -1,
		ReturnIndex:   source,
		EffectTarget:  target,
		Card:          trick,
		ResponseMode:  ResponseModeWuxiekTrick,
		AoeQueue:      rest,
		ResponseQueue: queue,
		ResponseIndex: 0,
		WuxiekChain:   nil,
	}
	g.advanceToNextWuxiekResponder(events)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "taoyuan_offer",
		PlayerIndex: source,
		TargetIndex: target,
		Message:     g.Message,
	})
}

// continueTaoYuanAfterTarget 桃园结义：当前目标处理完毕，继续下一个
func (g *Game) continueTaoYuanAfterTarget(source int, rest []int, events *[]GameEvent) {
	// 跳过满血玩家
	for len(rest) > 0 {
		next := rest[0]
		if g.Players[next].HP < g.Players[next].MaxHP {
			newRest := append([]int(nil), rest[1:]...)
			g.startTaoYuanHeal(source, next, newRest, events)
			return
		}
		// 满血跳过
		rest = rest[1:]
	}
	// 无人需要回复
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
	g.resetTimer()
}

func (g *Game) startCardResponse(seat, target int, card Card, requiredKind string, message string, events *[]GameEvent) error {
	g.Phase = PhaseResponse
	allowWuxiek := card.Kind == CardNanMan || card.Kind == CardWanJian
	g.appendWushuangMessage(seat, card.Kind, &message)
	
	// 计算伤害：决斗伤害可能受裸衣影响
	damage := 1
	if card.Kind == CardJueDou && g.getSkillCounter(seat, counterLuoyiActive) > 0 {
		damage = 2  // 裸衣+1
	}
	
	g.Pending = &PendingCombat{
		SourceIndex:     seat,
		TargetIndex:     target,
		ReturnIndex:     seat,
		Card:            card,
		RequiredKind:    requiredKind,
		Damage:          damage,
		AllowWuxiek:     allowWuxiek,
		ResponsesNeeded: g.wushuangResponsesNeeded(seat, card.Kind),
	}
	g.Message = message
	if allowWuxiek {
		g.Message = message + "，或出【无懈可击】仅抵消对自己的效果"
	}
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "trick_response",
		PlayerIndex: seat,
		TargetIndex: target,
		Message:     g.Message,
	})
	return nil
}

func (g *Game) startAoeTrick(source int, card Card, requiredKind string, events *[]GameEvent) error {
	queue := g.filterAoeQueue(g.aoeResponderQueue(source), card.Kind)
	if len(queue) == 0 {
		return nil
	}
	rest := append([]int(nil), queue[1:]...)
	if err := g.startAoeResponse(source, queue[0], card, requiredKind, rest, events); err != nil {
		return err
	}
	return nil
}

func (g *Game) startAoeResponse(source, target int, card Card, requiredKind string, rest []int, events *[]GameEvent) error {
	msg := fmt.Sprintf("%s 受到【%s】，需出%s", g.Players[target].Name, card.Name, cardLabel(requiredKind))
	g.Phase = PhaseResponse
	g.appendWushuangMessage(source, card.Kind, &msg)
	g.Pending = &PendingCombat{
		SourceIndex:     source,
		TargetIndex:     target,
		ReturnIndex:     source,
		Card:            card,
		RequiredKind:    requiredKind,
		Damage:          1,
		AllowWuxiek:     true,
		ResponsesNeeded: g.wushuangResponsesNeeded(source, card.Kind),
		AoeQueue:        rest,
	}
	g.Message = msg + "，或出【无懈可击】仅抵消对自己的效果"
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "trick_response",
		PlayerIndex: source,
		TargetIndex: target,
		Message:     g.Message,
	})
	return nil
}

func cardLabel(kind string) string {
	switch kind {
	case CardSha:
		return "【杀】"
	case CardShan:
		return "【闪】"
	default:
		return kind
	}
}

func (g *Game) continueAoeAfterTarget(source int, card Card, requiredKind string, queue []int, events *[]GameEvent) error {
	if len(queue) == 0 {
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = source
		g.Message = fmt.Sprintf("%s 继续出牌", g.Players[source].Name)
		g.resetTimer()
		return nil
	}
	rest := append([]int(nil), queue[1:]...)
	return g.startAoeResponse(source, queue[0], card, requiredKind, rest, events)
}

func (g *Game) playTao(seat int, cardID string, events *[]GameEvent) error {
	if seat != g.CurrentTurn {
		return ErrNotYourTurn
	}
	p := &g.Players[seat]
	if p.HP >= p.MaxHP {
		return ErrInvalidCard
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || cardObj.Kind != CardTao {
		return ErrInvalidCard
	}
	played := g.removeHandCard(seat, idx, events)
	return g.playTaoWithCard(seat, played, events)
}

// playTaoWithCard 用已移除的牌当桃使用（支持装备牌变牌）
func (g *Game) playTaoWithCard(seat int, played Card, events *[]GameEvent) error {
	p := &g.Players[seat]
	if p.HP >= p.MaxHP {
		return ErrInvalidCard
	}
	g.DiscardPile = append(g.DiscardPile, played)
	p.HP++
	g.Message = fmt.Sprintf("%s 使用【桃】，体力 %d/%d", p.Name, p.HP, p.MaxHP)
	*events = append(*events, GameEvent{
		Type:        "play_tao",
		PlayerIndex: seat,
		Card:        &played,
		Heal:        1,
		Message:     g.Message,
	})
	return nil
}
