package engine

import "fmt"

func (g *Game) weaponKind(seat int) string {
	if seat < 0 || seat >= len(g.Players) || g.Players[seat].Weapon == nil {
		return ""
	}
	return g.Players[seat].Weapon.Kind
}

func (g *Game) hasWeaponKind(seat int, kind string) bool {
	return g.weaponKind(seat) == kind
}

func (g *Game) canUseSha(seat int) bool {
	if g.skillUnlimitedSha(seat) {
		return true
	}
	p := &g.Players[seat]
	if !p.ShaUsedThisTurn {
		return true
	}
	if g.is3pDdz() && seat == g.LandlordSeat && !p.ShaExtraUsedThisTurn {
		return true
	}
	return false
}

func (g *Game) countShaInHand(seat int) int {
	n := 0
	for _, c := range g.Players[seat].Hand {
		if isSha(c.Kind) {
			n++
		}
	}
	return n
}

func (g *Game) targetHasHorse(seat int) bool {
	if seat < 0 || seat >= len(g.Players) {
		return false
	}
	p := &g.Players[seat]
	return p.PlusHorse != nil || p.MinusHorse != nil
}

func (g *Game) isGuanYuFollowPending() bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModeGuanYuFollow
}

func (g *Game) isQilinBowPending() bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModeQilinBow
}

func (g *Game) appendWeaponSkillEvent(events *[]GameEvent, source, target int, message string) {
	*events = append(*events, GameEvent{
		Type:        "weapon_skill",
		PlayerIndex: source,
		TargetIndex: target,
		Message:     message,
	})
}

func (g *Game) offerGuanYuFollowUp(source, target int, events *[]GameEvent) bool {
	if !g.hasWeaponKind(source, CardWeapon3) || g.countShaInHand(source) == 0 {
		return false
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  source,
		ReturnIndex:  source,
		EffectTarget: target,
		ResponseMode: ResponseModeGuanYuFollow,
		Card:         Card{Kind: CardSha, Name: "杀"},
	}
	g.Message = fmt.Sprintf("%s 可发动【青龙偃月刀】对 %s 再使用一张【杀】", g.Players[source].Name, g.Players[target].Name)
	g.appendWeaponSkillEvent(events, source, target, g.Message)
	FillPendingRoles(g.Pending)
	g.resetTimer()
	return true
}

func (g *Game) finishGuanYuFollowUp(seat int, events *[]GameEvent) error {
	if !g.isGuanYuFollowPending() || seat != g.Pending.TargetIndex {
		return ErrWrongPhase
	}
	returnIndex := g.Pending.ReturnIndex
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = returnIndex
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[returnIndex].Name)
	g.resetTimer()
	return nil
}

func (g *Game) offerQilinBow(source, target int, returnIndex int, events *[]GameEvent) bool {
	if !g.hasWeaponKind(source, CardWeapon5) || !g.targetHasHorse(target) {
		return false
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  source,
		ReturnIndex:  returnIndex,
		EffectTarget: target,
		ResponseMode: ResponseModeQilinBow,
		Card:         Card{Kind: CardSha, Name: "杀"},
	}
	g.Message = fmt.Sprintf("%s 可发动【麒麟弓】弃置 %s 的坐骑", g.Players[source].Name, g.Players[target].Name)
	g.appendWeaponSkillEvent(events, source, target, g.Message)
	g.resetTimer()
	return true
}

func (g *Game) qilinDiscardHorse(seat int, zone string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isQilinBowPending() || seat != g.Pending.TargetIndex {
		return ErrWrongPhase
	}
	target := g.Pending.EffectTarget
	card, label, ok := g.takeTargetCard(target, PlayTarget{Zone: zone}, events)
	if !ok {
		return ErrInvalidTarget
	}
	g.DiscardPile = append(g.DiscardPile, card)
	returnIndex := g.Pending.ReturnIndex
	*events = append(*events, GameEvent{
		Type:        "qilin_discard",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &card,
		Message:     fmt.Sprintf("%s 发动【麒麟弓】弃置 %s 的%s", g.Players[seat].Name, g.Players[target].Name, label),
	})
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = returnIndex
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[returnIndex].Name)
	g.resetTimer()
	return nil
}

func (g *Game) finishQilinBow(seat int, events *[]GameEvent) error {
	if !g.isQilinBowPending() || seat != g.Pending.TargetIndex {
		return ErrWrongPhase
	}
	returnIndex := g.Pending.ReturnIndex
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = returnIndex
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[returnIndex].Name)
	g.resetTimer()
	return nil
}

// isOppositeGender 判断两名角色是否异性（至少一方向有 gender 字段且不同）。
func isOppositeGender(a, b Character) bool {
	if a.Gender == "" || b.Gender == "" {
		return false
	}
	return a.Gender != b.Gender
}

// tryOfferChixiongOnSha 在出杀指定目标后尝试发动雌雄双股剑。
// 由 advanceShaBeforeTargetResponse 调用。
// 若触发，保存当前杀 Pending 到 SavedPending，并将 Pending 改为 Weapon8 响应窗口。
func (g *Game) tryOfferChixiongOnSha(events *[]GameEvent) bool {
	p := g.Pending
	if p == nil || p.Card.Kind != CardSha {
		return false
	}
	source := p.SourceIndex
	target := p.TargetIndex
	if !g.hasWeaponKind(source, CardWeapon8) {
		return false
	}
	if !isOppositeGender(g.Players[source].Character, g.Players[target].Character) {
		return false
	}
	// 保存当前杀流程的 Pending 到 SavedPending
	saved := *p
	saved.SavedPending = p.SavedPending // 保留原有的 SavedPending 链
	p.SavedPending = &saved

	if len(g.Players[target].Hand) == 0 {
		// 目标无手牌，直接让 source 摸一张，然后恢复杀流程
		g.drawCards(source, 1, events)
		*events = append(*events, GameEvent{
			Type:        "weapon_skill",
			PlayerIndex: source,
			TargetIndex: target,
			Message:     fmt.Sprintf("【雌雄双股剑】%s 无手牌，%s 摸一张牌", g.Players[target].Name, g.Players[source].Name),
		})
		g.resumeShaFromChixiong(events)
		return true
	}
	// 进入窗口：target 选择弃一张手牌，或跳过让 source 摸牌
	p.ResponseMode = ResponseModeWeapon8
	p.SourceIndex = source
	p.TargetIndex = target
	p.ReturnIndex = source
	g.Phase = PhaseResponse
	g.Message = fmt.Sprintf("【雌雄双股剑】%s 需弃一张手牌，或跳过让 %s 摸一张牌",
		g.Players[target].Name, g.Players[source].Name)
	g.appendWeaponSkillEvent(events, source, target, g.Message)
	g.resetTimer()
	return true
}

// resolveChixiongDiscard 目标选择弃牌。
func (g *Game) resolveChixiongDiscard(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isWeapon8Pending() || seat != g.Pending.TargetIndex {
		return ErrWrongPhase
	}
	idx, _, ok := g.findCard(seat, cardID)
	if !ok {
		return ErrInvalidCard
	}
	source := g.Pending.SourceIndex
	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	*events = append(*events, GameEvent{
		Type:        "weapon_skill",
		PlayerIndex: source,
		TargetIndex: seat,
		Card:        &discarded,
		Message:     fmt.Sprintf("【雌雄双股剑】%s 弃置%s", g.Players[seat].Name, discarded.Label),
	})
	g.resumeShaFromChixiong(events)
	return nil
}

// passChixiong 目标选择不弃牌，让 source 摸一张。
func (g *Game) passChixiong(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isWeapon8Pending() || seat != g.Pending.TargetIndex {
		return ErrWrongPhase
	}
	source := g.Pending.SourceIndex
	g.drawCards(source, 1, events)
	*events = append(*events, GameEvent{
		Type:        "weapon_skill",
		PlayerIndex: source,
		TargetIndex: seat,
		Message:     fmt.Sprintf("【雌雄双股剑】%s 摸一张牌", g.Players[source].Name),
	})
	g.resumeShaFromChixiong(events)
	return nil
}

func (g *Game) isWeapon8Pending() bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModeWeapon8
}

// resumeShaFromChixiong 雌雄双股剑处理完毕后，从 SavedPending 恢复杀流程。
func (g *Game) resumeShaFromChixiong(events *[]GameEvent) {
	saved := g.Pending.SavedPending
	if saved == nil {
		g.Pending = nil
		g.Phase = PhasePlaying
		g.TurnStep = StepPlay
		g.Message = fmt.Sprintf("%s 继续出牌", g.Players[g.CurrentTurn].Name)
		g.resetTimer()
		return
	}
	g.Pending = saved
	g.Phase = PhaseResponse
	if g.canOfferLiuli(g.Pending.TargetIndex) {
		g.offerLiuliWindow(g.Pending.TargetIndex, events)
	} else {
		_ = g.advanceShaBeforeTargetResponse(events)
	}
}

// offerGuanshifu 尝试发动贯石斧：source 的杀被 target 用闪抵消后，
// 若 source 装备贯石斧且手牌≥2，可弃两张手牌令此杀依然命中。
func (g *Game) offerGuanshifu(source, target int, pendingCard Card, damage int, returnIndex int, events *[]GameEvent) bool {
	if !g.hasWeaponKind(source, CardWeapon9) {
		return false
	}
	if len(g.Players[source].Hand) < 2 {
		return false
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  source,
		TargetIndex:  source,
		ReturnIndex:  returnIndex,
		EffectTarget: target,
		ResponseMode: ResponseModeWeapon9,
		Card:         pendingCard,
		Damage:       damage,
	}
	g.Message = fmt.Sprintf("【贯石斧】%s 可弃两张手牌，令此杀依然命中 %s", g.Players[source].Name, g.Players[target].Name)
	g.appendWeaponSkillEvent(events, source, target, g.Message)
	g.resetTimer()
	return true
}

// resolveGuanshifuDiscard 发动者确认弃两张牌，令杀命中。
func (g *Game) resolveGuanshifuDiscard(seat int, cardIDs []string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isWeapon9Pending() || seat != g.Pending.SourceIndex {
		return ErrWrongPhase
	}
	if len(cardIDs) != 2 {
		return ErrInvalidCard
	}
	// 验证两张牌都在手牌中
	idxMap := make(map[string]int)
	for i, c := range g.Players[seat].Hand {
		idxMap[c.ID] = i
	}
	for _, id := range cardIDs {
		if _, ok := idxMap[id]; !ok {
			return ErrInvalidCard
		}
	}
	// 按索引从大到小移除，避免索引偏移
	indices := []int{idxMap[cardIDs[0]], idxMap[cardIDs[1]]}
	if indices[0] < indices[1] {
		indices[0], indices[1] = indices[1], indices[0]
	}
	var discarded1, discarded2 Card
	discarded1 = g.removeHandCard(seat, indices[0], events)
	discarded2 = g.removeHandCard(seat, indices[1], events)
	g.DiscardPile = append(g.DiscardPile, discarded1, discarded2)

	target := g.Pending.EffectTarget
	damage := g.Pending.Damage
	if damage <= 0 {
		damage = 1
	}
	pendingCard := g.Pending.Card
	returnIndex := g.Pending.ReturnIndex
	g.Pending = nil
	g.applyDamageWithHook(seat, target, damage, pendingCard, events)
	*events = append(*events, GameEvent{
		Type:        "weapon_skill",
		PlayerIndex: seat,
		TargetIndex: target,
		Message:     fmt.Sprintf("【贯石斧】%s 弃两张牌，杀依然命中 %s", g.Players[seat].Name, g.Players[target].Name),
	})
	resume := DamageResume{
		Mode:        damageResumeShaHit,
		Card:        pendingCard,
		ReturnIndex: returnIndex,
		OfferQilin:  true,
		IgnoreArmor: false,
	}
	if g.Players[target].HP <= 0 {
		if g.afterDamageApplied(seat, target, damage, pendingCard, resume, events) {
			return nil
		}
	}
	g.damageAftermath = nil
	g.resumeAfterDamageNoSkill(resume, target, seat, events)
	return nil
}

// passGuanshifu 发动者选择不弃牌，杀正常被闪避。
func (g *Game) passGuanshifu(seat int, events *[]GameEvent) error {
	if !g.isWeapon9Pending() || seat != g.Pending.SourceIndex {
		return ErrWrongPhase
	}
	source := g.Pending.SourceIndex
	g.Pending = nil
	// 杀被闪避，流程交还 attacker 继续出牌
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = source
	g.Message = fmt.Sprintf("%s 的【杀】被闪避", g.Players[source].Name)
	g.resetTimer()
	return nil
}

func (g *Game) isWeapon9Pending() bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModeWeapon9
}
