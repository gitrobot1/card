package engine

import (
	"testing"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func TestDiscardWindow_PojunCampOne(t *testing.T) {
	g := newPojunDiscardGame(t, 1)
	var events []GameEvent
	if err := g.DiscardOne(1, "c1", &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[1].CampCards) != 1 || g.Players[1].CampCards[0].ID != "c2" {
		t.Fatalf("should discard only required count, camp=%+v", g.Players[1].CampCards)
	}
	if g.discardWindow != nil || g.Pending != nil {
		t.Fatal("discard window should be closed")
	}
}

func TestDiscardWindow_PojunCampPartial(t *testing.T) {
	g := newPojunDiscardGame(t, 2)
	var events []GameEvent
	if err := g.DiscardOne(1, "c1", &events); err != nil {
		t.Fatal(err)
	}
	if g.discardWindow == nil || g.discardWindow.remaining() != 1 {
		t.Fatalf("want 1 remaining, got window=%v", g.discardWindow)
	}
	if err := g.DiscardOne(1, "c2", &events); err != nil {
		t.Fatal(err)
	}
	if g.discardWindow != nil {
		t.Fatal("window should close after required discards")
	}
}

func TestDiscardWindow_AISweeps(t *testing.T) {
	g := newPojunDiscardGame(t, 2)
	g.Players[1].IsAI = true
	var events []GameEvent
	g.AutoDiscardWindow(1, &events)
	if g.discardWindow != nil {
		t.Fatal("AI should finish discard window")
	}
	if len(g.Players[1].CampCards) != 0 {
		t.Fatalf("camp should be empty, got %d", len(g.Players[1].CampCards))
	}
}

func newPojunDiscardGame(t *testing.T, need int) *Game {
	t.Helper()
	g := &Game{
		Players: []Player{
			{Index: 0, Name: "Attacker"},
			{Index: 1, Name: "Victim", CampCards: []Card{
				{ID: "c1", Label: "营牌1"},
				{ID: "c2", Label: "营牌2"},
			}},
		},
	}
	var events []GameEvent
	if err := g.OpenDiscardWindow(DiscardWindowConfig{
		SkillID:      skill.IDPojun,
		ResponseMode: ResponseModeSkillPojunDiscard,
		ActorSeat:    1,
		SourceZone:   ZoneCamp,
		MinDiscard:   need,
		MaxDiscard:   need,
		Message:      "须弃营",
		EventType:    "pojun_discard",
		OnEachDiscard: func(g *Game, card Card, events *[]GameEvent) error {
			*events = append(*events, GameEvent{
				Type:        "pojun_discard",
				PlayerIndex: 1,
				Card:        &card,
			})
			return nil
		},
		OnComplete: func(g *Game, events *[]GameEvent) error {
			g.Pending = nil
			g.Phase = PhasePlaying
			return nil
		},
	}, &events); err != nil {
		t.Fatal(err)
	}
	return g
}
