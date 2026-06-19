package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	ResponseModeSkillPojun = "skill_pojun"
)

// initPojunOnShaPending 用杀指定目标后，初始化破军：
// 1. 设置 PojunMax（目标体力值），用于后续拿牌
// 2. 若目标手牌数≤源且装备数≤源，伤害+1
func (g *Game) initPojunOnShaPending(source, target int, pending *PendingCombat) {
	if !g.hasSkill(source, SkillPojun) || pending == nil {
		return
	}
	pending.PojunMax = g.Players[target].HP
	if pending.PojunMax < 0 {
		pending.PojunMax = 0
	}
	// 效果2：对手牌数≤源 且 装备数≤源 的角色，伤害+1
	src := &g.Players[source]
	tgt := &g.Players[target]
	srcEquipCount := countEquip(src)
	if tgt.HandCount <= src.HandCount && countEquip(tgt) <= srcEquipCount {
		if pending.Damage <= 0 {
			pending.Damage = 2
		} else {
			pending.Damage++
		}
	}
}

// countEquip 统计玩家的装备数
func countEquip(p *Player) int {
	n := 0
	if p.Weapon != nil {
		n++
	}
	if p.Armor != nil {
		n++
	}
	if p.PlusHorse != nil {
		n++
	}
	if p.MinusHorse != nil {
		n++
	}
	return n
}

func (g *Game) advanceShaBeforeTargetResponse(events *[]GameEvent) error {
	p := g.Pending
	if p == nil || p.Card.Kind != CardSha {
		return nil
	}
	if p.TieqiPending || p.ResponseMode == ResponseModeSkillLiuli {
		return nil
	}
	if p.ResponseMode == ResponseModeSkillPojun {
		return nil
	}
	if p.PojunPlaced < p.PojunMax && g.hasSkill(p.SourceIndex, SkillPojun) {
		if !g.hasTakeableCard(p.TargetIndex) {
			// 目标没有可拿的牌，跳过拿牌直接继续
			return g.finishPojunPlacement(events)
		}
		return g.enterPojunPlacing(events)
	}
	// 雌雄双股剑：出杀指定目标后，若目标为异性则触发
	if g.tryOfferChixiongOnSha(events) {
		return nil
	}
	p.ResponseMode = ""
	p.SkillID = ""
	g.resetTimer()
	return nil
}

func (g *Game) enterPojunPlacing(events *[]GameEvent) error {
	p := g.Pending
	source := p.SourceIndex
	victim := p.TargetIndex
	maxTake := p.PojunMax - p.PojunPlaced
	if maxTake <= 0 {
		return g.finishPojunPlacement(events)
	}
	msg := fmt.Sprintf("%s 可发动【破军】，将 %s 至多 %d 张牌置于其武将牌上",
		g.Players[source].Name, g.Players[victim].Name, maxTake)
	return g.OpenTakeWindowOnPending(TakeWindowConfig{
		SkillID:          skill.IDPojun,
		ResponseMode:     ResponseModeSkillPojun,
		ActorSeat:        source,
		SubjectSeat:      victim,
		OriginSeat:       source,
		MaxTake:          maxTake,
		Destination:      TakeDestination{Zone: ZoneCamp, Seat: victim},
		Message:          msg,
		EventType:        "pojun_place",
		PassClosesWindow: true,
		PickTarget:       aiPickPojunTake,
		OnEachTake:       pojunOnEachTake,
		OnComplete:       pojunTakeComplete,
	}, events)
}

func pojunOnEachTake(g *Game, card Card, label string, events *[]GameEvent) error {
	p := g.Pending
	if p == nil {
		return ErrWrongPhase
	}
	source := p.SourceIndex
	victim := p.TargetIndex
	p.PojunPlaced++
	msg := fmt.Sprintf("%s 发动【破军】，将 %s 的%s置于「营」", g.Players[source].Name, g.Players[victim].Name, label)
	g.appendSkillEvent(events, skill.IDPojun, source, victim, msg)
	*events = append(*events, GameEvent{
		Type:        "pojun_place",
		PlayerIndex: source,
		TargetIndex: victim,
		Card:        &card,
		Message:     msg,
	})
	g.runHandEmptyHooks(victim, events)
	return nil
}

func pojunTakeComplete(g *Game, events *[]GameEvent) error {
	return g.finishPojunPlacement(events)
}

// PojunPlace 破军拿牌入「营」（TakeWindow 薄封装）。
func (g *Game) PojunPlace(source int, zone, cardID string, events *[]GameEvent) error {
	if zone == "" {
		zone = "hand"
	}
	return g.TakeOne(source, ZoneID(zone), cardID, events)
}

// PassPojun 结束破军拿牌窗口（TakeWindow 薄封装）。
func (g *Game) PassPojun(source int, events *[]GameEvent) error {
	return g.PassTake(source, events)
}

func (g *Game) finishPojunPlacement(events *[]GameEvent) error {
	p := g.Pending
	if p == nil {
		return ErrWrongPhase
	}
	victim := p.TargetIndex
	if p.PojunPlaced > 0 {
		// 新版破军：回合结束后，该角色获得「营」中的牌
		g.setSkillCounter(victim, "pojun_gain_pending", 1)
	}
	p.ResponseMode = ""
	p.SkillID = ""
	FillPendingRoles(p)
	g.Message = fmt.Sprintf("%s 对 %s 使用【杀】，等待出闪", g.Players[p.SourceIndex].Name, g.Players[victim].Name)
	g.resetTimer()
	return nil
}

// giveCampCardsToPlayer 将目标的「营」中牌移回其手牌（新版破军效果）
func (g *Game) giveCampCardsToPlayer(seat int, events *[]GameEvent) {
	p := &g.Players[seat]
	if len(p.CampCards) == 0 {
		return
	}
	for _, card := range p.CampCards {
		p.Hand = append(p.Hand, card)
		*events = append(*events, GameEvent{
			Type:        "pojun_gain",
			PlayerIndex: seat,
			Card:        &card,
			Message:     fmt.Sprintf("%s 获得「营」中的 %s", p.Name, card.Label),
		})
	}
	p.CampCards = nil
	g.syncCounts()
}

// startPojunGainIfNeeded 回合结束时，若目标有「营」中牌，获得这些牌
func (g *Game) startPojunGainIfNeeded(seat int, events *[]GameEvent) {
	if g.getSkillCounter(seat, "pojun_gain_pending") <= 0 {
		return
	}
	g.setSkillCounter(seat, "pojun_gain_pending", 0)
	g.giveCampCardsToPlayer(seat, events)
}

func aiPickPojunTake(g *Game, source, victim int) (zone, cardID string, ok bool) {
	p := &g.Players[victim]
	if len(p.Hand) > 0 {
		return "hand", p.Hand[0].ID, true
	}
	for _, slot := range []struct {
		zone string
		card **Card
	}{
		{EquipWeapon, &p.Weapon},
		{EquipArmor, &p.Armor},
		{EquipPlusHorse, &p.PlusHorse},
		{EquipMinusHorse, &p.MinusHorse},
	} {
		if *slot.card != nil {
			return slot.zone, (*slot.card).ID, true
		}
	}
	_ = source
	return "", "", false
}

func (r *gameSkillRuntime) PassPojun(seat int) error {
	return r.g.PassPojun(seat, r.events)
}

func (r *gameSkillRuntime) PojunPlace(seat int, zone, cardID string) error {
	return r.g.PojunPlace(seat, zone, cardID, r.events)
}

func (r *gameSkillRuntime) AutoPojunPlacing(seat int) error {
	r.g.AutoTakeWindow(seat, r.events)
	return nil
}

func (r *gameSkillRuntime) PendingPojunForSource(seat int) bool {
	if r.g.Pending == nil {
		return false
	}
	p := r.g.Pending
	return p.ResponseMode == ResponseModeSkillPojun && p.SourceIndex == seat
}

func aiAutoPojunPlacing(g *Game, source int, events *[]GameEvent) {
	g.AutoTakeWindow(source, events)
}
