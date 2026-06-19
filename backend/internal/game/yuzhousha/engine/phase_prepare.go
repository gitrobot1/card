package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) alivePlayerCount() int {
	n := 0
	for i := range g.Players {
		if g.Players[i].HP > 0 {
			n++
		}
	}
	if n < 1 {
		return 1
	}
	return n
}

func (g *Game) shouldEnterPreparePhase(seat int) bool {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(seat) {
		if h.OffersPreparePhase(rt, seat) {
			return true
		}
	}
	return false
}

func (g *Game) enterPreparePhase(seat int, events *[]GameEvent) bool {
	if !g.shouldEnterPreparePhase(seat) {
		return false
	}
	g.TurnStep = StepPrepare
	g.Pending = nil
	g.Message = fmt.Sprintf("%s 准备阶段", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "prepare_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	return true
}

func (g *Game) PassPrepare(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	return g.continueAfterPrepare(seat, events)
}

func (g *Game) peekCountForSkill(seat int, skillID string) int {
	h, ok := skill.Lookup(skillID)
	if !ok {
		return 0
	}
	return skill.PeekCountFor(g.skillRuntime(nil), seat, h)
}

func (g *Game) StartPeekDeck(seat int, skillID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	h, ok := skill.Lookup(skillID)
	if !ok || h.PeekDeckConfig() == nil || !g.hasSkill(seat, skillID) {
		return ErrInvalidCard
	}
	count := skill.PeekCountFor(g.skillRuntime(nil), seat, h)
	if count == 0 {
		return ErrWrongPhase
	}
	revealed := make([]Card, 0, count)
	for i := 0; i < count; i++ {
		c := g.DrawPile[0]
		g.DrawPile = g.DrawPile[1:]
		revealed = append(revealed, c)
	}
	g.syncCounts()

	meta := h.Meta()
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   seat,
		TargetIndex:   seat,
		ReturnIndex:   seat,
		ResponseMode:  ResponseModePeekDeck,
		RevealedCards: revealed,
		SkillID:       skillID,
	}
	g.Message = fmt.Sprintf("%s 发动【%s】，请分配 %d 张牌至牌堆顶/底", g.Players[seat].Name, meta.Name, len(revealed))
	FillPendingRoles(g.Pending)
	g.resetTimer()
	g.appendSkillEvent(events, skillID, seat, seat, g.Message)
	*events = append(*events, GameEvent{
		Type:        "peek_deck_reveal",
		PlayerIndex: seat,
		Amount:      len(revealed),
		SkillID:     skillID,
		Message:     g.Message,
	})
	for i := range revealed {
		c := revealed[i]
		*events = append(*events, GameEvent{
			Type:        "peek_deck_show",
			PlayerIndex: seat,
			Card:        &c,
			SkillID:     skillID,
			Message:     fmt.Sprintf("【%s】 %s", meta.Name, c.Label),
		})
	}
	return nil
}

type PeekDeckRequest struct {
	TopCardIDs    []string
	BottomCardIDs []string
}

// GuanxingRequest 兼容旧调用方。
type GuanxingRequest = PeekDeckRequest

func (g *Game) FinishPeekDeck(seat int, req PeekDeckRequest, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModePeekDeck {
		return ErrWrongPhase
	}
	if g.Pending.TargetIndex != seat {
		return ErrNotYourTurn
	}
	revealed := g.Pending.RevealedCards
	if err := validatePeekDeckPartition(revealed, req.TopCardIDs, req.BottomCardIDs); err != nil {
		return err
	}
	topCards := orderCardsByIDs(revealed, req.TopCardIDs)
	bottomCards := orderCardsByIDs(revealed, req.BottomCardIDs)

	g.DrawPile = append(topCards, g.DrawPile...)
	g.DrawPile = append(g.DrawPile, bottomCards...)
	g.syncCounts()

	skillID := g.Pending.SkillID
	skillName := skillID
	if h, ok := skill.Lookup(skillID); ok && h.Meta().Name != "" {
		skillName = h.Meta().Name
	}

	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPrepare

	msg := fmt.Sprintf("%s 完成【%s】", g.Players[seat].Name, skillName)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "peek_deck_finish",
		PlayerIndex: seat,
		Amount:      len(topCards),
		SkillID:     skillID,
		Message:     msg,
	})
	return g.continueAfterPrepare(seat, events)
}

func (g *Game) FinishGuanxing(seat int, req GuanxingRequest, events *[]GameEvent) error {
	return g.FinishPeekDeck(seat, req, events)
}

func (g *Game) StartGuanxing(seat int, events *[]GameEvent) error {
	return g.StartPeekDeck(seat, skill.IDGuanxing, events)
}

func validatePeekDeckPartition(revealed []Card, topIDs, bottomIDs []string) error {
	if len(topIDs)+len(bottomIDs) != len(revealed) {
		return ErrInvalidCard
	}
	seen := make(map[string]struct{}, len(revealed))
	for _, id := range append(append([]string{}, topIDs...), bottomIDs...) {
		if id == "" {
			return ErrInvalidCard
		}
		if _, dup := seen[id]; dup {
			return ErrInvalidCard
		}
		found := false
		for _, c := range revealed {
			if c.ID == id {
				found = true
				break
			}
		}
		if !found {
			return ErrInvalidCard
		}
		seen[id] = struct{}{}
	}
	return nil
}

func orderCardsByIDs(revealed []Card, ids []string) []Card {
	out := make([]Card, 0, len(ids))
	for _, id := range ids {
		for _, c := range revealed {
			if c.ID == id {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

func (g *Game) continueAfterPrepare(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	// 进入判定阶段
	return g.enterJudgePhase(seat, events)
}

func (g *Game) autoFinishPeekDeck(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModePeekDeck {
		return ErrWrongPhase
	}
	revealed := g.Pending.RevealedCards
	salt := seat*31 + g.CurrentTurn*17 + len(revealed)
	top, bottom := randomPeekPartition(revealed, salt)
	if err := validatePeekDeckPartition(revealed, top, bottom); err == nil {
		return g.FinishPeekDeck(seat, PeekDeckRequest{TopCardIDs: top, BottomCardIDs: bottom}, events)
	}
	topCards, bottomCards := splitPeekCardsByIndex(revealed, salt)
	return g.applyPeekDeckSplit(seat, topCards, bottomCards, events)
}

func splitPeekCardsByIndex(revealed []Card, salt int) (top, bottom []Card) {
	if len(revealed) == 0 {
		return nil, nil
	}
	for i, c := range revealed {
		if (i+salt)%2 == 0 {
			top = append(top, c)
		} else {
			bottom = append(bottom, c)
		}
	}
	if len(top) == 0 && len(bottom) > 0 {
		top = append(top, bottom[0])
		bottom = bottom[1:]
	}
	if len(bottom) == 0 && len(top) > 1 {
		bottom = append(bottom, top[len(top)-1])
		top = top[:len(top)-1]
	}
	return top, bottom
}

func (g *Game) applyPeekDeckSplit(seat int, topCards, bottomCards []Card, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModePeekDeck {
		return ErrWrongPhase
	}
	if g.Pending.TargetIndex != seat {
		return ErrNotYourTurn
	}
	g.DrawPile = append(topCards, g.DrawPile...)
	g.DrawPile = append(g.DrawPile, bottomCards...)
	g.syncCounts()

	skillID := g.Pending.SkillID
	skillName := skillID
	if h, ok := skill.Lookup(skillID); ok && h.Meta().Name != "" {
		skillName = h.Meta().Name
	}

	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPrepare

	msg := fmt.Sprintf("%s 完成【%s】", g.Players[seat].Name, skillName)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "peek_deck_finish",
		PlayerIndex: seat,
		Amount:      len(topCards),
		SkillID:     skillID,
		Message:     msg,
	})
	return g.continueAfterPrepare(seat, events)
}

func (g *Game) aiPartitionPeekDeck(seat int, revealed []Card) (topIDs, bottomIDs []string) {
	if g.Pending == nil {
		return nil, nil
	}
	skillID := g.Pending.SkillID
	h, ok := skill.Lookup(skillID)
	if !ok {
		return defaultAIPeekAllTop(revealed)
	}
	cfg := h.PeekDeckConfig()
	if cfg == nil || cfg.AIPartition == nil {
		return defaultAIPeekAllTop(revealed)
	}
	views := make([]skill.PeekCardView, len(revealed))
	for i, c := range revealed {
		views[i] = skill.PeekCardView{ID: c.ID, Kind: c.Kind}
	}
	return cfg.AIPartition(g.skillRuntime(nil), seat, views)
}

func (g *Game) finishPeekDeckAsAI(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModePeekDeck {
		return ErrWrongPhase
	}
	revealed := g.Pending.RevealedCards
	top, bottom := g.aiPartitionPeekDeck(seat, revealed)
	salt := seat*31 + g.CurrentTurn*17 + len(revealed)
	if len(top)+len(bottom) != len(revealed) {
		top, bottom = randomPeekPartition(revealed, salt)
	}
	if err := validatePeekDeckPartition(revealed, top, bottom); err == nil {
		return g.FinishPeekDeck(seat, PeekDeckRequest{TopCardIDs: top, BottomCardIDs: bottom}, events)
	}
	topCards, bottomCards := splitPeekCardsByIndex(revealed, salt)
	return g.applyPeekDeckSplit(seat, topCards, bottomCards, events)
}

func defaultAIPeekAllTop(revealed []Card) (topIDs, bottomIDs []string) {
	for _, c := range revealed {
		topIDs = append(topIDs, c.ID)
	}
	return topIDs, nil
}

// randomPeekPartition 将亮出牌伪随机分配至顶/底（sim 与人类强交互兜底，保证合法分区）。
func randomPeekPartition(revealed []Card, salt int) (topIDs, bottomIDs []string) {
	if len(revealed) == 0 {
		return nil, nil
	}
	for i, c := range revealed {
		if (i+salt)%2 == 0 {
			topIDs = append(topIDs, c.ID)
		} else {
			bottomIDs = append(bottomIDs, c.ID)
		}
	}
	if len(topIDs) == 0 {
		topIDs = append(topIDs, bottomIDs[0])
		bottomIDs = bottomIDs[1:]
	}
	if len(bottomIDs) == 0 && len(topIDs) > 1 {
		bottomIDs = append(bottomIDs, topIDs[len(topIDs)-1])
		topIDs = topIDs[:len(topIDs)-1]
	}
	return topIDs, bottomIDs
}

func (g *Game) runAIPreparePhase(seat int, events *[]GameEvent) {
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return
	}
	for attempt := 0; attempt < 8; attempt++ {
		if !g.runAIActiveSkills(seat, events) {
			break
		}
		if g.Phase == PhaseResponse && g.Pending != nil && g.Pending.ResponseMode == ResponseModePeekDeck {
			_ = g.finishPeekDeckAsAI(seat, events)
		}
		if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
			return
		}
	}
	_ = g.PassPrepare(seat, events)
}

func (g *Game) isPeekDeckPending() bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModePeekDeck
}

func (g *Game) peekDeckSkillID() string {
	if g.Pending == nil {
		return ""
	}
	return g.Pending.SkillID
}

// enterJudgePhase 进入判定阶段，处理判定区中的所有延时锦囊
func (g *Game) enterJudgePhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	// 检查是否有判定区的牌需要处理
	if g.judgeAreaCount(seat) == 0 {
		// 没有判定牌，直接进入摸牌阶段
		return g.advanceToDrawPhase(seat, events)
	}
	
	// 进入判定阶段
	g.TurnStep = StepJudge
	g.Pending = nil
	g.Message = fmt.Sprintf("%s 判定阶段", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "judge_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	
	// 开始处理第一张判定牌
	return g.processNextJudgeCard(seat, events)
}

// processNextJudgeCard 处理判定区中的下一张牌
func (g *Game) processNextJudgeCard(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	// 检查是否还有判定牌需要处理
	if g.judgeAreaCount(seat) == 0 {
		// 所有判定牌处理完毕，进入摸牌阶段
		return g.advanceToDrawPhase(seat, events)
	}
	
	// 获取第一张判定牌（按照标准规则：乐不思蜀 → 兵粮寸断 → 闪电）
	// 这里简单处理：按判定区数组顺序处理
	judgeCard := g.Players[seat].JudgeArea[0]
	
	// 先触发"判定前"的无懈可击窗口
	return g.startJudgeWuxiekWindow(seat, judgeCard, events)
}

// startJudgeWuxiekWindow 在判定前启动无懈可击响应窗口
func (g *Game) startJudgeWuxiekWindow(seat int, judgeCard Card, events *[]GameEvent) error {
	responseMode := ""
	switch judgeCard.Kind {
	case CardLeBu:
		responseMode = ResponseModeWuxiekLebu
	case CardBingLiang:
		responseMode = ResponseModeWuxiekBingliang
	case CardShanDian:
		responseMode = ResponseModeWuxiekShandian
	default:
		// 未知判定牌类型（可能是空牌或已被移除），直接移除并继续
		g.removeJudgeCard(seat, judgeCard.ID)
		g.DiscardPile = append(g.DiscardPile, judgeCard)
		g.syncCounts()
		return g.processNextJudgeCard(seat, events)
	}
	
	// 创建响应队列：从当前回合玩家开始，逆时针顺序
	responseQueue := g.createResponseQueue(seat)
	
	// 检查队列中是否有玩家可以使用无懈可击
	hasWuxiek := false
	for _, playerSeat := range responseQueue {
		if g.Players[playerSeat].HP > 0 {
			// 检查该玩家手牌中是否有无懈可击
			for _, card := range g.Players[playerSeat].Hand {
				if card.Kind == CardWuxiek {
					hasWuxiek = true
					break
				}
			}
			if hasWuxiek {
				break
			}
		}
	}
	
	// 如果没有人有无懈可击，直接执行判定
	if !hasWuxiek {
		return g.executeJudge(seat, judgeCard, events)
	}
	
	// 启动无懈可击响应窗口
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:    seat, // 判定牌的拥有者
		TargetIndex:    -1,   // -1 表示任何人都可以响应无懈可击
		ReturnIndex:    seat,
		EffectTarget:   seat,
		Card:           judgeCard,
		ResponseMode:   responseMode,
		AllowWuxiek:    true,
		ResponseQueue:  responseQueue,
		ResponseIndex:  0, // 从队列第一个玩家开始
	}
	// 保存当前判定的座位号，以便响应结束后恢复判定流程
	g.Pending.SavedPending = &PendingCombat{
		EffectTarget: seat,
		Card:         judgeCard,
	}
	
	// 设置当前响应者
	g.Pending.ActorSeat = responseQueue[0]
	
	g.Message = fmt.Sprintf("可对【%s】使用【无懈可击】", judgeCard.Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "wuxiek_offer",
		PlayerIndex: seat,
		TargetIndex:  -1,
		Card:        &judgeCard,
		Message:     g.Message,
	})
	
	return nil
}

// createResponseQueue 创建响应队列，按照三国杀规则：从指定座位开始，逆时针顺序
func (g *Game) createResponseQueue(startSeat int) []int {
	queue := []int{}
	currentSeat := startSeat
	
	// 添加所有存活的玩家到队列（逆时针顺序）
	for i := 0; i < len(g.Players); i++ {
		if g.Players[currentSeat].HP > 0 {
			queue = append(queue, currentSeat)
		}
		// 逆时针：下家是 (currentSeat + 1) % len(g.Players)
		currentSeat = (currentSeat + 1) % len(g.Players)
	}
	
	return queue
}

// executeJudge 执行判定（无懈可击处理后）
func (g *Game) executeJudge(seat int, judgeCard Card, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	// 根据判定牌的类型执行相应的判定
	switch judgeCard.Kind {
	case CardLeBu:
		return g.resolveLebuJudge(seat, events)
	case CardBingLiang:
		return g.resolveBingliangJudge(seat, events)
	case CardShanDian:
		return g.resolveShandianJudge(seat, events)
	default:
		// 未知判定牌类型，跳过
		return g.processNextJudgeCard(seat, events)
	}
}

// resumeJudgeAfterWuxiek 无懈可击响应后的恢复逻辑
// 当判定前的无懈可击窗口关闭后，调用此函数
func (g *Game) resumeJudgeAfterWuxiek(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.SavedPending == nil {
		// 没有保存的判定信息，直接进入摸牌阶段
		return g.advanceToDrawPhase(seat, events)
	}
	
	// 恢复保存的判定信息
	savedPending := g.Pending.SavedPending
	judgeSeat := savedPending.EffectTarget
	judgeCard := savedPending.Card
	
	// 清除Pending状态
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepJudge
	
	// 检查是否有第二张无懈可击被打出
	// 如果有，SavedPending会被更新，这里需要处理这种情况
	// 简化版本：直接执行判定
	return g.executeJudge(judgeSeat, judgeCard, events)
}

// handleWuxiekCounterPass 处理反无懈可击窗口的跳过
func (g *Game) handleWuxiekCounterPass(events *[]GameEvent) error {
	if g.Pending == nil {
		return g.advanceToDrawPhase(g.CurrentTurn, events)
	}
	// 递归窗口结束：SavedPending 不为 nil → 无懈可击生效（锦囊已被抵消），回到出牌阶段
	if g.Pending.SavedPending != nil {
		savedPending := g.Pending.SavedPending
		g.Pending = nil
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = savedPending.SourceIndex
		g.resetTimer()
		return nil
	}
	// 初始窗口结束：没人出无懈可击 → 锦囊生效
	if g.Pending.ResponseMode == ResponseModeWuxiekTrick {
		return g.continueTrickAfterWuxiekPass(events)
	}
	// 判定牌无懈窗口结束
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.resetTimer()
	return nil
}

// advanceToDrawPhase 从判定阶段进入摸牌阶段
func (g *Game) advanceToDrawPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	// 检查是否需要跳过摸牌阶段（如兵粮寸断）
	if g.Players[seat].SkipDraw {
		// 重置跳过标记
		g.Players[seat].SkipDraw = false
		
		msg := fmt.Sprintf("%s 的摸牌阶段被跳过", g.Players[seat].Name)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "draw_phase_skip",
			PlayerIndex: seat,
			Message:     msg,
		})
		
		// 跳过摸牌阶段，直接进入出牌阶段
		return g.advanceToPlayPhase(seat, events)
	}
	
	// 进入摸牌阶段
	g.TurnStep = StepDraw
	g.Pending = nil
	g.Message = fmt.Sprintf("%s 摸牌阶段", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "draw_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	
	// 触发摸牌阶段开始时的技能钩子
	g.runDrawPhaseStartHooks(seat, events)
	
	// 执行摸牌
	drawCount := g.drawCountFor(seat)
	g.drawCards(seat, drawCount, events)
	
	// 摸牌后进入出牌阶段
	return g.advanceToPlayPhase(seat, events)
}

// runDrawPhaseStartHooks 运行摸牌阶段开始时的技能钩子
func (g *Game) runDrawPhaseStartHooks(seat int, events *[]GameEvent) {
	// 目前简化实现：不触发任何技能
	// TODO: 未来可以在这里添加具体技能的钩子调用
	// 例如：英姿（周瑜）、突袭（张辽&张郃）等
	
	// 注意：当前 skill.Decl 中没有定义 OnDrawPhaseStart 钩子
	// 如果需要添加摸牌阶段开始的技能触发，需要在 skill 包中添加相应的钩子定义
}

// advanceToPlayPhase 从摸牌阶段进入出牌阶段
func (g *Game) advanceToPlayPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	g.TurnStep = StepPlay
	g.Pending = nil
	g.Message = fmt.Sprintf("%s 出牌阶段", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "play_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	
	return nil
}

// resolveLebuJudge 处理乐不思蜀的判定
func (g *Game) resolveLebuJudge(seat int, events *[]GameEvent) error {
	// 翻判定牌
	judgeCard, ok := g.flipJudgeCard(events, seat)
	if !ok {
		return g.processNextJudgeCard(seat, events)
	}
	
	// 检查判定结果（红桃则生效，跳过出牌阶段）
	isRed := isRedSuit(judgeCard.Suit)
	
	// 移除判定牌
	g.removeJudgeCard(seat, "")
	g.DiscardPile = append(g.DiscardPile, judgeCard)
	g.syncCounts()
	
	if isRed {
		// 判定成功，跳过出牌阶段
		g.Players[seat].SkipPlay = true
		msg := fmt.Sprintf("%s 的【乐不思蜀】判定生效，跳过出牌阶段", g.Players[seat].Name)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "judge_result",
			PlayerIndex: seat,
			Card:        &judgeCard,
			Message:     msg,
		})
		
		// 消耗 SkipPlay 标志，移除判定区的乐不思蜀，跳过出牌阶段
		g.Players[seat].SkipPlay = false
		if card, ok := g.removeJudgeByKind(seat, CardLeBu); ok {
			g.DiscardPile = append(g.DiscardPile, card)
		}
		
		// 直接跳到弃牌阶段
		return g.advanceToDiscardPhase(seat, events)
	} else {
		// 判定失败，正常进行出牌阶段
		g.Players[seat].SkipPlay = false
		msg := fmt.Sprintf("%s 的【乐不思蜀】判定无效", g.Players[seat].Name)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "judge_result",
			PlayerIndex: seat,
			Card:        &judgeCard,
			Message:     msg,
		})
		
		// 继续处理下一张判定牌
		return g.processNextJudgeCard(seat, events)
	}
}

// resolveBingliangJudge 处理兵粮寸断的判定
func (g *Game) resolveBingliangJudge(seat int, events *[]GameEvent) error {
	// 翻判定牌
	judgeCard, ok := g.flipJudgeCard(events, seat)
	if !ok {
		return g.processNextJudgeCard(seat, events)
	}
	
	// 检查判定结果（红桃则生效，跳过摸牌阶段）
	isRed := isRedSuit(judgeCard.Suit)
	
	// 移除判定牌
	g.removeJudgeCard(seat, "")
	g.DiscardPile = append(g.DiscardPile, judgeCard)
	g.syncCounts()
	
	if isRed {
		// 判定成功，跳过摸牌阶段
		g.Players[seat].SkipDraw = true
		msg := fmt.Sprintf("%s 的【兵粮寸断】判定生效，跳过摸牌阶段", g.Players[seat].Name)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "judge_result",
			PlayerIndex: seat,
			Card:        &judgeCard,
			Message:     msg,
		})
	} else {
		// 判定失败
		msg := fmt.Sprintf("%s 的【兵粮寸断】判定无效", g.Players[seat].Name)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "judge_result",
			PlayerIndex: seat,
			Card:        &judgeCard,
			Message:     msg,
		})
	}
	
	// 继续处理下一张判定牌
	return g.processNextJudgeCard(seat, events)
}

// resolveShandianJudge 处理闪电的判定
func (g *Game) resolveShandianJudge(seat int, events *[]GameEvent) error {
	// 翻判定牌
	judgeCard, ok := g.flipJudgeCard(events, seat)
	if !ok {
		return g.processNextJudgeCard(seat, events)
	}
	
	// 检查判定结果（黑桃2-9则雷击）
	struck := isLightningStrike(judgeCard.Suit, judgeCard.Rank)
	
	// 移除判定牌
	g.removeJudgeCard(seat, "")
	
	if struck {
		// 闪电生效，造成3点雷电伤害
		g.DiscardPile = append(g.DiscardPile, judgeCard)
		g.syncCounts()
		
		msg := fmt.Sprintf("%s 的【闪电】判定生效，受到 3 点雷电伤害", g.Players[seat].Name)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "judge_result",
			PlayerIndex: seat,
			Card:        &judgeCard,
			Message:     msg,
			Damage:      3,
		})
		
		// 应用伤害
		g.applyDamageWithHook(seat, seat, 3, Card{Kind: CardShanDian, Name: "闪电", DamageType: DamageTypeThunder}, events)
	} else {
		// 闪电转移到下家
		g.transferShandian(seat, judgeCard, events)
	}
	
	// 继续处理下一张判定牌（如果有的话）
	return g.processNextJudgeCard(seat, events)
}

// advanceToDiscardPhase 进入弃牌阶段
func (g *Game) advanceToDiscardPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	g.TurnStep = StepDiscard
	g.Pending = nil
	g.Message = fmt.Sprintf("%s 弃牌阶段", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "discard_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	
	// 执行弃牌
	g.autoDiscard(seat, events)
	
	// 进入回合结束阶段
	return g.enterFinishPhase(seat, events)
}

// enterFinishPhase 进入回合结束阶段
func (g *Game) enterFinishPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	g.TurnStep = StepFinish
	g.Pending = nil
	g.Message = fmt.Sprintf("%s 回合结束阶段", g.Players[seat].Name)
	g.resetTimer()
	
	*events = append(*events, GameEvent{
		Type:        "finish_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	
	// 触发回合结束时的技能钩子
	g.runTurnEndHooks(seat, events)
	
	// 完成回合结束阶段，进入下一个回合
	return g.finishTurn(seat, events)
}

// finishTurn 完成回合结束阶段，进入下一个回合
func (g *Game) finishTurn(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	// 触发回合结束后的清理工作
	g.runTurnEndCleanup(seat, events)
	
	// 重置玩家状态
	g.Players[seat].Drunk = false
	
	// 发送回合结束事件
	*events = append(*events, GameEvent{
		Type:        "turn_end",
		PlayerIndex: seat,
		Message:     fmt.Sprintf("%s 结束回合", g.Players[seat].Name),
	})
	
	// 切换到下一个玩家
	g.CurrentTurn = g.nextTurnSeat(g.CurrentTurn)
	
	// 开始下一个玩家的回合
	g.beginTurn(events)
	
	return nil
}

// runTurnEndCleanup 运行回合结束时的清理工作
func (g *Game) runTurnEndCleanup(seat int, events *[]GameEvent) {
	// 清理回合相关的状态
	// 例如：清理技能计数器、重置临时状态等
	
	// TODO: 未来可以在这里添加具体技能的清理逻辑
}

// flipJudgeCard 翻开判定牌
func (g *Game) flipJudgeCard(events *[]GameEvent, seat int) (Card, bool) {
	if len(g.DrawPile) == 0 {
		g.refillDrawPile()
	}
	if len(g.DrawPile) == 0 {
		return Card{}, false
	}
	
	card := g.DrawPile[0]
	g.DrawPile = g.DrawPile[1:]
	g.syncCounts()
	
	*events = append(*events, GameEvent{
		Type:        "judge_flip",
		PlayerIndex: seat,
		Card:        &card,
		Message:     fmt.Sprintf("%s 翻开判定牌：%s", g.Players[seat].Name, card.Label),
	})
	
	return card, true
}

// transferShandian 将闪电转移到下家
func (g *Game) transferShandian(seat int, shandianCard Card, events *[]GameEvent) {
	nextSeat := g.nextTurnSeat(seat)
	g.Players[nextSeat].JudgeArea = append(g.Players[nextSeat].JudgeArea, shandianCard)
	
	msg := fmt.Sprintf("【闪电】转移到 %s 的判定区", g.Players[nextSeat].Name)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "shandian_transfer",
		PlayerIndex: seat,
		TargetIndex: nextSeat,
		Card:        &shandianCard,
		Message:     msg,
	})
}

