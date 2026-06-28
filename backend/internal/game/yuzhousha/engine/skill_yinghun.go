package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterYinghunUsed           = "yinghun_used_prepare"
	ResponseModeSkillYinghun     = "skill_yinghun"
	ResponseModeSkillYinghunDiscard = "skill_yinghun_discard"
	YinghunOptionOppDrawXDiscard1 = "opp_draw_x_discard_1"
	YinghunOptionOppDraw1DiscardX = "opp_draw_1_discard_x"
)

// yinghunLostHp 返回孙坚已损失的体力值（X）
func yinghunLostHp(g *Game, seat int) int {
	return g.Players[seat].MaxHP - g.Players[seat].HP
}

func (g *Game) ActivateYinghun(seat int, target int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillYinghun) || g.getSkillCounter(seat, counterYinghunUsed) > 0 {
		return ErrWrongPhase
	}
	// 必须已受伤才能发动
	if g.Players[seat].HP >= g.Players[seat].MaxHP {
		return ErrWrongPhase
	}
	if target < 0 {
		target = g.opponentOf(seat)
	}
	if target == seat || target < 0 || target >= len(g.Players) {
		return ErrInvalidTarget
	}
	g.setSkillCounter(seat, counterYinghunUsed, 1)

	x := yinghunLostHp(g, seat)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  seat,
		TargetIndex:  target,
		ReturnIndex:  seat,
		EffectTarget: seat,
		ResponseMode: ResponseModeSkillYinghun,
		SkillID:      skill.IDYinghun,
		Extra:        map[string]int{"yinghun_x": x},
	}
	msg := fmt.Sprintf("%s 发动【英魂】（X=%d），请 %s 选择一项", g.Players[seat].Name, x, g.Players[target].Name)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDYinghun, seat, target, msg)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	return nil
}

func (g *Game) ResolveYinghunChoice(target int, option string, events *[]GameEvent) error {
	return g.resolveYinghunChoice(target, option, "", events)
}

// yinghunX 从 Pending.Extra 中读取 X 值
func yinghunX(p *PendingCombat) int {
	if p == nil || p.Extra == nil {
		return 1
	}
	if x, ok := p.Extra["yinghun_x"]; ok && x > 0 {
		return x
	}
	return 1
}

func (g *Game) YinghunDiscard(target int, cardIDs []string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillYinghunDiscard || g.Pending.TargetIndex != target {
		return ErrWrongPhase
	}
	source := g.Pending.SourceIndex
	x := yinghunX(g.Pending)
	option := g.Pending.Extra["yinghun_option"] // 1 或 2
	discardNeed := g.Pending.Extra["yinghun_discard_need"]
	discardDone := g.Pending.Extra["yinghun_discard_done"]

	// 执行弃牌（支持一次传多张）
	for _, cardID := range cardIDs {
		idx, _, ok := g.findCard(target, cardID)
		if !ok {
			return ErrInvalidCard
		}
		discarded := g.removeHandCard(target, idx, events)
		g.DiscardPile = append(g.DiscardPile, discarded)
		g.SyncCounts()
		g.runCardsDiscardedHooks(target, "yinghun", []Card{discarded}, events)
		discardDone++
	}

	// 检查是否还需要继续弃牌
	if discardDone < discardNeed {
		// 还需要继续弃牌，更新 Pending 状态
		g.Pending.Extra["yinghun_discard_done"] = discardDone
		remaining := discardNeed - discardDone
		g.Message = fmt.Sprintf("%s 请选择 %d 张手牌弃置（【英魂】）", g.Players[target].Name, remaining)
		g.resetTimer()
		return nil
	}

	// 弃牌完成，结算
	g.Pending = nil
	return g.finishYinghunDiscard(source, target, option, x, events)
}

// finishYinghunDiscard 弃牌完成后，生成消息并结束结算
// 摸牌已在 resolveYinghunChoice 中完成，此处只需生成消息
func (g *Game) finishYinghunDiscard(source, target, option, x int, events *[]GameEvent) error {
	switch option {
	case 1:
		// 选项1：令对手摸 X 张牌，然后其弃置一张牌（弃牌已完成）
		msg := fmt.Sprintf("%s 选择令 %s 摸 %d 张牌并弃置一张手牌", g.Players[target].Name, g.Players[source].Name, x)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "skill_yinghun",
			PlayerIndex: source,
			TargetIndex: target,
			SkillID:     skill.IDYinghun,
			Message:     msg,
		})
	default:
		// 选项2：令对手摸一张牌，然后其弃置 X 张牌（弃牌已完成）
		msg := fmt.Sprintf("%s 选择令 %s 摸一张牌并弃置 %d 张手牌", g.Players[target].Name, g.Players[source].Name, x)
		g.Message = msg
		*events = append(*events, GameEvent{
			Type:        "skill_yinghun",
			PlayerIndex: source,
			TargetIndex: target,
			SkillID:     skill.IDYinghun,
			Message:     msg,
		})
	}
	return g.finishYinghun(source, events)
}

func (g *Game) resolveYinghunChoice(target int, option, discardCardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillYinghun {
		return ErrNoPendingCombat
	}
	if target != g.Pending.TargetIndex {
		return ErrNotYourTurn
	}
	source := g.Pending.SourceIndex
	returnIndex := g.Pending.ReturnIndex
	x := yinghunX(g.Pending)

	switch option {
	case YinghunOptionOppDrawXDiscard1:
		// 选项1：令对手摸 X 张牌，然后其弃置一张牌
		g.drawCards(target, x, events)
		// 如果对手没有手牌，直接结算完毕
		if len(g.Players[target].Hand) == 0 {
			msg := fmt.Sprintf("%s 选择令 %s 摸 %d 张牌", g.Players[target].Name, g.Players[source].Name, x)
			g.Message = msg
			*events = append(*events, GameEvent{
				Type:        "skill_yinghun",
				PlayerIndex: source,
				TargetIndex: target,
				SkillID:     skill.IDYinghun,
				Message:     msg,
			})
			g.Pending = nil
			return g.finishYinghun(returnIndex, events)
		}
		// 需要对手弃置一张牌
		g.Phase = PhaseResponse
		g.Pending = &PendingCombat{
			SourceIndex:  source,
			TargetIndex:  target,
			ReturnIndex:  returnIndex,
			ResponseMode: ResponseModeSkillYinghunDiscard,
			SkillID:      skill.IDYinghun,
			Extra:        map[string]int{"yinghun_x": x, "yinghun_option": 1, "yinghun_discard_need": 1, "yinghun_discard_done": 0},
		}
		g.Message = fmt.Sprintf("%s 请选择一张手牌弃置（【英魂】）", g.Players[target].Name)
		FillPendingRoles(g.Pending)
		g.resetTimer()
		return nil

	case YinghunOptionOppDraw1DiscardX:
		// 选项2：令对手摸一张牌，然后其弃置 X 张牌
		g.drawCards(target, 1, events)
		handCount := len(g.Players[target].Hand)
		// 实际能弃的牌数 = min(X, 手牌数)
		discardNeed := x
		if handCount < discardNeed {
			discardNeed = handCount
		}
		if discardNeed <= 0 {
			// 无需弃牌，结算完毕
			msg := fmt.Sprintf("%s 选择令 %s 摸一张牌", g.Players[target].Name, g.Players[source].Name)
			g.Message = msg
			*events = append(*events, GameEvent{
				Type:        "skill_yinghun",
				PlayerIndex: source,
				TargetIndex: target,
				SkillID:     skill.IDYinghun,
				Message:     msg,
			})
			g.Pending = nil
			return g.finishYinghun(returnIndex, events)
		}
		// 需要弃置多张牌，进入弃牌流程
		g.Phase = PhaseResponse
		g.Pending = &PendingCombat{
			SourceIndex:  source,
			TargetIndex:  target,
			ReturnIndex:  returnIndex,
			ResponseMode: ResponseModeSkillYinghunDiscard,
			SkillID:      skill.IDYinghun,
			Extra:        map[string]int{"yinghun_x": x, "yinghun_option": 2, "yinghun_discard_need": discardNeed, "yinghun_discard_done": 0},
		}
		g.Message = fmt.Sprintf("%s 请选择 %d 张手牌弃置（【英魂】）", g.Players[target].Name, discardNeed)
		FillPendingRoles(g.Pending)
		g.resetTimer()
		return nil

	default:
		return ErrInvalidTarget
	}
}

func (g *Game) finishYinghun(returnIndex int, events *[]GameEvent) error {
	g.Phase = PhasePlaying
	g.TurnStep = StepPrepare
	g.CurrentTurn = returnIndex
	return g.continueAfterPrepare(returnIndex, events)
}

func (g *Game) aiPickYinghunOption(target, source int) string {
	_ = source
	// AI 简单策略：对手手牌多时选选项1（让对手摸多弃少），手牌少时选选项2（让对手摸少弃多）
	x := yinghunLostHp(g, source)
	handCount := len(g.Players[target].Hand)
	// 估算：选项1 对手净收益 = X-1 张牌；选项2 对手净收益 = 1-X 张牌（即损失 X-1 张）
	// AI 希望对手受损，所以选选项2（让对手弃更多）
	// 但如果对手手牌不足以弃 X 张，则选选项1
	if handCount < x {
		return YinghunOptionOppDrawXDiscard1
	}
	return YinghunOptionOppDraw1DiscardX
}
