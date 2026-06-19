package engine

import (
	"testing"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func TestDrawCountForYingzi(t *testing.T) {
	g := &Game{
		Players: []Player{
			{Character: Character{SkillIDs: []string{SkillYingzi}}},
		},
	}
	if got := g.drawCountFor(0); got != 3 {
		t.Fatalf("drawCountFor with yingzi=%d want 3", got)
	}
}

func TestYingziAndLianyingDeclHooks(t *testing.T) {
	y, ok := skill.Lookup(skill.IDYingzi)
	if !ok || y.Decl.DrawCountBonus == nil {
		t.Fatal("yingzi should register DrawCountBonus")
	}
	l, ok := skill.Lookup(skill.IDLianying)
	if !ok || l.Decl.OnHandEmpty == nil {
		t.Fatal("lianying should register OnHandEmpty")
	}
}

func TestBiyueDeclOnTurnEnd(t *testing.T) {
	h, ok := skill.Lookup(skill.IDBiyue)
	if !ok || h.Decl.OnTurnEnd == nil {
		t.Fatal("biyue should register OnTurnEnd in catalog")
	}
}

func TestCatalogPassiveDeclHooks(t *testing.T) {
	cases := []struct {
		id   string
		check func(skill.Handler) bool
	}{
		{skill.IDKeji, func(h skill.Handler) bool { return h.Decl.SkipsDiscardPhase != nil }},
		{skill.IDJiang, func(h skill.Handler) bool { return h.Decl.OnCardResolved != nil }},
		{skill.IDWushuang, func(h skill.Handler) bool { return h.Decl.ExtraResponsesNeeded != nil }},
		{skill.IDWeimu, func(h skill.Handler) bool { return h.Decl.BlocksTrickTarget != nil }},
		{skill.IDWansha, func(h skill.Handler) bool { return h.Decl.BlocksPeachUse != nil }},
		{skill.IDJueqing, func(h skill.Handler) bool { return h.Decl.DamageAsHPLoss != nil }},
		{skill.IDShangshi, func(h skill.Handler) bool { return h.Decl.OnHPChanged != nil || h.Decl.OnCardsDiscarded != nil || h.Decl.OnTurnEnd != nil }},
		{skill.IDHongyan, func(h skill.Handler) bool { return h.Decl.EffectiveSuit != nil }},
		{skill.IDXiaoji, func(h skill.Handler) bool { return h.Decl.OnEquipLost != nil }},
	}
	for _, c := range cases {
		h, ok := skill.Lookup(c.id)
		if !ok || !c.check(h) {
			t.Fatalf("%s should register catalog Decl hook", c.id)
		}
	}
}
