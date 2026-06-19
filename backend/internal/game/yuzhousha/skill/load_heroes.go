package skill

import (
	"encoding/json"
	"fmt"

	yzsdata "github.com/time/card/backend/internal/game/yuzhousha/data"
)

type heroPackFile struct {
	Pack   string          `json:"pack"`
	Heroes []heroFileEntry `json:"heroes"`
}

type heroFileEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	MaxHP       int      `json:"max_hp"`
	Kingdom     string   `json:"kingdom"`
	Gender      string   `json:"gender,omitempty"` // male | female
	SkillIDs    []string `json:"skill_ids"`
	Pickable    *bool    `json:"pickable,omitempty"`
	AccentColor string   `json:"accent_color,omitempty"`
	PortraitURL string   `json:"portrait_url,omitempty"`
}

// LoadEmbeddedHeroes registers heroes from embedded JSON packs.
func LoadEmbeddedHeroes() error {
	for _, data := range [][]byte{
		yzsdata.StandardHeroesJSON,
		yzsdata.SPHeroesJSON,
		yzsdata.ShenHeroesJSON,
	} {
		if err := RegisterHeroesJSON(data); err != nil {
			return err
		}
	}
	return nil
}

// RegisterHeroesJSON parses a hero pack file and registers each entry.
func RegisterHeroesJSON(data []byte) error {
	heroes, err := ParseHeroesJSON(data)
	if err != nil {
		return err
	}
	for _, h := range heroes {
		RegisterCharacter(h.def, h.pickable)
	}
	return nil
}

type parsedHero struct {
	def      CharacterDef
	pickable bool
}

// ParseHeroesJSON decodes hero pack JSON without touching the registry.
func ParseHeroesJSON(data []byte) ([]parsedHero, error) {
	var file heroPackFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse heroes json: %w", err)
	}
	out := make([]parsedHero, 0, len(file.Heroes))
	for _, h := range file.Heroes {
		if h.ID == "" {
			return nil, fmt.Errorf("hero entry missing id")
		}
		if h.MaxHP <= 0 {
			return nil, fmt.Errorf("hero %s: invalid max_hp", h.ID)
		}
		pickable := true
		if h.Pickable != nil {
			pickable = *h.Pickable
		}
		out = append(out, parsedHero{
			def: CharacterDef{
				ID:          h.ID,
				Name:        h.Name,
				MaxHP:       h.MaxHP,
				Kingdom:     h.Kingdom,
				Gender:      h.Gender,
				SkillIDs:    append([]string(nil), h.SkillIDs...),
				Pack:        file.Pack,
				AccentColor: h.AccentColor,
				PortraitURL: h.PortraitURL,
			},
			pickable: pickable,
		})
	}
	return out, nil
}
