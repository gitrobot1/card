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
	// 参考 noname: filter → event.target.isIn() && player.canUse("sha", event.target, false) && player.hasSha()
	if !g.hasWeaponKind(source, CardWeapon3) {
		return false
	}
	// 目标必须还活着
	if g.Players[target].HP <= 0 {
		return false
	}
	// 必须有杀
	if g.countShaInHand(source) == 0 {
		return false
	}
	// 必须能对目标使用杀（帷幕等阻止）
	if g.targetBlockedBySkill(target, CardSha) {
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

// equipMap 返回玩家装备区所有牌的 ID→装备槽映射。
func (g *Game) equipMap(seat int) map[string]string {
	m := make(map[string]string)
	p := &g.Players[seat]
	if p.Weapon != nil {
		m[p.Weapon.ID] = EquipWeapon
	}
	if p.Armor != nil {
		m[p.Armor.ID] = EquipArmor
	}
	if p.PlusHorse != nil {
		m[p.PlusHorse.ID] = EquipPlusHorse
	}
	if p.MinusHorse != nil {
		m[p.MinusHorse.ID] = EquipMinusHorse
	}
	return m
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
// 若 source 装备贯石斧且手牌+装备≥2，可弃两张牌令此杀依然命中。
// 参考 noname: guanshi_skill, filter: countCards("he") >= min (he = hand + equip)
func (g *Game) offerGuanshifu(source, target int, pendingCard Card, damage int, returnIndex int, events *[]GameEvent) bool {
	if !g.hasWeaponKind(source, CardWeapon9) {
		return false
	}
	// 手牌+装备总数≥2（参考 noname: countCards("he") >= 2，he = hand + equip）
	heCount := len(g.Players[source].Hand)
	if g.Players[source].Weapon != nil {
		heCount++
	}
	if g.Players[source].Armor != nil {
		heCount++
	}
	if g.Players[source].PlusHorse != nil {
		heCount++
	}
	if g.Players[source].MinusHorse != nil {
		heCount++
	}
	if heCount < 2 {
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
	g.Message = fmt.Sprintf("【贯石斧】%s 可弃两张牌，令此杀依然命中 %s", g.Players[source].Name, g.Players[target].Name)
	g.appendWeaponSkillEvent(events, source, target, g.Message)
	g.resetTimer()
	return true
}

// resolveGuanshifuDiscard 发动者确认弃两张牌（手牌或装备），令杀命中。
// 参考 noname: chooseToDiscard(2, "he") — he = hand + equip
// 限制：不能弃掉唯一的贯石斧（若装备区只有一张贯石斧且无其他贯石斧技能）
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
	// 不能弃掉唯一的贯石斧（参考 noname: 必须保留至少一张贯石斧）
	guanshiCount := g.countGuanshifu(seat)
	if guanshiCount <= 1 {
		for _, id := range cardIDs {
			if g.Players[seat].Weapon != nil && g.Players[seat].Weapon.ID == id {
				return ErrInvalidCard // 唯一的贯石斧不能弃
			}
		}
	}
	// 构建手牌+装备的索引映射
	handIdx := make(map[string]int)
	for i, c := range g.Players[seat].Hand {
		handIdx[c.ID] = i
	}
	equipMap := g.equipMap(seat)

	var discarded []Card
	for _, id := range cardIDs {
		if idx, ok := handIdx[id]; ok {
			// 先标记，稍后按索引从大到小移除
			discarded = append(discarded, g.Players[seat].Hand[idx])
		} else if equipZone, ok := equipMap[id]; ok {
			card := g.removeEquipCard(seat, equipZone, events)
			g.notifyEquipLost(seat, card, "discard", events)
			discarded = append(discarded, card)
		} else {
			return ErrInvalidCard
		}
	}
	// 移除手牌（按索引从大到小）
	handIDs := make(map[string]bool)
	for _, d := range discarded {
		handIDs[d.ID] = true
	}
	handIndices := make([]int, 0)
	for i, c := range g.Players[seat].Hand {
		if handIDs[c.ID] {
			handIndices = append(handIndices, i)
		}
	}
	// 从大到小排序
	for i := 0; i < len(handIndices); i++ {
		for j := i + 1; j < len(handIndices); j++ {
			if handIndices[i] < handIndices[j] {
				handIndices[i], handIndices[j] = handIndices[j], handIndices[i]
			}
		}
	}
	for _, idx := range handIndices {
		g.removeHandCard(seat, idx, events)
	}
	for _, d := range discarded {
		g.DiscardPile = append(g.DiscardPile, d)
	}

	target := g.Pending.EffectTarget
	damage := g.Pending.Damage
	if damage <= 0 {
		damage = 1
	}
	pendingCard := g.Pending.Card
	returnIndex := g.Pending.ReturnIndex
	g.Pending = nil
	resume := DamageResume{
		Mode:        damageResumeShaHit,
		Card:        pendingCard,
		ReturnIndex: returnIndex,
		OfferQilin:  true,
		IgnoreArmor: false,
	}
	if g.ApplyDamageAndCheckDeath(seat, target, damage, pendingCard, resume, events) {
		return nil
	}
	*events = append(*events, GameEvent{
		Type:        "weapon_skill",
		PlayerIndex: seat,
		TargetIndex: target,
		Message:     fmt.Sprintf("【贯石斧】%s 弃两张牌，杀依然命中 %s", g.Players[seat].Name, g.Players[target].Name),
	})
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

// countGuanshifu 统计玩家拥有的贯石斧数量（装备区1 + 技能途径获得的）。
// 参考 noname: player.hasSkill("guanshi_skill", null, false) 判断是否有额外技能。
func (g *Game) countGuanshifu(seat int) int {
	count := 0
	if g.hasWeaponKind(seat, CardWeapon9) {
		count++
	}
	// TODO: 将来如果有技能可以获得贯石斧技能，在此累加
	return count
}

// ============================================================================
// 丈八蛇矛（Zhangba）：2张手牌当杀使用或打出
// 参考 noname: zhangba_skill, viewAs: { name: "sha" }, selectCard: 2
// ============================================================================

// canZhangbaSha 检查玩家是否能用丈八蛇矛出杀（手牌≥2）。
func (g *Game) canZhangbaSha(seat int) bool {
	if !g.hasWeaponKind(seat, CardWeapon10) {
		return false
	}
	return len(g.Players[seat].Hand) >= 2
}

// TryZhangbaSha 丈八蛇矛出杀：选2张手牌当杀使用（导出供 service 调用）。
// 前端传入两个手牌ID，后端验证并创建虚拟杀牌。
func (g *Game) TryZhangbaSha(seat int, targetIndex int, cardIDs []string, events *[]GameEvent) error {
	if !g.hasWeaponKind(seat, CardWeapon10) {
		return ErrInvalidCard
	}
	if len(cardIDs) != 2 {
		return ErrInvalidCard
	}
	// 验证两张牌都在手牌中且不重复
	idxMap := make(map[string]int)
	for i, c := range g.Players[seat].Hand {
		idxMap[c.ID] = i
	}
	if cardIDs[0] == cardIDs[1] {
		return ErrInvalidCard
	}
	idx1, ok1 := idxMap[cardIDs[0]]
	idx2, ok2 := idxMap[cardIDs[1]]
	if !ok1 || !ok2 {
		return ErrInvalidCard
	}
	// 按索引从大到小移除，避免索引偏移
	if idx1 < idx2 {
		idx1, idx2 = idx2, idx1
	}
	card1 := g.removeHandCard(seat, idx1, events)
	card2 := g.removeHandCard(seat, idx2, events)
	g.DiscardPile = append(g.DiscardPile, card1, card2)
	g.syncCounts()

	// 创建虚拟杀牌（参考 noname: viewAs: { name: "sha" }）
	// 丈八蛇矛的杀无花色、无点数（丈八杀不被仁王盾阻挡，不参与拼点）
	shaCard := Card{
		ID:         fmt.Sprintf("zhangba_%s_%s", card1.ID, card2.ID),
		Kind:       CardSha,
		Name:       "杀",
		Suit:       "", // 无色
		Rank:       0,  // 无点数
		Label:      "丈八杀",
		DamageType: DamageTypeNormal,
	}

	msg := fmt.Sprintf("%s 发动【丈八蛇矛】，将 %s 和 %s 当【杀】使用", g.Players[seat].Name, card1.Label, card2.Label)
	g.appendWeaponSkillEvent(events, seat, targetIndex, msg)
	*events = append(*events, GameEvent{
		Type:        "zhangba_sha",
		PlayerIndex: seat,
		TargetIndex: targetIndex,
		Message:     msg,
	})

	return g.playShaWithCard(seat, shaCard, targetIndex, PlayTarget{}, events)
}
