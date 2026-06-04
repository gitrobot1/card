package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const ResponseModeSkillTianxiang = "skill_tianxiang"

func (g *Game) canOfferTianxiang(victim, amount int) bool {
	if amount <= 0 || !g.hasSkill(victim, SkillTianxiang) {
		return false
	}
	if !g.hasRedHandCard(victim) {
		return false
	}
	vhp, _ := g.PlayerHP(victim)
	for i := range g.Players {
		if i == victim || g.Players[i].HP <= 0 {
			continue
		}
		if g.Players[i].HP >= vhp {
			return true
		}
	}
	return false
}

func (g *Game) tianxiangRedirectTarget(victim int) int {
	vhp, _ := g.PlayerHP(victim)
	best := -1
	for i := range g.Players {
		if i == victim || g.Players[i].HP <= 0 {
			continue
		}
		if g.Players[i].HP >= vhp {
			if best < 0 || g.Players[i].HP > g.Players[best].HP {
				best = i
			}
		}
	}
	if best >= 0 {
		return best
	}
	return g.opponentOf(victim)
}

func (g *Game) offerTianxiangWindow(source, victim, amount int, card Card, resume DamageResume, events *[]GameEvent) error {
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  victim,
		ReturnIndex:  resume.ReturnIndex,
		EffectTarget: g.tianxiangRedirectTarget(victim),
		Card:         card,
		Damage:       amount,
		ResponseMode: ResponseModeSkillTianxiang,
		SkillID:      skill.IDTianxiang,
	}
	g.pendingDamageResume = &resume
	msg := fmt.Sprintf("%s 可发动【天香】转移 %d 点伤害", g.Players[victim].Name, amount)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDTianxiang, victim, g.Pending.EffectTarget, msg)
	g.resetTimer()
	return nil
}

func (g *Game) ApplyTianxiang(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillTianxiang {
		return ErrNoPendingCombat
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || !g.isRedHandCard(seat, cardObj) {
		return ErrInvalidCard
	}
	redirect := g.tianxiangRedirectTarget(seat)
	if redirect < 0 || g.Players[redirect].HP < g.Players[seat].HP {
		return ErrInvalidTarget
	}

	pending := *g.Pending
	resume := g.pendingDamageResume
	g.Pending = nil
	g.pendingDamageResume = nil

	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.syncCounts()

	msg := fmt.Sprintf("%s 发动【天香】，弃 %s，伤害转给 %s", g.Players[seat].Name, discarded.Label, g.Players[redirect].Name)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "skill_tianxiang",
		PlayerIndex: seat,
		TargetIndex: redirect,
		Card:        &discarded,
		SkillID:     skill.IDTianxiang,
		Message:     msg,
	})

	if resume == nil {
		r := DamageResume{ReturnIndex: pending.ReturnIndex}
		resume = &r
	}
	return g.finalizeDamageHit(pending.SourceIndex, redirect, pending.Damage, pending.Card, *resume, events)
}

func (g *Game) PassTianxiang(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillTianxiang {
		return ErrNoPendingCombat
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	pending := *g.Pending
	resume := g.pendingDamageResume
	g.Pending = nil
	g.pendingDamageResume = nil
	r := DamageResume{ReturnIndex: pending.ReturnIndex, SkipTianxiang: true}
	if resume != nil {
		r = *resume
		r.SkipTianxiang = true
	}
	return g.finalizeDamageHit(pending.SourceIndex, pending.TargetIndex, pending.Damage, pending.Card, r, events)
}

func (g *Game) finalizeDamageHit(source, target, damage int, card Card, resume DamageResume, events *[]GameEvent) error {
	if damage <= 0 {
		damage = 1
	}
	if !resume.SkipTianxiang && g.canOfferTianxiang(target, damage) {
		return g.offerTianxiangWindow(source, target, damage, card, resume, events)
	}

	g.applyDamage(source, target, damage, card, events)
	victim := &g.Players[target]

	eventType := "trick_hit"
	if card.Kind == CardSha {
		eventType = "sha_hit"
	}
	*events = append(*events, GameEvent{
		Type:        eventType,
		PlayerIndex: source,
		TargetIndex: target,
		Damage:      damage,
		Message:     g.damageMessage(victim, card.Name, damage),
	})

	if victim.HP <= 0 {
		if g.afterDamageApplied(source, target, damage, card, resume, events) {
			return nil
		}
	}

	if g.continueAfterDamage(source, target, damage, card, resume, events) {
		return nil
	}

	if resume.Mode == damageResumeFanjian {
		g.resumeAfterFanjianDamage(resume, events)
		return nil
	}
	if resume.ResumeLuanwu {
		return g.finishLuanwu(resume.LuanwuOwner, events)
	}
	if resume.LeijiResumeShan {
		return g.finishShanDodgeSuccess(resume.LeijiShanSeat, resume.LeijiSaved, events, "")
	}

	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = resume.ReturnIndex
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[resume.ReturnIndex].Name)
	g.resetTimer()
	return nil
}

func (g *Game) PlayerHP(seat int) (hp, maxHP int) {
	if seat < 0 || seat >= len(g.Players) {
		return 0, 0
	}
	p := &g.Players[seat]
	return p.HP, p.MaxHP
}
