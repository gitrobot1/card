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
		// 铁骑：出杀者可选择发动，但当前窗口仍是"目标需出闪"
		// ActorSeat 应为目标（需出闪的人），SourceIndex 在 PassResponse 中特殊处理
		p.WindowKind = WindowKindRespond
		p.ActorSeat = p.TargetIndex
		p.SubjectSeat = p.TargetIndex
		p.OriginSeat = p.SourceIndex
		return
	}

	switch p.ResponseMode {
	case ResponseModeWuxiekLebu,
		ResponseModeWuxiekBingliang, ResponseModeWuxiekShandian,
		ResponseModeWuxiekGuose:
		// 判定前无懈可击窗口：ActorSeat 由响应队列管理（startJudgeWuxiekWindow / advance*Queue），不覆盖
		// TargetIndex = -1 表示任何人都可以响应
		p.WindowKind = WindowKindRespond
		p.SubjectSeat = p.SourceIndex
		p.OriginSeat = p.SourceIndex
		return
	case ResponseModeWuxiekTrick:
		// 锦囊牌无懈可击窗口：
		// - TargetIndex >= 0：初始无懈可击窗口，TargetIndex 是当前可响应者
		// - TargetIndex = -1：反无懈可击窗口（startWuxiekRecursiveWindow），ActorSeat 由响应队列管理
		p.WindowKind = WindowKindRespond
		if p.TargetIndex >= 0 {
			p.ActorSeat = p.TargetIndex
			p.SubjectSeat = p.TargetIndex
		} else {
			p.SubjectSeat = p.SourceIndex
		}
		p.OriginSeat = p.SourceIndex
		return
	case ResponseModeSkillFankui, ResponseModeSkillTuxi:
		p.WindowKind = WindowKindTake
		p.ActorSeat = p.TargetIndex
		p.SubjectSeat = p.SourceIndex
		p.OriginSeat = p.TargetIndex
	case "skill_pojun":
		// 破军：OpenTakeWindowOnPending 已正确设置 ActorSeat/SubjectSeat，不覆盖
		return
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
	case ResponseModeGuoHe, ResponseModeTanNang:
		// 过河拆桥/顺手牵羊的 TakeWindow：Actor/Subject/WindowKind 由 OpenTakeWindow 已设置，不覆盖
		return
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

// CanRespondSeat 判断 seat 是否可以响应当前 Pending。
// 当 TargetIndex == -1 时，任何人都可以响应（用于无懈可击等全局响应）。
func (g *Game) CanRespondSeat(seat int) bool {
	if g.Pending == nil {
		return false
	}
	if g.Pending.TargetIndex == -1 {
		return seat >= 0 && seat < len(g.Players)
	}
	return g.Pending.TargetIndex == seat
}

// CanRespondWuxiek 检查是否可以出无懈可击（比 CanRespondSeat 更宽松，用于无懈可击专用检查）
func (g *Game) CanRespondWuxiek(seat int) bool {
	if g.Pending == nil {
		return false
	}
	// 无懈链期间：任何人都可以出无懈可击
	if g.Pending.ResponseMode == ResponseModeWuxiekTrick {
		return seat >= 0 && seat < len(g.Players)
	}
	return g.CanRespondSeat(seat)
}
