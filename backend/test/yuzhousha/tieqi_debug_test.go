package engine_test

import (
	"fmt"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

// TestDebug_MaChaoShaSimaYi 马超杀司马懿完整流程调试。
// 马超：手牌 1杀+1桃，发动铁骑
// 司马懿：1血，手牌 1黑桃2酒
func TestDebug_MaChaoShaSimaYi(t *testing.T) {
	fmt.Println("=== 马超杀司马懿 铁骑完整流程 ===")

	g, err := engine.NewSolo1v1("debug-mcsy", "玩家", engine.CharMaChao, engine.CharSimaYi)
	if err != nil {
		t.Fatal(err)
	}

	// 设置出牌阶段
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	// 马超手牌：1杀 + 1桃
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Suit: "S", Name: "杀"},
		{ID: "tao-1", Kind: engine.CardTao, Suit: "H", Name: "桃"},
	}
	// 司马懿：1血，手牌 1黑桃2酒
	g.Players[1].HP = 1
	g.Players[1].Hand = []engine.Card{
		{ID: "jiu-1", Kind: engine.CardJiu, Suit: "S", Name: "酒", Rank: 2},
	}
	// 牌堆：判定牌（黑桃，非红色 → 铁骑成功 → 不可出闪）
	g.DrawPile = []engine.Card{
		{ID: "judge-1", Suit: "S", Kind: engine.CardSha, Name: "杀", Label: "黑桃判定"},
	}
	g.SyncCounts()

	fmt.Printf("初始状态：马超 HP=%d 手牌=%d, 司马懿 HP=%d 手牌=%d\n",
		g.Players[0].HP, g.Players[0].HandCount,
		g.Players[1].HP, g.Players[1].HandCount)

	var events []engine.GameEvent

	// Step 1: 马超出杀
	fmt.Println("\n--- Step 1: 马超出杀 ---")
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	fmt.Printf("出杀后: Phase=%s Step=%s Pending=%+v\n", g.Phase, g.TurnStep, g.Pending)
	if g.Pending == nil {
		t.Fatal("出杀后 Pending 为 nil")
	}

	// 检查 TieqiPending
	if !g.Pending.TieqiPending {
		t.Fatalf("出杀后 TieqiPending 应为 true, 实际 %v", g.Pending.TieqiPending)
	}
	fmt.Println("✅ TieqiPending = true（OnUseCardToTarget Hook 已设置）")

	// Step 2: 马超发动铁骑 → 开始判定
	fmt.Println("\n--- Step 2: 马超发动铁骑 ---")
	if err := g.ApplyTieqi(0, &events); err != nil {
		t.Fatalf("发动铁骑失败: %v", err)
	}
	fmt.Printf("发动铁骑后: Phase=%s Pending.ResponseMode=%s JudgeCard=%v\n",
		g.Phase, g.Pending.ResponseMode, g.Pending.JudgeCard.Label)

	// 应该进入改判窗口（司马懿有鬼才）
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillGuicai {
		t.Fatalf("判定后应进入鬼才改判窗口, 实际 Pending.ResponseMode=%s", g.Pending.ResponseMode)
	}
	fmt.Println("✅ 进入鬼才改判窗口")

	// 司马懿有鬼才技能，可以改判或跳过
	// 但司马懿手牌是酒（黑桃2），不是手牌改判
	// 直接用 PassGuicai 跳过改判
	fmt.Println("\n--- Step 3: 司马懿不发动鬼才 ---")
	if err := g.PassGuicai(1, &events); err != nil {
		t.Fatalf("跳过鬼才失败: %v", err)
	}

	// 判定完成 → applyTieqiJudgeResult 执行
	// 黑桃判定 → 非红色 → ShaUnblockable=true → 目标不能出闪
	fmt.Printf("判定完成后: Phase=%s Pending.TieqiPending=%v Pending.ShaUnblockable=%v Pending.RequiredKind=%s\n",
		g.Phase, g.Pending.TieqiPending, g.Pending.ShaUnblockable, g.Pending.RequiredKind)

	if !g.Pending.ShaUnblockable {
		t.Fatalf("黑桃判定后 ShaUnblockable 应为 true")
	}
	if g.Pending.RequiredKind != "" {
		t.Fatalf("黑桃判定后 RequiredKind 应为空（不可出闪）, 实际 %q", g.Pending.RequiredKind)
	}
	fmt.Println("✅ 铁骑判定成功（黑桃），目标不可出闪")

	// 查看当前 Pending 的完整状态
	fmt.Printf("DEBUG 出闪前: Phase=%s TurnStep=%s Pending.ResponseMode=%q Pending.TieqiPending=%v Pending.ShaUnblockable=%v Pending.RequiredKind=%q\n",
		g.Phase, g.TurnStep, g.Pending.ResponseMode, g.Pending.TieqiPending, g.Pending.ShaUnblockable, g.Pending.RequiredKind)

	// Step 4: 司马懿被强制跳过出闪（杀不可闪避）
	fmt.Println("\n--- Step 4: 司马懿不出闪（被铁骑封锁） ---")
	err = g.PassResponse(1, &events)
	fmt.Printf("PassResponse 返回: err=%v\n", err)
	if err != nil {
		t.Fatalf("不出闪失败: %v", err)
	}

	fmt.Printf("PassResponse后: Phase=%s TurnStep=%s Pending=%v\n", g.Phase, g.TurnStep, g.Pending != nil)
	if g.Pending != nil {
		fmt.Printf("  Pending.ResponseMode=%q Pending.RequiredKind=%q\n", g.Pending.ResponseMode, g.Pending.RequiredKind)
	}
	fmt.Printf("扣血后: 司马懿 HP=%d\n", g.Players[1].HP)

	// 检查是否进入反馈窗口
	if g.Pending != nil {
		fmt.Printf("当前窗口: ResponseMode=%s\n", g.Pending.ResponseMode)
	}

	fmt.Println("\n=== 流程结束 ===")
	fmt.Printf("最终状态: 司马懿 HP=%d 手牌=%d, 马超 HP=%d 手牌=%d\n",
		g.Players[1].HP, g.Players[1].HandCount,
		g.Players[0].HP, g.Players[0].HandCount)
}
