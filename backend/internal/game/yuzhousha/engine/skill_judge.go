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

// startJudge 翻判定牌；然后按座位顺序询问改判技能。
func (g *Game) startJudge(judgeSeat int, reason skill.JudgeReason, resume string, events *[]GameEvent) error {
	card, ok := g.flipJudgeCard(events, judgeSeat)
	if !ok {
		return ErrInvalidCard
	}
	Logf("startJudge: seat=%d(%s) reason=%s card=%s", judgeSeat, g.Players[judgeSeat].Name, reason, card.Label)
	// 覆盖翻牌消息，带上判定类型和花色
	if len(*events) > 0 {
		last := &(*events)[len(*events)-1]
		if last.Type == "judge_flip" {
			judgeName := reasonToName(reason)
			label := suitSymbol(card.Suit) + rankLabel(card.Rank)
			last.Message = fmt.Sprintf("%s 的【%s】判定为 %s", g.Players[judgeSeat].Name, judgeName, label)
		}
	}
	return g.afterJudgeFlip(judgeSeat, reason, resume, card, events)
}

func reasonToName(reason skill.JudgeReason) string {
	switch reason {
	case skill.JudgeLebu:
		return "乐不思蜀"
	case skill.JudgeBingliang:
		return "兵粮寸断"
	case skill.JudgeShandian:
		return "闪电"
	case skill.JudgeBagua:
		return "八卦阵"
	case skill.JudgeTieqi:
		return "铁骑"
	case skill.JudgeGanglie:
		return "刚烈"
	case skill.JudgeLuoshen:
		return "洛神"
	case skill.JudgeLeiji:
		return "雷击"
	default:
		return string(reason)
	}
}

// afterJudgeFlip 翻牌后：按座位顺序收集所有可改判者，从当前回合玩家下家开始依次询问。
// 每人一次机会，发动后新牌继续问剩下的人，跳过则继续下一个。
func (g *Game) afterJudgeFlip(judgeSeat int, reason skill.JudgeReason, resume string, card Card, events *[]GameEvent) error {
	if g.offerDdzJudgeCancelWindow(judgeSeat, reason, resume, card, events) {
		return nil
	}

	// 收集所有可改判的座位（按座位顺序，从 judgeSeat 下家开始）
	candidates := g.collectModifyJudgeSeats(judgeSeat)
	if len(candidates) == 0 {
		return g.completeJudgeResume(resume, judgeSeat, reason, card, events)
	}

	return g.offerNextModifyJudge(judgeSeat, reason, resume, card, candidates, 0, events)
}

// collectModifyJudgeSeats 按座位顺序收集所有可改判者（从 startSeat 下家开始）
func (g *Game) collectModifyJudgeSeats(startSeat int) []int {
	var seats []int
	for i := 0; i < len(g.Players); i++ {
		seat := (startSeat + i + 1) % len(g.Players)
		if g.Players[seat].HP <= 0 {
			continue
		}
		canModify := false
		if g.hasSkill(seat, SkillGuicai) && len(g.Players[seat].Hand) > 0 {
			canModify = true
		}
		if g.hasSkill(seat, SkillGuidao) && g.hasBlackHandCard(seat) {
			canModify = true
		}
		if canModify {
			seats = append(seats, seat)
		}
	}
	return seats
}

// offerNextModifyJudge 询问候选人队列中的下一个人
func (g *Game) offerNextModifyJudge(judgeSeat int, reason skill.JudgeReason, resume string,
	card Card, candidates []int, idx int, events *[]GameEvent) error {

	if idx >= len(candidates) {
		// 所有人都问完了
		return g.completeJudgeResume(resume, judgeSeat, reason, card, events)
	}

	seat := candidates[idx]

	// 确定此人能用什么技能
	canGuicai := g.hasSkill(seat, SkillGuicai) && len(g.Players[seat].Hand) > 0
	canGuidao := g.hasSkill(seat, SkillGuidao) && g.hasBlackHandCard(seat)

	if !canGuicai && !canGuidao {
		// 条件不再满足，跳过此人
		return g.offerNextModifyJudge(judgeSeat, reason, resume, card, candidates, idx+1, events)
	}

	var respMode string
	var skillID string
	if canGuicai && canGuidao {
		respMode = ResponseModeSkillGuicaiGuidao
		skillID = "guicai_guidao"
	} else if canGuicai {
		respMode = ResponseModeSkillGuicai
		skillID = skill.IDGuicai
	} else {
		respMode = ResponseModeSkillGuidao
		skillID = skill.IDGuidao
	}

	var saved *PendingCombat
	if g.Pending != nil {
		copy := *g.Pending
		saved = &copy
	}

	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:     judgeSeat,
		TargetIndex:     seat,
		ResponseMode:    respMode,
		JudgeCard:       card,
		JudgeReason:     string(reason),
		GuicaiResume:    resume,
		GuicaiJudgeSeat: judgeSeat,
		SavedPending:    saved,
		// 保存改判队列状态
		ModifyCandidates: candidates,
		ModifyIndex:      idx,
	}
	judgeName := reasonToName(reason)
	label := suitSymbol(card.Suit) + rankLabel(card.Rank)
	g.Message = fmt.Sprintf("是否要对 %s 的【%s】判定为 %s 进行改判？", g.Players[judgeSeat].Name, judgeName, label)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	g.appendSkillEvent(events, skillID, seat, judgeSeat, g.Message)
	return nil
}

// resolveModifyReplace 统一的改判处理（鬼才或鬼道）
func (g *Game) resolveModifyReplace(seat int, handCardID string, skillType string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	if skillType == "guicai" && g.Pending.ResponseMode != ResponseModeSkillGuicai && g.Pending.ResponseMode != ResponseModeSkillGuicaiGuidao {
		return ErrWrongPhase
	}
	if skillType == "guidao" && g.Pending.ResponseMode != ResponseModeSkillGuidao && g.Pending.ResponseMode != ResponseModeSkillGuicaiGuidao {
		return ErrWrongPhase
	}

	idx, cardObj, ok := g.findCard(seat, handCardID)
	if !ok {
		return ErrInvalidCard
	}
	if skillType == "guidao" && !skill.IsBlackSuit(cardObj.Suit) {
		return ErrInvalidCard
	}

	oldJudge := g.Pending.JudgeCard
	played := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, oldJudge)
	g.syncCounts()

	resume := g.Pending.GuicaiResume
	judgeSeat := g.Pending.GuicaiJudgeSeat
	reason := skill.JudgeReason(g.Pending.JudgeReason)
	candidates := g.Pending.ModifyCandidates
	nextIdx := g.Pending.ModifyIndex + 1

	skillName := "鬼才"
	skillID := skill.IDGuicai
	eventType := "guicai_replace"
	if skillType == "guidao" {
		skillName = "鬼道"
		skillID = skill.IDGuidao
		eventType = "guidao_replace"
	}

	msg := fmt.Sprintf("%s 发动【%s】，以 %s 代替判定 %s", g.Players[seat].Name, skillName, played.Label, oldJudge.Label)
	g.appendSkillEvent(events, skillID, seat, judgeSeat, msg)
	*events = append(*events, GameEvent{
		Type:        eventType,
		PlayerIndex: seat,
		TargetIndex: judgeSeat,
		Card:        &played,
		Message:     msg,
	})

	// 清除 Pending，用新牌继续问剩下的人
	g.Pending = nil
	g.Phase = PhasePlaying
	return g.offerNextModifyJudge(judgeSeat, reason, resume, played, candidates, nextIdx, events)
}

func (g *Game) ApplyGuicaiReplace(seat int, handCardID string, events *[]GameEvent) error {
	return g.resolveModifyReplace(seat, handCardID, "guicai", events)
}

func (g *Game) ApplyGuidaoReplace(seat int, handCardID string, events *[]GameEvent) error {
	return g.resolveModifyReplace(seat, handCardID, "guidao", events)
}

// passModifyJudge 跳过当前改判询问，继续下一个
func (g *Game) passModifyJudge(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	card := g.Pending.JudgeCard
	resume := g.Pending.GuicaiResume
	judgeSeat := g.Pending.GuicaiJudgeSeat
	reason := skill.JudgeReason(g.Pending.JudgeReason)
	candidates := g.Pending.ModifyCandidates
	nextIdx := g.Pending.ModifyIndex + 1

	g.Pending = nil
	g.Phase = PhasePlaying
	return g.offerNextModifyJudge(judgeSeat, reason, resume, card, candidates, nextIdx, events)
}

func (g *Game) PassGuicai(seat int, events *[]GameEvent) error {
	return g.passModifyJudge(seat, events)
}

func (g *Game) PassGuidao(seat int, events *[]GameEvent) error {
	return g.passModifyJudge(seat, events)
}

func (g *Game) completeJudgeResume(resume string, judgeSeat int, reason skill.JudgeReason, card Card, events *[]GameEvent) error {
	Logf("completeJudgeResume: resume=%s judgeSeat=%d(%s) reason=%s", resume, judgeSeat, g.Players[judgeSeat].Name, reason)
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
	case "phase_judge":
		return g.applyPhaseJudgeResult(judgeSeat, reason, card, events)
	default:
		return nil
	}
}
