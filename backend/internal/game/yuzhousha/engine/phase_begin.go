package engine

import (
	"fmt"
)

// beginStartPhase 进入回合开始阶段
// 这是回合的第一个阶段，用于触发"回合开始时"的技能
func (g *Game) beginStartPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	// 设置阶段为回合开始阶段
	g.TurnStep = StepStart
	g.Pending = nil
	g.Message = fmt.Sprintf("%s 回合开始阶段", g.Players[seat].Name)
	g.resetTimer()
	
	// 触发回合开始事件
	*events = append(*events, GameEvent{
		Type:        "start_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	
	// 触发"回合开始时"的技能钩子
	// 目前简化：不触发任何技能，直接继续
	// TODO: 未来在这里添加技能触发逻辑
	
	// 自动继续到准备阶段（如果没有技能需要玩家决策）
	return g.continueAfterStart(seat, events)
}

// continueAfterStart 回合开始阶段结束后，进入准备阶段
func (g *Game) continueAfterStart(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	
	// 进入准备阶段
	if g.enterPreparePhase(seat, events) {
		if g.Players[seat].IsAI {
			g.runAIPreparePhase(seat, events)
		}
		return nil
	}
	
	// 如果没有准备阶段技能，直接进入判定阶段
	return g.continueAfterPrepare(seat, events)
}

// PassStart 跳过回合开始阶段（如果没有技能触发）
func (g *Game) PassStart(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepStart || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	
	return g.continueAfterStart(seat, events)
}
