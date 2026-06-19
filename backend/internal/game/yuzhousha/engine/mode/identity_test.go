package mode_test

import (
	"testing"

	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

type identityStub struct {
	stubCtx
	lord     int
	roles    []string
	revealed []bool
}

func (s identityStub) IdentityLordSeat() int { return s.lord }
func (s identityStub) IdentityOf(seat int) string {
	if seat < 0 || seat >= len(s.roles) {
		return ""
	}
	return s.roles[seat]
}
func (s identityStub) IdentityRevealed(seat int) bool {
	if seat < 0 || seat >= len(s.revealed) {
		return false
	}
	return s.revealed[seat]
}

func TestNormalizeID_Identity5(t *testing.T) {
	if got := mode.NormalizeID("identity_5"); got != mode.SoloIdentity5 {
		t.Fatalf("NormalizeID(identity_5) = %q, want %q", got, mode.SoloIdentity5)
	}
}

func TestIsIdentityMode(t *testing.T) {
	for _, id := range []string{mode.SoloIdentity5, mode.SoloIdentity8} {
		if !mode.IsIdentityMode(id) {
			t.Fatalf("IsIdentityMode(%q) = false, want true", id)
		}
	}
	if mode.IsIdentityMode(mode.Solo1v1) || mode.IsIdentityMode(mode.Solo3v3) {
		t.Fatal("non-identity modes should not match IsIdentityMode")
	}
}

func TestLordSkillsActive(t *testing.T) {
	if !mode.LordSkillsActive(mode.Solo2v2) {
		t.Fatal("2v2 should enable lord skills")
	}
	if !mode.LordSkillsActive(mode.SoloIdentity5) {
		t.Fatal("identity_5 should enable lord skills")
	}
	if !mode.LordSkillsActive(mode.SoloIdentity8) {
		t.Fatal("identity_8 should enable lord skills")
	}
	if mode.LordSkillsActive(mode.Solo1v1) {
		t.Fatal("1v1 should not enable lord skills")
	}
}

func TestValidateIdentity5Roles(t *testing.T) {
	ok := []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel}
	if err := mode.ValidateIdentity5Roles(ok); err != nil {
		t.Fatal(err)
	}
	bad := []string{mode.RoleLord, mode.RoleLord, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel}
	if err := mode.ValidateIdentity5Roles(bad); err == nil {
		t.Fatal("expected invalid role set error")
	}
}

func TestTeamOf_Identity5(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{5, 4, 4, 4, 4}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	if mode.TeamOf(ctx, 0) != 0 || mode.TeamOf(ctx, 1) != 0 {
		t.Fatal("lord faction team mismatch")
	}
	if mode.TeamOf(ctx, 2) != mode.IdentityTeamSpy {
		t.Fatal("spy team mismatch")
	}
	if mode.TeamOf(ctx, 3) != 1 || mode.TeamOf(ctx, 4) != 1 {
		t.Fatal("rebel team mismatch")
	}
}

func TestIsAlly_Identity5(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{5, 4, 4, 4, 4}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	if !mode.IsAlly(ctx, 0, 1) || mode.IsAlly(ctx, 0, 2) {
		t.Fatal("lord should ally loyalist only, not spy")
	}
	if mode.IsAlly(ctx, 2, 3) {
		t.Fatal("spy should not ally rebels")
	}
	if !mode.IsAlly(ctx, 3, 4) {
		t.Fatal("rebels should be allies")
	}
}

func TestValidPlayTargets_Identity5AnyOther(t *testing.T) {
	stub := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{5, 4, 4, 4, 4}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	targets := mode.ValidPlayTargets(identityTargetStub{identityStub: stub}, 0, mode.TargetSha)
	if len(targets) != 4 {
		t.Fatalf("lord should target 4 others, got %v", targets)
	}
	for _, tseat := range targets {
		if tseat == 0 {
			t.Fatal("should not include self")
		}
	}
}

func TestEvaluateIdentityWin_LordDeath(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{0, 4, 4, 4, 4}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	finished, winner, msg := mode.EvaluateIdentityWin(ctx, 0, 3)
	if !finished || winner != mode.IdentityTeamRebel || msg == "" {
		t.Fatalf("lord death: finished=%v winner=%d msg=%q", finished, winner, msg)
	}
}

func TestEvaluateIdentityWin_LordDeathNoLivingRebels(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{0, 4, 4, 0, 0}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	finished, winner, _ := mode.EvaluateIdentityWin(ctx, 0, -1)
	if !finished || winner != mode.IdentityTeamRebel {
		t.Fatalf("lord death with no rebels alive: finished=%v winner=%d", finished, winner)
	}
}

func TestEvaluateIdentityWin_LordDeathSpyDuelKilledBySpy(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{0, 0, 4, 0, 0}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	finished, winner, msg := mode.EvaluateIdentityWin(ctx, 0, 2)
	if !finished || winner != mode.IdentityTeamSpy || msg == "" {
		t.Fatalf("spy duel kill: finished=%v winner=%d msg=%q", finished, winner, msg)
	}
}

func TestEvaluateIdentityWin_LordDeathSpyDuelUnknownKiller(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{0, 0, 4, 0, 0}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	finished, winner, msg := mode.EvaluateIdentityWin(ctx, 0, -1)
	if !finished || winner != mode.IdentityTeamSpy || msg == "" {
		t.Fatalf("spy duel unknown killer: finished=%v winner=%d msg=%q", finished, winner, msg)
	}
}

func TestEvaluateIdentityWin_RebelsEliminated(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{5, 4, 0, 0, 0}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	finished, winner, _ := mode.EvaluateIdentityWin(ctx, 3, 0)
	if !finished || winner != mode.IdentityTeamLordFaction {
		t.Fatalf("rebels and spy eliminated: finished=%v winner=%d", finished, winner)
	}
}

func TestEvaluateIdentityWin_RebelsDeadSpyAlive(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{5, 4, 4, 0, 0}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel},
	}
	finished, _, _ := mode.EvaluateIdentityWin(ctx, 3, 0)
	if finished {
		t.Fatal("game should continue when spy still alive")
	}
}

func TestEvaluateIdentityWin_SpySolo(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity5, n: 5, hp: []int{0, 0, 0, 0, 4}},
		lord:    0,
		roles:   []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleRebel, mode.RoleRebel, mode.RoleSpy},
	}
	finished, winner, msg := mode.EvaluateIdentityWin(ctx, 3, 4)
	if !finished || winner != mode.IdentityTeamSpy || msg == "" {
		t.Fatalf("spy solo: finished=%v winner=%d msg=%q", finished, winner, msg)
	}
}

func TestValidateIdentity8Roles(t *testing.T) {
	ok := []string{
		mode.RoleLord, mode.RoleLoyalist, mode.RoleLoyalist, mode.RoleSpy,
		mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel,
	}
	if err := mode.ValidateIdentity8Roles(ok); err != nil {
		t.Fatal(err)
	}
	bad := []string{mode.RoleLord, mode.RoleLoyalist, mode.RoleSpy, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel}
	if err := mode.ValidateIdentity8Roles(bad); err == nil {
		t.Fatal("expected error for missing loyalist")
	}
}

func TestTeamOf_Identity8(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity8, n: 8, hp: []int{5, 4, 4, 4, 4, 4, 4, 4}},
		lord:    0,
		roles: []string{
			mode.RoleLord, mode.RoleLoyalist, mode.RoleLoyalist, mode.RoleSpy,
			mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel,
		},
	}
	if mode.TeamOf(ctx, 0) != 0 || mode.TeamOf(ctx, 1) != 0 || mode.TeamOf(ctx, 2) != 0 {
		t.Fatal("lord faction team mismatch")
	}
	if mode.TeamOf(ctx, 3) != mode.IdentityTeamSpy {
		t.Fatal("spy team mismatch")
	}
	for _, seat := range []int{4, 5, 6, 7} {
		if mode.TeamOf(ctx, seat) != 1 {
			t.Fatalf("rebel seat %d team mismatch", seat)
		}
	}
}

func TestIsAlly_Identity8(t *testing.T) {
	ctx := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity8, n: 8, hp: []int{5, 4, 4, 4, 4, 4, 4, 4}},
		lord:    0,
		roles: []string{
			mode.RoleLord, mode.RoleLoyalist, mode.RoleLoyalist, mode.RoleSpy,
			mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel,
		},
	}
	if !mode.IsAlly(ctx, 0, 1) || !mode.IsAlly(ctx, 0, 2) || mode.IsAlly(ctx, 0, 3) {
		t.Fatal("lord should ally loyalists only, not spy")
	}
	if mode.IsAlly(ctx, 3, 4) {
		t.Fatal("spy should not ally rebels")
	}
	if !mode.IsAlly(ctx, 4, 7) {
		t.Fatal("rebels should be allies")
	}
}

func TestValidPlayTargets_Identity8AnyOther(t *testing.T) {
	stub := identityStub{
		stubCtx: stubCtx{modeID: mode.SoloIdentity8, n: 8, hp: []int{5, 4, 4, 4, 4, 4, 4, 4}},
		lord:    0,
		roles: []string{
			mode.RoleLord, mode.RoleLoyalist, mode.RoleLoyalist, mode.RoleSpy,
			mode.RoleRebel, mode.RoleRebel, mode.RoleRebel, mode.RoleRebel,
		},
	}
	targets := mode.ValidPlayTargets(identityTargetStub{identityStub: stub}, 0, mode.TargetSha)
	if len(targets) != 7 {
		t.Fatalf("lord should target 7 others, got %v", targets)
	}
	for _, tseat := range targets {
		if tseat == 0 {
			t.Fatal("should not include self")
		}
	}
}

func TestNormalizeID_Identity8(t *testing.T) {
	if got := mode.NormalizeID("identity_8"); got != mode.SoloIdentity8 {
		t.Fatalf("NormalizeID(identity_8) = %q, want %q", got, mode.SoloIdentity8)
	}
	if got := mode.NormalizeID("8人身份局"); got != mode.SoloIdentity8 {
		t.Fatalf("NormalizeID(8人身份局) = %q, want %q", got, mode.SoloIdentity8)
	}
}

func TestLookup_Identity8Meta(t *testing.T) {
	meta, ok := mode.Lookup(mode.SoloIdentity8)
	if !ok {
		t.Fatal("identity_8 mode not registered")
	}
	if meta.PlayerCount != 8 || meta.LayoutKey != mode.LayoutOctagon8 {
		t.Fatalf("unexpected meta: %+v", meta)
	}
	if len(meta.SeatMap) != 7 {
		t.Fatalf("expected 7 seat_map entries, got %d", len(meta.SeatMap))
	}
}

func TestLookup_Identity5Meta(t *testing.T) {
	meta, ok := mode.Lookup(mode.SoloIdentity5)
	if !ok {
		t.Fatal("identity_5 mode not registered")
	}
	if meta.PlayerCount != 5 || meta.LayoutKey != mode.LayoutPentagon5 {
		t.Fatalf("unexpected meta: %+v", meta)
	}
}

// identityTargetStub implements TargetContext for identity targeting tests.
type identityTargetStub struct {
	identityStub
}

func (s identityTargetStub) CanAttack(_, _ int) bool            { return true }
func (s identityTargetStub) HasTakeableCard(_ int) bool         { return true }
func (s identityTargetStub) CanBingliangTarget(_, _ int) bool   { return true }
func (s identityTargetStub) TargetBlocked(_ int, _ string) bool { return false }
func (s identityTargetStub) PlayerHP(seat int) (int, int) {
	hp := s.AliveHP(seat)
	return hp, 5
}
func (s identityTargetStub) HandCount(_ int) int { return 1 }
func (s identityTargetStub) LimuActive(_ int) bool { return false }
