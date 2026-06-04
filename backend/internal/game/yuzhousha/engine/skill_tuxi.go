package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const ResponseModeSkillTuxi = "skill_tuxi"

func (g *Game) shouldOfferTuxiDrawChoice(seat int) bool {
	if !g.hasSkill(seat, SkillTuxi) {
		return false
	}
	for _, e := range g.enemiesOf(seat) {
		if g.hasTakeableCard(e) {
			return true
		}
	}
	return false
}

func (g *Game) shouldOfferDrawPhaseChoice(seat int) bool {
	if g.hasSkill(seat, SkillLuoyi) {
		return true
	}
	if g.hasSkill(seat, SkillShuangxiong) {
		return true
	}
	return g.shouldOfferTuxiDrawChoice(seat)
}

func (g *Game) offerDrawPhaseChoice(seat int, events *[]GameEvent) {
	g.setSkillCounter(seat, counterDrawChoicePending, 1)
	g.TurnStep = StepDraw
	g.Message = fmt.Sprintf("%s 摸牌阶段", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "draw_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
}

func (g *Game) isDrawPhaseChoicePending(seat int) bool {
	return g.Phase == PhasePlaying && g.TurnStep == StepDraw && g.CurrentTurn == seat &&
		g.getSkillCounter(seat, counterDrawChoicePending) > 0
}

func (g *Game) StartTuxi(seat, skipCount int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isDrawPhaseChoicePending(seat) {
		return ErrWrongPhase
	}
	drawN := g.drawCountFor(seat)
	if !g.hasSkill(seat, SkillTuxi) || skipCount < 1 || skipCount > drawN {
		return ErrInvalidTarget
	}
	opp := g.firstEnemyWhere(seat, g.hasTakeableCard)
	if !g.hasTakeableCard(opp) {
		return ErrInvalidTarget
	}
	g.setSkillCounter(seat, counterDrawChoicePending, 0)
	g.setSkillCounter(seat, counterTuxiDrawSkip, skipCount)

	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   opp,
		TargetIndex:   seat,
		ReturnIndex:   seat,
		ResponseMode:  ResponseModeSkillTuxi,
		SkillID:       skill.IDTuxi,
		TuxiRemaining: skipCount,
	}
	msg := fmt.Sprintf("%s 发动【突袭】，少摸 %d 张，请选择获得 %s 的牌", g.Players[seat].Name, skipCount, g.Players[opp].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDTuxi, seat, opp, msg)
	g.resetTimer()
	return nil
}

func (g *Game) TuxiTakeFrom(seat int, zone, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillTuxi || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	if g.Pending.TuxiRemaining <= 0 {
		return ErrWrongPhase
	}
	if zone == "" {
		zone = "hand"
	}
	source := g.Pending.SourceIndex
	spec := PlayTarget{SeatIndex: source, Zone: zone, CardID: cardID}
	card, label, ok := g.takeTargetCard(source, spec, events)
	if !ok {
		return g.finishTuxi(seat, events)
	}
	g.Players[seat].Hand = append(g.Players[seat].Hand, card)
	g.syncCounts()
	g.Pending.TuxiRemaining--
	msg := fmt.Sprintf("%s 发动【突袭】，获得 %s 的%s", g.Players[seat].Name, g.Players[source].Name, label)
	g.appendSkillEvent(events, skill.IDTuxi, seat, source, msg)
	*events = append(*events, GameEvent{
		Type:        "tuxi_take",
		PlayerIndex: seat,
		TargetIndex: source,
		Card:        &card,
		Message:     msg,
	})
	if g.Pending.TuxiRemaining > 0 && g.hasTakeableCard(source) {
		g.Message = fmt.Sprintf("%s 继续选择【突袭】目标牌", g.Players[seat].Name)
		g.resetTimer()
		return nil
	}
	return g.finishTuxi(seat, events)
}

func (g *Game) PassTuxi(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillTuxi || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	return g.finishTuxi(seat, events)
}

func (g *Game) finishTuxi(seat int, events *[]GameEvent) error {
	skipped := g.getSkillCounter(seat, counterTuxiDrawSkip)
	if skipped <= 0 && g.Pending != nil {
		skipped = g.Pending.TuxiRemaining
	}
	g.setSkillCounter(seat, counterTuxiDrawSkip, 0)
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepDraw
	drawLeft := g.drawCountFor(seat) - skipped
	if drawLeft > 0 {
		g.drawCards(seat, drawLeft, events)
	}
	return g.advanceTurnAfterDraw(seat, events)
}

func aiPickTakeTarget(g *Game, target int) (zone, cardID string) {
	p := &g.Players[target]
	if p.Weapon != nil {
		return EquipWeapon, p.Weapon.ID
	}
	if p.Armor != nil {
		return EquipArmor, p.Armor.ID
	}
	if p.MinusHorse != nil {
		return EquipMinusHorse, p.MinusHorse.ID
	}
	if p.PlusHorse != nil {
		return EquipPlusHorse, p.PlusHorse.ID
	}
	if len(p.Hand) > 0 {
		return "hand", ""
	}
	if len(p.JudgeArea) > 0 {
		return "judge", p.JudgeArea[0].ID
	}
	return "", ""
}
