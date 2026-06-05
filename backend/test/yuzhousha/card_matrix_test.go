package engine_test

import (
	"fmt"
	"strings"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
)

type cardMatrixEntry struct {
	kind         string
	category     string // basic | trick | equip
	playable     bool
	oppHandCount int
	prep         func(g *engine.Game)
}

func cardMatrixCatalog() []cardMatrixEntry {
	return []cardMatrixEntry{
		{kind: engine.CardSha, category: "basic", playable: true},
		{kind: engine.CardTao, category: "basic", playable: true, prep: func(g *engine.Game) {
			g.Players[0].HP = g.Players[0].MaxHP - 1
		}},
		{kind: engine.CardJiu, category: "basic", playable: true},
		{kind: engine.CardShan, category: "basic", playable: false},
		{kind: engine.CardGuoHe, category: "trick", playable: true, oppHandCount: 2},
		{kind: engine.CardTanNang, category: "trick", playable: true, oppHandCount: 2},
		{kind: engine.CardNanMan, category: "trick", playable: true},
		{kind: engine.CardWanJian, category: "trick", playable: true},
		{kind: engine.CardJueDou, category: "trick", playable: true},
		{kind: engine.CardLeBu, category: "trick", playable: true},
		{kind: engine.CardBingLiang, category: "trick", playable: true},
		{kind: engine.CardShanDian, category: "trick", playable: true},
		{kind: engine.CardWuGu, category: "trick", playable: true},
		{kind: engine.CardTaoYuan, category: "trick", playable: true},
		{kind: engine.CardWuZhong, category: "trick", playable: true},
		{kind: engine.CardWuxiek, category: "trick", playable: false},
		{kind: engine.CardWeapon1, category: "equip", playable: true},
		{kind: engine.CardWeapon2, category: "equip", playable: true},
		{kind: engine.CardWeapon3, category: "equip", playable: true},
		{kind: engine.CardWeapon4, category: "equip", playable: true},
		{kind: engine.CardWeapon5, category: "equip", playable: true},
		{kind: engine.CardWeapon6, category: "equip", playable: true},
		{kind: engine.CardArmor, category: "equip", playable: true},
		{kind: engine.CardArmorVine, category: "equip", playable: true},
		{kind: engine.CardHuoGong, category: "trick", playable: true, oppHandCount: 2},
		{kind: engine.CardTieSuo, category: "trick", playable: true},
		{kind: engine.CardPlusHorse, category: "equip", playable: true},
		{kind: engine.CardMinusHorse, category: "equip", playable: true},
	}
}

func pickMatrixCard(kind string) (engine.Card, []engine.Card, bool) {
	deck := engine.NewBasicDeck()
	var test engine.Card
	var rest []engine.Card
	found := false
	for _, c := range deck {
		if !found && c.Kind == kind {
			test = c
			found = true
			continue
		}
		rest = append(rest, c)
	}
	return test, rest, found
}

func resetMatrixBoard(g *engine.Game, testCard engine.Card, rest []engine.Card, oppHandCount int) {
	if oppHandCount < 1 {
		oppHandCount = 1
	}
	if oppHandCount > len(rest) {
		oppHandCount = len(rest)
	}
	g.Players[1].Hand = append([]engine.Card(nil), rest[:oppHandCount]...)
	g.DrawPile = append([]engine.Card(nil), rest[oppHandCount:]...)
	g.DiscardPile = nil
	g.Pending = nil
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.CurrentTurn = 0
	for i := range g.Players {
		p := &g.Players[i]
		if i == 0 {
			p.Hand = []engine.Card{testCard}
		}
		p.Weapon = nil
		p.Armor = nil
		p.PlusHorse = nil
		p.MinusHorse = nil
		p.JudgeArea = nil
		p.SkipDraw = false
		p.SkipPlay = false
		p.ShaUsedThisTurn = false
		p.Drunk = false
		if p.HP <= 0 {
			p.HP = p.MaxHP
		}
	}
	g.SyncCounts()
}

func playTargetForKind(kind string) int {
	switch kind {
	case engine.CardSha, engine.CardGuoHe, engine.CardTanNang,
		engine.CardJueDou, engine.CardLeBu, engine.CardBingLiang,
		engine.CardHuoGong:
		return 1
	default:
		return 0
	}
}

func drainMatrixPending(t *testing.T, g *engine.Game) {
	t.Helper()
	var events []engine.GameEvent
	for step := 0; step < 48; step++ {
		if g.IsFinished() {
			return
		}
		if g.Phase != engine.PhaseResponse || g.Pending == nil {
			return
		}
		seat, ok := responseActor(g)
		if !ok {
			t.Fatalf("matrix drain: no response actor pending=%+v", g.Pending)
		}
		switch g.Pending.ResponseMode {
		case engine.ResponseModeWuguPick:
			if len(g.Pending.RevealedCards) > 0 {
				if err := g.PickWuguCardForTest(seat, g.Pending.RevealedCards[0].ID, &events); err != nil {
					t.Fatalf("matrix wugu pick: %v", err)
				}
				continue
			}
		}
		if err := g.PassResponse(seat, &events); err != nil {
			t.Fatalf("matrix drain pass step %d: %v mode=%s", step, err, g.Pending.ResponseMode)
		}
	}
	t.Fatalf("matrix drain exceeded steps, pending=%+v phase=%s", g.Pending, g.Phase)
}

func runHeroCardMatrixCase(t *testing.T, heroID string, entry cardMatrixEntry) {
	t.Helper()
	testCard, rest, ok := pickMatrixCard(entry.kind)
	if !ok {
		t.Fatalf("deck missing kind %s", entry.kind)
	}

	g, err := engine.NewSolo1v1("matrix-"+heroID+"-"+entry.kind, "甲", heroID, engine.CharLiuBei)
	if err != nil {
		t.Fatal(err)
	}
	resetMatrixBoard(g, testCard, rest, entry.oppHandCount)
	if entry.prep != nil {
		entry.prep(g)
	}
	g.SyncCounts()
	assertGameInvariants(t, g)

	target := playTargetForKind(entry.kind)
	var events []engine.GameEvent
	err = g.PlayCard(0, testCard.ID, target, &events)

	if !entry.playable {
		if err == nil {
			t.Fatalf("expected play rejection for %s in play phase", entry.kind)
		}
		assertGameInvariants(t, g)
		return
	}
	if err != nil {
		t.Fatalf("play %s as %s: %v", entry.kind, heroID, err)
	}
	drainMatrixPending(t, g)
	assertGameInvariants(t, g)
}

// 全量矩阵：每个可选武将 × 基本/锦囊/装备牌种类，出牌或合法拒绝 + 牌数守恒。
func TestSmoke_HeroCardKindMatrix(t *testing.T) {
	heroes := engine.HeroesCatalog()
	catalog := cardMatrixCatalog()
	if len(heroes) == 0 {
		t.Fatal("empty hero catalog")
	}

	seenCat := map[string]int{}
	for _, e := range catalog {
		seenCat[e.category]++
	}
	for _, cat := range []string{"basic", "trick", "equip"} {
		if seenCat[cat] == 0 {
			t.Fatalf("matrix missing category %s", cat)
		}
	}

	for _, h := range heroes {
		h := h
		for _, entry := range catalog {
			entry := entry
			name := fmt.Sprintf("%s/%s/%s", h.ID, entry.category, entry.kind)
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				runHeroCardMatrixCase(t, h.ID, entry)
			})
		}
	}
}

func TestSmoke_CardMatrixCatalogCoversDeckKinds(t *testing.T) {
	catalog := cardMatrixCatalog()
	listed := make(map[string]bool, len(catalog))
	for _, e := range catalog {
		listed[e.kind] = true
	}
	deck := engine.NewBasicDeck()
	inDeck := map[string]bool{}
	for _, c := range deck {
		inDeck[c.Kind] = true
	}
	var missing []string
	for kind := range inDeck {
		if !listed[kind] {
			missing = append(missing, kind)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("matrix catalog missing deck kinds: %s", strings.Join(missing, ", "))
	}
}
