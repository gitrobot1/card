package skill

// HookKind 引擎广播的技能 hook 类型；play/judge 等只调 engine.runSkillHooks。
type HookKind string

// HookRole 技能触发时的角色维度（参考 noname: player/source/target/global）。
type HookRole int

const (
	RolePlayer HookRole = iota // 事件主体（如"自己受伤"）
	RoleSource                  // 事件来源（如"自己造成伤害"）
	RoleTarget                  // 事件目标（如"被指定为目标"）
	RoleGlobal                  // 全局监听（如"任何人判定"）
)

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
	HookModJudge            HookKind = "mod_judge"       // mod.judge 被动修改判定结果（参考 noname: mod.judge）
	HookJudgeFixing         HookKind = "judge_fixing"    // 判定修正后最终确认（参考 noname: judgeFixing）
	HookBlocksWuxiek        HookKind = "blocks_wuxiek"   // 阻止无懈可击（参考 noname: playernowuxie 技能标签）

	// ===== 阶段钩子（参考 noname: phaseBeforeStart/phaseBeforeEnd/phaseBegin/phaseEnd 等） =====
	HookPhaseBeforeStart HookKind = "phase_before_start"
	HookPhaseBeforeEnd   HookKind = "phase_before_end"
	HookPhaseBeginStart  HookKind = "phase_begin_start"
	HookPhaseBegin       HookKind = "phase_begin"
	HookPhaseChange      HookKind = "phase_change"
	HookPhaseEnd         HookKind = "phase_end"
	HookRoundStart       HookKind = "round_start"        // 新一轮开始（noname: roundStart）
	HookTurnBegin        HookKind = "turn_begin"         // 回合开始
	HookTurnEnd          HookKind = "turn_end"           // 回合结束

	// ===== 杀流程钩子 =====
	HookShaBegin         HookKind = "sha_begin"          // 杀开始结算（仁王盾触发点，noname: shaBegin）
	HookBecomeShaTarget  HookKind = "become_sha_target"  // 成为杀的目标后（琉璃等，noname: becomeTarget after shaBegin）
	HookShaMiss          HookKind = "sha_miss"           // 杀被闪抵消（青龙刀/贯石斧，noname: shaMiss）
	HookShaHit           HookKind = "sha_hit"            // 杀命中（麒麟弓，noname: shaHit）

	// ===== 伤害流程钩子 =====
	HookDamageBegin HookKind = "damage_begin" // 伤害开始结算（白银狮子，noname: damageBegin1~4）
	HookDamageEnd   HookKind = "damage_end"   // 伤害结算完毕（刚烈/反馈，noname: damageEnd）

	// ===== 锦囊/牌使用钩子 =====
	HookUseCard         HookKind = "use_card"          // 使用牌（集智，noname: useCard）
	HookUseCardToTarget HookKind = "use_card_to_target" // 牌指定目标后（雌雄双股剑，noname: useCardToTarget)
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

// JudgeResult 判定完整结果（参考 noname event.result）。
// 判定流程：取牌 → 构建 result → judge 函数计算 → mod.judge 修改 → judgeFixing → callback
type JudgeResult struct {
	// 判定牌信息（参考 noname: card / name / number / suit / color）
	Card   CardView `json:"card"`   // 判定牌
	Number int      `json:"number"` // 点数（三国杀归一化：A=1, J=11, Q=12, K=13）
	Suit   string   `json:"suit"`   // 花色: S/H/C/D (spade/heart/club/diamond)
	Color  string   `json:"color"`  // 颜色: red / black
	// 判定结果
	Judge int  `json:"judge"` // 判定函数返回值: >0 成功, <0 失败, 0 无结果
	Bool  *bool `json:"bool"` // true=成功, false=失败, nil=无结果
	// 来源信息
	Seat   int         `json:"seat"`   // 判定目标座位号
	Reason JudgeReason `json:"reason"` // 判定来源
}

// BoolPtr 辅助函数：创建 bool 指针。
func BoolPtr(b bool) *bool { return &b }

// JudgeFunc 判定函数类型（参考 noname: judge(card) → number）。
// 返回 >0 成功，<0 失败，0 无结果。
type JudgeFunc func(card CardView) int

// ModJudgeCtx mod.judge 被动修改上下文（参考 noname: mod.judge(player, result)）。
// 技能可在此修改判定结果的任意字段（suit/number/color/bool）。
type ModJudgeCtx struct {
	Seat   int         // 判定目标座位号
	Reason JudgeReason // 判定来源
	Result *JudgeResult // 可修改的判定结果
}

// JudgeCtx 翻判定牌后广播（旧接口，保留向后兼容）。
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
	Seat         int
	Card         CardView
	OriginalKind string // 牌使用前的原始类型（龙胆变牌检测用）
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

// ShaCtx 杀流程上下文（noname: shaBegin / shaMiss / shaHit）。
type ShaCtx struct {
	Source int      // 出杀者
	Target int      // 目标
	Card   CardView // 杀牌
	Damage int      // 伤害值
}

// UseCardCtx 使用牌上下文（noname: useCard / useCardToTarget）。
type UseCardCtx struct {
	Seat   int      // 使用者
	Target int      // 目标
	Card   CardView // 使用的牌
}

// HookCall 单次 hook 调用的参数；按 Kind 填对应字段。
type HookCall struct {
	Kind HookKind
	Role HookRole // 技能角色维度（默认 RolePlayer）

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
	ModJudge       *ModJudgeCtx
	ShaCtx         *ShaCtx      // 杀流程上下文
	UseCard        *UseCardCtx  // 使用牌上下文
	BecomeTarget   *BecomeTargetCtx // 成为目标上下文
	Discarded      *CardsDiscardedCtx
	EquipLost      *EquipLostCtx
}

// HookResult 聚合型 hook 的返回值。
type HookResult struct {
	Bool bool
	Int  int
	Err  error
}
