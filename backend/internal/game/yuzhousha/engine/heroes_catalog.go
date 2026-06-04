package engine

import (
	"fmt"
	"math"

	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	defaultHeroPage     = 1
	defaultHeroPageSize = 20
	maxHeroPageSize     = 100
)

// HeroesQuery filters and paginates the public hero catalog.
type HeroesQuery struct {
	Mode     string
	Kingdom  string
	Pack     string
	Page     int
	PageSize int
}

// HeroesPage is a paginated hero catalog response.
type HeroesPage struct {
	Heroes     []HeroPublic `json:"heroes"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// HeroPublic is API-facing hero metadata including display fields.
type HeroPublic struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	MaxHP         int               `json:"max_hp"`
	Kingdom       string            `json:"kingdom,omitempty"`
	SkillIDs      []string          `json:"skill_ids,omitempty"`
	Skills        []SkillMeta       `json:"skills,omitempty"`
	AccentColor   string            `json:"accent_color,omitempty"`
	PortraitURL   string            `json:"portrait_url,omitempty"`
	Pack          string            `json:"pack,omitempty"`
	DefaultSkinID string            `json:"default_skin_id,omitempty"`
	Display       skill.HeroDisplay `json:"display,omitempty"`
}

// ListHeroes returns heroes matching query and mode hero pool.
func ListHeroes(q HeroesQuery) HeroesPage {
	modeID := mode.NormalizeID(q.Mode)
	meta, _ := mode.Lookup(modeID)
	pool := meta.HeroPool

	filtered := make([]skill.CharacterDef, 0, len(skill.PickableCharacters()))
	for _, def := range skill.PickableCharacters() {
		if heroMatchesQuery(def, pool, q) {
			filtered = append(filtered, def)
		}
	}

	page, pageSize := normalizeHeroPage(q.Page, q.PageSize)
	total := len(filtered)
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	slice := filtered[start:end]
	heroes := make([]HeroPublic, len(slice))
	for i, def := range slice {
		heroes[i] = buildHeroPublic(def)
	}

	return HeroesPage{
		Heroes:     heroes,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// ValidateHeroForMode ensures a character may be picked in the given mode.
func ValidateHeroForMode(gameMode, charID string) error {
	def, ok := skill.CharacterByID(charID)
	if !ok {
		return fmt.Errorf("unknown character: %s", charID)
	}
	modeID := mode.NormalizeID(gameMode)
	meta, ok := mode.Lookup(modeID)
	if !ok {
		return nil
	}
	if !heroAllowedForPool(meta.HeroPool, def) {
		return fmt.Errorf("character not available in mode %s", modeID)
	}
	return nil
}

func heroMatchesQuery(def skill.CharacterDef, pool mode.HeroPoolSpec, q HeroesQuery) bool {
	if !heroAllowedForPool(pool, def) {
		return false
	}
	if q.Kingdom != "" && def.Kingdom != q.Kingdom {
		return false
	}
	if q.Pack != "" && def.Pack != q.Pack {
		return false
	}
	return true
}

func heroAllowedForPool(pool mode.HeroPoolSpec, def skill.CharacterDef) bool {
	if len(pool.Packs) > 0 {
		pack := def.Pack
		if pack == "" {
			pack = "standard"
		}
		found := false
		for _, p := range pool.Packs {
			if pack == p {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(pool.Kingdoms) > 0 {
		for _, k := range pool.Kingdoms {
			if def.Kingdom == k {
				return true
			}
		}
		return false
	}
	return true
}

func buildHeroPublic(def skill.CharacterDef) HeroPublic {
	c := buildCharacter(def.ID)
	display := skill.ResolveHeroDisplay(def.ID, "")
	return HeroPublic{
		ID:            c.ID,
		Name:          c.Name,
		MaxHP:         c.MaxHP,
		Kingdom:       c.Kingdom,
		SkillIDs:      c.SkillIDs,
		Skills:        c.Skills,
		AccentColor:   display.AccentColor,
		PortraitURL:   display.PortraitURL,
		Pack:          def.Pack,
		DefaultSkinID: display.SkinID,
		Display:       display,
	}
}

func normalizeHeroPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = defaultHeroPage
	}
	if pageSize < 1 {
		pageSize = defaultHeroPageSize
	}
	if pageSize > maxHeroPageSize {
		pageSize = maxHeroPageSize
	}
	return page, pageSize
}

// HeroesCatalog returns all pickable heroes (legacy flat list).
func HeroesCatalog() []Character {
	out := make([]Character, 0)
	for _, def := range skill.PickableCharacters() {
		out = append(out, buildCharacter(def.ID))
	}
	return out
}
