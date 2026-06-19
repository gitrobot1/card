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
	// 绝境：进入濒死状态时摸一张牌
	// 设 HP=1，手牌清空
	g.Players[0].HP = 1
	g.Players[0].Hand = nil
	g.SyncCounts()

	// 受到1点伤害，进入濒死状态（HP -> 0）
	var events []engine.GameEvent
	g.ApplyDamageForTest(0, 0, 1, "", "", &events)

	// 检查绝境事件是否产生（进入濒死时摸一张牌）
	hasJuejingEvent := false
	for _, ev := range events {
		if ev.SkillID == "juejing" {
			hasJuejingEvent = true
			break
		}
	}
	if !hasJuejingEvent {
		t.Fatal("expected juejing skill event when entering dying")
	}
}
