package uno_test

import (
	"testing"

	uno "github.com/time/card/backend/internal/game/uno"
)

func TestNewSoloGame(t *testing.T) {
	g, err := uno.NewSoloGame("test", "玩家", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(g.Players))
	}
	if g.Phase != uno.PhaseRollForFirst {
		t.Fatalf("expected roll phase, got %s", g.Phase)
	}
	g.SkipRollForFirst(0)
	if g.OpeningTurn != true {
		t.Fatal("expected opening turn after deal")
	}
	if g.TopCard.ID != "" {
		t.Fatal("expected no starter top card")
	}
	if g.DiscardCount != 0 {
		t.Fatalf("expected empty discard, got %d", g.DiscardCount)
	}
	for _, p := range g.Players {
		if len(p.Hand) != uno.InitialHand {
			t.Fatalf("expected %d cards, got %d", uno.InitialHand, len(p.Hand))
		}
	}
}

func TestOpeningTurnAnyCard(t *testing.T) {
	g, err := uno.NewSoloGame("test", "玩家", 1)
	if err != nil {
		t.Fatal(err)
	}
	g.SkipRollForFirst(0)
	if !g.OpeningTurn {
		t.Fatal("expected opening turn")
	}
	g.CurrentTurn = 0
	g.Players[0].Hand = []uno.Card{
		{ID: "b3", Color: uno.ColorBlue, Value: "3", Label: "3"},
		{ID: "y5", Color: uno.ColorYellow, Value: "5", Label: "5"},
	}
	if !g.CanPlayCardForTest(0, g.Players[0].Hand[0]) {
		t.Fatal("opening turn should allow any card")
	}
	var events []uno.GameEvent
	if err := g.Play(0, "b3", "", &events); err != nil {
		t.Fatal(err)
	}
	if g.OpeningTurn {
		t.Fatal("opening turn should end after first play")
	}
	if g.TopCard.ID != "b3" {
		t.Fatal("expected first played card as top")
	}
}

func TestPlayMatchingColor(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.Players[0].Hand = []uno.Card{{ID: "c1", Color: uno.ColorRed, Value: "7", Label: "7"}}
	var events []uno.GameEvent
	if err := g.Play(0, "c1", "", &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != 0 {
		t.Fatalf("expected empty hand win, got %d", len(g.Players[0].Hand))
	}
	if g.Phase != uno.PhaseFinished {
		t.Fatal("expected finished")
	}
}

func TestWild4PlayableAnytime(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorBlue, Value: "3", Label: "3"}
	g.Players[0].Hand = []uno.Card{
		{ID: "w4", Color: uno.ColorWild, Value: string(uno.ValueWild4), Label: "+4"},
		{ID: "r5", Color: uno.ColorRed, Value: "5", Label: "5"},
	}
	if !g.CanPlayCardForTest(0, g.Players[0].Hand[0]) {
		t.Fatal("wild4 should be playable even when matching color exists")
	}
}

func TestDrawAdvancesTurn(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "9", Label: "9"}
	g.Players[0].Hand = []uno.Card{{ID: "b1", Color: uno.ColorBlue, Value: "1", Label: "1"}}
	beforeHand := len(g.Players[0].Hand)
	var events []uno.GameEvent
	if err := g.Draw(0, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != beforeHand+1 {
		t.Fatalf("expected +1 card, got %d", len(g.Players[0].Hand)-beforeHand)
	}
	if g.CurrentTurn == 0 {
		t.Fatal("turn should advance after draw")
	}
	if len(events) != 1 || events[0].Type != uno.EventDraw {
		t.Fatalf("expected draw event, got %v", events)
	}
}

func TestDraw2Stack(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.Players[0].Hand = []uno.Card{
		{ID: "d2a", Color: uno.ColorRed, Value: string(uno.ValueDraw2), Label: "+2"},
		{ID: "x1", Color: uno.ColorRed, Value: "3", Label: "3"},
	}
	g.Players[1].Hand = []uno.Card{
		{ID: "d2b", Color: uno.ColorBlue, Value: string(uno.ValueDraw2), Label: "+2"},
		{ID: "x2", Color: uno.ColorBlue, Value: "4", Label: "4"},
	}
	var events []uno.GameEvent
	if err := g.Play(0, "d2a", "", &events); err != nil {
		t.Fatal(err)
	}
	if g.PendingDrawPenalty != 2 || g.CurrentTurn != 1 {
		t.Fatalf("expected pending 2 on seat 1, got %d turn %d", g.PendingDrawPenalty, g.CurrentTurn)
	}
	if err := g.Play(1, "d2b", "", &events); err != nil {
		t.Fatal(err)
	}
	if g.PendingDrawPenalty != 4 || g.CurrentTurn != 0 {
		t.Fatalf("expected pending 4 on seat 0, got %d turn %d", g.PendingDrawPenalty, g.CurrentTurn)
	}
	before := len(g.Players[0].Hand)
	if err := g.Draw(0, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != before+4 {
		t.Fatalf("expected +4 cards, got %d", len(g.Players[0].Hand)-before)
	}
	if g.PendingDrawPenalty != 0 {
		t.Fatal("stack should clear")
	}
	if len(events) < 2 {
		t.Fatalf("expected play+draw events, got %v", events)
	}
	drawEv := events[len(events)-1]
	if drawEv.Type != uno.EventDraw || drawEv.Card == nil || drawEv.Amount != 4 {
		t.Fatalf("expected stack draw event with card, got %+v", drawEv)
	}
	if g.CurrentTurn != 0 {
		t.Fatal("turn should stay after accepting stack draw")
	}
	if !g.MustPlayAfterStack {
		t.Fatal("expected must play after stack")
	}
}

func TestLastActionCardCannotPlay(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.Players[0].Hand = []uno.Card{
		{ID: "sk", Color: uno.ColorRed, Value: string(uno.ValueSkip), Label: "跳过"},
	}
	if g.CanPlayCardForTest(0, g.Players[0].Hand[0]) {
		t.Fatal("last action card should not be playable even when matching")
	}
	var events []uno.GameEvent
	if err := g.Play(0, "sk", "", &events); err != uno.ErrLastCardMustBeBasic {
		t.Fatalf("expected ErrLastCardMustBeBasic, got %v", err)
	}
}

func TestLastActionCardCannotStackDraw2(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.PendingDrawPenalty = 2
	g.Players[0].Hand = []uno.Card{
		{ID: "d2", Color: uno.ColorRed, Value: string(uno.ValueDraw2), Label: "+2"},
	}
	if g.CanPlayCardForTest(0, g.Players[0].Hand[0]) {
		t.Fatal("last +2 should not stack when it is the only card")
	}
}

func TestVoluntaryDrawWithPlayableCards(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.Players[0].Hand = []uno.Card{
		{ID: "r7", Color: uno.ColorRed, Value: "7", Label: "7"},
		{ID: "sk", Color: uno.ColorBlue, Value: string(uno.ValueSkip), Label: "跳过"},
	}
	before := len(g.Players[0].Hand)
	var events []uno.GameEvent
	if err := g.Draw(0, &events); err != nil {
		t.Fatalf("expected voluntary draw with playable cards, got %v", err)
	}
	if len(g.Players[0].Hand) != before+1 {
		t.Fatalf("expected +1 card, got %d", len(g.Players[0].Hand))
	}
	if g.CurrentTurn == 0 {
		t.Fatal("turn should advance after voluntary draw")
	}
}

func TestStackDrawThenPlay(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.PendingDrawPenalty = 2
	g.Players[0].Hand = []uno.Card{
		{ID: "r7", Color: uno.ColorRed, Value: "7", Label: "7"},
	}
	var events []uno.GameEvent
	if err := g.Draw(0, &events); err != nil {
		t.Fatal(err)
	}
	if !g.MustPlayAfterStack || g.CurrentTurn != 0 {
		t.Fatalf("expected human still playing after stack draw, turn=%d must=%v", g.CurrentTurn, g.MustPlayAfterStack)
	}
	if err := g.Draw(0, &events); err != uno.ErrCannotPlay {
		t.Fatalf("expected cannot draw while playable post-stack, got %v", err)
	}
	if err := g.Play(0, "r7", "", &events); err != nil {
		t.Fatal(err)
	}
	if g.MustPlayAfterStack {
		t.Fatal("flag should clear after play")
	}
	if g.CurrentTurn == 0 {
		t.Fatal("turn should advance after playing post-stack")
	}
}

func TestStackDrawThenVoluntaryDraw(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.PendingDrawPenalty = 4
	g.Players[0].Hand = []uno.Card{
		{ID: "g8", Color: uno.ColorGreen, Value: "8", Label: "8"},
	}
	var events []uno.GameEvent
	if err := g.Draw(0, &events); err != nil {
		t.Fatal(err)
	}
	if !g.MustPlayAfterStack {
		t.Fatal("expected must play after stack")
	}
	// 罚牌后强制设为无牌可出，验证「无牌可出则摸牌并结束回合」
	g.Players[0].Hand = []uno.Card{
		{ID: "g8", Color: uno.ColorGreen, Value: "8", Label: "8"},
		{ID: "g9", Color: uno.ColorGreen, Value: "9", Label: "9"},
	}
	g.SyncCountsForTest()
	if len(g.PlayableCards(0)) > 0 {
		t.Fatal("test setup: expected no playable cards")
	}
	before := len(g.Players[0].Hand)
	if err := g.Draw(0, &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[0].Hand) != before+1 {
		t.Fatalf("expected one voluntary draw, got %d cards", len(g.Players[0].Hand))
	}
	if g.CurrentTurn == 0 {
		t.Fatal("turn should end after voluntary draw post-stack")
	}
	if g.MustPlayAfterStack {
		t.Fatal("flag should clear after ending turn")
	}
}

func TestWild4StackBlocksDraw2(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 1
	g.PendingDrawPenalty = 4
	g.DrawStackWild4Only = true
	g.Players[1].Hand = []uno.Card{
		{ID: "d2", Color: uno.ColorRed, Value: string(uno.ValueDraw2), Label: "+2"},
		{ID: "w4", Color: uno.ColorWild, Value: string(uno.ValueWild4), Label: "+4"},
	}
	if g.CanPlayCardForTest(1, g.Players[1].Hand[0]) {
		t.Fatal("+2 cannot stack on +4")
	}
	if !g.CanPlayCardForTest(1, g.Players[1].Hand[1]) {
		t.Fatal("+4 should stack on +4")
	}
}

func TestDraw2StackAllowsWild4(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 1
	g.PendingDrawPenalty = 2
	g.DrawStackWild4Only = false
	g.Players[1].Hand = []uno.Card{
		{ID: "d2", Color: uno.ColorRed, Value: string(uno.ValueDraw2), Label: "+2"},
		{ID: "w4", Color: uno.ColorWild, Value: string(uno.ValueWild4), Label: "+4"},
	}
	if !g.CanPlayCardForTest(1, g.Players[1].Hand[0]) {
		t.Fatal("+2 should stack on +2")
	}
	if !g.CanPlayCardForTest(1, g.Players[1].Hand[1]) {
		t.Fatal("+4 should stack on +2 chain")
	}
}

func TestSkipForcesNextDraw(t *testing.T) {
	g := readyGame(t, 1)
	g.CurrentTurn = 0
	g.CurrentColor = uno.ColorRed
	g.TopCard = uno.Card{Color: uno.ColorRed, Value: "5", Label: "5"}
	g.Players[0].Hand = []uno.Card{
		{ID: "sk", Color: uno.ColorRed, Value: string(uno.ValueSkip), Label: "跳过"},
		{ID: "k", Color: uno.ColorRed, Value: "3", Label: "3"},
	}
	before := len(g.Players[1].Hand)
	var events []uno.GameEvent
	if err := g.Play(0, "sk", "", &events); err != nil {
		t.Fatal(err)
	}
	if len(g.Players[1].Hand) != before+1 {
		t.Fatalf("victim should draw 1, hand %d -> %d", before, len(g.Players[1].Hand))
	}
	if g.CurrentTurn != 0 {
		t.Fatalf("turn should return to player 0 in 2p, got %d", g.CurrentTurn)
	}
	if len(events) != 2 || events[0].Type != uno.EventPlay || events[1].Type != uno.EventDraw {
		t.Fatalf("expected play+draw events, got %v", events)
	}
}
