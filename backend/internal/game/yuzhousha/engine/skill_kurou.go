package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) ActivateKurou(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillKurou) {
		return ErrWrongPhase
	}
	p := &g.Players[seat]
	if p.HP <= 1 {
		return ErrInvalidTarget
	}

	p.HP--
	g.drawCards(seat, 2, events)

	msg := fmt.Sprintf("%s 发动【苦肉】，失去 1 点体力并摸 2 张牌（体力 %d/%d）", p.Name, p.HP, p.MaxHP)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDKurou, seat, seat, msg)
	*events = append(*events, GameEvent{
		Type:        "skill_kurou",
		PlayerIndex: seat,
		TargetIndex: seat,
		SkillID:     skill.IDKurou,
		Damage:      1,
		Message:     msg,
	})
	g.resetTimer()
	return nil
}
