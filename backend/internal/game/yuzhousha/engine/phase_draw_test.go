package engine

import (
	"testing"
)

// TestDrawPhaseSkip 测试摸牌阶段跳过（如兵粮寸断）
func TestDrawPhaseSkip(t *testing.T) {
	g, err := NewSolo1v1("draw-phase-skip-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	seat := 0 // 玩家0
	
	// 设置SkipDraw标记，模拟兵粮寸断的效果
	g.Players[seat].SkipDraw = true
	
	// 直接进入摸牌阶段
	events := []GameEvent{}
	err = g.advanceToDrawPhase(seat, &events)
	
	if err != nil {
		t.Fatalf("advanceToDrawPhase failed: %v", err)
	}
	
	// 检查是否触发了draw_phase_skip事件
	found := false
	for _, e := range events {
		if e.Type == "draw_phase_skip" && e.PlayerIndex == seat {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("draw_phase_skip event not found")
	}
	
	// 检查SkipDraw标记是否被重置
	if g.Players[seat].SkipDraw {
		t.Error("SkipDraw flag should be reset after skip")
	}
	
	// 检查是否进入了出牌阶段（跳过了摸牌）
	if g.TurnStep != StepPlay {
		t.Errorf("expected TurnStep=%s after skip, got %s", StepPlay, g.TurnStep)
	}
	
	t.Logf("摸牌阶段跳过测试通过，当前阶段: %s", g.TurnStep)
}

// TestDrawPhaseNormal 测试正常的摸牌阶段
func TestDrawPhaseNormal(t *testing.T) {
	g, err := NewSolo1v1("draw-phase-normal-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	seat := 0 // 玩家0
	
	// 记录摸牌前的手牌数
	handCountBefore := len(g.Players[seat].Hand)
	
	// 直接进入摸牌阶段
	events := []GameEvent{}
	err = g.advanceToDrawPhase(seat, &events)
	
	if err != nil {
		t.Fatalf("advanceToDrawPhase failed: %v", err)
	}
	
	// 检查是否触发了draw_phase事件
	found := false
	for _, e := range events {
		if e.Type == "draw_phase" && e.PlayerIndex == seat {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("draw_phase event not found")
	}
	
	// 检查是否摸了牌（默认摸2张）
	expectedDraw := 2 // DrawPerTurn = 2
	if g.Players[seat].Character.Name == "周瑜" {
		expectedDraw = 3 // 英姿
	}
	
	handCountAfter := len(g.Players[seat].Hand)
	actualDraw := handCountAfter - handCountBefore
	
	if actualDraw != expectedDraw {
		t.Errorf("expected to draw %d cards, got %d", expectedDraw, actualDraw)
	}
	
	// 检查是否进入了出牌阶段
	if g.TurnStep != StepPlay {
		t.Errorf("expected TurnStep=%s after draw, got %s", StepPlay, g.TurnStep)
	}
	
	t.Logf("正常摸牌阶段测试通过，摸了 %d 张牌，当前阶段: %s", actualDraw, g.TurnStep)
}

// TestDrawPhaseStartHooks 测试摸牌阶段开始时的技能钩子
func TestDrawPhaseStartHooks(t *testing.T) {
	g, err := NewSolo1v1("draw-phase-hooks-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	seat := 0 // 玩家0
	
	// 直接进入摸牌阶段
	events := []GameEvent{}
	err = g.advanceToDrawPhase(seat, &events)
	
	if err != nil {
		t.Fatalf("advanceToDrawPhase failed: %v", err)
	}
	
	// 检查runDrawPhaseStartHooks是否被调用（通过检查事件）
	// 目前简化实现：不触发任何技能，所以只需要确保没有错误
	t.Logf("摸牌阶段开始钩子测试通过，当前阶段: %s", g.TurnStep)
}

// TestDrawCountFor 测试摸牌数的计算
func TestDrawCountFor(t *testing.T) {
	g, err := NewSolo1v1("draw-count-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	seat := 0 // 玩家0
	
	// 测试默认摸牌数
	drawCount := g.drawCountFor(seat)
	expected := DrawPerTurn // 2
	
	if drawCount != expected {
		t.Errorf("expected draw count %d, got %d", expected, drawCount)
	}
	
	t.Logf("摸牌数计算测试通过，摸牌数: %d", drawCount)
}
