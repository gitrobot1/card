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
	SkillIDs    []string `json:"skill_ids"`
	Pack        string   `json:"pack,omitempty"`
	AccentColor string   `json:"accent_color,omitempty"`
	PortraitURL string   `json:"portrait_url,omitempty"`
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
	CardPlaysAs(seat int, cardKind, asKind, suit string) bool
	HandPlaysAs(seat int, asKind string) bool
	AlivePlayerCount() int
	DrawPileCount() int
	DrawCards(seat, count int) error
	DrawSkillCards(seat int, skillID string, count int, message string) error

	GiveRende(source, target int, cardIDs []string) error
	StartJijiangForUse(lord, target int) error
	StartJijiangForResponse(lord int) error
	ToggleWusheng(seat int) error
	StartPeekDeck(seat int, skillID string) error
	ApplyTieqi(seat int) error
	SkipTieqi(seat int) error
	PendingTieqiForSource(seat int) bool
	FankuiTakeFrom(seat int, zone, cardID string) error
	PassFankui(seat int) error
	PendingFankuiFor(seat int) bool
	FankuiSourceSeat(actor int) int
	FirstTakeableCardID(target int) string
	ApplyGuicaiReplace(seat int, handCardID string) error
	PassGuicai(seat int) error
	PendingGuicaiFor(seat int) bool
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
	StartTuxi(seat, skipCount int) error
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
	HasBlackHandCard(seat int) bool
	OpponentHasHandCard(seat int) bool
	ActivateQixi(seat int, cardID string) error
	QixiTakeFrom(seat int, cardID string) error
	ActivateYinghun(seat, target int) error
	ResolveYinghunChoice(seat int, option string) error
	YinghunDiscard(seat int, cardID string) error
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
}

// Decl 声明式技能：按需填字段，未填则使用默认零行为。
type Decl struct {
	Meta Meta

	PreparePhase PreparePhaseDecl
	PeekDeck     *PeekDeckConfig

	CanActivate  func(r Runtime, seat int) bool
	Activate     func(r Runtime, seat int, req ActivateReq) error
	OnTrigger    func(r Runtime, trigger Trigger, seat int) (handled bool, err error)
	CardPlaysAs   func(r Runtime, seat int, cardKind, asKind, suit string) bool
	UnlimitedSha  func(r Runtime, seat int) bool
	BlocksTarget       func(r Runtime, target int, cardKind string) bool
	DistanceDelta      func(r Runtime, from, to int) int
	TrickIgnoresDistance func(r Runtime, seat int, trickKind string) bool
	OnInstantTrickUsed func(r Runtime, seat int, trickKind string) error
	OnDamageDealt      func(r Runtime, ctx DamageCtx) error
	OnJudgeResult      func(r Runtime, ctx JudgeCtx) error
	OnCardsDiscarded   func(r Runtime, ctx CardsDiscardedCtx) error
	OnEquipLost        func(r Runtime, ctx EquipLostCtx) error
	DrawCountBonus     func(r Runtime, seat int) int
	OnTurnEnd          func(r Runtime, seat int) error
	OnHandEmpty        func(r Runtime, seat int) error
	EffectiveSuit      func(r Runtime, seat int, suit string) string
	BlocksTrickTarget  func(r Runtime, target int, trickKind, suit string) bool
	BlocksPeachUse     func(r Runtime, userSeat int) bool
	DamageAsHPLoss     func(r Runtime, source int) bool
	ExtraResponsesNeeded func(r Runtime, source int, cardKind string) int
	SkipsDiscardPhase  func(r Runtime, seat int) bool
	OnCardResolved     func(r Runtime, ctx CardResolvedCtx) error
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

func (h Handler) OnJudgeResult(r Runtime, ctx JudgeCtx) error {
	if h.Decl.OnJudgeResult == nil {
		return nil
	}
	return h.Decl.OnJudgeResult(r, ctx)
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
