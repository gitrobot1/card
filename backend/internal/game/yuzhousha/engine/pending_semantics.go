package engine

const (
	WindowKindRespond = "respond"
	WindowKindTake    = "take"
	WindowKindDiscard = "discard"
	WindowKindChoice  = "choice"
	WindowKindPeek    = "peek"
	WindowKindPick    = "pick"
)

// FillPendingRoles 按 ResponseMode 与 legacy 字段推导 Actor/Subject/Origin/WindowKind。
// 不修改技能行为，仅显式化 pending 语义供 JSON 与 PendingActorSeat 使用。
func FillPendingRoles(p *PendingCombat) {
	if p == nil {
		return
	}

	if p.TieqiPending {
		p.WindowKind = WindowKindChoice
		p.ActorSeat = p.SourceIndex
		p.SubjectSeat = p.TargetIndex
		p.OriginSeat = p.SourceIndex
		return
	}

	switch p.ResponseMode {
	case ResponseModeSkillFankui, ResponseModeSkillTuxi, ResponseModeSkillQixi:
		p.WindowKind = WindowKindTake
		p.ActorSeat = p.TargetIndex
		p.SubjectSeat = p.SourceIndex
		p.OriginSeat = p.TargetIndex
	case "skill_pojun":
		p.WindowKind = WindowKindTake
		p.ActorSeat = p.SourceIndex
		p.SubjectSeat = p.TargetIndex
		p.OriginSeat = p.SourceIndex
	case "skill_pojun_discard", ResponseModeSkillYinghunDiscard:
		p.WindowKind = WindowKindDiscard
		p.ActorSeat = p.TargetIndex
		p.SubjectSeat = p.TargetIndex
		p.OriginSeat = p.TargetIndex
	case ResponseModeSkillGanglieChoice, ResponseModeSkillYinghun:
		p.WindowKind = WindowKindChoice
		p.ActorSeat = p.TargetIndex
		p.SubjectSeat = p.TargetIndex
		p.OriginSeat = p.SourceIndex
	case ResponseModePeekDeck:
		p.WindowKind = WindowKindPeek
		p.ActorSeat = p.TargetIndex
		p.SubjectSeat = p.TargetIndex
		p.OriginSeat = p.SourceIndex
	case ResponseModeWuguPick:
		p.WindowKind = WindowKindPick
		p.ActorSeat = p.WuguPickSeat
		p.SubjectSeat = p.WuguPickSeat
		p.OriginSeat = p.SourceIndex
	case ResponseModeDying:
		p.WindowKind = WindowKindRespond
		p.ActorSeat = p.SourceIndex
		p.SubjectSeat = p.TargetIndex
		p.OriginSeat = p.TargetIndex
	default:
		p.WindowKind = WindowKindRespond
		p.ActorSeat = p.TargetIndex
		p.SubjectSeat = p.TargetIndex
		p.OriginSeat = p.SourceIndex
	}
}

func (g *Game) ensurePendingRoles() {
	if g.Pending != nil {
		FillPendingRoles(g.Pending)
	}
}

// PendingSubjectSeat 返回被操作座位；-1 表示当前无 response pending。
func (g *Game) PendingSubjectSeat() int {
	if g.Phase != PhaseResponse || g.Pending == nil {
		return -1
	}
	g.ensurePendingRoles()
	return g.Pending.SubjectSeat
}

// IsActorSeat 判断 seat 是否为当前 pending 的 Actor。
func (g *Game) IsActorSeat(seat int) bool {
	actor := g.PendingActorSeat()
	return actor >= 0 && actor == seat
}
