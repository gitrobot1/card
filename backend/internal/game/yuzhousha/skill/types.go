package skill

// Kind 技能分类。
type Kind string

const (
	KindPassive   Kind = "passive"
	KindActive    Kind = "active"
	KindLord      Kind = "lord"
	KindAwakening Kind = "awakening"
	KindLimited   Kind = "limited"
)

// SkillTag 技能行为标记（参考 noname: forced/limited/awaken/lord/equipSkill）。
// 用于控制技能的触发方式、生命周期和 UI 展示。
type SkillTag string

const (
	TagForced     SkillTag = "forced"      // 锁定技：自动触发不询问（noname: forced）
	TagLimited    SkillTag = "limited"     // 限定技：一局一次（noname: limited）
	TagAwaken     SkillTag = "awaken"      // 觉醒技：条件满足自动觉醒（noname: awaken）
	TagLord       SkillTag = "lord"        // 主公技（noname: lord）
	TagEquipSkill SkillTag = "equipSkill"  // 装备附带技能：卸下装备时自动移除（noname: equipSkill）
)

// Meta 对外公开的技能元数据。
type Meta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	Kind Kind   `json:"kind"`
	// InactiveIn1v1 主公技等在 1v1 中不生效，选将时仍展示但标记为不可用。
	InactiveIn1v1 bool `json:"inactive_in_1v1,omitempty"`
}

// CharacterDef 武将静态定义（仅数据，无逻辑）。
type CharacterDef struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	MaxHP       int      `json:"max_hp"`
	Kingdom     string   `json:"kingdom"`
	Gender      string   `json:"gender,omitempty"` // male | female
	SkillIDs    []string `json:"skill_ids"`
	Pack        string   `json:"pack,omitempty"`
	AccentColor string   `json:"accent_color,omitempty"`
	PortraitURL string   `json:"portrait_url,omitempty"`
}

// ViewAsConfig 变牌配置（参考 noname viewAs 机制）。
// 声明一个技能如何将牌"视为"另一种牌使用或打出。
// 系统不知道你有什么技能，只需告诉玩家"需要出什么牌"，
// 玩家的技能通过 ViewAs 自动把符合条件的牌标记为可用。
type ViewAsConfig struct {
	AsKind     string   `json:"as_kind"`     // 视为的牌类型（CardSha/CardShan/CardTao 等）
	SelectCard int      `json:"select_card"`  // 需要选几张牌（默认1，丈八蛇矛=2）
	Position   string   `json:"position"`     // 可选牌位置："h"=仅手牌, "he"=手牌+装备, "e"=仅装备
	Prompt     string   `json:"prompt"`       // UI 提示文本
	// FilterCard 哪些牌可选（返回 true 表示可选）。nil 表示所有牌可选。
	// 复杂过滤逻辑用此函数；简单过滤用 FilterSuits/FilterSuitColor/FilterKinds 声明即可。
	FilterCard func(r Runtime, seat int, card CardView) bool
	// ViewAsFilter 是否有可用的牌（过滤前检查）。返回 false 时此技能不显示。nil 表示始终可用。
	ViewAsFilter func(r Runtime, seat int) bool
	// OnResolve 选完牌后的处理逻辑。返回处理后的牌信息。
	OnResolve func(r Runtime, seat int, cardIDs []string, asKind string) (CardView, error)
	// IsActive 技能是否处于激活状态（如武圣已点击激活）。nil 表示不需要激活（始终生效，如龙胆）。
	IsActive func(r Runtime, seat int) bool

	// ===== 声明式过滤条件（用于前端渲染，无需硬编码技能名）=====
	// FilterSuits 过滤花色：为空表示不限花色（如 ["H","D"] 表示红色）
	FilterSuits []string
	// FilterSuitColor 过滤花色颜色："red"=红色, "black"=黑色, ""=不限
	FilterSuitColor string
	// FilterKinds 过滤牌类型：为空表示不限类型（如 ["shan"] 表示闪当杀，龙胆）
	FilterKinds []string
	// Passive 被动技能（始终生效，不需要激活）
	Passive bool
}

// ViewAsSkillInfo 前端渲染变牌技能所需的信息。
// 前端不需要知道具体技能名，只需比对 filter 条件来决定哪些牌可选。
type ViewAsSkillInfo struct {
	SkillID    string   `json:"skill_id"`
	SkillName  string   `json:"skill_name"`
	AsKind     string   `json:"as_kind"`
	SelectCard int      `json:"select_card"`
	Position   string   `json:"position"`
	Prompt     string   `json:"prompt"`
	IsActive   bool     `json:"is_active"`
	// FilterSuits 过滤花色：为空表示不限花色，非空表示仅限这些花色（如 ["H","D"] 表示红色）
	FilterSuits []string `json:"filter_suits,omitempty"`
	// FilterSuitColor 过滤花色颜色："red"=红色, "black"=黑色, ""=不限
	FilterSuitColor string `json:"filter_suit_color,omitempty"`
	// FilterKinds 过滤牌类型：为空表示不限类型（如 ["shan"] 表示闪当杀，龙胆）
	FilterKinds []string `json:"filter_kinds,omitempty"`
	// Passive 被动技能（始终生效，不需要激活）
	Passive bool `json:"passive"`
}

// ActivateReq 主动技请求。
type ActivateReq struct {
	SkillID      string
	TargetIndex  int
	CardIDs      []string
	TargetZone   string
	TargetCardID string
}

// Trigger 引擎广播时机（预留扩展）。
type Trigger string

const (
	TriggerTurnStart      Trigger = "turn_start"
	TriggerPlayPhaseStart Trigger = "play_phase_start"
)

// Runtime 技能逻辑访问对局的抽象接口；由 engine.Game 适配实现。
// 新技能优先使用 EnemiesOf/AlliesOf/DrawSkillCards 与 Decl hook 字段。
type Runtime interface {
	ModeID() string
	HasSkill(seat int, skillID string) bool
	Phase() string
	TurnStep() string
	CurrentTurn() int
	PlayerCount() int
	TeamOf(seat int) int
	PlayerHandCount(seat int) int
	PlayerHandCardIDs(seat int) []string
	PlayerHP(seat int) (hp, maxHP int)
	SkillCounter(seat int, key string) int
	OpponentOf(seat int) int
	EnemiesOf(seat int) []int
	AlliesOf(seat int) []int
	CanUseSha(seat int) bool
	CanAttack(from, to int) bool
	ShuAllies(lordSeat int) []int
	PendingRequiredKind() string
	PendingResponseMode() string
	PendingTargetSeat() int
	PendingWindowKind() string
	PendingActorSeat() int
	PendingSubjectSeat() int
	PendingOriginSeat() int
	CardPlaysAs(seat int, cardKind, asKind, suit string) bool
	HandPlaysAs(seat int, asKind string) bool
	HasBlackCard(seat int) bool
	CardSuit(seat int, cardID string) string // 返回指定手牌的花色
	AlivePlayerCount() int
	DrawPileCount() int
	DrawCards(seat, count int) error
	DrawSkillCards(seat int, skillID string, count int, message string) error

	GiveRende(source, target int, cardIDs []string) error
	StartJijiangForUse(lord, target int) error
	StartJijiangForResponse(lord int) error
	ToggleWusheng(seat int) error
	ToggleQixi(seat int) error
	StartPeekDeck(seat int, skillID string) error
	ApplyTieqi(seat int) error
	SkipTieqi(seat int) error
	PendingTieqiForSource(seat int) bool
	FankuiTakeFrom(seat int, zone, cardID string) error
	PassFankui(seat int) error
	TakeOne(actor int, zone, cardID string) error
	PassTake(actor int) error
	DiscardWindowOne(actor int, cardID string) error
	PendingFankuiFor(seat int) bool
	FankuiSourceSeat(actor int) int
	FirstTakeableCardID(target int) string
	ApplyGuicaiReplace(seat int, handCardID string) error
	PassGuicai(seat int) error
	PendingGuicaiFor(seat int) bool
	PendingJudgeReason() string
	StartLuoshen(seat int) error

	PendingJianxiongFor(seat int) bool
	ApplyJianxiong(seat int) error
	PassJianxiong(seat int) error
	PendingYijiOfferFor(seat int) bool
	PendingYijiGiveFor(seat int) bool
	ApplyYiji(seat int) error
	YijiGiveCards(seat, target int, cardIDs []string) error
	PassYijiOffer(seat int) error
	PassYijiGive(seat int) error
	PendingGanglieOfferFor(seat int) bool
	StartGanglieJudge(seat int) error
	PassGanglieOffer(seat int) error
	PendingGanglieChoiceFor(seat int) bool
	GanglieTakeDamage(seat int) error
	GanglieDiscard(seat int, cardIDs []string) error
	ActivateLuoyi(seat int) error
	PendingDrawPhaseChoiceFor(seat int) bool
	StartTuxi(seat int) error
	TuxiTakeFrom(seat int, zone, cardID string) error
	PassTuxi(seat int) error
	PendingTuxiTakeFor(seat int) bool
	TuxiSourceSeat(actor int) int
	OpponentHasTakeableCard(seat int) bool
	BestTakeTarget(target int) (zone, cardID string)
	ActivateZhiheng(seat int, cardIDs []string) error
	ActivateJieyin(seat, target int, cardIDs []string) error
	HasRedHandCard(seat int) bool
	ActivateFanjian(seat int, cardID string) error
	ResolveFanjianSuit(seat int, suit string) error
	ApplyTianxiang(seat int, cardID string) error
	PassTianxiang(seat int) error
	ActivateYinghun(seat, target int) error
	ResolveYinghunChoice(seat int, option string) error
	YinghunDiscard(seat int, cardIDs []string) error
	ActivateGuose(seat, target int, cardID string) error
	HasDiamondHandCard(seat int) bool
	ApplyLiuli(seat int, cardID string, redirect int) error
	PassLiuli(seat int) error
	PassPojun(seat int) error
	PojunPlace(seat int, zone, cardID string) error
	AutoPojunPlacing(seat int) error
	PendingPojunForSource(seat int) bool
	ActivateKurou(seat int) error
	AwakenHunzi(seat int) error
	ActivateShuangxiongDraw(seat int) error
	ActivateShuangxiongJuedou(seat int, cardID string) error
	HasShuangxiongJuedouCard(seat int) bool
	ActivateLuanwu(seat int) error
	PendingGuidaoFor(seat int) bool
	ApplyGuidaoReplace(seat int, handCardID string) error
	PassGuidao(seat int) error
	PendingLeijiOfferFor(seat int) bool
	StartLeijiJudge(seat int) error
	PassLeijiOffer(seat int) error
	IsSeatInDyingRescue(seat int) bool
	// 龙魂技能相关方法
	PlayerHandCards(seat int) []CardView
	UseLonghunCards(seat int, cardIDs []string, asKind string, useTwoCards, isRed, isBlack bool) error
	ResponseLonghunCards(seat int, cardIDs []string, asKind string, useTwoCards, isRed, isBlack bool) error
}

// Decl 声明式技能：按需填字段，未填则使用默认零行为。
type Decl struct {
	Meta Meta

	// ===== 技能标记 =====
	Tags     []SkillTag `json:"tags,omitempty"` // 技能行为标记（forced/limited/equipSkill 等）
	Priority int        `json:"priority"`       // 优先级：数字越大越先执行（默认 0）
	FirstDo  bool       `json:"firstDo"`        // 始终最先执行（如无懈可击）
	LastDo   bool       `json:"lastDo"`         // 始终最后执行

	PreparePhase PreparePhaseDecl
	PeekDeck     *PeekDeckConfig

	// ViewAs 变牌技能配置（参考 noname viewAs 机制）。
	// nil 表示不是变牌技能。非 nil 时，系统在需要出对应类型的牌时
	// 自动将此技能列为可选，玩家选中后走 OnResolve 流程。
	ViewAs *ViewAsConfig `json:"view_as,omitempty"`

	CanActivate  func(r Runtime, seat int) bool
	Activate     func(r Runtime, seat int, req ActivateReq) error
	OnTrigger    func(r Runtime, trigger Trigger, seat int) (handled bool, err error)
	CardPlaysAs   func(r Runtime, seat int, cardKind, asKind, suit string) bool
	UnlimitedSha  func(r Runtime, seat int) bool
	BlocksTarget       func(r Runtime, target int, cardKind string) bool
	DistanceDelta      func(r Runtime, from, to int) int
	TrickIgnoresDistance func(r Runtime, seat int, trickKind string) bool
	OnInstantTrickUsed func(r Runtime, seat int, trickKind string) error
	OnDamageCalculated func(r Runtime, ctx DamageCalculatedCtx) (int, error) // 伤害值计算完后（可修改）
	OnDamageDealt      func(r Runtime, ctx DamageCtx) error
	OnBeforeHPChange   func(r Runtime, ctx BeforeHPChangeCtx) (bool, error) // 扣血前（返回 true 可防止）
	OnHPLost           func(r Runtime, ctx HPLostCtx) error  // 血量流失后（非伤害）
	OnHPChanged        func(r Runtime, ctx HPChangedCtx) error // 血量变化后（伤害/流失/回复）
	OnJudgeResult      func(r Runtime, ctx JudgeCtx) error
	OnModJudge         func(r Runtime, ctx ModJudgeCtx) error // mod.judge 被动修改判定结果
	OnCardsDiscarded   func(r Runtime, ctx CardsDiscardedCtx) error
	OnEquipLost        func(r Runtime, ctx EquipLostCtx) error
	DrawCountBonus     func(r Runtime, seat int) int
	OnTurnEnd          func(r Runtime, seat int) error
	OnHandEmpty        func(r Runtime, seat int) error
	EffectiveSuit      func(r Runtime, seat int, suit string) string
	BlocksTrickTarget  func(r Runtime, target int, trickKind, suit string) bool
	BlocksPeachUse     func(r Runtime, userSeat int) bool
	DamageAsHPLoss     func(r Runtime, source int) bool
	OnLongdanActivate  func(r Runtime, seat int, target int) error // 龙胆发动时触发
	ExtraResponsesNeeded func(r Runtime, source int, cardKind string) int
	SkipsDiscardPhase  func(r Runtime, seat int) bool
	OnCardResolved     func(r Runtime, ctx CardResolvedCtx) error
	OnBecomeTarget      func(r Runtime, ctx BecomeTargetCtx) error // 成为某张牌的目标时
	OnBecomeShaTarget   func(r Runtime, ctx BecomeTargetCtx) error // 成为杀的目标后（琉璃等，仁王盾检查之后）
	OnDeath            func(r Runtime, ctx DeathCtx) error // 阵亡时（亡语，牌还在）
	OnAfterDeath       func(r Runtime, ctx DeathCtx) error // 阵亡后（牌已弃）
	BlocksWuxiek       func(r Runtime, seat int) bool // 阻止无懈可击（参考 noname: playernowuxie）

	// ===== 阶段/回合钩子 =====
	OnPhaseBeforeStart func(r Runtime, seat int) error // 回合开始前（noname: phaseBeforeStart）
	OnPhaseBeforeEnd   func(r Runtime, seat int) error // 回合开始阶段结束前（noname: phaseBeforeEnd）
	OnPhaseBeginStart  func(r Runtime, seat int) error // 回合开始(beginStart)（noname: phaseBeginStart）
	OnPhaseBegin       func(r Runtime, seat int) error // 回合正式开始（noname: phaseBegin）
	OnPhaseChange      func(r Runtime, seat int) error // 阶段切换（noname: phaseChange）
	OnPhaseEnd         func(r Runtime, seat int) error // 回合结束（noname: phaseEnd）
	OnTurnBegin        func(r Runtime, seat int) error // 回合开始（noname: turnBegin）
	// OnTurnEnd 已在上面第 217 行定义
	OnRoundStart       func(r Runtime, seat int) error // 新一轮开始（noname: roundStart）

	// ===== 杀流程钩子 =====
	OnShaBegin func(r Runtime, ctx ShaCtx) error // 杀开始结算（noname: shaBegin）
	OnShaMiss  func(r Runtime, ctx ShaCtx) error // 杀被闪（noname: shaMiss）
	OnShaHit   func(r Runtime, ctx ShaCtx) error // 杀命中（noname: shaHit）

	// ===== 伤害流程钩子 =====
	OnDamageBegin func(r Runtime, ctx DamageCtx) error // 伤害开始（noname: damageBegin）
	OnDamageEnd   func(r Runtime, ctx DamageCtx) error // 伤害结束（noname: damageEnd）

	// ===== 交互式改判 =====
	// CanModifyJudge 返回该座位是否可以进行交互式改判（鬼才/鬼道等需要询问替换牌）。
	// 返回的 skillID 用于 UI 显示技能名，返回 false 表示该座位不能改判。
	// 与 OnModJudge（被动修改）的区别：CanModifyJudge 是主动交互式改判，需要等待玩家选择替换牌。
	CanModifyJudge func(r Runtime, seat int) (canModify bool, skillID string)

	// ===== 锦囊使用钩子 =====
	OnUseCard         func(r Runtime, ctx UseCardCtx) error // 使用牌（noname: useCard）
	OnUseCardToTarget func(r Runtime, ctx UseCardCtx) error // 牌指定目标（noname: useCardToTarget）
	// 返回 true 时，该玩家使用的锦囊不可被无懈可击抵消。
	// 与 BlocksTrickTarget 的区别：BlocksTrickTarget 阻止锦囊指定目标，BlocksWuxiek 阻止别人对锦囊出无懈。
	//
	// 使用示例（将来扩展）：
	//   // 某技能使该角色使用的锦囊不可被无懈
	//   BlocksWuxiek: func(r Runtime, seat int) bool {
	//       return r.HasSkill(seat, "some_skill_id")
	//   }
	HandRetainLimit    func(r Runtime, seat int) int // 0=默认按体力；更大值提高留牌上限
	AIPriority         func(r Runtime, seat int) int
	AIActivate   func(r Runtime, seat int) error
}

// Handler 注册表中的技能实例（Decl 的包装）。
type Handler struct {
	Decl
}

func (h Handler) Meta() Meta { return h.Decl.Meta }

func (h Handler) PeekDeckConfig() *PeekDeckConfig { return h.Decl.PeekDeck }

// ViewAs 返回技能的变牌配置（nil 表示不是变牌技能）。
func (h Handler) ViewAs() *ViewAsConfig { return h.Decl.ViewAs }

// HasViewAs 检查技能是否有变牌能力。
func (h Handler) HasViewAs() bool { return h.Decl.ViewAs != nil }

func (h Handler) CanActivate(r Runtime, seat int) bool {
	if h.Decl.CanActivate == nil {
		return false
	}
	return h.Decl.CanActivate(r, seat)
}

func (h Handler) Activate(r Runtime, seat int, req ActivateReq) error {
	if h.Decl.Activate == nil {
		return ErrNotImplemented
	}
	return h.Decl.Activate(r, seat, req)
}

func (h Handler) OnTrigger(r Runtime, trigger Trigger, seat int) (bool, error) {
	if h.Decl.OnTrigger == nil {
		return false, nil
	}
	return h.Decl.OnTrigger(r, trigger, seat)
}

func (h Handler) CardPlaysAs(r Runtime, seat int, cardKind, asKind, suit string) bool {
	if h.Decl.CardPlaysAs == nil {
		return false
	}
	return h.Decl.CardPlaysAs(r, seat, cardKind, asKind, suit)
}

func (h Handler) UnlimitedSha(r Runtime, seat int) bool {
	if h.Decl.UnlimitedSha == nil {
		return false
	}
	return h.Decl.UnlimitedSha(r, seat)
}

func (h Handler) BlocksTarget(r Runtime, target int, cardKind string) bool {
	if h.Decl.BlocksTarget == nil {
	 return false
	}
	return h.Decl.BlocksTarget(r, target, cardKind)
}

func (h Handler) DistanceDelta(r Runtime, from, to int) int {
	if h.Decl.DistanceDelta == nil {
		return 0
	}
	return h.Decl.DistanceDelta(r, from, to)
}

func (h Handler) TrickIgnoresDistance(r Runtime, seat int, trickKind string) bool {
	if h.Decl.TrickIgnoresDistance == nil {
		return false
	}
	return h.Decl.TrickIgnoresDistance(r, seat, trickKind)
}

func (h Handler) OnInstantTrickUsed(r Runtime, seat int, trickKind string) error {
	if h.Decl.OnInstantTrickUsed == nil {
		return nil
	}
	return h.Decl.OnInstantTrickUsed(r, seat, trickKind)
}

func (h Handler) OnDamageDealt(r Runtime, ctx DamageCtx) error {
	if h.Decl.OnDamageDealt == nil {
		return nil
	}
	return h.Decl.OnDamageDealt(r, ctx)
}

func (h Handler) OnDamageCalculated(r Runtime, ctx DamageCalculatedCtx) (int, error) {
	if h.Decl.OnDamageCalculated == nil {
		return ctx.Amount, nil
	}
	return h.Decl.OnDamageCalculated(r, ctx)
}

func (h Handler) OnBeforeHPChange(r Runtime, ctx BeforeHPChangeCtx) (bool, error) {
	if h.Decl.OnBeforeHPChange == nil {
		return false, nil
	}
	return h.Decl.OnBeforeHPChange(r, ctx)
}

func (h Handler) OnHPLost(r Runtime, ctx HPLostCtx) error {
	if h.Decl.OnHPLost == nil {
		return nil
	}
	return h.Decl.OnHPLost(r, ctx)
}

func (h Handler) OnHPChanged(r Runtime, ctx HPChangedCtx) error {
	if h.Decl.OnHPChanged == nil {
		return nil
	}
	return h.Decl.OnHPChanged(r, ctx)
}

func (h Handler) OnJudgeResult(r Runtime, ctx JudgeCtx) error {
	if h.Decl.OnJudgeResult == nil {
		return nil
	}
	return h.Decl.OnJudgeResult(r, ctx)
}

func (h Handler) OnModJudge(r Runtime, ctx ModJudgeCtx) error {
	if h.Decl.OnModJudge == nil {
		return nil
	}
	return h.Decl.OnModJudge(r, ctx)
}

func (h Handler) OnCardsDiscarded(r Runtime, ctx CardsDiscardedCtx) error {
	if h.Decl.OnCardsDiscarded == nil {
		return nil
	}
	return h.Decl.OnCardsDiscarded(r, ctx)
}

func (h Handler) OnEquipLost(r Runtime, ctx EquipLostCtx) error {
	if h.Decl.OnEquipLost == nil {
		return nil
	}
	return h.Decl.OnEquipLost(r, ctx)
}

func (h Handler) DrawCountBonus(r Runtime, seat int) int {
	if h.Decl.DrawCountBonus == nil {
		return 0
	}
	return h.Decl.DrawCountBonus(r, seat)
}

func (h Handler) OnTurnEnd(r Runtime, seat int) error {
	if h.Decl.OnTurnEnd == nil {
		return nil
	}
	return h.Decl.OnTurnEnd(r, seat)
}

func (h Handler) OnHandEmpty(r Runtime, seat int) error {
	if h.Decl.OnHandEmpty == nil {
		return nil
	}
	return h.Decl.OnHandEmpty(r, seat)
}

func (h Handler) EffectiveSuit(r Runtime, seat int, suit string) string {
	if h.Decl.EffectiveSuit == nil {
		return suit
	}
	if s := h.Decl.EffectiveSuit(r, seat, suit); s != "" {
		return s
	}
	return suit
}

func (h Handler) BlocksTrickTarget(r Runtime, target int, trickKind, suit string) bool {
	if h.Decl.BlocksTrickTarget == nil {
		return false
	}
	return h.Decl.BlocksTrickTarget(r, target, trickKind, suit)
}

func (h Handler) BlocksPeachUse(r Runtime, userSeat int) bool {
	if h.Decl.BlocksPeachUse == nil {
		return false
	}
	return h.Decl.BlocksPeachUse(r, userSeat)
}

func (h Handler) DamageAsHPLoss(r Runtime, source int) bool {
	if h.Decl.DamageAsHPLoss == nil {
		return false
	}
	return h.Decl.DamageAsHPLoss(r, source)
}

func (h Handler) ExtraResponsesNeeded(r Runtime, source int, cardKind string) int {
	if h.Decl.ExtraResponsesNeeded == nil {
		return 0
	}
	return h.Decl.ExtraResponsesNeeded(r, source, cardKind)
}

func (h Handler) SkipsDiscardPhase(r Runtime, seat int) bool {
	if h.Decl.SkipsDiscardPhase == nil {
		return false
	}
	return h.Decl.SkipsDiscardPhase(r, seat)
}

func (h Handler) OnCardResolved(r Runtime, ctx CardResolvedCtx) error {
	if h.Decl.OnCardResolved == nil {
		return nil
	}
	return h.Decl.OnCardResolved(r, ctx)
}

func (h Handler) OnBecomeTarget(r Runtime, ctx BecomeTargetCtx) error {
	if h.Decl.OnBecomeTarget == nil {
		return nil
	}
	return h.Decl.OnBecomeTarget(r, ctx)
}

func (h Handler) OnBecomeShaTarget(r Runtime, ctx BecomeTargetCtx) error {
	if h.Decl.OnBecomeShaTarget == nil {
		return nil
	}
	return h.Decl.OnBecomeShaTarget(r, ctx)
}

func (h Handler) OnDeath(r Runtime, ctx DeathCtx) error {
	if h.Decl.OnDeath == nil {
		return nil
	}
	return h.Decl.OnDeath(r, ctx)
}

func (h Handler) OnAfterDeath(r Runtime, ctx DeathCtx) error {
	if h.Decl.OnAfterDeath == nil {
		return nil
	}
	return h.Decl.OnAfterDeath(r, ctx)
}

// BlocksWuxiek 返回 true 表示此技能阻止对该锦囊使用无懈可击（参考 noname: playernowuxie）。
func (h Handler) BlocksWuxiek(r Runtime, seat int) bool {
	if h.Decl.BlocksWuxiek == nil {
		return false
	}
	return h.Decl.BlocksWuxiek(r, seat)
}

func (h Handler) HandRetainLimit(r Runtime, seat int) int {
	if h.Decl.HandRetainLimit == nil {
		return 0
	}
	return h.Decl.HandRetainLimit(r, seat)
}

func (h Handler) AIPriority(r Runtime, seat int) int {
	if h.Decl.AIPriority == nil {
		return 0
	}
	return h.Decl.AIPriority(r, seat)
}

func (h Handler) AIActivate(r Runtime, seat int) error {
	if h.Decl.AIActivate == nil {
		return nil
	}
	return h.Decl.AIActivate(r, seat)
}
