package skill

// PreparePhaseDecl 准备阶段（判定/摸牌前）的可选技能入口。
type PreparePhaseDecl struct {
	// Offer 是否在本回合进入准备阶段等待（例如可发动观星）。
	Offer func(r Runtime, seat int) bool
}

// OffersPreparePhase 该技能是否使当前角色进入准备阶段。
func (h Handler) OffersPreparePhase(r Runtime, seat int) bool {
	if h.Decl.PreparePhase.Offer == nil {
		if h.PeekDeckConfig() != nil {
			return PeekCountFor(r, seat, h) > 0
		}
		return false
	}
	return h.Decl.PreparePhase.Offer(r, seat)
}
