package skill

// Default accent colors when hero JSON omits accent_color.
var kingdomAccent = map[string]string{
	KingdomShu: "#c45c3e",
	KingdomWei: "#4a5568",
	KingdomWu:  "#2b6cb0",
	KingdomQun: "#6b46c1",
}

// Per-hero accent overrides (from legacy frontend palette).
var heroAccent = map[string]string{
	CharLiuBei:          "#c45c3e",
	CharGuanYu:          "#3d7a52",
	CharZhaoYun:         "#5b6eae",
	CharCaoCao:          "#4a5568",
	CharXiahouDun:       "#8b4513",
	CharXuChu:           "#c53030",
	CharZhangLiao:       "#2b6cb0",
	CharGuoJia:          "#4a5568",
	CharSimaYi:          "#2c5282",
	CharZhenJi:          "#805ad5",
	CharSunQuan:         "#2b6cb0",
	CharSunShangxiang:   "#d53f8c",
	CharZhouYu:          "#805ad5",
	CharXiaoQiao:        "#ed64a6",
	CharGanNing:         "#2c5282",
	CharSunJian:         "#c05621",
	CharLuXun:           "#38a169",
	CharDaQiao:          "#d53f8c",
	CharHuangGai:        "#c05621",
	CharLvMeng:          "#2c5282",
	CharHuaTuo:          "#38a169",
	CharLvBu:            "#c53030",
	CharDiaoChan:        "#d53f8c",
	CharYanLiangWenChou: "#9b2c2c",
	CharJiaXu:           "#4a5568",
	CharZhangJiao:       "#6b46c1",
	CharZhangChunhua:    "#805ad5",
}

func ResolveAccentColor(def CharacterDef) string {
	if def.AccentColor != "" {
		return def.AccentColor
	}
	if c, ok := heroAccent[def.ID]; ok {
		return c
	}
	if c, ok := kingdomAccent[def.Kingdom]; ok {
		return c
	}
	return "#6b7280"
}
