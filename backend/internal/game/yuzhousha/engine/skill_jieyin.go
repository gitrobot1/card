package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const counterJieyinUsed = "jieyin_used_play"

func (g *Game) ActivateJieyin(seat, target int, cardIDs []string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillJieyin) || g.getSkillCounter(seat, counterJieyinUsed) > 0 {
		return ErrWrongPhase
	}
	if target < 0 || target >= len(g.Players) || target == seat || len(cardIDs) != 2 {
		return ErrInvalidTarget
	}
	tp := &g.Players[target]
	sp := &g.Players[seat]
	if tp.HP >= sp.HP || tp.HP >= tp.MaxHP || sp.HP >= sp.MaxHP {
		return ErrInvalidTarget
	}
	discarded := make([]Card, 0, 2)
	for _, id := range cardIDs {
		idx, _, ok := g.findCard(seat, id)
		if !ok {
			return ErrInvalidCard
		}
		discarded = append(discarded, g.removeHandCard(seat, idx, events))
	}
	g.DiscardPile = append(g.DiscardPile, discarded...)
	g.SyncCounts()
	g.runCardsDiscardedHooks(seat, "cost", discarded, events)

	sp.HP++
	tp.HP++
	g.setSkillCounter(seat, counterJieyinUsed, 1)
	msg := fmt.Sprintf("%s 发动【结姻】，与 %s 各回复 1 点体力", sp.Name, tp.Name)
	g.appendSkillEvent(events, skill.IDJieyin, seat, target, msg)
	for _, healed := range []int{seat, target} {
		p := &g.Players[healed]
		*events = append(*events, GameEvent{
			Type:        "skill_heal",
			PlayerIndex: healed,
			TargetIndex: target,
			SkillID:     skill.IDJieyin,
			Heal:        1,
			Message:     fmt.Sprintf("%s 体力 %d/%d", p.Name, p.HP, p.MaxHP),
		})
	}
	g.Message = msg
	g.resetTimer()
	return nil
}

func (g *Game) canJieyinTarget(actor, target int) bool {
	if target < 0 || target >= len(g.Players) || target == actor {
		return false
	}
	tp := &g.Players[target]
	sp := &g.Players[actor]
	return tp.HP < sp.HP && tp.HP < tp.MaxHP && sp.HP < sp.MaxHP
}
