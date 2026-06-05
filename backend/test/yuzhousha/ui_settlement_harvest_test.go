package engine_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func uiFixtureEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CARD_UI_FIXTURE") != "1" {
		t.Skip("settlement fixture harvest skipped (run: CARD_UI_FIXTURE=1 CARD_SIM=1 go test -tags cardtest ./test/yuzhousha/... -run TestHarvestYzsSettlementFixtures)")
	}
	if os.Getenv("CARD_SIM") != "1" {
		t.Skip("harvest requires CARD_SIM=1")
	}
}

func uiFixtureRounds() int {
	if v := os.Getenv("CARD_UI_FIXTURE_ROUNDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 15
}

func uiFixtureDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("..", "..", "frontend", "test", "fixtures", "yzs", "settlements")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "frontend", "test", "fixtures", "yzs", "settlements")
}

type settlementFixtureFile struct {
	Meta settlementFixtureMeta    `json:"meta"`
	State engine.PublicState      `json:"state"`
}

type settlementFixtureMeta struct {
	Mode       string `json:"mode"`
	Seed       int64  `json:"seed,omitempty"`
	Label      string `json:"label,omitempty"`
	WinnerTeam int    `json:"winner_team,omitempty"`
	ExportedAt string `json:"exported_at"`
}

func exportSettlementFixture(t *testing.T, g *engine.Game, ctx simContext) {
	t.Helper()
	if os.Getenv("CARD_UI_FIXTURE") != "1" || !g.IsFinished() || g.WinnerIndex == nil {
		return
	}
	dir := uiFixtureDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Logf("mkdir fixtures: %v", err)
		return
	}
	events := []engine.GameEvent{{
		Type:    "game_over",
		Message: g.Message,
	}}
	if g.WinnerIndex != nil {
		events[0].PlayerIndex = *g.WinnerIndex
	}
	pub := g.PublicViewForSeat(0, events)
	meta := settlementFixtureMeta{
		Mode:       g.Mode,
		Seed:       ctx.Seed,
		Label:      ctx.Label,
		ExportedAt: time.Now().Format(time.RFC3339),
	}
	if g.WinnerTeam != nil {
		meta.WinnerTeam = *g.WinnerTeam
	}
	payload := settlementFixtureFile{Meta: meta, State: pub}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Logf("marshal fixture: %v", err)
		return
	}
	name := sanitizeLogName(fmt.Sprintf("%s_%s_seed%d", g.Mode, ctx.Label, ctx.Seed))
	if name == "" || name == "_seed0" {
		name = sanitizeLogName(fmt.Sprintf("%s_%d", g.Mode, time.Now().UnixNano()))
	}
	path := filepath.Join(dir, name+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Logf("write fixture %s: %v", path, err)
		return
	}
	t.Logf("UI fixture → %s", path)
}

// 七模式随机自走，终局导出 PublicView 供前端 settlement 测试（CARD_SIM=1 CARD_UI_FIXTURE=1）。
func TestHarvestYzsSettlementFixtures(t *testing.T) {
	uiFixtureEnabled(t)
	rounds := uiFixtureRounds()
	ids := heroIDs()
	exported := 0

	t.Run("identity_8", func(t *testing.T) {
		for seed := int64(1); seed <= int64(rounds); seed++ {
			seed := seed
			r := rand.New(rand.NewSource(seed))
			lineup, err := pickRandomIdentity8Lineup(r, ids)
			if err != nil {
				t.Fatal(err)
			}
			roles := pickRandomIdentity8Roles(r)
			g, err := engine.NewSoloIdentity8WithHeroes("harvest-id8", lineup, roles)
			if err != nil {
				t.Fatal(err)
			}
			g.SetDeckSeedForTest(seed)
			label := fmt.Sprintf("id8_s%d", seed)
			run := runAISimulation(t, g, identity8SimMaxSteps)
			if run.result.finished && g.WinnerIndex != nil {
				exportSettlementFixture(t, g, simIdentity8Context(label, lineup, seed))
				exported++
			}
		}
	})

	t.Run("identity_5", func(t *testing.T) {
		for seed := int64(1); seed <= int64(rounds); seed++ {
			seed := seed
			r := rand.New(rand.NewSource(seed))
			lineup, err := pickRandomIdentity5Lineup(r, ids)
			if err != nil {
				t.Fatal(err)
			}
			roles := pickRandomIdentity5Roles(r)
			g, err := engine.NewSoloIdentity5WithHeroes("harvest-id5", lineup, roles)
			if err != nil {
				t.Fatal(err)
			}
			g.SetDeckSeedForTest(seed)
			label := fmt.Sprintf("id5_s%d", seed)
			run := runAISimulation(t, g, identitySimMaxSteps)
			if run.result.finished && g.WinnerIndex != nil {
				exportSettlementFixture(t, g, simIdentity5Context(label, lineup, seed))
				exported++
			}
		}
	})

	t.Run("2v2", func(t *testing.T) {
		for seed := int64(1); seed <= int64(min(rounds, 10)); seed++ {
			seed := seed
			r := rand.New(rand.NewSource(seed))
			lineup, err := pickRandom2v2Lineup(r, ids)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo2v2WithHeroes("harvest-2v2", lineup)
			if err != nil {
				t.Fatal(err)
			}
			g.SetDeckSeedForTest(seed)
			run := runAISimulation(t, g, defaultSimMaxSteps)
			if run.result.finished && g.WinnerIndex != nil {
				exportSettlementFixture(t, g, simContext{Label: fmt.Sprintf("2v2_s%d", seed), Mode: "2v2", Heroes: lineup[:], Seed: seed})
				exported++
			}
		}
	})

	t.Run("1v1", func(t *testing.T) {
		for seed := int64(1); seed <= int64(min(rounds, 10)); seed++ {
			seed := seed
			r := rand.New(rand.NewSource(seed + 100))
			h0 := ids[r.Intn(len(ids))]
			h1 := ids[r.Intn(len(ids))]
			if h0 == h1 {
				h1 = ids[(r.Intn(len(ids))+1)%len(ids)]
			}
			g, err := engine.NewSolo1v1("harvest-1v1", "测试", h0, h1)
			if err != nil {
				t.Fatal(err)
			}
			g.SetDeckSeedForTest(seed)
			run := runAISimulation(t, g, defaultSimMaxSteps)
			if run.result.finished && g.WinnerIndex != nil {
				exportSettlementFixture(t, g, simContext{Label: fmt.Sprintf("1v1_s%d", seed), Mode: "1v1", Hero0: h0, Hero1: h1, Seed: seed})
				exported++
			}
		}
	})

	t.Logf("exported %d settlement fixtures → %s", exported, uiFixtureDir())
	if exported == 0 {
		t.Fatal("no fixtures exported — sim may be stuck; check sim_logs")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
