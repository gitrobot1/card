package engine

// PhaseStack 阶段栈：实现电梯式阶段管理。
// 参考 noname GameEvent 生命周期：Before → Begin → Content → End → After
// 每个阶段 Push 时自动保存当前状态，Pop 时自动恢复并调用 OnResume 回调。
// 替换原来分散在 g.Pending / SavedPending / dyingContext / damageAftermath 的手动管理。
type PhaseStack struct {
	stack []PhaseFrame
}

// PhaseFrame 阶段帧：保存一个阶段的完整状态。
type PhaseFrame struct {
	Phase    string         // playing / response / hp_change
	TurnStep string         // 当前回合阶段
	Pending  *PendingCombat // 阶段挂起状态
	Resume   PhaseResume    // 恢复信息（含 OnResume 回调）
	// 额外保存的上下文（濒死、伤害链等），逐步迁移到 PhaseStack 管理
	DyingContext      *DyingContext
	DamageAftermath   *DamageAftermath
	LeijiSavedPending *PendingCombat
}

// PhaseResume 阶段恢复信息。
// OnResume 回调在 Pop 恢复状态后自动执行，用于继续被中断的上层流程。
type PhaseResume struct {
	// OnResume 阶段恢复回调：Pop 时自动调用。
	// 返回 error 表示恢复失败（如游戏已结束）。
	// 参考 noname: 每个事件的 content 函数，事件完成后自动继续上层。
	OnResume func(g *Game, events *[]GameEvent) error
}

// NewPhaseStack 创建空栈。
func NewPhaseStack() *PhaseStack {
	return &PhaseStack{}
}

// Depth 返回当前栈深度。
func (ps *PhaseStack) Depth() int {
	return len(ps.stack)
}

// Current 返回栈顶帧（不弹出）。
func (ps *PhaseStack) Current() *PhaseFrame {
	if len(ps.stack) == 0 {
		return nil
	}
	return &ps.stack[len(ps.stack)-1]
}

// Push 保存当前游戏状态并进入新阶段。
// 调用者负责在 Push 后设置 g.Phase / g.Pending / g.TurnStep。
// 参考 noname: event.manager.eventStack.push(event)
func (ps *PhaseStack) Push(g *Game, resume PhaseResume) {
	frame := PhaseFrame{
		Phase:            g.Phase,
		TurnStep:         g.TurnStep,
		Pending:          clonePending(g.Pending),
		Resume:           resume,
		DyingContext:      g.dyingContext,
		DamageAftermath:   g.damageAftermath,
		LeijiSavedPending: g.leijiSavedPending,
	}
	ps.stack = append(ps.stack, frame)
}

// Pop 恢复上一个阶段的游戏状态，并调用 OnResume 回调。
// 返回 true 表示 OnResume 执行成功（或无需执行），false 表示恢复失败。
// 参考 noname: event.manager.eventStack.pop() + 父事件继续
func (ps *PhaseStack) Pop(g *Game, events *[]GameEvent) bool {
	if len(ps.stack) == 0 {
		return false
	}
	frame := ps.stack[len(ps.stack)-1]
	ps.stack = ps.stack[:len(ps.stack)-1]

	// 恢复游戏状态
	g.Phase = frame.Phase
	g.TurnStep = frame.TurnStep
	g.Pending = frame.Pending
	g.dyingContext = frame.DyingContext
	g.damageAftermath = frame.DamageAftermath
	g.leijiSavedPending = frame.LeijiSavedPending

	return true
}

// PopAndResume 恢复上一个阶段并执行 OnResume 回调。
// 推荐使用此方法而非单独的 Pop。
func (ps *PhaseStack) PopAndResume(g *Game, events *[]GameEvent) error {
	if len(ps.stack) == 0 {
		return nil
	}
	frame := ps.stack[len(ps.stack)-1]
	resume := frame.Resume

	if !ps.Pop(g, events) {
		return nil
	}

	if resume.OnResume != nil {
		return resume.OnResume(g, events)
	}
	return nil
}

// PopTo 弹出直到（包含）指定深度的阶段。
func (ps *PhaseStack) PopTo(g *Game, depth int) *PhaseFrame {
	if depth < 0 || depth >= len(ps.stack) {
		return nil
	}
	for len(ps.stack) > depth {
		ps.Pop(g, nil)
	}
	return ps.Current()
}

// Clear 清空栈。
func (ps *PhaseStack) Clear() {
	ps.stack = nil
}

// Clone 深拷贝。
func (ps *PhaseStack) Clone() *PhaseStack {
	if ps == nil {
		return nil
	}
	c := &PhaseStack{stack: make([]PhaseFrame, len(ps.stack))}
	for i, f := range ps.stack {
		c.stack[i] = PhaseFrame{
			Phase:            f.Phase,
			TurnStep:         f.TurnStep,
			Pending:          clonePending(f.Pending),
			DyingContext:      f.DyingContext,
			DamageAftermath:   f.DamageAftermath,
			LeijiSavedPending: f.LeijiSavedPending,
		}
	}
	return c
}

func clonePending(p *PendingCombat) *PendingCombat {
	if p == nil {
		return nil
	}
	cp := *p
	if len(p.AoeQueue) > 0 {
		cp.AoeQueue = make([]int, len(p.AoeQueue))
		copy(cp.AoeQueue, p.AoeQueue)
	}
	if len(p.ResponseQueue) > 0 {
		cp.ResponseQueue = make([]int, len(p.ResponseQueue))
		copy(cp.ResponseQueue, p.ResponseQueue)
	}
	if len(p.WuxiekChain) > 0 {
		cp.WuxiekChain = make([]WuxiekEntry, len(p.WuxiekChain))
		copy(cp.WuxiekChain, p.WuxiekChain)
	}
	if len(p.RevealedCards) > 0 {
		cp.RevealedCards = make([]Card, len(p.RevealedCards))
		copy(cp.RevealedCards, p.RevealedCards)
	}
	if len(p.WuguRevealedAll) > 0 {
		cp.WuguRevealedAll = make([]Card, len(p.WuguRevealedAll))
		copy(cp.WuguRevealedAll, p.WuguRevealedAll)
	}
	if len(p.ModifyCandidates) > 0 {
		cp.ModifyCandidates = make([]int, len(p.ModifyCandidates))
		copy(cp.ModifyCandidates, p.ModifyCandidates)
	}
	cp.SavedPending = clonePending(p.SavedPending)
	return &cp
}

// ============================================================================
// Game 便捷方法
// ============================================================================

// PushPhase 保存当前阶段，设置新阶段。
// 参考 noname: game.createEvent(name) → event.start()
func (g *Game) PushPhase(phase, turnStep string, pending *PendingCombat, resume PhaseResume) {
	g.phaseStack.Push(g, resume)
	g.Phase = phase
	g.TurnStep = turnStep
	g.Pending = pending
}

// PopPhase 恢复上一个阶段并执行 OnResume 回调。
// 参考 noname: event.finish() → 自动回到父事件
func (g *Game) PopPhase(events *[]GameEvent) error {
	return g.phaseStack.PopAndResume(g, events)
}

// HasPushedPhase 检查是否有保存的阶段（栈非空）。
func (g *Game) HasPushedPhase() bool {
	return g.phaseStack.Depth() > 0
}

// PhaseStackDepth 返回当前栈深度（用于调试）。
func (g *Game) PhaseStackDepth() int {
	return g.phaseStack.Depth()
}
