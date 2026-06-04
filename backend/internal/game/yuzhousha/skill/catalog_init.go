package skill

func init() {
	for _, d := range catalogSkills() {
		Register(d)
	}
	for _, d := range catalogPeekDeckSkills() {
		Register(d)
	}
	if err := LoadEmbeddedHeroes(); err != nil {
		panic("yuzhousha: load heroes: " + err.Error())
	}
	if err := LoadEmbeddedSkins(); err != nil {
		panic("yuzhousha: load skins: " + err.Error())
	}
	if err := LoadEmbeddedPacks(); err != nil {
		panic("yuzhousha: load packs: " + err.Error())
	}
}
