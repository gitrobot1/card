package zhajinhua

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/time/card/backend/internal/game/card"
)

var (
	ErrNotYourTurn    = errors.New("not your turn")
	ErrWrongPhase     = errors.New("wrong phase")
	ErrInvalidAction  = errors.New("invalid action")
	ErrCompareNeedLook  = errors.New("compare need look")
	ErrCompareTargetNeedLook = errors.New("compare target need look")
	ErrTargetInvalid  = errors.New("invalid compare target")
	ErrInsufficientChips = errors.New("insufficient chips")
	ErrGameOver       = errors.New("game over")
)

type Player struct {
	Index      int         `json:"index"`
	Name       string      `json:"name"`
	IsAI       bool        `json:"is_ai"`
	Hand       []card.Card `json:"-"`
	Looked     bool        `json:"looked"`
	Folded     bool        `json:"folded"`
	Chips      int         `json:"chips"`
	BetRound   int         `json:"bet_round"`
	TotalBet   int         `json:"total_bet"`
	HandType   string      `json:"hand_type,omitempty"`
	HandLabel  string      `json:"hand_label,omitempty"`
	Multiplier int         `json:"multiplier,omitempty"`
}

type PlayerPublic struct {
	Index      int          `json:"index"`
	Name       string       `json:"name"`
	IsAI       bool         `json:"is_ai"`
	Looked     bool         `json:"looked"`
	Folded     bool         `json:"folded"`
	Chips      int          `json:"chips"`
	BetRound   int          `json:"bet_round"`
	TotalBet   int          `json:"total_bet"`
	HandType   string       `json:"hand_type,omitempty"`
	HandLabel  string       `json:"hand_label,omitempty"`
	Multiplier int          `json:"multiplier,omitempty"`
	Hand       []card.Card  `json:"hand,omitempty"`
	CardCount  int          `json:"card_count"`
}

type Game struct {
	ID             string      `json:"id"`
	Phase          string      `json:"phase"`
	Players        []Player    `json:"players"`
	HumanPlayer    int         `json:"human_player"`
	DealerIndex    int         `json:"dealer_index"`
	CurrentTurn    int         `json:"current_turn"`
	Pot            int         `json:"pot"`
	CurrentBet     int         `json:"current_bet"`
	BaseAnte       int         `json:"base_ante"`
	MinRaise       int         `json:"min_raise"`
	WinnerIndex    *int        `json:"winner_index,omitempty"`
	WinMultiplier  int         `json:"win_multiplier,omitempty"`
	WinHandLabel   string      `json:"win_hand_label,omitempty"`
	Message        string      `json:"message"`
	TurnDeadline   time.Time   `json:"-"`
	Events         []GameEvent `json:"events,omitempty"`
	lastAggressor     int
	hasActedThisRound map[int]bool
}

type PublicState struct {
	ID                string         `json:"id"`
	Phase             string         `json:"phase"`
	Players           []PlayerPublic `json:"players"`
	HumanPlayer       int            `json:"human_player"`
	DealerIndex       int            `json:"dealer_index"`
	CurrentTurn       int            `json:"current_turn"`
	Pot               int            `json:"pot"`
	CurrentBet        int            `json:"current_bet"`
	BaseAnte          int            `json:"base_ante"`
	MinRaise          int            `json:"min_raise"`
	CompareCost       int            `json:"compare_cost"`
	WinnerIndex       *int           `json:"winner_index,omitempty"`
	WinMultiplier     int            `json:"win_multiplier,omitempty"`
	WinHandLabel      string         `json:"win_hand_label,omitempty"`
	Message           string         `json:"message"`
	MyHand            []card.Card    `json:"my_hand,omitempty"`
	HandMultipliers   map[string]int `json:"hand_multipliers"`
	TurnDeadlineUnix  int64          `json:"turn_deadline_unix"`
	Events            []GameEvent    `json:"events,omitempty"`
}

func NewSoloGame(id, humanName string, botCount int) (*Game, error) {
	if botCount < 1 || botCount > MaxPlayers-1 {
		return nil, fmt.Errorf("bot count must be 1-%d", MaxPlayers-1)
	}
	total := botCount + 1
	names := make([]string, total)
	names[0] = humanName
	for i := 1; i < total; i++ {
		names[i] = fmt.Sprintf("电脑%d", i)
	}
	isAI := make([]bool, total)
	for i := 1; i < total; i++ {
		isAI[i] = true
	}
	return newGame(id, names, isAI, 0)
}

func NewOnlineGame(id string, names []string) (*Game, error) {
	if len(names) < MinPlayers || len(names) > MaxPlayers {
		return nil, fmt.Errorf("player count must be %d-%d", MinPlayers, MaxPlayers)
	}
	isAI := make([]bool, len(names))
	return newGame(id, names, isAI, 0)
}

func newGame(id string, names []string, isAI []bool, humanSeat int) (*Game, error) {
	g := &Game{
		ID:              id,
		Phase:           PhaseBetting,
		HumanPlayer:     humanSeat,
		DealerIndex:     0,
		BaseAnte:        DefaultAnte,
		MinRaise:        DefaultMinRaise,
		lastAggressor:     -1,
		hasActedThisRound: make(map[int]bool),
	}
	for i, name := range names {
		g.Players = append(g.Players, Player{
			Index: i,
			Name:  name,
			IsAI:  isAI[i],
			Chips: DefaultChips,
		})
	}
	g.dealAndPostAnte()
	g.CurrentTurn = (g.DealerIndex + 1) % len(g.Players)
	g.resetTurnTimer()
	g.Message = fmt.Sprintf("%s 开始行动", g.Players[g.CurrentTurn].Name)
	return g, nil
}

func (g *Game) dealAndPostAnte() {
	deck := card.ShuffleRandom(card.NewDeck52())
	need := len(g.Players) * 3
	if len(deck) < need {
		return
	}
	for i := range g.Players {
		g.Players[i].Hand = append([]card.Card(nil), deck[i*3:(i+1)*3]...)
		g.payAnte(i)
	}
}

func (g *Game) payAnte(i int) {
	ante := g.BaseAnte
	if g.Players[i].Chips < ante {
		ante = g.Players[i].Chips
	}
	g.Players[i].Chips -= ante
	g.Players[i].BetRound = ante
	g.Players[i].TotalBet = ante
	g.Pot += ante
	if g.CurrentBet < ante {
		g.CurrentBet = ante
	}
}

func (g *Game) Look(index int, events *[]GameEvent) error {
	if g.Phase != PhaseBetting {
		return ErrWrongPhase
	}
	p := &g.Players[index]
	if p.Folded {
		return ErrInvalidAction
	}
	if p.Looked {
		return nil
	}
	p.Looked = true
	msg := fmt.Sprintf("%s 看牌", p.Name)
	g.appendEvent(events, GameEvent{Type: "look", PlayerIndex: index, PlayerName: p.Name, Message: msg})
	g.Message = msg
	return nil
}

func (g *Game) Fold(index int, events *[]GameEvent) error {
	if err := g.ensureTurn(index); err != nil {
		return err
	}
	p := &g.Players[index]
	if p.Folded {
		return ErrInvalidAction
	}
	p.Folded = true
	msg := fmt.Sprintf("%s 弃牌", p.Name)
	g.appendEvent(events, GameEvent{Type: "fold", PlayerIndex: index, PlayerName: p.Name, Message: msg})
	g.Message = msg
	if g.activeCount() <= 1 {
		g.finishLastStanding(events)
		return nil
	}
	g.advanceTurn(events)
	return nil
}

func (g *Game) Follow(index int, events *[]GameEvent) error {
	if err := g.ensureTurn(index); err != nil {
		return err
	}
	cost := g.callCost(index)
	p := &g.Players[index]
	if cost <= 0 {
		g.markActed(index)
		msg := fmt.Sprintf("%s 过牌", p.Name)
		g.appendEvent(events, GameEvent{Type: "check", PlayerIndex: index, PlayerName: p.Name, Message: msg})
		g.Message = msg
		g.trySettleRound(events)
		return nil
	}
	if err := g.charge(index, cost, events, "follow"); err != nil {
		return err
	}
	msg := fmt.Sprintf("%s 跟注 +%d", p.Name, cost)
	if ev := lastEvent(events); ev != nil {
		ev.Message = msg
	}
	g.Message = msg
	g.markActed(index)
	g.trySettleRound(events)
	return nil
}

func (g *Game) Raise(index, toAmount int, events *[]GameEvent) error {
	if err := g.ensureTurn(index); err != nil {
		return err
	}
	if toAmount < g.CurrentBet+g.MinRaise {
		return ErrInvalidAction
	}
	need := toAmount - g.Players[index].BetRound
	if need <= 0 {
		return ErrInvalidAction
	}
	if err := g.charge(index, need, events, "raise"); err != nil {
		return err
	}
	g.CurrentBet = toAmount
	g.lastAggressor = index
	g.resetActionRound(index)
	msg := fmt.Sprintf("%s 加注至 %d（+%d）", g.Players[index].Name, toAmount, need)
	if ev := lastEvent(events); ev != nil {
		ev.Message = msg
	}
	g.Message = msg
	g.advanceTurn(events)
	return nil
}

func (g *Game) Compare(index, target int, events *[]GameEvent) error {
	if err := g.ensureTurn(index); err != nil {
		return err
	}
	if target < 0 || target >= len(g.Players) || target == index {
		return ErrTargetInvalid
	}
	a, b := &g.Players[index], &g.Players[target]
	if a.Folded || b.Folded {
		return ErrTargetInvalid
	}
	if !a.Looked {
		return ErrCompareNeedLook
	}
	if !b.Looked {
		return ErrCompareTargetNeedLook
	}
	if err := g.charge(index, CompareCost, events, "compare"); err != nil {
		return err
	}
	pa, err := AnalyzeHand(a.Hand)
	if err != nil {
		return err
	}
	pb, err := AnalyzeHand(b.Hand)
	if err != nil {
		return err
	}
	cmp := CompareHands(pa, pb)
	loser := target
	if cmp > 0 {
		loser = target
	} else if cmp < 0 {
		loser = index
	} else {
		loser = index // 先比为负
	}
	g.Players[loser].Folded = true
	msg := fmt.Sprintf("%s 与 %s 比牌，%s 出局", a.Name, b.Name, g.Players[loser].Name)
	if ev := lastEvent(events); ev != nil {
		ev.TargetIndex = target
		ev.TargetName = b.Name
		ev.Message = msg
	}
	g.Message = msg
	if g.activeCount() <= 1 {
		g.finishLastStanding(events)
		return nil
	}
	g.markActed(index)
	g.advanceTurn(events)
	return nil
}

func (g *Game) charge(index, amount int, events *[]GameEvent, action string) error {
	p := &g.Players[index]
	if p.Chips < amount {
		return ErrInsufficientChips
	}
	p.Chips -= amount
	p.BetRound += amount
	p.TotalBet += amount
	g.Pot += amount
	g.appendEvent(events, GameEvent{Type: action, PlayerIndex: index, PlayerName: p.Name, Amount: amount})
	return nil
}

func (g *Game) callCost(index int) int {
	p := &g.Players[index]
	gap := g.CurrentBet - p.BetRound
	if gap <= 0 {
		return 0
	}
	if !p.Looked {
		gap = (gap + 1) / 2 // 闷牌跟注半价（向上取整）
	}
	if gap > p.Chips {
		return p.Chips
	}
	return gap
}

func (g *Game) markActed(index int) {
	g.hasActedThisRound[index] = true
}

func (g *Game) resetActionRound(exceptIndex int) {
	g.hasActedThisRound = make(map[int]bool)
	if exceptIndex >= 0 {
		g.hasActedThisRound[exceptIndex] = true
	}
}

func (g *Game) trySettleRound(events *[]GameEvent) {
	if g.allActed() {
		g.showdown(events)
		return
	}
	g.advanceTurn(events)
}

func (g *Game) allActed() bool {
	for i, p := range g.Players {
		if p.Folded {
			continue
		}
		if p.BetRound < g.CurrentBet {
			return false
		}
		if !g.hasActedThisRound[i] {
			return false
		}
	}
	return g.activeCount() > 0
}

func lastEvent(events *[]GameEvent) *GameEvent {
	if events == nil || len(*events) == 0 {
		return nil
	}
	return &(*events)[len(*events)-1]
}

func (g *Game) showdown(events *[]GameEvent) {
	active := g.activeIndices()
	if len(active) == 0 {
		return
	}
	if len(active) == 1 {
		g.finishLastStanding(events)
		return
	}
	bestIdx := active[0]
	var bestPattern *HandPattern
	for _, i := range active {
		pat, err := AnalyzeHand(g.Players[i].Hand)
		if err != nil {
			continue
		}
		g.applyPattern(i, pat)
		if bestPattern == nil || CompareHands(pat, bestPattern) > 0 {
			bestPattern = pat
			bestIdx = i
		}
	}
	g.finishWinner(bestIdx, bestPattern, events)
}

func (g *Game) finishLastStanding(events *[]GameEvent) {
	for i, p := range g.Players {
		if !p.Folded {
			pat, _ := AnalyzeHand(p.Hand)
			g.finishWinner(i, pat, events)
			return
		}
	}
}

func (g *Game) finishWinner(index int, pat *HandPattern, events *[]GameEvent) {
	g.Phase = PhaseFinished
	g.WinnerIndex = &index
	mul := 1
	label := "胜利"
	if pat != nil {
		g.applyPattern(index, pat)
		mul = pat.Multiplier
		label = pat.TypeLabel
	}
	g.WinMultiplier = mul
	g.WinHandLabel = label
	win := g.Pot
	g.Players[index].Chips += win
	g.Pot = 0
	g.Message = fmt.Sprintf("%s 以%s x%d 赢得 %d", g.Players[index].Name, label, mul, win)
	evt := GameEvent{
		Type: "game_over", PlayerIndex: index, PlayerName: g.Players[index].Name,
		HandLabel: label, Multiplier: mul, Amount: win,
		Message: g.Message,
	}
	if pat != nil {
		evt.HandType = string(pat.Type)
	}
	g.appendEvent(events, evt)
}

func (g *Game) applyPattern(index int, pat *HandPattern) {
	if pat == nil {
		return
	}
	g.Players[index].HandType = string(pat.Type)
	g.Players[index].HandLabel = pat.TypeLabel
	g.Players[index].Multiplier = pat.Multiplier
}

func (g *Game) advanceTurn(events *[]GameEvent) {
	if g.Phase != PhaseBetting {
		return
	}
	n := len(g.Players)
	for step := 1; step <= n; step++ {
		next := (g.CurrentTurn + step) % n
		if !g.Players[next].Folded {
			g.CurrentTurn = next
			g.resetTurnTimer()
			g.Message = fmt.Sprintf("轮到 %s", g.Players[next].Name)
			return
		}
	}
	g.showdown(events)
}

func (g *Game) ensureTurn(index int) error {
	if g.Phase != PhaseBetting {
		return ErrWrongPhase
	}
	if g.CurrentTurn != index {
		return ErrNotYourTurn
	}
	if g.Players[index].Folded {
		return ErrInvalidAction
	}
	return nil
}

func (g *Game) activeCount() int {
	n := 0
	for _, p := range g.Players {
		if !p.Folded {
			n++
		}
	}
	return n
}

func (g *Game) activeIndices() []int {
	var out []int
	for i, p := range g.Players {
		if !p.Folded {
			out = append(out, i)
		}
	}
	sort.Ints(out)
	return out
}

func (g *Game) appendEvent(events *[]GameEvent, e GameEvent) {
	if events == nil {
		return
	}
	*events = append(*events, e)
}

func (g *Game) IsFinished() bool {
	return g.Phase == PhaseFinished
}

func (g *Game) HasAI() bool {
	for _, p := range g.Players {
		if p.IsAI {
			return true
		}
	}
	return false
}

func (g *Game) PublicViewForSeat(seat int, events []GameEvent) PublicState {
	mul := make(map[string]int, len(HandMultipliers))
	for k, v := range HandMultipliers {
		mul[string(k)] = v
	}
	pub := PublicState{
		ID:               g.ID,
		Phase:            g.Phase,
		HumanPlayer:      seat,
		DealerIndex:      g.DealerIndex,
		CurrentTurn:      g.CurrentTurn,
		Pot:              g.Pot,
		CurrentBet:       g.CurrentBet,
		BaseAnte:         g.BaseAnte,
		MinRaise:         g.MinRaise,
		CompareCost:      CompareCost,
		WinnerIndex:      g.WinnerIndex,
		WinMultiplier:    g.WinMultiplier,
		WinHandLabel:     g.WinHandLabel,
		Message:          g.Message,
		HandMultipliers:  mul,
		TurnDeadlineUnix: g.TurnDeadlineUnix(),
		Events:           events,
	}
	for _, p := range g.Players {
		pp := PlayerPublic{
			Index: p.Index, Name: p.Name, IsAI: p.IsAI,
			Looked: p.Looked, Folded: p.Folded, Chips: p.Chips,
			BetRound: p.BetRound, TotalBet: p.TotalBet,
			HandType: p.HandType, HandLabel: p.HandLabel, Multiplier: p.Multiplier,
			CardCount: 3,
		}
		showHand := seat == p.Index && p.Looked
		if g.Phase == PhaseFinished {
			showHand = true
		}
		if showHand {
			pp.Hand = append([]card.Card(nil), p.Hand...)
			pp.CardCount = len(p.Hand)
		}
		pub.Players = append(pub.Players, pp)
	}
	if seat >= 0 && seat < len(g.Players) && g.Players[seat].Looked {
		pub.MyHand = append([]card.Card(nil), g.Players[seat].Hand...)
	}
	return pub
}

func (g *Game) ApplyHumanTimeout(events *[]GameEvent) error {
	if g.Phase != PhaseBetting || !g.IsHumanTurn() {
		return nil
	}
	return g.Fold(g.CurrentTurn, events)
}
