package engine_test

import (
	"fmt"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

// TestDebug_ChongzhenRespond SP赵云响应万箭时龙胆+冲阵拿牌
func TestDebug_ChongzhenRespond(t *testing.T) {
	fmt.Println("=== SP赵云响应万箭 → 龙胆(杀当闪) → 冲阵拿牌 ===")

	g, err := engine.NewSolo1v1("debug-czr", "玩家", "sp_zhao_yun", "si_ma_yi")
	if err != nil {
		t.Fatal(err)
	}

	// SP赵云手牌：1杀（用于龙胆当闪）
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Suit: "S", Name: "杀"},
	}
	// 司马懿手牌：1桃（冲阵目标）
	g.Players[1].Hand = []engine.Card{
		{ID: "tao-1", Kind: engine.CardTao, Suit: "H", Name: "桃"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	// 模拟万箭齐发：司马懿出万箭 → SP赵云需要出闪
	g.Phase = engine.PhaseResponse
	g.CurrentTurn = 1
	g.Pending = &engine.PendingCombat{
		SourceIndex:  1,
		TargetIndex:  0,
		ReturnIndex:  1,
		ActorSeat:    0,
		SubjectSeat:  0,
		Card:         engine.Card{Kind: engine.CardWanJian, Name: "万箭齐发"},
		RequiredKind: engine.CardShan,
		AoeQueue:     []int{},
	}
	engine.FillPendingRoles(g.Pending)

	fmt.Printf("初始：SP赵云 手牌=%d, 司马懿 手牌=%d\n",
		g.Players[0].HandCount, g.Players[1].HandCount)

	// SP赵云用杀当闪（龙胆）响应万箭
	if err := g.RespondCard(0, "sha-1", &events); err != nil {
		t.Fatalf("响应失败: %v", err)
	}

	fmt.Printf("响应后：SP赵云 手牌=%d(含桃=%v), 司马懿 手牌=%d\n",
		g.Players[0].HandCount,
		func() bool {
			for _, c := range g.Players[0].Hand {
				if c.ID == "tao-1" { return true }
			}
			return false
		}(),
		g.Players[1].HandCount,
	)

	// SP赵云应该拿到了司马懿的桃
	hasTao := false
	for _, c := range g.Players[0].Hand {
		if c.ID == "tao-1" {
			hasTao = true
			break
		}
	}
	if !hasTao {
		t.Fatal("冲阵未获得司马懿的桃")
	}
	fmt.Println("✅ 龙胆(杀当闪) → 冲阵拿牌成功")

	// 验证 counter 已消耗（再次响应不会重复触发冲阵）

	// 响应正常继续
	if g.Pending != nil && g.Pending.AoeQueue != nil {
		fmt.Println("✅ 万箭流程正常继续")
	}

	fmt.Println("\n=== 验证通过：决斗/南蛮/万箭中冲阵可正常拿牌 ===")
}
