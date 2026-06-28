package mode

import "sort"

// SeatSlot describes a non-human seat position for client layout.
type SeatSlot struct {
	Seat       int    `json:"seat"`
	Area       string `json:"area"`
	Placement  string `json:"placement"`
	IsTeammate bool   `json:"is_teammate"`
	SeatRole   string `json:"seat_role,omitempty"` // chain: protect | mark
}

// HeroPoolSpec restricts which heroes are selectable in a mode.
type HeroPoolSpec struct {
	Packs    []string `json:"packs,omitempty"`
	Kingdoms []string `json:"kingdoms,omitempty"`
}

// Meta describes a playable mode for API / UI (no engine types).
type Meta struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Tag         string       `json:"tag,omitempty"`
	Description string       `json:"description"`
	Hint        string       `json:"hint,omitempty"`
	Subtitle    string       `json:"subtitle,omitempty"`
	LayoutKey   string       `json:"layout_key"`
	Tags        []string     `json:"tags"`
	Rules       []string     `json:"rules,omitempty"`
	PlayerCount int          `json:"player_count"`
	HumanSeats  []int        `json:"human_seats"`
	SeatMap     []SeatSlot   `json:"seat_map,omitempty"`
	HeroPool    HeroPoolSpec `json:"hero_pool"`
}

const (
	LayoutSolo1v1    = "solo_1v1"
	LayoutCross2v2   = "cross_2v2"
	LayoutTriangle3p = "triangle_3p"
	LayoutHex3v3     = "hex_3v3"
	LayoutPentagon5  = "pentagon_5"
	LayoutOctagon8   = "octagon_8"
)

// DefaultHeroPool 各模式默认可选武将扩展包。
var DefaultHeroPool = HeroPoolSpec{Packs: []string{"standard", "sp", "shen"}}

var registry = map[string]Meta{}

// Register adds a mode definition. Panics on duplicate ID.
func Register(meta Meta) {
	if meta.ID == "" {
		panic("mode: register without id")
	}
	if _, exists := registry[meta.ID]; exists {
		panic("mode: duplicate id " + meta.ID)
	}
	registry[meta.ID] = meta
}

// Lookup returns mode metadata by id.
func Lookup(id string) (Meta, bool) {
	m, ok := registry[NormalizeID(id)]
	return m, ok
}

// NormalizeID maps client mode strings to canonical IDs; unknown values default to 1v1.
func NormalizeID(id string) string {
	switch id {
	case "", Solo1v1, "1V1", "solo_1v1":
		return Solo1v1
	case Solo2v2, "2V2", "2V二", LayoutCross2v2:
		return Solo2v2
	case Solo3pChain, "3p", "3P", "杀上保下", LayoutTriangle3p:
		return Solo3pChain
	case Solo3pDdz, "斗地主":
		return Solo3pDdz
	case Solo3v3, "3V3", "3v3竞技", LayoutHex3v3:
		return Solo3v3
	case SoloIdentity5, "identity", "身份局", LayoutPentagon5:
		return SoloIdentity5
	case SoloIdentity8, "8人身份局", LayoutOctagon8:
		return SoloIdentity8
	default:
		if _, ok := registry[id]; ok {
			return id
		}
		return Solo1v1
	}
}

// All returns registered modes sorted by id.
func All() []Meta {
	out := make([]Meta, 0, len(registry))
	for _, m := range registry {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func init() {
	Register(Meta{
		ID:          Solo1v1,
		Name:        "对战电脑",
		Tag:         "单机 1v1",
		Description: "1v1 人机 · 基础规则 · 验证出牌与回合",
		Hint:        "4 体力 · 蜀将三选一 · 电脑随机对手",
		Subtitle:    "1v1 单机 · 基础牌验证",
		LayoutKey:   LayoutSolo1v1,
		Tags:        []string{"solo", "1v1"},
		Rules: []string{
			"刘备：仁德、激将",
			"关羽：武圣（闪当杀）",
			"张飞：咆哮（无限杀）",
		},
		PlayerCount: 2,
		HumanSeats:  []int{0},
		HeroPool:    DefaultHeroPool,
	})
	Register(Meta{
		ID:          Solo2v2,
		Name:        "十字阵对战",
		Tag:         "2v2 人机",
		Description: "2v2 人机 · 十字阵 · 1 真人 + 3 电脑",
		Hint:        "你（下）+ 队友（上） vs 两侧敌将 · 1 真人 + 3 电脑",
		Subtitle:    "2v2 十字阵 · 1 真人 + 3 电脑",
		LayoutKey:   LayoutCross2v2,
		Tags:        []string{"solo", "2v2", "team"},
		Rules: []string{
			"出牌顺序：你 → 敌左 → 队友 → 敌右",
			"队友坐对家，不可误伤",
			"消灭敌方两人即获胜",
		},
		PlayerCount: 4,
		HumanSeats:  []int{0},
		SeatMap: []SeatSlot{
			{Seat: 2, Area: "top", Placement: "top", IsTeammate: true},
			{Seat: 1, Area: "left", Placement: "left", IsTeammate: false},
			{Seat: 3, Area: "right", Placement: "right", IsTeammate: false},
		},
		HeroPool: DefaultHeroPool,
	})
	Register(Meta{
		ID:          Solo3pChain,
		Name:        "杀上保下",
		Tag:         "三人链式",
		Description: "3 人人机 · 杀上家、保下家",
		Hint:        "仅可攻击上家 · 下家阵亡即失败 · 上家阵亡即胜利",
		Subtitle:    "3 人链式 · 1 真人 + 2 电脑",
		LayoutKey:   LayoutTriangle3p,
		Tags:        []string{"solo", "3p", "chain"},
		Rules: []string{
			"回合顺序：你 → 下家（左）→ 上家（右）",
			"只能对上家使用杀和攻击性锦囊",
			"下家或自己阵亡 = 失败；上家阵亡 = 胜利",
		},
		PlayerCount: 3,
		HumanSeats:  []int{0},
		SeatMap: []SeatSlot{
			{Seat: 1, Area: "left", Placement: "left", IsTeammate: true, SeatRole: "protect"},
			{Seat: 2, Area: "right", Placement: "right", SeatRole: "mark"},
		},
		HeroPool: DefaultHeroPool,
	})
	Register(Meta{
		ID:          Solo3pDdz,
		Name:        "斗地主",
		Tag:         "三人地主",
		Description: "3 人人机 · 地主 vs 两农民",
		Hint:        "你是地主 · 每回合多摸一张、可额外出杀 · 判定时可弃两张取消",
		Subtitle:    "3 人斗地主 · 1 真人 + 2 电脑",
		LayoutKey:   LayoutTriangle3p,
		Tags:        []string{"solo", "3p", "team", "ddz"},
		Rules: []string{
			"你担任地主，对抗两名农民",
			"地主每回合额外摸 1 张牌",
			"地主每回合可额外出 1 张【杀】",
			"地主判定时可弃 2 张手牌，取消此次判定",
			"消灭对方阵营即获胜",
		},
		PlayerCount: 3,
		HumanSeats:  []int{0},
		SeatMap: []SeatSlot{
			{Seat: 1, Area: "left", Placement: "left", IsTeammate: false, SeatRole: "farmer"},
			{Seat: 2, Area: "right", Placement: "right", IsTeammate: false, SeatRole: "farmer"},
		},
		HeroPool: DefaultHeroPool,
	})
	Register(Meta{
		ID:          Solo3v3,
		Name:        "3v3 竞技",
		Tag:         "3v3 团队",
		Description: "6 人人机 · 参考三国杀 3v3 · 消灭敌方主帅获胜",
		Hint:        "你担任暖色主帅 · 两名 AI 前锋 · 对战冷色三人",
		Subtitle:    "3v3 竞技 · 1 真人 + 5 电脑",
		LayoutKey:   LayoutHex3v3,
		Tags:        []string{"solo", "3v3", "team"},
		Rules: []string{
			"暖色（你+两前锋）vs 冷色（主帅+两前锋）",
			"击败敌方主帅即获胜；前锋阵亡对局继续",
			"杀死任意角色可摸 3 张牌",
			"本模式不使用【闪电】与主公技",
		},
		PlayerCount: 6,
		HumanSeats:  []int{0},
		SeatMap: []SeatSlot{
			{Seat: 1, Area: "left-top", Placement: "left", IsTeammate: false, SeatRole: "forward"},
			{Seat: 2, Area: "top", Placement: "top", IsTeammate: false, SeatRole: "commander"},
			{Seat: 3, Area: "right-top", Placement: "right", IsTeammate: false, SeatRole: "forward"},
			{Seat: 4, Area: "left", Placement: "left", IsTeammate: true, SeatRole: "forward"},
			{Seat: 5, Area: "right", Placement: "right", IsTeammate: true, SeatRole: "forward"},
		},
		HeroPool: DefaultHeroPool,
	})
	Register(Meta{
		ID:          SoloIdentity5,
		Name:        "身份局",
		Tag:         "5 人身份",
		Description: "5 人人机 · 标准五人身份场",
		Hint:        "你担任主公（公开）· 消灭反贼与内奸获胜 · 与内奸单挑时主公阵亡则内奸胜 · 内奸独自存活亦内奸胜",
		Subtitle:    "5 人身份 · 1 真人 + 4 电脑",
		LayoutKey:   LayoutPentagon5,
		Tags:        []string{"solo", "identity", "5p"},
		Rules: []string{
			"1 主公（+1 体力）+ 1 忠臣 + 1 内奸 + 2 反贼",
			"主公身份公开，其余身份隐藏至阵亡",
			"可攻击除自己外任意角色",
			"主公阵亡 → 反贼胜（与内奸单挑时主公阵亡 → 内奸胜）；反贼与内奸全灭 → 主公阵营胜；仅剩内奸 → 内奸胜",
			"本模式不使用【闪电】",
		},
		PlayerCount: 5,
		HumanSeats:  []int{0},
		SeatMap: []SeatSlot{
			{Seat: 1, Area: "top-1", Placement: "top"},
			{Seat: 2, Area: "top-2", Placement: "top"},
			{Seat: 3, Area: "left", Placement: "left"},
			{Seat: 4, Area: "right-top", Placement: "right"},
		},
		HeroPool: DefaultHeroPool,
	})
	Register(Meta{
		ID:          SoloIdentity8,
		Name:        "八人身份局",
		Tag:         "8 人身份",
		Description: "8 人人机 · 标准八人身份场",
		Hint:        "你担任主公（公开）· 消灭反贼与内奸获胜 · 与内奸单挑时主公阵亡则内奸胜 · 内奸独自存活亦内奸胜",
		Subtitle:    "8 人身份 · 1 真人 + 7 电脑",
		LayoutKey:   LayoutOctagon8,
		Tags:        []string{"solo", "identity", "8p"},
		Rules: []string{
			"1 主公（+1 体力）+ 2 忠臣 + 1 内奸 + 4 反贼",
			"主公身份公开，其余身份隐藏至阵亡",
			"可攻击除自己外任意角色",
			"主公阵亡 → 反贼胜（与内奸单挑时主公阵亡 → 内奸胜）；反贼与内奸全灭 → 主公阵营胜；仅剩内奸 → 内奸胜",
			"本模式不使用【闪电】",
		},
		PlayerCount: 8,
		HumanSeats:  []int{0},
		SeatMap: []SeatSlot{
			{Seat: 1, Area: "top-1", Placement: "top"},
			{Seat: 2, Area: "top-2", Placement: "top"},
			{Seat: 3, Area: "top-3", Placement: "top"},
			{Seat: 4, Area: "top-4", Placement: "top"},
			{Seat: 5, Area: "top-5", Placement: "top"},
			{Seat: 6, Area: "left", Placement: "left"},
			{Seat: 7, Area: "right", Placement: "right"},
		},
		HeroPool: DefaultHeroPool,
	})
}
