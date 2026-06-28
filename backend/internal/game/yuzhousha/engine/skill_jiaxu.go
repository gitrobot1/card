package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	ResponseModeSkillLuanwu = "skill_luanwu"
	counterLuanwuUsed       = "luanwu_used"
)

func trickSingleTargetsPlayer(kind string) bool {
	switch kind {
	case CardGuoHe, CardTanNang, CardJueDou, CardLeBu, CardBingLiang:
		return true
	default:
		return false
	}
}

func (g *Game) weimuBlocksTrick(target int, card Card) bool {
	return g.trickBlockedViaHooks(target, card)
}

func (g *Game) targetBlockedByTrick(target int, card Card) bool {
	if g.targetBlockedBySkill(target, card.Kind) {
		return true
	}
	if !trickSingleTargetsPlayer(card.Kind) {
		return false
	}
	return g.weimuBlocksTrick(target, card)
}

func (g *Game) isSeatInDyingRescue(seat int) bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModeDying && g.Pending.TargetIndex == seat
}

func (g *Game) wanshaBlocksPeachUse(userSeat int) bool {
	return g.peachBlockedViaHooks(userSeat)
}

func (g *Game) ActivateLuanwu(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillLuanwu) || g.getSkillCounter(seat, counterLuanwuUsed) > 0 {
		return ErrWrongPhase
	}
	opp := g.opponentOf(seat)
	g.setSkillCounter(seat, counterLuanwuUsed, 1)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  seat,
		TargetIndex:  opp,
		EffectTarget: opp,
		ReturnIndex:  seat,
		ResponseMode: ResponseModeSkillLuanwu,
		SkillID:      skill.IDLuanwu,
	}
	msg := fmt.Sprintf("%s 发动【乱武】，%s 须对除 %s 外的一名角色使用【杀】，否则受到 1 点伤害",
		g.Players[seat].Name, g.Players[opp].Name, g.Players[seat].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDLuanwu, seat, opp, msg)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "skill_luanwu",
		PlayerIndex: seat,
		TargetIndex: opp,
		SkillID:     skill.IDLuanwu,
		Message:     msg,
	})
	return nil
}

func (g *Game) playLuanwuSha(seat int, cardID string, target int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillLuanwu {
		return ErrWrongPhase
	}
	owner := g.Pending.SourceIndex
	if seat != g.Pending.TargetIndex || target != g.Pending.EffectTarget || target == owner {
		return ErrInvalidTarget
	}
	g.Pending = nil
	return g.playShaLuanwu(seat, cardID, target, owner, events)
}

func (g *Game) playShaLuanwu(seat int, cardID string, targetIndex, owner int, events *[]GameEvent) error {
	if g.runSkillHooks(nil, skill.HookCall{Kind: skill.HookTargetBlocked, Target: targetIndex, CardKind: CardSha}).Bool {
		return ErrInvalidTarget
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || !g.cardPlaysAs(seat, cardObj, CardSha) {
		return ErrInvalidCard
	}

	played := g.removeHandCard(seat, idx, events)
	// 变牌统一转为普通杀
	if !isSha(played.Kind) {
		played = g.convertCardToKind(played, CardSha)
	}

	// 朱雀羽扇：将普通杀转为火杀
	if g.hasWeaponKind(seat, CardWeapon7) && played.DamageType == DamageTypeNormal {
		played.DamageType = DamageTypeFire
		played.Name = "火杀"
	}

	g.DiscardPile = append(g.DiscardPile, played)
	g.runCardsDiscardedHooks(seat, "play", []Card{played}, events)
	damage := g.shaBaseDamage(seat)

	ignoreArmor := g.hasWeaponKind(seat, CardWeapon2)
	msg := fmt.Sprintf("%s 【乱武】对 %s 使用【杀】，等待出闪", g.Players[seat].Name, g.Players[targetIndex].Name)
	if ignoreArmor {
		msg += "（【青釭剑】无视防具）"
	}
	g.appendWushuangMessage(seat, CardSha, &msg)

	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:     seat,
		TargetIndex:     targetIndex,
		ReturnIndex:     owner,
		Card:            played,
		RequiredKind:    CardShan,
		Damage:          damage,
		IgnoreArmor:     ignoreArmor,
		ResponsesNeeded: g.wushuangResponsesNeeded(seat, CardSha),
		LuanwuSha:       true,
		LuanwuOwner:     owner,
	}
	g.Message = msg
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "play_sha",
		PlayerIndex: seat,
		TargetIndex: targetIndex,
		Card:        &played,
		Message:     msg,
		SkillID:     skill.IDLuanwu,
	})
	return nil
}

func (g *Game) passLuanwu(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillLuanwu || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	owner := g.Pending.SourceIndex
	g.Pending = nil
	resume := DamageResume{ResumeLuanwu: true, LuanwuOwner: owner}
	if g.ApplyDamageAndCheckDeath(owner, seat, 1, Card{Kind: CardJueDou, Name: "乱武"}, resume, events) {
		return nil
	}
	msg := fmt.Sprintf("%s 未出【杀】，受到【乱武】1 点伤害", g.Players[seat].Name)
	*events = append(*events, GameEvent{
		Type:        "skill_luanwu_damage",
		PlayerIndex: owner,
		TargetIndex: seat,
		Damage:      1,
		SkillID:     skill.IDLuanwu,
		Message:     msg,
	})
	// 走统一伤害技能链（卖血技等），技能链完毕后由 resumeAfterDamageNoSkill 恢复乱武
	if g.continueAfterDamage(owner, seat, 1, Card{Kind: CardJueDou, Name: "乱武"}, resume, events) {
		return nil
	}
	return g.finishLuanwu(owner, events)
}

func (g *Game) finishLuanwu(owner int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = owner
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[owner].Name)
	g.resetTimer()
	return nil
}
