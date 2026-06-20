package engine

import (
	"time"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const maxAIActionsPerBurst = 500

// AIDelay AI每次行动之间的延迟，让前端有时间渲染动画
const AIDelay = 2 * time.Second

func RunAIActions(g *Game, events *[]GameEvent) {
	for i := 0; !g.IsFinished() && RunAIActionStep(g, events); i++ {
		if i >= maxAIActionsPerBurst {
			break
		}
		time.Sleep(AIDelay)
	}
}

// RunAIActionStep 执行一步 AI 决策；返回 false 表示本步无 AI 可行动作。
func RunAIActionStep(g *Game, events *[]GameEvent) bool {
	if g.IsFinished() {
		return false
	}
	LogGameState(g, "RunAIActionStep ENTER")
	acted := false

	if g.Phase == PhaseResponse && g.Pending != nil {
		pending := g.Pending
		if pending.TieqiPending {
			seat := pending.SourceIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if th, ok := skill.Lookup(SkillTieqi); ok && th.CanActivate(rt, seat) {
				_ = th.Activate(rt, seat, UseSkillRequest{SkillID: SkillTieqi})
			} else {
				_ = g.SkipTieqi(seat, events)
			}
			return true
		}
		if pending.WindowKind == WindowKindTake && g.takeWindow != nil {
			return g.autoTakeWindowIfNeeded(events)
		}
		if pending.WindowKind == WindowKindDiscard && g.discardWindow != nil {
			return g.autoDiscardWindowIfNeeded(events)
		}
		if pending.ResponseMode == ResponseModeSkillGuicai {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if gh, ok := skill.Lookup(SkillGuicai); ok && gh.CanActivate(rt, seat) {
				_ = gh.AIActivate(rt, seat)
			} else {
				_ = g.PassGuicai(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeSkillGuidao {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if gh, ok := skill.Lookup(SkillGuidao); ok && gh.CanActivate(rt, seat) {
				_ = gh.AIActivate(rt, seat)
			} else {
				_ = g.PassGuidao(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeSkillLeijiOffer {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if lh, ok := skill.Lookup(SkillLeiji); ok && lh.CanActivate(rt, seat) {
				_ = lh.AIActivate(rt, seat)
			} else {
				_ = g.PassLeijiOffer(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeSkillJianxiong {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if jh, ok := skill.Lookup(SkillJianxiong); ok && jh.CanActivate(rt, seat) {
				_ = jh.AIActivate(rt, seat)
			} else {
				_ = g.PassJianxiong(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeSkillYijiOffer || pending.ResponseMode == ResponseModeSkillYijiGive {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if yh, ok := skill.Lookup(SkillYiji); ok && yh.CanActivate(rt, seat) {
				if err := yh.AIActivate(rt, seat); err != nil {
					if pending.ResponseMode == ResponseModeSkillYijiOffer {
						_ = g.PassYijiOffer(seat, events)
					} else {
						_ = g.PassYijiGive(seat, events)
					}
				}
			} else if pending.ResponseMode == ResponseModeSkillYijiOffer {
				_ = g.PassYijiOffer(seat, events)
			} else {
				_ = g.PassYijiGive(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeSkillGanglieOffer {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if gh, ok := skill.Lookup(SkillGanglie); ok && gh.CanActivate(rt, seat) {
				if err := gh.AIActivate(rt, seat); err != nil {
					_ = g.PassGanglieOffer(seat, events)
				}
			} else {
				_ = g.PassGanglieOffer(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeSkillGanglieChoice {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if gh, ok := skill.Lookup(SkillGanglie); ok && gh.CanActivate(rt, seat) {
				_ = gh.AIActivate(rt, seat)
			} else {
				_ = g.GanglieTakeDamage(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeSkillFanjianSuit {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			_ = g.ResolveFanjianSuit(seat, g.aiPickFanjianSuit(), events)
			return true
		}
		if pending.ResponseMode == ResponseModeSkillTianxiang {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			rt := g.skillRuntime(events)
			if th, ok := skill.Lookup(SkillTianxiang); ok && th.CanActivate(rt, seat) {
				if err := th.AIActivate(rt, seat); err != nil {
					_ = g.PassTianxiang(seat, events)
				}
			} else {
				_ = g.PassTianxiang(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeSkillYinghun {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			_ = g.ResolveYinghunChoice(seat, g.aiPickYinghunOption(seat, pending.SourceIndex), events)
			return true
		}
		if pending.ResponseMode == ResponseModeSkillYinghunDiscard {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			ids := g.Players[seat].Hand
			if len(ids) == 0 {
				return false
			}
			// AI 逐张弃牌，每次弃最后一张
			_ = g.YinghunDiscard(seat, []string{ids[len(ids)-1].ID}, events)
			return true
		}
		if pending.ResponseMode == ResponseModeSkillLiuli {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			if g.runAIActiveSkills(seat, events) {
				return true
			}
			_ = g.PassLiuli(seat, events)
			return true
		}
		if pending.ResponseMode == ResponseModeWeapon9 {
			seat := pending.SourceIndex
			if !g.Players[seat].IsAI {
				return false
			}
			target := pending.EffectTarget
			hand := g.Players[seat].Hand
			if len(hand) < 2 {
				_ = g.passGuanshifu(seat, events)
				return true
			}
			shouldDiscard := g.Players[target].HP <= 2 || g.Players[target].HP >= g.Players[target].MaxHP
			if shouldDiscard {
				cardIDs := []string{hand[0].ID, hand[1].ID}
				_ = g.RespondDiscardCards(seat, cardIDs, events)
			} else {
				_ = g.passGuanshifu(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeWeapon8 {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			// AI 目标：若有手牌，优先弃一张牌（避免让对方摸牌）
			hand := g.Players[seat].Hand
			if len(hand) > 0 {
				_ = g.RespondCard(seat, hand[0].ID, events)
			} else {
				_ = g.PassResponse(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeDying {
			seat := pending.SourceIndex
			victim := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			if shouldAIDyingRescue(g, seat, victim) {
				if idx := firstPlaysAsCard(g, seat, CardTao); idx >= 0 {
					_ = g.RespondCard(seat, g.Players[seat].Hand[idx].ID, events)
					return true
				}
			}
			_ = g.PassResponse(seat, events)
			return true
		}
		// 用 ActorSeat 获取当前响应者（支持反无懈窗口 TargetIndex=-1 的情况）
		seat := g.PendingActorSeat()
		if seat < 0 || seat >= len(g.Players) {
			return false
		}
		if !g.Players[seat].IsAI {
			return false
		}
		if pending.ResponseMode == ResponseModeGuanYuFollow {
			if idx := firstShaLikeCard(g, seat); idx >= 0 {
				_ = g.playSha(seat, g.Players[seat].Hand[idx].ID, pending.EffectTarget, events)
			} else {
				_ = g.finishGuanYuFollowUp(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeQilinBow {
			target := pending.EffectTarget
			tp := &g.Players[target]
			zone := ""
			if tp.MinusHorse != nil {
				zone = EquipMinusHorse
			} else if tp.PlusHorse != nil {
				zone = EquipPlusHorse
			}
			if zone != "" {
				_ = g.qilinDiscardHorse(seat, zone, events)
			} else {
				_ = g.finishQilinBow(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeWuxiekTrick || pending.ResponseMode == ResponseModeWuxiekLebu ||
			pending.ResponseMode == ResponseModeWuxiekBingliang || pending.ResponseMode == ResponseModeWuxiekShandian {
			// 推进到下一个响应者（一次只处理一个AI）
			g.advanceToNextWuxiekResponder(events)
			// 如果 Pending 还在且还是无懈模式，说明还有AI需要处理
			return g.Pending != nil && (g.Pending.ResponseMode == ResponseModeWuxiekTrick ||
				g.Pending.ResponseMode == ResponseModeWuxiekLebu ||
				g.Pending.ResponseMode == ResponseModeWuxiekBingliang ||
				g.Pending.ResponseMode == ResponseModeWuxiekShandian)
		}
		if pending.ResponseMode == ResponseModeDdzJudgeCancel {
			if len(g.Players[seat].Hand) >= 2 {
				_ = g.ApplyDdzJudgeCancel(seat, []string{
					g.Players[seat].Hand[0].ID,
					g.Players[seat].Hand[1].ID,
				}, events)
			} else {
				_ = g.PassDdzJudgeCancel(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModePeekDeck {
			seat := pending.TargetIndex
			if !g.Players[seat].IsAI {
				return false
			}
			_ = g.finishPeekDeckAsAI(seat, events)
			return true
		}
		if pending.ResponseMode == ResponseModeWuguPick {
			// 选牌者是AI，自动选牌（无懈可击已在前面处理）
			picker := pending.WuguPickSeat
			if g.Players[picker].IsAI {
				if len(pending.RevealedCards) > 0 {
					if err := g.autoPickWuguCard(picker, events); err == nil {
						return true
					}
				}
				_ = g.autoPickWuguCard(picker, events)
				return true
			}
			return false
		}
		if pending.ResponseMode == ResponseModeSkillJijiang {
			ally := pending.TargetIndex
			lord := pending.JijiangLord
			if !g.Players[ally].IsAI {
				return false
			}
			if shouldAIRespondJijiang(g, ally, lord) {
				if idx := firstShaLikeCard(g, ally); idx >= 0 {
					_ = g.respondJijiangSha(ally, g.Players[ally].Hand[idx].ID, events)
					return true
				}
			}
			_ = g.passJijiang(ally, events)
			return true
		}
		if pending.ResponseMode == ResponseModeSkillLuanwu {
			seat := pending.TargetIndex
			target := pending.EffectTarget
			if target < 0 {
				target = seat
			}
			if idx := firstShaLikeCard(g, seat); idx >= 0 {
				_ = g.playLuanwuSha(seat, g.Players[seat].Hand[idx].ID, target, events)
			} else {
				_ = g.passLuanwu(seat, events)
			}
			return true
		}
		if pending.ResponseMode == ResponseModeHuoGong {
			seat := pending.TargetIndex
			if len(pending.RevealedCards) > 0 {
				suit := pending.RevealedCards[0].Suit
				for _, c := range g.Players[seat].Hand {
					if c.Suit == suit {
						_ = g.respondHuoGongDiscard(seat, c.ID, events)
						return true
					}
				}
			}
			_ = g.resolveHuoGongFail(seat, events)
			return true
		}
		required := pending.RequiredKind
		if required == "" {
			required = CardShan
		}
		if required == CardShan && g.hasBaguaArmor(seat) && !pending.BaguaUsed && !pending.IgnoreArmor {
			if err := g.TryBaguaJudge(seat, events); err == nil {
				return true
			}
			if g.Pending != nil {
				g.Pending.BaguaUsed = true
			}
		}
		// 尝试从手牌中找到可变牌的牌
		if idx := firstShaLikeCard(g, seat); idx >= 0 && required == CardSha {
			cardID := g.Players[seat].Hand[idx].ID
			_ = g.RespondCard(seat, cardID, events)
			acted = true
		} else if idx := firstPlaysAsCard(g, seat, required); idx >= 0 {
			cardID := g.Players[seat].Hand[idx].ID
			_ = g.RespondCard(seat, cardID, events)
			acted = true
		} else if _, equipCard, ok := findPlayableEquipCard(g, seat, required); ok {
			// 尝试从装备区变牌（武圣红色装备当杀、倾国黑色装备当闪等）
			_ = g.PlayCardWithTarget(seat, equipCard.ID, PlayTarget{SeatIndex: pending.SourceIndex}, events)
			acted = true
		} else if pending.AllowWuxiek && (firstCardKind(g.Players[seat].Hand, CardWuxiek) >= 0 || g.cardPlaysAs(seat, firstEquipCardFor(g, seat, CardWuxiek), CardWuxiek)) {
			if idx := firstCardKind(g.Players[seat].Hand, CardWuxiek); idx >= 0 {
				_ = g.RespondWuxiek(seat, g.Players[seat].Hand[idx].ID, events)
			} else {
				_ = g.RespondWuxiek(seat, firstEquipCardFor(g, seat, CardWuxiek).ID, events)
			}
			acted = true
		} else {
			_ = g.PassResponse(seat, events)
			acted = true
		}
	} else if g.Phase == PhasePlaying && g.TurnStep == StepPrepare && g.Players[g.CurrentTurn].IsAI {
		seat := g.CurrentTurn
		g.runAIPreparePhase(seat, events)
		acted = true
	} else if g.Phase == PhasePlaying && g.TurnStep == StepDraw && g.Players[g.CurrentTurn].IsAI {
		seat := g.CurrentTurn
		acted = runAIDrawPhase(g, seat, events)
	} else if g.Phase == PhasePlaying && g.TurnStep == StepPlay && g.Players[g.CurrentTurn].IsAI {
		if g.tryAIJijiHeal(events) {
			return true
		}
		seat := g.CurrentTurn
		acted = runAIPlayPhase(g, seat, events)
	} else if g.Phase == PhasePlaying && g.TurnStep == StepDiscard && g.Players[g.CurrentTurn].IsAI {
		seat := g.CurrentTurn
		g.autoDiscard(seat, events)
		_ = g.endTurn(events)
		acted = true
	}

	return acted
}

func runAIDrawPhase(g *Game, seat int, events *[]GameEvent) bool {
	if !g.isDrawPhaseChoicePending(seat) {
		return false
	}
	rt := g.skillRuntime(events)
	bestID := ""
	bestPri := 0
	for _, sid := range []string{SkillTuxi, SkillLuoyi, SkillShuangxiong} {
		h, ok := skill.Lookup(sid)
		if !ok || !h.CanActivate(rt, seat) {
			continue
		}
		if p := h.AIPriority(rt, seat); p > bestPri {
			bestPri = p
			bestID = sid
		}
	}
	if bestID != "" {
		if h, ok := skill.Lookup(bestID); ok {
			_ = h.AIActivate(rt, seat)
			return true
		}
	}
	_ = g.PassDrawPhase(seat, events)
	return true
}

func (g *Game) pickTrickTarget(seat int, cardKind string) int {
	valid := g.validPlayTargets(seat, cardKind)
	if len(valid) > 0 {
		return valid[0]
	}
	return g.opponentOf(seat)
}

func runAIPlayPhase(g *Game, seat int, events *[]GameEvent) bool {
	p := &g.Players[seat]
	target := g.pickAITarget(seat)

	if g.runAIActiveSkills(seat, events) {
		return true
	}

	for _, kind := range []string{CardWeapon9, CardWeapon8, CardWeapon7, CardWeapon6, CardWeapon5, CardWeapon4, CardWeapon3, CardWeapon2, CardWeapon1, CardArmorVine, CardArmor, CardPlusHorse, CardMinusHorse} {
		idx := firstCardKind(p.Hand, kind)
		if idx < 0 || !shouldAIEquip(p, kind) {
			continue
		}
		_ = g.playEquip(seat, p.Hand[idx].ID, events)
		return true
	}

	for _, kind := range []string{CardWuZhong, CardTaoYuan, CardWuGu, CardShanDian, CardGuoHe, CardTanNang, CardNanMan, CardWanJian, CardJueDou, CardLeBu, CardBingLiang, CardHuoGong, CardTieSuo} {
		idx := firstCardKind(p.Hand, kind)
		if idx < 0 {
			continue
		}
		if trickNeedsOpponentTarget(kind) {
			if len(g.validPlayTargets(seat, kind)) == 0 {
				continue
			}
			target = g.pickTrickTarget(seat, kind)
		}
		if kind == CardTaoYuan && !anyWounded(g) {
			continue
		}
		if kind == CardShanDian && p.hasJudgeKind(CardShanDian) {
			continue
		}
		playTarget := PlayTarget{SeatIndex: target, Zone: "hand"}
		if kind == CardShanDian || kind == CardWuGu || kind == CardTieSuo {
			playTarget = PlayTarget{SeatIndex: seat}
		}
		if err := g.playTrick(seat, p.Hand[idx].ID, playTarget, events); err == nil {
			return true
		}
	}

	if !g.canUseSha(seat) && !p.Drunk && len(g.validPlayTargets(seat, CardSha)) > 0 {
		hasSha := firstShaLikeCard(g, seat) >= 0
		if idx := firstCardKind(p.Hand, CardJiu); idx >= 0 && hasSha {
			_ = g.playJiu(seat, p.Hand[idx].ID, events)
			return true
		}
	}

	if g.canUseSha(seat) {
		if idx := firstShaLikeCard(g, seat); idx >= 0 {
			shaTarget := g.pickAITarget(seat)
			if len(g.validPlayTargets(seat, CardSha)) > 0 {
				_ = g.playSha(seat, p.Hand[idx].ID, shaTarget, events)
				return true
			}
		}
	}

	if p.HP < p.MaxHP {
		if idx := firstCardKind(p.Hand, CardTao); idx >= 0 {
			_ = g.playTao(seat, p.Hand[idx].ID, events)
			return true
		}
	}

	if err := g.EndPlay(seat, events); err == nil {
		return true
	}
	return false
}

func shouldAIEquip(p *Player, kind string) bool {
	switch equipSlot(kind) {
	case EquipWeapon:
		if kind == CardWeapon1 {
			return p.Weapon == nil || p.Weapon.Kind != CardWeapon1
		}
		return p.Weapon == nil || weaponRange(kind) > weaponRange(p.Weapon.Kind)
	case EquipArmor:
		return p.Armor == nil
	case EquipPlusHorse:
		return p.PlusHorse == nil
	case EquipMinusHorse:
		return p.MinusHorse == nil
	default:
		return false
	}
}

func anyWounded(g *Game) bool {
	for i := range g.Players {
		if g.Players[i].HP < g.Players[i].MaxHP {
			return true
		}
	}
	return false
}

func firstCardKind(hand []Card, kind string) int {
	for i, c := range hand {
		if c.Kind == kind {
			return i
		}
	}
	return -1
}

func firstShaLikeCard(g *Game, seat int) int {
	return firstPlaysAsCard(g, seat, CardSha)
}

func firstPlaysAsCard(g *Game, seat int, asKind string) int {
	// 先查手牌
	for i, c := range g.Players[seat].Hand {
		if g.cardPlaysAs(seat, c, asKind) {
			return i
		}
	}
	return -1
}

// firstShaLikeCardWithEquip 查找可变牌的牌（手牌或装备区），返回 (zone, idx, card) 或 ("", -1, Card{})
func findPlayableEquipCard(g *Game, seat int, asKind string) (zone string, card Card, ok bool) {
	p := &g.Players[seat]
	if p.Weapon != nil && g.cardPlaysAs(seat, *p.Weapon, asKind) {
		return EquipWeapon, *p.Weapon, true
	}
	if p.Armor != nil && g.cardPlaysAs(seat, *p.Armor, asKind) {
		return EquipArmor, *p.Armor, true
	}
	if p.PlusHorse != nil && g.cardPlaysAs(seat, *p.PlusHorse, asKind) {
		return EquipPlusHorse, *p.PlusHorse, true
	}
	if p.MinusHorse != nil && g.cardPlaysAs(seat, *p.MinusHorse, asKind) {
		return EquipMinusHorse, *p.MinusHorse, true
	}
	return "", Card{}, false
}

// firstEquipCardFor 从装备区找第一张可当 asKind 使用的牌
func firstEquipCardFor(g *Game, seat int, asKind string) Card {
	_, card, _ := findPlayableEquipCard(g, seat, asKind)
	return card
}

func shouldAIWuxiekTrick(g *Game, seat int, pending *PendingCombat) bool {
	switch pending.Card.Kind {
	case CardGuoHe, CardTanNang, CardJueDou:
		return true
	case CardTaoYuan:
		return g.Players[seat].HP < g.Players[seat].MaxHP
	case CardWuZhong:
		return true
	default:
		return pending.ResponseMode == ResponseModeWuxiekLebu ||
			pending.ResponseMode == ResponseModeWuxiekBingliang ||
			pending.ResponseMode == ResponseModeWuxiekShandian
	}
}

func shouldAIDyingRescue(g *Game, rescuer, victim int) bool {
	if rescuer == victim {
		return true
	}
	if g.is2v2() || g.is3v3() || g.isIdentity() || g.is3pChain() || g.is3pDdz() {
		return g.isAlly(rescuer, victim)
	}
	return false
}

func shouldAIRespondJijiang(g *Game, ally, lord int) bool {
	if !lordSkillsActive(g.Mode) {
		return false
	}
	return g.isAlly(ally, lord)
}
