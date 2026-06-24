package engine

import (
	"fmt"
)

// ============================================================================
// 方天画戟（Fangtian）：最后一张手牌出杀时，可额外指定最多2个目标
// 参考 noname: fangtian_skill, mod.selectTarget: range[1] += 2
//
// 规则：
//   1. 手牌只剩这张杀时触发
//   2. 额外目标必须在攻击范围内
//   3. 额外目标不能是原目标自己
//   4. 每个目标独立结算（各自出闪）
//   5. 不计额外出杀次数
// ============================================================================

// canFangtianMultiTarget 检查是否可用方天画戟多目标出杀。
func (g *Game) canFangtianMultiTarget(seat int) bool {
	if !g.hasWeaponKind(seat, CardWeapon4) {
		return false
	}
	// 必须只剩最后一张手牌（出杀后手牌为0）
	return len(g.Players[seat].Hand) == 0
}

// fangtianExtraTargets 返回方天画戟可选的额外目标列表（在攻击范围内且不是原目标）。
func (g *Game) fangtianExtraTargets(seat int, primaryTarget int) []int {
	var targets []int
	attackRange := weaponRange(g.Players[seat].Weapon.Kind)
	for i := 0; i < len(g.Players); i++ {
		if i == seat || i == primaryTarget {
			continue
		}
		if g.Players[i].HP <= 0 {
			continue
		}
		if g.targetBlockedBySkill(i, CardSha) {
			continue
		}
		dist := g.distanceBetween(seat, i)
		if dist <= attackRange {
			targets = append(targets, i)
		}
	}
	return targets
}

// startFangtianMultiSha 方天画戟多目标杀：对每个额外目标依次进入杀响应。
// primaryTarget 的杀已经创建了 Pending，这里处理额外目标。
func (g *Game) startFangtianMultiSha(seat int, extraTargets []int, shaCard Card, events *[]GameEvent) {
	if len(extraTargets) == 0 {
		return
	}
	// 保存额外目标队列到 Pending
	g.Pending.FangtianQueue = extraTargets
	g.Pending.FangtianIndex = 0

	msg := fmt.Sprintf("【方天画戟】%s 对 %s 使用【杀】（额外目标 %d 人）",
		g.Players[seat].Name, g.Players[extraTargets[0]].Name, len(extraTargets))
	g.appendWeaponSkillEvent(events, seat, extraTargets[0], msg)
}

// continueFangtianAfterTarget 当前目标结算完毕后，处理下一个额外目标。
func (g *Game) continueFangtianAfterTarget(seat int, events *[]GameEvent) {
	if g.Pending == nil || len(g.Pending.FangtianQueue) == 0 {
		return
	}
	g.Pending.FangtianIndex++
	if g.Pending.FangtianIndex >= len(g.Pending.FangtianQueue) {
		// 所有额外目标处理完毕
		g.Pending = nil
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.CurrentTurn = seat
		g.Message = fmt.Sprintf("%s 继续出牌", g.Players[seat].Name)
		g.resetTimer()
		return
	}
	// 下一个目标
	nextTarget := g.Pending.FangtianQueue[g.Pending.FangtianIndex]
	shaCard := g.Pending.Card
	returnIndex := g.Pending.ReturnIndex

	msg := fmt.Sprintf("【方天画戟】%s 对 %s 使用【杀】", g.Players[seat].Name, g.Players[nextTarget].Name)
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:     seat,
		TargetIndex:     nextTarget,
		ReturnIndex:     returnIndex,
		Card:            shaCard,
		RequiredKind:    CardShan,
		Damage:          1,
		ResponsesNeeded: g.wushuangResponsesNeeded(seat, CardSha), // 无双对每个目标单独生效
		FangtianQueue:   g.Pending.FangtianQueue,
		FangtianIndex:   g.Pending.FangtianIndex,
	}
	FillPendingRoles(g.Pending)
	g.Message = msg
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "fangtian_sha",
		PlayerIndex: seat,
		TargetIndex: nextTarget,
		Card:        &shaCard,
		Message:     msg,
	})
}
