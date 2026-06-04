package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func TestSmoke_3pChain_Bootstrap(t *testing.T) {
	g, err := engine.NewSolo3pChain("smoke-3p", "甲", engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	if g.Mode != engine.Mode3pChain {
		t.Fatalf("mode=%q", g.Mode)
	}
	if len(g.Players) != 3 {
		t.Fatalf("players=%d", len(g.Players))
	}
	if mode.MarkTarget(g, 0) != 2 || mode.ProtectTarget(g, 0) != 1 {
		t.Fatal("unexpected chain seats for human")
	}
	assertGameInvariants(t, g)
	var events []engine.GameEvent
	_ = g.RunSkillHooks(&events, skill.HookCall{Kind: skill.HookTargetBlocked, Target: 2, CardKind: engine.CardSha})
}

func TestSmoke_3pChain_AllHeroesBootstrap(t *testing.T) {
	heroes := engine.HeroesCatalog()
	if len(heroes) < 3 {
		t.Fatal("need at least 3 heroes for 3p chain smoke")
	}
	filler := make([]string, 0, len(heroes))
	for _, h := range heroes {
		filler = append(filler, h.ID)
	}
	for _, h := range heroes {
		h := h
		t.Run(h.ID, func(t *testing.T) {
			lineup, err := pick3pLineup(h.ID, filler)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo3pChainWithHeroes("smoke-3p-chain-"+h.ID, lineup)
			if err != nil {
				t.Fatal(err)
			}
			if g.Mode != engine.Mode3pChain {
				t.Fatalf("mode=%q want %q", g.Mode, engine.Mode3pChain)
			}
			if len(g.Players) != 3 {
				t.Fatalf("players=%d want 3", len(g.Players))
			}
			if mode.MarkTarget(g, 0) != 2 || mode.ProtectTarget(g, 0) != 1 {
				t.Fatal("unexpected chain seats for seat 0")
			}
			if mode.IsEnemy(g, 0, 2) != true || mode.IsAlly(g, 0, 1) != true {
				t.Fatal("chain ally/enemy mismatch")
			}
			assertGameInvariants(t, g)
			var events []engine.GameEvent
			_ = g.RunSkillHooks(&events, skill.HookCall{Kind: skill.HookTargetBlocked, Target: 2, CardKind: engine.CardSha})
		})
	}
}

func TestSmoke_3pDdz_AllHeroesBootstrap(t *testing.T) {
	heroes := engine.HeroesCatalog()
	if len(heroes) < 3 {
		t.Fatal("need at least 3 heroes for 3p ddz smoke")
	}
	filler := make([]string, 0, len(heroes))
	for _, h := range heroes {
		filler = append(filler, h.ID)
	}
	for _, h := range heroes {
		h := h
		t.Run(h.ID, func(t *testing.T) {
			lineup, err := pick3pLineup(h.ID, filler)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo3pDdzWithHeroes("smoke-3p-ddz-"+h.ID, lineup)
			if err != nil {
				t.Fatal(err)
			}
			if g.Mode != engine.Mode3pDdz {
				t.Fatalf("mode=%q want %q", g.Mode, engine.Mode3pDdz)
			}
			if g.LandlordSeat != 0 {
				t.Fatalf("landlord=%d want 0", g.LandlordSeat)
			}
			if mode.TeamOf(g, 0) != 0 {
				t.Fatal("seat 0 should be landlord team")
			}
			if mode.TeamOf(g, 1) != 1 || mode.TeamOf(g, 2) != 1 {
				t.Fatal("seats 1/2 should be farmer team")
			}
			if !mode.IsEnemy(g, 0, 1) || !mode.IsAlly(g, 1, 2) {
				t.Fatal("ddz team relations mismatch")
			}
			assertGameInvariants(t, g)
		})
	}
}
