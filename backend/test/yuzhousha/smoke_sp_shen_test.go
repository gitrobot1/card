package engine_test

import (
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func TestSmoke_SpShenZhaoYun_Bootstrap(t *testing.T) {
	cases := []struct {
		id   string
		name string
	}{
		{skill.CharSpZhaoYun, "SP赵云"},
		{skill.CharShenZhaoYun, "神赵云"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.id, func(t *testing.T) {
			g, err := engine.NewSolo1v1("smoke-"+tc.id, "玩家", tc.id, engine.CharGuanYu)
			if err != nil {
				t.Fatal(err)
			}
			if g.Players[0].Character.ID != tc.id {
				t.Fatalf("seat0 hero=%s want %s", g.Players[0].Character.ID, tc.id)
			}
			if len(g.Players[0].Character.Skills) == 0 {
				t.Fatal("expected skills on new hero")
			}
			assertGameInvariants(t, g)
		})
	}
}

func TestSmoke_ShenZhaoYun_JuejingDrawBonus(t *testing.T) {
	g, err := engine.NewSolo1v1("smoke-shen-draw", "玩家", skill.CharShenZhaoYun, engine.CharGuanYu)
	if err != nil {
		t.Fatal(err)
	}
	g.Players[0].HP = 1
	if got := g.DrawCountForTest(0); got != engine.DrawPerTurn+2 {
		t.Fatalf("juejing draw=%d want %d", got, engine.DrawPerTurn+2)
	}
}
