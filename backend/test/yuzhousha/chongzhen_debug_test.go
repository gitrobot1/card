package engine_test

import (
	"fmt"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// TestDebug_ChongzhenTakeWindow SP赵云龙胆变牌杀→冲阵选牌完整流程。
// SP赵云：手牌 1闪（当杀用），触发冲阵 → TakeWindow 拿对手一张牌。
func TestDebug_ChongzhenTakeWindow(t *testing.T) {
	fmt.Println("=== SP赵云龙胆变牌→冲阵选牌 ===")

	g, err := engine.NewSolo1v1("debug-cz", "玩家", skill.CharSpZhaoYun, skill.CharSimaYi)
	if err != nil {
		t.Fatal(err)
	}

	// 设置出牌阶段
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	// SP赵云手牌：1闪（用于龙胆变杀）
	g.Players[0].Hand = []engine.Card{
		{ID: "shan-1", Kind: engine.CardShan, Suit: "D", Name: "闪"},
	}
	// 司马懿手牌：1桃（冲阵目标）+ 1闪（用于响应杀）
	g.Players[1].Hand = []engine.Card{
		{ID: "tao-1", Kind: engine.CardTao, Suit: "H", Name: "桃"},
		{ID: "shan-2", Kind: engine.CardShan, Suit: "D", Name: "闪"},
	}
	g.SyncCounts()

	fmt.Printf("初始：SP赵云 手牌=%d, 司马懿 手牌=%d\n",
		g.Players[0].HandCount, g.Players[1].HandCount)

	var events []engine.GameEvent

	// Step 1: SP赵云使用闪（龙胆当杀）
	fmt.Println("\n--- Step 1: SP赵云使用闪（龙胆当杀） ---")
	if err := g.PlayCard(0, "shan-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	fmt.Printf("出杀后: Phase=%s ResponseMode=%q\n", g.Phase, g.Pending.ResponseMode)

	// 应该进入冲阵选牌窗口
	if g.Pending == nil {
		t.Fatal("出杀后 Pending 为 nil")
	}
	fmt.Printf("Pending: ResponseMode=%q WindowKind=%q ChongzhenDone=%v OriginalKind=%q\n",
		g.Pending.ResponseMode, g.Pending.WindowKind, g.Pending.ChongzhenDone, g.Pending.OriginalKind)

	if g.Pending.ResponseMode != engine.ResponseModeSkillChongzhen {
		t.Fatalf("期望冲阵窗口, 实际 ResponseMode=%q", g.Pending.ResponseMode)
	}
	fmt.Println("✅ 进入冲阵选牌窗口（复用手顺手牵羊的 TakeWindow）")

	// Step 2: SP赵云选司马懿的第一张手牌
	fmt.Println("\n--- Step 2: SP赵云选牌 ---")
	if err := g.TakeOne(0, engine.ZoneHand, "tao-1", &events); err != nil {
		t.Fatalf("选牌失败: %v", err)
	}

	fmt.Printf("选牌后: SP赵云 手牌=%d(含桃=%v), 司马懿 手牌=%d\n",
		g.Players[0].HandCount,
		func() bool {
			for _, c := range g.Players[0].Hand {
				if c.ID == "tao-1" {
					return true
				}
			}
			return false
		}(),
		g.Players[1].HandCount,
	)

	// Step 3: 冲阵完成 → 回到杀响应窗口
	fmt.Println("\n--- Step 3: 冲阵完成 → 司马懿出闪 ---")
	fmt.Printf("冲阵后: Phase=%s ResponseMode=%q RequiredKind=%q\n",
		g.Phase, g.Pending.ResponseMode, g.Pending.RequiredKind)

	// 司马懿出闪
	if g.Pending == nil {
		t.Fatal("冲阵完成后 Pending 为 nil")
	}
	if err := g.RespondCard(1, "shan-2", &events); err != nil {
		t.Fatalf("出闪失败: %v", err)
	}
	fmt.Println("✅ 司马懿出闪成功")

	// 验证冲阵拿到了桃
	hasTao := false
	for _, c := range g.Players[0].Hand {
		if c.ID == "tao-1" {
			hasTao = true
			break
		}
	}
	if !hasTao {
		t.Fatal("冲阵未成功获得司马懿的桃")
	}
	fmt.Println("✅ 冲阵成功获得对方手牌")

	fmt.Println("\n=== 流程结束 ===")
	fmt.Printf("最终: SP赵云 手牌=%d, 司马懿 手牌=%d\n",
		g.Players[0].HandCount, g.Players[1].HandCount)
}
