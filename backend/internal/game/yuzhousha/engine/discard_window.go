package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

type DiscardWindowConfig struct {
	SkillID      string
	ResponseMode string
	ActorSeat    int

	SourceZone ZoneID
	MinDiscard int
	MaxDiscard int

	Message   string
	EventType string

	OnEachDiscard func(g *Game, card Card, events *[]GameEvent) error
	OnComplete    func(g *Game, events *[]GameEvent) error
}

type discardWindowState struct {
	cfg       DiscardWindowConfig
	discarded int
}

func (dw *discardWindowState) remaining() int {
	if dw.cfg.MaxDiscard <= 0 {
		return 0
	}
	return dw.cfg.MaxDiscard - dw.discarded
}

func (g *Game) OpenDiscardWindow(cfg DiscardWindowConfig, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if cfg.MaxDiscard <= 0 {
		return ErrInvalidTarget
	}
	if cfg.MinDiscard <= 0 {
		cfg.MinDiscard = cfg.MaxDiscard
	}
	if cfg.ActorSeat < 0 {
		return ErrInvalidTarget
	}
	if cfg.SourceZone == "" {
		cfg.SourceZone = ZoneHand
	}
	if cfg.EventType == "" {
		cfg.EventType = "discard_window"
	}

	g.Phase = PhaseResponse
	seat := cfg.ActorSeat
	g.Pending = &PendingCombat{
		SourceIndex:    seat,
		TargetIndex:    seat,
		ReturnIndex:    seat,
		ResponseMode:   cfg.ResponseMode,
		SkillID:        cfg.SkillID,
		WindowKind:     WindowKindDiscard,
		ActorSeat:      seat,
		SubjectSeat:    seat,
		OriginSeat:     seat,
		PojunRemaining: cfg.MaxDiscard,
	}
	FillPendingRoles(g.Pending)
	g.discardWindow = &discardWindowState{cfg: cfg}
	if cfg.Message != "" {
		g.Message = cfg.Message
	}
	if events != nil && cfg.Message != "" {
		offerType := cfg.EventType + "_offer"
		*events = append(*events, GameEvent{
			Type:        offerType,
			PlayerIndex: seat,
			Message:     cfg.Message,
		})
	}
	g.resetTimer()
	return nil
}

func (g *Game) syncLegacyDiscardRemaining() {
	if g.discardWindow == nil || g.Pending == nil {
		return
	}
	// 新版破军不再需要 DiscardWindow，此处留空
	_ = g.Pending.ResponseMode
}

func (g *Game) validateDiscardActor(actor int) error {
	if g.Phase != PhaseResponse || g.Pending == nil || g.discardWindow == nil {
		return ErrWrongPhase
	}
	if g.Pending.WindowKind != WindowKindDiscard {
		return ErrWrongPhase
	}
	if !g.IsActorSeat(actor) {
		return ErrNotYourTurn
	}
	if g.discardWindow.remaining() <= 0 {
		return ErrWrongPhase
	}
	return nil
}

func (g *Game) HasDiscardableInWindow(actor int, zone ZoneID) bool {
	if zone == "" {
		zone = ZoneHand
	}
	p := &g.Players[actor]
	switch zone {
	case ZoneHand:
		return len(p.Hand) > 0
	case ZoneCamp:
		return len(p.CampCards) > 0
	default:
		return false
	}
}

func (g *Game) removeCardFromActorZone(actor int, zone ZoneID, cardID string) (Card, bool) {
	p := &g.Players[actor]
	switch zone {
	case ZoneCamp:
		for i, c := range p.CampCards {
			if c.ID == cardID {
				card := c
				p.CampCards = append(p.CampCards[:i], p.CampCards[i+1:]...)
				g.syncCounts()
				return card, true
			}
		}
	case ZoneHand:
		for i, c := range p.Hand {
			if c.ID == cardID {
				card := c
				p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
				g.syncCounts()
				return card, true
			}
		}
	}
	return Card{}, false
}

func (g *Game) DiscardOne(actor int, cardID string, events *[]GameEvent) error {
	if err := g.validateDiscardActor(actor); err != nil {
		return err
	}
	dw := g.discardWindow
	card, ok := g.removeCardFromActorZone(actor, dw.cfg.SourceZone, cardID)
	if !ok {
		return ErrInvalidCard
	}
	g.DiscardPile = append(g.DiscardPile, card)
	msg := fmt.Sprintf("%s 弃置 %s", g.Players[actor].Name, card.Label)
	if dw.cfg.OnEachDiscard != nil {
		if err := dw.cfg.OnEachDiscard(g, card, events); err != nil {
			return err
		}
	} else if events != nil {
		if dw.cfg.SkillID != "" {
			g.appendSkillEvent(events, dw.cfg.SkillID, actor, actor, msg)
		}
		*events = append(*events, GameEvent{
			Type:        dw.cfg.EventType,
			PlayerIndex: actor,
			Card:        &card,
			Message:     msg,
		})
	}
	dw.discarded++
	g.syncLegacyDiscardRemaining()
	if dw.remaining() > 0 && g.HasDiscardableInWindow(actor, dw.cfg.SourceZone) {
		g.Message = fmt.Sprintf("%s 还须弃置 %d 张", g.Players[actor].Name, dw.remaining())
		g.resetTimer()
		return nil
	}
	return g.finishDiscardWindow(events)
}

func (g *Game) PassDiscardWindow(actor int, events *[]GameEvent) error {
	if err := g.validateDiscardActor(actor); err != nil {
		return err
	}
	dw := g.discardWindow
	if dw.discarded < dw.cfg.MinDiscard && g.HasDiscardableInWindow(actor, dw.cfg.SourceZone) {
		return ErrWrongPhase
	}
	return g.finishDiscardWindow(events)
}

func (g *Game) finishDiscardWindow(events *[]GameEvent) error {
	dw := g.discardWindow
	onComplete := dw.cfg.OnComplete
	g.discardWindow = nil
	if onComplete != nil {
		return onComplete(g, events)
	}
	g.Pending = nil
	g.Phase = PhasePlaying
	return nil
}

func (g *Game) AutoDiscardWindow(actor int, events *[]GameEvent) {
	if g.discardWindow == nil || !g.IsActorSeat(actor) {
		return
	}
	zone := g.discardWindow.cfg.SourceZone
	for g.discardWindow != nil && g.discardWindow.remaining() > 0 && g.HasDiscardableInWindow(actor, zone) {
		p := &g.Players[actor]
		var cardID string
		switch zone {
		case ZoneCamp:
			if len(p.CampCards) == 0 {
				break
			}
			cardID = p.CampCards[0].ID
		case ZoneHand:
			if len(p.Hand) == 0 {
				break
			}
			cardID = p.Hand[0].ID
		default:
			break
		}
		if cardID == "" {
			break
		}
		if err := g.DiscardOne(actor, cardID, events); err != nil {
			break
		}
	}
}

func (g *Game) autoDiscardWindowIfNeeded(events *[]GameEvent) bool {
	if g.discardWindow == nil || g.Pending == nil {
		return false
	}
	seat := g.PendingActorSeat()
	if seat < 0 || !g.Players[seat].IsAI {
		return false
	}
	if g.Pending.SkillID != "" {
		rt := g.skillRuntime(events)
		if h, ok := skill.Lookup(g.Pending.SkillID); ok && h.CanActivate(rt, seat) {
			if err := h.AIActivate(rt, seat); err != nil {
				_ = g.PassDiscardWindow(seat, events)
			}
			return true
		}
	}
	g.AutoDiscardWindow(seat, events)
	return true
}
