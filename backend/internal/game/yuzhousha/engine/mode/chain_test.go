package mode

import "testing"

func Test3pChainRelations(t *testing.T) {
	ctx := stubCtx{mode: Solo3pChain, players: []int{4, 4, 4}}
	if UpperSeat(0, 3) != 2 {
		t.Fatalf("seat0 upper want 2 got %d", UpperSeat(0, 3))
	}
	if LowerSeat(0, 3) != 1 {
		t.Fatalf("seat0 lower want 1 got %d", LowerSeat(0, 3))
	}
	if !IsEnemy(ctx, 0, 2) || IsEnemy(ctx, 0, 1) {
		t.Fatal("seat0 enemy should be upper only")
	}
	if !IsAlly(ctx, 0, 1) || IsAlly(ctx, 0, 2) {
		t.Fatal("seat0 ally should be lower only")
	}
	if got := DefaultEnemy(ctx, 0); got != 2 {
		t.Fatalf("default enemy want 2 got %d", got)
	}
}

func Test3pChainAoeSkipsProtect(t *testing.T) {
	ctx := stubCtx{mode: Solo3pChain, players: []int{4, 4, 4}}
	q := AoeResponderQueue(ctx, 0)
	if len(q) != 1 || q[0] != 2 {
		t.Fatalf("seat0 AOE queue want [2], got %v", q)
	}
}

func TestEvaluateHumanChainDeath(t *testing.T) {
	ctx := stubCtx{mode: Solo3pChain, players: []int{4, 4, 4}}
	if o, _ := EvaluateHumanChainDeath(ctx, 0, 2); o != ChainHumanWin {
		t.Fatal("upper death should win")
	}
	if o, _ := EvaluateHumanChainDeath(ctx, 0, 1); o != ChainHumanLose {
		t.Fatal("lower death should lose")
	}
	if o, _ := EvaluateHumanChainDeath(ctx, 0, 0); o != ChainHumanLose {
		t.Fatal("self death should lose")
	}
}

func TestValidPlayTargets3pChainSha(t *testing.T) {
	ctx := stubTargetCtx{
		stubCtx: stubCtx{mode: Solo3pChain, players: []int{4, 4, 4}},
		canAttack: map[[2]int]bool{{0, 2}: true},
	}
	valid := ValidPlayTargets(ctx, 0, TargetSha)
	if len(valid) != 1 || valid[0] != 2 {
		t.Fatalf("sha targets want [2], got %v", valid)
	}
	if IsValidPlayTarget(ctx, 0, 1, TargetSha) {
		t.Fatal("should not sha protect target")
	}
}
