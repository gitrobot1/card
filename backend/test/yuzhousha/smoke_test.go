package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// 快速冒烟：每个可选武将组合都能正常开局，hook 不 panic。
func TestSmoke_AllHeroPairsBootstrap(t *testing.T) {
	heroes := engine.HeroesCatalog()
	if len(heroes) == 0 {
		t.Fatal("empty hero catalog")
	}
	for _, h0 := range heroes {
		for _, h1 := range heroes {
			name := h0.ID + "_vs_" + h1.ID
			t.Run(name, func(t *testing.T) {
				g, err := engine.NewSolo1v1("smoke-"+name, "甲", h0.ID, h1.ID)
				if err != nil {
					t.Fatal(err)
				}
				if g.Phase != engine.PhasePlaying {
					t.Fatalf("expected playing phase, got %s", g.Phase)
				}
				for i, p := range g.Players {
					if len(p.Hand) == 0 {
						t.Fatalf("player %d has empty hand after deal", i)
					}
					if len(p.Character.Skills) == 0 {
						t.Fatalf("player %d (%s) has no skills", i, p.Character.ID)
					}
				}
				assertGameInvariants(t, g)
				var events []engine.GameEvent
				_ = g.RunSkillHooks(&events, skill.HookCall{Kind: skill.HookDistanceDelta, From: 0, To: 1})
				_ = g.RunSkillHooks(&events, skill.HookCall{Kind: skill.HookTargetBlocked, Target: 1, CardKind: engine.CardSha})
				_ = g.RunSkillHooks(&events, skill.HookCall{Kind: skill.HookUnlimitedSha, Seat: 0})
			})
		}
	}
}
