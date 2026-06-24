package mode_test

import (
	"testing"

	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func TestDeckProfileFor_LegacyHasShanDian(t *testing.T) {
	p := mode.DeckProfileFor(mode.Solo1v1)
	if !p.HasKind(mode.DeckKindShanDian) {
		t.Fatal("legacy deck should include shandian")
	}
	if p.TotalCards() != 72 {
		t.Fatalf("legacy total=%d want 72", p.TotalCards())
	}
}

func TestDeckProfileFor_3v3NoShanDian(t *testing.T) {
	p := mode.DeckProfileFor(mode.Solo3v3)
	if p.HasKind(mode.DeckKindShanDian) {
		t.Fatal("3v3 deck must not include shandian")
	}
	if p.TotalCards() != 66 {
		t.Fatalf("3v3 total=%d want 66", p.TotalCards())
	}
	if p.ID != mode.DeckProfileComp3v3 {
		t.Fatalf("profile id=%q want %q", p.ID, mode.DeckProfileComp3v3)
	}
}

func TestDeckProfileFor_Identity8LargeDeck(t *testing.T) {
	p := mode.DeckProfileFor(mode.SoloIdentity8)
	if p.HasKind(mode.DeckKindShanDian) {
		t.Fatal("identity_8 deck must not include shandian")
	}
	if p.TotalCards() != 93 {
		t.Fatalf("identity_8 total=%d want 93", p.TotalCards())
	}
	if p.ID != mode.DeckProfileIdentity8 {
		t.Fatalf("profile id=%q want %q", p.ID, mode.DeckProfileIdentity8)
	}
	id5 := mode.DeckProfileFor(mode.SoloIdentity5)
	if p.TotalCards() <= id5.TotalCards() {
		t.Fatalf("identity_8 total=%d should exceed identity_5 %d", p.TotalCards(), id5.TotalCards())
	}
}

func TestDeckProfileFor_DdzExtraSha(t *testing.T) {
	p := mode.DeckProfileFor(mode.Solo3pDdz)
	// 普通杀 5+3=8，另加 火杀 3 雷杀 2，共 13 张杀
	if p.CountKind(mode.DeckKindSha) != 8 {
		t.Fatalf("ddz sha=%d want 8", p.CountKind(mode.DeckKindSha))
	}
	if p.CountKind(mode.DeckKindShaFire) != 3 {
		t.Fatalf("ddz sha_fire=%d want 3", p.CountKind(mode.DeckKindShaFire))
	}
	if p.TotalCards() != 70 {
		t.Fatalf("ddz total=%d want 70", p.TotalCards())
	}
	if p.ID != mode.DeckProfileDdz3p {
		t.Fatalf("profile id=%q want %q", p.ID, mode.DeckProfileDdz3p)
	}
	legacy := mode.DeckProfileFor(mode.Solo1v1)
	if p.CountKind(mode.DeckKindSha) <= legacy.CountKind(mode.DeckKindSha) {
		t.Fatal("ddz should have more sha than legacy")
	}
}

func TestDeckProfileFor_Identity5TunedKeepsShanDian(t *testing.T) {
	p := mode.DeckProfileFor(mode.SoloIdentity5)
	if !p.HasKind(mode.DeckKindShanDian) {
		t.Fatal("identity_5 should keep shandian for now")
	}
	// 普通杀 5+2=7，另加 火杀 3 雷杀 2，共 12 张杀
	if p.CountKind(mode.DeckKindSha) != 7 {
		t.Fatalf("identity_5 sha=%d want 7", p.CountKind(mode.DeckKindSha))
	}
	if p.CountKind(mode.DeckKindTao) != 5 {
		t.Fatalf("identity_5 tao=%d want 5", p.CountKind(mode.DeckKindTao))
	}
	if p.TotalCards() != 70 {
		t.Fatalf("identity_5 total=%d want 70", p.TotalCards())
	}
	if p.ID != mode.DeckProfileIdentity5 {
		t.Fatalf("profile id=%q want %q", p.ID, mode.DeckProfileIdentity5)
	}
}

func TestDeckProfileFor_UnknownUsesLegacy(t *testing.T) {
	p := mode.DeckProfileFor("not-a-mode")
	if p.TotalCards() != 67 {
		t.Fatalf("unknown mode total=%d want 67", p.TotalCards())
	}
}
