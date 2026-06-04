//go:build cardtest

package uno

// 以下导出仅供 backend/test 下的外部测试使用（需 -tags cardtest）。

func (g *Game) CanPlayCardForTest(seat int, card Card) bool { return g.canPlayCard(seat, card) }

func (g *Game) SyncCountsForTest() { g.syncCounts() }

func (g *Game) FinalizeRollRoundForTest(events *[]GameEvent) error { return g.finalizeRollRound(events) }

func (g *Game) SetRollRoundSum(seat, sum int) { g.rollRoundSums[seat] = sum }

func (g *Game) RollRoundSum(seat int) int { return g.rollRoundSums[seat] }

func (g *Game) RollContenders() []int { return append([]int(nil), g.rollContenders...) }

func (g *Game) CheckAfterEliminationForTest(events *[]GameEvent) { g.checkAfterElimination(events) }

func FilterEventsForSeat(events []GameEvent, seat int) []GameEvent {
	return filterEventsForSeat(events, seat)
}
