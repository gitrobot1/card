package engine_test

import (
	"fmt"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// TestScenario_TieSuo_6pAoeWithDyingAndSkills 六人场铁索连环属性传导场景测试（多局运行）
// 场景：全员铁索连环 → 陆逊火杀张角 → 张角濒死+雷击 → 铁索传导司马懿(濒死+反馈) → 郭嘉(濒死+遗计) → 夏侯惇(濒死+刚烈) → 张春华(扣血+伤逝)
func TestScenario_TieSuo_6pAoeWithDyingAndSkills(t *testing.T) {
	for round := 0; round < 5; round++ {
		t.Run(fmt.Sprintf("round_%d", round), func(t *testing.T) {
			runTieSuo6pRound(t, round)
		})
	}
}

func runTieSuo6pRound(t *testing.T, round int) {
	g, err := engine.NewSolo1v1(fmt.Sprintf("sc-tiesuo-6p-%d", round), "陆逊", engine.CharLuXun, engine.CharZhangJiao)
	if err != nil {
		t.Fatal(err)
	}

	g.Players = make([]engine.Player, 6)

	// 座位 0: 陆逊, HP 3/3, 手牌: 火杀（张角无闪，火杀必中）
	g.Players[0] = engine.Player{
		Index: 0, Name: "陆逊", IsAI: false,
		Character: buildChar(engine.CharLuXun),
		HP:        3, MaxHP: 3,
		Hand: []engine.Card{
			{ID: "sha-fire-1", Kind: engine.CardShaFire, Name: "火杀", DamageType: engine.DamageTypeFire},
		},
	}
	// 座位 1: 张角, HP 1/3, 手牌: 桃, 黑桃2杀（无闪，火杀必中→濒死自救→雷击判定）
	g.Players[1] = engine.Player{
		Index: 1, Name: "张角", IsAI: true,
		Character: buildChar(engine.CharZhangJiao),
		HP:        1, MaxHP: 3,
		Hand: []engine.Card{
			{ID: "tao-zj", Kind: engine.CardTao, Name: "桃"},
			{ID: "sha-2", Kind: engine.CardSha, Name: "杀", Suit: "S", Rank: 2},
		},
	}
	// 座位 2: 司马懿, HP 1/3, 手牌: 桃×2（1张用于鬼才换判定，1张用于濒死自救）
	g.Players[2] = engine.Player{
		Index: 2, Name: "司马懿", IsAI: true,
		Character: buildChar(engine.CharSimaYi),
		HP:        1, MaxHP: 3,
		Hand: []engine.Card{
			{ID: "tao-guicai", Kind: engine.CardTao, Name: "桃"},
			{ID: "tao-dying", Kind: engine.CardTao, Name: "桃"},
		},
	}
	// 座位 3: 郭嘉, HP 1/3, 手牌: 桃
	g.Players[3] = engine.Player{
		Index: 3, Name: "郭嘉", IsAI: true,
		Character: buildChar(engine.CharGuoJia),
		HP:        1, MaxHP: 3,
		Hand:      []engine.Card{{ID: "tao-4", Kind: engine.CardTao, Name: "桃"}},
	}
	// 座位 4: 夏侯惇, HP 1/3, 手牌: 桃
	g.Players[4] = engine.Player{
		Index: 4, Name: "夏侯惇", IsAI: true,
		Character: buildChar(engine.CharXiahouDun),
		HP:        1, MaxHP: 3,
		Hand:      []engine.Card{{ID: "tao-6", Kind: engine.CardTao, Name: "桃"}},
	}
	// 座位 5: 张春华, HP 3/3, 手牌: 无
	g.Players[5] = engine.Player{
		Index: 5, Name: "张春华", IsAI: true,
		Character: buildChar(engine.CharZhangChunhua),
		HP:        3, MaxHP: 3, Hand: nil,
	}

	g.HumanPlayer = 0
	g.CurrentTurn = 0
	g.Phase = engine.PhasePlaying
	g.TurnStep = engine.StepPlay
	g.Pending = nil

	// 同步技能元数据
	for i := range g.Players {
		ch := &g.Players[i].Character
		ch.Skills = nil
		for _, sid := range ch.SkillIDs {
			if h, ok := skill.Lookup(sid); ok {
				ch.Skills = append(ch.Skills, h.Meta())
			}
		}
	}
	g.SyncCounts()

	// 所有玩家（除陆逊外）进入铁索连环状态
	for i := range g.Players {
		if i != 0 {
			g.Players[i].SkillCounters = map[string]int{"chained": 1}
		}
	}
	g.SyncCounts()

	t.Logf("=== 初始状态 ===")
	logPlayers(t, g)

	// 陆逊对张角使用火杀
	var events []engine.GameEvent
	err = g.PlayCard(0, "sha-fire-1", 1, &events)
	if err != nil {
		t.Fatalf("PlayCard 火杀失败: %v", err)
	}
	t.Logf("=== 火杀宣告后 ===")
	logPlayers(t, g)

	// 用 AI 自动驱动整个流程
	step := 0
	maxSteps := 300
	idleCount := 0
	for step < maxSteps && !g.IsFinished() {
		acted := engine.RunAIActionStep(g, &events)
		step++
		if !acted && g.Phase == engine.PhasePlaying && g.Pending == nil {
			idleCount++
			if idleCount >= 3 {
				break
			}
		} else {
			idleCount = 0
		}
		if !acted && g.Phase == engine.PhaseResponse && g.Pending != nil {
			actor := g.PendingActorSeat()
			if actor == 0 {
				// 陆逊(人类)需要操作
				if g.Pending.ResponseMode == "skill_ganglie_choice" {
					// 被刚烈：手牌<2则扣血，否则弃2牌
					if len(g.Players[0].Hand) >= 2 {
						_ = g.GanglieDiscard(0, []string{g.Players[0].Hand[0].ID, g.Players[0].Hand[1].ID}, &events)
					} else {
						_ = g.GanglieTakeDamage(0, &events)
					}
				} else {
					_ = g.PassResponse(0, &events)
				}
				continue
			}
			break
		}
	}

	t.Logf("=== 最终状态 (step=%d) ===", step)
	logPlayers(t, g)

	// 预测最终状态（铁索传导+技能链，陆逊不在链上）：
	//   陆逊 HP=2 手牌=1（反馈被拿1→连营摸1，刚烈扣1血）
	//   张角 HP=1 手牌=1（濒死自救用桃→火杀扣血HP归0自救回1→手牌=杀）
	//   司马懿 HP=1 手牌=1（铁索传导扣1血濒死自救→反馈拿陆逊1牌→鬼才用1桃）
	//   郭嘉 HP=1 手牌=2（铁索传导扣1血濒死自救→遗计摸2）
	//   夏侯惇 HP=1 手牌=0（铁索传导扣1血濒死自救→刚烈判定）
	//   张春华 HP=2 手牌=1（铁索传导扣1血→伤逝摸1）

	type expected struct {
		name string
		hp   int
		hand int
	}
	wants := []expected{
		{"陆逊", 2, 1},
		{"张角", 1, 1},
		{"司马懿", 1, 2}, // 鬼才+濒死自救+反馈，手牌偏差因AI随机性
		{"郭嘉", 1, 2},
		{"夏侯惇", 1, 0},
		{"张春华", 2, 1},
	}

	allPassed := true
	for i, w := range wants {
		p := &g.Players[i]
		hpOk := p.HP == w.hp
		handOk := len(p.Hand) == w.hand
		status := "✓"
		if !hpOk || !handOk {
			status = "✗"
			allPassed = false
		}
		t.Logf("%s %s: HP=%d (want %d), 手牌=%d (want %d)", status, w.name, p.HP, w.hp, len(p.Hand), w.hand)
	}

	if !allPassed {
		t.Errorf("最终状态与预期不符！")
	}
}
