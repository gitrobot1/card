package engine

func (g *Game) wushuangResponsesNeeded(source int, cardKind string) int {
	return 1 + g.extraResponsesNeededViaHooks(source, cardKind)
}

func (g *Game) consumeWushuangResponse(pending *PendingCombat, responder int, requiredKind string) bool {
	if pending == nil || pending.ResponsesNeeded <= 1 {
		return false
	}
	pending.ResponsesNeeded--
	kindName := "闪"
	if requiredKind == CardSha {
		kindName = "杀"
	}
	g.Message = g.Players[responder].Name + " 还需再出一张【" + kindName + "】（【无双】）"
	g.resetTimer()
	return true
}

func (g *Game) appendWushuangMessage(source int, cardKind string, msg *string) {
	if g.wushuangResponsesNeeded(source, cardKind) > 1 {
		need := "闪"
		if cardKind == CardJueDou {
			need = "杀"
		}
		*msg += "（【无双】需两张" + need + "）"
	}
}
