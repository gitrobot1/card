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
	if p.TotalCards() != 64 {
		t.Fatalf("legacy total=%d want 64", p.TotalCards())
	}
}

func TestDeckProfileFor_3v3NoShanDian(t *testing.T) {
	p := mode.DeckProfileFor(mode.Solo3v3)
	if p.HasKind(mode.DeckKindShanDian) {
		t.Fatal("3v3 deck must not include shandian")
	}
	if p.TotalCards() != 63 {
		t.Fatalf("3v3 total=%d want 63", p.TotalCards())
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
	if p.TotalCards() != 90 {
		t.Fatalf("identity_8 total=%d want 90", p.TotalCards())
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
	if p.CountKind(mode.DeckKindSha) != 13 {
		t.Fatalf("ddz sha=%d want 13", p.CountKind(mode.DeckKindSha))
	}
	if p.TotalCards() != 67 {
		t.Fatalf("ddz total=%d want 67", p.TotalCards())
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
	if p.CountKind(mode.DeckKindSha) != 12 {
		t.Fatalf("identity_5 sha=%d want 12", p.CountKind(mode.DeckKindSha))
	}
	if p.CountKind(mode.DeckKindTao) != 5 {
		t.Fatalf("identity_5 tao=%d want 5", p.CountKind(mode.DeckKindTao))
	}
	if p.TotalCards() != 67 {
		t.Fatalf("identity_5 total=%d want 67", p.TotalCards())
	}
	if p.ID != mode.DeckProfileIdentity5 {
		t.Fatalf("profile id=%q want %q", p.ID, mode.DeckProfileIdentity5)
	}
}

func TestDeckProfileFor_UnknownUsesLegacy(t *testing.T) {
	p := mode.DeckProfileFor("not-a-mode")
	if p.TotalCards() != 64 {
		t.Fatalf("unknown mode total=%d want 64", p.TotalCards())
	}
}
