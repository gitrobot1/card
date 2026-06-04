package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) drawCountFor(seat int) int {
	base := DrawPerTurn
	rt := g.skillRuntime(nil)
	bonus := g.landlordDrawBonus(seat)
	for _, h := range g.playerSkillHandlers(seat) {
		bonus += h.DrawCountBonus(rt, seat)
	}
	return base + bonus
}

func (g *Game) drawSkillCards(seat int, skillID string, count int, message string, events *[]GameEvent) error {
	if count <= 0 {
		return nil
	}
	p := &g.Players[seat]
	if message == "" {
		skillName := skillID
		if h, ok := skill.Lookup(skillID); ok {
			skillName = h.Meta().Name
		}
		if count == 1 {
			message = fmt.Sprintf("%s 发动【%s】，摸一张牌", p.Name, skillName)
		} else {
			message = fmt.Sprintf("%s 发动【%s】，摸 %d 张牌", p.Name, skillName, count)
		}
	}
	g.Message = message
	g.appendSkillEvent(events, skillID, seat, seat, message)
	*events = append(*events, GameEvent{
		Type:        "skill_" + skillID,
		PlayerIndex: seat,
		SkillID:     skillID,
		Message:     message,
	})
	g.drawCards(seat, count, events)
	return nil
}

func (g *Game) runTurnEndHooks(seat int, events *[]GameEvent) {
	rt := g.skillRuntime(events)
	for _, h := range g.playerSkillHandlers(seat) {
		if err := h.OnTurnEnd(rt, seat); err != nil {
			return
		}
	}
}

func (g *Game) runHandEmptyHooks(seat int, events *[]GameEvent) {
	if seat < 0 || seat >= len(g.Players) || len(g.Players[seat].Hand) > 0 {
		return
	}
	rt := g.skillRuntime(events)
	for _, h := range g.playerSkillHandlers(seat) {
		if err := h.OnHandEmpty(rt, seat); err != nil {
			return
		}
	}
}
