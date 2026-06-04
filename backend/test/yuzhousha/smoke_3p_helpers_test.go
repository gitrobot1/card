package engine_test

import (
	"fmt"
	"math/rand"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func pick3pLineup(fixedSeat0 string, others []string) ([3]string, error) {
	var lineup [3]string
	lineup[0] = fixedSeat0
	used := map[string]bool{fixedSeat0: true}
	idx := 1
	for _, h := range others {
		if used[h] {
			continue
		}
		if idx >= 3 {
			break
		}
		lineup[idx] = h
		used[h] = true
		idx++
	}
	for _, h := range engine.HeroesCatalog() {
		if idx >= 3 {
			break
		}
		if used[h.ID] {
			continue
		}
		lineup[idx] = h.ID
		used[h.ID] = true
		idx++
	}
	if idx < 3 {
		return lineup, fmt.Errorf("not enough distinct heroes for 3p lineup")
	}
	return lineup, nil
}

func pickRandom3pLineup(r *rand.Rand, ids []string) ([3]string, error) {
	if len(ids) < 3 {
		return [3]string{}, fmt.Errorf("need at least 3 heroes")
	}
	perm := r.Perm(len(ids))
	var lineup [3]string
	for i := 0; i < 3; i++ {
		lineup[i] = ids[perm[i]]
	}
	return lineup, nil
}
