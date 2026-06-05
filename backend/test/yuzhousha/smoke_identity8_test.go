package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func TestSmoke_Identity8_AllHeroesBootstrap(t *testing.T) {
	heroes := engine.HeroesCatalog()
	if len(heroes) < 8 {
		t.Fatal("need at least 8 heroes for identity_8 smoke")
	}
	others := make([]string, 0, 7)
	for _, h := range heroes {
		if h.ID != heroes[0].ID {
			others = append(others, h.ID)
		}
		if len(others) >= 7 {
			break
		}
	}
	roles := defaultIdentity8Roles()
	for _, h := range heroes {
		lineup, err := pickIdentity8Lineup(h.ID, others)
		if err != nil {
			t.Fatalf("pick lineup for %s: %v", h.ID, err)
		}
		g, err := engine.NewSoloIdentity8WithHeroes("smoke-id8-"+h.ID, lineup, roles)
		if err != nil {
			t.Fatalf("NewSoloIdentity8WithHeroes(%s): %v", h.ID, err)
		}
		if g.Mode != mode.SoloIdentity8 || len(g.Players) != 8 {
			t.Fatalf("unexpected game mode/players for %s", h.ID)
		}
		if g.LordSeat != 0 || g.Identities[0] != mode.RoleLord {
			t.Fatalf("seat0 should be lord for %s", h.ID)
		}
		assertGameInvariants(t, g)
	}
}

func TestSmoke_Identity8_SingleQuick(t *testing.T) {
	lineup, err := pickIdentity8Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	roles := defaultIdentity8Roles()
	g, err := engine.NewSoloIdentity8WithHeroes("smoke-id8-quick", lineup, roles)
	if err != nil {
		t.Fatal(err)
	}
	if g.Mode != mode.SoloIdentity8 {
		t.Fatal("expected identity_8 mode")
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
	if counts[mode.RoleLord] != 1 || counts[mode.RoleLoyalist] != 2 || counts[mode.RoleSpy] != 1 || counts[mode.RoleRebel] != 4 {
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

func TestSmoke_Identity8_LargeDeckNoShanDian(t *testing.T) {
	lineup, err := pickIdentity8Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	roles := defaultIdentity8Roles()
	g, err := engine.NewSoloIdentity8WithHeroes("smoke-id8-deck", lineup, roles)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range allCardsInGame(g) {
		if c.Kind == engine.CardShanDian {
			t.Fatalf("identity_8 deck must not contain shandian, found %s", c.ID)
		}
	}
	want := mode.DeckProfileFor(mode.SoloIdentity8).TotalCards()
	if got := countCardsInPlay(g); got != want {
		t.Fatalf("identity_8 card total=%d want %d", got, want)
	}
	// 8 人 × 4 起手 + 主公首轮摸 2 张 → 牌堆 83−32−2=49（明显大于 legacy 八人局 ~23）
	if len(g.DrawPile) < 40 {
		t.Fatalf("identity_8 draw pile too thin: %d", len(g.DrawPile))
	}
}

func TestSmoke_Identity8_SetupShuffle(t *testing.T) {
	g, err := engine.NewSoloIdentity8("smoke-id8-shuffle", "测试", engine.CharLiuBei)
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
	if counts[mode.RoleLoyalist] != 2 || counts[mode.RoleSpy] != 1 || counts[mode.RoleRebel] != 4 {
		t.Fatalf("AI seats should be 2 loyalist + 1 spy + 4 rebels, got %v", counts)
	}
}
