package skill

import (
	"testing"

	yzsdata "github.com/time/card/backend/internal/game/yuzhousha/data"
)

const standardHeroCount = 32

var standardKingdomCounts = map[string]int{
	"shu": 7,
	"wei": 8,
	"wu":  11,
	"qun": 6,
}

func TestParseHeroesJSONStandardPack(t *testing.T) {
	parsed, err := ParseHeroesJSON(yzsdata.StandardHeroesJSON)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != standardHeroCount {
		t.Fatalf("count: got %d want %d", len(parsed), standardHeroCount)
	}

	seen := make(map[string]struct{}, len(parsed))
	kingdomCounts := make(map[string]int)
	for _, h := range parsed {
		if h.def.ID == "" || h.def.Name == "" {
			t.Fatalf("hero missing id or name: %+v", h.def)
		}
		if h.def.MaxHP <= 0 {
			t.Fatalf("hero %s invalid max_hp: %d", h.def.ID, h.def.MaxHP)
		}
		if h.def.Kingdom == "" {
			t.Fatalf("hero %s missing kingdom", h.def.ID)
		}
		if len(h.def.SkillIDs) == 0 {
			t.Fatalf("hero %s has no skills", h.def.ID)
		}
		if !h.pickable {
			t.Fatalf("standard pack hero %s should be pickable", h.def.ID)
		}
		if _, dup := seen[h.def.ID]; dup {
			t.Fatalf("duplicate hero id: %s", h.def.ID)
		}
		seen[h.def.ID] = struct{}{}
		kingdomCounts[h.def.Kingdom]++
	}

	for kingdom, want := range standardKingdomCounts {
		if got := kingdomCounts[kingdom]; got != want {
			t.Fatalf("kingdom %s count: got %d want %d", kingdom, got, want)
		}
	}
}

func TestLoadEmbeddedHeroesPickableSet(t *testing.T) {
	loaded := PickableCharacters()
	if len(loaded) != 35 {
		t.Fatalf("pickable count: got %d want 35", len(loaded))
	}

	loadedSet := make(map[string]struct{}, len(loaded))
	for _, c := range loaded {
		loadedSet[c.ID] = struct{}{}
	}

	parsed, err := ParseHeroesJSON(yzsdata.StandardHeroesJSON)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range parsed {
		if _, ok := loadedSet[h.def.ID]; !ok {
			t.Fatalf("standard hero %s missing from pickable set", h.def.ID)
		}
	}
}

func TestLoadEmbeddedAllHeroesPickableCount(t *testing.T) {
	loaded := PickableCharacters()
	if len(loaded) != 35 {
		t.Fatalf("pickable count: got %d want 35", len(loaded))
	}
	ids := map[string]bool{}
	for _, c := range loaded {
		ids[c.ID] = true
	}
	for _, want := range []string{CharSpZhaoYun, CharShenZhaoYun} {
		if !ids[want] {
			t.Fatalf("missing hero %s in pickable set", want)
		}
	}
}

func TestParseHeroesJSONSPAndShen(t *testing.T) {
	sp, err := ParseHeroesJSON(yzsdata.SPHeroesJSON)
	if err != nil {
		t.Fatal(err)
	}
	if len(sp) != 2 {
		t.Fatalf("sp pack: want 2 heroes, got %d: %+v", len(sp), sp)
	}
	foundSP := false
	foundJie := false
	for _, h := range sp {
		if h.def.ID == CharSpZhaoYun {
			foundSP = true
		}
		if h.def.ID == "jie_xu_sheng" {
			foundJie = true
		}
	}
	if !foundSP || !foundJie {
		t.Fatalf("sp pack missing heroes: %+v", sp)
	}
	shen, err := ParseHeroesJSON(yzsdata.ShenHeroesJSON)
	if err != nil {
		t.Fatal(err)
	}
	if len(shen) != 1 || shen[0].def.ID != CharShenZhaoYun {
		t.Fatalf("shen pack: %+v", shen)
	}
}
