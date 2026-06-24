package engine

import "fmt"

// ============================================================================
// GameEvent 核心状态机
// 参考 noname GameEvent: Before → Begin → Content → End → After
// 源码: noname/library/element/gameEvent.js
// ============================================================================

// EventPhase 事件生命周期阶段（参考 noname _triggered: 0-4）
type EventPhase int

const (
	EventPhaseCreated  EventPhase = 0 // 事件被创建，等待 checkSkipped
	EventPhaseBefore   EventPhase = 1 // Before 阶段（触发 {name}Before 钩子）
	EventPhaseBegin    EventPhase = 2 // Begin 阶段（触发 {name}Begin 钩子）
	EventPhaseContent  EventPhase = 3 // Content 阶段（执行事件实际内容）
	EventPhaseEnd      EventPhase = 4 // End 阶段（触发 {name}End 钩子）
	EventPhaseAfter    EventPhase = 5 // After 阶段（触发 {name}After 钩子）
	EventPhaseFinished EventPhase = 6 // 事件已完成
)

// EventType 事件类型（参考 noname event.type: "card" / "player" / "phase"）
type EventType string

const (
	EventTypeCard   EventType = "card"   // 卡牌事件（使用卡牌）
	EventTypePlayer EventType = "player" // 玩家事件（受伤、回复等）
	EventTypePhase  EventType = "phase"  // 阶段事件（回合阶段）
)

// GameEventInstance 游戏事件实例。
// 参考 noname GameEvent: name, player, source, target, card, num, type, _triggered,
// finished, cancelled, parent, childEvents, skill, forced, notrigger, nature, result
type GameEventInstance struct {
	// 基本属性（参考 noname 事件属性表）
	Name   string    // 事件名（如 "phaseUse", "damage", "useCard"）
	Player int       // 事件主体玩家座位号
	Source int       // 事件来源座位号（如伤害来源）
	Target int       // 事件目标座位号
	Targets []int    // 多目标数组（如 AOE 锦囊）
	Card   *Card     // 关联卡牌
	Cards  []Card    // 关联卡牌数组
	Num    int       // 数值（伤害值、摸牌数等）
	Type   EventType // 事件类型: "card" / "player" / "phase"

	// 生命周期状态（参考 noname _triggered: 0-4, finished, cancelled）
	Phase     EventPhase // 当前生命周期阶段
	Finished  bool       // 是否已完成（进入 End→After 流程）
	Cancelled bool       // 是否被取消（Begin 前取消走 Omitted）
	Skipped   bool       // 是否被跳过（skipList 检查结果）
	InContent bool       // 是否正在执行 content（参考 noname #inContent）

	// 事件结果（参考 noname event.result）
	Result *EventResult // 事件结果（判定结果、出牌结果等）

	// 事件内容钩子
	Content  func(g *Game, ev *GameEventInstance, events *[]GameEvent) error // 核心逻辑
	OnBefore func(g *Game, ev *GameEventInstance, events *[]GameEvent) error // Before 钩子
	OnBegin  func(g *Game, ev *GameEventInstance, events *[]GameEvent) error // Begin 钩子
	OnEnd    func(g *Game, ev *GameEventInstance, events *[]GameEvent) error // End 钩子
	OnAfter  func(g *Game, ev *GameEventInstance, events *[]GameEvent) error // After 钩子

	// 技能相关（参考 noname: skill, forced, notrigger, nature）
	Skill    string // 关联技能 ID
	Forced   bool   // 是否强制（锁定技）
	NoTrigger bool  // 是否跳过触发（参考 noname notrigger）
	Nature   string // 属性（fire/thunder/poison/ice 等）

	// 栈关系（参考 noname: parent, childEvents）
	Parent      *GameEventInstance   // 父事件
	ChildEvents []*GameEventInstance // 子事件列表

	// Step 控制（参考 noname: goto(N), redo()）
	Step      int // 当前 step 编号（配合 Step 编译器）
	NextStep  int // 下一个 step 编号（-1 表示顺序执行）
	RedoStep  bool // 是否重复当前 step

	// 额外数据
	Data map[string]interface{} // 事件携带的额外数据
}

// EventResult 事件结果（参考 noname event.result）
type EventResult struct {
	Bool   bool   // 判定结果 true/false
	Card   *Card  // 判定牌
	Number int    // 判定点数
	Suit   string // 判定花色
	Color  string // 判定颜色
	Judge  int    // 判定函数返回值（>0 成功, <0 失败, 0 无结果）
}

// ============================================================================
// skipList 跳过列表（参考 noname: player.skipList）
// ============================================================================

// SkipList 阶段跳过列表。
// 参考 noname: player.skipList 数组，checkSkipped() 检查 event.name 是否在列表中。
type SkipList struct {
	phases []string // 被跳过的阶段名列表
}

// Add 添加跳过标记（参考 noname: player.skip("phaseUse")）
func (sl *SkipList) Add(phase string) {
	if !sl.Contains(phase) {
		sl.phases = append(sl.phases, phase)
	}
}

// Remove 移除跳过标记（参考 noname: skipList.remove(name)）
func (sl *SkipList) Remove(phase string) {
	for i, p := range sl.phases {
		if p == phase {
			sl.phases = append(sl.phases[:i], sl.phases[i+1:]...)
			return
		}
	}
}

// Contains 检查是否在跳过列表中（参考 noname: skipList.includes(name)）
func (sl *SkipList) Contains(phase string) bool {
	for _, p := range sl.phases {
		if p == phase {
			return true
		}
	}
	return false
}

// Clear 清空跳过列表。
func (sl *SkipList) Clear() {
	sl.phases = nil
}

// ============================================================================
// 事件栈管理（参考 noname GameEventManager.eventStack）
// ============================================================================

// EventManager 事件管理器。
// 参考 noname GameEventManager: eventStack, rootEvent, tempEvent
type EventManager struct {
	stack     []*GameEventInstance // 事件栈，栈顶=当前正在处理的事件
	rootEvent *GameEventInstance   // 根事件（回合事件）
	tempEvent *GameEventInstance   // 临时事件（高优先级插入，参考 noname tempEvent）
}

// NewEventManager 创建事件管理器。
func NewEventManager() *EventManager {
	return &EventManager{}
}

// Depth 返回当前栈深度。
func (em *EventManager) Depth() int {
	return len(em.stack)
}

// Current 返回栈顶事件（当前正在处理的事件）。
func (em *EventManager) Current() *GameEventInstance {
	if len(em.stack) == 0 {
		return nil
	}
	return em.stack[len(em.stack)-1]
}

// push 事件入栈（参考 noname: event.manager.eventStack.push(event)）
func (em *EventManager) push(ev *GameEventInstance) {
	if em.rootEvent == nil {
		em.rootEvent = ev
	}
	em.stack = append(em.stack, ev)
}

// pop 事件出栈（参考 noname: event.manager.eventStack.pop()）
func (em *EventManager) pop() {
	if len(em.stack) > 0 {
		em.stack = em.stack[:len(em.stack)-1]
	}
}

// clear 清空栈。
func (em *EventManager) clear() {
	em.stack = nil
	em.rootEvent = nil
	em.tempEvent = nil
}

// ============================================================================
// 事件创建与启动（参考 noname: constructor + start()）
// ============================================================================

// NewGameEvent 创建游戏事件。
// 参考 noname: new GameEvent(name, trigger, manager)
func (g *Game) NewGameEvent(name string, player int) *GameEventInstance {
	return &GameEventInstance{
		Name:  name,
		Player: player,
		Phase: EventPhaseCreated,
		Type:  EventTypePhase, // 默认阶段事件类型
		NextStep: -1,          // -1 表示不使用 step 跳转
	}
}

// NewCardEvent 创建卡牌事件（参考 noname: type="card"）
func (g *Game) NewCardEvent(name string, player int, card Card) *GameEventInstance {
	ev := g.NewGameEvent(name, player)
	ev.Type = EventTypeCard
	ev.Card = &card
	return ev
}

// NewPlayerEvent 创建玩家事件（参考 noname: type="player"）
func (g *Game) NewPlayerEvent(name string, player int) *GameEventInstance {
	ev := g.NewGameEvent(name, player)
	ev.Type = EventTypePlayer
	return ev
}

// StartEvent 启动事件（参考 noname: event.start()）。
// 事件入栈 → 执行生命周期循环 → 出栈。
// 防重入：同一事件不能重复启动（参考 noname: if (this.#start) return this.#start）
func (g *Game) StartEvent(ev *GameEventInstance, events *[]GameEvent) error {
	if g.eventManager == nil {
		g.eventManager = NewEventManager()
	}

	// 防重入检查（参考 noname: if (this.#start) return this.#start）
	if ev.Phase > EventPhaseCreated {
		return nil // 事件已启动过，不重复执行
	}

	// 建立父子关系（参考 noname: if (this.parent) this.parent.childEvents.push(this)）
	ev.Parent = g.eventManager.Current()
	if ev.Parent != nil {
		ev.Parent.ChildEvents = append(ev.Parent.ChildEvents, ev)
	}

	// 入栈（参考 noname: this.manager.eventStack.push(this)）
	g.eventManager.push(ev)

	// 执行生命周期循环
	err := g.runEventLoop(ev, events)

	// 出栈（参考 noname: this.manager.eventStack.pop()）
	g.eventManager.pop()
	return err
}

// StartTempEvent 启动临时事件（高优先级插入）。
// 参考 noname: manager.tempEvent
func (g *Game) StartTempEvent(ev *GameEventInstance, events *[]GameEvent) error {
	if g.eventManager == nil {
		g.eventManager = NewEventManager()
	}
	// 保存当前事件为 tempEvent
	g.eventManager.tempEvent = g.eventManager.Current()
	return g.StartEvent(ev, events)
}

// ============================================================================
// 事件生命周期循环（参考 noname: loop() 行1082-1115）
// ============================================================================

// runEventLoop 执行事件生命周期循环。
// 参考 noname loop():
//
//	loop() {
//	    if checkSkipped() return;
//	    while(true) {
//	        waitNext();  ← 等待子事件完成
//	        if !finished:
//	            _triggered===0 → Before → 1
//	            _triggered===1 → Begin → 2
//	            else → content()
//	        else:
//	            _triggered===1 → Omitted
//	            _triggered===2 → End → 3
//	            _triggered===3 → After → 4
//	            else return
//	    }
//	}
func (g *Game) runEventLoop(ev *GameEventInstance, events *[]GameEvent) error {
	// Step 0: 检查是否被跳过（参考 noname: checkSkipped()）
	if g.checkEventSkipped(ev, events) {
		return nil
	}

	// 生命周期循环
	for {
		if g.IsFinished() {
			return nil
		}

		if !ev.Finished {
			// 事件未完成，推进生命周期
			switch ev.Phase {
			case EventPhaseCreated:
				// Before 阶段（参考 noname: _triggered===0 → Before → 1）
				ev.Phase = EventPhaseBefore
				if ev.OnBefore != nil {
					if err := ev.OnBefore(g, ev, events); err != nil {
						return err
					}
				}

			case EventPhaseBefore:
				// Begin 阶段（参考 noname: _triggered===1 → Begin → 2）
				ev.Phase = EventPhaseBegin
				if ev.OnBegin != nil {
					if err := ev.OnBegin(g, ev, events); err != nil {
						return err
					}
				}

			case EventPhaseBegin:
				// Content 阶段：执行事件核心逻辑（参考 noname: content(this)）
				ev.Phase = EventPhaseContent
				ev.InContent = true
				if ev.Content != nil {
					if err := ev.Content(g, ev, events); err != nil {
						ev.InContent = false
						return err
					}
				}
				ev.InContent = false
				// content 执行完毕后自动标记完成
				if !ev.Finished {
					ev.FinishEvent()
				}

			case EventPhaseContent:
				// 如果 content 没有调用 FinishEvent，手动推进
				ev.FinishEvent()

			default:
				return nil
			}

		} else {
			// 事件已完成，走 End → After 流程
			switch ev.Phase {
			case EventPhaseBefore:
				// Begin 前被取消/跳过 → Omitted（参考 noname: trigger("Omitted")）
				ev.Phase = EventPhaseFinished
				return nil

			case EventPhaseBegin, EventPhaseContent:
				// End 阶段（参考 noname: _triggered===2 → End → 3）
				if ev.Phase <= EventPhaseContent {
					ev.Phase = EventPhaseEnd
					if ev.OnEnd != nil {
						if err := ev.OnEnd(g, ev, events); err != nil {
							return err
						}
					}
				}

			case EventPhaseEnd:
				// After 阶段（参考 noname: _triggered===3 → After → 4）
				ev.Phase = EventPhaseAfter
				if ev.OnAfter != nil {
					if err := ev.OnAfter(g, ev, events); err != nil {
						return err
					}
				}

			case EventPhaseAfter:
				// 事件完成（参考 noname: else return）
				ev.Phase = EventPhaseFinished
				return nil

			default:
				return nil
			}
		}
	}
}

// ============================================================================
// 事件控制方法（参考 noname: finish/cancel/goto/redo/trigger）
// ============================================================================

// FinishEvent 标记事件完成（参考 noname: event.finish()）。
// 标记后事件将进入 End → After 流程。
func (ev *GameEventInstance) FinishEvent() {
	ev.Finished = true
}

// CancelEvent 取消事件（参考 noname: event.cancel()）。
// 如果在 Begin 前取消，跳过 End/After 直接结束（Omitted）。
// 如果在 Begin 后取消，仍走 End/After 流程。
func (ev *GameEventInstance) CancelEvent() {
	ev.Cancelled = true
	if ev.Phase < EventPhaseBegin {
		// Begin 前取消：直接标记完成，跳过 End/After（参考 noname Omitted）
		ev.Finished = true
	}
}

// Goto 跳转到指定 step（参考 noname: event.goto(N)）。
// 配合 Step 编译器使用，NextStep 在 runEventLoop 中消费。
func (ev *GameEventInstance) Goto(step int) {
	ev.NextStep = step
}

// Redo 重复当前 step（参考 noname: event.redo()）。
func (ev *GameEventInstance) Redo() {
	ev.RedoStep = true
}

// ============================================================================
// 子事件等待（参考 noname: waitNext()）
// ============================================================================

// WaitNext 等待所有子事件完成（参考 noname: await this.waitNext()）。
// 在 content 执行期间，如果有子事件正在处理，暂停当前事件直到子事件完成。
// Go 中通过同步调用 StartEvent 天然实现了 waitNext（子事件在 StartEvent 返回前已完成）。
// 此方法为占位，供后续异步场景使用。
func (ev *GameEventInstance) WaitNext(g *Game) {
	// Go 的同步模型：子事件在 StartEvent 中同步执行完毕后才返回，
	// 因此不需要显式 waitNext。保留此方法供后续异步场景扩展。
	_ = g
}

// ============================================================================
// 事件跳过检查（参考 noname: checkSkipped() 行1117-1124）
// ============================================================================

// checkEventSkipped 检查事件是否应被跳过。
// 参考 noname:
//
//	checkSkipped() {
//	    if (!player.skipList.includes(this.name)) return false;
//	    player.skipList.remove(this.name);
//	    if (lib.phaseName.includes(this.name))
//	        player.getHistory("skipped").add(this.name);
//	    this.finish();
//	    trigger(this.name + "Skipped");
//	    return true;
//	}
func (g *Game) checkEventSkipped(ev *GameEventInstance, events *[]GameEvent) bool {
	if ev.Player < 0 || ev.Player >= len(g.Players) {
		return false
	}
	p := &g.Players[ev.Player]

	// 使用 skipList 检查（参考 noname: player.skipList.includes(this.name)）
	// 兼容旧的 SkipPlay/SkipDraw 字段，逐步迁移到 skipList
	skipped := false

	// 旧字段兼容（将被 skipList 替代）
	switch ev.Name {
	case "phaseUse":
		if p.SkipPlay {
			p.SkipPlay = false
			skipped = true
		}
	case "phaseDraw":
		if p.SkipDraw {
			p.SkipDraw = false
			skipped = true
		}
	}

	// skipList 检查（新机制）
	if p.SkipPhases != nil && p.SkipPhases.Contains(ev.Name) {
		p.SkipPhases.Remove(ev.Name)
		skipped = true
	}

	if !skipped {
		return false
	}

	// 标记完成，跳过 content（参考 noname: this.finish()）
	ev.Finished = true
	ev.Skipped = true

	// 触发跳过事件（参考 noname: trigger(this.name + "Skipped")）
	skipMsg := fmt.Sprintf("%s 的%s被跳过", p.Name, ev.Name)
	*events = append(*events, GameEvent{
		Type:        ev.Name + "_skipped",
		PlayerIndex: ev.Player,
		Message:     skipMsg,
	})

	return true
}

// ============================================================================
// 子事件创建辅助（参考 noname: player.phase() → game.createEvent(...)）
// ============================================================================

// PushChildEvent 创建子事件并入栈。
// 参考 noname: parent.childEvents.push(child); eventStack.push(child)
func (g *Game) PushChildEvent(ev *GameEventInstance, events *[]GameEvent) error {
	return g.StartEvent(ev, events)
}

// FinishCurrentPhaseEvent 标记当前阶段事件完成。
// 用于外部（如 EndPlay HTTP 请求）触发阶段切换。
// 参考 noname: event.finish() → End → After → 下一阶段
func (g *Game) FinishCurrentPhaseEvent(events *[]GameEvent) error {
	current := g.eventManager.Current()
	if current == nil {
		return nil
	}
	current.FinishEvent()
	// 继续执行 runEventLoop 的剩余循环（End → After）
	return g.runEventLoop(current, events)
}
