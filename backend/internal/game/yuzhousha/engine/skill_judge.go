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

// startJudge 判定统一入口（参考 noname: player.judge(judgeFunc) → content.judge）。
// 完整流程：
//   step 0: 取牌 → 亮出 → trigger("judge") 改判介入
//   step 1: 构建 JudgeResult → judge 函数计算 → mod.judge 修改 → judgeFixing → callback
func (g *Game) startJudge(judgeSeat int, reason skill.JudgeReason, judgeFunc skill.JudgeFunc, resume string, events *[]GameEvent) error {
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
	return g.afterJudgeFlip(judgeSeat, reason, judgeFunc, resume, card, events)
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

// afterJudgeFlip 翻牌后：按座位顺序收集所有可改判者，从当前回合角色开始依次询问。
// 参考 noname: trigger {global:"judge"}，触发顺序从当前回合角色开始逆时针。
// 每人一次机会，发动后新牌继续问剩下的人，跳过则继续下一个。
func (g *Game) afterJudgeFlip(judgeSeat int, reason skill.JudgeReason, judgeFunc skill.JudgeFunc, resume string, card Card, events *[]GameEvent) error {
	if g.offerDdzJudgeCancelWindow(judgeSeat, reason, resume, card, events) {
		return nil
	}

	// 收集所有可改判的座位（按座位顺序，从当前回合角色开始，逆时针）
	// 参考 noname: event.trigger("judge")，current 为当前回合角色
	startSeat := g.CurrentTurn
	candidates := g.collectModifyJudgeSeatsFrom(startSeat)
	Logf("afterJudgeFlip: startSeat=%d(%s) candidates=%v reason=%s", startSeat, g.Players[startSeat].Name, candidates, reason)
	if len(candidates) == 0 {
		return g.completeJudgeResume(judgeSeat, reason, judgeFunc, resume, card, events)
	}

	return g.offerNextModifyJudge(judgeSeat, reason, judgeFunc, resume, card, candidates, 0, events)
}

// collectModifySkillIDs 收集指定座位所有可用的改判技能 ID。
// 通过 Decl 注册表的 CanModifyJudge 回调查询，不再硬编码技能 ID。
func (g *Game) collectModifySkillIDs(seat int) []string {
	var ids []string
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(seat) {
		if h.CanModifyJudge != nil {
			if canModify, skillID := h.CanModifyJudge(rt, seat); canModify && skillID != "" {
				ids = append(ids, skillID)
			}
		}
	}
	return ids
}

// collectModifyJudgeSeats 按座位顺序收集所有可改判者（从 startSeat 下家开始）。
// 通过 Decl 注册表的 CanModifyJudge 回调查询，不再硬编码技能 ID。
// 已废弃：请使用 collectModifyJudgeSeatsFrom（从当前回合角色开始，符合三国杀规则）。
func (g *Game) collectModifyJudgeSeats(startSeat int) []int {
	var seats []int
	rt := g.skillRuntime(nil)
	for i := 0; i < len(g.Players); i++ {
		seat := (startSeat + i + 1) % len(g.Players)
		if g.Players[seat].HP <= 0 {
			continue
		}
		for _, h := range g.playerSkillHandlers(seat) {
			if h.CanModifyJudge != nil {
				if canModify, _ := h.CanModifyJudge(rt, seat); canModify {
					seats = append(seats, seat)
					break
				}
			}
		}
	}
	return seats
}

// collectModifyJudgeSeatsFrom 按座位顺序收集所有可改判者（从 startSeat 自身开始，逆时针）。
// 参考 noname: event.trigger("judge")，从当前回合角色开始依次询问。
// 通过 Decl 注册表的 CanModifyJudge 回调查询，不再硬编码技能 ID。
func (g *Game) collectModifyJudgeSeatsFrom(startSeat int) []int {
	var seats []int
	rt := g.skillRuntime(nil)
	for i := 0; i < len(g.Players); i++ {
		seat := (startSeat + i) % len(g.Players) // 从 startSeat 自身开始
		if g.Players[seat].HP <= 0 {
			continue
		}
		for _, h := range g.playerSkillHandlers(seat) {
			if h.CanModifyJudge != nil {
				if canModify, _ := h.CanModifyJudge(rt, seat); canModify {
					seats = append(seats, seat)
					break
				}
			}
		}
	}
	return seats
}

// offerNextModifyJudge 询问候选人队列中的下一个人
func (g *Game) offerNextModifyJudge(judgeSeat int, reason skill.JudgeReason, judgeFunc skill.JudgeFunc,
	resume string, card Card, candidates []int, idx int, events *[]GameEvent) error {

	if idx >= len(candidates) {
		// 所有人都问完了
		return g.completeJudgeResume(judgeSeat, reason, judgeFunc, resume, card, events)
	}

	seat := candidates[idx]

	// 通过 Decl 注册表的 CanModifyJudge 回调查询改判能力（替代硬编码 hasSkill）。
	// 收集该座位所有可用的改判技能 ID。
	modifySkills := g.collectModifySkillIDs(seat)

	if len(modifySkills) == 0 {
		// 条件不再满足，跳过此人
		return g.offerNextModifyJudge(judgeSeat, reason, judgeFunc, resume, card, candidates, idx+1, events)
	}

	var respMode string
	var skillID string
	if len(modifySkills) >= 2 {
		respMode = ResponseModeSkillGuicaiGuidao
		skillID = "guicai_guidao"
	} else {
		skillID = modifySkills[0]
		if skillID == skill.IDGuicai {
			respMode = ResponseModeSkillGuicai
		} else {
			respMode = ResponseModeSkillGuidao
		}
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
	g.SyncCounts()

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

	// 用新牌继续问剩下的人（不在这里清 Pending，由 offerNextModifyJudge 管理）
	g.Phase = PhasePlaying
	return g.offerNextModifyJudge(judgeSeat, reason, nil, resume, played, candidates, nextIdx, events)
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

	g.Phase = PhasePlaying
	return g.offerNextModifyJudge(judgeSeat, reason, nil, resume, card, candidates, nextIdx, events)
}

func (g *Game) PassGuicai(seat int, events *[]GameEvent) error {
	return g.passModifyJudge(seat, events)
}

func (g *Game) PassGuidao(seat int, events *[]GameEvent) error {
	return g.passModifyJudge(seat, events)
}

// buildJudgeResult 构建完整判定结果对象（参考 noname: event.result = {card, name, number, suit, color, bool, judge}）。
func (g *Game) buildJudgeResult(judgeSeat int, reason skill.JudgeReason, card Card, judgeFunc skill.JudgeFunc) *skill.JudgeResult {
	cv := cardView(card)
	result := &skill.JudgeResult{
		Card:   cv,
		Number: normalizeRank(card.Rank),
		Suit:   card.Suit,
		Color:  suitColor(card.Suit),
		Seat:   judgeSeat,
		Reason: reason,
	}

	// 执行判定函数（参考 noname: event.result.judge = event.judge(event.result)）
	if judgeFunc != nil {
		result.Judge = judgeFunc(cv)
		if result.Judge > 0 {
			result.Bool = skill.BoolPtr(true)
		} else if result.Judge < 0 {
			result.Bool = skill.BoolPtr(false)
		}
		// result.Judge == 0 → result.Bool = nil（无结果）
	}

	return result
}

// suitColor 返回花色对应的颜色。
func suitColor(suit string) string {
	switch suit {
	case "H", "D":
		return "red"
	case "S", "C":
		return "black"
	default:
		return ""
	}
}

// ============================================================================
// 判定函数（参考 noname: judge(card) → number, >0成功 <0失败 0无结果）
// ============================================================================

// judgeFuncLebu 乐不思蜀判定：红桃 → 1 (失效/成功), 其他 → -2 (生效/失败)
func judgeFuncLebu(card skill.CardView) int {
	if card.Suit == "H" {
		return 1
	}
	return -2
}

// judgeFuncBingliang 兵粮寸断判定：梅花 → 1 (失效/成功), 其他 → -2 (生效/失败)
func judgeFuncBingliang(card skill.CardView) int {
	if card.Suit == "C" {
		return 1
	}
	return -2
}

// judgeFuncShandian 闪电判定：黑桃2-9 → -5 (生效/失败), 其他 → 1 (失效/成功)
func judgeFuncShandian(card skill.CardView) int {
	if card.Suit == "S" && card.Rank >= 2 && card.Rank <= 9 {
		return -5
	}
	return 1
}

// judgeFuncBagua 八卦阵判定：红色 → 1 (成功，视为出闪), 黑色 → -1 (失败)
func judgeFuncBagua(card skill.CardView) int {
	if card.Suit == "H" || card.Suit == "D" {
		return 1
	}
	return -1
}

// judgeFuncTieqi 铁骑判定：红色 → 1 (成功), 黑色 → -1 (失败)
func judgeFuncTieqi(card skill.CardView) int {
	if card.Suit == "H" || card.Suit == "D" {
		return 1
	}
	return -1
}

// judgeFuncGanglie 刚烈判定：红桃 → -2 (失败), 其他 → 2 (成功)
func judgeFuncGanglie(card skill.CardView) int {
	if card.Suit == "H" {
		return -2
	}
	return 2
}

// judgeFuncLuoshen 洛神判定：黑色 → 1 (获得), 红色 → -1 (停止)
func judgeFuncLuoshen(card skill.CardView) int {
	if card.Suit == "S" || card.Suit == "C" {
		return 1
	}
	return -1
}

// judgeFuncLeiji 雷击判定：黑色 → 2 (成功), 红色 → -2 (失败)
func judgeFuncLeiji(card skill.CardView) int {
	if card.Suit == "S" || card.Suit == "C" {
		return 2
	}
	return -2
}

// completeJudgeResume 改判阶段结束后的统一回调。
// 构建 JudgeResult → mod.judge 修改 → judgeFixing → 回调处理。
func (g *Game) completeJudgeResume(judgeSeat int, reason skill.JudgeReason, judgeFunc skill.JudgeFunc, resume string, card Card, events *[]GameEvent) error {
	Logf("completeJudgeResume: resume=%s judgeSeat=%d(%s) reason=%s card=%s", resume, judgeSeat, g.Players[judgeSeat].Name, reason, card.Label)

	// 1. 构建完整判定结果（参考 noname: event.result = {card, name, number, suit, color, bool, judge}）
	result := g.buildJudgeResult(judgeSeat, reason, card, judgeFunc)

	// 2. mod.judge 被动修改（参考 noname: game.checkMod(player, event.result, "judge", player)）
	// 技能可在此设置 result.KeepCard = true 来保留判定牌（如"获得判定牌"类技能）
	g.runModJudgeHooks(judgeSeat, reason, result, events)

	// 3. judgeFixing 最终确认（参考 noname: event.trigger("judgeFixing")）
	g.runJudgeFixingHooks(judgeSeat, reason, result, events)

	// 4. 触发旧的 OnJudgeResult 钩子（向后兼容）
	g.runJudgeResultHooks(skill.JudgeCtx{
		Seat: judgeSeat, Reason: reason, Card: cardView(card), IsRed: isRedSuit(card.Suit),
	}, events)

	// 5. 根据 resume 分发到具体回调
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
		// 改判阶段结束 → PopPhase 触发 OnResume 回调（applyPhaseJudgeResult）
		// 参考 noname: phaseJudge step 3 → goto(1) 循环
		return g.PopPhase(events)
	default:
		return nil
	}
}
