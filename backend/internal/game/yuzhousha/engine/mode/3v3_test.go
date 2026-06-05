package mode_test

import (
	"testing"

	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

type stubCtx struct {
	modeID string
	n      int
	hp     []int
}

func (s stubCtx) ModeID() string       { return s.modeID }
func (s stubCtx) PlayerCount() int     { return s.n }
func (s stubCtx) AliveHP(seat int) int { return s.hp[seat] }

func TestNormalizeID_3v3(t *testing.T) {
	if got := mode.NormalizeID("3V3"); got != mode.Solo3v3 {
		t.Fatalf("NormalizeID(3V3) = %q, want %q", got, mode.Solo3v3)
	}
}

func TestTeamOf_3v3(t *testing.T) {
	ctx := stubCtx{modeID: mode.Solo3v3, n: 6, hp: []int{4, 4, 4, 4, 4, 4}}
	cases := map[int]int{0: 0, 4: 0, 5: 0, 1: 1, 2: 1, 3: 1}
	for seat, want := range cases {
		if got := mode.TeamOf(ctx, seat); got != want {
			t.Fatalf("TeamOf seat %d = %d, want %d", seat, got, want)
		}
	}
}

func TestIsCommander3v3(t *testing.T) {
	if !mode.IsCommander3v3(0) || !mode.IsCommander3v3(2) {
		t.Fatal("seats 0 and 2 should be commanders")
	}
	if mode.IsCommander3v3(1) || mode.IsCommander3v3(4) {
		t.Fatal("forwards should not be commanders")
	}
}

func TestDefaultEnemy_3v3PrefersCommander(t *testing.T) {
	ctx := stubCtx{modeID: mode.Solo3v3, n: 6, hp: []int{4, 4, 4, 4, 4, 4}}
	if got := mode.DefaultEnemy(ctx, 0); got != 2 {
		t.Fatalf("DefaultEnemy warm commander = %d, want cold commander 2", got)
	}
}

func TestLookup_3v3Meta(t *testing.T) {
	meta, ok := mode.Lookup(mode.Solo3v3)
	if !ok {
		t.Fatal("3v3 mode not registered")
	}
	if meta.PlayerCount != 6 || meta.LayoutKey != mode.LayoutHex3v3 {
		t.Fatalf("unexpected meta: %+v", meta)
	}
}

func TestEvaluateCommanderDeath(t *testing.T) {
	ctx := stubCtx{modeID: mode.Solo3v3, n: 6, hp: []int{4, 4, 0, 4, 4, 4}}
	finished, winner, _ := mode.EvaluateCommanderDeath(ctx, 0, 2)
	if !finished || winner != 0 {
		t.Fatalf("cold commander death: finished=%v winner=%d", finished, winner)
	}
}
