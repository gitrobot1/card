package engine_test

import (
	"testing"
	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func TestDebugWuxiekLebu(t *testing.T) {
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
	t.Logf("After PlayCard: Phase=%s, JudgeArea=%+v, SkipPlay=%v", g.Phase, g.Players[1].JudgeArea, g.Players[1].SkipPlay)
	
	if err := g.EndPlay(0, &events); err != nil {
		t.Fatal(err)
	}
	t.Logf("After EndPlay: Phase=%s, Pending=%+v", g.Phase, g.Pending)
	
	if g.Phase == engine.PhaseResponse && g.Pending != nil {
		t.Logf("ResponseMode=%s", g.Pending.ResponseMode)
		
		if err := g.RespondWuxiek(1, "wx-1", &events); err != nil {
			t.Fatal(err)
		}
		t.Logf("After RespondWuxiek: JudgeArea=%+v, SkipPlay=%v", g.Players[1].JudgeArea, g.Players[1].SkipPlay)
		t.Logf("After RespondWuxiek: Phase=%s, TurnStep=%s", g.Phase, g.TurnStep)
	}
}
