package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	ResponseModeSkillGanglieOffer  = "skill_ganglie_offer"
	ResponseModeSkillGanglieChoice = "skill_ganglie_choice"
	guicaiResumeGanglie          = "ganglie"
)

func isHeartSuit(suit string) bool { return suit == "H" }

func (g *Game) offerGanglieWindow(a *DamageAftermath, events *[]GameEvent) bool {
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		TargetIndex:  a.Target,
		SourceIndex:  a.Source,
		ReturnIndex:  a.Resume.ReturnIndex,
		ResponseMode: ResponseModeSkillGanglieOffer,
		GanglieOwner: a.Target,
		GanglieIndex: a.GanglieLeft,
	}
	g.Message = fmt.Sprintf("%s 可发动【刚烈】进行判定", g.Players[a.Target].Name)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDGanglie, a.Target, a.Source, g.Message)
	return true
}

func (g *Game) StartGanglieJudge(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillGanglieOffer || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	a := g.damageAftermath
	if a == nil {
		return ErrWrongPhase
	}
	a.GanglieLeft--
	g.Pending = nil
	return g.startJudge(seat, skill.JudgeGanglie, guicaiResumeGanglie, events)
}

func (g *Game) PassGanglieOffer(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillGanglieOffer || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	a := g.damageAftermath
	if a != nil && a.GanglieLeft > 0 {
		a.GanglieLeft--
	}
	g.Pending = nil
	if g.advanceDamageAftermath(events) {
		return nil
	}
	return nil
}

func (g *Game) applyGanglieJudgeResult(owner int, judgeCard Card, events *[]GameEvent) error {
	a := g.damageAftermath
	if a == nil {
		return nil
	}
	source := a.Source
	msg := fmt.Sprintf("【刚烈】判定 %s", judgeCard.Label)
	if isHeartSuit(judgeCard.Suit) {
		msg += "，红桃无效"
		g.appendSkillEvent(events, skill.IDGanglie, owner, source, msg)
		g.Pending = nil
		if g.advanceDamageAftermath(events) {
			return nil
		}
		return nil
	}
	msg += "，请选择弃两张牌或受到1点伤害"
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  owner,
		TargetIndex:  source,
		ReturnIndex:  a.Resume.ReturnIndex,
		ResponseMode: ResponseModeSkillGanglieChoice,
		GanglieOwner: owner,
	}
	g.Message = fmt.Sprintf("%s 的【刚烈】生效，%s 需弃2张手牌或受到1点伤害", g.Players[owner].Name, g.Players[source].Name)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDGanglie, owner, source, msg)
	return nil
}

func (g *Game) GanglieTakeDamage(source int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillGanglieChoice || g.Pending.TargetIndex != source {
		return ErrWrongPhase
	}
	owner := g.Pending.GanglieOwner
	g.Pending = nil
	g.applyDamageWithHook(owner, source, 1, Card{Name: "刚烈"}, events)
	msg := fmt.Sprintf("%s 受到【刚烈】1 点伤害", g.Players[source].Name)
	*events = append(*events, GameEvent{
		Type: "ganglie_damage", PlayerIndex: owner, TargetIndex: source, Damage: 1, Message: msg,
	})
	if g.Players[source].HP <= 0 {
		if g.afterDamageApplied(owner, source, 1, Card{Name: "刚烈"}, DamageResume{}, events) {
			g.damageAftermath = nil
			return nil
		}
	}
	if g.advanceDamageAftermath(events) {
		return nil
	}
	return nil
}

func (g *Game) GanglieDiscard(source int, cardIDs []string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillGanglieChoice || g.Pending.TargetIndex != source {
		return ErrWrongPhase
	}
	if len(cardIDs) != 2 {
		return ErrInvalidDiscardCount
	}
	owner := g.Pending.GanglieOwner
	seen := make(map[string]struct{}, 2)
	for _, id := range cardIDs {
		if _, dup := seen[id]; dup {
			return ErrInvalidCard
		}
		seen[id] = struct{}{}
		if _, _, ok := g.findCard(source, id); !ok {
			return ErrInvalidCard
		}
	}
	discarded := make([]Card, 0, 2)
	for _, id := range cardIDs {
		idx, _, ok := g.findCard(source, id)
		if !ok {
			return ErrInvalidCard
		}
		c := g.removeHandCard(source, idx, events)
		g.DiscardPile = append(g.DiscardPile, c)
		discarded = append(discarded, c)
	}
	g.syncCounts()
	g.Pending = nil
	msg := fmt.Sprintf("%s 弃置2张牌以响应【刚烈】", g.Players[source].Name)
	g.appendSkillEvent(events, skill.IDGanglie, owner, source, msg)
	for i := range discarded {
		*events = append(*events, GameEvent{
			Type: "discard", PlayerIndex: source, Card: &discarded[i], Message: msg,
		})
	}
	if g.advanceDamageAftermath(events) {
		return nil
	}
	return nil
}
