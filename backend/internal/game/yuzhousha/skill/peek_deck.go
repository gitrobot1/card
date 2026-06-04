package skill

// PeekCardView 亮牌堆顶时的单牌视图（供 AI 分配顶/底，不依赖 engine.Card）。
type PeekCardView struct {
	ID   string
	Kind string
}

// PeekDeckConfig 「观看牌堆顶并分配至顶/底」类技能的通用配置。
// 观星及同类技能在 Decl 上挂载此字段，由 engine 的 PhasePrepare 流程驱动。
type PeekDeckConfig struct {
	// MaxPeekCap 与存活角色数取 min 的上限；0 表示不限（仅按存活数）。
	MaxPeekCap int
	// Count 自定义观看张数；nil 时使用 DefaultPeekCount。
	Count func(r Runtime, seat int) int
	// AIPartition AI 将亮出牌分配至顶/底；nil 时 engine 默认全部置于牌堆顶。
	AIPartition func(r Runtime, seat int, cards []PeekCardView) (topIDs, bottomIDs []string)
}

// DefaultPeekCount 默认观看张数：min(存活数, MaxPeekCap, 牌堆剩余)。
func DefaultPeekCount(r Runtime, seat int, maxCap int) int {
	_ = seat
	n := r.AlivePlayerCount()
	if maxCap > 0 && n > maxCap {
		n = maxCap
	}
	if n > r.DrawPileCount() {
		n = r.DrawPileCount()
	}
	if n < 0 {
		return 0
	}
	return n
}

// PeekCountFor 按技能配置计算本次可观看张数。
func PeekCountFor(r Runtime, seat int, h Handler) int {
	cfg := h.PeekDeckConfig()
	if cfg == nil {
		return 0
	}
	if cfg.Count != nil {
		return cfg.Count(r, seat)
	}
	return DefaultPeekCount(r, seat, cfg.MaxPeekCap)
}
