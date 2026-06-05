package engine

// PendingActorSeat 返回当前应由哪一座位玩家（或 AI）行动；-1 表示无待操作。
func (g *Game) PendingActorSeat() int {
	if g.IsFinished() {
		return -1
	}
	if g.Phase == PhaseResponse && g.Pending != nil {
		g.ensurePendingRoles()
		if g.Pending.ActorSeat >= 0 {
			return g.Pending.ActorSeat
		}
		// fallback：未识别 mode 时沿用旧推导（迁移完成后删除）
		p := g.Pending
		if p.TieqiPending {
			return p.SourceIndex
		}
		switch p.ResponseMode {
		case "skill_pojun":
			return p.SourceIndex
		case ResponseModeDying:
			return p.SourceIndex
		case ResponseModeWuguPick:
			return p.WuguPickSeat
		case "skill_pojun_discard":
			return p.TargetIndex
		default:
			return p.TargetIndex
		}
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
