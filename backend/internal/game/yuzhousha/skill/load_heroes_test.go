package skill

import (
	"sort"
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
	if len(loaded) != standardHeroCount {
		t.Fatalf("pickable count: got %d want %d", len(loaded), standardHeroCount)
	}

	loadedIDs := make([]string, len(loaded))
	for i, c := range loaded {
		loadedIDs[i] = c.ID
	}
	sort.Strings(loadedIDs)

	parsed, err := ParseHeroesJSON(yzsdata.StandardHeroesJSON)
	if err != nil {
		t.Fatal(err)
	}
	jsonIDs := make([]string, len(parsed))
	for i, h := range parsed {
		jsonIDs[i] = h.def.ID
	}
	sort.Strings(jsonIDs)

	if len(loadedIDs) != len(jsonIDs) {
		t.Fatalf("id count mismatch: loaded=%d json=%d", len(loadedIDs), len(jsonIDs))
	}
	for i := range loadedIDs {
		if loadedIDs[i] != jsonIDs[i] {
			t.Fatalf("pickable ids mismatch at %d: loaded=%s json=%s", i, loadedIDs[i], jsonIDs[i])
		}
	}
}
