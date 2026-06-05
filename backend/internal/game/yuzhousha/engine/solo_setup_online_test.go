package engine

import (
	"testing"

	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

func TestNewOnline2v2_FourHumans(t *testing.T) {
	g, err := NewOnline2v2("t", [4]string{"A", "B", "C", "D"}, [4]string{CharLiuBei, CharGuanYu, CharZhangFei, CharZhaoYun})
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Players) != 4 {
		t.Fatalf("players = %d, want 4", len(g.Players))
	}
	for i, p := range g.Players {
		if p.IsAI {
			t.Fatalf("seat %d should be human", i)
		}
	}
	if g.Mode != Mode2v2 {
		t.Fatalf("mode = %q", g.Mode)
	}
}

func TestNewOnline3pChain_ThreeHumans(t *testing.T) {
	g, err := NewOnline3pChain("t", [3]string{"A", "B", "C"}, [3]string{CharLiuBei, CharGuanYu, CharZhangFei})
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Players) != 3 {
		t.Fatalf("players = %d, want 3", len(g.Players))
	}
	for i, p := range g.Players {
		if p.IsAI {
			t.Fatalf("seat %d should be human", i)
		}
	}
	if g.Mode != Mode3pChain {
		t.Fatalf("mode = %q", g.Mode)
	}
}

func TestNewOnline3v3_SixHumans(t *testing.T) {
	g, err := NewOnline3v3(
		"t",
		[6]string{"A", "B", "C", "D", "E", "F"},
		[6]string{CharLiuBei, CharGuanYu, CharZhangFei, CharZhaoYun, CharMaChao, CharHuangYueying},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Players) != 6 {
		t.Fatalf("players = %d, want 6", len(g.Players))
	}
	for i, p := range g.Players {
		if p.IsAI {
			t.Fatalf("seat %d should be human", i)
		}
	}
	if g.Mode != Mode3v3 {
		t.Fatalf("mode = %q", g.Mode)
	}
}

func TestNewOnlineIdentity5_FiveHumans(t *testing.T) {
	g, err := NewOnlineIdentity5(
		"t",
		[5]string{"A", "B", "C", "D", "E"},
		[5]string{CharLiuBei, CharGuanYu, CharZhangFei, CharZhaoYun, CharMaChao},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Players) != 5 {
		t.Fatalf("players = %d, want 5", len(g.Players))
	}
	for i, p := range g.Players {
		if p.IsAI {
			t.Fatalf("seat %d should be human", i)
		}
	}
	if g.Mode != ModeIdentity5 {
		t.Fatalf("mode = %q", g.Mode)
	}
	if err := mode.ValidateIdentity5Roles(g.Identities); err != nil {
		t.Fatal(err)
	}
	if g.LordSeat < 0 || g.LordSeat >= 5 {
		t.Fatalf("lord seat = %d", g.LordSeat)
	}
	if g.Identities[g.LordSeat] != mode.RoleLord {
		t.Fatalf("lord seat role = %q", g.Identities[g.LordSeat])
	}
	lord := g.Players[g.LordSeat]
	if lord.MaxHP != lord.Character.MaxHP+1 {
		t.Fatalf("lord maxHP = %d want %d", lord.MaxHP, lord.Character.MaxHP+1)
	}
}

func TestNewOnlineIdentity8_EightHumans(t *testing.T) {
	g, err := NewOnlineIdentity8(
		"t",
		[8]string{"A", "B", "C", "D", "E", "F", "G", "H"},
		[8]string{
			CharLiuBei, CharGuanYu, CharZhangFei, CharZhaoYun,
			CharMaChao, CharHuangYueying, CharSunQuan, CharZhouYu,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Players) != 8 {
		t.Fatalf("players = %d, want 8", len(g.Players))
	}
	for i, p := range g.Players {
		if p.IsAI {
			t.Fatalf("seat %d should be human", i)
		}
	}
	if g.Mode != ModeIdentity8 {
		t.Fatalf("mode = %q", g.Mode)
	}
	if err := mode.ValidateIdentity8Roles(g.Identities); err != nil {
		t.Fatal(err)
	}
	if g.Identities[g.LordSeat] != mode.RoleLord {
		t.Fatalf("lord seat role = %q", g.Identities[g.LordSeat])
	}
}
