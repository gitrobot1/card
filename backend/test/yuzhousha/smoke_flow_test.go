package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

// ============================================================
// 流程冒烟测试：快速验证核心流程的完整性
// 不测"对不对"，只测"通不通"
// 目标：30 秒内跑完，覆盖所有主要流程节点
// ============================================================

// ------------------------------------------------------------------
// 1. 基础出牌流程
// ------------------------------------------------------------------

func TestFlow_PlayShaAndShan(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-sha", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = []engine.Card{
		{ID: "shan-1", Kind: engine.CardShan, Name: "闪"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	if g.Pending == nil {
		t.Fatal("出杀后没有进入响应窗口")
	}
	if err := g.RespondCard(1, "shan-1", &events); err != nil {
		t.Fatalf("出闪失败: %v", err)
	}
	t.Log("✅ 杀→闪流程通过")
}

// ------------------------------------------------------------------
// 2. 锦囊牌使用流程
// ------------------------------------------------------------------

func TestFlow_PlayGuoHe(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-gh", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "gh-1", Kind: engine.CardGuoHe, Name: "过河拆桥"},
	}
	g.Players[1].Hand = []engine.Card{
		{ID: "h-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.PlayCard(0, "gh-1", 1, &events); err != nil {
		t.Fatalf("使用过河拆桥失败: %v", err)
	}
	if g.Pending == nil {
		t.Fatal("使用过河拆桥后没有进入任何窗口")
	}
	t.Logf("✅ 过河拆桥流程通过 (pending=%s)", g.Pending.ResponseMode)
}

// ------------------------------------------------------------------
// 3. AOE 锦囊流程
// ------------------------------------------------------------------

func TestFlow_PlayNanman(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-nm", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "nm-1", Kind: engine.CardNanMan, Name: "南蛮入侵"},
	}
	g.Players[1].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.PlayCard(0, "nm-1", 1, &events); err != nil {
		t.Fatalf("使用南蛮入侵失败: %v", err)
	}
	t.Logf("✅ 南蛮入侵流程通过 (pending=%s)", func() string {
		if g.Pending != nil {
			return g.Pending.ResponseMode
		}
		return "nil"
	}())
}

// ------------------------------------------------------------------
// 4. 延时锦囊流程
// ------------------------------------------------------------------

func TestFlow_PlayLebu(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-lb", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "lb-1", Kind: engine.CardLeBu, Name: "乐不思蜀"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.PlayCard(0, "lb-1", 1, &events); err != nil {
		t.Fatalf("使用乐不思蜀失败: %v", err)
	}
	if len(g.Players[1].JudgeArea) == 0 {
		t.Fatal("乐不思蜀未进入判定区")
	}
	t.Log("✅ 乐不思蜀流程通过")
}

// ------------------------------------------------------------------
// 5. 装备武器流程
// ------------------------------------------------------------------

func TestFlow_EquipWeapon(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-wp", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "wp-1", Kind: engine.CardWeapon1, Name: "诸葛连弩"},
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.PlayCard(0, "wp-1", 0, &events); err != nil {
		t.Fatalf("装备武器失败: %v", err)
	}
	if g.Players[0].Weapon == nil {
		t.Fatal("装备武器后 Weapon 为 nil")
	}
	t.Log("✅ 装备武器流程通过")
}

// ------------------------------------------------------------------
// 6. 卖血技能流程
// ------------------------------------------------------------------

func TestFlow_GanglieTrigger(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-gl", "玩家", engine.CharGuanYu, engine.CharXiahouDun)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatalf("不出闪失败: %v", err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillGanglieOffer {
		t.Fatalf("受伤后未触发刚烈窗口, pending=%+v", g.Pending)
	}
	t.Log("✅ 刚烈触发流程通过")
}

// ------------------------------------------------------------------
// 7. 改判技能流程（铁骑→改判）
// ------------------------------------------------------------------

func TestFlow_TieqiThenGuicai(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-gc", "玩家", engine.CharMaChao, engine.CharSimaYi)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = []engine.Card{
		{ID: "h-1", Kind: engine.CardSha, Name: "杀", Label: "改判用"},
	}
	g.DrawPile = []engine.Card{
		{ID: "j-1", Suit: "S", Kind: engine.CardSha, Name: "杀", Label: "黑桃"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	if g.Pending == nil || !g.Pending.TieqiPending {
		t.Fatalf("出杀后未进入铁骑窗口, pending=%+v", g.Pending)
	}
	if err := g.ApplyTieqi(0, &events); err != nil {
		t.Fatalf("发动铁骑失败: %v", err)
	}
	if g.Pending == nil {
		t.Fatal("判定后未进入改判窗口")
	}
	t.Logf("✅ 铁骑→改判流程通过 (pending=%s)", g.Pending.ResponseMode)
}

// ------------------------------------------------------------------
// 8. 濒死求桃流程
// ------------------------------------------------------------------

func TestFlow_DyingRescue(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-dy", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].HP = 1
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatalf("不出闪失败: %v", err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeDying {
		t.Fatalf("扣血至0后未进入濒死窗口, pending=%+v", g.Pending)
	}
	t.Log("✅ 濒死流程通过")
}

// ------------------------------------------------------------------
// 9. 主动技能流程
// ------------------------------------------------------------------

func TestFlow_RendeActivate(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-rd", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "c-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "c-2", Kind: engine.CardShan, Name: "闪"},
	}
	g.Players[0].HP = 3
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.UseSkill(0, engine.UseSkillRequest{
		SkillID: engine.SkillRende, TargetIndex: 1, CardIDs: []string{"c-1", "c-2"},
	}, &events); err != nil {
		t.Fatalf("使用仁德失败: %v", err)
	}
	if g.Players[0].HP != 4 {
		t.Fatalf("仁德后未回血, hp=%d", g.Players[0].HP)
	}
	t.Log("✅ 仁德技能流程通过")
}

// ------------------------------------------------------------------
// 10. 牌当牌技能流程（武圣）
// ------------------------------------------------------------------

func TestFlow_WushengRedAsSha(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-ws", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "shan-r", Kind: engine.CardShan, Suit: "H", Name: "闪", Label: "红桃闪"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	if err := g.UseSkill(0, engine.UseSkillRequest{SkillID: engine.SkillWusheng}, &events); err != nil {
		t.Fatalf("发动武圣失败: %v", err)
	}
	redShan := engine.Card{ID: "shan-r", Kind: engine.CardShan, Suit: "H", Name: "闪"}
	if !g.CardPlaysAsForTest(0, redShan, engine.CardSha) {
		t.Fatal("发动武圣后红色闪不能当杀")
	}
	t.Log("✅ 武圣变牌流程通过")
}

// ------------------------------------------------------------------
// 11. 综合流程：完整一回合（含准备→出牌→弃牌→结束）
// 注：此测试依赖引擎内部阶段流转细节，如果失败检查阶段流转是否正常
// ------------------------------------------------------------------

func TestFlow_FullTurn(t *testing.T) {
	g, err := engine.NewSolo1v1("flow-turn", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-2", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-3", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-4", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-5", Kind: engine.CardSha, Name: "杀"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	// 设置到出牌阶段
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	t.Logf("初始: phase=%s step=%s", g.Phase, g.TurnStep)

	// 出一张杀
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatalf("不出闪失败: %v", err)
	}

	// 结束出牌
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatalf("EndPlay 失败: %v", err)
	}
	t.Logf("结束出牌后: phase=%s step=%s", g.Phase, g.TurnStep)

	t.Log("✅ 出牌→结束流程通过")
}
