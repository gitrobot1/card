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
			target := g.Pending.EffectTarget
			if target < 0 {
				target = seat
			}
			return g.playLuanwuSha(seat, cardID, target, events)
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

	_, cardObj, ok := g.findCard(seat, cardID)
	if !ok {
		return ErrInvalidCard
	}

	if cardObj.Kind != CardSha && g.cardPlaysAs(seat, cardObj, CardSha) {
		if g.canUseSha(seat) && g.isValidPlayTarget(seat, target.SeatIndex, CardSha) {
			return g.playSha(seat, cardID, target.SeatIndex, events)
		}
	}

	switch cardObj.Kind {
	case CardSha:
		return g.playSha(seat, cardID, target.SeatIndex, events)
	case CardShan:
		return ErrInvalidCard
	case CardTao:
		return g.playTao(seat, cardID, events)
	case CardJiu:
		return g.playJiu(seat, cardID, events)
	case CardGuoHe, CardTanNang, CardNanMan, CardWanJian, CardJueDou, CardLeBu, CardBingLiang, CardShanDian, CardWuGu, CardTaoYuan, CardWuZhong, CardHuoGong, CardTieSuo:
		return g.playTrick(seat, cardID, target, events)
	case CardWeapon1, CardWeapon2, CardWeapon3, CardWeapon4, CardWeapon5, CardWeapon6, CardArmor, CardArmorVine, CardPlusHorse, CardMinusHorse:
		return g.playEquip(seat, cardID, events)
	default:
		return ErrInvalidCard
	}
}

func (g *Game) playSha(seat int, cardID string, targetIndex int, events *[]GameEvent) error {
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
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || !g.cardPlaysAs(seat, cardObj, CardSha) {
		return ErrInvalidCard
	}

	played := g.removeHandCard(seat, idx, events)
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
	}
	g.initPojunOnShaPending(seat, targetIndex, g.Pending)
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
	if g.is3v3() && cardObj.Kind == CardShanDian {
		return ErrInvalidCard
	}
	if g.isIdentity() && cardObj.Kind == CardShanDian {
		return ErrInvalidCard
	}

	target := targetSpec.SeatIndex
	if trickNeedsOpponentTarget(cardObj.Kind) {
		if target < 0 || target >= len(g.Players) || !g.isValidPlayTarget(seat, target, cardObj.Kind) {
			return ErrInvalidTarget
		}
	} else if target < 0 || target >= len(g.Players) {
		target = g.opponentOf(seat)
	}
	if (cardObj.Kind == CardGuoHe || cardObj.Kind == CardTanNang) && !g.hasTakeableCard(target) {
		return ErrInvalidTarget
	}
	if cardObj.Kind == CardBingLiang && !g.canBingliangTarget(seat, target) {
		return ErrInvalidTarget
	}
	if cardObj.Kind == CardHuoGong && len(g.Players[target].Hand) == 0 {
		return ErrInvalidTarget
	}
	if cardObj.Kind == CardTieSuo && target == seat {
		played := g.removeHandCard(seat, idx, events)
		g.runCardsDiscardedHooks(seat, "play", []Card{played}, events)
		*events = append(*events, GameEvent{
			Type:        "play_trick",
			PlayerIndex: seat,
			TargetIndex: seat,
			Card:        &played,
			Message:     fmt.Sprintf("%s 使用【%s】", g.Players[seat].Name, played.Name),
		})
		return g.playTieSuoRecast(seat, played, events)
	}
	effectTarget := targetSpec.SeatIndex
	if effectTarget < 0 || effectTarget >= len(g.Players) {
		effectTarget = target
	}
	if g.targetBlockedByTrick(effectTarget, cardObj) {
		return ErrInvalidTarget
	}
	if cardObj.Kind == CardJueDou && g.runSkillHooks(nil, skill.HookCall{Kind: skill.HookTargetBlocked, Target: effectTarget, CardKind: CardJueDou}).Bool {
		return ErrInvalidTarget
	}

	played := g.removeHandCard(seat, idx, events)
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
	case CardGuoHe, CardTanNang, CardJueDou, CardTaoYuan, CardWuZhong:
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
		return g.startAoeTrick(seat, played, CardSha, events)
	case CardWanJian:
		return g.startAoeTrick(seat, played, CardShan, events)
	case CardHuoGong:
		return g.playHuoGong(seat, played, target, events)
	case CardTieSuo:
		g.notifyBecameTarget(target, seat, played, events)
		return g.resolveTieSuoChain(seat, target, played, events)
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
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  responder,
		ReturnIndex:  source,
		EffectTarget: effectTarget,
		Card:         trick,
		ResponseMode: ResponseModeWuxiekTrick,
		TargetZone:   spec.Zone,
		TargetCardID: spec.CardID,
	}
	g.Message = fmt.Sprintf("%s 可使用【无懈可击】抵消【%s】", g.Players[responder].Name, trick.Name)
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

func (g *Game) startWuxiekLebuJudgeWindow(seat int, events *[]GameEvent) {
	jc := g.Players[seat].judgeCardByKind(CardLeBu)
	if jc == nil {
		return
	}
	lebu := *jc
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  g.opponentOf(seat),
		TargetIndex:  seat,
		ReturnIndex:  seat,
		EffectTarget: seat,
		Card:         lebu,
		ResponseMode: ResponseModeWuxiekLebu,
	}
	g.Message = fmt.Sprintf("%s 可对【乐不思蜀】使用【无懈可击】（判定前）", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: g.opponentOf(seat),
		TargetIndex: seat,
		Card:        &lebu,
		Message:     g.Message,
	})
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
		if err == nil {
			g.notifyInstantTrickUsed(source, CardGuoHe, events)
		}
	case CardTanNang:
		err = g.resolveTanNang(source, target, spec, events)
		if err == nil {
			g.notifyInstantTrickUsed(source, CardTanNang, events)
		}
	case CardJueDou:
		err = g.startCardResponse(source, target, pending.Card, CardSha, fmt.Sprintf("%s 对 %s 发起【决斗】，%s 需出杀", g.Players[source].Name, g.Players[target].Name, g.Players[target].Name), events)
		if err == nil {
			g.notifyInstantTrickUsed(source, CardJueDou, events)
		}
	case CardTaoYuan:
		g.resolveTaoYuan(source, events)
		g.notifyInstantTrickUsed(source, CardTaoYuan, events)
	case CardWuZhong:
		g.Message = fmt.Sprintf("%s 使用【无中生有】，摸两张牌", g.Players[source].Name)
		*events = append(*events, GameEvent{Type: "trick_effect", PlayerIndex: source, TargetIndex: source, Message: g.Message})
		g.drawCards(source, 2, events)
		g.notifyInstantTrickUsed(source, CardWuZhong, events)
	default:
		return ErrInvalidCard
	}
	if err != nil {
		return err
	}
	g.resetTimer()
	return nil
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

func (g *Game) cancelTrickWithWuxiek(pending PendingCombat, events *[]GameEvent) error {
	g.Pending = nil
	seat := pending.TargetIndex

	switch pending.ResponseMode {
	case ResponseModeWuxiekLebu:
		p := &g.Players[seat]
		p.SkipPlay = false
		g.removeJudgeByKind(seat, CardLeBu)
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = seat
		g.Message = fmt.Sprintf("【乐不思蜀】被【无懈可击】抵消，%s 可正常出牌", p.Name)
	case ResponseModeWuxiekBingliang:
		p := &g.Players[seat]
		p.SkipDraw = false
		g.removeJudgeByKind(seat, CardBingLiang)
		g.Phase = PhasePlaying
		g.TurnStep = StepDraw
		g.CurrentTurn = seat
		g.Message = fmt.Sprintf("【兵粮寸断】被【无懈可击】抵消，%s 正常摸牌", p.Name)
		g.drawCards(seat, g.drawCountFor(seat), events)
		if g.IsFinished() {
			return nil
		}
		if p.SkipPlay {
			if p.hasJudgeKind(CardLeBu) {
				g.startWuxiekLebuJudgeWindow(seat, events)
				return nil
			}
			g.applyLebuSkipDirect(seat, events)
			return nil
		}
		g.TurnStep = StepPlay
	case ResponseModeWuxiekShandian:
		g.removeJudgeByKind(seat, CardShanDian)
		g.Phase = PhasePlaying
		g.TurnStep = StepDraw
		g.CurrentTurn = seat
		g.Message = fmt.Sprintf("【闪电】被【无懈可击】抵消")
		*events = append(*events, GameEvent{
			Type:        "trick_cancelled",
			PlayerIndex: pending.SourceIndex,
			TargetIndex: pending.TargetIndex,
			Card:        &pending.Card,
			Message:     g.Message,
		})
		return g.resumeBeginTurnAfterLightning(seat, events)
	default:
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = pending.SourceIndex
		g.Message = fmt.Sprintf("【%s】被【无懈可击】抵消", pending.Card.Name)
	}
	*events = append(*events, GameEvent{
		Type:        "trick_cancelled",
		PlayerIndex: pending.SourceIndex,
		TargetIndex: pending.TargetIndex,
		Card:        &pending.Card,
		Message:     g.Message,
	})
	if pending.Card.Kind == CardJueDou {
		g.tryJiangDraw(pending.SourceIndex, pending.Card, events)
	}
	g.resetTimer()
	return nil
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
	case CardWeapon1, CardWeapon2, CardWeapon3, CardWeapon4, CardWeapon5, CardWeapon6:
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
	card, label, ok := g.takeTargetCard(target, spec, events)
	if !ok {
		return ErrInvalidTarget
	}
	g.DiscardPile = append(g.DiscardPile, card)
	g.Message = fmt.Sprintf("%s 拆掉 %s 的%s", g.Players[seat].Name, g.Players[target].Name, label)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &card,
		Message:     g.Message,
	})
	return nil
}

func (g *Game) resolveTanNang(seat, target int, spec PlayTarget, events *[]GameEvent) error {
	card, label, ok := g.takeTargetCard(target, spec, events)
	if !ok {
		return ErrInvalidTarget
	}
	g.Players[seat].Hand = append(g.Players[seat].Hand, card)
	g.syncCounts()
	g.Message = fmt.Sprintf("%s 获得 %s 的%s", g.Players[seat].Name, g.Players[target].Name, label)
	*events = append(*events, GameEvent{
		Type:        "trick_effect",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &card,
		Message:     g.Message,
		Amount:      1,
	})
	return nil
}

func (g *Game) resolveTaoYuan(seat int, events *[]GameEvent) {
	g.Message = fmt.Sprintf("%s 使用【桃园结义】", g.Players[seat].Name)
	for i := range g.Players {
		if g.Players[i].HP >= g.Players[i].MaxHP {
			continue
		}
		g.Players[i].HP++
		*events = append(*events, GameEvent{
			Type:        "trick_heal",
			PlayerIndex: seat,
			TargetIndex: i,
			Heal:        1,
			Message:     fmt.Sprintf("%s 回复 1 点体力", g.Players[i].Name),
		})
	}
}

func (g *Game) startCardResponse(seat, target int, card Card, requiredKind string, message string, events *[]GameEvent) error {
	g.Phase = PhaseResponse
	allowWuxiek := card.Kind == CardNanMan || card.Kind == CardWanJian
	g.appendWushuangMessage(seat, card.Kind, &message)
	g.Pending = &PendingCombat{
		SourceIndex:     seat,
		TargetIndex:     target,
		ReturnIndex:     seat,
		Card:            card,
		RequiredKind:    requiredKind,
		Damage:          1,
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
		g.notifyInstantTrickUsed(source, card.Kind, events)
		return nil
	}
	rest := append([]int(nil), queue[1:]...)
	if err := g.startAoeResponse(source, queue[0], card, requiredKind, rest, events); err != nil {
		return err
	}
	g.notifyInstantTrickUsed(source, card.Kind, events)
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
