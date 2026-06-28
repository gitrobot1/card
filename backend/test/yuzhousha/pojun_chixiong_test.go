package engine_test

import (
	"fmt"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

// TestDebug_PojunChixiong 界徐盛装备雌雄双股剑杀女角色。
// 验证：雌雄 → 破军的顺序正确，不互相覆盖。
func TestDebug_PojunChixiong(t *testing.T) {
	fmt.Println("=== 界徐盛装备雌雄双股剑 → 杀女角色 ===")

	g, err := engine.NewSolo1v1("debug-pc", "玩家", "jie_xu_sheng", "zhen_ji")
	if err != nil {
		t.Fatal(err)
	}

	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	// 界徐盛手牌：1杀 + 雌雄双股剑
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Suit: "S", Name: "杀"},
		{ID: "chixiong", Kind: engine.CardWeapon8, Name: "雌雄双股剑"},
	}
	// 甄姬手牌：1闪 + 1桃
	g.Players[1].Hand = []engine.Card{
		{ID: "shan-1", Kind: engine.CardShan, Suit: "D", Name: "闪"},
		{ID: "tao-1", Kind: engine.CardTao, Suit: "H", Name: "桃"},
	}
	g.Players[1].HP = 3
	g.SyncCounts()

	var events []engine.GameEvent

	// 装备雌雄双股剑
	fmt.Println("\n--- Step 0: 装备雌雄双股剑 ---")
	if err := g.PlayCardWithTarget(0, "chixiong", engine.PlayTarget{}, &events); err != nil {
		t.Fatalf("装备失败: %v", err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	// 出杀
	fmt.Println("\n--- Step 1: 界徐盛出杀 ---")
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}

	if g.Pending == nil {
		t.Fatal("出杀后 Pending 为 nil")
	}
	fmt.Printf("出杀后: ResponseMode=%q WindowKind=%q PojunMax=%d chixiong_done=%v\n",
		g.Pending.ResponseMode, g.Pending.WindowKind, g.Pending.PojunMax, g.Pending.Extra["chixiong_done"])

	// AI 目标：雌雄自动弃牌（甄姬弃最后一张手牌），不弹窗
	// 然后破军打开选牌窗口
	if g.Pending.ResponseMode == engine.ResponseModeWeapon8 {
		// 人类目标路径：雌雄弹窗
		fmt.Println("  [人类目标] 雌雄弹窗 → 跳过")
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatalf("跳过雌雄失败: %v", err)
		}
		fmt.Printf("  雌雄后: ResponseMode=%q\n", g.Pending.ResponseMode)
	}

	// 破军应该打开选牌窗口
	if g.Pending.ResponseMode != engine.ResponseModeSkillPojun {
		t.Fatalf("期望破军窗口, 实际 ResponseMode=%q", g.Pending.ResponseMode)
	}
	fmt.Println("✅ 破军选牌窗口（雌雄已完成，未覆盖破军）")

	// 破军跳过（不实际拿牌，验证窗口正确打开即可）
	if err := g.PassPojun(0, &events); err != nil {
		t.Fatalf("跳过破军失败: %v", err)
	}

	if g.Pending.RequiredKind != engine.CardShan {
		t.Fatalf("期望等待出闪, 实际 RequiredKind=%q", g.Pending.RequiredKind)
	}
	fmt.Printf("✅ 破军完成 → 等待出闪 (RequiredKind=%q)\n", g.Pending.RequiredKind)

	// 甄姬出闪
	if err := g.RespondCard(1, "shan-1", &events); err != nil {
		t.Fatalf("出闪失败: %v", err)
	}
	fmt.Println("✅ 甄姬出闪，杀被抵消")

	fmt.Println("\n=== 流程验证通过: 雌雄 → 破军 → 目标出闪 ===")
}
