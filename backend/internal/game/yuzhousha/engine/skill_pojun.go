package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	ResponseModeSkillPojun        = "skill_pojun"
	ResponseModeSkillPojunDiscard = "skill_pojun_discard"
	counterPojunEndDiscard        = "pojun_end_discard"
)

func (g *Game) initPojunOnShaPending(source, target int, pending *PendingCombat) {
	if !g.hasSkill(source, SkillPojun) || pending == nil {
		return
	}
	pending.PojunMax = g.Players[target].HP
	if pending.PojunMax < 0 {
		pending.PojunMax = 0
	}
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
	if p.PojunPlaced < p.PojunMax && g.hasSkill(p.SourceIndex, SkillPojun) && g.hasTakeableCard(p.TargetIndex) {
		return g.enterPojunPlacing(events)
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
		need := 1
		if g.isSoleShaTarget(p.SourceIndex, victim) {
			need = p.PojunMax
			if need < 1 {
				need = 1
			}
		}
		g.setSkillCounter(victim, counterPojunEndDiscard, need)
	}
	p.ResponseMode = ""
	p.SkillID = ""
	FillPendingRoles(p)
	g.Message = fmt.Sprintf("%s 对 %s 使用【杀】，等待出闪", g.Players[p.SourceIndex].Name, g.Players[victim].Name)
	g.resetTimer()
	return nil
}

func (g *Game) isSoleShaTarget(source, target int) bool {
	_ = source
	_ = target
	return true
}

func (g *Game) discardCampCards(seat int, count int, events *[]GameEvent) {
	if count <= 0 || seat < 0 || seat >= len(g.Players) {
		return
	}
	p := &g.Players[seat]
	for count > 0 && len(p.CampCards) > 0 {
		card := p.CampCards[0]
		p.CampCards = p.CampCards[1:]
		g.DiscardPile = append(g.DiscardPile, card)
		*events = append(*events, GameEvent{
			Type:        "pojun_discard",
			PlayerIndex: seat,
			Card:        &card,
			Message:     fmt.Sprintf("%s 弃置「营」中 %s", p.Name, card.Label),
		})
		count--
	}
	g.syncCounts()
}

func (g *Game) startPojunCampDiscardIfNeeded(seat int, events *[]GameEvent) bool {
	need := g.getSkillCounter(seat, counterPojunEndDiscard)
	if need <= 0 || len(g.Players[seat].CampCards) == 0 {
		g.setSkillCounter(seat, counterPojunEndDiscard, 0)
		return false
	}
	discardN := need
	if discardN > len(g.Players[seat].CampCards) {
		discardN = len(g.Players[seat].CampCards)
	}
	if g.Players[seat].IsAI {
		g.discardCampCards(seat, discardN, events)
		g.setSkillCounter(seat, counterPojunEndDiscard, 0)
		return false
	}
	msg := fmt.Sprintf("%s 回合结束，须弃置 %d 张「营」中牌", g.Players[seat].Name, discardN)
	return g.OpenDiscardWindow(DiscardWindowConfig{
		SkillID:      skill.IDPojun,
		ResponseMode: ResponseModeSkillPojunDiscard,
		ActorSeat:    seat,
		SourceZone:   ZoneCamp,
		MinDiscard:   discardN,
		MaxDiscard:   discardN,
		Message:      msg,
		EventType:    "pojun_discard",
		OnEachDiscard: func(g *Game, card Card, events *[]GameEvent) error {
			msg := fmt.Sprintf("%s 弃置「营」中 %s", g.Players[seat].Name, card.Label)
			*events = append(*events, GameEvent{
				Type:        "pojun_discard",
				PlayerIndex: seat,
				Card:        &card,
				Message:     msg,
			})
			return nil
		},
		OnComplete: func(g *Game, events *[]GameEvent) error {
			g.setSkillCounter(seat, counterPojunEndDiscard, 0)
			return g.endTurnAfterPojunDiscard(seat, events)
		},
	}, events) == nil
}

// PojunDiscardCamp 破军弃「营」（DiscardWindow 薄封装）。
func (g *Game) PojunDiscardCamp(seat int, cardID string, events *[]GameEvent) error {
	return g.DiscardOne(seat, cardID, events)
}

func (g *Game) finishPojunCampDiscardPhase(seat int, events *[]GameEvent) error {
	g.Pending = nil
	g.Phase = PhasePlaying
	return g.endTurnAfterPojunDiscard(seat, events)
}

func (g *Game) endTurnAfterPojunDiscard(seat int, events *[]GameEvent) error {
	g.runTurnEndHooks(seat, events)
	g.Players[seat].Drunk = false
	*events = append(*events, GameEvent{
		Type:        "turn_end",
		PlayerIndex: seat,
		Message:     fmt.Sprintf("%s 结束回合", g.Players[seat].Name),
	})
	g.CurrentTurn = g.nextTurnSeat(seat)
	g.beginTurn(events)
	g.Message = fmt.Sprintf("轮到 %s", g.Players[g.CurrentTurn].Name)
	return nil
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
