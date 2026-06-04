package skill

// IsRedSuit 红桃 H、方块 D 为红色。
func IsRedSuit(suit string) bool {
	return suit == "H" || suit == "D"
}

// IsBlackSuit 黑桃 S、梅花 C 为黑色。
func IsBlackSuit(suit string) bool {
	return suit == "S" || suit == "C"
}
