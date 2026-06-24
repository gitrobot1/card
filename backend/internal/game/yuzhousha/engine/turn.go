package engine

import "fmt"

func (g *Game) drawCards(seat, count int, events *[]GameEvent) {
	p := &g.Players[seat]
	for i := 0; i < count; i++ {
		if len(g.DrawPile) == 0 {
			g.refillDrawPile()
		}
		if len(g.DrawPile) == 0 {
			break
		}
		c := g.DrawPile[0]
		g.DrawPile = g.DrawPile[1:]
		p.Hand = append(p.Hand, c)
		*events = append(*events, GameEvent{
			Type:        "draw",
			PlayerIndex: seat,
			Card:        &c,
			Message:     fmt.Sprintf("%s 摸牌", p.Name),
		})
	}
	g.syncCounts()
}

func (g *Game) refillDrawPile() {
	if len(g.DiscardPile) <= 1 {
		return
	}
	rest := append([]Card(nil), g.DiscardPile[:len(g.DiscardPile)-1]...)
	g.DrawPile = g.shuffleCards(rest)
	g.DiscardPile = g.DiscardPile[len(g.DiscardPile)-1:]
}

func (g *Game) beginTurn(events *[]GameEvent) {
	if events == nil {
		events = &[]GameEvent{}
	}
	seat := g.CurrentTurn

	// 初始化阶段栈和事件管理器（每个回合开始时重置）
	g.phaseStack = NewPhaseStack()
	g.eventManager = NewEventManager()

	// 重置回合状态
	g.Players[seat].ShaUsedThisTurn = false
	g.Players[seat].ShaExtraUsedThisTurn = false
	g.Players[seat].Drunk = false
	g.setSkillCounter(seat, counterShaInPlayPhase, 0)
	g.resetPlayPhaseSkillCounters(seat)

	// 创建回合事件（参考 noname: player.phase() step 0-13）
	phaseEv := g.NewGameEvent("phase", seat)
	phaseEv.Type = EventTypePhase

	// OnBefore: 回合开始前（参考 noname: step 0 → trigger("phaseBefore")）
	phaseEv.OnBefore = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		// 预留：phaseBefore 钩子
		// 触发时机：回合正式开始前，所有玩家都能响应
		return nil
	}

	// Content: 回合主体逻辑（参考 noname: step 1-13）
	phaseEv.Content = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		// ============================================================
		// step 1: 轮数检测 + 构建阶段列表
		// 参考 noname: game.phaseNumber++; roundStart 检测
		// ============================================================
		g.tryAdvanceRound(seat, evs)

		// ============================================================
		// step 2: trigger("phaseBeforeStart")
		// 用途：1v1 武将登场、回合开始前技能
		// 相关技能：当先（廖化）- 在此插入额外出牌阶段
		// ============================================================
		g.triggerPhaseHook(seat, "phaseBeforeStart", evs)

		// ============================================================
		// step 3: trigger("phaseBeforeEnd")
		// 用途：纵傀（sp贾诩）- 回合开始时给其他角色贴标记
		// ============================================================
		g.triggerPhaseHook(seat, "phaseBeforeEnd", evs)

		// ============================================================
		// step 4: 翻面检测（参考 noname: 翻面则 event.cancel() 跳过回合）
		// 相关技能：据守（曹仁）、潜袭（sp曹仁）、冲阵（sp赵云）等
		// ============================================================
		if g.IsTurnedOver(seat) {
			g.clearTurnOver(seat)
			msg := fmt.Sprintf("%s 翻面，跳过本回合", g.Players[seat].Name)
			g.Message = msg
			*evs = append(*evs, GameEvent{
				Type:        "turn_skip",
				PlayerIndex: seat,
				Message:     msg,
			})
			ev.CancelEvent()
			return nil
		}

		// ============================================================
		// step 5-6: trigger("phaseBeginStart") → trigger("phaseBegin")
		// 用途：国战明置武将、回合正式开始钩子
		// 相关技能：化身（左慈）- 回合开始时获得新技能
		// ============================================================
		g.triggerPhaseHook(seat, "phaseBeginStart", evs)
		g.triggerPhaseHook(seat, "phaseBegin", evs)

		// step 7: 回合开始阶段（准备阶段技能）
		g.beginStartPhase(seat, evs)

		// ============================================================
		// step 8-11: 阶段循环（参考 noname: phaseChange → 各阶段 → goto(8)）
		// 每个阶段是独立 GameEvent，可被 skipList 跳过
		// 插入未知阶段：在 phaseList 加一项 + executePhaseContent 加分支
		// ============================================================
		g.triggerPhaseHook(seat, "phaseChange", evs)
		g.startPhaseLoop(seat, evs)
		return nil
	}

	// OnEnd: 回合结束（参考 noname: trigger("phaseEnd")）
	phaseEv.OnEnd = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		g.triggerPhaseHook(seat, "phaseEnd", evs)
		return nil
	}

	// OnAfter: 回合结束后（参考 noname: step 12 → trigger("phaseAfter")）
	phaseEv.OnAfter = func(g *Game, ev *GameEventInstance, evs *[]GameEvent) error {
		// 预留：phaseAfter 钩子
		// 用途：回合完全结束后清理、触发延迟效果
		return nil
	}

	g.StartEvent(phaseEv, events)
}

// ============================================================================
// 轮数管理（参考 noname: game.roundNumber++; roundStart 检测）
// ============================================================================

// tryAdvanceRound 检测是否新的一轮，触发 roundStart 事件。
// 参考 noname: content.js 行3650-3676
func (g *Game) tryAdvanceRound(seat int, events *[]GameEvent) {
	// 当 seat=0（首位玩家）且这是新一轮的第一个回合时
	if seat == 0 && g.isNewRound() {
		g.RoundNumber++
		g.triggerPhaseHook(seat, "roundStart", events)
	}
}

// isNewRound 判断是否新的一轮（首位玩家开始新回合）。
func (g *Game) isNewRound() bool {
	// 简单判断：seat=0 的回合开始 = 新一轮
	return true
}

// ============================================================================
// 阶段钩子触发（预留，供后续 trigger 系统接入）
// ============================================================================

// triggerPhaseHook 触发阶段钩子（预留接口）。
// 当前为空实现，后续接入 trigger 系统后，会遍历所有监听此事件名的技能并执行。
// 参考 noname: event.trigger(eventName)
func (g *Game) triggerPhaseHook(seat int, hookName string, events *[]GameEvent) {
	// TODO: 接入 trigger 系统后，替换为：
	// g.eventManager.Current().trigger(hookName)
	_ = seat
	_ = hookName
	_ = events
}
func (g *Game) EndPlay(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase == PhaseResponse {
		return ErrPendingCombat
	}
	if g.Phase != PhasePlaying || g.CurrentTurn != seat {
		return ErrNotYourTurn
	}
	if g.TurnStep != StepPlay {
		return ErrWrongPhase
	}
	// 标记当前阶段事件完成（参考 noname: event.finish() → 进入 End→After → 下一阶段）
	return g.FinishCurrentPhaseEvent(events)
}

func (g *Game) DiscardCards(seat int, cardIDs []string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepDiscard || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	p := &g.Players[seat]
	cap := g.handRetainLimit(seat)
	need := len(p.Hand) - cap
	if need <= 0 {
		return ErrWrongPhase
	}
	if len(cardIDs) != need {
		return ErrInvalidDiscardCount
	}

	seen := make(map[string]struct{}, len(cardIDs))
	for _, id := range cardIDs {
		if id == "" {
			return ErrInvalidCard
		}
		if _, dup := seen[id]; dup {
			return ErrInvalidCard
		}
		seen[id] = struct{}{}
		if _, _, ok := g.findCard(seat, id); !ok {
			return ErrInvalidCard
		}
	}

	discarded := make([]Card, 0, len(cardIDs))
	for _, id := range cardIDs {
		idx, _, ok := g.findCard(seat, id)
		if !ok {
			return ErrInvalidCard
		}
		played := g.removeHandCard(seat, idx, events)
		g.DiscardPile = append(g.DiscardPile, played)
		discarded = append(discarded, played)
	}
	g.syncCounts()

	discardMsg := fmt.Sprintf("%s 弃牌", p.Name)
	for i := range discarded {
		*events = append(*events, GameEvent{
			Type:        "discard",
			PlayerIndex: seat,
			Card:        &discarded[i],
			Message:     discardMsg,
			Amount:      0,
		})
	}
	g.runCardsDiscardedHooks(seat, "discard_phase", discarded, events)
	g.Message = discardMsg
	return g.endTurn(events)
}

func (g *Game) autoDiscard(seat int, events *[]GameEvent) {
	p := &g.Players[seat]
	cap := g.handRetainLimit(seat)
	need := len(p.Hand) - cap
	if need <= 0 {
		return
	}
	discarded := make([]Card, 0, need)
	for len(p.Hand) > cap {
		c := p.Hand[len(p.Hand)-1]
		p.Hand = p.Hand[:len(p.Hand)-1]
		g.DiscardPile = append(g.DiscardPile, c)
		discarded = append(discarded, c)
	}
	g.syncCounts()
	discardMsg := fmt.Sprintf("%s 弃牌", p.Name)
	for i := range discarded {
		*events = append(*events, GameEvent{
			Type:        "discard",
			PlayerIndex: seat,
			Card:        &discarded[i],
			Message:     discardMsg,
			Amount:      0,
		})
	}
	if len(discarded) > 0 {
		g.Message = discardMsg
		g.runCardsDiscardedHooks(seat, "discard_phase", discarded, events)
	}
}

func (g *Game) endTurn(events *[]GameEvent) error {
	seat := g.CurrentTurn

	// 高达1号：绝境手牌补到4
	g.gundamSyncHandSize(seat, events)

	// 破军：回合结束后，获得「营」中的牌
	g.startPojunGainIfNeeded(seat, events)

	// 进入回合结束阶段
	return g.enterFinishPhase(seat, events)
}
