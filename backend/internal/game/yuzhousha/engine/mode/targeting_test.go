package mode

import (
	"fmt"
	"testing"
)

type stubTargetCtx struct {
	stubCtx
	canAttack    map[[2]int]bool
	hasTakeable  map[int]bool
	canBingliang map[[2]int]bool
	blocked      map[string]bool
	maxHP        []int
}

func (s stubTargetCtx) CanAttack(from, to int) bool {
	if s.canAttack != nil {
		return s.canAttack[[2]int{from, to}]
	}
	return true
}

func (s stubTargetCtx) HasTakeableCard(target int) bool {
	if s.hasTakeable != nil {
		return s.hasTakeable[target]
	}
	return true
}

func (s stubTargetCtx) CanBingliangTarget(from, to int) bool {
	if s.canBingliang != nil {
		return s.canBingliang[[2]int{from, to}]
	}
	return true
}

func (s stubTargetCtx) TargetBlocked(target int, cardKind string) bool {
	if s.blocked != nil {
		return s.blocked[fmt.Sprintf("%d:%s", target, cardKind)]
	}
	return false
}

func (s stubTargetCtx) PlayerHP(seat int) (hp, maxHP int) {
	hp = s.AliveHP(seat)
	maxHP = 4
	if s.maxHP != nil && seat < len(s.maxHP) {
		maxHP = s.maxHP[seat]
	}
	return hp, maxHP
}

func TestValidPlayTargets2v2Sha(t *testing.T) {
	ctx := stubTargetCtx{
		stubCtx: stubCtx{mode: Solo2v2, players: []int{4, 4, 4, 4}},
		canAttack: map[[2]int]bool{
			{0, 1}: true,
			{0, 3}: false,
		},
	}
	targets := ValidPlayTargets(ctx, 0, TargetSha)
	if len(targets) != 1 || targets[0] != 1 {
		t.Fatalf("want [1], got %v", targets)
	}
	if IsValidPlayTarget(ctx, 0, 2, TargetSha) {
		t.Fatal("teammate must not be valid sha target")
	}
}

func TestValidPlayTargetsGuoheNeedsCards(t *testing.T) {
	ctx := stubTargetCtx{
		stubCtx:     stubCtx{mode: Solo1v1, players: []int{4, 4}},
		hasTakeable: map[int]bool{1: false},
	}
	if IsValidPlayTarget(ctx, 0, 1, TargetGuohe) {
		t.Fatal("guohe should fail without takeable cards")
	}
	ctx.hasTakeable[1] = true
	if !IsValidPlayTarget(ctx, 0, 1, TargetGuohe) {
		t.Fatal("guohe should succeed with takeable cards")
	}
}

func TestPickAITargetPrefersWounded(t *testing.T) {
	ctx := stubTargetCtx{
		stubCtx: stubCtx{mode: Solo2v2, players: []int{4, 2, 4, 4}},
		canAttack: map[[2]int]bool{
			{0, 1}: true,
			{0, 3}: true,
		},
		maxHP: []int{4, 4, 4, 4},
	}
	if got := PickAITarget(ctx, 0, TargetSha); got != 1 {
		t.Fatalf("want wounded seat 1, got %d", got)
	}
}
