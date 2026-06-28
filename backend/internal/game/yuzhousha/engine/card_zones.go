package engine

import "fmt"

// placeTakenCard 将取到的牌放入目标区（TakeWindow 用）。
func (g *Game) placeTakenCard(dest TakeDestination, card Card, events *[]GameEvent) error {
	seat := dest.Seat
	if seat < 0 || seat >= len(g.Players) {
		return fmt.Errorf("invalid destination seat %d", seat)
	}
	p := &g.Players[seat]
	switch dest.Zone {
	case ZoneHand:
		p.Hand = append(p.Hand, card)
		g.SyncCounts()
		return nil
	case ZoneCamp:
		p.CampCards = append(p.CampCards, card)
		g.SyncCounts()
		return nil
	case ZoneDiscard:
		g.DiscardPile = append(g.DiscardPile, card)
		g.SyncCounts()
		return nil
	case ZoneVoid:
		g.SyncCounts()
		return nil
	default:
		return fmt.Errorf("unsupported take destination zone %q", dest.Zone)
	}
}
