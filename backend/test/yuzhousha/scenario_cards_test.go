package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

func TestScenario_GudingDaoExtraDamageWhenTargetEmptyHand(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-guding", "甲", engine.CharLiuBei, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Weapon = &engine.Card{ID: "w6", Kind: engine.CardWeapon6, Name: "古锭刀"}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Hand = nil
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-2 {
		t.Fatalf("guding dao want 2 damage, hp=%d", g.Players[1].HP)
	}
}

func TestScenario_VineArmorSkipsNanman(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-vine-nm", "甲", engine.CharLiuBei, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "nm-1", Kind: engine.CardNanMan, Name: "南蛮入侵"}}
	g.Players[1].Armor = &engine.Card{ID: "vine", Kind: engine.CardArmorVine, Name: "藤甲"}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlayCard(0, "nm-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Phase != engine.PhasePlaying || g.CurrentTurn != 0 {
		t.Fatalf("vine should skip nanman response, phase=%s turn=%d pending=%+v", g.Phase, g.CurrentTurn, g.Pending)
	}
}

func TestScenario_HuoGongFireDamageOnPass(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-huogong", "甲", engine.CharLiuBei, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "hg-1", Kind: engine.CardHuoGong, Name: "火攻"}}
	g.Players[1].Hand = []engine.Card{{ID: "shown", Kind: engine.CardShan, Name: "闪", Suit: "heart"}}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlayCard(0, "hg-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeHuoGong)
	if err := g.PassResponse(0, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-1 {
		t.Fatalf("huogong pass want 1 damage, hp=%d", g.Players[1].HP)
	}
}

func TestScenario_TieSuoRecastDraws(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-tiesuo", "甲", engine.CharLiuBei, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "ts-1", Kind: engine.CardTieSuo, Name: "铁索连环"}}
	g.DrawPile = []engine.Card{{ID: "draw-1", Kind: engine.CardShan, Name: "闪"}}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlayCardWithTarget(0, "ts-1", engine.PlayTarget{SeatIndex: 0}, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 1 || g.Players[0].Hand[0].ID != "draw-1" {
		t.Fatalf("tiesuo recast want 1 drawn card, hand=%+v", g.Players[0].Hand)
	}
}

func TestScenario_VineArmorPlusOneShaDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-vine-sha", "甲", engine.CharLiuBei, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Armor = &engine.Card{ID: "vine", Kind: engine.CardArmorVine, Name: "藤甲"}
	g.Players[1].Hand = []engine.Card{{ID: "h1", Kind: engine.CardShan, Name: "闪"}}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-2 {
		t.Fatalf("vine + sha want 2 damage, hp=%d", g.Players[1].HP)
	}
}

func TestScenario_PojunJiuGudingDaoThreeDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-pojun-guding", "甲", engine.CharJieXuSheng, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Weapon = &engine.Card{ID: "w6", Kind: engine.CardWeapon6, Name: "古锭刀"}
	g.Players[0].Hand = []engine.Card{
		{ID: "jiu-1", Kind: engine.CardJiu, Name: "酒"},
		{ID: "sha-1", Kind: engine.CardSha, Name: "杀"},
	}
	g.Players[1].Hand = []engine.Card{{ID: "h1", Kind: engine.CardShan, Name: "闪"}}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlayCard(0, "jiu-1", 0, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeSkillPojun)
	if err := g.UseSkill(0, engine.UseSkillRequest{
		SkillID: engine.SkillPojun, TargetZone: "hand", TargetCardID: "h1",
	}, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-3 {
		t.Fatalf("jiu+sha+guding after pojun want 3 damage, hp=%d", g.Players[1].HP)
	}
}

func TestScenario_PojunMovesVineArmorRemovesShaBonus(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-pojun-vine", "甲", engine.CharJieXuSheng, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Armor = &engine.Card{ID: "vine", Kind: engine.CardArmorVine, Name: "藤甲"}
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	assertPendingMode(t, g, engine.ResponseModeSkillPojun)
	if err := g.UseSkill(0, engine.UseSkillRequest{
		SkillID: engine.SkillPojun, TargetZone: engine.EquipArmor, TargetCardID: "vine",
	}, &events); err != nil {
		t.Fatal(err)
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-1 {
		t.Fatalf("vine moved to camp want 1 damage, hp=%d armor=%+v camp=%d",
			g.Players[1].HP, g.Players[1].Armor, len(g.Players[1].CampCards))
	}
}

func TestScenario_VineArmorKeptPlusOneShaDamage(t *testing.T) {
	g, err := engine.NewSolo1v1("sc-vine-kept", "甲", engine.CharJieXuSheng, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].Hand = []engine.Card{{ID: "sha-1", Kind: engine.CardSha, Name: "杀"}}
	g.Players[1].Armor = &engine.Card{ID: "vine", Kind: engine.CardArmorVine, Name: "藤甲"}
	g.Players[1].Hand = nil
	setupPlayingTurn(g, 0)

	var events []engine.GameEvent
	if err := g.PlaySha(0, "sha-1", 1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Pending != nil && g.Pending.ResponseMode == engine.ResponseModeSkillPojun {
		if err := g.PassResponse(0, &events); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.PassResponse(1, &events); err != nil {
		t.Fatal(err)
	}
	if g.Players[1].HP != engine.DefaultMaxHP-2 {
		t.Fatalf("vine equipped want 2 sha damage, hp=%d", g.Players[1].HP)
	}
}
