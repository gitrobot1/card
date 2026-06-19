package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	ResponseModeSkillGuidao     = "skill_guidao"
	ResponseModeSkillLeijiOffer = "skill_leiji_offer"
	guicaiResumeLeiji           = "leiji"
)

func (g *Game) guidaoHolderSeat() int {
	for i := range g.Players {
		if g.hasSkill(i, SkillGuidao) && g.hasBlackHandCard(i) {
			return i
		}
	}
	return -1
}

func (g *Game) offerGuidaoWindow(judgeSeat int, reason skill.JudgeReason, resume string, card Card, events *[]GameEvent) bool {
	guidaoSeat := g.guidaoHolderSeat()
	if guidaoSeat < 0 {
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
		TargetIndex:     guidaoSeat,
		ResponseMode:    ResponseModeSkillGuidao,
		JudgeCard:       card,
		JudgeReason:     string(reason),
		GuicaiResume:    resume,
		GuicaiJudgeSeat: judgeSeat,
		SavedPending:    saved,
	}
	g.Message = fmt.Sprintf("%s 可对判定 %s 发动【鬼道】", g.Players[guidaoSeat].Name, card.Label)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDGuidao, guidaoSeat, judgeSeat, g.Message)
	return true
}

func (g *Game) ApplyGuidaoReplace(seat int, handCardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillGuidao || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	idx, cardObj, ok := g.findCard(seat, handCardID)
	if !ok || !skill.IsBlackSuit(cardObj.Suit) {
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

	msg := fmt.Sprintf("%s 发动【鬼道】，以 %s 代替判定 %s", g.Players[seat].Name, played.Label, oldJudge.Label)
	g.appendSkillEvent(events, skill.IDGuidao, seat, judgeSeat, msg)
	*events = append(*events, GameEvent{
		Type:        "guidao_replace",
		PlayerIndex: seat,
		TargetIndex: judgeSeat,
		Card:        &played,
		Message:     msg,
	})

	g.Pending = saved
	return g.completeJudgeResume(resume, judgeSeat, reason, played, events)
}

func (g *Game) PassGuidao(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillGuidao {
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

func (g *Game) canOfferLeijiJudge() bool {
	if len(g.DrawPile) == 0 {
		g.refillDrawPile()
	}
	return len(g.DrawPile) > 0
}

func (g *Game) offerLeijiAfterShan(shanSeat int, dodgePending *PendingCombat, events *[]GameEvent) bool {
	if dodgePending == nil || dodgePending.RequiredKind != CardShan || !g.hasSkill(shanSeat, SkillLeiji) {
		return false
	}
	if !g.canOfferLeijiJudge() {
		return false
	}
	saved := *dodgePending
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  shanSeat,
		TargetIndex:  shanSeat,
		ResponseMode: ResponseModeSkillLeijiOffer,
		SavedPending: &saved,
	}
	g.Message = fmt.Sprintf("%s 可发动【雷击】", g.Players[shanSeat].Name)
	g.resetTimer()
	g.appendSkillEvent(events, skill.IDLeiji, shanSeat, shanSeat, g.Message)
	return true
}

func (g *Game) StartLeijiJudge(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillLeijiOffer || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	g.leijiSavedPending = g.Pending.SavedPending
	g.leijiShanSeat = seat
	g.Pending = nil
	return g.startJudge(seat, skill.JudgeLeiji, guicaiResumeLeiji, events)
}

func (g *Game) PassLeijiOffer(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillLeijiOffer || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	saved := g.Pending.SavedPending
	g.Pending = nil
	return g.finishShanDodgeSuccess(seat, saved, events, "")
}

func (g *Game) applyLeijiJudgeResult(judgeSeat int, card Card, events *[]GameEvent) error {
	seat := g.leijiShanSeat
	saved := g.leijiSavedPending
	g.leijiSavedPending = nil
	g.leijiShanSeat = 0

	if !skill.IsBlackSuit(card.Suit) {
		return g.finishShanDodgeSuccess(seat, saved, events, "")
	}

	target := g.opponentOf(seat)
	lightning := Card{Kind: CardShanDian, Name: "雷击"}
	g.applyDamageWithHook(seat, target, 2, lightning, events)
	msg := fmt.Sprintf("%s 【雷击】判定黑色，%s 受到 2 点雷电伤害", g.Players[seat].Name, g.Players[target].Name)
	g.appendSkillEvent(events, skill.IDLeiji, seat, target, msg)
	*events = append(*events, GameEvent{
		Type:        "skill_leiji_hit",
		PlayerIndex: seat,
		TargetIndex: target,
		Damage:      2,
		SkillID:     skill.IDLeiji,
		Message:     msg,
	})

	resume := DamageResume{
		LeijiResumeShan: true,
		LeijiSaved:      saved,
		LeijiShanSeat:   seat,
	}
	if g.Players[target].HP <= 0 {
		if g.afterDamageApplied(seat, target, 2, lightning, resume, events) {
			return nil
		}
	}
	if g.continueAfterDamage(seat, target, 2, lightning, resume, events) {
		return nil
	}
	return g.finishShanDodgeSuccess(seat, saved, events, "")
}

func (g *Game) finishShanDodgeSuccess(seat int, pending *PendingCombat, events *[]GameEvent, messageOverride string) error {
	if pending == nil {
		g.Pending = nil
		g.Phase = PhasePlaying
		return nil
	}
	source := pending.SourceIndex
	dodgeMsg := messageOverride
	if dodgeMsg == "" {
		if pending.Card.Kind == CardSha {
			dodgeMsg = fmt.Sprintf("%s 打出【闪】，【杀】无效", g.Players[seat].Name)
		} else {
			dodgeMsg = fmt.Sprintf("%s 响应【%s】", g.Players[seat].Name, pending.Card.Name)
		}
	}
	if pending.Card.Kind == CardSha && g.offerGuanYuFollowUp(source, seat, events) {
		return nil
	}
	g.Pending = nil
	if pending.LuanwuSha {
		g.Message = dodgeMsg
		return g.finishLuanwu(pending.LuanwuOwner, events)
	}
	// 雌雄双股剑：杀被闪后也可能触发（标准规则是命中后，此处改为命中或被闪后均可）
	// 标准规则：只有杀命中后才触发雌雄双股剑
	// 此处按标准规则，在 finalizeDamageHit 中触发；若需被闪也触发，取消下行注释
	// g.tryOfferChixiongAfterSha(source, seat, source, events)
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = dodgeMsg
	g.resetTimer()
	return nil
}
