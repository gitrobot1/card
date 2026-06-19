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
	HookDamageCalculated    HookKind = "damage_calculated"    // 伤害值计算完后（可修改）
	HookDamageDealt         HookKind = "damage_dealt"
	HookBeforeHPChange      HookKind = "before_hp_change"    // 扣血前（可防止）
	HookHPLost              HookKind = "hp_lost"              // 血量流失后
	HookHPChanged           HookKind = "hp_changed"           // 血量变化后
	HookJudgeResult         HookKind = "judge_result"
	HookCardsDiscarded      HookKind = "cards_discarded"
	HookEquipLost           HookKind = "equip_lost"
	HookOnDeath             HookKind = "on_death"       // 阵亡时（亡语，牌还在）
	HookAfterDeath          HookKind = "after_death"    // 阵亡后（牌已弃）
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
	Source      int
	Target      int
	Amount      int
	FinalAmount int // 经过 OnDamageCalculated 修改后的最终伤害值
	CardKind    string
	CardName    string
	Card        CardView // 造成伤害的牌（若有）
}

// DamageCalculatedCtx 伤害值计算完后广播（可修改伤害值）。
type DamageCalculatedCtx struct {
	Source   int
	Target   int
	Amount   int    // 当前伤害值，监听器可修改此值
	CardKind string
	CardName string
	Card     CardView
}

// BeforeHPChangeCtx 扣血前广播（可防止扣血）。
type BeforeHPChangeCtx struct {
	Source  int
	Target  int
	Amount  int // 即将造成的伤害值
	Cancel  bool // 监听器设为 true 可防止扣血
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

// HPLostCtx 血量流失后广播（非伤害导致的扣血，如【蛊惑】、【刚烈】等）。
type HPLostCtx struct {
	Seat   int
	Amount int
	Reason string // skill | card_effect
	Source int    // 伤害来源（若有）
}

// HPChangedCtx 血量变化后广播（伤害/流失/回复）。
type HPChangedCtx struct {
	Seat      int
	OldHP     int
	NewHP     int
	Delta     int // 变化量：正数=回复，负数=扣血
	Reason    string // damage | hp_loss | heal | skill
	Source    int    // 来源（若有）
	SkillID   string // 技能ID（若是技能导致）
}

// CardResolvedCtx 【杀】/【决斗】生效或被无懈抵消后广播（如【激昂】）。
type CardResolvedCtx struct {
	Seat int
	Card CardView
}

// BecomeTargetCtx 成为某张牌的目标时广播（如【激昂】）。
type BecomeTargetCtx struct {
	Seat   int       // 成为目标的座位号
	Source int       // 使用牌的座位号
	Card   CardView  // 牌的信息
}

// DeathCtx 阵亡时广播（亡语时机，牌还在）。
type DeathCtx struct {
	Victim int    // 阵亡者座位号
	Killer int    // 凶手座位号（若无则为 -1）
	Reason string // damage | hp_loss | skill（死亡原因）
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

	Damage         *DamageCtx
	DamageCalculated *DamageCalculatedCtx
	BeforeHPChange  *BeforeHPChangeCtx
	Death          *DeathCtx
	HPLost         *HPLostCtx
	HPChanged      *HPChangedCtx
	Judge          *JudgeCtx
	Discarded      *CardsDiscardedCtx
	EquipLost      *EquipLostCtx
}

// HookResult 聚合型 hook 的返回值。
type HookResult struct {
	Bool bool
	Int  int
	Err  error
}
