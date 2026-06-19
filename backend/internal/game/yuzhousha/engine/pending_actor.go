package engine

// PendingActorSeat 返回当前应由哪一座位玩家（或 AI）行动；-1 表示无待操作。
func (g *Game) PendingActorSeat() int {
	if g.IsFinished() {
		return -1
	}
	if g.Phase == PhaseResponse && g.Pending != nil {
		g.ensurePendingRoles()
		return g.Pending.ActorSeat
	}
	if g.Phase == PhasePlaying {
		if g.TurnStep == StepPrepare || g.TurnStep == StepPlay || g.TurnStep == StepDiscard {
			return g.CurrentTurn
		}
		if g.TurnStep == StepDraw && g.isDrawPhaseChoicePending(g.CurrentTurn) {
			return g.CurrentTurn
		}
	}
	return -1
}
