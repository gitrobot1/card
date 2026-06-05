package engine_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func simIdentity5Enabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CARD_SIM") != "1" {
		t.Skip("identity_5 AI sim skipped (run: CARD_SIM=1 ./scripts/test.sh simidentity)")
	}
}

func simIdentity5Context(label string, lineup [5]string, seed int64) simContext {
	return simContext{
		Label:  label,
		Mode:   "identity_5",
		Heroes: lineup[:],
		Hero0:  lineup[0],
		Hero1:  lineup[1],
		Hero2:  lineup[2],
		Hero3:  lineup[3],
		Seed:   seed,
	}
}

func TestSim_Identity5_SingleQuick(t *testing.T) {
	lineup, err := pickIdentity5Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	roles := defaultIdentity5Roles()
	g, err := engine.NewSoloIdentity5WithHeroes("sim-id5-quick", lineup, roles)
	if err != nil {
		t.Fatal(err)
	}
	run := runAISimulation(t, g, identitySimMaxSteps)
	assertSimFinished(t, g, simIdentity5Context("id5_liu_lord", lineup, 0), run)
}

// 每位武将作为 0 号位主公，其余四人随机不重复（CARD_SIM=1）。
func TestSim_Identity5_AllHeroesAsSeat0(t *testing.T) {
	simIdentity5Enabled(t)
	if testing.Short() {
		t.Skip("skip all-hero identity_5 sim in short mode")
	}
	ids := heroIDs()
	for _, h := range engine.HeroesCatalog() {
		h := h
		t.Run(h.ID, func(t *testing.T) {
			lineup, err := pickIdentity5Lineup(h.ID, ids)
			if err != nil {
				t.Fatal(err)
			}
			roles := defaultIdentity5Roles()
			g, err := engine.NewSoloIdentity5WithHeroes("sim-id5-"+h.ID, lineup, roles)
			if err != nil {
				t.Fatal(err)
			}
			run := runAISimulation(t, g, identitySimMaxSteps)
			assertSimFinished(t, g, simIdentity5Context(h.ID+"_lord", lineup, 0), run)
		})
	}
}

// 固定种子的随机五人阵容与身份分配；失败日志含 seed 便于复现（CARD_SIM=1）。
func TestSim_Identity5_RandomPentasSeeded(t *testing.T) {
	simIdentity5Enabled(t)
	ids := heroIDs()
	rounds := simRandomRounds()
	for seed := int64(1); seed <= int64(rounds); seed++ {
		seed := seed
		t.Run(strconv.FormatInt(seed, 10), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			lineup, err := pickRandomIdentity5Lineup(r, ids)
			if err != nil {
				t.Fatal(err)
			}
			roles := pickRandomIdentity5Roles(r)
			g, err := engine.NewSoloIdentity5WithHeroes("sim-id5-rand", lineup, roles)
			if err != nil {
				t.Fatal(err)
			}
			g.SetDeckSeedForTest(seed)
			label := fmt.Sprintf("seed_%d_%s_%s_%s_%s_%s", seed, lineup[0], lineup[1], lineup[2], lineup[3], lineup[4])
			run := runAISimulation(t, g, identitySimMaxSteps)
			assertSimFinished(t, g, simIdentity5Context(label, lineup, seed), run)
		})
	}
}
