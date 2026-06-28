package engine

import "fmt"

// ============================================================================
// 阶段流转系统（参考 noname 03-phase-flow.md）
// phaseList 是数据数组，每个阶段是 GameEvent，通过 phaseLoop 循环驱动。
// 参考 noname: phase() 函数 step 8-11 阶段循环
// ============================================================================

// PhaseDef 阶段定义（参考 noname phaseList 数组）
type PhaseDef struct {
	Name    string // 阶段名（如 "phaseJudge", "phaseDraw", "phaseUse"）
	StepKey string // TurnStep 键（如 "judge", "draw", "play"）
	Message string // 阶段提示消息模板（%s 替换为玩家名）
}

// phaseList 六大阶段列表（参考 noname: phaseList = ["phaseZhunbei","phaseJudge","phaseDraw","phaseUse","phaseDiscard","phaseJieshu"]）
var phaseList = []PhaseDef{
	{Name: "phaseZhunbei", StepKey: StepPrepare, Message: "%s 准备阶段"},
	{Name: "phaseJudge", StepKey: StepJudge, Message: "%s 判定阶段"},
	{Name: "phaseDraw", StepKey: StepDraw, Message: "%s 摸牌阶段"},
	{Name: "phaseUse", StepKey: StepPlay, Message: "%s 出牌阶段"},
	{Name: "phaseDiscard", StepKey: StepDiscard, Message: "%s 弃牌阶段"},
	{Name: "phaseJieshu", StepKey: StepFinish, Message: "%s 回合结束阶段"},
}

// ============================================================================
// phaseLoop：全局回合循环（参考 noname: phaseLoop step 1-3）
// ============================================================================

// startPhaseLoop 启动回合阶段循环。
// 参考 noname: phase() step 8-11 阶段循环
// num 是阶段索引，从 0 开始（跳过 phaseZhunbei，直接从 phaseJudge 开始）
func (g *Game) startPhaseLoop(seat int, events *[]GameEvent) {
	// 从判定阶段开始（准备阶段已在 beginTurn 前处理）
	g.runPhaseStep(seat, 1, events) // phaseList[1] = phaseJudge
}

// runPhaseStep 执行当前阶段。
// 参考 noname: step 9 → player[phaseList[num]]() → step 10 cleanup → step 11 goto(8)
func (g *Game) runPhaseStep(seat, num int, events *[]GameEvent) {
	if g.IsFinished() {
		return
	}
	if num >= len(phaseList) {
		// 所有阶段完成 → 回合结束（参考 noname: trigger("phaseEnd")）
		g.finishPhaseLoop(seat, events)
		return
	}

	phase := phaseList[num]

	// 发送阶段切换事件（参考 noname: trigger("phaseChange")）
	g.TurnStep = phase.StepKey
	g.Pending = nil
	g.Message = fmt.Sprintf(phase.Message, g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        phase.Name,
		PlayerIndex: seat,
		Message:     g.Message,
	})

	// 创建阶段 GameEvent（参考 noname: player[phaseName]()）
	phaseEv := g.NewGameEvent(phase.Name, seat)
	phaseEv.Content = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		return g.executePhaseContent(seat, num, evs)
	}

	// 阶段完成后继续下一阶段（参考 noname: step 11 goto(8)）
	// 人类玩家的出牌阶段不自动推进，等待前端 EndPlay 请求
	phaseEv.OnAfter = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		if phase.StepKey == StepPlay && !g.Players[seat].IsAI && g.Pending == nil {
			// 人类玩家出牌阶段：不推进，等待 EndPlay
			return nil
		}
		g.runPhaseStep(seat, num+1, evs)
		return nil
	}

	// 启动阶段事件（参考 noname: player[phaseName]() → event.start()）
	g.StartEvent(phaseEv, events)
}

// SkipToPhase 跳过中间阶段，直接跳转到指定阶段。
// 参考 noname: player.skip("phaseUse") → checkSkipped() 在事件开始时自动跳过
// 用于乐不思蜀/兵粮寸断等跳过出牌/摸牌阶段后直接进入弃牌阶段。
func (g *Game) SkipToPhase(seat int, phaseName string, events *[]GameEvent) {
	for i, phase := range phaseList {
		if phase.Name == phaseName {
			g.runPhaseStep(seat, i, events)
			return
		}
	}
	// 未找到阶段，直接结束回合
	g.finishPhaseLoop(seat, events)
}

// ============================================================================
// 翻面机制（参考 noname: step 4 翻面检测）
// ============================================================================

// IsTurnedOver 检查玩家是否翻面（参考 noname: 翻面检测）
func (g *Game) IsTurnedOver(seat int) bool {
	if seat < 0 || seat >= len(g.Players) {
		return false
	}
	return g.Players[seat].TurnedOver
}

// TurnOver 将玩家翻面（参考 noname: 翻面，下回合跳过）
func (g *Game) TurnOver(seat int) {
	if seat >= 0 && seat < len(g.Players) {
		g.Players[seat].TurnedOver = true
	}
}

// clearTurnOver 清除翻面状态（翻面回合结束后自动清除）
func (g *Game) clearTurnOver(seat int) {
	if seat >= 0 && seat < len(g.Players) {
		g.Players[seat].TurnedOver = false
	}
}

// executePhaseContent 执行阶段的具体内容。
// 参考 noname: 各阶段 content 函数
func (g *Game) executePhaseContent(seat, num int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	phase := phaseList[num]

	switch phase.Name {
	case "phaseZhunbei":
		return g.executePreparePhase(seat, events)
	case "phaseJudge":
		return g.executeJudgePhaseContent(seat, events)
	case "phaseDraw":
		return g.executeDrawPhase(seat, events)
	case "phaseUse":
		return g.executePlayPhase(seat, events)
	case "phaseDiscard":
		return g.executeDiscardPhase(seat, events)
	case "phaseJieshu":
		return g.executeFinishPhase(seat, events)
	default:
		return nil
	}
}

// ============================================================================
// 各阶段 content 实现（参考 noname: phaseZhunbei/phaseJudge/phaseDraw/phaseUse/phaseDiscard/phaseJieshu）
// ============================================================================

// executePreparePhase 准备阶段（参考 noname: phaseZhunbei → trigger("phaseZhunbei")）
func (g *Game) executePreparePhase(seat int, events *[]GameEvent) error {
	if g.enterPreparePhase(seat, events) {
		if g.Players[seat].IsAI {
			g.runAIPreparePhase(seat, events)
		}
	}
	return nil
}

// executeJudgePhaseContent 判定阶段 content（参考 noname: phaseJudge step 0-3）
func (g *Game) executeJudgePhaseContent(seat int, events *[]GameEvent) error {
	if g.judgeAreaCount(seat) == 0 {
		return nil // 无判定牌，直接完成
	}
	return g.processNextJudgeCard(seat, events)
}

// executeDrawPhase 摸牌阶段（参考 noname: phaseDraw step 0-2）
func (g *Game) executeDrawPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}

	// 高达1号：绝境跳过摸牌阶段
	if g.gundamSkipDrawPhase(seat) {
		g.Message = fmt.Sprintf("%s 【绝境-高达一号】跳过摸牌阶段", g.Players[seat].Name)
		*events = append(*events, GameEvent{
			Type:        "draw_phase_skip",
			PlayerIndex: seat,
			Message:     g.Message,
		})
		return nil
	}

	// 摸牌阶段选择（突袭/洛义/双雄等）
	if g.isDrawPhaseChoicePending(seat) {
		if g.Players[seat].IsAI {
			runAIDrawPhase(g, seat, events)
		}
		return nil
	}

	// 正常摸牌
	drawCount := g.drawCountFor(seat)
	g.drawCards(seat, drawCount, events)

	// 触发摸牌阶段开始钩子
	g.runDrawPhaseStartHooks(seat, events)

	return nil
}

// executePlayPhase 出牌阶段（参考 noname: phaseUse step 0-6）
func (g *Game) executePlayPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}

	// 高达1号：斩将
	g.gundamZhanjiang(seat, events)

	// AI 自动出牌；人类玩家由前端驱动（Phase=Playing, TurnStep=StepPlay, Pending=nil）
	if g.Players[seat].IsAI {
		runAIPlayPhase(g, seat, events)
	}

	return nil
}

// executeDiscardPhase 弃牌阶段（参考 noname: phaseDiscard step 0-1）
func (g *Game) executeDiscardPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}

	// 清理阶段跳过标记（参考 noname: skipList.remove）
	g.Players[seat].SkipPlay = false
	g.Players[seat].SkipDraw = false

	// AI 自动弃牌
	if g.Players[seat].IsAI {
		g.autoDiscard(seat, events)
	}

	return nil
}

// executeFinishPhase 结束阶段（参考 noname: phaseJieshu → trigger("phaseJieshu")）
func (g *Game) executeFinishPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}

	// 高达1号：绝境手牌补到4
	g.gundamSyncHandSize(seat, events)

	// 破军：回合结束后，获得「营」中的牌
	g.startPojunGainIfNeeded(seat, events)

	// 触发回合结束钩子
	g.runTurnEndHooks(seat, events)

	// 清理回合状态
	g.runTurnEndCleanup(seat, events)
	g.Players[seat].Drunk = false

	return nil
}

// ============================================================================
// 回合结束与下一玩家（参考 noname: phaseEnd → phaseAfter → 下一玩家）
// ============================================================================

// finishPhaseLoop 所有阶段完成后，切换到下一玩家。
// 参考 noname: trigger("phaseEnd") → trigger("phaseAfter") → goto(1) 下一玩家
func (g *Game) finishPhaseLoop(seat int, events *[]GameEvent) {
	if g.IsFinished() {
		return
	}

	// 发送回合结束事件
	*events = append(*events, GameEvent{
		Type:        "turn_end",
		PlayerIndex: seat,
		Message:     fmt.Sprintf("%s 结束回合", g.Players[seat].Name),
	})

	// 切换到下一个玩家，但不自动开始回合。
	// 回合启动由外部 NextTurn 请求驱动（前端/AI），确保每个回合都是独立的"电梯旅程"。
	g.CurrentTurn = g.nextTurnSeat(seat)
	g.TurnStep = "" // 清空 TurnStep，标记回合已结束，AutoBeginTurnIfNeeded 据此初始化新回合
}
