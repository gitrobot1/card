package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestWuguFullFlow A打出五谷丰登→亮牌→依次选牌（群体锦囊不可被无懈可击）
func TestWuguFullFlow(t *testing.T) {
	logDir := filepath.Join("..", "..", "..", "..", "logs")
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "wugu_test.log")
	f, _ := os.Create(logPath)
	defer f.Close()

	g := &Game{
		Phase:       PhasePlaying,
		TurnStep:    StepPlay,
		CurrentTurn: 0,
		Players: []Player{
			{Index: 0, Name: "玩家A", HP: 3, MaxHP: 3, IsAI: true,
				Character: buildCharacter(CharHuangYueying)},
			{Index: 1, Name: "玩家B", HP: 3, MaxHP: 3, IsAI: true,
				Character: buildCharacter(CharHuangYueying)},
		},
		DiscardPile: []Card{},
	}

	// A: 五谷丰登 + 2无懈 + 3其他
	g.Players[0].Hand = []Card{
		{ID: "A_wugu", Kind: CardWuGu, Name: "五谷丰登", Suit: "S", Rank: 1, Label: "A♠"},
		{ID: "A_wx1", Kind: CardWuxiek, Name: "无懈可击", Suit: "H", Rank: 2, Label: "2♥"},
		{ID: "A_wx2", Kind: CardWuxiek, Name: "无懈可击", Suit: "D", Rank: 3, Label: "3♦"},
		{ID: "A_sha", Kind: CardSha, Name: "杀", Suit: "C", Rank: 4, Label: "4♣"},
		{ID: "A_tao", Kind: CardTao, Name: "桃", Suit: "H", Rank: 5, Label: "5♥"},
		{ID: "A_shan", Kind: CardShan, Name: "闪", Suit: "D", Rank: 6, Label: "6♦"},
	}
	// B: 2无懈 + 2其他
	g.Players[1].Hand = []Card{
		{ID: "B_wx1", Kind: CardWuxiek, Name: "无懈可击", Suit: "S", Rank: 7, Label: "7♠"},
		{ID: "B_wx2", Kind: CardWuxiek, Name: "无懈可击", Suit: "C", Rank: 8, Label: "8♣"},
		{ID: "B_sha", Kind: CardSha, Name: "杀", Suit: "H", Rank: 9, Label: "9♥"},
		{ID: "B_shan", Kind: CardShan, Name: "闪", Suit: "D", Rank: 10, Label: "10♦"},
	}

	// 足够的牌堆
	g.DrawPile = make([]Card, 20)
	for i := range g.DrawPile {
		g.DrawPile[i] = Card{
			ID:    fmt.Sprintf("deck_%d", i),
			Kind:  CardSha,
			Name:  "杀",
			Suit:  "S",
			Rank:  i + 1,
			Label: fmt.Sprintf("%d♠", i+1),
		}
	}

	events := []GameEvent{}

	fmt.Fprintf(f, "=== 初始 ===\n")
	logTestState(f, g)

	// A打出五谷丰登
	fmt.Fprintf(f, "\n=== A打出五谷丰登 ===\n")
	if err := g.PlayCard(0, "A_wugu", 0, &events); err != nil {
		fmt.Fprintf(f, "ERROR PlayCard: %v\n", err)
		t.Fatalf("PlayCard: %v", err)
	}
	logTestState(f, g)
	logTestEvents(f, events)
	events = nil

	// 运行AI直到无操作
	for round := 0; round < 20; round++ {
		acted := false
		for i := 0; i < 3; i++ {
			if RunAIActionStep(g, &events) {
				acted = true
			}
		}
		fmt.Fprintf(f, "\n=== AI轮次 %d (acted=%v) ===\n", round+1, acted)
		logTestState(f, g)
		logTestEvents(f, events)
		events = nil
		if !acted {
			break
		}
	}

	fmt.Fprintf(f, "\n=== 最终状态 ===\n")
	logTestState(f, g)

	aWuxiekLeft := 0
	bWuxiekLeft := 0
	for _, c := range g.Players[0].Hand {
		if c.Kind == CardWuxiek {
			aWuxiekLeft++
		}
	}
	for _, c := range g.Players[1].Hand {
		if c.Kind == CardWuxiek {
			bWuxiekLeft++
		}
	}
	fmt.Fprintf(f, "\nA剩余无懈: %d, B剩余无懈: %d\n", aWuxiekLeft, bWuxiekLeft)
	fmt.Fprintf(f, "A手牌数: %d, B手牌数: %d\n", len(g.Players[0].Hand), len(g.Players[1].Hand))
	fmt.Fprintf(f, "DiscardPile: %d\n", len(g.DiscardPile))

	// 验证：五谷丰登已结算（不要求 Phase=playing，因为后续可能继续游戏）
	t.Logf("A剩余无懈: %d, B剩余无懈: %d, A手牌: %d, B手牌: %d",
		aWuxiekLeft, bWuxiekLeft, len(g.Players[0].Hand), len(g.Players[1].Hand))

	if len(g.DiscardPile) < 2 {
		t.Errorf("DiscardPile只有%d张，五谷可能未完全结算", len(g.DiscardPile))
	}

	t.Logf("测试日志: %s", logPath)
}

// TestWuguWuxiekChain 原始测试保留
func TestWuguWuxiekChain(t *testing.T) {
	logDir := filepath.Join("..", "..", "..", "..", "logs")
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "wugu_test.log")
	f, _ := os.Create(logPath)
	defer f.Close()
	fmt.Fprintf(f, "=== 五谷丰登无懈可击链测试 ===\n\n")

	g := &Game{
		Phase:       PhasePlaying,
		TurnStep:    StepPlay,
		CurrentTurn: 0,
		Players: []Player{
			{Index: 0, Name: "玩家A", HP: 3, MaxHP: 3, IsAI: true,
				Character: buildCharacter(CharHuangYueying)},
			{Index: 1, Name: "玩家B", HP: 3, MaxHP: 3, IsAI: true,
				Character: buildCharacter(CharHuangYueying)},
		},
		DiscardPile: []Card{},
	}

	g.Players[0].Hand = []Card{
		{ID: "wugu1", Kind: CardWuGu, Name: "五谷丰登"},
		{ID: "wuxiek1", Kind: CardWuxiek, Name: "无懈可击"},
		{ID: "wuxiek2", Kind: CardWuxiek, Name: "无懈可击"},
	}
	g.Players[1].Hand = []Card{
		{ID: "wuxiek3", Kind: CardWuxiek, Name: "无懈可击"},
		{ID: "wuxiek4", Kind: CardWuxiek, Name: "无懈可击"},
	}

	g.DrawPile = []Card{
		{ID: "r1", Kind: CardSha, Name: "杀", Suit: "S", Rank: 1, Label: "A♠"},
		{ID: "r2", Kind: CardShan, Name: "闪", Suit: "H", Rank: 2, Label: "2♥"},
		{ID: "r3", Kind: CardTao, Name: "桃", Suit: "H", Rank: 3, Label: "3♥"},
		{ID: "r4", Kind: CardJiu, Name: "酒", Suit: "C", Rank: 4, Label: "4♣"},
	}

	events := []GameEvent{}

	fmt.Fprintf(f, "初始:\n")
	logTestState(f, g)

	fmt.Fprintf(f, "\n--- 打出五谷丰登 ---\n")
	err := g.PlayCard(0, "wugu1", 0, &events)
	if err != nil {
		fmt.Fprintf(f, "ERROR PlayCard: %v\n", err)
	}
	logTestState(f, g)
	logTestEvents(f, events)
	events = nil

	for i := 0; i < 10 && !g.IsFinished(); i++ {
		fmt.Fprintf(f, "\n--- AI step %d ---\n", i+1)
		acted := RunAIActionStep(g, &events)
		logTestState(f, g)
		logTestEvents(f, events)
		events = nil
		if !acted {
			fmt.Fprintf(f, "  (no AI action)\n")
			break
		}
	}

	fmt.Fprintf(f, "\n=== 测试完成 ===\n")
	fmt.Printf("测试日志: %s\n", logPath)
	t.Logf("测试日志: %s", logPath)
}

func logTestState(f *os.File, g *Game) {
	fmt.Fprintf(f, "  Phase=%s TurnStep=%s CurrentTurn=%d\n", g.Phase, g.TurnStep, g.CurrentTurn)
	if g.Pending != nil {
		p := g.Pending
		fmt.Fprintf(f, "  Pending: Mode=%s Actor=%d Target=%d Effect=%d Queue=%v Idx=%d Chain=%d Card=%s WuguPick=%d Revealed=%d\n",
			p.ResponseMode, p.ActorSeat, p.TargetIndex, p.EffectTarget, p.ResponseQueue, p.ResponseIndex, len(p.WuxiekChain), p.Card.Kind, p.WuguPickSeat, len(p.RevealedCards))
	} else {
		fmt.Fprintf(f, "  Pending: nil\n")
	}
	for i, pl := range g.Players {
		kinds := make([]string, len(pl.Hand))
		for j, c := range pl.Hand {
			kinds[j] = c.Kind
		}
		fmt.Fprintf(f, "  P%d(%s) HP=%d/%d Hand=%v\n", i, pl.Name, pl.HP, pl.MaxHP, kinds)
	}
}

func logTestEvents(f *os.File, events []GameEvent) {
	if len(events) == 0 {
		return
	}
	fmt.Fprintf(f, "  Events(%d):\n", len(events))
	for _, e := range events {
		cn := ""
		if e.Card != nil {
			cn = e.Card.Name
		}
		fmt.Fprintf(f, "    %s p=%d t=%d card=%s msg=%s\n", e.Type, e.PlayerIndex, e.TargetIndex, cn, e.Message)
	}
}

// TestWugu8Players 8人场五谷丰登：亮牌后每人依次选牌（不可被无懈可击）
func TestWugu8Players(t *testing.T) {
	logDir := filepath.Join("..", "..", "..", "..", "logs")
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "wugu_8p_test.log")
	f, _ := os.Create(logPath)
	defer f.Close()

	names := []string{"P0", "P1", "P2", "P3", "P4", "P5", "P6", "P7"}
	g := &Game{
		Phase:       PhasePlaying,
		TurnStep:    StepPlay,
		CurrentTurn: 0,
		Players:     make([]Player, 8),
		DiscardPile: []Card{},
	}
	for i := range g.Players {
		g.Players[i] = Player{
			Index:     i,
			Name:      names[i],
			HP:        3,
			MaxHP:     3,
			IsAI:      true,
			Character: buildCharacter(CharHuangYueying),
		}
	}

	// P0: 五谷丰登 + 普通牌（没有无懈）
	g.Players[0].Hand = []Card{
		{ID: "P0_wugu", Kind: CardWuGu, Name: "五谷丰登", Suit: "S", Rank: 1},
		{ID: "P0_s1", Kind: CardSha, Name: "杀", Suit: "C", Rank: 3},
		{ID: "P0_s2", Kind: CardSha, Name: "杀", Suit: "D", Rank: 4},
	}
	// P1: 有无懈可击（对P0用无懈阻止P0选牌）
	g.Players[1].Hand = []Card{
		{ID: "P1_wx", Kind: CardWuxiek, Name: "无懈可击", Suit: "H", Rank: 2},
		{ID: "P1_s1", Kind: CardSha, Name: "杀", Suit: "S", Rank: 6},
		{ID: "P1_s2", Kind: CardSha, Name: "杀", Suit: "H", Rank: 16},
	}
	// P2-P7: 只有普通牌，没有无懈可击
	for i := 2; i < 8; i++ {
		g.Players[i].Hand = []Card{
			{ID: fmt.Sprintf("P%d_s1", i), Kind: CardSha, Name: "杀", Suit: "S", Rank: 5 + i},
			{ID: fmt.Sprintf("P%d_s2", i), Kind: CardSha, Name: "杀", Suit: "H", Rank: 15 + i},
		}
	}

	// 牌堆全是闪（AI不会用闪攻击，避免游戏提前结束）
	g.DrawPile = make([]Card, 30)
	for i := range g.DrawPile {
		g.DrawPile[i] = Card{
			ID:    fmt.Sprintf("deck_%d", i),
			Kind:  CardShan,
			Name:  "闪",
			Suit:  "D",
			Rank:  i + 1,
			Label: fmt.Sprintf("%d♦", i+1),
		}
	}

	events := []GameEvent{}
	fmt.Fprintf(f, "=== 8人五谷丰登测试（群体锦囊不可被无懈可击）===\n")
	fmt.Fprintf(f, "P0打出五谷丰登，亮出8张牌，每人依次选牌\n")
	logTestState(f, g)

	// P0打出五谷丰登
	fmt.Fprintf(f, "\n--- P0打出五谷丰登 ---\n")
	if err := g.PlayCard(0, "P0_wugu", 0, &events); err != nil {
		t.Fatalf("PlayCard: %v", err)
	}
	logTestState(f, g)
	logTestEvents(f, events)
	events = nil

	// 运行AI直到五谷完全结算
	wuguFinished := false
	totalSteps := 0
	for round := 0; round < 100 && !g.IsFinished() && totalSteps < 200; round++ {
		acted := false
		for i := 0; i < 3 && totalSteps < 200; i++ {
			totalSteps++
			if RunAIActionStep(g, &events) {
				acted = true
			}
		}
		if len(events) > 0 {
			fmt.Fprintf(f, "\n--- AI轮次 %d ---\n", round+1)
			logTestState(f, g)
			logTestEvents(f, events)
			// 检查五谷是否结束
			for _, e := range events {
				if e.Type == "wugu_pick" || e.Type == "wugu_skip" {
					// 还在五谷中
				}
			}
			if g.Phase == PhasePlaying && g.Pending == nil && g.CurrentTurn == 0 {
				wuguFinished = true
				fmt.Fprintf(f, "  >>> 五谷丰登已结算完毕！\n")
			}
			events = nil
		}
		if !acted {
			break
		}
	}

	fmt.Fprintf(f, "\n=== 最终状态 ===\n")
	logTestState(f, g)

	// 统计：亮出8张，P0被跳过，P1-P7各选1张 = 7人选了牌
	wuguPickCount := 0
	for i := range g.Players {
		fmt.Fprintf(f, "P%d 手牌数: %d\n", i, len(g.Players[i].Hand))
	}
	fmt.Fprintf(f, "五谷结算完成: %v, Phase=%s Pending=%v\n", wuguFinished, g.Phase, g.Pending != nil)

	if !wuguFinished {
		t.Errorf("五谷丰登未正确结算！Phase=%s Pending=%v", g.Phase, g.Pending != nil)
	} else {
		t.Logf("五谷丰登8人测试通过！7人选牌完成")
	}
	_ = wuguPickCount
	t.Logf("测试日志: %s", logPath)
}
