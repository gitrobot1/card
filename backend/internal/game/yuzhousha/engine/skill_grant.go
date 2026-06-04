package engine

import (
	"sort"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) removeSkillFromPlayer(seat int, skillID string) {
	p := &g.Players[seat]
	out := p.Character.SkillIDs[:0]
	for _, id := range p.Character.SkillIDs {
		if id != skillID {
			out = append(out, id)
		}
	}
	p.Character.SkillIDs = out
	g.syncPlayerSkillsMeta(seat)
}

func (g *Game) grantSkillsToPlayer(seat int, skillIDs []string) {
	p := &g.Players[seat]
	have := make(map[string]struct{}, len(p.Character.SkillIDs))
	for _, id := range p.Character.SkillIDs {
		have[id] = struct{}{}
	}
	for _, id := range skillIDs {
		if _, ok := have[id]; ok {
			continue
		}
		p.Character.SkillIDs = append(p.Character.SkillIDs, id)
		have[id] = struct{}{}
	}
	sort.Strings(p.Character.SkillIDs)
	g.syncPlayerSkillsMeta(seat)
}

func (g *Game) syncPlayerSkillsMeta(seat int) {
	p := &g.Players[seat]
	out := make([]SkillMeta, 0, len(p.Character.SkillIDs))
	for _, id := range p.Character.SkillIDs {
		h, ok := skill.Lookup(id)
		if !ok {
			continue
		}
		m := h.Meta()
		if m.Kind == skill.KindLord {
			m.InactiveIn1v1 = true
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	p.Character.Skills = out
}
