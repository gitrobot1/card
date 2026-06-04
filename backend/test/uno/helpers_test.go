package uno_test

import (
	"testing"

	uno "github.com/time/card/backend/internal/game/uno"
)

func readyGame(t *testing.T, bots int) *uno.Game {
	t.Helper()
	g, err := uno.NewSoloGame("test", "玩家", bots)
	if err != nil {
		t.Fatal(err)
	}
	g.SkipRollForFirst(0)
	g.OpeningTurn = false
	return g
}
