package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	fankuiResumeShaHit    = "sha_hit"
	fankuiResumeLightning = "lightning"
)

func (g *Game) offerFankuiFromAftermath(a *DamageAftermath, events *[]GameEvent) bool {
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:       a.Source,
		TargetIndex:       a.Target,
		ReturnIndex:       a.Resume.ReturnIndex,
		ResponseMode:      ResponseModeSkillFankui,
		FankuiRemaining:   1,
		FankuiResumeMode:  a.Resume.Mode,
		FankuiResumeCard:  a.Resume.Card,
		FankuiReturnIndex: a.Resume.ReturnIndex,
	}
	a.FankuiLeft--
	g.Message = fmt.Sprintf("%s 可发动【反馈】获得 %s 一张牌", g.Players[a.Target].Name, g.Players[a.Source].Name)
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDFankui, a.Target, a.Source, g.Message)
	return true
}

func (g *Game) offerFankuiAfterDamage(target, source, damage int, resumeMode string, resumeCard Card, returnIndex int, events *[]GameEvent) bool {
	if damage <= 0 || !g.hasSkill(target, SkillFankui) || !g.hasTakeableCard(source) {
		return false
	}
	g.initDamageAftermath(source, target, damage, resumeCard, DamageResume{
		Mode: resumeMode, Card: resumeCard, ReturnIndex: returnIndex,
	})
	if g.damageAftermath == nil {
		return false
	}
	g.damageAftermath.OfferJianxiong = false
	g.damageAftermath.GanglieLeft = 0
	if g.damageAftermath.FankuiLeft <= 0 {
		g.damageAftermath = nil
		return false
	}
	return g.offerFankuiFromAftermath(g.damageAftermath, events)
}

func (g *Game) FankuiTakeFrom(seat int, zone, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillFankui || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	if g.Pending.FankuiRemaining <= 0 {
		return ErrWrongPhase
	}
	if zone == "" {
		zone = "hand"
	}
	source := g.Pending.SourceIndex
	spec := PlayTarget{SeatIndex: source, Zone: zone, CardID: cardID}
	card, label, ok := g.takeTargetCard(source, spec, events)
	if !ok {
		return ErrInvalidTarget
	}
	g.Players[seat].Hand = append(g.Players[seat].Hand, card)
	g.syncCounts()
	g.Pending.FankuiRemaining--
	msg := fmt.Sprintf("%s 发动【反馈】，获得 %s 的%s", g.Players[seat].Name, g.Players[source].Name, label)
	g.appendSkillEvent(events, skill.IDFankui, seat, source, msg)
	*events = append(*events, GameEvent{
		Type:        "fankui_take",
		PlayerIndex: seat,
		TargetIndex: source,
		Card:        &card,
		Message:     msg,
	})
	if g.Pending.FankuiRemaining > 0 && g.hasTakeableCard(source) {
		g.Message = fmt.Sprintf("%s 可继续发动【反馈】", g.Players[seat].Name)
		g.resetTimer()
		return nil
	}
	return g.resumeAfterFankui(events)
}

func (g *Game) PassFankui(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillFankui {
		return ErrWrongPhase
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	if g.Pending.FankuiRemaining > 0 {
		g.Pending.FankuiRemaining--
	}
	if g.Pending.FankuiRemaining > 0 && g.hasTakeableCard(g.Pending.SourceIndex) {
		g.Message = fmt.Sprintf("%s 跳过【反馈】，仍可再发动", g.Players[seat].Name)
		g.resetTimer()
		return nil
	}
	return g.resumeAfterFankui(events)
}

func (g *Game) resumeAfterFankui(events *[]GameEvent) error {
	g.Pending = nil
	if g.advanceDamageAftermath(events) {
		return nil
	}
	return nil
}

func (g *Game) applyTieqiJudgeResult(seat int, judgeCard Card, events *[]GameEvent) error {
	pending := g.Pending
	if pending == nil {
		return ErrWrongPhase
	}
	pending.TieqiPending = false
	msg := fmt.Sprintf("【铁骑】判定 %s", judgeCard.Label)
	if !isRedSuit(judgeCard.Suit) {
		pending.ShaUnblockable = true
		pending.RequiredKind = ""
		msg += "，目标不能出【闪】"
	} else {
		msg += "，目标仍可出【闪】"
	}
	g.Message = msg
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDTieqi, seat, pending.TargetIndex, msg)
	return g.advanceShaBeforeTargetResponse(events)
}

func (g *Game) applyBaguaJudgeResult(seat int, judgeCard Card, events *[]GameEvent) error {
	pending := g.Pending
	if pending == nil || seat != pending.TargetIndex {
		return ErrWrongPhase
	}
	pending.BaguaUsed = true
	red := isRedSuit(judgeCard.Suit)
	msg := fmt.Sprintf("%s 发动【八卦阵】判定 %s，", g.Players[seat].Name, judgeCard.Label)
	if red {
		msg += "红色，视为出【闪】"
	} else {
		msg += "黑色，仍需出【闪】"
	}
	*events = append(*events, GameEvent{
		Type:        "bagua_judge",
		PlayerIndex: seat,
		TargetIndex: pending.SourceIndex,
		Card:        &judgeCard,
		Message:     msg,
		Amount:      boolAmount(red),
	})
	if red {
		if g.consumeWushuangResponse(pending, seat, CardShan) {
			dodgeMsg := fmt.Sprintf("%s 的【八卦阵】视为出【闪】", g.Players[seat].Name)
			g.Message = dodgeMsg + "，" + g.Message
			return nil
		}
		dodgeMsg := fmt.Sprintf("%s 的【八卦阵】生效，【杀】无效", g.Players[seat].Name)
		return g.resolvePendingDodgeSuccess(seat, pending, events, dodgeMsg)
	}
	g.Message = fmt.Sprintf("%s 八卦阵判定失败，请出【闪】或点「取消」", g.Players[seat].Name)
	g.resetTimer()
	return nil
}

func (g *Game) applyShandianJudgeResult(seat int, judgeCard Card, events *[]GameEvent) error {
	strike := isLightningStrike(judgeCard.Suit, judgeCard.Rank)
	msg := fmt.Sprintf("%s 【闪电】判定 %s", g.Players[seat].Name, judgeCard.Label)
	if strike {
		msg += "，生效"
	} else {
		msg += "，传递"
	}
	*events = append(*events, GameEvent{
		Type:        "shandian_judge",
		PlayerIndex: seat,
		Card:        &judgeCard,
		Message:     msg,
		Amount:      boolAmount(strike),
	})
	if strike {
		lightning := g.Pending.Card
		if lightning.Kind == CardShanDian {
			g.DiscardPile = append(g.DiscardPile, lightning)
		}
		p := &g.Players[seat]
		source := g.opponentOf(seat)
		g.applyDamage(source, seat, 3, lightning, events)
		*events = append(*events, GameEvent{
			Type:        "trick_hit",
			PlayerIndex: source,
			TargetIndex: seat,
			Damage:      3,
			Message:     fmt.Sprintf("%s 受到【闪电】3 点伤害，体力 %d/%d", p.Name, p.HP, p.MaxHP),
		})
		g.Pending = nil
		if p.HP <= 0 {
			if g.afterDamageApplied(source, seat, 3, lightning, DamageResume{
				Mode: damageResumeLightning,
			}, events) {
				return nil
			}
		}
		if g.continueAfterDamage(source, seat, 3, lightning, DamageResume{
			Mode: damageResumeLightning,
		}, events) {
			return nil
		}
		return g.continueTurnAfterJudge(seat, events)
	}
	lightning := g.Pending.Card
	next := g.opponentOf(seat)
	if lightning.Kind == CardShanDian {
		g.setJudgeCard(next, lightning)
	}
	g.Pending = nil
	g.Message = fmt.Sprintf("【闪电】传至 %s 的判定区", g.Players[next].Name)
	return g.continueTurnAfterJudge(seat, events)
}
