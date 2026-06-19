package engine_test

import (
	"testing"
	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func TestDebugLeBu(t *testing.T) {
	g, err := engine.NewSolo1v1("g9", "玩家", engine.CharLiuBei, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "lb-1", Kind: engine.CardLeBu, Name: "乐不思蜀"}}
	g.Players[1].Hand = nil
	g.DrawPile = []engine.Card{
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀", Suit: "H", Rank: 1},
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
	t.Logf("After PlayCard: Phase=%s, TurnStep=%s, CurrentTurn=%d", g.Phase, g.TurnStep, g.CurrentTurn)
	t.Logf("Player1 JudgeArea: %+v", g.Players[1].JudgeArea)
	t.Logf("Player1 SkipPlay: %v", g.Players[1].SkipPlay)
	
	if g.Phase == engine.PhaseResponse {
		t.Fatal("expected lebu to apply immediately without wuxiek window")
	}
	if !g.HasJudgeKindForTest(1, engine.CardLeBu) || !g.Players[1].SkipPlay {
		t.Fatal("expected lebu placed in judge zone")
	}
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	t.Logf("After EndPlay: Phase=%s, TurnStep=%s, CurrentTurn=%d", g.Phase, g.TurnStep, g.CurrentTurn)
	t.Logf("Player1 SkipPlay: %v", g.Players[1].SkipPlay)
	t.Logf("Player1 JudgeArea: %+v", g.Players[1].JudgeArea)
	
	if g.CurrentTurn != 0 {
		t.Logf("Expected CurrentTurn=0, got %d", g.CurrentTurn)
	}
}
