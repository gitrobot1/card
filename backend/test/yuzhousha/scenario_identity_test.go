package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func newIdentityScenario(t *testing.T, id string) *engine.Game {
	t.Helper()
	lineup, err := pickIdentity5Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	roles := defaultIdentity5Roles()
	g, err := engine.NewSoloIdentity5WithHeroes(id, lineup, roles)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func assertIdentityGameOver(t *testing.T, g *engine.Game, wantTeam int) {
	t.Helper()
	if g.Phase != engine.PhaseFinished {
		t.Fatalf("expected finished, phase=%s msg=%q", g.Phase, g.Message)
	}
	if g.WinnerTeam == nil {
		t.Fatal("expected winner_team")
	}
	if *g.WinnerTeam != wantTeam {
		t.Fatalf("winner_team=%d want=%d msg=%q", *g.WinnerTeam, wantTeam, g.Message)
	}
	if g.WinnerIndex == nil {
		t.Fatal("expected winner_index")
	}
}

func assertIdentityContinues(t *testing.T, g *engine.Game) {
	t.Helper()
	if g.Phase == engine.PhaseFinished {
		t.Fatalf("game should continue, got finished: %q", g.Message)
	}
}

// 主公阵亡时反贼阵营胜，即使反贼已全部阵亡（标准规则）。
func TestScenario_Identity_LordDeath_RebelsWinWithoutLivingRebels(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-lord-death")
	g.Players[3].HP = 0
	g.Players[4].HP = 0

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(0, 3, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamRebel)
}

// 反贼与内奸全灭后主公阵营胜。
func TestScenario_Identity_LordFactionWin_AfterSpyEliminated(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-lord-win")
	g.Players[3].HP = 0
	g.Players[4].HP = 0

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(2, 0, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamLordFaction)
}

// 内奸独自存活时内奸胜。
func TestScenario_Identity_SpySoloWin(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-spy-solo")
	g.Players[0].HP = 0
	g.Players[1].HP = 0
	g.Players[3].HP = 0
	// seat2=spy, seat4=rebel 仍存活

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(4, 2, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamSpy)
}

// 反贼阵亡但内奸存活 → 对局继续。
func TestScenario_Identity_ContinuesWhenSpyStillAlive(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-continue-spy")
	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(3, 0, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityContinues(t, g)
}

// 主公与内奸双存活、反贼全灭 → 对局继续（内奸尚未独活）。
func TestScenario_Identity_ContinuesLordAndSpyRemaining(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-lord-spy")
	g.Players[3].HP = 0
	g.Players[4].HP = 0

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(1, 2, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityContinues(t, g)
}

// 阵亡后身份揭示事件。
func TestScenario_Identity_RevealOnDeath(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-reveal")
	if g.RoleRevealed[1] {
		t.Fatal("loyalist should start hidden")
	}

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(1, 3, &events); err != nil {
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

// 主公与内奸单挑，内奸击杀主公 → 内奸胜。
func TestScenario_Identity_SpyDuelKillsLord(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-spy-duel")
	g.Players[1].HP = 0
	g.Players[3].HP = 0
	g.Players[4].HP = 0

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(0, 2, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamSpy)
}

// 主公与内奸单挑，主公无论何种来源死亡 → 内奸胜。
func TestScenario_Identity_SpyDuelLordDiesToUnknown_SpyWins(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-spy-duel-unknown")
	g.Players[1].HP = 0
	g.Players[3].HP = 0
	g.Players[4].HP = 0

	var events []engine.GameEvent
	if err := g.ResolveDyingDeathForTest(0, -1, &events); err != nil {
		t.Fatal(err)
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamSpy)
}

// 绝情致死也走身份胜负判定（非 1v1 直接判 killer 胜）。
func TestScenario_Identity_JueqingLordDeath_RebelsWin(t *testing.T) {
	g := newIdentityScenario(t, "sc-id5-jueqing")
	g.Players[0].HP = 0
	g.Players[3].HP = 0
	g.Players[4].HP = 0

	var events []engine.GameEvent
	if !g.FinishJueqingDeathForTest(2, 0, &events) {
		t.Fatal("expected jueqing death handled")
	}
	assertIdentityGameOver(t, g, mode.IdentityTeamRebel)
}
