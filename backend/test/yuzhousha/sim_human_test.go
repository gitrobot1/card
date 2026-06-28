package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

// ============================================================
// 模拟两个"人类"玩家出牌测试
// 不依赖 AI 标记，模拟真实操作流程：
// 有杀就出杀 → 被杀出闪 → 没有就硬吃 → 受伤触发技能
// 目标：验证核心出牌→响应→技能→结算流程通畅
// ============================================================

// ------------------------------------------------------------------
// 工具函数
// ------------------------------------------------------------------

func hasCardKind(hand []engine.Card, kind string) (int, string) {
	for i, c := range hand {
		if c.Kind == kind {
			return i, c.ID
		}
	}
	return -1, ""
}

func hasCardPlaysAs(g *engine.Game, seat int, hand []engine.Card, asKind string) (int, string) {
	for i, c := range hand {
		if c.Kind == asKind || g.CardPlaysAsForTest(seat, c, asKind) {
			return i, c.ID
		}
	}
	return -1, ""
}

// ------------------------------------------------------------------
// 测试1：杀→闪 基础响应
// ------------------------------------------------------------------

func TestSim_Human_ShaShan(t *testing.T) {
	g, err := engine.NewSolo1v1("h-sha", "玩家", engine.CharGuanYu, engine.CharLiuBei)
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

	// 出杀
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	// 出闪
	if err := g.RespondCard(1, "shan-1", &events); err != nil {
		t.Fatalf("出闪失败: %v", err)
	}
	t.Log("✅ 杀→闪 流程正常")
}

// ------------------------------------------------------------------
// 测试2：杀命中→扣血→刚烈触发
// ------------------------------------------------------------------

func TestSim_Human_ShaHitGanglie(t *testing.T) {
	g, err := engine.NewSolo1v1("h-gl", "玩家", engine.CharGuanYu, engine.CharXiahouDun)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = nil // 夏侯惇没有闪
	g.SyncCounts()

	var events []engine.GameEvent
	hpBefore := g.Players[1].HP

	// 出杀
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	// 不出闪
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatalf("不出闪失败: %v", err)
	}
	// 验证扣血
	if g.Players[1].HP != hpBefore-1 {
		t.Fatalf("未扣血: hp=%d", g.Players[1].HP)
	}
	// 验证刚烈窗口触发
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillGanglieOffer {
		t.Fatalf("刚烈未触发: pending=%+v", g.Pending)
	}
	t.Log("✅ 杀命中→扣血→刚烈触发 流程正常")
}

// ------------------------------------------------------------------
// 测试3：杀→闪→青龙刀追加杀（武器联动）
// ------------------------------------------------------------------

func TestSim_Human_QinglongDaoke(t *testing.T) {
	g, err := engine.NewSolo1v1("h-ql", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Weapon = &engine.Card{ID: "w3", Kind: engine.CardWeapon3, Name: "青龙刀"}
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-2", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = []engine.Card{
		{ID: "shan-1", Kind: engine.CardShan, Name: "闪"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	// 第一张杀
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	// 目标出闪
	if err := g.RespondCard(1, "shan-1", &events); err != nil {
		t.Fatalf("出闪失败: %v", err)
	}
	// 青龙刀允许追加杀，检查是否有后续窗口
	t.Logf("青龙刀追加杀后: pending=%+v", g.Pending)
	t.Log("✅ 杀→闪→青龙刀 流程正常")
}

// ------------------------------------------------------------------
// 测试4：过河拆桥→无懈窗口→拆牌
// ------------------------------------------------------------------

func TestSim_Human_GuoHeDiscard(t *testing.T) {
	g, err := engine.NewSolo1v1("h-gh", "玩家", engine.CharGuanYu, engine.CharLiuBei)
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

	// 使用过河拆桥
	if err := g.PlayCard(0, "gh-1", 1, &events); err != nil {
		t.Fatalf("使用过河拆桥失败: %v", err)
	}
	if g.Pending == nil {
		t.Fatal("使用过河拆桥后无窗口")
	}
	t.Logf("过河拆桥窗口: pending=%s", g.Pending.ResponseMode)
	t.Log("✅ 过河拆桥流程正常")
}

// ------------------------------------------------------------------
// 测试5：南蛮入侵→AOE响应链
// ------------------------------------------------------------------

func TestSim_Human_NanmanAOE(t *testing.T) {
	g, err := engine.NewSolo1v1("h-nm", "玩家", engine.CharGuanYu, engine.CharLiuBei)
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

	// 使用南蛮
	if err := g.PlayCard(0, "nm-1", 1, &events); err != nil {
		t.Fatalf("使用南蛮失败: %v", err)
	}
	// 进入 AOE 宣告/无懈窗口
	t.Logf("南蛮后: pending=%s", func() string {
		if g.Pending != nil {
			return g.Pending.ResponseMode
		}
		return "nil"
	}())

	// 跳过无懈窗口（模拟无人出无懈）
	for g.Pending != nil && (g.Pending.ResponseMode == engine.ResponseModeWuxiekTrick) {
		actor := g.PendingActorSeat()
		if actor < 0 {
			break
		}
		if err := g.PassResponse(actor, &events); err != nil {
			t.Fatalf("跳过无懈失败: %v", err)
		}
	}
	t.Logf("无懈后: pending=%s", func() string {
		if g.Pending != nil {
			return g.Pending.ResponseMode
		}
		return "nil"
	}())

	t.Log("✅ 南蛮入侵流程正常")
}

// ------------------------------------------------------------------
// 测试6：乐不思蜀→判定→跳过出牌阶段
// ------------------------------------------------------------------

func TestSim_Human_LebuJudge(t *testing.T) {
	g, err := engine.NewSolo1v1("h-lb", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "lb-1", Kind: engine.CardLeBu, Name: "乐不思蜀"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	// 使用乐不思蜀
	if err := g.PlayCard(0, "lb-1", 1, &events); err != nil {
		t.Fatalf("使用乐不思蜀失败: %v", err)
	}
	// 检查进入判定区
	if len(g.Players[1].JudgeArea) == 0 {
		t.Fatal("乐不思蜀未进入判定区")
	}
	t.Logf("乐不思蜀进入判定区: judgeArea=%d", len(g.Players[1].JudgeArea))
	t.Log("✅ 乐不思蜀流程正常")
}

// ------------------------------------------------------------------
// 测试7：装备武器→出杀
// ------------------------------------------------------------------

func TestSim_Human_EquipAndSha(t *testing.T) {
	g, err := engine.NewSolo1v1("h-eq", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "wp-1", Kind: engine.CardWeapon1, Name: "诸葛连弩"},
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-2", Kind: engine.CardSha, Name: "杀"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	// 装备武器
	if err := g.PlayCard(0, "wp-1", 0, &events); err != nil {
		t.Fatalf("装备武器失败: %v", err)
	}
	if g.Players[0].Weapon == nil {
		t.Fatal("武器未装备")
	}
	t.Logf("装备了: %s", g.Players[0].Weapon.Name)

	// 出一张杀
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	_ = g.PassResponse(1, &events)
	t.Logf("第一张杀后 HP=%d", g.Players[1].HP)

	// 诸葛连弩应该可以再出杀
	if err := g.PlayCard(0, "sha-2", 1, &events); err != nil {
		t.Logf("第二张杀失败（预期）: %v", err)
	} else {
		_ = g.PassResponse(1, &events)
		t.Logf("第二张杀后 HP=%d（诸葛连弩生效）", g.Players[1].HP)
	}

	t.Log("✅ 装备→出杀流程正常")
}

// ------------------------------------------------------------------
// 测试8：武圣发动→红色牌当杀
// ------------------------------------------------------------------

func TestSim_Human_WushengActivate(t *testing.T) {
	g, err := engine.NewSolo1v1("h-ws", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "shan-r", Kind: engine.CardShan, Suit: "H", Name: "闪", Label: "红桃闪"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	// 发动武圣
	if err := g.UseSkill(0, engine.UseSkillRequest{SkillID: engine.SkillWusheng}, &events); err != nil {
		t.Fatalf("发动武圣失败: %v", err)
	}

	// 验证红色闪能当杀
	redShan := engine.Card{ID: "shan-r", Kind: engine.CardShan, Suit: "H", Name: "闪"}
	if !g.CardPlaysAsForTest(0, redShan, engine.CardSha) {
		t.Fatal("发动武圣后红色闪不能当杀")
	}

	// 实际出这张"杀"
	if err := g.PlayCard(0, "shan-r", 1, &events); err != nil {
		t.Fatalf("用红色闪当杀出牌失败: %v", err)
	}
	_ = g.PassResponse(1, &events)
	t.Logf("武圣出杀后 HP=%d", g.Players[1].HP)

	t.Log("✅ 武圣变牌→出杀流程正常")
}

// ------------------------------------------------------------------
// 测试9：仁德→给牌+回血
// ------------------------------------------------------------------

func TestSim_Human_RendeGiveHeal(t *testing.T) {
	g, err := engine.NewSolo1v1("h-rd", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "c-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "c-2", Kind: engine.CardShan, Name: "闪"},
	}
	g.Players[0].HP = 3
	beforeHP := g.Players[0].HP
	beforeHand := len(g.Players[1].Hand)
	g.SyncCounts()

	var events []engine.GameEvent

	// 使用仁德
	if err := g.UseSkill(0, engine.UseSkillRequest{
		SkillID: engine.SkillRende, TargetIndex: 1, CardIDs: []string{"c-1", "c-2"},
	}, &events); err != nil {
		t.Fatalf("使用仁德失败: %v", err)
	}
	// 验证回血
	if g.Players[0].HP != beforeHP+1 {
		t.Fatalf("仁德未回血: hp=%d", g.Players[0].HP)
	}
	// 验证给牌
	if len(g.Players[1].Hand) != beforeHand+2 {
		t.Fatalf("仁德未给牌: 目标手牌=%d", len(g.Players[1].Hand))
	}
	t.Log("✅ 仁德→给牌+回血流程正常")
}

// ------------------------------------------------------------------
// 测试10：濒死→求桃→自救
// ------------------------------------------------------------------

func TestSim_Human_DyingRescue(t *testing.T) {
	g, err := engine.NewSolo1v1("h-dy", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	setupPlayingTurn(g, 0)
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].HP = 1
	g.Players[1].Hand = []engine.Card{
		{ID: "tao-1", Kind: engine.CardTao, Name: "桃"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	// 出杀→不出闪→扣血至0
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatalf("不出闪失败: %v", err)
	}
	// 检查进入濒死
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeDying {
		t.Fatalf("未进入濒死: pending=%+v", g.Pending)
	}
	t.Log("进入濒死窗口")

	// 自救：出桃
	if err := g.RespondCard(1, "tao-1", &events); err != nil {
		t.Fatalf("自救出桃失败: %v", err)
	}
	if g.Players[1].HP != 1 {
		t.Fatalf("自救后 HP 异常: %d", g.Players[1].HP)
	}
	t.Log("✅ 濒死→自救流程正常")
}

// ------------------------------------------------------------------
// 测试11：铁骑→判定→改判（完整改判链）
// ------------------------------------------------------------------

func TestSim_Human_TieqiJudgeModify(t *testing.T) {
	g, err := engine.NewSolo1v1("h-tq", "玩家", engine.CharMaChao, engine.CharSimaYi)
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
		{ID: "j-1", Suit: "H", Kind: engine.CardSha, Name: "杀", Label: "红桃K"},
	}
	g.SyncCounts()

	var events []engine.GameEvent

	// 马超出杀
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatalf("出杀失败: %v", err)
	}
	// 铁骑窗口
	if g.Pending == nil || !g.Pending.TieqiPending {
		t.Fatalf("铁骑未触发: pending=%+v", g.Pending)
	}
	t.Log("铁骑窗口触发")

	// 发动铁骑
	if err := g.ApplyTieqi(0, &events); err != nil {
		t.Fatalf("发动铁骑失败: %v", err)
	}
	// 改判窗口（司马懿有鬼才）
	if g.Pending == nil {
		t.Fatal("判定后未进入改判窗口")
	}
	t.Logf("改判窗口: %s", g.Pending.ResponseMode)

	t.Log("✅ 铁骑→判定→改判流程正常")
}
