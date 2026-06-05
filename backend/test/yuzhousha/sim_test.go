package engine_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func simEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CARD_SIM") != "1" {
		t.Skip("AI self-play sim skipped (run: ./scripts/test.sh sim)")
	}
}

func simRandomRounds() int {
	const defaultRounds = 80
	if s := os.Getenv("CARD_SIM_ROUNDS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return defaultRounds
}

// 全武将两两 AI 自对弈（CARD_SIM=1；失败见 test-output/sim/*.log）。
func TestSim_AllHeroPairsAIVsAI(t *testing.T) {
	simEnabled(t)
	if testing.Short() {
		t.Skip("skip all-pairs sim in short mode")
	}
	heroes := engine.HeroesCatalog()
	for _, h0 := range heroes {
		for _, h1 := range heroes {
			name := h0.ID + "_vs_" + h1.ID
			t.Run(name, func(t *testing.T) {
				g, err := engine.NewSolo1v1("sim-"+name, "甲", h0.ID, h1.ID)
				if err != nil {
					t.Fatal(err)
				}
				run := runAISimulation(t, g, defaultSimMaxSteps)
				assertSimFinished(t, g, simContext{
					Label: name, Hero0: h0.ID, Hero1: h1.ID,
				}, run)
			})
		}
	}
}

// 固定种子的随机武将组合；失败日志含 seed 便于复现。
func TestSim_RandomHeroMixSeeded(t *testing.T) {
	simEnabled(t)
	heroes := engine.HeroesCatalog()
	ids := make([]string, len(heroes))
	for i, h := range heroes {
		ids[i] = h.ID
	}
	rounds := simRandomRounds()
	for seed := int64(1); seed <= int64(rounds); seed++ {
		seed := seed
		t.Run(strconv.FormatInt(seed, 10), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			h0 := ids[r.Intn(len(ids))]
			h1 := ids[r.Intn(len(ids))]
			g, err := engine.NewSolo1v1("sim-rand", "甲", h0, h1)
			if err != nil {
				t.Fatal(err)
			}
			run := runAISimulation(t, g, defaultSimMaxSteps)
			assertSimFinished(t, g, simContext{
				Label:  fmt.Sprintf("seed_%d_%s_vs_%s", seed, h0, h1),
				Hero0:  h0,
				Hero1:  h1,
				Seed:   seed,
			}, run)
		})
	}
}

func TestSim_SinglePairQuick(t *testing.T) {
	g, err := engine.NewSolo1v1("sim-quick", "甲", engine.CharMaChao, engine.CharSimaYi)
	if err != nil {
		t.Fatal(err)
	}
	run := runAISimulation(t, g, defaultSimMaxSteps)
	assertSimFinished(t, g, simContext{
		Label: "ma_chao_vs_si_ma_yi",
		Hero0: engine.CharMaChao, Hero1: engine.CharSimaYi,
	}, run)
}
