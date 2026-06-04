package skill

import (
	"encoding/json"
	"fmt"
	"sort"
)

// SkinDef is cosmetic metadata for a hero appearance (future shop / inventory).
type SkinDef struct {
	ID          string `json:"id"`
	HeroID      string `json:"hero_id"`
	Name        string `json:"name,omitempty"`
	Pack        string `json:"pack,omitempty"`
	PortraitURL string `json:"portrait_url,omitempty"`
	AccentColor string `json:"accent_color,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

// HeroDisplay is resolved appearance sent to clients.
type HeroDisplay struct {
	SkinID      string `json:"skin_id"`
	PortraitURL string `json:"portrait_url,omitempty"`
	AccentColor string `json:"accent_color,omitempty"`
}

type skinPackFile struct {
	Pack  string    `json:"pack"`
	Skins []SkinDef `json:"skins"`
}

var (
	skinsByID       = map[string]SkinDef{}
	defaultSkinHero = map[string]string{}
)

// DefaultSkinID returns the canonical default skin id for a hero.
func DefaultSkinID(heroID string) string {
	return heroID + ":default"
}

// RegisterSkin adds a skin definition. Panics on duplicate id.
func RegisterSkin(s SkinDef) {
	if s.ID == "" {
		panic("skin: register without id")
	}
	if s.HeroID == "" {
		panic("skin: register without hero_id for " + s.ID)
	}
	if _, exists := skinsByID[s.ID]; exists {
		panic("skin: duplicate id " + s.ID)
	}
	skinsByID[s.ID] = s
	if s.IsDefault {
		defaultSkinHero[s.HeroID] = s.ID
	}
}

// SkinByID looks up a skin by id.
func SkinByID(id string) (SkinDef, bool) {
	s, ok := skinsByID[id]
	return s, ok
}

// DefaultSkinForHero returns the default skin for a hero, if registered.
func DefaultSkinForHero(heroID string) (SkinDef, bool) {
	if id, ok := defaultSkinHero[heroID]; ok {
		return SkinByID(id)
	}
	s, ok := SkinByID(DefaultSkinID(heroID))
	return s, ok
}

// SkinsForHero returns all skins for a hero sorted by id.
func SkinsForHero(heroID string) []SkinDef {
	out := make([]SkinDef, 0)
	for _, s := range skinsByID {
		if s.HeroID == heroID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// RegisterSkinsJSON parses a skin pack file.
func RegisterSkinsJSON(data []byte) error {
	var file skinPackFile
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("parse skins json: %w", err)
	}
	for _, s := range file.Skins {
		if s.Pack == "" {
			s.Pack = file.Pack
		}
		RegisterSkin(s)
	}
	return nil
}

// EnsureDefaultSkins registers a classic skin for every known hero missing one.
func EnsureDefaultSkins() {
	for _, def := range PickableCharacters() {
		if _, ok := defaultSkinHero[def.ID]; ok {
			continue
		}
		if _, ok := skinsByID[DefaultSkinID(def.ID)]; ok {
			defaultSkinHero[def.ID] = DefaultSkinID(def.ID)
			continue
		}
		RegisterSkin(SkinDef{
			ID:          DefaultSkinID(def.ID),
			HeroID:      def.ID,
			Name:        "经典",
			Pack:        def.Pack,
			PortraitURL: def.PortraitURL,
			AccentColor: ResolveAccentColor(def),
			IsDefault:   true,
		})
	}
}

// ResolveHeroDisplay resolves portrait/accent for a hero + optional skin override.
// Empty skinID uses the hero default skin.
func ResolveHeroDisplay(heroID, skinID string) HeroDisplay {
	def, ok := CharacterByID(heroID)
	if !ok {
		return HeroDisplay{SkinID: skinID, AccentColor: "#6b7280"}
	}
	if skinID == "" {
		if ds, ok := DefaultSkinForHero(heroID); ok {
			skinID = ds.ID
		} else {
			skinID = DefaultSkinID(heroID)
		}
	}
	if skin, ok := SkinByID(skinID); ok && skin.HeroID == heroID {
		accent := skin.AccentColor
		if accent == "" {
			accent = ResolveAccentColor(def)
		}
		portrait := skin.PortraitURL
		if portrait == "" {
			portrait = def.PortraitURL
		}
		return HeroDisplay{
			SkinID:      skin.ID,
			PortraitURL: portrait,
			AccentColor: accent,
		}
	}
	return HeroDisplay{
		SkinID:      skinID,
		PortraitURL: def.PortraitURL,
		AccentColor: ResolveAccentColor(def),
	}
}
