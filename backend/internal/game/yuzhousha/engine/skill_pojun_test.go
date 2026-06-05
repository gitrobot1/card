package engine

import "testing"

func TestPojunTakeArmorThenPassShan(t *testing.T) {
	g, err := NewSolo1v1("pojun-vine-debug", "甲", CharJieXuSheng, CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []Card{{ID: "sha-1", Kind: CardSha, Name: "杀"}}
	g.Players[1].Armor = &Card{ID: "vine", Kind: CardArmorVine, Name: "藤甲"}
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = 0

	var events []GameEvent
	if err := g.PlayCard(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModeSkillPojun {
		t.Fatalf("pending after sha: mode=%q takeWindow=%v", g.Pending.ResponseMode, g.takeWindow != nil)
	}
	if err := g.UseSkill(0, UseSkillRequest{
		SkillID: SkillPojun, TargetZone: EquipArmor, TargetCardID: "vine",
	}, &events); err != nil {
		t.Fatalf("UseSkill: %v", err)
	}
	if len(g.Players[1].CampCards) != 1 {
		t.Fatalf("camp want 1 after armor take, got %d", len(g.Players[1].CampCards))
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatalf("PassResponse(1): %v", err)
	}
	if g.Players[1].HP != DefaultMaxHP-1 {
		t.Fatalf("vine in camp want 1 sha damage, hp=%d", g.Players[1].HP)
	}
}
