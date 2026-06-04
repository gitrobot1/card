package skill

import "testing"

func TestEnsureDefaultSkinsForAllHeroes(t *testing.T) {
	for _, def := range PickableCharacters() {
		skin, ok := DefaultSkinForHero(def.ID)
		if !ok {
			t.Fatalf("hero %s missing default skin", def.ID)
		}
		if skin.HeroID != def.ID {
			t.Fatalf("skin %s hero mismatch: got %s want %s", skin.ID, skin.HeroID, def.ID)
		}
		display := ResolveHeroDisplay(def.ID, "")
		if display.SkinID == "" || display.AccentColor == "" {
			t.Fatalf("hero %s unresolved display: %+v", def.ID, display)
		}
	}
}

func TestStandardPackManifestRegistered(t *testing.T) {
	m, ok := PackByID("standard")
	if !ok {
		t.Fatal("standard pack not registered")
	}
	if m.HeroPack != "standard" || m.SkinPack != "standard" {
		t.Fatalf("unexpected pack links: %+v", m)
	}
}
