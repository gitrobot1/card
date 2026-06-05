package mode

// 牌种类（与 engine Card* 常量字符串一致，mode 包不依赖 engine）
const (
	DeckKindSha       = "sha"
	DeckKindShan      = "shan"
	DeckKindWuxiek    = "wuxiek"
	DeckKindTao       = "tao"
	DeckKindJiu       = "jiu"
	DeckKindGuohe     = "guohe"
	DeckKindTanNang   = "tannang"
	DeckKindWuZhong   = "wuzhong"
	DeckKindNanMan    = "nanman"
	DeckKindWanJian   = "wanjian"
	DeckKindJueDou    = "juedou"
	DeckKindLeBu      = "lebu"
	DeckKindBingLiang = "bingliang"
	DeckKindShanDian  = "shandian"
	DeckKindWuGu      = "wugu"
	DeckKindTaoYuan   = "taoyuan"
	DeckKindWeapon1   = "weapon_1"
	DeckKindWeapon2   = "weapon_2"
	DeckKindWeapon3   = "weapon_3"
	DeckKindWeapon4   = "weapon_4"
	DeckKindWeapon5   = "weapon_5"
	DeckKindWeapon6   = "weapon_6"
	DeckKindArmor     = "armor"
	DeckKindArmorVine = "armor_vine"
	DeckKindHuoGong   = "huogong"
	DeckKindTieSuo    = "tiesuo"
	DeckKindPlusHorse = "plus_horse"
	DeckKindMinusHorse = "minus_horse"
)

const (
	DeckProfileLegacy    = "legacy"
	DeckProfileComp3v3   = "comp_3v3"
	DeckProfileIdentity5 = "identity_5p"
	DeckProfileIdentity8 = "identity_8p"
	DeckProfileDdz3p     = "ddz_3p"
)

// DeckSpec 描述一种牌的数量。
type DeckSpec struct {
	Kind  string
	Count int
}

// DeckProfile 某模式的牌堆与发牌参数。
type DeckProfile struct {
	ID              string
	Specs           []DeckSpec
	InitialHandSize int
}

// BasicDeckSpecs 当前全员默认牌堆（64 张，含 1 闪电）。
func BasicDeckSpecs() []DeckSpec {
	return []DeckSpec{
		{DeckKindSha, 10},
		{DeckKindShan, 4},
		{DeckKindWuxiek, 3},
		{DeckKindTao, 4},
		{DeckKindJiu, 3},
		{DeckKindGuohe, 2},
		{DeckKindTanNang, 2},
		{DeckKindWuZhong, 2},
		{DeckKindNanMan, 2},
		{DeckKindWanJian, 2},
		{DeckKindJueDou, 2},
		{DeckKindLeBu, 2},
		{DeckKindBingLiang, 2},
		{DeckKindShanDian, 1},
		{DeckKindWuGu, 2},
		{DeckKindTaoYuan, 2},
		{DeckKindWeapon1, 1},
		{DeckKindWeapon2, 1},
		{DeckKindWeapon3, 1},
		{DeckKindWeapon4, 1},
		{DeckKindWeapon5, 1},
		{DeckKindWeapon6, 1},
		{DeckKindArmor, 3},
		{DeckKindArmorVine, 2},
		{DeckKindHuoGong, 2},
		{DeckKindTieSuo, 2},
		{DeckKindPlusHorse, 2},
		{DeckKindMinusHorse, 2},
	}
}

// adjustDeckSpecs 在 base 上按 kind 增减张数（只改已存在的 kind）。
func adjustDeckSpecs(base []DeckSpec, delta map[string]int) []DeckSpec {
	out := make([]DeckSpec, len(base))
	copy(out, base)
	for i, s := range out {
		if d, ok := delta[s.Kind]; ok {
			out[i].Count += d
		}
	}
	return out
}

// Ddz3pDeckSpecs 三人斗地主：legacy +3 杀，偏进攻节奏。
func Ddz3pDeckSpecs() []DeckSpec {
	return adjustDeckSpecs(BasicDeckSpecs(), map[string]int{
		DeckKindSha: 3,
	})
}

// Identity5DeckSpecs 五人身份局微调：+2 杀 +1 桃，保留闪电。
func Identity5DeckSpecs() []DeckSpec {
	return adjustDeckSpecs(BasicDeckSpecs(), map[string]int{
		DeckKindSha: 2,
		DeckKindTao: 1,
	})
}

// Identity8DeckSpecs 八人身份局专用牌堆（90 张，无闪电；较 legacy 增基础牌与锦囊以支撑 8 人消耗）。
func Identity8DeckSpecs() []DeckSpec {
	return []DeckSpec{
		{DeckKindSha, 18},
		{DeckKindShan, 10},
		{DeckKindWuxiek, 4},
		{DeckKindTao, 7},
		{DeckKindJiu, 4},
		{DeckKindGuohe, 3},
		{DeckKindTanNang, 3},
		{DeckKindWuZhong, 2},
		{DeckKindNanMan, 3},
		{DeckKindWanJian, 3},
		{DeckKindJueDou, 2},
		{DeckKindLeBu, 3},
		{DeckKindBingLiang, 3},
		{DeckKindWuGu, 2},
		{DeckKindTaoYuan, 2},
		{DeckKindWeapon1, 1},
		{DeckKindWeapon2, 1},
		{DeckKindWeapon3, 1},
		{DeckKindWeapon4, 1},
		{DeckKindWeapon5, 1},
		{DeckKindWeapon6, 1},
		{DeckKindArmor, 3},
		{DeckKindArmorVine, 2},
		{DeckKindHuoGong, 2},
		{DeckKindTieSuo, 2},
		{DeckKindPlusHorse, 3},
		{DeckKindMinusHorse, 3},
	}
}

func specsWithoutKinds(specs []DeckSpec, omit ...string) []DeckSpec {
	ban := make(map[string]bool, len(omit))
	for _, k := range omit {
		ban[k] = true
	}
	out := make([]DeckSpec, 0, len(specs))
	for _, s := range specs {
		if ban[s.Kind] {
			continue
		}
		out = append(out, s)
	}
	return out
}

var deckProfiles = map[string]DeckProfile{
	DeckProfileLegacy: {
		ID:              DeckProfileLegacy,
		Specs:           BasicDeckSpecs(),
		InitialHandSize: 4,
	},
	DeckProfileComp3v3: {
		ID:              DeckProfileComp3v3,
		Specs:           specsWithoutKinds(BasicDeckSpecs(), DeckKindShanDian),
		InitialHandSize: 4,
	},
	DeckProfileDdz3p: {
		ID:              DeckProfileDdz3p,
		Specs:           Ddz3pDeckSpecs(),
		InitialHandSize: 4,
	},
	DeckProfileIdentity5: {
		ID:              DeckProfileIdentity5,
		Specs:           Identity5DeckSpecs(),
		InitialHandSize: 4,
	},
	DeckProfileIdentity8: {
		ID:              DeckProfileIdentity8,
		Specs:           Identity8DeckSpecs(),
		InitialHandSize: 4,
	},
}

// DeckProfileFor 返回模式对应的牌堆配置；未知模式用 legacy（与改前 NewBasicDeck 一致）。
func DeckProfileFor(modeID string) DeckProfile {
	switch NormalizeID(modeID) {
	case Solo3v3:
		return deckProfiles[DeckProfileComp3v3]
	case Solo3pDdz:
		return deckProfiles[DeckProfileDdz3p]
	case SoloIdentity5:
		return deckProfiles[DeckProfileIdentity5]
	case SoloIdentity8:
		return deckProfiles[DeckProfileIdentity8]
	default:
		return deckProfiles[DeckProfileLegacy]
	}
}

// TotalCards 牌堆总张数。
func (p DeckProfile) TotalCards() int {
	n := 0
	for _, s := range p.Specs {
		n += s.Count
	}
	return n
}

// HasKind 牌堆是否含某类牌。
func (p DeckProfile) HasKind(kind string) bool {
	return p.CountKind(kind) > 0
}

// CountKind 某类牌的张数。
func (p DeckProfile) CountKind(kind string) int {
	n := 0
	for _, s := range p.Specs {
		if s.Kind == kind {
			n += s.Count
		}
	}
	return n
}
