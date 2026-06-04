package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	ResponseModeDdzJudgeCancel = "ddz_judge_cancel"
	SkillDdzJudgeCancel        = "ddz_judge_cancel"

	ddzResumeLebu      = "ddz_lebu"
	ddzResumeBingliang  = "ddz_bingliang"
)

func (g *Game) DdzLandlordSeat() int { return g.LandlordSeat }

func (g *Game) landlordDrawBonus(seat int) int {
	if g.is3pDdz() && seat == g.LandlordSeat {
		return 1
	}
	return 0
}

func (g *Game) canLandlordCancelJudge(seat int) bool {
	return g.is3pDdz() && seat == g.LandlordSeat && len(g.Players[seat].Hand) >= 2
}

func (g *Game) offerDdzJudgeCancelWindow(judgeSeat int, reason skill.JudgeReason, resume string, card Card, events *[]GameEvent) bool {
	if !g.canLandlordCancelJudge(judgeSeat) {
		return false
	}
	var saved *PendingCombat
	if g.Pending != nil {
		copy := *g.Pending
		saved = &copy
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:     judgeSeat,
		TargetIndex:     g.LandlordSeat,
		ResponseMode:    ResponseModeDdzJudgeCancel,
		JudgeCard:       card,
		JudgeReason:     string(reason),
		GuicaiResume:    resume,
		GuicaiJudgeSeat: judgeSeat,
		SavedPending:    saved,
	}
	g.Message = fmt.Sprintf("%s 可弃置两张手牌，取消此次判定", g.Players[g.LandlordSeat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "ddz_judge_cancel_offer",
		PlayerIndex: g.LandlordSeat,
		TargetIndex: judgeSeat,
		Card:        &card,
		Message:     g.Message,
	})
	return true
}

func (g *Game) offerDdzTrickCancelWindow(seat int, resume string, events *[]GameEvent) bool {
	if !g.canLandlordCancelJudge(seat) {
		return false
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:     seat,
		TargetIndex:     g.LandlordSeat,
		ResponseMode:    ResponseModeDdzJudgeCancel,
		GuicaiResume:    resume,
		GuicaiJudgeSeat: seat,
	}
	switch resume {
	case ddzResumeLebu:
		g.Message = fmt.Sprintf("%s 可弃置两张手牌，抵消【乐不思蜀】", g.Players[g.LandlordSeat].Name)
	case ddzResumeBingliang:
		g.Message = fmt.Sprintf("%s 可弃置两张手牌，抵消【兵粮寸断】", g.Players[g.LandlordSeat].Name)
	default:
		g.Message = fmt.Sprintf("%s 可弃置两张手牌，取消此次判定", g.Players[g.LandlordSeat].Name)
	}
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "ddz_judge_cancel_offer",
		PlayerIndex: g.LandlordSeat,
		TargetIndex: seat,
		Message:     g.Message,
	})
	return true
}

func (g *Game) ApplyDdzJudgeCancel(seat int, cardIDs []string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeDdzJudgeCancel || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	if len(cardIDs) != 2 {
		return ErrInvalidDiscardCount
	}
	seen := make(map[string]struct{}, 2)
	for _, id := range cardIDs {
		if id == "" {
			return ErrInvalidCard
		}
		if _, dup := seen[id]; dup {
			return ErrInvalidCard
		}
		seen[id] = struct{}{}
		if _, _, ok := g.findCard(seat, id); !ok {
			return ErrInvalidCard
		}
	}
	for _, id := range cardIDs {
		idx, _, ok := g.findCard(seat, id)
		if !ok {
			return ErrInvalidCard
		}
		played := g.removeHandCard(seat, idx, events)
		g.DiscardPile = append(g.DiscardPile, played)
	}
	g.syncCounts()

	resume := g.Pending.GuicaiResume
	judgeSeat := g.Pending.GuicaiJudgeSeat
	judgeCard := g.Pending.JudgeCard
	saved := g.Pending.SavedPending
	g.Pending = nil

	msg := fmt.Sprintf("%s 弃置两张牌，取消此次判定", g.Players[seat].Name)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "ddz_judge_cancel",
		PlayerIndex: seat,
		TargetIndex: judgeSeat,
		Message:     msg,
	})
	return g.finishDdzJudgeCancel(resume, judgeSeat, judgeCard, saved, events)
}

func (g *Game) PassDdzJudgeCancel(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeDdzJudgeCancel {
		return ErrWrongPhase
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	judgeCard := g.Pending.JudgeCard
	resume := g.Pending.GuicaiResume
	judgeSeat := g.Pending.GuicaiJudgeSeat
	reason := skill.JudgeReason(g.Pending.JudgeReason)
	saved := g.Pending.SavedPending
	g.Pending = saved

	switch resume {
	case ddzResumeLebu:
		g.Pending = nil
		g.Phase = PhasePlaying
		return g.applyLebuSkipDirectContinue(judgeSeat, events)
	case ddzResumeBingliang:
		g.Pending = nil
		g.Phase = PhasePlaying
		return g.applyBingliangSkipDrawContinue(judgeSeat, events)
	default:
		if g.offerGuicaiWindow(judgeSeat, reason, resume, judgeCard, events) {
			return nil
		}
		if g.offerGuidaoWindow(judgeSeat, reason, resume, judgeCard, events) {
			return nil
		}
		return g.completeJudgeResume(resume, judgeSeat, reason, judgeCard, events)
	}
}

func (g *Game) finishDdzJudgeCancel(resume string, judgeSeat int, judgeCard Card, saved *PendingCombat, events *[]GameEvent) error {
	if judgeCard.ID != "" {
		g.DiscardPile = append(g.DiscardPile, judgeCard)
	}
	g.Pending = saved

	switch resume {
	case guicaiResumeTieqi:
		if g.Pending != nil {
			g.Pending.TieqiPending = false
			g.Phase = PhaseResponse
			g.Message = fmt.Sprintf("%s 取消【铁骑】判定，目标仍可出【闪】", g.Players[judgeSeat].Name)
			g.resetTimer()
			return nil
		}
		return nil
	case guicaiResumeBagua:
		if g.Pending != nil {
			g.Pending.BaguaUsed = true
			g.Phase = PhaseResponse
			g.Message = fmt.Sprintf("%s 取消判定，请出【闪】或点「取消」", g.Players[judgeSeat].Name)
			g.resetTimer()
			return nil
		}
		return nil
	case guicaiResumeShandian:
		g.Pending = nil
		if card, ok := g.removeJudgeByKind(judgeSeat, CardShanDian); ok {
			g.DiscardPile = append(g.DiscardPile, card)
		}
		g.Phase = PhasePlaying
		g.TurnStep = StepDraw
		g.CurrentTurn = judgeSeat
		g.Message = fmt.Sprintf("【闪电】判定被取消")
		return g.resumeBeginTurnAfterLightning(judgeSeat, events)
	case guicaiResumeLuoshen:
		return g.continueAfterPrepare(judgeSeat, events)
	case guicaiResumeGanglie:
		g.Pending = nil
		if g.advanceDamageAftermath(events) {
			return nil
		}
		return nil
	case guicaiResumeLeiji:
		seat := g.leijiShanSeat
		savedLeiji := g.leijiSavedPending
		g.leijiSavedPending = nil
		g.leijiShanSeat = 0
		return g.finishShanDodgeSuccess(seat, savedLeiji, events, "")
	case ddzResumeLebu:
		p := &g.Players[judgeSeat]
		p.SkipPlay = false
		g.removeJudgeByKind(judgeSeat, CardLeBu)
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = judgeSeat
		g.Message = fmt.Sprintf("【乐不思蜀】被抵消，%s 可正常出牌", p.Name)
		g.resetTimer()
		return nil
	case ddzResumeBingliang:
		p := &g.Players[judgeSeat]
		p.SkipDraw = false
		g.removeJudgeByKind(judgeSeat, CardBingLiang)
		g.Phase = PhasePlaying
		g.TurnStep = StepDraw
		g.CurrentTurn = judgeSeat
		g.Message = fmt.Sprintf("【兵粮寸断】被抵消，%s 正常摸牌", p.Name)
		g.drawCards(judgeSeat, g.drawCountFor(judgeSeat), events)
		if g.IsFinished() {
			return nil
		}
		if p.SkipPlay {
			if p.hasJudgeKind(CardLeBu) {
				g.startWuxiekLebuJudgeWindow(judgeSeat, events)
				return nil
			}
			return g.applyLebuSkipDirectContinue(judgeSeat, events)
		}
		g.TurnStep = StepPlay
		g.resetTimer()
		return nil
	default:
		return nil
	}
}

func (g *Game) applyLebuSkipDirectContinue(seat int, events *[]GameEvent) error {
	g.applyLebuSkipDirect(seat, events)
	return nil
}

func (g *Game) applyBingliangSkipDrawContinue(seat int, events *[]GameEvent) error {
	g.applyBingliangSkipDraw(seat, events)
	if g.IsFinished() {
		return nil
	}
	if g.Players[seat].SkipPlay {
		if g.Players[seat].hasJudgeKind(CardLeBu) {
			g.startWuxiekLebuJudgeWindow(seat, events)
			return nil
		}
		return g.applyLebuSkipDirectContinue(seat, events)
	}
	g.TurnStep = StepPlay
	g.resetTimer()
	return nil
}
