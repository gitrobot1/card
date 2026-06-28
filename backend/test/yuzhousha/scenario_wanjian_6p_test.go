package engine_test

import (
	"fmt"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// TestScenario_WanJian_6pAoeWithDyingAndSkills 六人场万箭齐发场景测试（多局运行）
// 场景：陆逊放万箭 → 张角出闪(雷击) → 司马懿濒死+反馈 → 郭嘉濒死+遗计 → 张春华扣血+伤逝 → 夏侯惇濒死+刚烈
func TestScenario_WanJian_6pAoeWithDyingAndSkills(t *testing.T) {
	// 跑多局，因为鬼才可能把刚烈判定改成红桃导致刚烈不生效
	// 多跑几局总会遇到司马懿手里没有红桃的情况
	for round := 0; round < 5; round++ {
		t.Run(fmt.Sprintf("round_%d", round), func(t *testing.T) {
			runWanJian6pRound(t, round)
		})
	}
}

func runWanJian6pRound(t *testing.T, round int) {
	g, err := engine.NewSolo1v1(fmt.Sprintf("sc-wanjian-6p-%d", round), "陆逊", engine.CharLuXun, engine.CharZhangJiao)
	if err != nil {
		t.Fatal(err)
	}

	g.Players = make([]engine.Player, 6)

	// 座位 0: 陆逊, HP 3/3, 手牌: 万箭齐发
	g.Players[0] = engine.Player{
		Index: 0, Name: "陆逊", IsAI: false,
		Character: buildChar(engine.CharLuXun),
		HP: 3, MaxHP: 3,
		Hand: []engine.Card{{ID: "wanjian-1", Kind: engine.CardWanJian, Name: "万箭齐发"}},
	}
	// 座位 1: 张角, HP 3/3, 手牌: 闪, 黑桃2杀
	g.Players[1] = engine.Player{
		Index: 1, Name: "张角", IsAI: true,
		Character: buildChar(engine.CharZhangJiao),
		HP: 3, MaxHP: 3,
		Hand: []engine.Card{
			{ID: "shan-1", Kind: engine.CardShan, Name: "闪"},
			{ID: "sha-2", Kind: engine.CardSha, Name: "杀", Suit: "S", Rank: 2},
		},
	}
	// 座位 2: 司马懿, HP 1/3, 手牌: 桃×2（1张用于鬼才换判定，1张用于濒死自救）
	g.Players[2] = engine.Player{
		Index: 2, Name: "司马懿", IsAI: true,
		Character: buildChar(engine.CharSimaYi),
		HP: 1, MaxHP: 3,
		Hand: []engine.Card{
			{ID: "tao-guicai", Kind: engine.CardTao, Name: "桃"},
			{ID: "tao-dying", Kind: engine.CardTao, Name: "桃"},
		},
	}
	// 座位 3: 郭嘉, HP 1/3, 手牌: 桃
	g.Players[3] = engine.Player{
		Index: 3, Name: "郭嘉", IsAI: true,
		Character: buildChar(engine.CharGuoJia),
		HP: 1, MaxHP: 3,
		Hand: []engine.Card{{ID: "tao-4", Kind: engine.CardTao, Name: "桃"}},
	}
	// 座位 4: 夏侯惇, HP 1/3, 手牌: 桃
	g.Players[4] = engine.Player{
		Index: 4, Name: "夏侯惇", IsAI: true,
		Character: buildChar(engine.CharXiahouDun),
		HP: 1, MaxHP: 3,
		Hand: []engine.Card{{ID: "tao-6", Kind: engine.CardTao, Name: "桃"}},
	}
	// 座位 5: 张春华, HP 3/3, 手牌: 无
	g.Players[5] = engine.Player{
		Index: 5, Name: "张春华", IsAI: true,
		Character: buildChar(engine.CharZhangChunhua),
		HP: 3, MaxHP: 3, Hand: nil,
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

	t.Logf("=== 初始状态 ===")
	logPlayers(t, g)

	// 陆逊打出万箭齐发
	var events []engine.GameEvent
	err = g.PlayCard(0, "wanjian-1", 0, &events)
	if err != nil {
		t.Fatalf("PlayCard 万箭失败: %v", err)
	}
	t.Logf("=== 万箭宣告后（陆逊连营） ===")
	logPlayers(t, g)

	// 用 AI 自动驱动整个万箭流程
	step := 0
	maxSteps := 300
	idleCount := 0
	for step < maxSteps && !g.IsFinished() {
		acted := engine.RunAIActionStep(g, &events)
		step++
		// 连续多步无动作且 Phase=Playing 时退出（万箭流程已结束）
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

	// 验证关键事件是否发生
	// 预测最终状态：
	//   陆逊 HP=2 手牌=1（反馈被拿1→连营摸1，刚烈受1伤）
	//   张角 HP=3 手牌=1（出闪剩黑桃2杀）
	//   司马懿 HP=1 手牌=1（鬼才用1桃→濒死用1桃→反馈拿陆逊1牌）
	//   郭嘉 HP=1 手牌=2（濒死用桃→遗计摸2）
	//   张春华 HP=2 手牌=1（扣1血→伤逝摸1）
	//   夏侯惇 HP=1 手牌=0（濒死用桃→刚烈判定）

	type expected struct {
		name string
		hp   int
		hand int
	}
	wants := []expected{
		{"陆逊", 2, 1},
		{"张角", 3, 1},
		{"司马懿", 1, 1},
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

func buildChar(charID string) engine.Character {
	def, ok := skill.CharacterByID(charID)
	if !ok {
		panic(fmt.Sprintf("unknown character: %s", charID))
	}
	return engine.Character{
		ID: def.ID, Name: def.Name, MaxHP: def.MaxHP, SkillIDs: def.SkillIDs,
	}
}

func logPlayers(t *testing.T, g *engine.Game) {
	t.Helper()
	for i := range g.Players {
		p := &g.Players[i]
		handIDs := make([]string, len(p.Hand))
		for j, c := range p.Hand {
			handIDs[j] = c.Name
		}
		t.Logf("  [%d] %s HP=%d/%d 手牌=%v chained=%v",
			i, p.Name, p.HP, p.MaxHP, handIDs, p.SkillCounters["chained"] > 0)
	}
}
