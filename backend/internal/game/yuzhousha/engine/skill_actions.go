package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) executeRendeGive(source, target int, cardIDs []string, events *[]GameEvent) error {
	if !g.hasSkill(source, skill.IDRende) {
		return ErrInvalidCard
	}
	if len(cardIDs) == 0 {
		return ErrInvalidCard
	}
	given := make([]Card, 0, len(cardIDs))
	for _, id := range cardIDs {
		idx, _, ok := g.findCard(source, id)
		if !ok {
			return ErrInvalidCard
		}
		given = append(given, g.removeHandCard(source, idx, events))
	}
	g.Players[target].Hand = append(g.Players[target].Hand, given...)
	g.syncCounts()

	g.addSkillCounter(source, counterRendeGiven, len(given))
	total := g.getSkillCounter(source, counterRendeGiven)
	msg := fmt.Sprintf("%s 发动【仁德】，交给 %s %d 张牌", g.Players[source].Name, g.Players[target].Name, len(given))
	g.appendSkillEvent(events, skill.IDRende, source, target, msg)
	for i := range given {
		c := given[i]
		*events = append(*events, GameEvent{
			Type:        "skill_give_card",
			PlayerIndex: source,
			TargetIndex: target,
			Card:        &c,
			SkillID:     skill.IDRende,
			Message:     fmt.Sprintf("给出 %s", c.Label),
		})
	}

	if total >= 2 && g.getSkillCounter(source, counterRendeHealed) == 0 {
		p := &g.Players[source]
		if p.HP < p.MaxHP {
			p.HP++
			g.addSkillCounter(source, counterRendeHealed, 1)
			*events = append(*events, GameEvent{
				Type:        "skill_heal",
				PlayerIndex: source,
				SkillID:     skill.IDRende,
				Heal:        1,
				Message:     fmt.Sprintf("%s 【仁德】回复 1 点体力，%d/%d", p.Name, p.HP, p.MaxHP),
			})
		}
	}

	g.Message = msg
	g.resetTimer()
	return nil
}

func (g *Game) startJijiangForUse(lord, target int, events *[]GameEvent) error {
	allies := g.shuAlliesOf(lord)
	if len(allies) == 0 {
		return ErrInvalidTarget
	}
	ally := allies[0]
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  lord,
		TargetIndex:  ally,
		ReturnIndex:  lord,
		EffectTarget: target,
		Card:         Card{Kind: CardSha, Name: "杀"},
		RequiredKind: CardSha,
		ResponseMode: ResponseModeSkillJijiang,
		SkillID:      skill.IDJijiang,
		JijiangLord:  lord,
		JijiangUse:   true,
	}
	g.Message = fmt.Sprintf("%s 发动【激将】，请 %s 出【杀】", g.Players[lord].Name, g.Players[ally].Name)
	g.appendSkillEvent(events, skill.IDJijiang, lord, ally, g.Message)
	g.resetTimer()
	return nil
}

func (g *Game) startJijiangForResponse(lord int, events *[]GameEvent) error {
	allies := g.shuAlliesOf(lord)
	if len(allies) == 0 {
		return ErrInvalidTarget
	}
	pending := *g.Pending
	ally := allies[0]
	g.Pending = &PendingCombat{
		SourceIndex:  pending.SourceIndex,
		TargetIndex:  ally,
		ReturnIndex:  pending.ReturnIndex,
		EffectTarget: pending.EffectTarget,
		Card:         pending.Card,
		RequiredKind: CardSha,
		ResponseMode: ResponseModeSkillJijiang,
		SkillID:      skill.IDJijiang,
		JijiangLord:  lord,
		JijiangUse:   false,
		Damage:       pending.Damage,
		AllowWuxiek:  pending.AllowWuxiek,
	}
	g.Message = fmt.Sprintf("%s 发动【激将】，请 %s 打出【杀】", g.Players[lord].Name, g.Players[ally].Name)
	g.appendSkillEvent(events, skill.IDJijiang, lord, ally, g.Message)
	g.resetTimer()
	return nil
}

func (g *Game) respondJijiangSha(ally int, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillJijiang {
		return ErrWrongPhase
	}
	if ally != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	lord := g.Pending.JijiangLord
	target := g.Pending.EffectTarget
	idx, cardObj, ok := g.findCard(ally, cardID)
	if !ok || !g.cardPlaysAs(ally, cardObj, CardSha) {
		return ErrInvalidCard
	}
	played := g.removeHandCard(ally, idx, events)
	// 变牌统一转为普通杀
	if !isSha(played.Kind) {
		played = g.convertCardToKind(played, CardSha)
	}
	g.DiscardPile = append(g.DiscardPile, played)

	*events = append(*events, GameEvent{
		Type:        "skill_jijiang_sha",
		PlayerIndex: ally,
		TargetIndex: target,
		Card:        &played,
		SkillID:     skill.IDJijiang,
		Message:     fmt.Sprintf("%s 响应【激将】打出 %s", g.Players[ally].Name, played.Label),
	})

	if g.Pending.JijiangUse {
		if !g.skillUnlimitedSha(lord) {
			g.Players[lord].ShaUsedThisTurn = true
		}
		return g.finishJijiangUseSha(lord, target, played, events)
	}
	return g.finishJijiangResponseSha(lord, played, events)
}

func (g *Game) finishJijiangUseSha(lord, target int, played Card, events *[]GameEvent) error {
	// 朱雀羽扇：将普通杀转为火杀
	if g.hasWeaponKind(lord, CardWeapon7) && played.DamageType == DamageTypeNormal {
		played.DamageType = DamageTypeFire
		played.Name = "火杀"
	}
	damage := g.shaBaseDamage(lord)
	ignoreArmor := g.hasWeaponKind(lord, CardWeapon2)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  lord,
		TargetIndex:  target,
		ReturnIndex:  lord,
		Card:         played,
		RequiredKind: CardShan,
		Damage:       damage,
		IgnoreArmor:  ignoreArmor,
	}
	g.Message = fmt.Sprintf("%s 的【激将】杀指向 %s，等待出闪", g.Players[lord].Name, g.Players[target].Name)
	*events = append(*events, GameEvent{
		Type:        "play_sha",
		PlayerIndex: lord,
		TargetIndex: target,
		Card:        &played,
		Message:     g.Message,
	})
	g.resetTimer()
	return nil
}

func (g *Game) finishJijiangResponseSha(lord int, _ Card, events *[]GameEvent) error {
	pending := *g.Pending
	ally := pending.TargetIndex
	msg := fmt.Sprintf("%s 通过【激将】由 %s 打出【杀】", g.Players[lord].Name, g.Players[ally].Name)
	pending.TargetIndex = lord
	return g.resolvePendingDodgeSuccess(lord, &pending, events, msg)
}

func (g *Game) passJijiang(ally int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillJijiang {
		return ErrWrongPhase
	}
	if ally != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	lord := g.Pending.JijiangLord
	jijiangUse := g.Pending.JijiangUse
	saved := *g.Pending
	g.Pending = nil

	if jijiangUse {
		g.addSkillCounter(lord, counterJijiangUseFailed, 1)
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = lord
		g.Message = fmt.Sprintf("%s 未响应【激将】，请自行使用【杀】", g.Players[lord].Name)
		g.resetTimer()
		return nil
	}

	g.Pending = &PendingCombat{
		SourceIndex:  saved.SourceIndex,
		TargetIndex:  lord,
		ReturnIndex:  saved.ReturnIndex,
		EffectTarget: saved.EffectTarget,
		Card:         saved.Card,
		RequiredKind: CardSha,
		Damage:       saved.Damage,
		AllowWuxiek:  saved.AllowWuxiek,
	}
	g.Message = fmt.Sprintf("%s 未响应【激将】，%s 需自行出【杀】", g.Players[ally].Name, g.Players[lord].Name)
	g.resetTimer()
	return nil
}

func (g *Game) toggleWusheng(seat int, events *[]GameEvent) error {
	if !g.hasSkill(seat, skill.IDWusheng) {
		return ErrInvalidCard
	}
	active := g.getSkillCounter(seat, counterWushengActive) > 0
	if active {
		g.setSkillCounter(seat, counterWushengActive, 0)
		g.Message = fmt.Sprintf("%s 取消【武圣】", g.Players[seat].Name)
	} else {
		g.setSkillCounter(seat, counterWushengActive, 1)
		g.Message = fmt.Sprintf("%s 发动【武圣】，可将红色牌当【杀】", g.Players[seat].Name)
	}
	g.appendSkillEvent(events, skill.IDWusheng, seat, -1, g.Message)
	g.resetTimer()
	return nil
}

func (g *Game) toggleQixi(seat int, events *[]GameEvent) error {
	if !g.hasSkill(seat, skill.IDQixi) {
		return ErrInvalidCard
	}
	active := g.getSkillCounter(seat, counterQixiActive) > 0
	if active {
		g.setSkillCounter(seat, counterQixiActive, 0)
		g.Message = fmt.Sprintf("%s 取消【奇袭】", g.Players[seat].Name)
	} else {
		g.setSkillCounter(seat, counterQixiActive, 1)
		g.Message = fmt.Sprintf("%s 发动【奇袭】，可将黑色牌当【过河拆桥】", g.Players[seat].Name)
	}
	g.appendSkillEvent(events, skill.IDQixi, seat, -1, g.Message)
	g.resetTimer()
	return nil
}
