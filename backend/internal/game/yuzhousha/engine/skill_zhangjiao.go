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
	// 雷击判定：黑色成功（参考 noname: 雷击 → 判定黑色）
	return g.startJudge(seat, skill.JudgeLeiji, judgeFuncLeiji, guicaiResumeLeiji, events)
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
	resume := DamageResume{
		LeijiResumeShan: true,
		LeijiSaved:      saved,
		LeijiShanSeat:   seat,
	}
	if g.ApplyDamageAndCheckDeath(seat, target, 2, lightning, resume, events) {
		return nil
	}
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
	if pending.Card.Kind == CardSha {
		// HookShaMiss：杀被闪抵消（noname: shaMiss）
		// RolePlayer: 出闪者技能（如雷击）
		g.runSkillHooks(events, skill.HookCall{
			Kind: skill.HookShaMiss, Seat: seat, Role: skill.RolePlayer,
			ShaCtx: &skill.ShaCtx{Source: source, Target: seat, Card: cardView(pending.Card), Damage: pending.Damage},
		})
		// RoleSource: 出杀者装备技能（青龙刀/贯石斧等 TagEquipSkill）
		g.runSkillHooks(events, skill.HookCall{
			Kind: skill.HookShaMiss, From: source, Role: skill.RoleSource,
			ShaCtx: &skill.ShaCtx{Source: source, Target: seat, Card: cardView(pending.Card), Damage: pending.Damage},
		})
	}
	// ★ 南蛮/万箭 AOE 恢复：出闪后继续下一个目标
	if pending.Card.Kind == CardNanMan || pending.Card.Kind == CardWanJian {
		g.Pending = nil
		queue := pending.AoeQueue
		if pending.Card.Kind == CardNanMan {
			g.continueNanManAfterTarget(source, queue, events)
		} else {
			g.continueWanJianAfterTarget(source, queue, events)
		}
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
	// 方天画戟：当前目标结算完毕，处理下一个额外目标
	if g.Pending != nil && len(g.Pending.FangtianQueue) > 0 {
		g.continueFangtianAfterTarget(source, events)
		return nil
	}
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = dodgeMsg
	g.resetTimer()
	return nil
}
