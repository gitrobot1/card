package engine_test

import (
	"errors"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func TestSolo1v1TurnFlow(t *testing.T) {
	g, err := engine.NewSolo1v1("g1", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhasePlaying || g.TurnStep != engine.StepPlay {
		t.Fatalf("expected play step, got phase=%s step=%s", g.Phase, g.TurnStep)
	}
	if len(g.Players[0].Hand) != engine.InitialHandSize+engine.DrawPerTurn {
		t.Fatalf("expected %d cards after opening draw, got %d", engine.InitialHandSize+engine.DrawPerTurn, len(g.Players[0].Hand))
	}
}

func TestShaShanResolution(t *testing.T) {
	g, err := engine.NewSolo1v1("g2", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = []engine.Card{{ID: "shan-1", Kind: engine.CardShan, Name: "闪"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	g.Players[0].ShaUsedThisTurn = false

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhaseResponse {
		t.Fatalf("expected response phase, got %s", g.Phase)
	}
	if err := g.RespondShan(1, "shan-1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhasePlaying || g.CurrentTurn != 0 {
		t.Fatalf("expected source continue play, phase=%s turn=%d", g.Phase, g.CurrentTurn)
	}
	if g.Players[1].HP != engine.DefaultMaxHP {
		t.Fatalf("expected no damage, hp=%d", g.Players[1].HP)
	}
}

func TestManualDiscardPhase(t *testing.T) {
	g, err := engine.NewSolo1v1("g4", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "c-0", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-1", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-2", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-3", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-4", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-5", Kind: engine.CardShan, Name: "闪"},
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.TurnStep != engine.StepDiscard {
		t.Fatalf("expected discard step, got %s", g.TurnStep)
	}
	if err := g.DiscardCards(0, []string{"c-0", "c-1"}, &events); err != nil {
		t.Fatal(err)
	}
	if g.TurnStep == engine.StepDiscard {
		t.Fatalf("expected turn ended after batch discard, step=%s", g.TurnStep)
	}
	if len(g.Players[0].Hand) != g.Players[0].HP {
		t.Fatalf("expected hand limit %d, got %d", g.Players[0].HP, len(g.Players[0].Hand))
	}
}

func TestDiscardCardsWrongCount(t *testing.T) {
	g, err := engine.NewSolo1v1("g5", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "c-0", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-1", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-2", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-3", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-4", Kind: engine.CardShan, Name: "闪"},
		{ID: "c-5", Kind: engine.CardShan, Name: "闪"},
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepDiscard
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.DiscardCards(0, []string{"c-0"}, &events); !errors.Is(err, engine.ErrInvalidDiscardCount) {
		t.Fatalf("expected invalid discard count, got %v", err)
	}
	if len(g.Players[0].Hand) != 6 {
		t.Fatalf("expected hand unchanged, got %d", len(g.Players[0].Hand))
	}
}

func TestShaHitFinishesAtZeroHP(t *testing.T) {
	g, err := engine.NewSolo1v1("g3", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = nil
	g.Players[1].HP = 1
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeDying {
		t.Fatalf("expected dying window, pending=%+v", g.Pending)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhaseFinished || g.WinnerIndex == nil || *g.WinnerIndex != 0 {
		t.Fatalf("expected player 0 win, phase=%s winner=%v", g.Phase, g.WinnerIndex)
	}
}

func TestWuZhongDrawsTwoCards(t *testing.T) {
	g, err := engine.NewSolo1v1("g6", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "wz-1", Kind: engine.CardWuZhong, Name: "无中生有"}}
	g.DrawPile = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "shan-1", Kind: engine.CardShan, Name: "闪"},
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "wz-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.Players[0].Hand) != 2 {
		t.Fatalf("expected two drawn cards, got %d", len(g.Players[0].Hand))
	}
}

func TestNanManRequiresShaResponse(t *testing.T) {
	g, err := engine.NewSolo1v1("g7", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "nm-1", Kind: engine.CardNanMan, Name: "南蛮入侵"}}
	g.Players[1].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "nm-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.RequiredKind != engine.CardSha {
		t.Fatalf("expected sha response, pending=%+v", g.Pending)
	}
	if err := g.RespondCard(1, "sha-1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != g.Players[1].MaxHP {
		t.Fatalf("expected no damage after sha response, hp=%d", g.Players[1].HP)
	}
}

func TestTanNangTakesOpponentCard(t *testing.T) {
	g, err := engine.NewSolo1v1("g8", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "tn-1", Kind: engine.CardTanNang, Name: "顺手牵羊"}}
	g.Players[1].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "tn-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.Players[0].Hand) != 1 || g.Players[0].Hand[0].ID != "sha-1" {
		t.Fatalf("expected stolen card in player hand, hand=%+v", g.Players[0].Hand)
	}
	if len(g.Players[1].Hand) != 0 {
		t.Fatalf("expected opponent hand empty, got %d", len(g.Players[1].Hand))
	}
}

func TestLeBuSkipsPlayPhase(t *testing.T) {
	g, err := engine.NewSolo1v1("g9", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "lb-1", Kind: engine.CardLeBu, Name: "乐不思蜀"}}
	g.Players[1].Hand = nil
	g.DrawPile = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "shan-1", Kind: engine.CardShan, Name: "闪"},
		{ID: "tao-1", Kind: engine.CardTao, Name: "桃"},
		{ID: "sha-2", Kind: engine.CardSha, Name: "杀"},
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "lb-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		t.Fatal("expected lebu to apply immediately without wuxiek window")
	}
	if !g.HasJudgeKindForTest(1, engine.CardLeBu) || !g.Players[1].SkipPlay {
		t.Fatal("expected lebu placed in judge zone")
	}
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatal(err)
		}
	}
	if g.CurrentTurn != 0 {
		t.Fatalf("expected opponent skipped back to player 0, current=%d", g.CurrentTurn)
	}
	if g.Players[1].SkipPlay {
		t.Fatal("expected skip flag consumed")
	}
}

func TestEquipWeaponExtendsShaRange(t *testing.T) {
	g, err := engine.NewSolo1v1("g10", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "w2-1", Kind: engine.CardWeapon2, Name: "长枪"},
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = nil
	g.Players[1].PlusHorse = &engine.Card{ID: "horse-1", Kind: engine.CardPlusHorse, Name: "+1马"}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); !errors.Is(err, engine.ErrInvalidTarget) {
		t.Fatalf("expected invalid target before weapon, got %v", err)
	}
	if err := g.PlayCard(0, "w2-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
}

func TestBaguaRedJudgeBlocksSha(t *testing.T) {
	g, err := engine.NewSolo1v1("g11", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = nil
	g.Players[1].Armor = &engine.Card{ID: "bagua-1", Kind: engine.CardArmor, Name: "八卦阵"}
	g.DrawPile = []engine.Card{{ID: "judge-1", Kind: engine.CardSha, Suit: "H", Label: "红桃2", Name: "杀"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.TryBaguaJudge(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhasePlaying || g.Players[1].HP != engine.DefaultMaxHP {
		t.Fatalf("expected bagua to block sha, phase=%s hp=%d", g.Phase, g.Players[1].HP)
	}
}

func TestBaguaBlackJudgeStillTakesDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("g11b", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = nil
	g.Players[1].Armor = &engine.Card{ID: "bagua-1", Kind: engine.CardArmor, Name: "八卦阵"}
	g.DrawPile = []engine.Card{{ID: "judge-1", Kind: engine.CardSha, Suit: "S", Label: "黑桃2", Name: "杀"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.TryBaguaJudge(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhaseResponse || !g.Pending.BaguaUsed {
		t.Fatalf("expected still awaiting shan after black judge, phase=%s bagua=%v", g.Phase, g.Pending.BaguaUsed)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-1 {
		t.Fatalf("expected 1 damage after failed bagua, hp=%d", g.Players[1].HP)
	}
}

func TestJiuNextShaAddsOneDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("g12", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "jiu-1", Kind: engine.CardJiu, Name: "酒"},
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = nil
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "jiu-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if !g.Players[0].Drunk {
		t.Fatal("expected drunk state after jiu")
	}
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[0].Drunk {
		t.Fatal("expected drunk state consumed by sha")
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-2 {
		t.Fatalf("expected 2 damage, hp=%d", g.Players[1].HP)
	}
}

func TestNanManDoesNotRequireTargetIndex(t *testing.T) {
	g, err := engine.NewSolo1v1("g13", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "nm-1", Kind: engine.CardNanMan, Name: "南蛮入侵"}}
	g.Players[1].Hand = nil
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "nm-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.TargetIndex != 1 || g.Pending.RequiredKind != engine.CardSha {
		t.Fatalf("expected opponent sha response, pending=%+v", g.Pending)
	}
}

func TestGuoHeCanDiscardEquipment(t *testing.T) {
	g, err := engine.NewSolo1v1("g14", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "gh-1", Kind: engine.CardGuoHe, Name: "过河拆桥"}}
	g.Players[1].Hand = nil
	g.Players[1].Weapon = &engine.Card{ID: "w-1", Kind: engine.CardWeapon3, Name: "强弩"}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	err = g.PlayCardWithTarget(0, "gh-1", engine.PlayTarget{SeatIndex: 1, Zone: engine.EquipWeapon, CardID: "w-1"}, &events)
	if err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatal(err)
		}
	}
	if g.Players[1].Weapon != nil {
		t.Fatal("expected weapon discarded")
	}
	if len(g.DiscardPile) != 2 {
		t.Fatalf("expected trick and weapon in discard pile, got %d", len(g.DiscardPile))
	}
}

func TestTanNangCanTakeJudgementCard(t *testing.T) {
	g, err := engine.NewSolo1v1("g15", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "tn-1", Kind: engine.CardTanNang, Name: "顺手牵羊"}}
	g.Players[1].Hand = nil
	g.Players[1].JudgeArea = []engine.Card{{ID: "lb-1", Kind: engine.CardLeBu, Name: "乐不思蜀"}}
	g.Players[1].SkipPlay = true
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	err = g.PlayCardWithTarget(0, "tn-1", engine.PlayTarget{SeatIndex: 1, Zone: "judge", CardID: "lb-1"}, &events)
	if err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatal(err)
		}
	}
	if g.HasJudgeKindForTest(1, engine.CardLeBu) || g.Players[1].SkipPlay {
		t.Fatalf("expected judgement cleared, judge=%+v skip=%v", g.Players[1].JudgeArea, g.Players[1].SkipPlay)
	}
	if len(g.Players[0].Hand) != 1 || g.Players[0].Hand[0].ID != "lb-1" {
		t.Fatalf("expected lebu in player hand, hand=%+v", g.Players[0].Hand)
	}
}

func TestWuxiekCancelsGuoHe(t *testing.T) {
	g, err := engine.NewSolo1v1("g16", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "gh-1", Kind: engine.CardGuoHe, Name: "过河拆桥"}}
	g.Players[1].Hand = []engine.Card{{ID: "wx-1", Kind: engine.CardWuxiek, Name: "无懈可击"}}
	g.Players[1].Weapon = &engine.Card{ID: "w-1", Kind: engine.CardWeapon3, Name: "强弩"}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	err = g.PlayCardWithTarget(0, "gh-1", engine.PlayTarget{SeatIndex: 1, Zone: engine.EquipWeapon, CardID: "w-1"}, &events)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.RespondWuxiek(1, "wx-1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].Weapon == nil {
		t.Fatal("expected weapon kept after wuxiek")
	}
}

func TestWuxiekSelfOnNanMan(t *testing.T) {
	g, err := engine.NewSolo1v1("g17", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "nm-1", Kind: engine.CardNanMan, Name: "南蛮入侵"}}
	g.Players[1].Hand = []engine.Card{{ID: "wx-1", Kind: engine.CardWuxiek, Name: "无懈可击"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "nm-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if !g.Pending.AllowWuxiek {
		t.Fatal("expected nanman to allow wuxiek on self")
	}
	if err := g.RespondWuxiek(1, "wx-1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP {
		t.Fatalf("expected no damage after self wuxiek, hp=%d", g.Players[1].HP)
	}
}

func TestWuxiekLebuBeforeJudge(t *testing.T) {
	g, err := engine.NewSolo1v1("g18", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "lb-1", Kind: engine.CardLeBu, Name: "乐不思蜀"}}
	g.Players[1].Hand = []engine.Card{{ID: "wx-1", Kind: engine.CardWuxiek, Name: "无懈可击"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "lb-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhasePlaying {
		t.Fatalf("expected lebu placed without wuxiek window, phase=%s", g.Phase)
	}
	if !g.HasJudgeKindForTest(1, engine.CardLeBu) || !g.Players[1].SkipPlay {
		t.Fatal("expected lebu in judge zone after play")
	}
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeWuxiekLebu {
		t.Fatalf("expected wuxiek window before lebu judge, pending=%+v phase=%s", g.Pending, g.Phase)
	}
	if err := g.RespondWuxiek(1, "wx-1", &events); err != nil {
		t.Fatal(err)
	}
	if g.HasJudgeKindForTest(1, engine.CardLeBu) || g.Players[1].SkipPlay {
		t.Fatalf("expected lebu cancelled before judge, judge=%+v skip=%v", g.Players[1].JudgeArea, g.Players[1].SkipPlay)
	}
	if g.TurnStep != engine.StepPlay {
		t.Fatalf("expected play step after cancelling lebu, step=%s", g.TurnStep)
	}
}

func TestLebuCannotWuxiekOnPlay(t *testing.T) {
	g, err := engine.NewSolo1v1("g18b", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "lb-1", Kind: engine.CardLeBu, Name: "乐不思蜀"}}
	g.Players[1].Hand = []engine.Card{{ID: "wx-1", Kind: engine.CardWuxiek, Name: "无懈可击"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "lb-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		t.Fatal("lebu should not open wuxiek window on play")
	}
	if len(g.Players[1].Hand) != 1 {
		t.Fatalf("expected opponent wuxiek unused, hand=%+v", g.Players[1].Hand)
	}
}

func TestJiuShaDealsTwoDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("g19", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "jiu-1", Kind: engine.CardJiu, Name: "酒"},
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = nil
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

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
}

func TestLianNuAllowsMultipleSha(t *testing.T) {
	g, err := engine.NewSolo1v1("g20", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-2", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[0].Weapon = &engine.Card{ID: "w1", Kind: engine.CardWeapon1, Name: "诸葛连弩"}
	g.Players[1].Hand = nil
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[0].ShaUsedThisTurn {
		t.Fatal("liannu should not mark sha used")
	}
	if err := g.PlayCard(0, "sha-2", 1, &events); err != nil {
		t.Fatal(err)
	}
}

func TestQingGangIgnoresBagua(t *testing.T) {
	g, err := engine.NewSolo1v1("g21", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[0].Weapon = &engine.Card{ID: "w2", Kind: engine.CardWeapon2, Name: "青釭剑"}
	g.Players[1].Hand = nil
	g.Players[1].Armor = &engine.Card{ID: "bagua", Kind: engine.CardArmor, Name: "八卦阵"}
	g.DrawPile = []engine.Card{{ID: "j1", Kind: engine.CardSha, Suit: "H", Label: "红桃2", Name: "杀"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if !g.Pending.IgnoreArmor {
		t.Fatal("expected qinggang pending ignore armor")
	}
	if err := g.TryBaguaJudge(1, &events); !errors.Is(err, engine.ErrInvalidCard) {
		t.Fatalf("expected bagua blocked by qinggang, got %v", err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-1 {
		t.Fatalf("expected damage through ignored bagua, hp=%d", g.Players[1].HP)
	}
}

func TestGuanYuFollowUpSha(t *testing.T) {
	g, err := engine.NewSolo1v1("g22", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "sha-2", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[0].Weapon = &engine.Card{ID: "w3", Kind: engine.CardWeapon3, Name: "青龙偃月刀"}
	g.Players[1].Hand = []engine.Card{{ID: "shan-1", Kind: engine.CardShan, Name: "闪"}}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.RespondCard(1, "shan-1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeGuanYuFollow {
		t.Fatalf("expected guanyu follow pending, got %+v", g.Pending)
	}
	if err := g.PlayCard(0, "sha-2", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-1 {
		t.Fatalf("expected follow-up damage, hp=%d", g.Players[1].HP)
	}
}

func TestQilinBowDiscardsHorse(t *testing.T) {
	g, err := engine.NewSolo1v1("g23", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[0].Weapon = &engine.Card{ID: "w5", Kind: engine.CardWeapon5, Name: "麒麟弓"}
	g.Players[1].Hand = nil
	g.Players[1].MinusHorse = &engine.Card{ID: "horse-1", Kind: engine.CardMinusHorse, Name: "-1马"}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeQilinBow {
		t.Fatalf("expected qilin pending, got %+v", g.Pending)
	}
	if err := g.QilinDiscardHorseForTest(0, engine.EquipMinusHorse, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].MinusHorse != nil {
		t.Fatal("expected minus horse discarded")
	}
}

func TestBingliangSkipsDrawPhase(t *testing.T) {
	g, err := engine.NewSolo1v1("g24", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "bl-1", Kind: engine.CardBingLiang, Name: "兵粮寸断"}}
	g.Players[1].Hand = nil
	handBefore := len(g.Players[1].Hand)
	g.DrawPile = []engine.Card{
		{ID: "draw-1", Kind: engine.CardSha, Name: "杀"},
		{ID: "draw-2", Kind: engine.CardShan, Name: "闪"},
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "bl-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if !g.HasJudgeKindForTest(1, engine.CardBingLiang) || !g.Players[1].SkipDraw {
		t.Fatal("expected bingliang in judge zone with skip_draw")
	}
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase == engine.PhaseResponse {
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.Players[1].Hand) != handBefore {
		t.Fatalf("expected draw skipped, hand=%d", len(g.Players[1].Hand))
	}
	if g.Players[1].SkipDraw {
		t.Fatal("expected skip_draw consumed")
	}
}

func TestShandianStrikeDamages(t *testing.T) {
	g, err := engine.NewSolo1v1("g25", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[1].JudgeArea = []engine.Card{{ID: "sd-1", Kind: engine.CardShanDian, Name: "闪电"}}
	g.Players[1].Hand = nil
	g.DrawPile = []engine.Card{{ID: "judge-1", Kind: engine.CardSha, Suit: "S", Rank: 5, Label: "黑桃5", Name: "杀"}}
	g.CurrentTurn = 1
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay

	var events []engine.GameEvent
	g.BeginTurnForTest(&events)
	if g.Phase == engine.PhaseResponse && g.Pending != nil && g.Pending.ResponseMode == engine.ResponseModeWuxiekShandian {
		if err := g.PassResponse(1, &events); err != nil {
			t.Fatal(err)
		}
	}
	if g.Players[1].HP != engine.DefaultMaxHP-3 {
		t.Fatalf("expected 3 lightning damage, hp=%d", g.Players[1].HP)
	}
	if g.HasJudgeKindForTest(1, engine.CardShanDian) {
		t.Fatal("expected lightning removed after strike")
	}
}

func TestWuguDistributesCards(t *testing.T) {
	g, err := engine.NewSolo1v1("g26", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "wg-1", Kind: engine.CardWuGu, Name: "五谷丰登"}}
	g.Players[1].Hand = nil
	g.DrawPile = []engine.Card{
		{ID: "c1", Kind: engine.CardSha, Name: "杀"},
		{ID: "c2", Kind: engine.CardShan, Name: "闪"},
	}
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0

	var events []engine.GameEvent
	if err := g.PlayCard(0, "wg-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeWuguPick {
		t.Fatalf("expected wugu pick pending, got %+v", g.Pending)
	}
	if err := g.PickWuguCardForTest(0, "c1", &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModeWuguPick {
		t.Fatalf("expected second wugu pick, got %+v", g.Pending)
	}
	if err := g.PickWuguCardForTest(1, "c2", &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 1 || len(g.Players[1].Hand) != 1 {
		t.Fatalf("expected each player got one card, hands=%d/%d", len(g.Players[0].Hand), len(g.Players[1].Hand))
	}
}

func TestDeckCanExceed52Cards(t *testing.T) {
	deck := engine.NewBasicDeck()
	if len(deck) <= 52 {
		t.Fatalf("expected deck larger than 52 cards, got %d", len(deck))
	}
}

func TestNewDeckForMode_3v3NoShanDian(t *testing.T) {
	deck := engine.NewDeckForMode(mode.Solo3v3)
	if len(deck) != 63 {
		t.Fatalf("3v3 deck size=%d want 63", len(deck))
	}
	for _, c := range deck {
		if c.Kind == engine.CardShanDian {
			t.Fatalf("3v3 deck contains shandian: %s", c.ID)
		}
	}
}

func TestNewDeckForMode_Identity8LargeDeck(t *testing.T) {
	deck := engine.NewDeckForMode(mode.SoloIdentity8)
	if len(deck) != 90 {
		t.Fatalf("identity_8 deck size=%d want 90", len(deck))
	}
	for _, c := range deck {
		if c.Kind == engine.CardShanDian {
			t.Fatalf("identity_8 deck contains shandian: %s", c.ID)
		}
	}
}

func TestNewDeckForMode_DdzExtraSha(t *testing.T) {
	deck := engine.NewDeckForMode(mode.Solo3pDdz)
	if len(deck) != 67 {
		t.Fatalf("ddz deck size=%d want 67", len(deck))
	}
	sha := 0
	for _, c := range deck {
		if c.Kind == engine.CardSha {
			sha++
		}
	}
	if sha != 13 {
		t.Fatalf("ddz sha=%d want 13", sha)
	}
}

func TestNewDeckForMode_Identity5Tuned(t *testing.T) {
	deck := engine.NewDeckForMode(mode.SoloIdentity5)
	if len(deck) != 67 {
		t.Fatalf("identity_5 deck size=%d want 67", len(deck))
	}
	sha, tao, shandian := 0, 0, 0
	for _, c := range deck {
		switch c.Kind {
		case engine.CardSha:
			sha++
		case engine.CardTao:
			tao++
		case engine.CardShanDian:
			shandian++
		}
	}
	if sha != 12 || tao != 5 || shandian != 1 {
		t.Fatalf("identity_5 sha=%d tao=%d shandian=%d want 12/5/1", sha, tao, shandian)
	}
}
