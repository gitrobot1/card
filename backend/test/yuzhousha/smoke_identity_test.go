package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func TestSmoke_Identity5_AllHeroesBootstrap(t *testing.T) {
	heroes := engine.HeroesCatalog()
	if len(heroes) < 5 {
		t.Fatal("need at least 5 heroes for identity smoke")
	}
	others := make([]string, 0, 4)
	for _, h := range heroes {
		if h.ID != heroes[0].ID {
			others = append(others, h.ID)
		}
		if len(others) >= 4 {
			break
		}
	}
	roles := defaultIdentity5Roles()
	for _, h := range heroes {
		lineup, err := pickIdentity5Lineup(h.ID, others)
		if err != nil {
			t.Fatalf("pick lineup for %s: %v", h.ID, err)
		}
		g, err := engine.NewSoloIdentity5WithHeroes("smoke-id5-"+h.ID, lineup, roles)
		if err != nil {
			t.Fatalf("NewSoloIdentity5WithHeroes(%s): %v", h.ID, err)
		}
		if g.Mode != mode.SoloIdentity5 || len(g.Players) != 5 {
			t.Fatalf("unexpected game mode/players for %s", h.ID)
		}
		if g.LordSeat != 0 || g.Identities[0] != mode.RoleLord {
			t.Fatalf("seat0 should be lord for %s", h.ID)
		}
		assertGameInvariants(t, g)
	}
}

func TestSmoke_Identity5_SingleQuick(t *testing.T) {
	lineup, err := pickIdentity5Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	roles := defaultIdentity5Roles()
	g, err := engine.NewSoloIdentity5WithHeroes("smoke-id5-quick", lineup, roles)
	if err != nil {
		t.Fatal(err)
	}
	if !mode.IsIdentity(g) {
		t.Fatal("expected identity_5 mode")
	}
	if g.Players[0].MaxHP != g.Players[0].Character.MaxHP+1 {
		t.Fatalf("lord max hp = %d, want base+1=%d", g.Players[0].MaxHP, g.Players[0].Character.MaxHP+1)
	}
	if g.CurrentTurn != g.LordSeat {
		t.Fatalf("first turn should be lord seat %d, got %d", g.LordSeat, g.CurrentTurn)
	}
	counts := map[string]int{}
	for _, role := range g.Identities {
		counts[role]++
	}
	if counts[mode.RoleLord] != 1 || counts[mode.RoleLoyalist] != 1 || counts[mode.RoleSpy] != 1 || counts[mode.RoleRebel] != 2 {
		t.Fatalf("bad role distribution: %v", counts)
	}
	targets := g.ValidPlayTargetsForTest(0, engine.CardSha)
	if len(targets) == 0 {
		t.Fatal("lord should have at least one sha target in range")
	}
	foundAlly := false
	for _, seat := range targets {
		if g.Identities[seat] == mode.RoleLoyalist {
			foundAlly = true
		}
	}
	if !foundAlly {
		t.Fatalf("identity mode should allow sha on ally in range; targets=%v roles=%v", targets, g.Identities)
	}
}

func TestSmoke_Identity5_TunedDeckKeepsShanDian(t *testing.T) {
	lineup, err := pickIdentity5Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	roles := defaultIdentity5Roles()
	g, err := engine.NewSoloIdentity5WithHeroes("smoke-id5-deck", lineup, roles)
	if err != nil {
		t.Fatal(err)
	}
	want := mode.DeckProfileFor(mode.SoloIdentity5)
	if got := countCardsInPlay(g); got != want.TotalCards() {
		t.Fatalf("identity_5 card total=%d want %d", got, want.TotalCards())
	}
	hasShandian := false
	sha := 0
	for _, c := range allCardsInGame(g) {
		if c.Kind == engine.CardShanDian {
			hasShandian = true
		}
		if c.Kind == engine.CardSha {
			sha++
		}
	}
	if !hasShandian {
		t.Fatal("identity_5 deck should still include shandian")
	}
	if sha != want.CountKind(mode.DeckKindSha) {
		t.Fatalf("identity_5 sha in play=%d want %d", sha, want.CountKind(mode.DeckKindSha))
	}
}

func TestSmoke_Identity5_SetupShuffle(t *testing.T) {
	g, err := engine.NewSoloIdentity5("smoke-id5-shuffle", "测试", engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	if g.Identities[0] != mode.RoleLord {
		t.Fatal("human seat must be lord")
	}
	counts := map[string]int{}
	for _, role := range g.Identities {
		counts[role]++
	}
	if counts[mode.RoleLoyalist] != 1 || counts[mode.RoleSpy] != 1 || counts[mode.RoleRebel] != 2 {
		t.Fatalf("AI seats should be 1 loyalist + 1 spy + 2 rebels, got %v", counts)
	}
}
