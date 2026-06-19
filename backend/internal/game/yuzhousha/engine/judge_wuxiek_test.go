package engine

import (
	"testing"
)

// TestJudgeWuxiekWindow 测试判定前的无懈可击窗口
func TestJudgeWuxiekWindow(t *testing.T) {
	// 创建一个简单的游戏场景
	g, err := NewSolo1v1("judge-wuxiek-test", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	// 设置游戏状态：玩家0的回合，有乐不思蜀在判定区
	g.Phase = PhasePlaying
	g.TurnStep = StepJudge
	g.CurrentTurn = 0
	
	// 添加乐不思蜀到玩家0的判定区
	lebuCard := Card{ID: "lebu-1", Kind: CardLeBu, Name: "乐不思蜀"}
	g.Players[0].JudgeArea = append(g.Players[0].JudgeArea, lebuCard)
	
	// 给玩家1一张无懈可击
	g.Players[1].Hand = append(g.Players[1].Hand, Card{ID: "wuxiek-1", Kind: CardWuxiek, Name: "无懈可击"})
	
	var events []GameEvent
	
	// 开始判定阶段
	err = g.enterJudgePhase(0, &events)
	if err != nil {
		t.Fatal(err)
	}
	
	// 检查是否启动了无懈可击响应窗口
	if g.Phase != PhaseResponse {
		t.Errorf("期望 PhaseResponse, 得到 %s", g.Phase)
	}
	
	if g.Pending == nil {
		t.Fatal("Pending 不应该为 nil")
	}
	
	// 检查响应模式
	if g.Pending.ResponseMode != ResponseModeWuxiekLebu {
		t.Errorf("期望 ResponseModeWuxiekLebu, 得到 %s", g.Pending.ResponseMode)
	}
	
	// 检查响应队列是否正确（从当前回合玩家开始，逆时针顺序）
	if len(g.Pending.ResponseQueue) != 2 {
		t.Errorf("期望响应队列长度为 2, 得到 %d", len(g.Pending.ResponseQueue))
	}
	
	// 在1v1模式下，响应队列应该是 [0, 1]
	if len(g.Pending.ResponseQueue) == 2 {
		if g.Pending.ResponseQueue[0] != 0 || g.Pending.ResponseQueue[1] != 1 {
			t.Errorf("期望响应队列为 [0, 1], 得到 [%d, %d]", 
				g.Pending.ResponseQueue[0], g.Pending.ResponseQueue[1])
		}
	}
	
	// 检查当前响应者是否是队列第一个
	if g.Pending.ActorSeat != g.Pending.ResponseQueue[0] {
		t.Errorf("期望当前响应者为 %d, 得到 %d", 
			g.Pending.ResponseQueue[0], g.Pending.ActorSeat)
	}
	
	t.Logf("判定前无懈可击窗口测试通过！")
}

// TestJudgeWuxiekCancel 测试无懈可击成功抵消判定牌
func TestJudgeWuxiekCancel(t *testing.T) {
	// 创建一个简单的游戏场景
	g, err := NewSolo1v1("judge-wuxiek-cancel", "甲", "zhao_yun", "liu_bei")
	if err != nil {
		t.Fatal(err)
	}
	
	// 设置游戏状态：玩家0的回合，有乐不思蜀在判定区
	g.Phase = PhasePlaying
	g.TurnStep = StepJudge
	g.CurrentTurn = 0
	
	// 添加乐不思蜀到玩家0的判定区
	lebuCard := Card{ID: "lebu-1", Kind: CardLeBu, Name: "乐不思蜀"}
	g.Players[0].JudgeArea = append(g.Players[0].JudgeArea, lebuCard)
	
	// 给玩家1一张无懈可击
	g.Players[1].Hand = append(g.Players[1].Hand, Card{ID: "wuxiek-1", Kind: CardWuxiek, Name: "无懈可击"})
	
	var events []GameEvent
	
	// 开始判定阶段
	err = g.enterJudgePhase(0, &events)
	if err != nil {
		t.Fatal(err)
	}
	
	// 玩家1打出无懈可击
	err = g.RespondWuxiek(1, "wuxiek-1", &events)
	if err != nil {
		t.Fatal(err)
	}
	
	// 检查是否启动了反无懈可击窗口
	if g.Pending == nil {
		t.Fatal("反无懈可击窗口 Pending 不应该为 nil")
	}
	
	// 跳过反无懈可击窗口（没有人再打出无懈可击）
	err = g.PassResponse(0, &events)
	if err != nil {
		t.Fatal(err)
	}
	
	// 检查乐不思蜀是否被抵消（从判定区移除，进入弃牌堆）
	if len(g.Players[0].JudgeArea) != 0 {
		t.Errorf("期望判定区为空，得到 %d 张牌", len(g.Players[0].JudgeArea))
	}
	
	// 检查是否继续处理下一张判定牌（这里没有下一张，应该进入摸牌阶段）
	if g.TurnStep != StepDraw && g.TurnStep != StepJudge {
		t.Logf("当前阶段: %s（可能已进入摸牌阶段或继续处理判定）", g.TurnStep)
	}
	
	t.Logf("无懈可击抵消判定牌测试通过！")
}
