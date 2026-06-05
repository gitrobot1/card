package engine_test

import (
	"fmt"
	"math/rand"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
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

func pick3v3Lineup(fixedSeat0 string, others []string) ([6]string, error) {
	var lineup [6]string
	lineup[0] = fixedSeat0
	used := map[string]bool{fixedSeat0: true}
	idx := 1
	for _, h := range others {
		if used[h] {
			continue
		}
		if idx >= 6 {
			break
		}
		lineup[idx] = h
		used[h] = true
		idx++
	}
	for _, h := range engine.HeroesCatalog() {
		if idx >= 6 {
			break
		}
		if used[h.ID] {
			continue
		}
		lineup[idx] = h.ID
		used[h.ID] = true
		idx++
	}
	if idx < 6 {
		return lineup, fmt.Errorf("not enough distinct heroes for 3v3 lineup")
	}
	return lineup, nil
}

func pickRandom3v3Lineup(r *rand.Rand, ids []string) ([6]string, error) {
	if len(ids) < 6 {
		return [6]string{}, fmt.Errorf("need at least 6 heroes")
	}
	perm := r.Perm(len(ids))
	var lineup [6]string
	for i := 0; i < 6; i++ {
		lineup[i] = ids[perm[i]]
	}
	return lineup, nil
}

func pickRandomIdentity5Lineup(r *rand.Rand, ids []string) ([5]string, error) {
	if len(ids) < 5 {
		return [5]string{}, fmt.Errorf("need at least 5 heroes")
	}
	perm := r.Perm(len(ids))
	var lineup [5]string
	for i := 0; i < 5; i++ {
		lineup[i] = ids[perm[i]]
	}
	return lineup, nil
}

func pickRandomIdentity5Roles(r *rand.Rand) [5]string {
	pool := []string{
		mode.RoleLord,
		mode.RoleLoyalist,
		mode.RoleSpy,
		mode.RoleRebel,
		mode.RoleRebel,
	}
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	var roles [5]string
	copy(roles[:], pool)
	return roles
}

func defaultIdentity8Roles() [8]string {
	return [8]string{
		mode.RoleLord,
		mode.RoleLoyalist, mode.RoleLoyalist,
		mode.RoleSpy,
		mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel,
	}
}

func pickRandomIdentity8Lineup(r *rand.Rand, ids []string) ([8]string, error) {
	if len(ids) < 8 {
		return [8]string{}, fmt.Errorf("need at least 8 heroes")
	}
	perm := r.Perm(len(ids))
	var lineup [8]string
	for i := 0; i < 8; i++ {
		lineup[i] = ids[perm[i]]
	}
	return lineup, nil
}

func pickRandomIdentity8Roles(r *rand.Rand) [8]string {
	pool := []string{
		mode.RoleLord,
		mode.RoleLoyalist, mode.RoleLoyalist,
		mode.RoleSpy,
		mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel,
	}
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	var roles [8]string
	copy(roles[:], pool)
	return roles
}

func pickIdentity8Lineup(fixedSeat0 string, others []string) ([8]string, error) {
	var lineup [8]string
	lineup[0] = fixedSeat0
	used := map[string]bool{fixedSeat0: true}
	idx := 1
	for _, h := range others {
		if used[h] {
			continue
		}
		if idx >= 8 {
			break
		}
		lineup[idx] = h
		used[h] = true
		idx++
	}
	for _, h := range engine.HeroesCatalog() {
		if idx >= 8 {
			break
		}
		if used[h.ID] {
			continue
		}
		lineup[idx] = h.ID
		used[h.ID] = true
		idx++
	}
	if idx < 8 {
		return lineup, fmt.Errorf("not enough distinct heroes for identity_8 lineup")
	}
	return lineup, nil
}

func pickIdentity5Lineup(fixedSeat0 string, others []string) ([5]string, error) {
	var lineup [5]string
	lineup[0] = fixedSeat0
	used := map[string]bool{fixedSeat0: true}
	idx := 1
	for _, h := range others {
		if used[h] {
			continue
		}
		if idx >= 5 {
			break
		}
		lineup[idx] = h
		used[h] = true
		idx++
	}
	for _, h := range engine.HeroesCatalog() {
		if idx >= 5 {
			break
		}
		if used[h.ID] {
			continue
		}
		lineup[idx] = h.ID
		used[h.ID] = true
		idx++
	}
	if idx < 5 {
		return lineup, fmt.Errorf("not enough distinct heroes for identity_5 lineup")
	}
	return lineup, nil
}

func defaultIdentity5Roles() [5]string {
	return [5]string{
		mode.RoleLord,
		mode.RoleLoyalist,
		mode.RoleSpy,
		mode.RoleRebel,
		mode.RoleRebel,
	}
}
