package engine_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func sim3v3Enabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CARD_SIM") != "1" {
		t.Skip("3v3 AI sim skipped (run: CARD_SIM=1 ./scripts/test.sh sim3v3)")
	}
}

func sim3v3Context(label string, lineup [6]string, seed int64) simContext {
	return simContext{
		Label:  label,
		Mode:   "3v3",
		Heroes: lineup[:],
		Hero0:  lineup[0],
		Hero1:  lineup[1],
		Hero2:  lineup[2],
		Hero3:  lineup[3],
		Seed:   seed,
	}
}

func TestSim_3v3_SingleQuick(t *testing.T) {
	lineup, err := pick3v3Lineup(engine.CharLiuBei, nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := engine.NewSolo3v3WithHeroes("sim-3v3-quick", lineup)
	if err != nil {
		t.Fatal(err)
	}
	run := runAISimulation(t, g, defaultSimMaxSteps)
	assertSimFinished(t, g, sim3v3Context("3v3_liu_line", lineup, 0), run)
}

// 每位武将作为 0 号位（暖主帅），其余五人随机不重复（CARD_SIM=1）。
func TestSim_3v3_AllHeroesAsSeat0(t *testing.T) {
	sim3v3Enabled(t)
	if testing.Short() {
		t.Skip("skip all-hero 3v3 sim in short mode")
	}
	ids := heroIDs()
	for _, h := range engine.HeroesCatalog() {
		h := h
		t.Run(h.ID, func(t *testing.T) {
			lineup, err := pick3v3Lineup(h.ID, ids)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo3v3WithHeroes("sim-3v3-"+h.ID, lineup)
			if err != nil {
				t.Fatal(err)
			}
			run := runAISimulation(t, g, defaultSimMaxSteps)
			assertSimFinished(t, g, sim3v3Context(h.ID+"_seat0", lineup, 0), run)
		})
	}
}

// 固定种子的随机六人阵容；失败日志含 seed 便于复现（CARD_SIM=1）。
func TestSim_3v3_RandomHexesSeeded(t *testing.T) {
	sim3v3Enabled(t)
	ids := heroIDs()
	rounds := simRandomRounds()
	for seed := int64(1); seed <= int64(rounds); seed++ {
		seed := seed
		t.Run(strconv.FormatInt(seed, 10), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			lineup, err := pickRandom3v3Lineup(r, ids)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo3v3WithHeroes("sim-3v3-rand", lineup)
			if err != nil {
				t.Fatal(err)
			}
			label := fmt.Sprintf("seed_%d_%s_%s_%s_%s_%s_%s", seed, lineup[0], lineup[1], lineup[2], lineup[3], lineup[4], lineup[5])
			run := runAISimulation(t, g, defaultSimMaxSteps)
			assertSimFinished(t, g, sim3v3Context(label, lineup, seed), run)
		})
	}
}
