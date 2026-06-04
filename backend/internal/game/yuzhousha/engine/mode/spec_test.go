package mode

import "testing"

type stubCtx struct {
	mode    string
	players []int // HP per seat
}

func (s stubCtx) ModeID() string     { return s.mode }
func (s stubCtx) PlayerCount() int   { return len(s.players) }
func (s stubCtx) AliveHP(seat int) int { return s.players[seat] }

func TestDefaultEnemy1v1(t *testing.T) {
	ctx := stubCtx{mode: Solo1v1, players: []int{4, 4}}
	if got := DefaultEnemy(ctx, 0); got != 1 {
		t.Fatalf("expected opponent 1, got %d", got)
	}
}

func TestDefaultEnemy2v2Clockwise(t *testing.T) {
	ctx := stubCtx{mode: Solo2v2, players: []int{4, 4, 4, 4}}
	if got := DefaultEnemy(ctx, 0); got != 1 {
		t.Fatalf("seat 0 default enemy want 1, got %d", got)
	}
	if got := DefaultEnemy(ctx, 1); got != 2 {
		t.Fatalf("seat 1 default enemy want 2 (ally), got %d", got)
	}
}

func TestEnemiesOf2v2(t *testing.T) {
	ctx := stubCtx{mode: Solo2v2, players: []int{4, 3, 0, 4}}
	enemies := EnemiesOf(ctx, 0)
	if len(enemies) != 2 || enemies[0] != 1 || enemies[1] != 3 {
		t.Fatalf("unexpected enemies: %v", enemies)
	}
}

func TestAlliesOf2v2(t *testing.T) {
	ctx := stubCtx{mode: Solo2v2, players: []int{4, 3, 2, 0}}
	allies := AlliesOf(ctx, 0)
	if len(allies) != 1 || allies[0] != 2 {
		t.Fatalf("unexpected allies: %v", allies)
	}
	if len(AlliesOf(ctx, 1)) != 0 {
		t.Fatal("dead ally should not appear")
	}
}
