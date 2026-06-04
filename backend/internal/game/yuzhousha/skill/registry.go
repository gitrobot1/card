package skill

import (
	"math/rand"
	"sort"
	"time"
)

var (
	byID          = map[string]Handler{}
	characters    = map[string]CharacterDef{}
	pickableOrder []string
)

// Register 注册一个技能声明。
func Register(d Decl) {
	if d.Meta.ID == "" {
		panic("skill: register without id")
	}
	byID[d.Meta.ID] = Handler{Decl: d}
}

// Unregister 移除技能注册（仅测试用）。
func Unregister(id string) {
	delete(byID, id)
}

// RegisterCharacter 注册可选武将。
func RegisterCharacter(c CharacterDef, pickable bool) {
	characters[c.ID] = c
	if pickable {
		pickableOrder = append(pickableOrder, c.ID)
	}
}

func Lookup(id string) (Handler, bool) {
	h, ok := byID[id]
	return h, ok
}

func CharacterByID(id string) (CharacterDef, bool) {
	c, ok := characters[id]
	return c, ok
}

func PickableCharacters() []CharacterDef {
	out := make([]CharacterDef, 0, len(pickableOrder))
	for _, id := range pickableOrder {
		if c, ok := characters[id]; ok {
			out = append(out, c)
		}
	}
	return out
}

func RandomPickableCharacter(excludeID string) string {
	candidates := make([]string, 0, len(pickableOrder))
	for _, id := range pickableOrder {
		if id != excludeID {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return CharGuanYu
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return candidates[r.Intn(len(candidates))]
}

func MetasForCharacter(charID string) []Meta {
	c, ok := characters[charID]
	if !ok {
		return nil
	}
	out := make([]Meta, 0, len(c.SkillIDs))
	for _, id := range c.SkillIDs {
		if h, ok := byID[id]; ok {
			m := h.Meta()
			if m.Kind == KindLord {
				m.InactiveIn1v1 = true
			}
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func HandlersForCharacter(charID string) []Handler {
	c, ok := characters[charID]
	if !ok {
		return nil
	}
	out := make([]Handler, 0, len(c.SkillIDs))
	for _, id := range c.SkillIDs {
		if h, ok := byID[id]; ok {
			out = append(out, h)
		}
	}
	return out
}
