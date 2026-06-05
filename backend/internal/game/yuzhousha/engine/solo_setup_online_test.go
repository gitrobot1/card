package engine

import "testing"

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
