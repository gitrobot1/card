package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func cardView(c Card) skill.CardView {
	return skill.CardView{
		ID: c.ID, Kind: c.Kind, Suit: c.Suit, Label: c.Label, Name: c.Name, Rank: c.Rank,
	}
}

// runSkillHooks 统一技能 hook 入口；新增时机只扩展 skill.HookKind 与 switch 分支。
func (g *Game) runSkillHooks(events *[]GameEvent, call skill.HookCall) skill.HookResult {
	rt := g.skillRuntime(events)

	// 按角色维度收集 handlers（参考 noname: player/source/target/global）
	handlers := g.collectRoleHandlers(call)

	switch call.Kind {
	case skill.HookTargetBlocked:
		for _, h := range handlers {
			if h.BlocksTarget(rt, call.Target, call.CardKind) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookDistanceDelta:
		sum := 0
		for _, h := range handlers {
			sum += h.DistanceDelta(rt, call.From, call.To)
		}
		return skill.HookResult{Int: sum}

	case skill.HookTrickIgnoresDistance:
		for _, h := range handlers {
			if h.TrickIgnoresDistance(rt, call.Seat, call.TrickKind) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookInstantTrickUsed:
		if trickStaysInJudge(call.TrickKind) {
			return skill.HookResult{}
		}
		for _, h := range handlers {
			if err := h.OnInstantTrickUsed(rt, call.Seat, call.TrickKind); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookCardPlaysAs:
		for _, h := range handlers {
			if h.CardPlaysAs(rt, call.Seat, call.CardKind, call.AsKind, call.Suit) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookUnlimitedSha:
		for _, h := range handlers {
			if h.UnlimitedSha(rt, call.Seat) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookDamageCalculated:
		if call.DamageCalculated == nil {
			return skill.HookResult{}
		}
		ctx := *call.DamageCalculated
		for _, h := range handlers {
			modified, err := h.OnDamageCalculated(rt, ctx)
			if err != nil {
				return skill.HookResult{Err: err}
			}
			if modified != ctx.Amount {
				ctx.Amount = modified // 让后续 handler 看到修改后的伤害值
			}
		}
		return skill.HookResult{Int: ctx.Amount}

	case skill.HookDamageDealt:
		if call.Damage == nil {
			return skill.HookResult{}
		}
		ctx := *call.Damage
		for _, h := range handlers {
			if err := h.OnDamageDealt(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookBeforeHPChange:
		if call.BeforeHPChange == nil {
			return skill.HookResult{}
		}
		ctx := *call.BeforeHPChange
		for _, h := range handlers {
			cancelled, err := h.OnBeforeHPChange(rt, ctx)
			if err != nil {
				return skill.HookResult{Err: err}
			}
			if cancelled {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	case skill.HookHPLost:
		if call.HPLost == nil {
			return skill.HookResult{}
		}
		ctx := *call.HPLost
		for _, h := range handlers {
			if err := h.OnHPLost(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookHPChanged:
		if call.HPChanged == nil {
			return skill.HookResult{}
		}
		ctx := *call.HPChanged
		for _, h := range handlers {
			if err := h.OnHPChanged(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookJudgeResult:
		if call.Judge == nil {
			return skill.HookResult{}
		}
		ctx := *call.Judge
		for _, h := range handlers {
			if err := h.OnJudgeResult(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookModJudge:
		// mod.judge 被动修改：遍历所有玩家的技能，按座位顺序执行
		// 参考 noname: game.checkMod(player, event.result, "judge", player)
		if call.ModJudge == nil {
			return skill.HookResult{}
		}
		ctx := *call.ModJudge
		for _, h := range handlers {
			if err := h.OnModJudge(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookCardsDiscarded:
		if call.Discarded == nil {
			return skill.HookResult{}
		}
		ctx := *call.Discarded
		for _, h := range handlers {
			if err := h.OnCardsDiscarded(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookEquipLost:
		if call.EquipLost == nil {
			return skill.HookResult{}
		}
		ctx := *call.EquipLost
		for _, h := range handlers {
			if err := h.OnEquipLost(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookOnDeath:
		if call.Death == nil {
			return skill.HookResult{}
		}
		ctx := *call.Death
		for _, h := range handlers {
			if err := h.OnDeath(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookAfterDeath:
		if call.Death == nil {
			return skill.HookResult{}
		}
		ctx := *call.Death
		for _, h := range handlers {
			if err := h.OnAfterDeath(rt, ctx); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookBlocksWuxiek:
		// 检查是否有技能阻止无懈可击（参考 noname: playernowuxie）
		// 遍历所有存活玩家，任一玩家的技能可阻止则返回 true
		seat := call.Seat
		for _, h := range handlers {
			if h.BlocksWuxiek(rt, seat) {
				return skill.HookResult{Bool: true}
			}
		}
		return skill.HookResult{}

	// ===== 阶段钩子 =====
	case skill.HookPhaseBeforeStart:
		for _, h := range handlers {
			if h.OnPhaseBeforeStart != nil {
				if err := h.OnPhaseBeforeStart(rt, call.Seat); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookPhaseBegin:
		for _, h := range handlers {
			if h.OnPhaseBegin != nil {
				if err := h.OnPhaseBegin(rt, call.Seat); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookPhaseEnd:
		for _, h := range handlers {
			if h.OnPhaseEnd != nil {
				if err := h.OnPhaseEnd(rt, call.Seat); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	// ===== 阶段/回合钩子（通过 Decl 回调注册） =====
	case skill.HookPhaseBeforeEnd:
		for _, h := range handlers {
			if h.OnPhaseBeforeEnd != nil {
				if err := h.OnPhaseBeforeEnd(rt, call.Seat); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookPhaseBeginStart:
		for _, h := range handlers {
			if h.OnPhaseBeginStart != nil {
				if err := h.OnPhaseBeginStart(rt, call.Seat); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookPhaseChange:
		for _, h := range handlers {
			if h.OnPhaseChange != nil {
				if err := h.OnPhaseChange(rt, call.Seat); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookTurnBegin:
		for _, h := range handlers {
			if h.OnTurnBegin != nil {
				if err := h.OnTurnBegin(rt, call.Seat); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookTurnEnd:
		for _, h := range handlers {
			if err := h.OnTurnEnd(rt, call.Seat); err != nil {
				return skill.HookResult{Err: err}
			}
		}
		return skill.HookResult{}

	case skill.HookRoundStart:
		for _, h := range handlers {
			if h.OnRoundStart != nil {
				if err := h.OnRoundStart(rt, call.Seat); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	// ===== 杀流程钩子 =====
	case skill.HookShaBegin:
		if call.ShaCtx == nil {
			return skill.HookResult{}
		}
		ctx := *call.ShaCtx
		for _, h := range handlers {
			if h.OnShaBegin != nil {
				if err := h.OnShaBegin(rt, ctx); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookBecomeShaTarget:
		if call.BecomeTarget == nil {
			return skill.HookResult{}
		}
		ctx := *call.BecomeTarget
		for _, h := range handlers {
			if h.OnBecomeShaTarget != nil {
				if err := h.OnBecomeShaTarget(rt, ctx); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookShaMiss:
		if call.ShaCtx == nil {
			return skill.HookResult{}
		}
		ctx := *call.ShaCtx
		for _, h := range handlers {
			if h.OnShaMiss != nil {
				if err := h.OnShaMiss(rt, ctx); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookShaHit:
		if call.ShaCtx == nil {
			return skill.HookResult{}
		}
		ctx := *call.ShaCtx
		for _, h := range handlers {
			if h.OnShaHit != nil {
				if err := h.OnShaHit(rt, ctx); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	// ===== 伤害流程钩子 =====
	case skill.HookDamageBegin:
		if call.Damage == nil {
			return skill.HookResult{}
		}
		ctx := *call.Damage
		for _, h := range handlers {
			if h.OnDamageBegin != nil {
				if err := h.OnDamageBegin(rt, ctx); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookDamageEnd:
		if call.Damage == nil {
			return skill.HookResult{}
		}
		ctx := *call.Damage
		for _, h := range handlers {
			if h.OnDamageEnd != nil {
				if err := h.OnDamageEnd(rt, ctx); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	// ===== 锦囊钩子 =====
	case skill.HookUseCard:
		if call.UseCard == nil {
			return skill.HookResult{}
		}
		ctx := *call.UseCard
		for _, h := range handlers {
			if h.OnUseCard != nil {
				if err := h.OnUseCard(rt, ctx); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	case skill.HookUseCardToTarget:
		if call.UseCard == nil {
			return skill.HookResult{}
		}
		ctx := *call.UseCard
		for _, h := range handlers {
			if h.OnUseCardToTarget != nil {
				if err := h.OnUseCardToTarget(rt, ctx); err != nil {
					return skill.HookResult{Err: err}
				}
			}
		}
		return skill.HookResult{}

	default:
		return skill.HookResult{}
	}
}

// collectRoleHandlers 按角色维度收集技能 handlers（参考 noname: player/source/target/global）。
// 支持优先级排序（Priority 降序，FirstDo 排最前，LastDo 排最后）。
//
// 兼容旧调用：如果未显式设置 Role 且 Seat 为 0 但 Target 非 0，自动回退为 RoleTarget。
// 后续新增 HookCall 应显式设置 Role 字段。
func (g *Game) collectRoleHandlers(call skill.HookCall) []skill.Handler {
	role := call.Role
	seat := call.Seat

	// 兼容旧调用：未显式设置 Role/Seat 时，从上下文字段自动推断
	if role == 0 && call.Seat == 0 {
		if call.Target > 0 {
			role = skill.RoleTarget
			seat = call.Target
		} else if call.Damage != nil && call.Damage.Target > 0 {
			seat = call.Damage.Target
		} else if call.BeforeHPChange != nil && call.BeforeHPChange.Target > 0 {
			seat = call.BeforeHPChange.Target
		} else if call.HPLost != nil && call.HPLost.Seat > 0 {
			seat = call.HPLost.Seat
		} else if call.HPChanged != nil && call.HPChanged.Seat > 0 {
			seat = call.HPChanged.Seat
		} else if call.Judge != nil && call.Judge.Seat > 0 {
			seat = call.Judge.Seat
		} else if call.Discarded != nil && call.Discarded.Seat > 0 {
			seat = call.Discarded.Seat
		} else if call.EquipLost != nil && call.EquipLost.Seat > 0 {
			seat = call.EquipLost.Seat
		} else if call.Death != nil && call.Death.Victim > 0 {
			seat = call.Death.Victim
		}
	}

	var handlers []skill.Handler
	switch role {
	case skill.RoleSource:
		handlers = g.playerSkillHandlers(call.From)
	case skill.RoleTarget:
		handlers = g.playerSkillHandlers(call.Target)
	case skill.RoleGlobal:
		// 遍历所有存活玩家
		for i := range g.Players {
			if g.Players[i].HP > 0 {
				handlers = append(handlers, g.playerSkillHandlers(i)...)
			}
		}
	default: // RolePlayer
		// 优先用兼容推断出的 seat，否则用 call.Seat
		if seat > 0 {
			handlers = g.playerSkillHandlers(seat)
		} else {
			handlers = g.playerSkillHandlers(call.Seat)
		}
	}

	// 按优先级排序：Priority 降序，FirstDo 排最前，LastDo 排最后
	if len(handlers) > 1 {
		handlers = sortHandlersByPriority(handlers)
	}
	return handlers
}

// sortHandlersByPriority 按技能优先级排序（Priority 降序，FirstDo/LostDo 首尾优先）。
func sortHandlersByPriority(handlers []skill.Handler) []skill.Handler {
	firstDo := make([]skill.Handler, 0)
	lastDo := make([]skill.Handler, 0)
	middle := make([]skill.Handler, 0)

	for _, h := range handlers {
		decl := h.Decl
		if decl.FirstDo {
			firstDo = append(firstDo, h)
		} else if decl.LastDo {
			lastDo = append(lastDo, h)
		} else {
			middle = append(middle, h)
		}
	}

	// 中间部分按 Priority 降序
	for i := 0; i < len(middle); i++ {
		for j := i + 1; j < len(middle); j++ {
			if middle[j].Decl.Priority > middle[i].Decl.Priority {
				middle[i], middle[j] = middle[j], middle[i]
			}
		}
	}

	result := make([]skill.Handler, 0, len(handlers))
	result = append(result, firstDo...)
	result = append(result, middle...)
	result = append(result, lastDo...)
	return result
}

func (g *Game) targetBlockedBySkill(target int, cardKind string) bool {
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookTargetBlocked, Target: target, CardKind: cardKind,
	}).Bool
}

func (g *Game) skillDistanceDelta(from, to int) int {
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookDistanceDelta, From: from, To: to,
	}).Int
}

func (g *Game) trickIgnoresDistance(seat int, trickKind string) bool {
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookTrickIgnoresDistance, Seat: seat, TrickKind: trickKind,
	}).Bool
}

func (g *Game) notifyInstantTrickUsed(seat int, trickKind string, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookInstantTrickUsed, Seat: seat, TrickKind: trickKind,
	})
}

func (g *Game) cardPlaysAsViaHooks(seat int, card Card, asKind string) bool {
	if card.Kind == asKind {
		return true
	}
	// 属性杀（火杀/雷杀）视为"杀"使用
	if asKind == CardSha && isSha(card.Kind) {
		return true
	}
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookCardPlaysAs, Seat: seat,
		CardKind: card.Kind, AsKind: asKind, Suit: card.Suit,
	}).Bool
}

// triggerChongzhen 检查是否因龙胆转化而触发冲阵
func (g *Game) triggerChongzhen(seat int, card Card, asKind string) {
	// 如果牌本身就是目标类型，不需要龙胆转化
	if card.Kind == asKind {
		return
	}
	
	// 检查是否有冲阵技能（只有SP赵云有冲阵）
	if !g.hasSkill(seat, SkillChongzhen) {
		return
	}
	
	// 检查是否是将闪当杀，或将杀当闪（龙胆的效果）
	isValidLongdan := (card.Kind == CardShan && asKind == CardSha) || 
		(card.Kind == CardSha && asKind == CardShan)
	
	if !isValidLongdan {
		return
	}
	
	// 触发冲阵：获得对方一张手牌
	opponent := g.opponentOf(seat)
	if opponent < 0 || len(g.Players[opponent].Hand) == 0 {
		return
	}
	
	// 获得对方一张随机手牌
	g.takeRandomHandCard(seat, opponent)
}

// takeRandomHandCard 获得目标角色的一张随机手牌
func (g *Game) takeRandomHandCard(seat, target int) {
	if target < 0 || target >= len(g.Players) {
		return
	}
	
	p := &g.Players[target]
	if len(p.Hand) == 0 {
		return
	}
	
	// 简单实现：获得第一张牌（实际游戏中应该是让玩家选择或随机获得）
	taken := p.Hand[0]
	p.Hand = p.Hand[1:]
	g.Players[seat].Hand = append(g.Players[seat].Hand, taken)
	
	g.Message = fmt.Sprintf("%s 发动【冲阵】，获得 %s 的一张手牌", g.Players[seat].Name, p.Name)
}

// triggerChongzhenWithEvents 同 triggerChongzhen，但发出 GameEvent 供前端动画
func (g *Game) triggerChongzhenWithEvents(seat int, card Card, asKind string, events *[]GameEvent) {
	// 如果牌本身就是目标类型，不需要龙胆转化
	if card.Kind == asKind {
		return
	}
	
	if !g.hasSkill(seat, SkillChongzhen) {
		return
	}
	
	isValidLongdan := (card.Kind == CardShan && asKind == CardSha) || 
		(card.Kind == CardSha && asKind == CardShan)
	
	if !isValidLongdan {
		return
	}
	
	opponent := g.opponentOf(seat)
	if opponent < 0 || len(g.Players[opponent].Hand) == 0 {
		return
	}
	
	taken := g.Players[opponent].Hand[0]
	g.Players[opponent].Hand = g.Players[opponent].Hand[1:]
	g.Players[seat].Hand = append(g.Players[seat].Hand, taken)
	g.SyncCounts()
	
	msg := fmt.Sprintf("%s 发动【冲阵】，获得 %s 的一张手牌", g.Players[seat].Name, g.Players[opponent].Name)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "chongzhen_take",
		PlayerIndex: seat,
		TargetIndex: opponent,
		Card:        &taken,
		SkillID:     SkillChongzhen,
		Message:     msg,
	})
}

func (g *Game) skillUnlimitedShaViaHooks(seat int) bool {
	return g.runSkillHooks(nil, skill.HookCall{
		Kind: skill.HookUnlimitedSha, Seat: seat,
	}).Bool
}

func (g *Game) runDamageDealtHooks(ctx skill.DamageCtx, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookDamageDealt, Damage: &ctx,
	})
}

func (g *Game) runJudgeResultHooks(ctx skill.JudgeCtx, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookJudgeResult, Judge: &ctx,
	})
}

// runModJudgeHooks 执行 mod.judge 被动修改（参考 noname: game.checkMod(player, event.result, "judge", player)）。
// 遍历所有玩家的技能，按座位顺序依次调用 OnModJudge，技能可修改 result 的 suit/number/color/bool。
func (g *Game) runModJudgeHooks(seat int, reason skill.JudgeReason, result *skill.JudgeResult, events *[]GameEvent) {
	ctx := skill.ModJudgeCtx{Seat: seat, Reason: reason, Result: result}
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookModJudge, ModJudge: &ctx,
	})
}

// runJudgeFixingHooks 执行 judgeFixing 最终确认（参考 noname: event.trigger("judgeFixing")）。
// 在所有修改完成后触发，给技能最后一次确认/修改判定结果的机会。
func (g *Game) runJudgeFixingHooks(seat int, reason skill.JudgeReason, result *skill.JudgeResult, events *[]GameEvent) {
	// judgeFixing 通过 OnJudgeResult 钩子实现（参考 noname: judgeFixing 后执行 callback）
	// 当前标准包没有必须依赖 judgeFixing 的技能，预留接口供扩展包使用。
	_ = seat
	_ = reason
	_ = result
	_ = events
}

func (g *Game) runCardsDiscardedHooks(seat int, reason string, cards []Card, events *[]GameEvent) {
	if len(cards) == 0 {
		return
	}
	views := make([]skill.CardView, len(cards))
	for i, c := range cards {
		views[i] = cardView(c)
	}
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookCardsDiscarded,
		Seat: seat,
		Discarded: &skill.CardsDiscardedCtx{
			Seat: seat, Reason: reason, Cards: views,
		},
	})
}

func (g *Game) runCardResolvedHooks(seat int, card Card, originalKind string, events *[]GameEvent) {
	rt := g.skillRuntime(events)
	ctx := skill.CardResolvedCtx{Seat: seat, Card: cardView(card), OriginalKind: originalKind}
	for _, h := range g.playerSkillHandlers(seat) {
		if err := h.OnCardResolved(rt, ctx); err != nil {
			return
		}
	}
}

func (g *Game) runBecomeTargetHooks(target int, source int, card Card, events *[]GameEvent) {
	rt := g.skillRuntime(events)
	ctx := skill.BecomeTargetCtx{Seat: target, Source: source, Card: cardView(card)}
	for _, h := range g.playerSkillHandlers(target) {
		if err := h.OnBecomeTarget(rt, ctx); err != nil {
			return
		}
	}
}

func (g *Game) effectiveSuitViaHooks(seat int, suit string) string {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(seat) {
		suit = h.EffectiveSuit(rt, seat, suit)
	}
	return suit
}

func (g *Game) trickBlockedViaHooks(target int, card Card) bool {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(target) {
		if h.BlocksTrickTarget(rt, target, card.Kind, card.Suit) {
			return true
		}
	}
	return false
}

func (g *Game) peachBlockedViaHooks(userSeat int) bool {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(g.CurrentTurn) {
		if h.BlocksPeachUse(rt, userSeat) {
			return true
		}
	}
	return false
}

func (g *Game) damageAsHPLossViaHooks(source int) bool {
	if source < 0 || source >= len(g.Players) {
		return false
	}
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(source) {
		if h.DamageAsHPLoss(rt, source) {
			return true
		}
	}
	return false
}

func (g *Game) extraResponsesNeededViaHooks(source int, cardKind string) int {
	if source < 0 || source >= len(g.Players) {
		return 0
	}
	rt := g.skillRuntime(nil)
	extra := 0
	for _, h := range g.playerSkillHandlers(source) {
		extra += h.ExtraResponsesNeeded(rt, source, cardKind)
	}
	return extra
}

func (g *Game) skipsDiscardViaHooks(seat int) bool {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(seat) {
		if h.SkipsDiscardPhase(rt, seat) {
			return true
		}
	}
	return false
}

func (g *Game) notifyEquipLost(seat int, card Card, reason string, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookEquipLost,
		Seat: seat,
		EquipLost: &skill.EquipLostCtx{
			Seat: seat, Reason: reason, Card: cardView(card),
		},
	})
}

func isEquipZone(zone string) bool {
	return zone == EquipWeapon || zone == EquipArmor || zone == EquipPlusHorse || zone == EquipMinusHorse
}

// applyDamage 统一扣血并广播 OnDamageDealt。
func (g *Game) applyDamage(source, target, amount int, damageCard Card, events *[]GameEvent) int {
	if amount <= 0 || target < 0 || target >= len(g.Players) {
		return 0
	}
	p := &g.Players[target]
	if p.HP <= 0 {
		return 0
	}

	// 藤甲效果：普通杀对装备藤甲的角色无效（伤害被防止）
	if damageCard.Kind == CardSha && damageCard.DamageType == DamageTypeNormal && g.hasVineArmor(target) {
		*events = append(*events, GameEvent{
			Type:        "skill_trigger",
			PlayerIndex: target,
			TargetIndex: source,
			Message:     fmt.Sprintf("【藤甲】生效，%s 的【杀】对 %s 无效", g.Players[source].Name, p.Name),
		})
		return 0
	}

	// HOOK: 伤害值计算完后（圣光/天盾等可修改伤害值）
	finalAmount := amount
	if !g.isJueqingHarm(source) {
		result := g.runSkillHooks(events, skill.HookCall{
			Kind: skill.HookDamageCalculated,
			DamageCalculated: &skill.DamageCalculatedCtx{
				Source:   source,
				Target:   target,
				Amount:   amount,
				CardKind: damageCard.Kind,
				CardName: damageCard.Name,
				Card:     cardView(damageCard),
			},
		})
		// 总是用返回的伤害值（若无 handler 修改，result.Int == amount）
		finalAmount = result.Int
	}

	// HOOK: 扣血前（防止扣血等可取消）
	if !g.isJueqingHarm(source) {
		beforeResult := g.runSkillHooks(events, skill.HookCall{
			Kind: skill.HookBeforeHPChange,
			BeforeHPChange: &skill.BeforeHPChangeCtx{
				Source: source,
				Target: target,
				Amount: finalAmount,
			},
		})
		if beforeResult.Bool {
			// 扣血被防止，不执行扣血
			return 0
		}
	}

	oldHP := p.HP
	p.HP -= finalAmount
	// 不钳制 HP 到 0，保留负值以支持多桃濒死（HP=-1 需要 2 桃等）
	newHP := p.HP
	delta := newHP - oldHP

	if !g.isJueqingHarm(source) {
		ctx := skill.DamageCtx{
			Source: source, Target: target, Amount: finalAmount,
			FinalAmount: finalAmount,
			CardKind: damageCard.Kind, CardName: damageCard.Name,
		}
		if damageCard.ID != "" {
			ctx.Card = cardView(damageCard)
		}
		g.runDamageDealtHooks(ctx, events)
	}

	// 广播血量变化
	g.runHPChangedHooks(target, oldHP, newHP, delta, "damage", source, "", events)
	return finalAmount
}

// runHPLostHooks 广播血量流失事件（非伤害扣血）。
func (g *Game) runHPLostHooks(seat, amount int, reason string, source int, events *[]GameEvent) {
	if amount <= 0 {
		return
	}
	oldHP := g.Players[seat].HP
	g.Players[seat].HP -= amount
	if g.Players[seat].HP < 0 {
		g.Players[seat].HP = 0
	}
	newHP := g.Players[seat].HP
	delta := newHP - oldHP

	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookHPLost,
		HPLost: &skill.HPLostCtx{
			Seat:   seat,
			Amount: amount,
			Reason: reason,
			Source: source,
		},
	})

	// 广播血量变化
	g.runHPChangedHooks(seat, oldHP, newHP, delta, "hp_loss", source, "", events)
}

// runHPChangedHooks 广播血量变化事件。
func (g *Game) runHPChangedHooks(seat, oldHP, newHP, delta int, reason string, source int, skillID string, events *[]GameEvent) {
	_ = g.runSkillHooks(events, skill.HookCall{
		Kind: skill.HookHPChanged,
		HPChanged: &skill.HPChangedCtx{
			Seat:    seat,
			OldHP:   oldHP,
			NewHP:   newHP,
			Delta:   delta,
			Reason:  reason,
			Source:  source,
			SkillID: skillID,
		},
	})
}

// applyHeal 统一回复血量并广播 HPChanged。
func (g *Game) applyHeal(seat, amount int, reason string, source int, skillID string, events *[]GameEvent) {
	if amount <= 0 || seat < 0 || seat >= len(g.Players) {
		return
	}
	p := &g.Players[seat]
	oldHP := p.HP
	if p.HP+amount > p.MaxHP {
		p.HP = p.MaxHP
	} else {
		p.HP += amount
	}
	newHP := p.HP
	delta := newHP - oldHP
	if delta <= 0 {
		return
	}

	// 广播血量变化
	g.runHPChangedHooks(seat, oldHP, newHP, delta, "heal", source, skillID, events)
}
