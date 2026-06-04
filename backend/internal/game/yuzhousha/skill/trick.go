package skill

// IsJinnangKind 是否为锦囊牌（含延时与非延时）。
func IsJinnangKind(kind string) bool {
	switch kind {
	case "guohe", "tannang", "nanman", "wanjian", "juedou",
		"lebu", "bingliang", "shandian", "wugu", "taoyuan", "wuzhong", "wuxiek":
		return true
	default:
		return false
	}
}

// IsJiangCard 【激昂】触发的牌：【决斗】或红色【杀】。
func IsJiangCard(kind, suit string) bool {
	if kind == "juedou" {
		return true
	}
	return kind == "sha" && IsRedSuit(suit)
}
