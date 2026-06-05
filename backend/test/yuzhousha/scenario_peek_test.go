package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

// 观星 pending 时 AI 应能完成分配并离开 response 窗（sim forceProgress 兜底同路径）。
func TestScenario_PeekDeck_AIResolvesPending(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-peek-ai", "玩家", engine.CharZhugeLiang, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].IsAI = true
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPrepare
	g.CurrentTurn = 0
	g.DrawPile = append([]engine.Card{
		{ID: "pk1", Kind: engine.CardSha, Name: "杀"},
		{ID: "pk2", Kind: engine.CardShan, Name: "闪"},
		{ID: "pk3", Kind: engine.CardTao, Name: "桃"},
	}, g.DrawPile...)

	var events []engine.GameEvent
	if err := g.StartPeekDeck(0, engine.SkillGuanxing, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != engine.ResponseModePeekDeck {
		t.Fatalf("expected peek_deck pending, got %+v", g.Pending)
	}

	if !engine.RunAIActionStep(g, &events) {
		t.Fatal("AI should resolve peek_deck")
	}
	if g.Pending != nil && g.Pending.ResponseMode == engine.ResponseModePeekDeck {
		t.Fatal("peek_deck should be cleared after AI step")
	}
	if g.Phase == engine.PhaseResponse {
		t.Fatalf("should leave response phase after peek, still pending=%+v", g.Pending)
	}
}
