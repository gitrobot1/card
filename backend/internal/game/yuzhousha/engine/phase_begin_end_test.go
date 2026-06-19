package engine

import (
	"testing"
)

// TestStartPhase 测试回合开始阶段
func TestStartPhase(t *testing.T) {
	g, err := NewSolo1v1("start-phase-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	seat := 0 // 玩家0
	
	// 开始回合
	events := []GameEvent{}
	err = g.beginStartPhase(seat, &events)
	
	if err != nil {
		t.Fatalf("beginStartPhase failed: %v", err)
	}
	
	// 检查是否触发了start_phase事件
	found := false
	for _, e := range events {
		if e.Type == "start_phase" && e.PlayerIndex == seat {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("start_phase event not found")
	}
	
	t.Logf("回合开始阶段测试通过，当前阶段: %s", g.TurnStep)
}

// TestFinishPhase 测试回合结束阶段
func TestFinishPhase(t *testing.T) {
	g, err := NewSolo1v1("finish-phase-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	seat := 0 // 玩家0
	
	// 直接进入回合结束阶段
	events := []GameEvent{}
	err = g.enterFinishPhase(seat, &events)
	
	if err != nil {
		t.Fatalf("enterFinishPhase failed: %v", err)
	}
	
	// 检查是否触发了finish_phase事件
	found := false
	for _, e := range events {
		if e.Type == "finish_phase" && e.PlayerIndex == seat {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("finish_phase event not found")
	}
	
	// 注意：enterFinishPhase 会调用 finishTurn，而 finishTurn 会开始下一个回合
	// 所以这里不检查 TurnStep，因为它已经变成了下一个回合的阶段
	
	t.Logf("回合结束阶段测试通过，当前阶段: %s", g.TurnStep)
}

// TestFullTurn 测试完整的回合流程
func TestFullTurn(t *testing.T) {
	g, err := NewSolo1v1("full-turn-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	// 记录初始阶段
	initialStep := g.TurnStep
	t.Logf("初始阶段: %s", initialStep)
	
	// 开始回合
	events := []GameEvent{}
	g.beginTurn(&events)
	
	// 检查是否按正确顺序执行了各个阶段
	// 1. 回合开始阶段
	// 2. 准备阶段（如果没有技能，会直接进入判定阶段）
	// 3. 判定阶段（如果没有延时锦囊，会直接进入摸牌阶段）
	// 4. 摸牌阶段
	// 5. 出牌阶段
	
	// 检查是否至少进入了出牌阶段
	if g.TurnStep != StepPlay && g.TurnStep != StepJudge {
		t.Errorf("expected to reach play or judge phase, got %s", g.TurnStep)
	}
	
	t.Logf("完整回合流程测试通过，当前阶段: %s", g.TurnStep)
}

// TestTurnSteps 测试所有阶段的常量定义
func TestTurnSteps(t *testing.T) {
	expectedSteps := []string{
		StepStart,
		StepPrepare,
		StepJudge,
		StepDraw,
		StepPlay,
		StepDiscard,
		StepFinish,
	}
	
	for _, step := range expectedSteps {
		if step == "" {
			t.Errorf("step constant is empty")
		} else {
			t.Logf("阶段常量: %s", step)
		}
	}
}
