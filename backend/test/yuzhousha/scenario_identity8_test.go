package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

// 默认八人身份：0 主 1/2 忠 3 内 4–7 反。
func newIdentity8Scenario(t *testing.T, id string) *engine.Game {
	t.Helper()
	lineup, err := pickIdentity8Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	roles := defaultIdentity8Roles()
	g, err := engine.NewSoloIdentity8WithHeroes(id, lineup, roles)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestScenario_Identity8_LordDeath_RebelsWinWithoutLivingRebels(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-lord-death")
	for _, seat := range []int{4, 5, 6, 7} {
		g.Players[seat].HP = 0
	}

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(0, 4, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamRebel)
}

func TestScenario_Identity8_LordFactionWin_AfterSpyEliminated(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-lord-win")
	for _, seat := range []int{4, 5, 6, 7} {
		g.Players[seat].HP = 0
	}

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(3, 0, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamLordFaction)
}

func TestScenario_Identity8_SpySoloWin(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-spy-solo")
	for _, seat := range []int{0, 1, 2, 4, 5, 6} {
		g.Players[seat].HP = 0
	}

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(7, 3, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamSpy)
}

func TestScenario_Identity8_ContinuesWhenSpyStillAlive(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-continue-spy")
	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(4, 0, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityContinues(t, g)
}

func TestScenario_Identity8_ContinuesLordAndSpyRemaining(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-lord-spy")
	for _, seat := range []int{4, 5, 6, 7} {
		g.Players[seat].HP = 0
	}

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(1, 3, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityContinues(t, g)
}

func TestScenario_Identity8_RevealOnDeath(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-reveal")
	if g.RoleRevealed[1] {
		t.Fatal("loyalist should start hidden")
	}

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(1, 4, &events); err != nil {
		t.Fatal(err)
	}
	if !g.RoleRevealed[1] {
		t.Fatal("loyalist should be revealed after death")
	}
	found := false
	for _, e := range events {
		if e.Type == "identity_revealed" && e.PlayerIndex == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected identity_revealed event")
	}
}

func TestScenario_Identity8_SpyDuelKillsLord(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-spy-duel")
	for _, seat := range []int{1, 2, 4, 5, 6, 7} {
		g.Players[seat].HP = 0
	}

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(0, 3, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamSpy)
}

func TestScenario_Identity8_SpyDuelLordDiesToUnknown_SpyWins(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-spy-duel-unknown")
	for _, seat := range []int{1, 2, 4, 5, 6, 7} {
		g.Players[seat].HP = 0
	}

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(0, -1, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamSpy)
}

func TestScenario_Identity8_JueqingLordDeath_RebelsWin(t *testing.T) {
	g := newIdentity8Scenario(t, "sc-id8-jueqing")
	g.Players[0].HP = 0
	for _, seat := range []int{4, 5, 6, 7} {
		g.Players[seat].HP = 0
	}

	var events []engine.GameEvent
	if !g.FinishJueqingDeathForTest(3, 0, &events) {
		t.Fatal("expected jueqing death handled")
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamRebel)
}
