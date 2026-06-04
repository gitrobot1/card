package engine

func (g *Game) tryJiangDraw(source int, card Card, events *[]GameEvent) {
	if source < 0 || source >= len(g.Players) {
		return
	}
	g.runCardResolvedHooks(source, card, events)
}

func (g *Game) isJueqingHarm(source int) bool {
	return g.damageAsHPLossViaHooks(source)
}
