package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) ApplyTieqi(seat int, events *[]GameEvent) error {
	pending := g.Pending
	if pending == nil || !pending.TieqiPending || pending.SourceIndex != seat {
		return ErrWrongPhase
	}
	return g.startJudge(seat, skill.JudgeTieqi, guicaiResumeTieqi, events)
}

func (g *Game) SkipTieqi(seat int, events *[]GameEvent) error {
	if g.Pending == nil || !g.Pending.TieqiPending || g.Pending.SourceIndex != seat {
		return ErrWrongPhase
	}
	g.Pending.TieqiPending = false
	g.Message = fmt.Sprintf("%s 未发动【铁骑】", g.Players[seat].Name)
	g.resetTimer()
	return g.advanceShaBeforeTargetResponse(events)
}

func (r *gameSkillRuntime) DrawCards(seat, count int) error {
	r.g.drawCards(seat, count, r.events)
	return nil
}

func (r *gameSkillRuntime) DrawSkillCards(seat int, skillID string, count int, message string) error {
	return r.g.drawSkillCards(seat, skillID, count, message, r.events)
}

func (r *gameSkillRuntime) IsSeatInDyingRescue(seat int) bool {
	return r.g.isSeatInDyingRescue(seat)
}

func (r *gameSkillRuntime) ApplyTieqi(seat int) error {
	return r.g.ApplyTieqi(seat, r.events)
}

func (r *gameSkillRuntime) SkipTieqi(seat int) error {
	return r.g.SkipTieqi(seat, r.events)
}

func (r *gameSkillRuntime) PendingTieqiForSource(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	p := r.g.Pending
	return p.TieqiPending && p.SourceIndex == seat && p.Card.Kind == CardSha
}

func (r *gameSkillRuntime) FankuiTakeFrom(seat int, zone, cardID string) error {
	return r.g.FankuiTakeFrom(seat, zone, cardID, r.events)
}

func (r *gameSkillRuntime) PassFankui(seat int) error {
	return r.g.PassFankui(seat, r.events)
}

func (r *gameSkillRuntime) PendingFankuiFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillFankui &&
		r.g.Pending.TargetIndex == seat &&
		r.g.Pending.FankuiRemaining > 0
}

func (r *gameSkillRuntime) FankuiSourceSeat(actor int) int {
	if r.g.Pending == nil || r.g.Pending.ResponseMode != ResponseModeSkillFankui {
		return -1
	}
	if r.g.Pending.TargetIndex != actor {
		return -1
	}
	return r.g.Pending.SourceIndex
}

func (r *gameSkillRuntime) FirstTakeableCardID(target int) string {
	if len(r.g.Players[target].Hand) > 0 {
		return r.g.Players[target].Hand[0].ID
	}
	p := &r.g.Players[target]
	for _, slot := range []*Card{p.Weapon, p.Armor, p.PlusHorse, p.MinusHorse} {
		if slot != nil {
			return slot.ID
		}
	}
	if len(p.JudgeArea) > 0 {
		return p.JudgeArea[0].ID
	}
	return ""
}

func (r *gameSkillRuntime) ApplyGuicaiReplace(seat int, handCardID string) error {
	return r.g.ApplyGuicaiReplace(seat, handCardID, r.events)
}

func (r *gameSkillRuntime) PassGuicai(seat int) error {
	return r.g.PassGuicai(seat, r.events)
}

func (r *gameSkillRuntime) PendingGuicaiFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillGuicai && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) StartLuoshen(seat int) error {
	return r.g.StartLuoshen(seat, r.events)
}

func (r *gameSkillRuntime) PendingJianxiongFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillJianxiong && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) ApplyJianxiong(seat int) error {
	return r.g.ApplyJianxiong(seat, r.events)
}

func (r *gameSkillRuntime) PassJianxiong(seat int) error {
	return r.g.PassJianxiong(seat, r.events)
}

func (r *gameSkillRuntime) PendingYijiOfferFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillYijiOffer && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) PendingYijiGiveFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillYijiGive && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) ApplyYiji(seat int) error {
	return r.g.ApplyYiji(seat, r.events)
}

func (r *gameSkillRuntime) YijiGiveCards(seat, target int, cardIDs []string) error {
	return r.g.YijiGiveCards(seat, target, cardIDs, r.events)
}

func (r *gameSkillRuntime) PassYijiOffer(seat int) error {
	return r.g.PassYijiOffer(seat, r.events)
}

func (r *gameSkillRuntime) PassYijiGive(seat int) error {
	return r.g.PassYijiGive(seat, r.events)
}

func (r *gameSkillRuntime) PendingGanglieOfferFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillGanglieOffer && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) StartGanglieJudge(seat int) error {
	return r.g.StartGanglieJudge(seat, r.events)
}

func (r *gameSkillRuntime) PassGanglieOffer(seat int) error {
	return r.g.PassGanglieOffer(seat, r.events)
}

func (r *gameSkillRuntime) PendingGanglieChoiceFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillGanglieChoice && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) GanglieTakeDamage(seat int) error {
	return r.g.GanglieTakeDamage(seat, r.events)
}

func (r *gameSkillRuntime) GanglieDiscard(seat int, cardIDs []string) error {
	return r.g.GanglieDiscard(seat, cardIDs, r.events)
}

func (r *gameSkillRuntime) ActivateLuoyi(seat int) error {
	return r.g.ActivateLuoyi(seat, r.events)
}

func (r *gameSkillRuntime) PendingDrawPhaseChoiceFor(seat int) bool {
	return r.g.isDrawPhaseChoicePending(seat)
}

func (r *gameSkillRuntime) StartTuxi(seat int) error {
	return r.g.StartTuxi(seat, r.events)
}

func (r *gameSkillRuntime) TuxiTakeFrom(seat int, zone, cardID string) error {
	return r.g.TuxiTakeFrom(seat, zone, cardID, r.events)
}

func (r *gameSkillRuntime) PassTuxi(seat int) error {
	return r.g.PassTuxi(seat, r.events)
}

func (r *gameSkillRuntime) PendingTuxiTakeFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillTuxi && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) TuxiSourceSeat(actor int) int {
	if r.g.Pending == nil || r.g.Pending.ResponseMode != ResponseModeSkillTuxi {
		return -1
	}
	if r.g.Pending.TargetIndex != actor {
		return -1
	}
	return r.g.Pending.SourceIndex
}

func (r *gameSkillRuntime) OpponentHasTakeableCard(seat int) bool {
	return r.g.hasTakeableCard(r.g.opponentOf(seat))
}

func (r *gameSkillRuntime) BestTakeTarget(target int) (zone, cardID string) {
	return aiPickTakeTarget(r.g, target)
}

func (r *gameSkillRuntime) ActivateZhiheng(seat int, cardIDs []string) error {
	return r.g.ActivateZhiheng(seat, cardIDs, r.events)
}

func (r *gameSkillRuntime) ActivateJieyin(seat, target int, cardIDs []string) error {
	return r.g.ActivateJieyin(seat, target, cardIDs, r.events)
}

func (r *gameSkillRuntime) HasRedHandCard(seat int) bool {
	return r.g.hasRedHandCard(seat)
}

func (r *gameSkillRuntime) ActivateFanjian(seat int, cardID string) error {
	return r.g.ActivateFanjian(seat, cardID, r.events)
}

func (r *gameSkillRuntime) ResolveFanjianSuit(seat int, suit string) error {
	return r.g.ResolveFanjianSuit(seat, suit, r.events)
}

func (r *gameSkillRuntime) ApplyTianxiang(seat int, cardID string) error {
	return r.g.ApplyTianxiang(seat, cardID, r.events)
}

func (r *gameSkillRuntime) PassTianxiang(seat int) error {
	return r.g.PassTianxiang(seat, r.events)
}





func (r *gameSkillRuntime) ActivateYinghun(seat, target int) error {
	return r.g.ActivateYinghun(seat, target, r.events)
}

func (r *gameSkillRuntime) ResolveYinghunChoice(seat int, option string) error {
	return r.g.ResolveYinghunChoice(seat, option, r.events)
}

func (r *gameSkillRuntime) YinghunDiscard(seat int, cardIDs []string) error {
	return r.g.YinghunDiscard(seat, cardIDs, r.events)
}

func (r *gameSkillRuntime) ActivateGuose(seat, target int, cardID string) error {
	return r.g.ActivateGuose(seat, target, cardID, r.events)
}

func (r *gameSkillRuntime) HasDiamondHandCard(seat int) bool {
	return r.g.hasDiamondHandCard(seat)
}

func (r *gameSkillRuntime) ApplyLiuli(seat int, cardID string, redirect int) error {
	return r.g.ApplyLiuli(seat, cardID, redirect, r.events)
}

func (r *gameSkillRuntime) PassLiuli(seat int) error {
	return r.g.PassLiuli(seat, r.events)
}

func (r *gameSkillRuntime) ActivateKurou(seat int) error {
	return r.g.ActivateKurou(seat, r.events)
}

func (r *gameSkillRuntime) AwakenHunzi(seat int) error {
	return r.g.AwakenHunzi(seat, r.events)
}

func (r *gameSkillRuntime) ActivateShuangxiongDraw(seat int) error {
	return r.g.ActivateShuangxiongDraw(seat, r.events)
}

func (r *gameSkillRuntime) ActivateShuangxiongJuedou(seat int, cardID string) error {
	return r.g.ActivateShuangxiongJuedou(seat, cardID, r.events)
}

func (r *gameSkillRuntime) HasShuangxiongJuedouCard(seat int) bool {
	return r.g.hasShuangxiongJuedouCard(seat)
}

func (r *gameSkillRuntime) ActivateLuanwu(seat int) error {
	return r.g.ActivateLuanwu(seat, r.events)
}

func (r *gameSkillRuntime) PendingGuidaoFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillGuidao && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) ApplyGuidaoReplace(seat int, handCardID string) error {
	return r.g.ApplyGuidaoReplace(seat, handCardID, r.events)
}

func (r *gameSkillRuntime) PassGuidao(seat int) error {
	return r.g.PassGuidao(seat, r.events)
}

func (r *gameSkillRuntime) PendingLeijiOfferFor(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	return r.g.Pending.ResponseMode == ResponseModeSkillLeijiOffer && r.g.Pending.TargetIndex == seat
}

func (r *gameSkillRuntime) StartLeijiJudge(seat int) error {
	return r.g.StartLeijiJudge(seat, r.events)
}

func (r *gameSkillRuntime) PassLeijiOffer(seat int) error {
	return r.g.PassLeijiOffer(seat, r.events)
}
