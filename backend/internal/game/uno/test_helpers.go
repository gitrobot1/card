package uno

import "testing"

func readyGame(t *testing.T, bots int) *Game {
	t.Helper()
	g, err := NewSoloGame("test", "玩家", bots)
	if err != nil {
		t.Fatal(err)
	}
	g.SkipRollForFirst(0)
	g.OpeningTurn = false
	return g
}
