package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func TestSmoke_3pDdz_Bootstrap(t *testing.T) {
	g, err := engine.NewSolo3pDdz("smoke-ddz", "甲", engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	if g.Mode != engine.Mode3pDdz {
		t.Fatalf("mode=%q", g.Mode)
	}
	if g.LandlordSeat != 0 {
		t.Fatalf("landlord=%d", g.LandlordSeat)
	}
	if len(g.Players) != 3 {
		t.Fatalf("players=%d", len(g.Players))
	}
	if mode.TeamOf(g, 0) != 0 {
		t.Fatal("human should be landlord team")
	}
	if mode.TeamOf(g, 1) != 1 || mode.TeamOf(g, 2) != 1 {
		t.Fatal("AI should be farmer team")
	}
	if !mode.IsEnemy(g, 0, 1) {
		t.Fatal("landlord and farmer should be enemies")
	}
	assertGameInvariants(t, g)
}

func TestSmoke_3pDdz_LandlordPerks(t *testing.T) {
	g, err := engine.NewSolo3pDdz("smoke-ddz-perks", "甲", engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	if g.DrawCountForTest(0) != engine.DrawPerTurn+1 {
		t.Fatalf("landlord draw=%d want %d", g.DrawCountForTest(0), engine.DrawPerTurn+1)
	}
	if g.DrawCountForTest(1) != engine.DrawPerTurn {
		t.Fatalf("farmer draw=%d", g.DrawCountForTest(1))
	}
	if !g.CanUseShaForTest(0) {
		t.Fatal("landlord should be able to use first sha")
	}
	g.Players[0].ShaUsedThisTurn = true
	if !g.CanUseShaForTest(0) {
		t.Fatal("landlord should be able to use second sha")
	}
	g.Players[0].ShaExtraUsedThisTurn = true
	if g.CanUseShaForTest(0) {
		t.Fatal("landlord should be out of sha")
	}
}
