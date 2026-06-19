package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	guicaiResumeTieqi    = "tieqi"
	guicaiResumeBagua    = "bagua"
	guicaiResumeShandian = "shandian"
	guicaiResumeLuoshen  = "luoshen"
)

// startJudge 翻判定牌；若有人可发动【鬼才】则挂起，否则立即完成后续。
func (g *Game) startJudge(judgeSeat int, reason skill.JudgeReason, resume string, events *[]GameEvent) error {
	card, ok := g.flipJudgeCard(events, judgeSeat)
	if !ok {
		return ErrInvalidCard
	}
	return g.afterJudgeFlip(judgeSeat, reason, resume, card, events)
}

func (g *Game) afterJudgeFlip(judgeSeat int, reason skill.JudgeReason, resume string, card Card, events *[]GameEvent) error {
	if g.offerDdzJudgeCancelWindow(judgeSeat, reason, resume, card, events) {
		return nil
	}
	if g.offerGuicaiWindow(judgeSeat, reason, resume, card, events) {
		return nil
	}
	if g.offerGuidaoWindow(judgeSeat, reason, resume, card, events) {
		return nil
	}
	return g.completeJudgeResume(resume, judgeSeat, reason, card, events)
}

func (g *Game) guicaiHolderSeat() int {
	for i := range g.Players {
		if g.hasSkill(i, SkillGuicai) && len(g.Players[i].Hand) > 0 {
			return i
		}
	}
	return -1
}

func (g *Game) offerGuicaiWindow(judgeSeat int, reason skill.JudgeReason, resume string, card Card, events *[]GameEvent) bool {
	guicaiSeat := g.guicaiHolderSeat()
	if guicaiSeat < 0 {
		return false
	}
	var saved *PendingCombat
	if g.Pending != nil {
		copy := *g.Pending
		saved = &copy
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:    judgeSeat,
		TargetIndex:    guicaiSeat,
		ResponseMode:   ResponseModeSkillGuicai,
		JudgeCard:      card,
		JudgeReason:    string(reason),
		GuicaiResume:   resume,
		GuicaiJudgeSeat: judgeSeat,
		SavedPending:   saved,
	}
	g.Message = fmt.Sprintf("%s 可对判定 %s 发动【鬼才】", g.Players[guicaiSeat].Name, card.Label)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDGuicai, guicaiSeat, judgeSeat, g.Message)
	return true
}

func (g *Game) ApplyGuicaiReplace(seat int, handCardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillGuicai || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	idx, _, ok := g.findCard(seat, handCardID)
	if !ok {
		return ErrInvalidCard
	}
	oldJudge := g.Pending.JudgeCard
	played := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, oldJudge)
	g.syncCounts()

	resume := g.Pending.GuicaiResume
	judgeSeat := g.Pending.GuicaiJudgeSeat
	reason := skill.JudgeReason(g.Pending.JudgeReason)
	saved := g.Pending.SavedPending

	msg := fmt.Sprintf("%s 发动【鬼才】，以 %s 代替判定 %s", g.Players[seat].Name, played.Label, oldJudge.Label)
	g.appendSkillEvent(events, skill.IDGuicai, seat, judgeSeat, msg)
	*events = append(*events, GameEvent{
		Type:        "guicai_replace",
		PlayerIndex: seat,
		TargetIndex: judgeSeat,
		Card:        &played,
		Message:     msg,
	})

	g.Pending = saved
	return g.completeJudgeResume(resume, judgeSeat, reason, played, events)
}

func (g *Game) PassGuicai(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillGuicai {
		return ErrWrongPhase
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	card := g.Pending.JudgeCard
	resume := g.Pending.GuicaiResume
	judgeSeat := g.Pending.GuicaiJudgeSeat
	reason := skill.JudgeReason(g.Pending.JudgeReason)
	saved := g.Pending.SavedPending
	g.Pending = saved
	return g.completeJudgeResume(resume, judgeSeat, reason, card, events)
}

func (g *Game) completeJudgeResume(resume string, judgeSeat int, reason skill.JudgeReason, card Card, events *[]GameEvent) error {
	g.runJudgeResultHooks(skill.JudgeCtx{
		Seat: judgeSeat, Reason: reason, Card: cardView(card), IsRed: isRedSuit(card.Suit),
	}, events)
	switch resume {
	case guicaiResumeTieqi:
		return g.applyTieqiJudgeResult(judgeSeat, card, events)
	case guicaiResumeBagua:
		return g.applyBaguaJudgeResult(judgeSeat, card, events)
	case guicaiResumeShandian:
		return g.applyShandianJudgeResult(judgeSeat, card, events)
	case guicaiResumeLuoshen:
		return g.applyLuoshenJudgeResult(judgeSeat, card, events)
	case guicaiResumeGanglie:
		return g.applyGanglieJudgeResult(judgeSeat, card, events)
	case guicaiResumeLeiji:
		return g.applyLeijiJudgeResult(judgeSeat, card, events)
	default:
		return nil
	}
}
