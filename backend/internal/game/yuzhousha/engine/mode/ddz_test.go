package mode

import "testing"

type ddzStubCtx struct {
	stubCtx
	landlord int
}

func (d ddzStubCtx) DdzLandlordSeat() int { return d.landlord }

func TestTeamOf_3pDdz(t *testing.T) {
	ctx := ddzStubCtx{stubCtx: stubCtx{mode: Solo3pDdz, players: []int{4, 4, 4}}, landlord: 0}
	if TeamOf(ctx, 0) != 0 {
		t.Fatal("seat 0 should be landlord team")
	}
	if TeamOf(ctx, 1) != 1 || TeamOf(ctx, 2) != 1 {
		t.Fatal("farmers should be team 1")
	}
	if !IsEnemy(ctx, 0, 1) || !IsEnemy(ctx, 0, 2) {
		t.Fatal("landlord should be enemy of farmers")
	}
	if !IsAlly(ctx, 1, 2) {
		t.Fatal("farmers should be allies")
	}
}

func TestIs3pDdz(t *testing.T) {
	if !Is3pDdz(stubCtx{mode: Solo3pDdz, players: []int{4, 4, 4}}) {
		t.Fatal("expected 3p ddz")
	}
	if Is3pDdz(stubCtx{mode: Solo3pChain, players: []int{4, 4, 4}}) {
		t.Fatal("chain is not ddz")
	}
}

func TestDefaultEnemy3pDdz(t *testing.T) {
	ctx := ddzStubCtx{stubCtx: stubCtx{mode: Solo3pDdz, players: []int{4, 4, 4}}, landlord: 0}
	if DefaultEnemy(ctx, 2) < 0 || DefaultEnemy(ctx, 2) >= ctx.PlayerCount() {
		t.Fatalf("seat 2 default enemy=%d out of range", DefaultEnemy(ctx, 2))
	}
	if DefaultEnemy(ctx, 2) != 0 {
		t.Fatalf("farmer seat 2 should target landlord, got %d", DefaultEnemy(ctx, 2))
	}
}
