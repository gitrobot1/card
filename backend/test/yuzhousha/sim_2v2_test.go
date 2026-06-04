package engine_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func sim2v2Enabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CARD_SIM") != "1" {
		t.Skip("2v2 AI sim skipped (run: CARD_SIM=1 ./scripts/test.sh sim2v2)")
	}
}

func heroIDs() []string {
	heroes := engine.HeroesCatalog()
	ids := make([]string, len(heroes))
	for i, h := range heroes {
		ids[i] = h.ID
	}
	return ids
}

func pickRandom2v2Lineup(r *rand.Rand, ids []string) ([4]string, error) {
	if len(ids) < 4 {
		return [4]string{}, fmt.Errorf("need at least 4 heroes")
	}
	perm := r.Perm(len(ids))
	var lineup [4]string
	for i := 0; i < 4; i++ {
		lineup[i] = ids[perm[i]]
	}
	return lineup, nil
}

func sim2v2Context(label string, lineup [4]string, seed int64) simContext {
	return simContext{
		Label:  label,
		Mode:   "2v2",
		Heroes: lineup[:],
		Hero0:  lineup[0],
		Hero1:  lineup[1],
		Hero2:  lineup[2],
		Hero3:  lineup[3],
		Seed:   seed,
	}
}

func TestSim_2v2_SingleQuick(t *testing.T) {
	lineup := [4]string{
		engine.CharLiuBei,
		engine.CharGuanYu,
		engine.CharZhangFei,
		engine.CharZhaoYun,
	}
	g, err := engine.NewSolo2v2WithHeroes("sim-2v2-quick", lineup)
	if err != nil {
		t.Fatal(err)
	}
	run := runAISimulation(t, g, defaultSimMaxSteps)
	assertSimFinished(t, g, sim2v2Context("2v2_liu_bei_line", lineup, 0), run)
}

// 每位武将作为 0 号位，其余三人随机不重复（CARD_SIM=1）。
func TestSim_2v2_AllHeroesAsSeat0(t *testing.T) {
	sim2v2Enabled(t)
	if testing.Short() {
		t.Skip("skip all-hero 2v2 sim in short mode")
	}
	ids := heroIDs()
	for _, h := range engine.HeroesCatalog() {
		h := h
		t.Run(h.ID, func(t *testing.T) {
			lineup, err := pick2v2Lineup(h.ID, ids)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo2v2WithHeroes("sim-2v2-"+h.ID, lineup)
			if err != nil {
				t.Fatal(err)
			}
			run := runAISimulation(t, g, defaultSimMaxSteps)
			assertSimFinished(t, g, sim2v2Context(h.ID+"_seat0", lineup, 0), run)
		})
	}
}

// 固定种子的随机四人阵容；失败日志含 seed 便于复现（CARD_SIM=1）。
func TestSim_2v2_RandomQuadsSeeded(t *testing.T) {
	sim2v2Enabled(t)
	ids := heroIDs()
	rounds := simRandomRounds()
	for seed := int64(1); seed <= int64(rounds); seed++ {
		seed := seed
		t.Run(strconv.FormatInt(seed, 10), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			lineup, err := pickRandom2v2Lineup(r, ids)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo2v2WithHeroes("sim-2v2-rand", lineup)
			if err != nil {
				t.Fatal(err)
			}
			label := fmt.Sprintf("seed_%d_%s_%s_%s_%s", seed, lineup[0], lineup[1], lineup[2], lineup[3])
			run := runAISimulation(t, g, defaultSimMaxSteps)
			assertSimFinished(t, g, sim2v2Context(label, lineup, seed), run)
		})
	}
}
