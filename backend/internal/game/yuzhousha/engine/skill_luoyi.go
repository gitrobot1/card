package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) shouldOfferLuoyiDrawChoice(seat int) bool {
	return g.hasSkill(seat, SkillLuoyi)
}

func (g *Game) PassDrawPhase(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isDrawPhaseChoicePending(seat) {
		return ErrWrongPhase
	}
	g.setSkillCounter(seat, counterDrawChoicePending, 0)
	g.drawCards(seat, g.drawCountFor(seat), events)
	return g.advanceTurnAfterDraw(seat, events)
}

func (g *Game) ActivateLuoyi(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isDrawPhaseChoicePending(seat) {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillLuoyi) {
		return ErrInvalidCard
	}
	g.setSkillCounter(seat, counterDrawChoicePending, 0)
	g.setSkillCounter(seat, counterLuoyiActive, 1)
	msg := fmt.Sprintf("%s 发动【裸衣】，跳过摸牌，本回合【杀】伤害+1", g.Players[seat].Name)
	g.appendSkillEvent(events, skill.IDLuoyi, seat, seat, msg)
	g.Message = msg
	return g.advanceTurnAfterDraw(seat, events)
}

func (g *Game) advanceTurnAfterDraw(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	if g.Players[seat].SkipPlay {
		if g.Players[seat].hasJudgeKind(CardLeBu) {
			g.startWuxiekLebuJudgeWindow(seat, events)
			return nil
		}
		g.applyLebuSkipDirect(seat, events)
		return nil
	}
	g.TurnStep = StepPlay
	g.resetTimer()
	return nil
}

func (g *Game) shaBaseDamage(seat int) int {
	damage := 1
	if g.Players[seat].Drunk {
		damage++
		g.Players[seat].Drunk = false
	}
	if g.getSkillCounter(seat, counterLuoyiActive) > 0 {
		damage++
	}
	return damage
}
