// 复杂结算场景示例：特定手牌/装备/牌堆 + 逐步推进 + 断言 Pending 中间态。
// 新技能或改伤害链时，可复制本文件中的模式编写用例。
package engine_test

import (
	"errors"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func setupPlayingTurn(g *engine.Game, turn int) {
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = turn
	g.SyncCounts()
}

func assertPendingMode(t *testing.T, g *engine.Game, mode string) {
	t.Helper()
	if g.Pending == nil || g.Pending.ResponseMode != mode {
		t.Fatalf("expected pending mode %q, got %+v", mode, g.Pending)
	}
}

// 酒+杀造成 2 点伤害 → 刚烈按伤害值排队两次（奸雄→刚烈×N 链）。
func TestScenario_JiuShaOffersGanglieTwice(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-ganglie2", "玩家", engine.CharLiuBei, engine.CharXiahouDun)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "jiu-1", Kind: engine.CardJiu, Name: "酒"},
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = nil
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlayCard(0, "jiu-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-2 {
		t.Fatalf("expected 2 damage from jiu sha, hp=%d", g.Players[1].HP)
	}

	assertPendingMode(t, g, engine.ResponseModeSkillGanglieOffer)
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeSkillGanglieOffer)
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhasePlaying || g.CurrentTurn != 0 {
		t.Fatalf("expected source to resume play, phase=%s turn=%d", g.Phase, g.CurrentTurn)
	}
}

// 杀命中 → 反馈拿牌 → 伤害链结束后 → 麒麟弓弃马（装备结算顺序）。
func TestScenario_FankuiThenQilinBow(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-fk-qilin", "玩家", engine.CharLiuBei, engine.CharSimaYi)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "extra", Kind: engine.CardShan, Name: "闪", Label: "留牌"},
	}
	g.Players[0].Weapon = &engine.Card{ID: "w5", Kind: engine.CardWeapon5, Name: "麒麟弓"}
	g.Players[1].Hand = []engine.Card{{ID: "spare", Kind: engine.CardSha, Name: "杀", Label: "反馈用"}}
	g.Players[1].MinusHorse = &engine.Card{ID: "horse-1", Kind: engine.CardMinusHorse, Name: "-1马"}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}

	assertPendingMode(t, g, engine.ResponseModeSkillFankui)
	if err := g.FankuiTakeFrom(1, "hand", "extra", &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeQilinBow)
	if err := g.QilinDiscardHorseForTest(0, engine.EquipMinusHorse, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].MinusHorse != nil {
		t.Fatal("expected minus horse discarded by qilin")
	}
	foundExtra := false
	for _, c := range g.Players[1].Hand {
		if c.ID == "extra" {
			foundExtra = true
		}
	}
	if !foundExtra {
		t.Fatalf("expected sima yi to gain extra via fankui, hand=%+v", g.Players[1].Hand)
	}
}

// 青釭剑无视八卦 → 命中 → 反馈（装备 + 防具 + 技能链）。
func TestScenario_QingGangIgnoresBaguaThenFankui(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-qg-bg-fk", "玩家", engine.CharLiuBei, engine.CharSimaYi)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "extra", Kind: engine.CardShan, Name: "闪", Label: "留牌"},
	}
	g.Players[0].Weapon = &engine.Card{ID: "w2", Kind: engine.CardWeapon2, Name: "青釭剑"}
	g.Players[1].Hand = []engine.Card{{ID: "spare", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Armor = &engine.Card{ID: "bagua", Kind: engine.CardArmor, Name: "八卦阵"}
	g.DrawPile = []engine.Card{{ID: "j1", Kind: engine.CardSha, Suit: "H", Label: "红桃2", Name: "杀"}}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if !g.Pending.IgnoreArmor {
		t.Fatal("expected qinggang to ignore armor on pending sha")
	}
	if err := g.TryBaguaJudge(1, &events); !errors.Is(err, engine.ErrInvalidCard) {
		t.Fatalf("expected bagua blocked by qinggang, got %v", err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeSkillFankui)
	if err := g.FankuiTakeFrom(1, "hand", "extra", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != g.Players[1].MaxHP-1 {
		t.Fatalf("expected 1 damage through ignored bagua, hp=%d max=%d", g.Players[1].HP, g.Players[1].MaxHP)
	}
}

// 铁骑判黑封锁闪 → 命中 → 目标刚烈判定链。
func TestScenario_TieqiBlackHitThenGanglieJudge(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-tq-gl", "玩家", engine.CharMaChao, engine.CharXiahouDun)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{{ID: "shan-1", Kind: engine.CardShan, Name: "闪"}}
	g.DrawPile = []engine.Card{
		{ID: "tieqi-j", Suit: "S", Kind: engine.CardSha, Name: "杀", Label: "黑桃7"},
		{ID: "ganglie-j", Suit: "S", Rank: 3, Kind: engine.CardSha, Name: "杀", Label: "黑桃3"},
	}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.ApplyTieqi(0, &events); err != nil {
		t.Fatal(err)
	}
	if !g.Pending.ShaUnblockable {
		t.Fatal("expected tieqi black to make sha unblockable")
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}

	assertPendingMode(t, g, engine.ResponseModeSkillGanglieOffer)
	if err := g.UseSkill(1, engine.UseSkillRequest{SkillID: engine.SkillGanglie}, &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeSkillGanglieChoice)
	hpBefore := g.Players[0].HP
	if err := g.UseSkill(0, engine.UseSkillRequest{SkillID: engine.SkillGanglie, TargetZone: "take_damage"}, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[0].HP != hpBefore-1 {
		t.Fatalf("expected source to take ganglie damage, hp=%d", g.Players[0].HP)
	}
}

// 奸雄窗口 → 放弃 → 不拿牌，继续伤害链（此处目标无后续技能则回到出牌）。
func TestScenario_JianxiongPassThenResumePlay(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-jx-pass", "玩家", engine.CharLiuBei, engine.CharCaoCao)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀", Label: "杀"}}
	g.Players[1].Hand = nil
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeSkillJianxiong)
	handBefore := len(g.Players[1].Hand)
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[1].Hand) != handBefore {
		t.Fatalf("expected cao cao not to gain card after passing jianxiong, hand=%+v", g.Players[1].Hand)
	}
	if g.Phase != engine.PhasePlaying || g.CurrentTurn != 0 {
		t.Fatalf("expected play to resume, phase=%s turn=%d", g.Phase, g.CurrentTurn)
	}
}

// 青龙偃月：闪掉第一刀 → 跟进窗口 → 放弃第二刀 → 回到出牌。
func TestScenario_GuanYuFollowPassThenResume(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-gy-pass", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-2", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[0].Weapon = &engine.Card{ID: "w3", Kind: engine.CardWeapon3, Name: "青龙偃月刀"}
	g.Players[1].Hand = []engine.Card{{ID: "shan-1", Kind: engine.CardShan, Name: "闪"}}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.RespondCard(1, "shan-1", &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeGuanYuFollow)
	if err := g.PassResponse(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhasePlaying || g.CurrentTurn != 0 {
		t.Fatalf("expected source to resume after skipping follow-up sha, phase=%s turn=%d", g.Phase, g.CurrentTurn)
	}
	if g.Players[1].HP != g.Players[1].MaxHP {
		t.Fatalf("expected no damage when follow-up skipped, hp=%d", g.Players[1].HP)
	}
}

// 乐不思蜀进判定 → 对手回合开始时无懈 → 乐被拆，仍可正常出牌。
func TestScenario_LebuWuxiekOnJudgeCancelsSkip(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-lb-wx", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "lb-1", Kind: engine.CardLeBu, Name: "乐不思蜀"}}
	g.Players[1].Hand = []engine.Card{{ID: "wx-1", Kind: engine.CardWuxiek, Name: "无懈可击"}}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlayCard(0, "lb-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if !g.HasJudgeKindForTest(1, engine.CardLeBu) {
		t.Fatal("expected lebu in judge zone")
	}
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeWuxiekLebu)
	// 玩家 1 打出无懈可击
	if err := g.RespondWuxiekForTest(1, "wx-1", &events); err != nil {
		t.Fatal(err)
	}
	// 跳过反无懈可击窗口，使第一张无懈可击生效
	if g.Pending != nil && g.Pending.ResponseMode == engine.ResponseModeWuxiekLebu {
		if err := g.PassResponse(g.CurrentTurn, &events); err != nil {
			t.Fatal(err)
		}
	}
	// 现在乐不思蜀应该被取消了
	if g.HasJudgeKindForTest(1, engine.CardLeBu) || g.Players[1].SkipPlay {
		t.Fatalf("expected lebu cancelled, judge=%+v skip=%v", g.Players[1].JudgeArea, g.Players[1].SkipPlay)
	}
}

// TestScenario_SimaYiShaXiahouDun 司马懿杀夏侯惇场景：
// 司马懿 HP=1 手牌=[酒(黑桃2), 杀]
// 夏侯惇 HP=1 手牌=[桃, 杀]
// 司马懿出杀 → 夏侯惇出闪？不，夏侯惇没有闪 → 受到1点伤害 → HP=0濒死
// → 夏侯惇用桃自救 → 刚烈判定 → 司马懿鬼才改判 → 根据判定结果：
//   - 非红桃 → 司马懿弃2牌或扣1血 → 司马懿HP=1扣血→HP=0濒死→酒自救→最终HP=1
func TestScenario_SimaYiShaXiahouDun(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-sima-xiahou", "司马懿", engine.CharSimaYi, engine.CharXiahouDun)
	if err != nil {
		t.Fatal(err)
	}

	// 司马懿 HP=1 手牌=[酒(黑桃2), 杀]
	g.Players[0].Hand = []engine.Card{
		{ID: "jiu-sima", Kind: engine.CardJiu, Name: "酒", Suit: "S", Rank: 2, Label: "黑桃2"},
		{ID: "sha-sima", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[0].HP = 1

	// 夏侯惇 HP=1 手牌=[桃, 杀]
	g.Players[1].Hand = []engine.Card{
		{ID: "tao-xiahou", Kind: engine.CardTao, Name: "桃"},
		{ID: "sha-xiahou", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].HP = 1

	// 牌堆：刚烈判定牌（黑桃3 → 刚烈生效）
	g.DrawPile = []engine.Card{
		{ID: "judge1", Suit: "S", Kind: engine.CardSha, Name: "杀", Label: "黑桃3"},
	}

	setupPlayingTurn(g, 0)

	t.Logf("司马懿 Skills=%v, 夏侯惇 Skills=%v", g.Players[0].Character.SkillIDs, g.Players[1].Character.SkillIDs)
	t.Logf("=== 初始: 司马懿 HP=%d 手牌=%d, 夏侯惇 HP=%d 手牌=%d ===",
		g.Players[0].HP, len(g.Players[0].Hand), g.Players[1].HP, len(g.Players[1].Hand))

	var events []engine.GameEvent

	// 1. 司马懿出杀
	if err := g.PlaySha(0, "sha-sima", 1, &events); err != nil {
		t.Fatalf("PlaySha 失败: %v", err)
	}
	t.Logf("出杀后: Pending=%s", g.Pending.ResponseMode)

	// 2. 夏侯惇没有闪 → PassResponse（杀命中）
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatalf("夏侯惇 PassResponse 失败: %v", err)
	}
	t.Logf("夏侯惇无闪后: HP=%d Pending=%s", g.Players[1].HP, pendingMode(g))

	// 3. 夏侯惇 HP=0 → 濒死窗口
	assertPendingMode(t, g, engine.ResponseModeDying)
	t.Logf("濒死窗口: Actor=%d", g.PendingActorSeat())

	// 4. 夏侯惇自己出桃自救
	if err := g.RespondCard(1, "tao-xiahou", &events); err != nil {
		t.Fatalf("夏侯惇出桃失败: %v", err)
	}
	t.Logf("夏侯惇自救后: HP=%d Pending=%s", g.Players[1].HP, pendingMode(g))

	// 5. 刚烈窗口
	assertPendingMode(t, g, engine.ResponseModeSkillGanglieOffer)
	t.Logf("刚烈窗口: Actor=%d", g.PendingActorSeat())

	// 6. 夏侯惇发动刚烈
	if err := g.UseSkill(1, engine.UseSkillRequest{SkillID: engine.SkillGanglie}, &events); err != nil {
		t.Fatalf("发动刚烈失败: %v", err)
	}
	t.Logf("刚烈判定后: Pending=%s", pendingMode(g))

	// 7. 鬼才改判窗口（司马懿，当前回合角色优先）
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillGuicai {
		t.Fatalf("期望鬼才窗口，实际 Pending=%v", pendingMode(g))
	}
	t.Logf("鬼才窗口: 判定牌=%s(黑桃→刚烈生效)", g.Pending.JudgeCard.Label)

	// 8. ★ 司马懿的正确选择：pass鬼才
	// 改判黑桃→黑桃没意义，不如保留酒用于濒死自救。
	if err := g.PassGuicai(0, &events); err != nil {
		t.Fatalf("PassGuicai 失败: %v", err)
	}
	t.Logf("司马懿pass鬼才后: Pending=%s", pendingMode(g))

	// 9. 刚烈判定黑桃 → 生效 → 司马懿手牌<2 → 扣血
	assertPendingMode(t, g, engine.ResponseModeSkillGanglieChoice)
	t.Logf("刚烈choice: 司马懿手牌=%d(只剩杀)", len(g.Players[0].Hand))

	// 10. 司马懿扣血
	if err := g.GanglieTakeDamage(0, &events); err != nil {
		t.Fatalf("刚烈扣血失败: %v", err)
	}
	t.Logf("刚烈扣血后: 司马懿 HP=%d Pending=%s", g.Players[0].HP, pendingMode(g))

	// 11. 司马懿 HP=0 → 濒死 → 用酒自救
	assertPendingMode(t, g, engine.ResponseModeDying)
	t.Logf("司马懿濒死: 用酒自救")

	if err := g.RespondCard(0, "jiu-sima", &events); err != nil {
		t.Fatalf("司马懿酒自救失败: %v", err)
	}
	t.Logf("司马懿自救后: HP=%d Pending=%s", g.Players[0].HP, pendingMode(g))

	// 12. 反馈窗口：司马懿（seat=0）从夏侯惇拿牌
	// 司马懿有反馈技能！ActorSeat=0(司马懿), SubjectSeat=1(夏侯惇)
	t.Logf("反馈Pending: RespMode=%s ActorSeat=%d TargetIdx=%d SrcIdx=%d WinKind=%s",
		g.Pending.ResponseMode, g.Pending.ActorSeat, g.Pending.TargetIndex, g.Pending.SourceIndex, g.Pending.WindowKind)
	assertPendingMode(t, g, engine.ResponseModeSkillFankui)

	// 司马懿(seat=0)从夏侯惇手牌拿"sha-sima"（夏侯惇手牌里的杀）
	if err := g.FankuiTakeFrom(0, "hand", "sha-xiahou", &events); err != nil {
		t.Fatalf("反馈拿牌失败: %v", err)
	}
	t.Logf("反馈拿牌后: 司马懿手牌=%d, 夏侯惇手牌=%d", len(g.Players[0].Hand), len(g.Players[1].Hand))

	t.Logf("=== 最终: 司马懿 HP=%d 手牌=%d, 夏侯惇 HP=%d 手牌=%d ===",
		g.Players[0].HP, len(g.Players[0].Hand), g.Players[1].HP, len(g.Players[1].Hand))

	// 司马懿: 刚烈扣1血→濒死→酒自救→HP=1, 反馈拿夏侯惇的杀→手牌+1
	if g.Players[0].HP != 1 {
		t.Errorf("司马懿 HP=%d, want 1", g.Players[0].HP)
	}
	// 夏侯惇: HP=1, 手牌=0(杀被反馈拿走)
	if g.Players[1].HP != 1 {
		t.Errorf("夏侯惇 HP=%d, want 1", g.Players[1].HP)
	}
}

func pendingMode(g *engine.Game) string {
	if g.Pending == nil {
		return "<nil>"
	}
	return g.Pending.ResponseMode
}
