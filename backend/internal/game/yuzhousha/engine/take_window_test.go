package engine

import (
	"testing"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func TestTakeWindow_FankuiTakeOne(t *testing.T) {
	g := newTakeWindowFankuiGame(t)
	var events []GameEvent
	if err := g.TakeOne(1, ZoneHand, "extra", &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[1].Hand) != 1 || g.Players[1].Hand[0].ID != "extra" {
		t.Fatalf("actor hand = %+v, want extra", g.Players[1].Hand)
	}
	if len(g.Players[0].Hand) != 0 {
		t.Fatalf("subject hand should be empty, got %+v", g.Players[0].Hand)
	}
}

func TestTakeWindow_PassEarly(t *testing.T) {
	g := newTakeWindowTuxiGame(t, 3)
	g.Players[1].Hand = []Card{{ID: "h1", Label: "手牌"}}
	var events []GameEvent
	if err := g.TakeOne(0, ZoneWeapon, "w1", &events); err != nil {
		t.Fatal(err)
	}
	if g.takeWindow == nil {
		t.Fatal("window should stay open after first take")
	}
	if err := g.PassTake(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.takeWindow != nil || g.Pending != nil {
		t.Fatal("take window should be closed after pass")
	}
}

func TestTakeWindow_AISweeps(t *testing.T) {
	g := newTakeWindowTuxiGame(t, 2)
	g.Players[0].IsAI = true
	var events []GameEvent
	g.AutoTakeWindow(0, &events)
	if g.takeWindow != nil {
		t.Fatal("take window should be closed after AI sweep")
	}
	if len(g.Players[0].Hand) < 1 {
		t.Fatalf("AI should take at least weapon into hand, hand=%+v", g.Players[0].Hand)
	}
}

func newTakeWindowFankuiGame(t *testing.T) *Game {
	t.Helper()
	g := &Game{
		Phase: PhaseResponse,
		Players: []Player{
			{Index: 0, Name: "A", Hand: []Card{{ID: "extra", Kind: CardSha, Label: "杀"}}},
			{Index: 1, Name: "B", Character: buildCharacter(CharSimaYi)},
		},
	}
	if err := g.OpenTakeWindow(TakeWindowConfig{
		SkillID:         skill.IDFankui,
		ResponseMode:    ResponseModeSkillFankui,
		ActorSeat:       1,
		SubjectSeat:     0,
		OriginSeat:      0,
		MaxTake:         1,
		Destination:     TakeDestination{Zone: ZoneHand, Seat: 1},
		EventType:       "fankui_take",
		SkillEventLabel: "反馈",
		OnComplete:      fankuiTakeComplete,
	}, nil); err != nil {
		t.Fatal(err)
	}
	return g
}

func newTakeWindowTuxiGame(t *testing.T, maxTake int) *Game {
	t.Helper()
	g := &Game{
		Phase:       PhasePlaying,
		TurnStep:    StepDraw,
		CurrentTurn: 0,
		Players: []Player{
			{Index: 0, Name: "Attacker", Character: buildCharacter(CharZhangLiao)},
			{Index: 1, Name: "Victim", Weapon: &Card{ID: "w1", Kind: CardWeapon1, Label: "刀"}},
		},
	}
	g.setSkillCounter(0, counterTuxiDrawSkip, maxTake)
	var events []GameEvent
	actor := 0
	if err := g.OpenTakeWindow(TakeWindowConfig{
		SkillID:          skill.IDTuxi,
		ResponseMode:     ResponseModeSkillTuxi,
		ActorSeat:        0,
		SubjectSeat:      1,
		OriginSeat:       0,
		MaxTake:          maxTake,
		Destination:      TakeDestination{Zone: ZoneHand, Seat: 0},
		EventType:        "tuxi_take",
		SkillEventLabel:  "突袭",
		PassClosesWindow: true,
		OnComplete: func(g *Game, events *[]GameEvent) error {
			return g.finishTuxi(actor, events)
		},
	}, &events); err != nil {
		t.Fatal(err)
	}
	return g
}
