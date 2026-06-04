package engine_test

import (
	"fmt"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func pick2v2Lineup(fixedSeat0 string, others []string) ([4]string, error) {
	var lineup [4]string
	lineup[0] = fixedSeat0
	used := map[string]bool{fixedSeat0: true}
	idx := 1
	for _, h := range others {
		if used[h] {
			continue
		}
		if idx >= 4 {
			break
		}
		lineup[idx] = h
		used[h] = true
		idx++
	}
	for _, h := range engine.HeroesCatalog() {
		if idx >= 4 {
			break
		}
		if used[h.ID] {
			continue
		}
		lineup[idx] = h.ID
		used[h.ID] = true
		idx++
	}
	if idx < 4 {
		return lineup, fmt.Errorf("not enough distinct heroes for 2v2 lineup")
	}
	return lineup, nil
}

// 每位可选武将作为 0 号位（你）都能正常开局 2v2，hook 不 panic。
func TestSmoke_2v2_AllHeroesBootstrap(t *testing.T) {
	heroes := engine.HeroesCatalog()
	if len(heroes) < 4 {
		t.Fatal("need at least 4 heroes for 2v2 smoke")
	}
	filler := make([]string, 0, len(heroes))
	for _, h := range heroes {
		filler = append(filler, h.ID)
	}
	for _, h := range heroes {
		h := h
		t.Run(h.ID, func(t *testing.T) {
			lineup, err := pick2v2Lineup(h.ID, filler)
			if err != nil {
				t.Fatal(err)
			}
			g, err := engine.NewSolo2v2WithHeroes("smoke-2v2-"+h.ID, lineup)
			if err != nil {
				t.Fatal(err)
			}
			if g.Mode != engine.Mode2v2 {
				t.Fatalf("mode=%q want %q", g.Mode, engine.Mode2v2)
			}
			if len(g.Players) != 4 {
				t.Fatalf("players=%d want 4", len(g.Players))
			}
			for i, p := range g.Players {
				if len(p.Hand) == 0 {
					t.Fatalf("player %d has empty hand after deal", i)
				}
				if len(p.Character.Skills) == 0 {
					t.Fatalf("player %d (%s) has no skills", i, p.Character.ID)
				}
			}
			if mode.TeamOf(g, 0) != mode.TeamOf(g, 2) {
				t.Fatal("seat 0 and 2 should be allies")
			}
			if mode.TeamOf(g, 1) != mode.TeamOf(g, 3) {
				t.Fatal("seat 1 and 3 should be allies")
			}
			if mode.TeamOf(g, 0) == mode.TeamOf(g, 1) {
				t.Fatal("seat 0 and 1 should be enemies")
			}
			assertGameInvariants(t, g)
			var events []engine.GameEvent
			_ = g.RunSkillHooks(&events, skill.HookCall{Kind: skill.HookDistanceDelta, From: 0, To: 1})
			_ = g.RunSkillHooks(&events, skill.HookCall{Kind: skill.HookTargetBlocked, Target: 1, CardKind: engine.CardSha})
			_ = g.RunSkillHooks(&events, skill.HookCall{Kind: skill.HookUnlimitedSha, Seat: 0})
		})
	}
}
