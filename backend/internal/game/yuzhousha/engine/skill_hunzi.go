package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) AwakenHunzi(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillHunzi) {
		return ErrWrongPhase
	}
	p := &g.Players[seat]
	if p.HP > 1 {
		return ErrInvalidTarget
	}
	if p.MaxHP <= 1 {
		return ErrInvalidTarget
	}

	p.MaxHP--
	p.Character.MaxHP = p.MaxHP
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	g.removeSkillFromPlayer(seat, SkillHunzi)
	g.grantSkillsToPlayer(seat, []string{SkillYingzi, SkillYinghun})

	msg := fmt.Sprintf("%s 觉醒【魂姿】，体力上限 -1，获得【英姿】【英魂】", p.Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDHunzi, seat, seat, msg)
	*events = append(*events, GameEvent{
		Type:        "skill_awaken",
		PlayerIndex: seat,
		SkillID:     skill.IDHunzi,
		Message:     msg,
	})
	g.resetTimer()
	return nil
}
