package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) RespondShan(seat int, cardID string, events *[]GameEvent) error {
	return g.RespondCard(seat, cardID, events)
}

func (g *Game) RespondCard(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	if g.Pending.ResponseMode == ResponseModeDying {
		return g.playTaoForDying(seat, cardID, events)
	}
	if !g.CanRespondSeat(seat) {
		return ErrNotYourTurn
	}

	// 支持手牌和装备区变牌
	fallbackKind := g.Pending.RequiredKind
	if fallbackKind == "" && g.Pending.ResponseMode == ResponseModeWuxiekTrick {
		fallbackKind = CardWuxiek
	}
	Logf("RespondCard: seat=%d cardID=%s fallbackKind=%s pendingMode=%s", seat, cardID, fallbackKind, g.Pending.ResponseMode)
	zone, handIdx, cardObj, ok := g.findCardInHandOrEquipKind(seat, cardID, fallbackKind)
	if !ok {
		Logf("RespondCard: card not found seat=%d cardID=%s", seat, cardID)
		return ErrInvalidCard
	}
	if cardObj.Kind == CardWuxiek {
		return g.respondWuxiekWithCard(seat, cardObj, handIdx, zone, events)
	}

	pending := g.Pending
	if pending.ResponseMode == ResponseModeWuxiekTrick || pending.ResponseMode == ResponseModeWuxiekLebu ||
		pending.ResponseMode == ResponseModeWuxiekBingliang || pending.ResponseMode == ResponseModeWuxiekShandian {
		return ErrInvalidCard
	}
	if pending.ShaUnblockable {
		return ErrInvalidCard
	}

	requiredKind := pending.RequiredKind
	if requiredKind == "" {
		requiredKind = CardShan
	}
	if cardObj.Kind != requiredKind && !g.cardPlaysAs(seat, cardObj, requiredKind) {
		return ErrInvalidCard
	}

	// 检查是否因龙胆而将杀当闪打出，如果是则触发冲阵
	g.triggerChongzhen(seat, cardObj, requiredKind)

	// 从对应区域移除牌
	var played Card
	if zone == string(ZoneHand) || zone == "" {
		played = g.removeHandCard(seat, handIdx, events)
	} else {
		played = g.removeEquipCard(seat, zone, events)
		g.notifyEquipLost(seat, played, "skill", events)
	}
	// 变牌：如果牌本身不是目标类型，统一转为目标类型
	if played.Kind != requiredKind && !isSha(played.Kind) {
		played = g.convertCardToKind(played, requiredKind)
	}
	g.DiscardPile = append(g.DiscardPile, played)
	g.runCardsDiscardedHooks(seat, "play", []Card{played}, events)
	source := pending.SourceIndex
	responseType := "respond_" + requiredKind

	*events = append(*events, GameEvent{
		Type:        responseType,
		PlayerIndex: seat,
		TargetIndex: source,
		Card:        &played,
		Message:     fmt.Sprintf("%s 打出【%s】", g.Players[seat].Name, played.Name),
	})
	if requiredKind == CardSha {
		g.markShaInPlayPhase(seat)
	}

	if g.consumeWushuangResponse(pending, seat, requiredKind) {
		return nil
	}

	if pending.Card.Kind == CardJueDou {
		g.Pending = &PendingCombat{
			SourceIndex:  seat,
			TargetIndex:  source,
			ReturnIndex:  pending.ReturnIndex,
			Card:         pending.Card,
			RequiredKind: CardSha,
			Damage:       pending.Damage,
		}
		g.Message = fmt.Sprintf("【决斗】继续，%s 需出杀", g.Players[source].Name)
		g.resetTimer()
		return nil
	}

	return g.resolvePendingDodgeSuccess(seat, pending, events, "")
}

func (g *Game) resolvePendingDodgeSuccess(seat int, pending *PendingCombat, events *[]GameEvent, messageOverride string) error {
	if pending.Card.Kind == CardNanMan || pending.Card.Kind == CardWanJian {
		required := pending.RequiredKind
		if required == "" {
			required = CardShan
		}
		queue := pending.AoeQueue
		g.Pending = nil
		if g.offerLeijiAfterShan(seat, pending, events) {
			return nil
		}
		// 南蛮/万箭用新的逐人流程
		if pending.Card.Kind == CardNanMan {
			g.continueNanManAfterTarget(pending.SourceIndex, queue, events)
			return nil
		}
		if pending.Card.Kind == CardWanJian {
			g.continueWanJianAfterTarget(pending.SourceIndex, queue, events)
			return nil
		}
		return g.continueAoeAfterTarget(pending.SourceIndex, pending.Card, required, queue, events)
	}
	// 贯石斧：杀被闪抵消后， attacker 可弃两张牌令杀命中
	if pending.Card.Kind == CardSha {
		source := pending.SourceIndex
		if g.offerGuanshifu(source, seat, pending.Card, pending.Damage, pending.ReturnIndex, events) {
			return nil
		}
	}
	if g.offerLeijiAfterShan(seat, pending, events) {
		return nil
	}
	return g.finishShanDodgeSuccess(seat, pending, events, messageOverride)
}

func isRedSuit(suit string) bool {
	return skill.IsRedSuit(suit)
}

func (g *Game) TryBaguaJudge(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	pending := g.Pending
	if pending.ResponseMode == ResponseModeWuxiekTrick || pending.ResponseMode == ResponseModeWuxiekLebu ||
		pending.ResponseMode == ResponseModeWuxiekBingliang || pending.ResponseMode == ResponseModeWuxiekShandian {
		return ErrInvalidCard
	}
	requiredKind := pending.RequiredKind
	if requiredKind == "" {
		requiredKind = CardShan
	}
	if requiredKind != CardShan {
		return ErrInvalidCard
	}
	if g.Players[seat].Armor == nil || !g.hasBaguaArmor(seat) {
		return ErrInvalidCard
	}
	if pending.IgnoreArmor {
		return ErrInvalidCard
	}
	if pending.BaguaUsed {
		return ErrAlreadyActed
	}

	return g.startJudge(seat, skill.JudgeBagua, guicaiResumeBagua, events)
}

func (g *Game) RespondWuxiek(seat int, cardID string, events *[]GameEvent) error {
	// 容错：按 cardID 查找，找不到时用手牌中第一张无懈可击
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || cardObj.Kind != CardWuxiek {
		for i, c := range g.Players[seat].Hand {
			if c.Kind == CardWuxiek {
				idx = i
				cardObj = c
				ok = true
				break
			}
		}
		if !ok {
			return ErrInvalidCard
		}
	}
	return g.respondWuxiekWithCard(seat, cardObj, idx, string(ZoneHand), events)
}

// respondWuxiekWithCard 用已找到的牌打出无懈可击
func (g *Game) respondWuxiekWithCard(seat int, cardObj Card, handIdx int, zone string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	if !g.CanRespondWuxiek(seat) {
		return ErrNotYourTurn
	}

	pending := *g.Pending
	switch pending.ResponseMode {
	case ResponseModeWuxiekTrick, ResponseModeWuxiekLebu, ResponseModeWuxiekBingliang, ResponseModeWuxiekShandian, ResponseModeWuxiekGuose:
	case ResponseModeWeapon8:
		return g.resolveChixiongDiscard(seat, cardObj.ID, events)
	case "":
		if !pending.AllowWuxiek {
			return ErrInvalidCard
		}
	default:
		return ErrInvalidCard
	}

	var played Card
	if zone == string(ZoneHand) || zone == "" {
		played = g.removeHandCard(seat, handIdx, events)
	} else {
		played = g.removeEquipCard(seat, zone, events)
		g.notifyEquipLost(seat, played, "skill", events)
	}
	g.DiscardPile = append(g.DiscardPile, played)
	g.notifyInstantTrickUsed(seat, CardWuxiek, events)
	wuxiekTarget := pending.SourceIndex
	if len(g.Pending.WuxiekChain) > 0 {
		wuxiekTarget = g.Pending.WuxiekChain[len(g.Pending.WuxiekChain)-1].Seat
	}
	*events = append(*events, GameEvent{
		Type:        "play_wuxiek",
		PlayerIndex: seat,
		TargetIndex: wuxiekTarget,
		Card:        &played,
		Message:     fmt.Sprintf("%s 打出【无懈可击】", g.Players[seat].Name),
	})

	if pending.ResponseMode == ResponseModeWuxiekLebu ||
		pending.ResponseMode == ResponseModeWuxiekBingliang ||
		pending.ResponseMode == ResponseModeWuxiekShandian ||
		pending.ResponseMode == ResponseModeWuxiekGuose {
		return g.handleJudgeWuxiekResponse(seat, pending, events)
	}

	if pending.AllowWuxiek && (pending.ResponseMode == "" || pending.ResponseMode == ResponseModeWuguPick) {
		prevMode := pending.ResponseMode
		g.Pending.ResponseMode = ResponseModeWuxiekTrick
		g.Pending.TargetIndex = -1
		if prevMode == ResponseModeWuguPick {
			g.Pending.EffectTarget = pending.WuguPickSeat
			g.Pending.SavedPending = &pending
		} else {
			g.Pending.EffectTarget = pending.TargetIndex
		}
		g.Pending.WuxiekChain = append(g.Pending.WuxiekChain, WuxiekEntry{Seat: seat, Card: played})
		g.rebuildWuxiekQueue(seat)
		g.advanceToNextWuxiekResponder(events)
		return nil
	}

	g.Pending.WuxiekChain = append(g.Pending.WuxiekChain, WuxiekEntry{Seat: seat, Card: played})
	g.rebuildWuxiekQueue(seat)
	g.advanceToNextWuxiekResponder(events)
	return nil
}

// rebuildWuxiekQueue 重建无懈可击响应队列。
// 从刚出无懈者的下家开始，排除刚出无懈的人（他刚出完），但包含链中其他人。
func (g *Game) rebuildWuxiekQueue(fromSeat int) {
	newQueue := make([]int, 0, len(g.Players))
	for i := 0; i < len(g.Players); i++ {
		s := (fromSeat + i) % len(g.Players)
		// 只排除刚出无懈的人，链中其他人可以再出（如果手牌还有无懈可击）
		if s != fromSeat && g.Players[s].HP > 0 {
			newQueue = append(newQueue, s)
		}
	}
	g.Pending.ResponseQueue = newQueue
	g.Pending.ResponseIndex = 0
}

// handleJudgeWuxiekResponse 处理判定前无懈可击的响应
func (g *Game) handleJudgeWuxiekResponse(seat int, pending PendingCombat, events *[]GameEvent) error {
	// 无懈可击打出来了，需要启动一个反无懈可击的窗口
	// 保存当前的判定信息
	judgeSeat := pending.EffectTarget
	judgeCard := pending.Card
	
	// 创建反无懈可击的响应队列：从打出无懈可击的玩家下家开始，逆时针顺序
	responseQueue := g.createResponseQueue((seat + 1) % len(g.Players))
	
	// 启动反无懈可击窗口
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:    seat, // 刚刚打出无懈可击的人
		TargetIndex:    -1,    // 任何人都可以响应
		ReturnIndex:    judgeSeat,
		EffectTarget:   judgeSeat,
		Card:           judgeCard,
		ResponseMode:   pending.ResponseMode, // 保持相同的响应模式
		AllowWuxiek:    true,
		SavedPending:   pending.SavedPending, // 保存原始判定信息
		ResponseQueue:  responseQueue,
		ResponseIndex:  0, // 从队列第一个玩家开始
	}
	
	// 设置当前响应者
	if len(responseQueue) > 0 {
		g.Pending.ActorSeat = responseQueue[0]
	}
	
	g.Message = fmt.Sprintf("可对【无懈可击】使用【无懈可击】")
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_counter_offer",
		PlayerIndex: seat,
		TargetIndex:  -1,
		Card:        &judgeCard,
		Message:     g.Message,
	})
	
	return nil
}

// advanceWuxiekResponseQueue 推进反无懈可击的响应队列
func (g *Game) advanceWuxiekResponseQueue(events *[]GameEvent) error {
	if g.Pending == nil || len(g.Pending.ResponseQueue) == 0 {
		// 响应队列为空，第一张无懈可击生效
		return g.handleWuxiekCounterPass(events)
	}
	
	// 移动到下一个响应者
	g.Pending.ResponseIndex++
	
	// 如果所有玩家都响应过了
	if g.Pending.ResponseIndex >= len(g.Pending.ResponseQueue) {
		// 第一张无懈可击生效
		return g.handleWuxiekCounterPass(events)
	}
	
	// 设置下一个响应者
	nextSeat := g.Pending.ResponseQueue[g.Pending.ResponseIndex]
	g.Pending.ActorSeat = nextSeat
	
	// 如果下一个响应者是AI，自动处理
	if g.Players[nextSeat].IsAI {
		return g.handleAIWuxiekResponse(nextSeat, events)
	}
	
	return nil
}

// advanceJudgeWuxiekQueue 推进判定前无懈可击的响应队列
func (g *Game) advanceJudgeWuxiekQueue(seat int, events *[]GameEvent) error {
	if g.Pending == nil || len(g.Pending.ResponseQueue) == 0 {
		// 响应队列为空，执行判定
		return g.executeJudge(seat, g.Pending.Card, events)
	}
	
	// 移动到下一个响应者
	g.Pending.ResponseIndex++
	
	// 如果所有玩家都响应过了
	if g.Pending.ResponseIndex >= len(g.Pending.ResponseQueue) {
		// 没有人使用无懈可击，执行判定
		return g.executeJudge(seat, g.Pending.Card, events)
	}
	
	// 设置下一个响应者
	nextSeat := g.Pending.ResponseQueue[g.Pending.ResponseIndex]
	g.Pending.ActorSeat = nextSeat
	
	// 如果下一个响应者是AI，自动处理
	if g.Players[nextSeat].IsAI {
		return g.handleAIWuxiekResponse(nextSeat, events)
	}
	
	return nil
}

// handleAIWuxiekResponse 处理AI的无懈可击响应
func (g *Game) handleAIWuxiekResponse(seat int, events *[]GameEvent) error {
	// 检查AI是否有无懈可击
	hasWuxiek := false
	var wuxiekCardID string
	for _, card := range g.Players[seat].Hand {
		if card.Kind == CardWuxiek {
			hasWuxiek = true
			wuxiekCardID = card.ID
			break
		}
	}
	
	if hasWuxiek && g.shouldAIUseWuxiek(seat) {
		// AI决定使用无懈可击
		return g.RespondWuxiek(seat, wuxiekCardID, events)
	}
	// AI决定不使用无懈可击，继续队列
	return g.PassResponse(seat, events)
}

// shouldAIUseWuxiek AI是否应该使用无懈可击（简化版本）
func (g *Game) shouldAIUseWuxiek(seat int) bool {
	// 简化逻辑：随机决定是否使用
	// 实际应该根据游戏状态、策略等综合判断
	return false // 暂时返回false，避免AI过于频繁使用
}

func (g *Game) PassResponse(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	g.ensurePendingRoles()
	if g.Pending.WindowKind == WindowKindTake && g.takeWindow != nil {
		if g.IsActorSeat(seat) {
			return g.PassTake(seat, events)
		}
		g.abandonTakeWindow()
	}
	if g.Pending.WindowKind == WindowKindDiscard && g.discardWindow != nil {
		if g.IsActorSeat(seat) {
			return g.PassDiscardWindow(seat, events)
		}
	}
	if g.Pending.TieqiPending && seat == g.Pending.SourceIndex {
		return g.SkipTieqi(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillGuicai && seat == g.Pending.TargetIndex {
		return g.PassGuicai(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillGuidao && seat == g.Pending.TargetIndex {
		return g.PassGuidao(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillLeijiOffer && seat == g.Pending.TargetIndex {
		return g.PassLeijiOffer(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillFankui && seat == g.Pending.TargetIndex {
		return g.PassFankui(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillJianxiong && seat == g.Pending.TargetIndex {
		return g.PassJianxiong(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillYijiOffer && seat == g.Pending.TargetIndex {
		return g.PassYijiOffer(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillYijiGive && seat == g.Pending.TargetIndex {
		return g.PassYijiGive(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillGanglieOffer && seat == g.Pending.TargetIndex {
		return g.PassGanglieOffer(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillTuxi && seat == g.Pending.TargetIndex {
		return g.PassTuxi(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillFanjianSuit && seat == g.Pending.TargetIndex {
		return g.ResolveFanjianSuit(seat, g.aiPickFanjianSuit(), events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillTianxiang && seat == g.Pending.TargetIndex {
		return g.PassTianxiang(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillYinghun && seat == g.Pending.TargetIndex {
		return g.ResolveYinghunChoice(seat, g.aiPickYinghunOption(seat, g.Pending.SourceIndex), events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillYinghunDiscard && seat == g.Pending.TargetIndex {
		if len(g.Players[seat].Hand) == 0 {
			return ErrInvalidCard
		}
		return g.YinghunDiscard(seat, []string{g.Players[seat].Hand[0].ID}, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillLiuli && seat == g.Pending.TargetIndex {
		return g.PassLiuli(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeDdzJudgeCancel && seat == g.Pending.TargetIndex {
		return g.PassDdzJudgeCancel(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillLuanwu {
		return g.passLuanwu(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeHuoGong && seat == g.Pending.TargetIndex {
		return g.resolveHuoGongFail(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeDying {
		return g.passDying(seat, events)
	}
	if !g.CanRespondSeat(seat) {
		return ErrNotYourTurn
	}
	switch g.Pending.ResponseMode {
	case ResponseModeWuxiekTrick:
		// 只有当前 Actor 可以跳过（PassResponse）
		if seat != g.Pending.ActorSeat {
			return ErrNotYourTurn
		}
		g.advanceWuxiekQueueAfterPass(seat, events)
		return nil
	case ResponseModeWuxiekLebu, ResponseModeWuxiekBingliang, ResponseModeWuxiekShandian:
		// 检查是否是反无懈可击窗口
		if g.Pending.SavedPending != nil {
			// 这是反无懈可击窗口的跳过
			// 移动到响应队列的下一个玩家
			return g.advanceWuxiekResponseQueue(events)
		} else {
			// 这是判定前无懈可击窗口的跳过
			// 移动到响应队列的下一个玩家
			return g.advanceJudgeWuxiekQueue(seat, events)
		}
	case ResponseModeWuxiekGuose:
		// 国色无懈可击响应窗口的跳过
		// 检查是否是反无懈可击窗口
		if g.Pending.SavedPending != nil {
			// 这是反无懈可击窗口的跳过
			return g.advanceWuxiekResponseQueue(events)
		} else {
			// 这是国色无懈可击窗口的跳过，响应完成，乐不思蜀成功置入
			// 返回出牌阶段
			g.Pending = nil
			g.Phase = PhasePlaying
			g.TurnStep = StepPlay
			g.resetTimer()
			return nil
		}
	case ResponseModeGuanYuFollow:
		return g.finishGuanYuFollowUp(seat, events)
	case ResponseModeQilinBow:
		return g.finishQilinBow(seat, events)
	case ResponseModeWeapon8:
		return g.passChixiong(seat, events)
	case ResponseModeWeapon9:
		return g.passGuanshifu(seat, events)
	case ResponseModeSkillJijiang:
		return g.passJijiang(seat, events)
	case ResponseModeWuguPick:
		// 五谷丰登：只有当前选牌者可以操作（选牌），其他人不能做任何事
		return ErrNotYourTurn
	case ResponseModePeekDeck:
		return ErrWrongPhase
	default:
		return g.resolvePendingMiss(events)
	}
}

// RespondDiscardCards 处理需要弃置多张牌的响应窗口（如贯石斧）。
func (g *Game) RespondDiscardCards(seat int, cardIDs []string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	if !g.CanRespondSeat(seat) && seat != g.Pending.SourceIndex {
		return ErrNotYourTurn
	}
	switch g.Pending.ResponseMode {
	case ResponseModeWeapon9:
		return g.resolveGuanshifuDiscard(seat, cardIDs, events)
	default:
		return ErrInvalidCard
	}
}

func (g *Game) resolvePendingMiss(events *[]GameEvent) error {
	pending := *g.Pending
	if len(pending.AoeQueue) >= 0 && (pending.Card.Kind == CardNanMan || pending.Card.Kind == CardWanJian) {
		required := pending.RequiredKind
		if required == "" {
			required = CardShan
		}
		damage := pending.Damage
		if damage <= 0 {
			damage = 1
		}
		g.applyDamageWithHook(pending.SourceIndex, pending.TargetIndex, damage, pending.Card, events)
		*events = append(*events, GameEvent{
			Type:        "trick_hit",
			PlayerIndex: pending.SourceIndex,
			TargetIndex: pending.TargetIndex,
			Damage:      damage,
			Message:     g.damageMessage(&g.Players[pending.TargetIndex], pending.Card.Name, damage),
		})
		if g.Players[pending.TargetIndex].HP <= 0 {
			if g.afterDamageApplied(pending.SourceIndex, pending.TargetIndex, damage, pending.Card, DamageResume{}, events) {
				// 濒死阶段启动，startDyingWindow 已保存当前 Pending 到 SavedPending
				// 不要清空 Pending，等濒死结束恢复
				return nil
			}
			if g.IsFinished() {
				g.Pending = nil
				return nil
			}
		}
		g.Pending = nil
		// 南蛮/万箭用新的逐人流程
		if pending.Card.Kind == CardNanMan {
			g.continueNanManAfterTarget(pending.SourceIndex, pending.AoeQueue, events)
			return nil
		}
		if pending.Card.Kind == CardWanJian {
			g.continueWanJianAfterTarget(pending.SourceIndex, pending.AoeQueue, events)
			return nil
		}
		return g.continueAoeAfterTarget(pending.SourceIndex, pending.Card, required, pending.AoeQueue, events)
	}
	g.Pending = nil
	damage := pending.Damage
	if damage <= 0 {
		damage = 1
	}
	resume := DamageResume{
		Mode:        damageResumeShaHit,
		Card:        pending.Card,
		ReturnIndex: pending.ReturnIndex,
		OfferQilin:  pending.Card.Kind == CardSha,
		IgnoreArmor: pending.IgnoreArmor,
	}
	if pending.LuanwuSha {
		resume.Mode = ""
		resume.OfferQilin = false
		resume.ResumeLuanwu = true
		resume.LuanwuOwner = pending.LuanwuOwner
	}
	return g.finalizeDamageHit(pending.SourceIndex, pending.TargetIndex, damage, pending.Card, resume, events)
}

func (g *Game) damageMessage(target *Player, sourceName string, damage int) string {
	return fmt.Sprintf("%s 受到【%s】%d 点伤害，体力 %d/%d", target.Name, sourceName, damage, target.HP, target.MaxHP)
}
