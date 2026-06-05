package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) RespondShan(seat int, cardID string, events *[]GameEvent) error {
	return g.RespondCard(seat, cardID, events)
}

func (g *Game) RespondCard(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	if g.Pending.ResponseMode == ResponseModeDying {
		return g.playTaoForDying(seat, cardID, events)
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}

	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok {
		return ErrInvalidCard
	}
	if cardObj.Kind == CardWuxiek {
		return g.RespondWuxiek(seat, cardID, events)
	}

	pending := g.Pending
	if pending.ResponseMode == ResponseModeWuxiekTrick || pending.ResponseMode == ResponseModeWuxiekLebu ||
		pending.ResponseMode == ResponseModeWuxiekBingliang || pending.ResponseMode == ResponseModeWuxiekShandian {
		return ErrInvalidCard
	}
	if pending.ShaUnblockable {
		return ErrInvalidCard
	}

	requiredKind := pending.RequiredKind
	if requiredKind == "" {
		requiredKind = CardShan
	}
	if cardObj.Kind != requiredKind && !g.cardPlaysAs(seat, cardObj, requiredKind) {
		return ErrInvalidCard
	}

	played := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, played)
	g.runCardsDiscardedHooks(seat, "play", []Card{played}, events)
	source := pending.SourceIndex
	responseType := "respond_" + requiredKind

	*events = append(*events, GameEvent{
		Type:        responseType,
		PlayerIndex: seat,
		TargetIndex: source,
		Card:        &played,
		Message:     fmt.Sprintf("%s 打出【%s】", g.Players[seat].Name, played.Name),
	})
	if requiredKind == CardSha {
		g.markShaInPlayPhase(seat)
	}

	if g.consumeWushuangResponse(pending, seat, requiredKind) {
		return nil
	}

	if pending.Card.Kind == CardJueDou {
		g.Pending = &PendingCombat{
			SourceIndex:  seat,
			TargetIndex:  source,
			ReturnIndex:  pending.ReturnIndex,
			Card:         pending.Card,
			RequiredKind: CardSha,
			Damage:       pending.Damage,
		}
		g.Message = fmt.Sprintf("【决斗】继续，%s 需出杀", g.Players[source].Name)
		g.resetTimer()
		return nil
	}

	return g.resolvePendingDodgeSuccess(seat, pending, events, "")
}

func (g *Game) resolvePendingDodgeSuccess(seat int, pending *PendingCombat, events *[]GameEvent, messageOverride string) error {
	if pending.Card.Kind == CardNanMan || pending.Card.Kind == CardWanJian {
		required := pending.RequiredKind
		if required == "" {
			required = CardShan
		}
		queue := pending.AoeQueue
		g.Pending = nil
		if g.offerLeijiAfterShan(seat, pending, events) {
			return nil
		}
		return g.continueAoeAfterTarget(pending.SourceIndex, pending.Card, required, queue, events)
	}
	if g.offerLeijiAfterShan(seat, pending, events) {
		return nil
	}
	return g.finishShanDodgeSuccess(seat, pending, events, messageOverride)
}

func isRedSuit(suit string) bool {
	return skill.IsRedSuit(suit)
}

func (g *Game) flipJudgeCard(events *[]GameEvent, seat int) (Card, bool) {
	if len(g.DrawPile) == 0 {
		g.refillDrawPile()
	}
	if len(g.DrawPile) == 0 {
		return Card{}, false
	}
	card := g.DrawPile[0]
	g.DrawPile = g.DrawPile[1:]
	g.DiscardPile = append(g.DiscardPile, card)
	*events = append(*events, GameEvent{
		Type:        "judge_flip",
		PlayerIndex: seat,
		Card:        &card,
		Message:     fmt.Sprintf("判定牌 %s", card.Label),
	})
	return card, true
}

func (g *Game) TryBaguaJudge(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	pending := g.Pending
	if pending.ResponseMode == ResponseModeWuxiekTrick || pending.ResponseMode == ResponseModeWuxiekLebu ||
		pending.ResponseMode == ResponseModeWuxiekBingliang || pending.ResponseMode == ResponseModeWuxiekShandian {
		return ErrInvalidCard
	}
	requiredKind := pending.RequiredKind
	if requiredKind == "" {
		requiredKind = CardShan
	}
	if requiredKind != CardShan {
		return ErrInvalidCard
	}
	if g.Players[seat].Armor == nil || !g.hasBaguaArmor(seat) {
		return ErrInvalidCard
	}
	if pending.IgnoreArmor {
		return ErrInvalidCard
	}
	if pending.BaguaUsed {
		return ErrAlreadyActed
	}

	return g.startJudge(seat, skill.JudgeBagua, guicaiResumeBagua, events)
}

func (g *Game) RespondWuxiek(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}

	pending := *g.Pending
	switch pending.ResponseMode {
	case ResponseModeWuxiekTrick, ResponseModeWuxiekLebu, ResponseModeWuxiekBingliang, ResponseModeWuxiekShandian:
		// allowed
	case "":
		if !pending.AllowWuxiek {
			return ErrInvalidCard
		}
	default:
		return ErrInvalidCard
	}

	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || cardObj.Kind != CardWuxiek {
		return ErrInvalidCard
	}

	played := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, played)
	*events = append(*events, GameEvent{
		Type:        "play_wuxiek",
		PlayerIndex: seat,
		TargetIndex: pending.SourceIndex,
		Card:        &played,
		Message:     fmt.Sprintf("%s 打出【无懈可击】", g.Players[seat].Name),
	})

	if pending.AllowWuxiek && pending.ResponseMode == "" {
		return g.cancelAoeSelfWithWuxiek(pending, events)
	}
	return g.cancelTrickWithWuxiek(pending, events)
}

func (g *Game) PassResponse(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil {
		return ErrNoPendingCombat
	}
	g.ensurePendingRoles()
	if g.Pending.WindowKind == WindowKindTake && g.takeWindow != nil {
		if g.IsActorSeat(seat) {
			return g.PassTake(seat, events)
		}
		g.abandonTakeWindow()
	}
	if g.Pending.WindowKind == WindowKindDiscard && g.discardWindow != nil {
		if g.IsActorSeat(seat) {
			return g.PassDiscardWindow(seat, events)
		}
	}
	if g.Pending.TieqiPending && seat == g.Pending.SourceIndex {
		return g.SkipTieqi(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillGuicai && seat == g.Pending.TargetIndex {
		return g.PassGuicai(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillGuidao && seat == g.Pending.TargetIndex {
		return g.PassGuidao(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillLeijiOffer && seat == g.Pending.TargetIndex {
		return g.PassLeijiOffer(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillFankui && seat == g.Pending.TargetIndex {
		return g.PassFankui(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillJianxiong && seat == g.Pending.TargetIndex {
		return g.PassJianxiong(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillYijiOffer && seat == g.Pending.TargetIndex {
		return g.PassYijiOffer(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillYijiGive && seat == g.Pending.TargetIndex {
		return g.PassYijiGive(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillGanglieOffer && seat == g.Pending.TargetIndex {
		return g.PassGanglieOffer(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillTuxi && seat == g.Pending.TargetIndex {
		return g.PassTuxi(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillFanjianSuit && seat == g.Pending.TargetIndex {
		return g.ResolveFanjianSuit(seat, g.aiPickFanjianSuit(), events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillTianxiang && seat == g.Pending.TargetIndex {
		return g.PassTianxiang(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillYinghun && seat == g.Pending.TargetIndex {
		return g.ResolveYinghunChoice(seat, g.aiPickYinghunOption(seat, g.Pending.SourceIndex), events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillYinghunDiscard && seat == g.Pending.TargetIndex {
		if len(g.Players[seat].Hand) == 0 {
			return ErrInvalidCard
		}
		return g.YinghunDiscard(seat, g.Players[seat].Hand[0].ID, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillLiuli && seat == g.Pending.TargetIndex {
		return g.PassLiuli(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeDdzJudgeCancel && seat == g.Pending.TargetIndex {
		return g.PassDdzJudgeCancel(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeSkillLuanwu {
		return g.passLuanwu(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeHuoGong && seat == g.Pending.TargetIndex {
		return g.resolveHuoGongFail(seat, events)
	}
	if g.Pending.ResponseMode == ResponseModeDying {
		return g.passDying(seat, events)
	}
	if seat != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	switch g.Pending.ResponseMode {
	case ResponseModeWuxiekTrick:
		return g.continueTrickAfterWuxiekPass(events)
	case ResponseModeWuxiekLebu:
		return g.applyLebuSkip(seat, events)
	case ResponseModeWuxiekBingliang:
		if g.offerDdzTrickCancelWindow(seat, ddzResumeBingliang, events) {
			return nil
		}
		g.Pending = nil
		g.Phase = PhasePlaying
		g.applyBingliangSkipDraw(seat, events)
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
	case ResponseModeWuxiekShandian:
		g.Pending = nil
		return g.resolveShandianJudge(seat, events)
	case ResponseModeGuanYuFollow:
		return g.finishGuanYuFollowUp(seat, events)
	case ResponseModeQilinBow:
		return g.finishQilinBow(seat, events)
	case ResponseModeSkillJijiang:
		return g.passJijiang(seat, events)
	case ResponseModePeekDeck, ResponseModeWuguPick:
		return ErrWrongPhase
	default:
		return g.resolvePendingMiss(events)
	}
}

func (g *Game) resolvePendingMiss(events *[]GameEvent) error {
	pending := *g.Pending
	if len(pending.AoeQueue) >= 0 && (pending.Card.Kind == CardNanMan || pending.Card.Kind == CardWanJian) {
		required := pending.RequiredKind
		if required == "" {
			required = CardShan
		}
		g.Pending = nil
		damage := pending.Damage
		if damage <= 0 {
			damage = 1
		}
		g.applyDamage(pending.SourceIndex, pending.TargetIndex, damage, pending.Card, events)
		*events = append(*events, GameEvent{
			Type:        "trick_hit",
			PlayerIndex: pending.SourceIndex,
			TargetIndex: pending.TargetIndex,
			Damage:      damage,
			Message:     g.damageMessage(&g.Players[pending.TargetIndex], pending.Card.Name, damage),
		})
		if g.Players[pending.TargetIndex].HP <= 0 {
			if g.afterDamageApplied(pending.SourceIndex, pending.TargetIndex, damage, pending.Card, DamageResume{}, events) {
				return nil
			}
		}
		return g.continueAoeAfterTarget(pending.SourceIndex, pending.Card, required, pending.AoeQueue, events)
	}
	g.Pending = nil
	damage := pending.Damage
	if damage <= 0 {
		damage = 1
	}
	resume := DamageResume{
		Mode:        damageResumeShaHit,
		Card:        pending.Card,
		ReturnIndex: pending.ReturnIndex,
		OfferQilin:  pending.Card.Kind == CardSha,
		IgnoreArmor: pending.IgnoreArmor,
	}
	if pending.LuanwuSha {
		resume.Mode = ""
		resume.OfferQilin = false
		resume.ResumeLuanwu = true
		resume.LuanwuOwner = pending.LuanwuOwner
	}
	return g.finalizeDamageHit(pending.SourceIndex, pending.TargetIndex, damage, pending.Card, resume, events)
}

func (g *Game) damageMessage(target *Player, sourceName string, damage int) string {
	return fmt.Sprintf("%s 受到【%s】%d 点伤害，体力 %d/%d", target.Name, sourceName, damage, target.HP, target.MaxHP)
}
