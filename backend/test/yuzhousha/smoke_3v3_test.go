package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func TestSmoke_3v3_AllHeroesBootstrap(t *testing.T) {
	heroes := engine.HeroesCatalog()
	if len(heroes) < 6 {
		t.Fatal("need at least 6 heroes for 3v3 smoke")
	}
	others := make([]string, 0, 5)
	for _, h := range heroes {
		if h.ID != heroes[0].ID {
			others = append(others, h.ID)
		}
		if len(others) >= 5 {
			break
		}
	}
	for _, h := range heroes {
		lineup, err := pick3v3Lineup(h.ID, others)
		if err != nil {
			t.Fatalf("pick lineup for %s: %v", h.ID, err)
		}
		g, err := engine.NewSolo3v3WithHeroes("smoke-3v3-"+h.ID, lineup)
		if err != nil {
			t.Fatalf("NewSolo3v3WithHeroes(%s): %v", h.ID, err)
		}
		if g.Mode != mode.Solo3v3 || len(g.Players) != 6 {
			t.Fatalf("unexpected game mode/players for %s", h.ID)
		}
		if g.Players[0].Character.ID != h.ID {
			t.Fatalf("seat0 hero mismatch for %s", h.ID)
		}
		assertGameInvariants(t, g)
	}
}

func TestSmoke_3v3_DeckNoShanDian(t *testing.T) {
	lineup, err := pick3v3Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := engine.NewSolo3v3WithHeroes("smoke-3v3-deck", lineup)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range allCardsInGame(g) {
		if c.Kind == engine.CardShanDian {
			t.Fatalf("3v3 deck must not contain shandian, found %s", c.ID)
		}
	}
	want := mode.DeckProfileFor(mode.Solo3v3).TotalCards()
	if got := countCardsInPlay(g); got != want {
		t.Fatalf("3v3 card total=%d want %d", got, want)
	}
}

func allCardsInGame(g *engine.Game) []engine.Card {
	out := append([]engine.Card(nil), g.DrawPile...)
	out = append(out, g.DiscardPile...)
	for _, p := range g.Players {
		out = append(out, p.Hand...)
		out = append(out, p.JudgeArea...)
		for _, slot := range []*engine.Card{p.Weapon, p.Armor, p.PlusHorse, p.MinusHorse} {
			if slot != nil {
				out = append(out, *slot)
			}
		}
	}
	return out
}

func TestSmoke_3v3_SingleQuick(t *testing.T) {
	lineup, err := pick3v3Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := engine.NewSolo3v3WithHeroes("smoke-3v3-quick", lineup)
	if err != nil {
		t.Fatal(err)
	}
	if !mode.Is3v3(g) {
		t.Fatal("expected 3v3 mode")
	}
	if mode.TeamOf(g, 0) != mode.TeamOf(g, 4) {
		t.Fatal("warm team mismatch")
	}
	if mode.TeamOf(g, 1) != mode.TeamOf(g, 2) {
		t.Fatal("cold team mismatch")
	}
	if !mode.IsCommander3v3(0) || !mode.IsCommander3v3(2) {
		t.Fatal("commander seats wrong")
	}
}
