package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

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
	g.SyncCounts()
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

	// 跳过已死亡玩家（HP <= 0），找到下一个存活玩家
	if g.AliveHP(seat) <= 0 {
		next := g.nextTurnSeat(seat)
		if next == seat {
			// 没有其他存活玩家，游戏应该结束
			return
		}
		g.CurrentTurn = next
		seat = next
	}

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

	// 回合事件完成。不在这里自动推进到下一回合。
	// 每个回合是独立的"电梯旅程"：beginTurn 启动，阶段执行完毕后由外部
	// （前端 NextTurn / AI finalize）发起新的 beginTurn。
}

// AutoBeginTurnIfNeeded 检测是否需要为新回合初始化状态。
// 由 finalize 在每次操作后调用。如果当前回合玩家的 TurnStep 为空
// （说明 finishPhaseLoop 已切换 CurrentTurn 但新回合还没初始化），
// 则初始化新回合状态（重置出杀次数、醉酒等），然后进入出牌阶段。
func (g *Game) AutoBeginTurnIfNeeded(events *[]GameEvent) {
	if g.IsFinished() || g.Phase != PhasePlaying {
		return
	}
	if g.TurnStep != "" || g.Pending != nil {
		return
	}
	seat := g.CurrentTurn
	if g.AliveHP(seat) <= 0 {
		return
	}

	// 初始化新回合状态（不执行完整回合循环，只重置状态）
	g.phaseStack = NewPhaseStack()
	g.eventManager = NewEventManager()
	g.Players[seat].ShaUsedThisTurn = false
	g.Players[seat].ShaExtraUsedThisTurn = false
	g.Players[seat].Drunk = false
	g.TurnStep = StepStart

	// 调用 beginTurn 初始化回合（创建 phaseEv，执行摸牌，进入出牌阶段）
	// 人类玩家在出牌阶段暂停，AI 自动完成整个回合循环
	g.beginTurn(events)
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

// triggerPhaseHook 触发阶段钩子（参考 noname: event.trigger(eventName)）。
// 通过 HookCall 分发到注册了对应 HookKind 的技能。
func (g *Game) triggerPhaseHook(seat int, hookName string, events *[]GameEvent) {
	switch hookName {
	case "phaseBeforeStart":
		g.runSkillHooks(events, skill.HookCall{Kind: skill.HookPhaseBeforeStart, Seat: seat, Role: skill.RolePlayer})
	case "phaseBeforeEnd":
		g.runSkillHooks(events, skill.HookCall{Kind: skill.HookPhaseBeforeEnd, Seat: seat, Role: skill.RolePlayer})
	case "phaseBeginStart":
		g.runSkillHooks(events, skill.HookCall{Kind: skill.HookPhaseBeginStart, Seat: seat, Role: skill.RolePlayer})
	case "phaseBegin":
		g.runSkillHooks(events, skill.HookCall{Kind: skill.HookPhaseBegin, Seat: seat, Role: skill.RolePlayer})
	case "phaseChange":
		g.runSkillHooks(events, skill.HookCall{Kind: skill.HookPhaseChange, Seat: seat, Role: skill.RolePlayer})
	case "phaseEnd":
		g.runSkillHooks(events, skill.HookCall{Kind: skill.HookPhaseEnd, Seat: seat, Role: skill.RolePlayer})
	case "roundStart":
		g.runSkillHooks(events, skill.HookCall{Kind: skill.HookRoundStart, Seat: seat, Role: skill.RoleGlobal})
	}
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
	// 人类玩家：手动推进到弃牌阶段（跳过 FinishCurrentPhaseEvent，直接调用 runPhaseStep）
	// 找到当前阶段在 phaseList 中的索引，推进到下一个阶段
	for i, phase := range phaseList {
		if phase.StepKey == StepPlay {
			g.runPhaseStep(seat, i+1, events)
			return nil
		}
	}
	// fallback
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
	g.SyncCounts()

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
	g.SyncCounts()
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
