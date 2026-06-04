package skill

import (
	"encoding/json"
	"fmt"
	"sort"

	yzsdata "github.com/time/card/backend/internal/game/yuzhousha/data"
)

// PackManifest describes a content pack (heroes + skins + future card sets).
type PackManifest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	HeroPack    string `json:"hero_pack"`
	SkinPack    string `json:"skin_pack,omitempty"`
}

var packManifests = map[string]PackManifest{}

// RegisterPack adds a pack manifest. Panics on duplicate id.
func RegisterPack(m PackManifest) {
	if m.ID == "" {
		panic("pack: register without id")
	}
	if _, exists := packManifests[m.ID]; exists {
		panic("pack: duplicate id " + m.ID)
	}
	packManifests[m.ID] = m
}

// PackByID returns a pack manifest.
func PackByID(id string) (PackManifest, bool) {
	m, ok := packManifests[id]
	return m, ok
}

// AllPacks returns registered packs sorted by id.
func AllPacks() []PackManifest {
	out := make([]PackManifest, 0, len(packManifests))
	for _, m := range packManifests {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// RegisterPackJSON parses and registers a pack manifest file.
func RegisterPackJSON(data []byte) error {
	var m PackManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("parse pack manifest: %w", err)
	}
	RegisterPack(m)
	return nil
}

// LoadEmbeddedPacks registers embedded pack manifests.
func LoadEmbeddedPacks() error {
	return RegisterPackJSON(yzsdata.StandardPackManifestJSON)
}

// LoadEmbeddedSkins registers embedded skins then ensures defaults for all heroes.
func LoadEmbeddedSkins() error {
	if err := RegisterSkinsJSON(yzsdata.StandardSkinsJSON); err != nil {
		return err
	}
	EnsureDefaultSkins()
	return nil
}
