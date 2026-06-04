package engine_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func sim3pDdzEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CARD_SIM") != "1" {
		t.Skip("3p ddz AI sim skipped (run: CARD_SIM=1 ./scripts/test.sh sim3p_ddz)")
	}
}

func sim3pDdzContext(label string, lineup [3]string, seed int64) simContext {
	return simContext{
		Label:  label,
		Mode:   "3p_ddz",
		Heroes: lineup[:],
		Hero0:  lineup[0],
		Hero1:  lineup[1],
		Hero2:  lineup[2],
		Seed:   seed,
	}
}

func TestSim_3pDdz_SingleQuick(t *testing.T) {
	lineup := [3]string{
		engine.CharLiuBei,
		engine.CharGuanYu,
		engine.CharZhangFei,
	}
	g, err := engine.NewSolo3pDdzWithHeroes("sim-3p-ddz-quick", lineup)
	if err != nil {
		t.Fatal(err)
	}
	run := runAISimulation(t, g, defaultSimMaxSteps)
	assertSimFinished(t, g, sim3pDdzContext("3p_ddz_liu_line", lineup, 0), run)
}

func TestSim_3pDdz_AllHeroesAsSeat0(t *testing.T) {
	sim3pDdzEnabled(t)
	if testing.Short() {
		t.Skip("skip all-hero 3p ddz sim in short mode")
	}
	ids := heroIDs()
	for _, h := range engine.HeroesCatalog() {
		h := h
		t.Run(h.ID, func(t *testing.T) {
			lineup, err := pick3pLineup(h.ID, ids)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo3pDdzWithHeroes("sim-3p-ddz-"+h.ID, lineup)
			if err != nil {
				t.Fatal(err)
			}
			run := runAISimulation(t, g, defaultSimMaxSteps)
			assertSimFinished(t, g, sim3pDdzContext(h.ID+"_seat0", lineup, 0), run)
		})
	}
}

func TestSim_3pDdz_RandomTriosSeeded(t *testing.T) {
	sim3pDdzEnabled(t)
	ids := heroIDs()
	rounds := simRandomRounds()
	for seed := int64(1); seed <= int64(rounds); seed++ {
		seed := seed
		t.Run(strconv.FormatInt(seed, 10), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			lineup, err := pickRandom3pLineup(r, ids)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo3pDdzWithHeroes("sim-3p-ddz-rand", lineup)
			if err != nil {
				t.Fatal(err)
			}
			label := fmt.Sprintf("seed_%d_%s_%s_%s", seed, lineup[0], lineup[1], lineup[2])
			run := runAISimulation(t, g, defaultSimMaxSteps)
			assertSimFinished(t, g, sim3pDdzContext(label, lineup, seed), run)
		})
	}
}
