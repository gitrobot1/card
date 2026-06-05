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
	p.ResponseMode = ResponseModeSkillPojun
	p.SkillID = skill.IDPojun
	msg := fmt.Sprintf("%s 可发动【破军】，将 %s 至多 %d 张牌置于其武将牌上",
		g.Players[source].Name, g.Players[victim].Name, p.PojunMax-p.PojunPlaced)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDPojun, source, victim, msg)
	g.resetTimer()
	return nil
}

func (g *Game) PojunPlace(source int, zone, cardID string, events *[]GameEvent) error {
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillPojun {
		return ErrWrongPhase
	}
	if source != g.Pending.SourceIndex {
		return ErrNotYourTurn
	}
	if g.Pending.PojunPlaced >= g.Pending.PojunMax {
		return g.finishPojunPlacement(events)
	}
	victim := g.Pending.TargetIndex
	if zone == "" {
		zone = "hand"
	}
	spec := PlayTarget{SeatIndex: victim, Zone: zone, CardID: cardID}
	card, label, ok := g.takeTargetCard(victim, spec, events)
	if !ok {
		return ErrInvalidTarget
	}
	g.Players[victim].CampCards = append(g.Players[victim].CampCards, card)
	g.syncCounts()
	g.Pending.PojunPlaced++
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
	if g.Pending.PojunPlaced >= g.Pending.PojunMax || !g.hasTakeableCard(victim) {
		return g.finishPojunPlacement(events)
	}
	g.Message = fmt.Sprintf("%s 继续发动【破军】", g.Players[source].Name)
	g.resetTimer()
	return nil
}

func (g *Game) PassPojun(source int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillPojun || g.Pending.SourceIndex != source {
		return ErrWrongPhase
	}
	return g.finishPojunPlacement(events)
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
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:    seat,
		TargetIndex:    seat,
		ReturnIndex:    seat,
		ResponseMode:   ResponseModeSkillPojunDiscard,
		SkillID:        skill.IDPojun,
		PojunRemaining: discardN,
	}
	g.Message = fmt.Sprintf("%s 回合结束，须弃置 %d 张「营」中牌", g.Players[seat].Name, discardN)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "pojun_discard_offer",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	return true
}

func (g *Game) PojunDiscardCamp(seat int, cardID string, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillPojunDiscard || g.Pending.TargetIndex != seat {
		return ErrWrongPhase
	}
	idx := -1
	for i, c := range g.Players[seat].CampCards {
		if c.ID == cardID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrInvalidCard
	}
	card := g.Players[seat].CampCards[idx]
	g.Players[seat].CampCards = append(g.Players[seat].CampCards[:idx], g.Players[seat].CampCards[idx+1:]...)
	g.DiscardPile = append(g.DiscardPile, card)
	g.Pending.PojunRemaining--
	*events = append(*events, GameEvent{
		Type:        "pojun_discard",
		PlayerIndex: seat,
		Card:        &card,
		Message:     fmt.Sprintf("%s 弃置「营」中 %s", g.Players[seat].Name, card.Label),
	})
	if g.Pending.PojunRemaining > 0 && len(g.Players[seat].CampCards) > 0 {
		g.Message = fmt.Sprintf("%s 还须弃置 %d 张「营」中牌", g.Players[seat].Name, g.Pending.PojunRemaining)
		g.resetTimer()
		return nil
	}
	g.setSkillCounter(seat, counterPojunEndDiscard, 0)
	return g.finishPojunCampDiscardPhase(seat, events)
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
	aiAutoPojunPlacing(r.g, seat, r.events)
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
	for step := 0; step < 16; step++ {
		if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillPojun {
			return
		}
		if g.Pending.PojunPlaced >= g.Pending.PojunMax {
			_ = g.finishPojunPlacement(events)
			return
		}
		victim := g.Pending.TargetIndex
		if !g.hasTakeableCard(victim) {
			_ = g.finishPojunPlacement(events)
			return
		}
		zone, cardID, ok := aiPickPojunTake(g, source, victim)
		if !ok {
			_ = g.finishPojunPlacement(events)
			return
		}
		if err := g.PojunPlace(source, zone, cardID, events); err != nil {
			_ = g.finishPojunPlacement(events)
			return
		}
	}
}
