package engine_test

import (
	"fmt"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func TestPaoxiaoUnlimitedSha(t *testing.T) {
	g, err := engine.NewSolo1v1("sp1", "玩家", engine.CharZhangFei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	if !g.CanUseSha(0) {
		t.Fatal("expected paoxiao to allow sha")
	}
	g.Players[0].ShaUsedThisTurn = true
	if !g.CanUseSha(0) {
		t.Fatal("expected paoxiao to ignore sha_used flag")
	}
}

func TestWushengRedAsShaRequiresActivate(t *testing.T) {
	g, err := engine.NewSolo1v1("sw1", "玩家", engine.CharGuanYu, engine.CharZhangFei)
	if err != nil {
		t.Fatal(err)
	}
	redShan := engine.Card{ID: "shan-r", Kind: engine.CardShan, Suit: "H", Name: "闪"}
	redTrick := engine.Card{ID: "bl-r", Kind: engine.CardBingLiang, Suit: "D", Name: "兵粮寸断"}
	blackShan := engine.Card{ID: "shan-b", Kind: engine.CardShan, Suit: "S", Name: "闪"}

	if g.CardPlaysAsForTest(0, redShan, engine.CardSha) {
		t.Fatal("wusheng should not apply before activation")
	}
	if g.CardPlaysAsForTest(0, redTrick, engine.CardSha) {
		t.Fatal("red trick should not be sha before activation")
	}

	var events []engine.GameEvent
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	if err := g.UseSkill(0, engine.UseSkillRequest{SkillID: engine.SkillWusheng}, &events); err != nil {
		t.Fatal(err)
	}
	if !g.CardPlaysAsForTest(0, redShan, engine.CardSha) {
		t.Fatal("expected red shan as sha after wusheng activated")
	}
	if !g.CardPlaysAsForTest(0, redTrick, engine.CardSha) {
		t.Fatal("expected red trick as sha when wusheng active")
	}
	if g.CardPlaysAsForTest(0, blackShan, engine.CardSha) {
		t.Fatal("black shan should not be sha for wusheng")
	}
}

func TestLongdanShaShan(t *testing.T) {
	g, err := engine.NewSolo1v1("sl1", "玩家", engine.CharZhaoYun, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	shan := engine.Card{ID: "shan-1", Kind: engine.CardShan, Name: "闪"}
	sha := engine.Card{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}
	if !g.CardPlaysAsForTest(0, shan, engine.CardSha) {
		t.Fatal("expected longdan shan as sha")
	}
	if !g.CardPlaysAsForTest(0, sha, engine.CardShan) {
		t.Fatal("expected longdan sha as shan")
	}
}

func TestRendeGiveAndHeal(t *testing.T) {
	g, err := engine.NewSolo1v1("sr1", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "c1", Kind: engine.CardSha, Name: "杀"},
		{ID: "c2", Kind: engine.CardShan, Name: "闪"},
	}
	g.Players[0].HP = 3
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	before := len(g.Players[1].Hand)
	var events []engine.GameEvent
	err = g.UseSkill(0, engine.UseSkillRequest{
		SkillID: engine.SkillRende, TargetIndex: 1, CardIDs: []string{"c1", "c2"},
	}, &events)
	if err != nil {
		t.Fatal(err)
	}
	if g.Players[0].HP != 4 {
		t.Fatalf("expected rende heal, hp=%d", g.Players[0].HP)
	}
	if len(g.Players[1].Hand) != before+2 {
		t.Fatalf("expected 2 cards given, got %d total", len(g.Players[1].Hand))
	}
}

func TestHeroesCatalog(t *testing.T) {
	heroes := engine.HeroesCatalog()
	if len(heroes) < 32 {
		t.Fatalf("expected at least 32 heroes, got %d", len(heroes))
	}
	for _, h := range heroes {
		if len(h.Skills) == 0 {
			t.Fatalf("hero %s should have skills", h.ID)
		}
	}
	_ = skill.CharZhaoYun
}

func TestKongchengBlocksShaAndJuedou(t *testing.T) {
	g, err := engine.NewSolo1v1("sk1", "玩家", engine.CharGuanYu, engine.CharZhugeLiang)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[1].Hand = nil
	g.SyncCounts()
	if !g.TargetBlockedBySkillForTest(1, engine.CardSha) {
		t.Fatal("expected kongcheng to block sha")
	}
	if !g.TargetBlockedBySkillForTest(1, engine.CardJueDou) {
		t.Fatal("expected kongcheng to block juedou")
	}
	g.Players[1].Hand = []engine.Card{{ID: "h1", Kind: engine.CardShan, Name: "闪"}}
	g.SyncCounts()
	if g.TargetBlockedBySkillForTest(1, engine.CardSha) {
		t.Fatal("kongcheng should not block when hand not empty")
	}

	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = nil
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != engine.ErrInvalidTarget {
		t.Fatalf("expected invalid target for kongcheng, got %v", err)
	}
}

func TestPeekDeckReorderDrawPile(t *testing.T) {
	g, err := engine.NewSolo1v1("sg1", "玩家", engine.CharZhugeLiang, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPrepare
	g.CurrentTurn = 0
	g.DrawPile = []engine.Card{
		{ID: "a", Kind: engine.CardSha, Name: "杀", Label: "A"},
		{ID: "b", Kind: engine.CardShan, Name: "闪", Label: "B"},
		{ID: "c", Kind: engine.CardTao, Name: "桃", Label: "C"},
		{ID: "d", Kind: engine.CardSha, Name: "杀", Label: "D"},
	}
	var events []engine.GameEvent
	if err := g.StartPeekDeck(0, skill.IDGuanxing, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModePeekDeck || len(g.Pending.RevealedCards) != 2 {
		t.Fatalf("expected 2 peeked cards in 1v1, got %+v", g.Pending)
	}
	top := []string{g.Pending.RevealedCards[1].ID}
	bottom := []string{g.Pending.RevealedCards[0].ID}
	g.Players[0].SkipDraw = true
	if err := g.FinishPeekDeck(0, engine.PeekDeckRequest{TopCardIDs: top, BottomCardIDs: bottom}, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.DrawPile) < 4 {
		t.Fatalf("draw pile too short: %d", len(g.DrawPile))
	}
	if g.DrawPile[0].ID != top[0] {
		t.Fatalf("expected top card %s, got %s", top[0], g.DrawPile[0].ID)
	}
	if g.DrawPile[len(g.DrawPile)-1].ID != bottom[0] {
		t.Fatalf("expected bottom card %s, got %s", bottom[0], g.DrawPile[len(g.DrawPile)-1].ID)
	}
}

func TestSkillHookInfrastructure(t *testing.T) {
	g, err := engine.NewSolo1v1("hk1", "玩家", engine.CharMaChao, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[1].PlusHorse = &engine.Card{ID: "horse1", Kind: engine.CardPlusHorse, Name: "+1马"}
	delta := g.RunSkillHooks(nil, skill.HookCall{Kind: skill.HookDistanceDelta, From: 0, To: 1})
	if delta.Int != -1 {
		t.Fatalf("expected mashi delta -1, got %d", delta.Int)
	}
	g2, _ := engine.NewSolo1v1("hk2", "玩家", engine.CharGuanYu, engine.CharZhugeLiang)
	g2.Players[1].Hand = nil
	g2.SyncCounts()
	if !g2.RunSkillHooks(nil, skill.HookCall{Kind: skill.HookTargetBlocked, Target: 1, CardKind: engine.CardSha}).Bool {
		t.Fatal("expected kongcheng via hook")
	}
}

func TestApplyDamageRunsHooks(t *testing.T) {
	var got skill.DamageCtx
	skill.Register(skill.Decl{
		Meta: skill.Meta{ID: "_test_damage_hook", Name: "测", Kind: skill.KindPassive},
		OnDamageDealt: func(_ skill.Runtime, ctx skill.DamageCtx) error {
			got = ctx
			return nil
		},
	})
	t.Cleanup(func() { skill.Unregister("_test_damage_hook") })

	g, err := engine.NewSolo1v1("dmg1", "玩家", engine.CharGuanYu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[1].Character.SkillIDs = append(g.Players[1].Character.SkillIDs, "_test_damage_hook")
	var events []engine.GameEvent
	g.ApplyDamageForTest(0, 1, 2, engine.CardSha, "杀", &events)
	if got.Target != 1 || got.Amount != 2 || got.CardKind != engine.CardSha {
		t.Fatalf("unexpected damage ctx: %+v", got)
	}
}

func TestMashiReducesDistance(t *testing.T) {
	g, err := engine.NewSolo1v1("ms1", "玩家", engine.CharMaChao, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[1].PlusHorse = &engine.Card{ID: "horse1", Kind: engine.CardPlusHorse, Name: "+1马"}
	if g.DistanceBetween(0, 1) != 1 {
		t.Fatalf("expected mashi distance 1, got %d", g.DistanceBetween(0, 1))
	}
	g2, _ := engine.NewSolo1v1("ms2", "玩家", engine.CharGuanYu, engine.CharMaChao)
	g2.Players[1].PlusHorse = &engine.Card{ID: "horse1", Kind: engine.CardPlusHorse, Name: "+1马"}
	if g2.DistanceBetween(0, 1) != 2 {
		t.Fatalf("expected distance 2 without mashi, got %d", g2.DistanceBetween(0, 1))
	}
}

func TestQicaiIgnoresBingliangDistance(t *testing.T) {
	g, err := engine.NewSolo1v1("qc1", "玩家", engine.CharHuangYueying, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[1].PlusHorse = &engine.Card{ID: "horse1", Kind: engine.CardPlusHorse, Name: "+1马"}
	if !g.CanBingliangTargetForTest(0, 1) {
		t.Fatal("expected qicai to ignore bingliang distance")
	}
	g2, _ := engine.NewSolo1v1("qc2", "玩家", engine.CharGuanYu, engine.CharHuangYueying)
	g2.Players[1].PlusHorse = &engine.Card{ID: "horse1", Kind: engine.CardPlusHorse, Name: "+1马"}
	if g2.CanBingliangTargetForTest(0, 1) {
		t.Fatal("expected bingliang blocked at distance 2 without qicai")
	}
}

func TestJizhiDrawsAfterInstantTrick(t *testing.T) {
	g, err := engine.NewSolo1v1("jz1", "玩家", engine.CharHuangYueying, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.DrawPile = []engine.Card{{ID: "draw1", Kind: engine.CardSha, Name: "杀", Label: "摸牌"}}
	handBefore := len(g.Players[0].Hand)
	var events []engine.GameEvent
	g.NotifyInstantTrickUsedForTest(0, engine.CardWuZhong, &events)
	if len(g.Players[0].Hand) != handBefore+1 {
		t.Fatalf("expected jizhi draw, hand=%d", len(g.Players[0].Hand))
	}
}

func TestTieqiBlocksShan(t *testing.T) {
	g, err := engine.NewSolo1v1("tq1", "玩家", engine.CharMaChao, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{{ID: "shan1", Kind: engine.CardShan, Name: "闪"}}
	g.DrawPile = []engine.Card{{ID: "judge1", Suit: "S", Kind: engine.CardSha, Name: "杀", Label: "黑桃7"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if !g.Pending.TieqiPending {
		t.Fatal("expected tieqi pending")
	}
	if err := g.ApplyTieqi(0, &events); err != nil {
		t.Fatal(err)
	}
	if !g.Pending.ShaUnblockable {
		t.Fatal("expected sha unblockable after black judge")
	}
	if err := g.RespondCard(1, "shan1", &events); err == nil {
		t.Fatal("expected shan blocked by tieqi")
	}
	hpBefore := g.Players[1].HP
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != hpBefore-1 {
		t.Fatalf("expected damage, hp=%d", g.Players[1].HP)
	}
}

func TestQingguoBlackCardAsShan(t *testing.T) {
	g, err := engine.NewSolo1v1("qg1", "玩家", engine.CharZhenJi, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	black := engine.Card{ID: "b1", Kind: engine.CardSha, Suit: "S", Name: "杀"}
	if !g.CardPlaysAsForTest(0, black, engine.CardShan) {
		t.Fatal("expected black card to play as shan via qingguo")
	}
	red := engine.Card{ID: "r1", Kind: engine.CardSha, Suit: "H", Name: "杀"}
	if g.CardPlaysAsForTest(0, red, engine.CardShan) {
		t.Fatal("red card should not play as shan via qingguo")
	}
}

func TestQingguoRespondShaWithBlackCard(t *testing.T) {
	g, err := engine.NewSolo1v1("qg2", "玩家", engine.CharZhenJi, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 1
	g.Players[0].Hand = []engine.Card{{ID: "black1", Kind: engine.CardSha, Suit: "S", Name: "杀", Label: "黑桃杀"}}
	g.Players[1].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀", Label: "杀"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(1, "sha1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.RespondCard(0, "black1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending != nil {
		t.Fatalf("expected sha dodged, pending=%+v", g.Pending)
	}
	found := false
	for _, c := range g.DiscardPile {
		if c.ID == "black1" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected black card discarded as shan")
	}
}

func TestFankuiTakesCardAfterShaHit(t *testing.T) {
	g, err := engine.NewSolo1v1("fk1", "玩家", engine.CharGuanYu, engine.CharSimaYi)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{
		{ID: "sha1", Kind: engine.CardSha, Name: "杀"},
		{ID: "extra", Kind: engine.CardShan, Name: "闪", Label: "留牌"},
	}
	g.Players[1].Hand = []engine.Card{{ID: "src1", Kind: engine.CardShan, Name: "闪", Label: "闪"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillFankui {
		t.Fatalf("expected fankui window, pending=%+v", g.Pending)
	}
	if err := g.FankuiTakeFrom(1, "hand", "extra", &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 0 {
		t.Fatalf("expected source to lose taken card, hand=%+v", g.Players[0].Hand)
	}
	found := false
	for _, c := range g.Players[1].Hand {
		if c.ID == "extra" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected sima yi to gain card")
	}
}

func TestGuicaiReplacesJudge(t *testing.T) {
	g, err := engine.NewSolo1v1("gc1", "玩家", engine.CharMaChao, engine.CharSimaYi)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{
		{ID: "replace", Kind: engine.CardShan, Suit: "H", Name: "闪", Label: "红闪"},
	}
	g.DrawPile = []engine.Card{{ID: "judge1", Suit: "S", Kind: engine.CardSha, Name: "杀", Label: "黑桃7"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.ApplyTieqi(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillGuicai {
		t.Fatalf("expected guicai window, pending=%+v", g.Pending)
	}
	if err := g.ApplyGuicaiReplace(1, "replace", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending.ShaUnblockable {
		t.Fatal("expected red guicai replace to allow shan")
	}
}

func TestLuoshenBlackGainRedStop(t *testing.T) {
	g, err := engine.NewSolo1v1("ls1", "玩家", engine.CharZhenJi, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPrepare
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{}
	g.DrawPile = []engine.Card{
		{ID: "black1", Suit: "S", Kind: engine.CardSha, Name: "杀", Label: "黑桃7"},
		{ID: "red1", Suit: "H", Kind: engine.CardShan, Name: "闪", Label: "红桃2"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.StartLuoshen(0, &events); err != nil {
		t.Fatal(err)
	}
	hasBlack := false
	for _, c := range g.Players[0].Hand {
		if c.ID == "black1" {
			hasBlack = true
		}
	}
	if !hasBlack {
		t.Fatalf("expected black1 in hand, hand=%+v", g.Players[0].Hand)
	}
	if g.TurnStep != engine.StepPrepare {
		t.Fatalf("expected prepare after black, step=%s", g.TurnStep)
	}
	handLen := len(g.Players[0].Hand)
	if err := g.StartLuoshen(0, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != handLen {
		t.Fatalf("expected no gain on red stop, hand=%+v", g.Players[0].Hand)
	}
	if g.TurnStep != engine.StepDraw && g.TurnStep != engine.StepPlay {
		t.Fatalf("expected turn to advance after red luoshen, step=%s", g.TurnStep)
	}
}

func TestJianxiongGainDamageCard(t *testing.T) {
	g, err := engine.NewSolo1v1("jx1", "玩家", engine.CharCaoCao, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 1
	g.Players[1].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀", Label: "杀"}}
	g.Players[0].Hand = nil
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(1, "sha1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != "skill_jianxiong" {
		t.Fatalf("expected jianxiong window, pending=%+v", g.Pending)
	}
	if err := g.UseSkill(0, engine.UseSkillRequest{SkillID: engine.SkillJianxiong}, &events); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range g.Players[0].Hand {
		if c.ID == "sha1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected cao cao to gain sha1, hand=%+v", g.Players[0].Hand)
	}
}

func TestGanglieJudgeNotHeartDamagesSource(t *testing.T) {
	g, err := engine.NewSolo1v1("gl1", "玩家", engine.CharXiahouDun, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 1
	g.Players[1].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀", Label: "杀"}}
	g.Players[0].Hand = nil
	g.DrawPile = []engine.Card{{ID: "judge1", Suit: "S", Rank: 3, Kind: engine.CardSha, Name: "杀", Label: "黑桃3"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(1, "sha1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.UseSkill(0, engine.UseSkillRequest{SkillID: engine.SkillGanglie}, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != "skill_ganglie_choice" {
		t.Fatalf("expected ganglie choice, pending=%+v", g.Pending)
	}
	hpBefore := g.Players[1].HP
	if err := g.UseSkill(1, engine.UseSkillRequest{SkillID: engine.SkillGanglie, TargetZone: "take_damage"}, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != hpBefore-1 {
		t.Fatalf("expected source to take 1 ganglie damage, hp=%d", g.Players[1].HP)
	}
}

func TestLuoyiSkipDrawShaExtraDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("ly1", "玩家", engine.CharXuChu, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepDraw
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀", Label: "杀"}}
	g.Players[1].Hand = []engine.Card{}
	g.DrawPile = []engine.Card{{ID: "d1", Kind: engine.CardShan, Name: "闪", Label: "闪"}}
	g.DiscardPile = nil
	g.SyncCounts()
	g.Players[0].SkillCounters = map[string]int{"draw_choice_pending": 1}

	var events []engine.GameEvent
	if err := g.ActivateLuoyi(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.TurnStep != engine.StepPlay {
		t.Fatalf("expected play after luoyi, step=%s", g.TurnStep)
	}
	if len(g.Players[0].Hand) != 1 {
		t.Fatalf("expected no draw, hand=%d", len(g.Players[0].Hand))
	}
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.Damage != 2 {
		t.Fatalf("expected luoyi sha damage 2, pending=%+v", g.Pending)
	}
}

func TestTuxiSkipDrawTakeCard(t *testing.T) {
	g, err := engine.NewSolo1v1("tx1", "玩家", engine.CharZhangLiao, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepDraw
	g.CurrentTurn = 0
	g.Players[0].Hand = nil
	g.Players[1].Hand = []engine.Card{
		{ID: "h1", Kind: engine.CardShan, Name: "闪", Label: "闪"},
		{ID: "h2", Kind: engine.CardTao, Name: "桃", Label: "桃"},
	}
	g.Players[1].Weapon = &engine.Card{ID: "w1", Kind: engine.CardWeapon1, Name: "青釭剑", Label: "青釭剑"}
	g.DrawPile = []engine.Card{
		{ID: "d1", Kind: engine.CardSha, Name: "杀", Label: "杀"},
		{ID: "d2", Kind: engine.CardSha, Name: "杀", Label: "杀"},
	}
	g.DiscardPile = nil
	g.SyncCounts()
	g.Players[0].SkillCounters = map[string]int{"draw_choice_pending": 1}

	var events []engine.GameEvent
	if err := g.StartTuxi(0, 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillTuxi {
		t.Fatalf("expected tuxi pending, got %+v", g.Pending)
	}
	if err := g.TuxiTakeFrom(0, engine.EquipWeapon, "w1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].Weapon != nil {
		t.Fatalf("expected opponent weapon taken, still has %v", g.Players[1].Weapon)
	}
	if g.TurnStep != engine.StepPlay {
		t.Fatalf("expected play after tuxi finish, step=%s", g.TurnStep)
	}
	if len(g.Players[0].Hand) != 2 {
		t.Fatalf("expected 1 draw + 1 taken = 2 cards, hand=%d", len(g.Players[0].Hand))
	}
	hasWeapon := false
	for _, c := range g.Players[0].Hand {
		if c.ID == "w1" {
			hasWeapon = true
		}
	}
	if !hasWeapon {
		t.Fatal("expected taken weapon in hand")
	}
}

func TestTuxiAIMirrorSim(t *testing.T) {
	g, err := engine.NewSolo1v1("tx-mirror", "甲", engine.CharZhangLiao, engine.CharZhangLiao)
	if err != nil {
		t.Fatal(err)
	}
	run := runAISimulation(t, g, defaultSimMaxSteps)
	if run.result.stuck {
		t.Fatalf("zhang_liao mirror sim stuck at %q forceErr=%q", run.stuckAtFP, run.forceErr)
	}
	if !run.result.finished {
		t.Fatal("sim did not finish")
	}
}

func TestYijiDrawAndGive(t *testing.T) {
	g, err := engine.NewSolo1v1("yj1", "玩家", engine.CharGuoJia, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 1
	g.Players[0].Hand = []engine.Card{
		{ID: "h1", Kind: engine.CardSha, Name: "杀", Label: "杀"},
		{ID: "h2", Kind: engine.CardShan, Name: "闪", Label: "闪"},
	}
	g.Players[1].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀", Label: "杀"}}
	g.DrawPile = []engine.Card{
		{ID: "d1", Kind: engine.CardSha, Name: "杀", Label: "杀"},
		{ID: "d2", Kind: engine.CardShan, Name: "闪", Label: "闪"},
	}
	g.DiscardPile = nil
	g.SyncCounts()

	var events []engine.GameEvent
	if err := g.PlaySha(1, "sha1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillYijiOffer {
		t.Fatalf("expected yiji offer after damage, pending=%+v", g.Pending)
	}
	if err := g.ApplyYiji(0, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 4 {
		t.Fatalf("expected 2+2 hand after yiji draw, got %d", len(g.Players[0].Hand))
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillYijiGive {
		t.Fatalf("expected yiji give pending, got %+v", g.Pending)
	}
	if err := g.YijiGiveCards(0, 1, []string{"h1", "h2"}, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 2 || len(g.Players[1].Hand) != 2 {
		t.Fatalf("expected give transfer, p0=%d p1=%d", len(g.Players[0].Hand), len(g.Players[1].Hand))
	}
}

func TestZhihengDiscardDraw(t *testing.T) {
	g, err := engine.NewSolo1v1("zh1", "玩家", engine.CharSunQuan, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{
		{ID: "h1", Kind: engine.CardSha, Name: "杀", Label: "杀"},
		{ID: "h2", Kind: engine.CardShan, Name: "闪", Label: "闪"},
	}
	g.DrawPile = []engine.Card{
		{ID: "d1", Kind: engine.CardTao, Name: "桃", Label: "桃"},
		{ID: "d2", Kind: engine.CardTao, Name: "桃", Label: "桃"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.ActivateZhiheng(0, []string{"h1", "h2"}, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 2 {
		t.Fatalf("expected 2 cards after zhiheng, hand=%d", len(g.Players[0].Hand))
	}
	if g.Players[0].SkillCounters["zhiheng_used_play"] != 1 {
		t.Fatal("expected zhiheng marked used")
	}
}

func TestJieyinHealBoth(t *testing.T) {
	g, err := engine.NewSolo1v1("jy1", "玩家", engine.CharSunShangxiang, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].HP = 2
	g.Players[1].HP = 1
	g.Players[0].Hand = []engine.Card{
		{ID: "h1", Kind: engine.CardSha, Name: "杀", Label: "杀"},
		{ID: "h2", Kind: engine.CardShan, Name: "闪", Label: "闪"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.ActivateJieyin(0, 1, []string{"h1", "h2"}, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[0].HP != 3 || g.Players[1].HP != 2 {
		t.Fatalf("expected both healed, p0=%d p1=%d", g.Players[0].HP, g.Players[1].HP)
	}
}

func TestXiaojiDrawOnEquipLost(t *testing.T) {
	g, err := engine.NewSolo1v1("xj1", "玩家", engine.CharSunShangxiang, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 1
	g.Players[0].Weapon = &engine.Card{ID: "w1", Kind: engine.CardWeapon1, Name: "诸葛连弩", Label: "诸葛连弩"}
	g.Players[0].Hand = nil
	g.Players[1].Hand = []engine.Card{{ID: "gh1", Kind: engine.CardGuoHe, Name: "过河拆桥", Label: "过河拆桥"}}
	g.DrawPile = []engine.Card{
		{ID: "d1", Kind: engine.CardSha, Name: "杀", Label: "杀"},
		{ID: "d2", Kind: engine.CardShan, Name: "闪", Label: "闪"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlayCardWithTarget(1, "gh1", engine.PlayTarget{SeatIndex: 0, Zone: engine.EquipWeapon, CardID: "w1"}, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		if err := g.PassResponse(0, &events); err != nil {
			t.Fatal(err)
		}
	}
	if g.Players[0].Weapon != nil {
		t.Fatal("expected weapon taken")
	}
	if len(g.Players[0].Hand) != 2 {
		t.Fatalf("expected xiaoji draw 2, hand=%d", len(g.Players[0].Hand))
	}
}

func TestYingziDrawThree(t *testing.T) {
	g, err := engine.NewSolo1v1("yz1", "玩家", engine.CharZhouYu, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPrepare
	g.CurrentTurn = 0
	handBefore := len(g.Players[0].Hand)
	var events []engine.GameEvent
	if err := g.PassPrepare(0, &events); err != nil {
		t.Fatal(err)
	}
	if got := len(g.Players[0].Hand) - handBefore; got != 3 {
		t.Fatalf("expected yingzi draw 3, got %d", got)
	}
}

func TestFanjianMatchDamages(t *testing.T) {
	g, err := engine.NewSolo1v1("fj1", "玩家", engine.CharZhouYu, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[1].HP = 4
	g.Players[0].Hand = []engine.Card{
		{ID: "h1", Kind: engine.CardSha, Name: "杀", Label: "杀", Suit: "H"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.ActivateFanjian(0, "h1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillFanjianSuit {
		t.Fatalf("expected fanjian suit pending, got %+v", g.Pending)
	}
	if err := g.ResolveFanjianSuit(1, "H", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != 3 {
		t.Fatalf("expected fanjian damage, hp=%d", g.Players[1].HP)
	}
}

func TestTianxiangRedirectsDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("tx1", "玩家", engine.CharXiaoQiao, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 1
	g.Players[0].HP = 3
	g.Players[1].HP = 4
	g.Players[0].Hand = []engine.Card{
		{ID: "r1", Kind: engine.CardSha, Name: "杀", Label: "杀", Suit: "H"},
	}
	g.Players[1].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀", Label: "杀"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(1, "sha1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillTianxiang {
		t.Fatalf("expected tianxiang offer, pending=%+v", g.Pending)
	}
	if err := g.ApplyTianxiang(0, "r1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[0].HP != 3 {
		t.Fatalf("xiao qiao should avoid damage, hp=%d", g.Players[0].HP)
	}
	if g.Players[1].HP != 3 {
		t.Fatalf("opponent should take redirected damage, hp=%d", g.Players[1].HP)
	}
}

func TestQixiTakeHandCard(t *testing.T) {
	g, err := engine.NewSolo1v1("qx1", "玩家", engine.CharGanNing, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{
		{ID: "b1", Kind: engine.CardSha, Name: "杀", Label: "杀", Suit: "S"},
	}
	g.Players[1].Hand = []engine.Card{
		{ID: "h1", Kind: engine.CardTao, Name: "桃", Label: "桃"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.ActivateQixi(0, "b1", &events); err != nil {
		t.Fatal(err)
	}
	if err := g.QixiTakeFrom(0, "h1", &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 1 || g.Players[0].Hand[0].ID != "h1" {
		t.Fatalf("expected qixi take, hand=%v", g.Players[0].Hand)
	}
	if len(g.Players[1].Hand) != 0 {
		t.Fatalf("expected opponent hand empty, got %d", len(g.Players[1].Hand))
	}
}

func TestYinghunDrawBoth(t *testing.T) {
	g, err := engine.NewSolo1v1("yh1", "玩家", engine.CharSunJian, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPrepare
	g.CurrentTurn = 0
	drawBefore := len(g.DrawPile)
	h0, h1 := len(g.Players[0].Hand), len(g.Players[1].Hand)
	_ = drawBefore
	var events []engine.GameEvent
	if err := g.ActivateYinghun(0, 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.ResolveYinghunChoice(1, engine.YinghunOptionDrawBoth, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) < h0+1 || len(g.Players[1].Hand) < h1+1 {
		t.Fatalf("expected both draw at least 1 from yinghun, p0=%d p1=%d", len(g.Players[0].Hand), len(g.Players[1].Hand))
	}
	if g.TurnStep != engine.StepPlay && g.TurnStep != engine.StepDraw {
		t.Fatalf("expected turn advanced after yinghun, step=%s", g.TurnStep)
	}
}

func TestLianyingDrawOnEmptyHand(t *testing.T) {
	g, err := engine.NewSolo1v1("ly1", "玩家", engine.CharLuXun, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{
		{ID: "sha1", Kind: engine.CardSha, Name: "杀", Label: "杀"},
	}
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 1 {
		t.Fatalf("expected lianying draw after playing last hand card, hand=%d", len(g.Players[0].Hand))
	}
}

func TestGuoseBlocksSha(t *testing.T) {
	g, err := engine.NewSolo1v1("gs1", "玩家", engine.CharDaQiao, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{
		{ID: "d1", Kind: engine.CardSha, Name: "杀", Label: "杀", Suit: "D"},
	}
	var events []engine.GameEvent
	if err := g.ActivateGuose(0, 1, "d1", &events); err != nil {
		t.Fatal(err)
	}
	if g.CanUseSha(1) {
		t.Fatal("expected guose to block opponent sha")
	}
}

func TestLiuliRedirectSha(t *testing.T) {
	g, err := engine.NewSolo1v1("ll1", "玩家", engine.CharLiuBei, engine.CharDaQiao)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{
		{ID: "sha1", Kind: engine.CardSha, Name: "杀", Label: "杀"},
	}
	g.Players[1].Hand = []engine.Card{
		{ID: "c1", Kind: engine.CardShan, Name: "闪", Label: "闪"},
		{ID: "c2", Kind: engine.CardTao, Name: "桃", Label: "桃"},
	}
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeSkillLiuli {
		t.Fatalf("expected liuli window, pending=%v", g.Pending)
	}
	if err := g.ApplyLiuli(1, "c2", 0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.TargetIndex != 0 {
		t.Fatalf("expected sha redirected to attacker, target=%d", g.Pending.TargetIndex)
	}
}

func TestKurouLoseHPDrawTwo(t *testing.T) {
	g, err := engine.NewSolo1v1("kr1", "玩家", engine.CharHuangGai, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].HP = 4
	handBefore := len(g.Players[0].Hand)
	var events []engine.GameEvent
	if err := g.ActivateKurou(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[0].HP != 3 {
		t.Fatalf("expected hp 3, got %d", g.Players[0].HP)
	}
	if len(g.Players[0].Hand) != handBefore+2 {
		t.Fatalf("expected +2 cards, hand=%d before=%d", len(g.Players[0].Hand), handBefore)
	}
}

func TestKejiSkipDiscard(t *testing.T) {
	g, err := engine.NewSolo1v1("kj1", "玩家", engine.CharLvMeng, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].HP = 3
	g.Players[0].Hand = make([]engine.Card, 6)
	for i := range g.Players[0].Hand {
		g.Players[0].Hand[i] = engine.Card{ID: fmt.Sprintf("c%d", i), Kind: engine.CardShan, Name: "闪"}
	}
	g.Players[0].ShaUsedThisTurn = false
	var events []engine.GameEvent
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.TurnStep == engine.StepDiscard {
		t.Fatal("expected keji to skip discard phase")
	}
	if g.CurrentTurn != 1 {
		t.Fatalf("expected turn ended, current=%d", g.CurrentTurn)
	}
}

func TestKejiNoSkipAfterSha(t *testing.T) {
	g, err := engine.NewSolo1v1("kj2", "玩家", engine.CharLvMeng, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].HP = 3
	g.Players[0].Hand = make([]engine.Card, 6)
	for i := range g.Players[0].Hand {
		g.Players[0].Hand[i] = engine.Card{ID: fmt.Sprintf("c%d", i), Kind: engine.CardShan, Name: "闪"}
	}
	g.SetSkillCounterForTest(0, "sha_in_play_phase", 1)
	var events []engine.GameEvent
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.TurnStep != engine.StepDiscard {
		t.Fatalf("expected discard phase after sha, step=%s", g.TurnStep)
	}
}

func TestHunziAwaken(t *testing.T) {
	g, err := engine.NewSolo1v1("hz1", "玩家", engine.CharSunCe, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPrepare
	g.CurrentTurn = 0
	g.Players[0].HP = 1
	g.Players[0].MaxHP = 4
	var events []engine.GameEvent
	if err := g.AwakenHunziForTest(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[0].MaxHP != 3 {
		t.Fatalf("expected max hp 3, got %d", g.Players[0].MaxHP)
	}
	if !g.HasSkillForTest(0, engine.SkillYingzi) || !g.HasSkillForTest(0, engine.SkillYinghun) {
		t.Fatalf("expected yingzi and yinghun, skills=%v", g.Players[0].Character.SkillIDs)
	}
	if g.HasSkillForTest(0, engine.SkillHunzi) {
		t.Fatal("hunzi should be removed after awaken")
	}
}

func TestJiangDrawOnRedShaHit(t *testing.T) {
	g, err := engine.NewSolo1v1("jg1", "玩家", engine.CharSunCe, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	handBefore := len(g.Players[0].Hand)
	var events []engine.GameEvent
	g.TryJiangDrawForTest(0, engine.Card{Kind: engine.CardSha, Suit: "H", Name: "杀"}, &events)
	if len(g.Players[0].Hand) != handBefore+1 {
		t.Fatalf("expected jiang draw, hand=%d before=%d", len(g.Players[0].Hand), handBefore)
	}
}

func TestJiangDrawOnJuedouWuxiek(t *testing.T) {
	g, err := engine.NewSolo1v1("jg2", "玩家", engine.CharSunCe, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[1].Hand = []engine.Card{
		{ID: "wx1", Kind: engine.CardWuxiek, Name: "无懈可击"},
	}
	g.Phase = engine.PhaseResponse
	g.Pending = &engine.PendingCombat{
		SourceIndex:  0,
		TargetIndex:  1,
		ReturnIndex:  0,
		ResponseMode: engine.ResponseModeWuxiekTrick,
		Card:         engine.Card{Kind: engine.CardJueDou, Name: "决斗"},
	}
	handBefore := len(g.Players[0].Hand)
	var events []engine.GameEvent
	if err := g.RespondWuxiekForTest(1, "wx1", &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != handBefore+1 {
		t.Fatalf("expected jiang draw on wuxiek juedou, hand=%d before=%d", len(g.Players[0].Hand), handBefore)
	}
}

func TestJijiRedAsTaoOutsideTurn(t *testing.T) {
	g, err := engine.NewSolo1v1("jj1", "玩家", engine.CharGuanYu, engine.CharHuaTuo)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[1].HP = 2
	g.Players[1].Hand = []engine.Card{
		{ID: "red-sha", Kind: engine.CardSha, Suit: "H", Label: "红桃杀", Name: "杀"},
	}
	g.SyncCounts()
	redSha := engine.Card{ID: "red-sha", Kind: engine.CardSha, Suit: "H", Name: "杀"}
	if !g.CardPlaysAsForTest(1, redSha, engine.CardTao) {
		t.Fatal("expected red card as tao outside own turn")
	}
	if g.CardPlaysAsForTest(1, engine.Card{ID: "black", Kind: engine.CardSha, Suit: "S", Name: "杀"}, engine.CardTao) {
		t.Fatal("black should not play as tao for jiji")
	}
	g.CurrentTurn = 1
	if g.CardPlaysAsForTest(1, redSha, engine.CardTao) {
		t.Fatal("jiji should not apply on own turn")
	}
	g.CurrentTurn = 0
	var events []engine.GameEvent
	if err := g.PlayJijiHealForTest(1, "red-sha", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != 3 {
		t.Fatalf("expected heal to 3, got %d", g.Players[1].HP)
	}
}

func TestBiyueDrawsAtEndTurn(t *testing.T) {
	g, err := engine.NewSolo1v1("by1", "玩家", engine.CharDiaoChan, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].HP = 3
	g.Players[0].Hand = make([]engine.Card, 3)
	for i := range g.Players[0].Hand {
		g.Players[0].Hand[i] = engine.Card{ID: fmt.Sprintf("h%d", i), Kind: engine.CardShan, Name: "闪"}
	}
	drawBefore := len(g.DrawPile)
	var events []engine.GameEvent
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 4 {
		t.Fatalf("expected biyue draw hand=4, got %d", len(g.Players[0].Hand))
	}
	foundBiyue := false
	for _, ev := range events {
		if ev.Type == "skill_biyue" {
			foundBiyue = true
		}
	}
	if !foundBiyue {
		t.Fatal("expected skill_biyue event")
	}
	_ = drawBefore
}

func TestWushuangDoubleShan(t *testing.T) {
	g, err := engine.NewSolo1v1("ws1", "玩家", engine.CharLvBu, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{
		{ID: "s1", Kind: engine.CardShan, Name: "闪"},
		{ID: "s2", Kind: engine.CardShan, Name: "闪"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponsesNeeded != 2 {
		t.Fatalf("expected wushuang responses_needed=2, got %+v", g.Pending)
	}
	if err := g.RespondCard(1, "s1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponsesNeeded != 1 {
		t.Fatalf("expected one shan remaining, pending=%+v", g.Pending)
	}
	if err := g.RespondCard(1, "s2", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending != nil {
		t.Fatal("sha should be dodged after two shans")
	}
}

func TestWushuangDoubleShaJuedou(t *testing.T) {
	g, err := engine.NewSolo1v1("ws2", "玩家", engine.CharLvBu, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "jd1", Kind: engine.CardJueDou, Name: "决斗"}}
	g.Players[1].Hand = []engine.Card{
		{ID: "k1", Kind: engine.CardSha, Name: "杀"},
		{ID: "k2", Kind: engine.CardSha, Name: "杀"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlayCard(0, "jd1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse && g.Pending != nil && g.Pending.ResponseMode == engine.ResponseModeWuxiekTrick {
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatal(err)
		}
	}
	if g.Pending == nil {
		t.Fatal("expected juedou pending")
	}
	if g.Pending.ResponsesNeeded != 2 {
		t.Fatalf("expected wushuang juedou need 2 sha, got %d", g.Pending.ResponsesNeeded)
	}
	if err := g.RespondCard(1, "k1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending.ResponsesNeeded != 1 {
		t.Fatalf("expected 1 sha left, got %d", g.Pending.ResponsesNeeded)
	}
	if err := g.RespondCard(1, "k2", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.TargetIndex != 0 {
		t.Fatalf("juedou should continue to lv bu, pending=%+v", g.Pending)
	}
}

func TestDyingRescueWithTao(t *testing.T) {
	g, err := engine.NewSolo1v1("dy1", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{{ID: "tao1", Kind: engine.CardTao, Name: "桃"}}
	g.Players[1].HP = 1
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeDying {
		t.Fatalf("expected dying, got %+v", g.Pending)
	}
	if err := g.RespondCard(1, "tao1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != 1 {
		t.Fatalf("expected victim at 1 hp, got %d", g.Players[1].HP)
	}
	if g.Phase == engine.PhaseFinished {
		t.Fatal("game should continue after rescue")
	}
}

func TestDyingJijiRescue(t *testing.T) {
	g, err := engine.NewSolo1v1("dy2", "玩家", engine.CharLiuBei, engine.CharHuaTuo)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{{ID: "red1", Kind: engine.CardSha, Suit: "H", Label: "红桃杀", Name: "杀"}}
	g.Players[1].HP = 1
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.RespondCard(1, "red1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != 1 {
		t.Fatalf("expected jiji rescue to 1 hp, got %d", g.Players[1].HP)
	}
}

func TestShuangxiongDrawAndJuedou(t *testing.T) {
	g, err := engine.NewSolo1v1("sx1", "玩家", engine.CharYanLiangWenChou, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepDraw
	g.CurrentTurn = 0
	g.SetSkillCounterForTest(0, "draw_choice_pending", 1)
	redTop := engine.Card{ID: "top-r", Kind: engine.CardSha, Suit: "H", Label: "红桃杀", Name: "杀"}
	blackHand := engine.Card{ID: "bh1", Kind: engine.CardShan, Suit: "S", Label: "黑桃闪", Name: "闪"}
	g.DrawPile = []engine.Card{redTop}
	g.Players[0].Hand = []engine.Card{blackHand}
	g.Players[1].Hand = nil
	g.SyncCounts()

	var events []engine.GameEvent
	if err := g.UseSkill(0, engine.UseSkillRequest{SkillID: engine.SkillShuangxiong}, &events); err != nil {
		t.Fatal(err)
	}
	if g.GetSkillCounterForTest(0, "shuangxiong_active") != 1 {
		t.Fatal("expected shuangxiong active after draw")
	}
	if g.GetSkillCounterForTest(0, "shuangxiong_ref_red") != 1 {
		t.Fatal("expected red reference card")
	}
	if g.TurnStep != engine.StepPlay {
		t.Fatalf("expected play phase, step=%s", g.TurnStep)
	}
	if err := g.UseSkill(0, engine.UseSkillRequest{SkillID: engine.SkillShuangxiong, CardIDs: []string{"bh1"}}, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhaseResponse {
		t.Fatalf("expected wuxiek/juedou response, phase=%s pending=%+v", g.Phase, g.Pending)
	}
	if g.Pending == nil || g.Pending.Card.Kind != engine.CardJueDou {
		t.Fatalf("expected juedou pending, got %+v", g.Pending)
	}
}

func TestWanshaBlocksJijiHeal(t *testing.T) {
	g, err := engine.NewSolo1v1("ws1", "玩家", engine.CharJiaXu, engine.CharHuaTuo)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[1].HP = 2
	g.Players[1].Hand = []engine.Card{
		{ID: "red-sha", Kind: engine.CardSha, Suit: "H", Label: "红桃杀", Name: "杀"},
	}
	g.SyncCounts()
	redSha := engine.Card{ID: "red-sha", Kind: engine.CardSha, Suit: "H", Name: "杀"}
	if g.CanUseJijiHealForTest(1, redSha) {
		t.Fatal("wansha should block jiji heal during jia xu turn")
	}
	g2, err := engine.NewSolo1v1("ws2", "玩家", engine.CharHuaTuo, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g2.Phase = engine.PhasePlaying
	g2.TurnStep = engine.StepPlay
	g2.CurrentTurn = 1
	g2.Players[0].HP = 2
	g2.Players[0].Hand = []engine.Card{redSha}
	g2.SyncCounts()
	if !g2.CanUseJijiHealForTest(0, redSha) {
		t.Fatal("jiji should work when current turn is not jia xu")
	}
}

func TestWeimuBlocksBlackTrick(t *testing.T) {
	g, err := engine.NewSolo1v1("wm1", "玩家", engine.CharGuanYu, engine.CharJiaXu)
	if err != nil {
		t.Fatal(err)
	}
	blackGuohe := engine.Card{ID: "gh1", Kind: engine.CardGuoHe, Suit: "S", Name: "过河拆桥"}
	redGuohe := engine.Card{ID: "gh2", Kind: engine.CardGuoHe, Suit: "H", Name: "过河拆桥"}
	if !g.WeimuBlocksTrickForTest(1, blackGuohe) {
		t.Fatal("weimu should block black guohe")
	}
	if g.WeimuBlocksTrickForTest(1, redGuohe) {
		t.Fatal("weimu should not block red guohe")
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{blackGuohe}
	g.Players[1].Hand = []engine.Card{{ID: "h1", Kind: engine.CardShan, Name: "闪"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlayCard(0, "gh1", 1, &events); err != engine.ErrInvalidTarget {
		t.Fatalf("expected invalid target for black guohe on weimu, got %v", err)
	}
}

func TestLuanwuShaOrDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("lw1", "玩家", engine.CharJiaXu, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[1].HP = 4
	g.Players[1].Hand = []engine.Card{
		{ID: "sha1", Kind: engine.CardSha, Name: "杀"},
	}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.ActivateLuanwuForTest(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != "skill_luanwu" {
		t.Fatalf("expected luanwu pending, got %+v", g.Pending)
	}
	if err := g.PlayLuanwuShaForTest(1, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.RequiredKind != engine.CardShan {
		t.Fatalf("expected sha response pending, got %+v", g.Pending)
	}

	g2, err := engine.NewSolo1v1("lw2", "玩家", engine.CharJiaXu, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g2.Phase = engine.PhasePlaying
	g2.TurnStep = engine.StepPlay
	g2.CurrentTurn = 0
	g2.Players[1].HP = 4
	g2.Players[1].Hand = nil
	g2.SyncCounts()
	events = nil
	if err := g2.ActivateLuanwuForTest(0, &events); err != nil {
		t.Fatal(err)
	}
	hpBefore := g2.Players[1].HP
	if err := g2.PassLuanwuForTest(1, &events); err != nil {
		t.Fatal(err)
	}
	if g2.Players[1].HP != hpBefore-1 {
		t.Fatalf("expected luanwu pass damage, hp=%d", g2.Players[1].HP)
	}
	if g2.CurrentTurn != 0 || g2.TurnStep != engine.StepPlay {
		t.Fatalf("expected jia xu to continue play, turn=%d step=%s", g2.CurrentTurn, g2.TurnStep)
	}
}

func TestLeijiAfterShanBlackJudge(t *testing.T) {
	g, err := engine.NewSolo1v1("lj1", "玩家", engine.CharGuanYu, engine.CharZhangJiao)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{{ID: "shan1", Kind: engine.CardShan, Name: "闪"}}
	g.DrawPile = []engine.Card{{ID: "judge-black", Suit: "S", Rank: 5, Label: "黑桃5", Name: "5"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.RespondCardForTest(1, "shan1", &events); err != nil {
		t.Fatal(err)
	}
	if g.PendingResponseModeForTest() != "skill_leiji_offer" {
		t.Fatalf("expected leiji offer, mode=%q", g.PendingResponseModeForTest())
	}
	oppHP := g.Players[0].HP
	if err := g.StartLeijiJudgeForTest(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[0].HP != oppHP-2 {
		t.Fatalf("expected leiji 2 damage, opp hp=%d", g.Players[0].HP)
	}
}

func TestGuidaoReplacesWithBlackCard(t *testing.T) {
	g, err := engine.NewSolo1v1("gd1", "玩家", engine.CharZhangJiao, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "black1", Kind: engine.CardSha, Suit: "S", Label: "黑桃7", Name: "杀"},
	}
	g.SetLeijiContextForTest(0)
	g.SyncCounts()
	var events []engine.GameEvent
	redJudge := engine.Card{ID: "j-red", Suit: "H", Label: "红桃3", Name: "3"}
	if err := g.AfterJudgeFlipForTest(1, redJudge, &events); err != nil {
		t.Fatal(err)
	}
	if g.PendingResponseModeForTest() != "skill_guidao" {
		t.Fatalf("expected guidao window, mode=%q", g.PendingResponseModeForTest())
	}
	if err := g.ApplyGuidaoReplaceForTest(0, "black1", &events); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ev := range events {
		if ev.Type == "guidao_replace" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected guidao_replace event")
	}
}

func TestJueqingSkipsDying(t *testing.T) {
	g, err := engine.NewSolo1v1("jq1", "玩家", engine.CharZhangChunhua, engine.CharHuaTuo)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{{ID: "tao1", Kind: engine.CardTao, Name: "桃"}}
	g.Players[1].HP = 1
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if !g.IsFinished() {
		t.Fatal("expected game over from jueqing lose hp")
	}
	if g.PendingResponseModeForTest() == engine.ResponseModeDying {
		t.Fatal("jueqing should skip dying rescue")
	}
	if g.WinnerIndex == nil || *g.WinnerIndex != 0 {
		t.Fatalf("expected zhang chunhua win, winner=%v", g.WinnerIndex)
	}
}

func TestShangshiDrawsWhenLosingHandAtOneHP(t *testing.T) {
	g, err := engine.NewSolo1v1("ss1", "玩家", engine.CharZhangChunhua, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].HP = 1
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.DrawPile = []engine.Card{{ID: "d1", Kind: engine.CardShan, Name: "闪"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 1 {
		t.Fatalf("expected shangshi draw after playing card at 1 hp, hand=%d", len(g.Players[0].Hand))
	}
	found := false
	for _, ev := range events {
		if ev.SkillID == "shangshi" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected shangshi skill event")
	}
}

func TestShangshiNoDrawAboveOneHP(t *testing.T) {
	g, err := engine.NewSolo1v1("ss2", "玩家", engine.CharZhangChunhua, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].HP = 2
	g.Players[0].Hand = []engine.Card{{ID: "sha1", Kind: engine.CardSha, Name: "杀"}}
	g.DrawPile = []engine.Card{{ID: "d1", Kind: engine.CardShan, Name: "闪"}}
	g.SyncCounts()
	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 0 {
		t.Fatalf("shangshi should not draw above 1 hp, hand=%d", len(g.Players[0].Hand))
	}
}

