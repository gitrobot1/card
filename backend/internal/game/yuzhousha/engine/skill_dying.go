package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const ResponseModeDying = "dying_rescue"

// DyingContext 濒死结算完成后需恢复的伤害链上下文。
type DyingContext struct {
	Victim int
	Killer int
	Damage int
	Card   Card
	Resume DamageResume
}

func (g *Game) startDyingWindow(victim int, ctx DyingContext, events *[]GameEvent) bool {
	if victim < 0 || victim >= len(g.Players) || g.Players[victim].HP > 0 {
		return false
	}
	if g.Pending != nil && g.Pending.ResponseMode == ResponseModeDying {
		return true
	}
	ctx.Victim = victim
	g.dyingContext = &ctx
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		ResponseMode: ResponseModeDying,
		SourceIndex:  victim,
		TargetIndex:  victim,
		EffectTarget: victim,
		ReturnIndex:  victim,
		RequiredKind: CardTao,
	}
	victimName := g.Players[victim].Name
	askName := g.Players[victim].Name
	g.Message = fmt.Sprintf("%s 进入濒死，%s 是否出【桃】", victimName, askName)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "dying_start",
		PlayerIndex: victim,
		TargetIndex: victim,
		Message:     g.Message,
	})
	return true
}

func (g *Game) afterDamageApplied(source, target, damage int, card Card, resume DamageResume, events *[]GameEvent) bool {
	if target < 0 || target >= len(g.Players) || g.Players[target].HP > 0 {
		return false
	}
	if g.isJueqingHarm(source) {
		return g.finishJueqingDeath(source, target, events)
	}
	return g.startDyingWindow(target, DyingContext{
		Killer: source,
		Damage: damage,
		Card:   card,
		Resume: resume,
	}, events)
}

func (g *Game) canPlayTaoForDying(askSeat int, card Card) bool {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeDying {
		return false
	}
	if askSeat != g.Pending.SourceIndex {
		return false
	}
	victim := g.Pending.TargetIndex
	if askSeat != victim && g.wanshaBlocksPeachUse(askSeat) {
		return false
	}
	if card.Kind == CardTao {
		return true
	}
	return g.cardPlaysAs(askSeat, card, CardTao)
}

func (g *Game) playTaoForDying(askSeat int, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeDying {
		return ErrNoPendingCombat
	}
	if askSeat != g.Pending.SourceIndex {
		return ErrNotYourTurn
	}
	idx, cardObj, ok := g.findCard(askSeat, cardID)
	if !ok || !g.canPlayTaoForDying(askSeat, cardObj) {
		return ErrInvalidCard
	}
	victim := g.Pending.TargetIndex
	played := g.removeHandCard(askSeat, idx, events)
	g.DiscardPile = append(g.DiscardPile, played)
	g.runCardsDiscardedHooks(askSeat, "play", []Card{played}, events)

	p := &g.Players[victim]
	p.HP++
	viaJiji := cardObj.Kind != CardTao && g.hasSkill(askSeat, SkillJiji)
	msg := fmt.Sprintf("%s 对 %s 使用【桃】，体力 %d/%d", g.Players[askSeat].Name, p.Name, p.HP, p.MaxHP)
	if viaJiji {
		msg = fmt.Sprintf("%s 发动【急救】，将 %s 当【桃】救 %s，体力 %d/%d",
			g.Players[askSeat].Name, played.Label, p.Name, p.HP, p.MaxHP)
	}
	g.Message = msg
	eventType := "play_tao"
	if viaJiji {
		eventType = "skill_jiji"
		g.appendSkillEvent(events, skill.IDJiji, askSeat, victim, msg)
	}
	*events = append(*events, GameEvent{
		Type:        eventType,
		PlayerIndex: askSeat,
		TargetIndex: victim,
		Card:        &played,
		Heal:        1,
		SkillID:     skillIDIf(viaJiji, skill.IDJiji),
		Message:     msg,
	})
	*events = append(*events, GameEvent{
		Type:        "dying_saved",
		PlayerIndex: victim,
		TargetIndex: askSeat,
		Heal:        1,
		Message:     fmt.Sprintf("%s 脱离濒死", p.Name),
	})
	return g.resolveDyingSaved(events)
}

func skillIDIf(cond bool, id string) string {
	if cond {
		return id
	}
	return ""
}

func (g *Game) passDying(askSeat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeDying {
		return ErrNoPendingCombat
	}
	if askSeat != g.Pending.SourceIndex {
		return ErrNotYourTurn
	}
	victim := g.Pending.TargetIndex
	*events = append(*events, GameEvent{
		Type:        "dying_pass",
		PlayerIndex: askSeat,
		TargetIndex: victim,
		Message:     fmt.Sprintf("%s 不出【桃】", g.Players[askSeat].Name),
	})
	next, roundDone := g.nextDyingAskSeat(askSeat, victim)
	if !roundDone {
		g.Pending.SourceIndex = next
		g.Message = fmt.Sprintf("%s 濒死，%s 是否出【桃】", g.Players[victim].Name, g.Players[next].Name)
		g.resetTimer()
		return nil
	}
	if g.Players[victim].HP > 0 {
		return g.resolveDyingSaved(events)
	}
	return g.resolveDyingDeath(events)
}

func (g *Game) nextDyingAskSeat(current, victim int) (next int, roundDone bool) {
	n := len(g.Players)
	if n <= 2 {
		if current == victim {
			return g.opponentOf1v1(victim), false
		}
		return -1, true
	}
	cand := (current + 1) % n
	for steps := 0; steps < n; steps++ {
		if cand == victim {
			return -1, true
		}
		if g.Players[cand].HP > 0 {
			return cand, false
		}
		cand = (cand + 1) % n
	}
	return -1, true
}

func (g *Game) resolveDyingSaved(events *[]GameEvent) error {
	ctx := g.dyingContext
	g.dyingContext = nil
	g.Pending = nil
	if ctx == nil {
		g.Phase = PhasePlaying
		return nil
	}
	victim := ctx.Victim
	source := ctx.Killer
	if g.continueAfterDamage(source, victim, ctx.Damage, ctx.Card, ctx.Resume, events) {
		return nil
	}
	if ctx.Resume.Mode == damageResumeFanjian {
		g.resumeAfterFanjianDamage(ctx.Resume, events)
		return nil
	}
	if ctx.Resume.ResumeLuanwu {
		return g.finishLuanwu(ctx.Resume.LuanwuOwner, events)
	}
	if ctx.Resume.LeijiResumeShan {
		return g.finishShanDodgeSuccess(ctx.Resume.LeijiShanSeat, ctx.Resume.LeijiSaved, events, "")
	}
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = ctx.Resume.ReturnIndex
	if ctx.Resume.ReturnIndex >= 0 && ctx.Resume.ReturnIndex < len(g.Players) {
		g.Message = fmt.Sprintf("%s 继续出牌", g.Players[ctx.Resume.ReturnIndex].Name)
	}
	g.resetTimer()
	return nil
}

// scatterPlayerCardsOnDeath 阵亡后将手牌、装备、判定区与「营」中牌置入弃牌堆。
func (g *Game) scatterPlayerCardsOnDeath(seat int, events *[]GameEvent) {
	if seat < 0 || seat >= len(g.Players) {
		return
	}
	p := &g.Players[seat]
	var toDiscard []Card
	if len(p.Hand) > 0 {
		toDiscard = append(toDiscard, p.Hand...)
		p.Hand = nil
	}
	for _, slot := range []*Card{p.Weapon, p.Armor, p.PlusHorse, p.MinusHorse} {
		if slot != nil {
			toDiscard = append(toDiscard, *slot)
		}
	}
	p.Weapon, p.Armor, p.PlusHorse, p.MinusHorse = nil, nil, nil, nil
	if len(p.JudgeArea) > 0 {
		toDiscard = append(toDiscard, p.JudgeArea...)
		p.JudgeArea = nil
	}
	if len(p.CampCards) > 0 {
		toDiscard = append(toDiscard, p.CampCards...)
		p.CampCards = nil
	}
	if len(toDiscard) == 0 {
		return
	}
	g.DiscardPile = append(g.DiscardPile, toDiscard...)
	g.syncCounts()
	*events = append(*events, GameEvent{
		Type:        "death_scatter",
		PlayerIndex: seat,
		Message:     fmt.Sprintf("%s 阵亡，弃置所有牌", p.Name),
		Amount:      len(toDiscard),
	})
}

func (g *Game) resolveDyingDeath(events *[]GameEvent) error {
	ctx := g.dyingContext
	victim := 0
	killer := 0
	if ctx != nil {
		victim = ctx.Victim
		killer = ctx.Killer
	}
	g.dyingContext = nil
	g.Pending = nil
	if victim >= 0 && victim < len(g.Players) {
		// HOOK: 阵亡时（亡语，牌还在）
		if !g.isJueqingHarm(killer) {
			g.runSkillHooks(events, skill.HookCall{
				Kind: skill.HookOnDeath,
				Death: &skill.DeathCtx{
					Victim: victim,
					Killer: killer,
					Reason: "damage",
				},
			})
		}

		g.scatterPlayerCardsOnDeath(victim, events)

		// HOOK: 阵亡后（牌已弃）
		if !g.isJueqingHarm(killer) {
			g.runSkillHooks(events, skill.HookCall{
				Kind: skill.HookAfterDeath,
				Death: &skill.DeathCtx{
					Victim: victim,
					Killer: killer,
					Reason: "damage",
				},
			})
		}

		*events = append(*events, GameEvent{
			Type:        "dying_death",
			PlayerIndex: victim,
			TargetIndex: killer,
			Message:     fmt.Sprintf("%s 濒死无人救，阵亡", g.Players[victim].Name),
		})
	}
	if g.checkTeamElimination(events) {
		return nil
	}
	if g.checkChainDeath(victim, events) {
		return nil
	}
	if g.is3v3() {
		if killer >= 0 && killer < len(g.Players) && g.AliveHP(killer) > 0 {
			g.drawCards(killer, 3, events)
			*events = append(*events, GameEvent{
				Type:        "kill_draw",
				PlayerIndex: killer,
				TargetIndex: victim,
				Amount:      3,
				Message:     fmt.Sprintf("%s 杀死 %s，摸 3 张牌", g.Players[killer].Name, g.Players[victim].Name),
			})
		}
		if g.checkCommanderDeath(victim, events) {
			return nil
		}
		g.Phase = PhasePlaying
		if g.AliveHP(g.CurrentTurn) <= 0 {
			g.CurrentTurn = g.nextTurnSeat(g.CurrentTurn)
			g.beginTurn(events)
		}
		g.Message = fmt.Sprintf("%s 阵亡，对局继续", g.Players[victim].Name)
		g.resetTimer()
		return nil
	}
	if g.isIdentity() {
		if g.checkIdentityDeath(victim, killer, events) {
			return nil
		}
		g.Phase = PhasePlaying
		if g.AliveHP(g.CurrentTurn) <= 0 {
			g.CurrentTurn = g.nextTurnSeat(g.CurrentTurn)
			g.beginTurn(events)
		}
		g.Message = fmt.Sprintf("%s 阵亡，对局继续", g.Players[victim].Name)
		g.resetTimer()
		return nil
	}
	if g.is2v2() {
		g.Phase = PhasePlaying
		if g.AliveHP(g.CurrentTurn) <= 0 {
			g.CurrentTurn = g.nextTurnSeat(g.CurrentTurn)
			g.beginTurn(events)
		}
		g.Message = fmt.Sprintf("%s 阵亡，对局继续", g.Players[victim].Name)
		g.resetTimer()
		return nil
	}
	g.finishGame(killer, events)
	return nil
}
