package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func TestIdentity5_LordSkillsNotInactiveIn1v1(t *testing.T) {
	g, err := engine.NewSoloIdentity5("lord-skill-meta", "玩家", engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	lord := &g.Players[g.LordSeat]
	for _, s := range lord.Character.Skills {
		if s.Kind != skill.KindLord {
			continue
		}
		if s.InactiveIn1v1 {
			t.Fatalf("lord skill %q should be active in identity_5, got inactive_in_1v1", s.ID)
		}
	}
}

func TestListHeroes_Identity5_LordSkillsActive(t *testing.T) {
	page := engine.ListHeroes(engine.HeroesQuery{Mode: "identity_5", PageSize: 100})
	for _, h := range page.Heroes {
		if h.ID != engine.CharLiuBei {
			continue
		}
		for _, s := range h.Skills {
			if s.Kind == skill.KindLord && s.InactiveIn1v1 {
				t.Fatalf("catalog liu_bei lord skill %q marked inactive in identity_5", s.ID)
			}
		}
		return
	}
	t.Fatal("liu_bei not found in identity_5 hero list")
}

func TestIdentity8_LordSkillsNotInactiveIn1v1(t *testing.T) {
	g, err := engine.NewSoloIdentity8("lord-skill-meta-8", "玩家", engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	lord := &g.Players[g.LordSeat]
	for _, s := range lord.Character.Skills {
		if s.Kind != skill.KindLord {
			continue
		}
		if s.InactiveIn1v1 {
			t.Fatalf("lord skill %q should be active in identity_8, got inactive_in_1v1", s.ID)
		}
	}
}

func TestListHeroes_Identity8_LordSkillsActive(t *testing.T) {
	page := engine.ListHeroes(engine.HeroesQuery{Mode: "identity_8", PageSize: 100})
	for _, h := range page.Heroes {
		if h.ID != engine.CharLiuBei {
			continue
		}
		for _, s := range h.Skills {
			if s.Kind == skill.KindLord && s.InactiveIn1v1 {
				t.Fatalf("catalog liu_bei lord skill %q marked inactive in identity_8", s.ID)
			}
		}
		return
	}
	t.Fatal("liu_bei not found in identity_8 hero list")
}
