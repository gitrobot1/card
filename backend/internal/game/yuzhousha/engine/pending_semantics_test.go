package engine

import "testing"

func TestFillPendingRoles_PojunActor(t *testing.T) {
	p := &PendingCombat{
		SourceIndex:  0,
		TargetIndex:  1,
		ResponseMode: "skill_pojun",
	}
	FillPendingRoles(p)
	if p.WindowKind != WindowKindTake {
		t.Fatalf("window_kind = %q, want take", p.WindowKind)
	}
	if p.ActorSeat != 0 {
		t.Fatalf("actor_seat = %d, want 0 (source)", p.ActorSeat)
	}
	if p.SubjectSeat != 1 {
		t.Fatalf("subject_seat = %d, want 1 (target)", p.SubjectSeat)
	}
	if p.OriginSeat != 0 {
		t.Fatalf("origin_seat = %d, want 0", p.OriginSeat)
	}
}

func TestFillPendingRoles_TuxiActor(t *testing.T) {
	p := &PendingCombat{
		SourceIndex:  1,
		TargetIndex:  0,
		ResponseMode: ResponseModeSkillTuxi,
	}
	FillPendingRoles(p)
	if p.WindowKind != WindowKindTake {
		t.Fatalf("window_kind = %q, want take", p.WindowKind)
	}
	if p.ActorSeat != 0 {
		t.Fatalf("actor_seat = %d, want 0 (attacker seat)", p.ActorSeat)
	}
	if p.SubjectSeat != 1 {
		t.Fatalf("subject_seat = %d, want 1 (victim seat)", p.SubjectSeat)
	}
}

func TestFillPendingRoles_PojunDiscard(t *testing.T) {
	p := &PendingCombat{
		SourceIndex:  1,
		TargetIndex:  1,
		ResponseMode: "skill_pojun_discard",
	}
	FillPendingRoles(p)
	if p.WindowKind != WindowKindDiscard {
		t.Fatalf("window_kind = %q, want discard", p.WindowKind)
	}
	if p.ActorSeat != 1 || p.SubjectSeat != 1 {
		t.Fatalf("actor/subject = %d/%d, want 1/1", p.ActorSeat, p.SubjectSeat)
	}
}

func TestFillPendingRoles_DyingAskSeat(t *testing.T) {
	p := &PendingCombat{
		SourceIndex:  1,
		TargetIndex:  0,
		ResponseMode: ResponseModeDying,
	}
	FillPendingRoles(p)
	if p.ActorSeat != 1 {
		t.Fatalf("actor_seat = %d, want 1 (ask seat)", p.ActorSeat)
	}
	if p.SubjectSeat != 0 {
		t.Fatalf("subject_seat = %d, want 0 (dying)", p.SubjectSeat)
	}
}

func TestPendingActorSeat_PrefersActorSeat(t *testing.T) {
	g := &Game{
		Phase: PhaseResponse,
		Pending: &PendingCombat{
			SourceIndex:  0,
			TargetIndex:  1,
			ResponseMode: "skill_pojun",
		},
	}
	if got := g.PendingActorSeat(); got != 0 {
		t.Fatalf("PendingActorSeat() = %d, want 0", got)
	}
}

func TestPublicViewForSeat_IncludesPendingRoles(t *testing.T) {
	g := &Game{
		ID:    "test",
		Phase: PhaseResponse,
		Players: []Player{
			{Index: 0, Name: "A"},
			{Index: 1, Name: "B"},
		},
		Pending: &PendingCombat{
			SourceIndex:  0,
			TargetIndex:  1,
			ResponseMode: "skill_pojun",
			Card:         Card{Kind: CardSha, Name: "杀"},
		},
	}
	pub := g.PublicViewForSeat(0, nil)
	if pub.Pending == nil {
		t.Fatal("pending is nil")
	}
	if pub.Pending.ActorSeat != 0 {
		t.Fatalf("json actor_seat = %d, want 0", pub.Pending.ActorSeat)
	}
	if pub.Pending.SubjectSeat != 1 {
		t.Fatalf("json subject_seat = %d, want 1", pub.Pending.SubjectSeat)
	}
	if pub.Pending.WindowKind != WindowKindTake {
		t.Fatalf("json window_kind = %q, want take", pub.Pending.WindowKind)
	}
}
