package skill

// HookKind 引擎广播的技能 hook 类型；play/judge 等只调 engine.runSkillHooks。
type HookKind string

const (
	HookTargetBlocked       HookKind = "target_blocked"
	HookDistanceDelta       HookKind = "distance_delta"
	HookTrickIgnoresDistance HookKind = "trick_ignores_distance"
	HookInstantTrickUsed    HookKind = "instant_trick_used"
	HookCardPlaysAs         HookKind = "card_plays_as"
	HookUnlimitedSha        HookKind = "unlimited_sha"
	HookDamageDealt         HookKind = "damage_dealt"
	HookJudgeResult         HookKind = "judge_result"
	HookCardsDiscarded      HookKind = "cards_discarded"
	HookEquipLost           HookKind = "equip_lost"
)

// JudgeReason 判定来源，供 OnJudgeResult 区分铁骑/八卦/延时锦囊等。
type JudgeReason string

const (
	JudgeTieqi     JudgeReason = "tieqi"
	JudgeBagua     JudgeReason = "bagua"
	JudgeShandian  JudgeReason = "shandian"
	JudgeLebu      JudgeReason = "lebu"
	JudgeBingliang JudgeReason = "bingliang"
	JudgeLuoshen   JudgeReason = "luoshen"
	JudgeGanglie   JudgeReason = "ganglie"
	JudgeLeiji     JudgeReason = "leiji"
)

// CardView 传给技能逻辑的最小牌面信息（避免 engine 类型泄漏）。
type CardView struct {
	ID    string
	Kind  string
	Suit  string
	Label string
	Name  string
	Rank  int
}

// DamageCtx 造成伤害后广播。
type DamageCtx struct {
	Source   int
	Target   int
	Amount   int
	CardKind string
	CardName string
	Card     CardView // 造成伤害的牌（若有）
}

// JudgeCtx 翻判定牌后广播。
type JudgeCtx struct {
	Seat   int
	Reason JudgeReason
	Card   CardView
	IsRed  bool
}

// CardsDiscardedCtx 牌进入弃牌堆后广播（弃牌阶段、技能弃牌等）。
type CardsDiscardedCtx struct {
	Seat   int
	Reason string // discard_phase | cost | effect
	Cards  []CardView
}

// EquipLostCtx 装备区失去牌时广播（被拆、被拿、替换等）。
type EquipLostCtx struct {
	Seat   int
	Reason string // taken | replace | discard
	Card   CardView
}

// CardResolvedCtx 【杀】/【决斗】生效或被无懈抵消后广播（如【激昂】）。
type CardResolvedCtx struct {
	Seat int
	Card CardView
}

// HookCall 单次 hook 调用的参数；按 Kind 填对应字段。
type HookCall struct {
	Kind HookKind

	Seat   int
	From   int
	To     int
	Target int

	CardKind  string
	AsKind    string
	Suit      string
	TrickKind string

	Damage   *DamageCtx
	Judge    *JudgeCtx
	Discarded *CardsDiscardedCtx
	EquipLost *EquipLostCtx
}

// HookResult 聚合型 hook 的返回值。
type HookResult struct {
	Bool bool
	Int  int
	Err  error
}
