package douniu

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/time/card/backend/internal/game/card"
)

var (
	ErrNotInGame      = errors.New("not in game")
	ErrWrongPhase     = errors.New("wrong phase")
	ErrInvalidGrab    = errors.New("invalid grab multiplier")
	ErrInvalidBet     = errors.New("invalid bet multiplier")
	ErrAlreadyActed   = errors.New("already acted this phase")
	ErrGameOver       = errors.New("game over")
)

type Player struct {
	Index          int         `json:"index"`
	Name           string      `json:"name"`
	IsAI           bool        `json:"is_ai"`
	Hand           []card.Card `json:"-"`
	Chips          int         `json:"chips"`
	GrabMult       int         `json:"grab_mult"`
	BetMult        int         `json:"bet_mult"`
	HandType       string      `json:"hand_type,omitempty"`
	HandLabel      string      `json:"hand_label,omitempty"`
	HandMultiplier int         `json:"hand_multiplier,omitempty"`
	RoundDelta     int         `json:"round_delta,omitempty"`
}

type PlayerPublic struct {
	Index          int         `json:"index"`
	Name           string      `json:"name"`
	IsAI           bool        `json:"is_ai"`
	Chips          int         `json:"chips"`
	GrabMult       int         `json:"grab_mult"`
	BetMult        int         `json:"bet_mult"`
	HandType       string      `json:"hand_type,omitempty"`
	HandLabel      string      `json:"hand_label,omitempty"`
	HandMultiplier int         `json:"hand_multiplier,omitempty"`
	HandLayout     *HandLayout `json:"hand_layout,omitempty"`
	RoundDelta     int         `json:"round_delta,omitempty"`
	Hand           []card.Card `json:"hand,omitempty"`
	CardCount      int         `json:"card_count"`
	GrabDone       bool        `json:"grab_done"`
	BetDone        bool        `json:"bet_done"`
}

type Game struct {
	ID               string
	Phase            string
	Players          []Player
	HumanPlayer      int
	BankerIndex      int
	BaseAnte         int
	Message          string
	TurnDeadline     time.Time
	TurnDeadlineUnix int64
}

type PublicState struct {
	ID               string         `json:"id"`
	Phase            string         `json:"phase"`
	Players          []PlayerPublic `json:"players"`
	HumanPlayer      int            `json:"human_player"`
	BankerIndex      int            `json:"banker_index"`
	BaseAnte         int            `json:"base_ante"`
	Message          string         `json:"message"`
	MyHand           []card.Card    `json:"my_hand,omitempty"`
	MyHandLabel      string         `json:"my_hand_label,omitempty"`
	MyHandType       string         `json:"my_hand_type,omitempty"`
	MyHandMultiplier int            `json:"my_hand_multiplier,omitempty"`
	MyHandLayout     *HandLayout    `json:"my_hand_layout,omitempty"`
	HandMultipliers  map[string]int `json:"hand_multipliers"`
	GrabOptions      []int          `json:"grab_options"`
	BetOptions       []int          `json:"bet_options"`
	TurnDeadlineUnix int64          `json:"turn_deadline_unix"`
	Events           []GameEvent    `json:"events"`
}

func NewSoloGame(id, humanName string, botCount int, chips []int) (*Game, error) {
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
	return newGame(id, names, isAI, 0, chips)
}

func NewOnlineGame(id string, names []string, chips []int) (*Game, error) {
	if len(names) < MinPlayers || len(names) > MaxPlayers {
		return nil, fmt.Errorf("player count must be %d-%d", MinPlayers, MaxPlayers)
	}
	isAI := make([]bool, len(names))
	return newGame(id, names, isAI, 0, chips)
}

func CarryChipsFromGame(prev *Game, names []string) []int {
	if prev == nil || prev.Phase != PhaseFinished {
		return nil
	}
	out := make([]int, len(names))
	for i, name := range names {
		out[i] = DefaultChips
		for _, p := range prev.Players {
			if p.Name == name {
				out[i] = p.Chips
				break
			}
		}
	}
	return out
}

func newGame(id string, names []string, isAI []bool, humanSeat int, chips []int) (*Game, error) {
	g := &Game{
		ID:           id,
		Phase:        PhaseGrabBanker,
		HumanPlayer:  humanSeat,
		BankerIndex:  -1,
		BaseAnte:     DefaultAnte,
	}
	for i, name := range names {
		chip := DefaultChips
		if chips != nil && i < len(chips) {
			chip = chips[i]
		}
		g.Players = append(g.Players, Player{
			Index:    i,
			Name:     name,
			IsAI:     isAI[i],
			Chips:    chip,
			GrabMult: GrabUnset,
			BetMult:  BetUnset,
		})
	}
	g.deal()
	g.resetPhaseTimer()
	g.Message = "看牌后选择是否抢庄"
	return g, nil
}

func (g *Game) deal() {
	deck := card.ShuffleRandom(card.NewDeck52())
	need := len(g.Players) * 5
	if need > len(deck) {
		return
	}
	for i := range g.Players {
		g.Players[i].Hand = append([]card.Card(nil), deck[i*5:(i+1)*5]...)
		g.Players[i].GrabMult = GrabUnset
		g.Players[i].BetMult = BetUnset
		g.Players[i].HandType = ""
		g.Players[i].HandLabel = ""
		g.Players[i].HandMultiplier = 0
		g.Players[i].RoundDelta = 0
	}
}

func (g *Game) HasAI() bool {
	for _, p := range g.Players {
		if p.IsAI {
			return true
		}
	}
	return false
}

func (g *Game) IsFinished() bool {
	return g.Phase == PhaseFinished
}

func (g *Game) IsHumanPending() bool {
	if g.IsFinished() {
		return false
	}
	h := &g.Players[g.HumanPlayer]
	switch g.Phase {
	case PhaseGrabBanker:
		return h.GrabMult == GrabUnset
	case PhaseBetting:
		return g.HumanPlayer != g.BankerIndex && h.BetMult == BetUnset
	default:
		return false
	}
}

func (g *Game) GrabBanker(seat, mult int, events *[]GameEvent) error {
	if g.Phase != PhaseGrabBanker {
		return ErrWrongPhase
	}
	if seat < 0 || seat >= len(g.Players) {
		return ErrNotInGame
	}
	if !validGrab(mult) {
		return ErrInvalidGrab
	}
	p := &g.Players[seat]
	if p.GrabMult != GrabUnset {
		return ErrAlreadyActed
	}
	p.GrabMult = mult
	*events = append(*events, GameEvent{
		Type:        "grab_banker",
		PlayerIndex: seat,
		PlayerName:  p.Name,
		GrabMult:    mult,
		Message:     formatGrabMessage(p.Name, mult),
	})
	if g.allGrabDone() {
		g.resolveBanker(events)
	}
	return nil
}

func (g *Game) PlaceBet(seat, mult int, events *[]GameEvent) error {
	if g.Phase != PhaseBetting {
		return ErrWrongPhase
	}
	if seat < 0 || seat >= len(g.Players) {
		return ErrNotInGame
	}
	if seat == g.BankerIndex {
		return ErrInvalidBet
	}
	if !validBet(mult) {
		return ErrInvalidBet
	}
	p := &g.Players[seat]
	if p.BetMult != BetUnset {
		return ErrAlreadyActed
	}
	p.BetMult = mult
	*events = append(*events, GameEvent{
		Type:        "place_bet",
		PlayerIndex: seat,
		PlayerName:  p.Name,
		BetMult:     mult,
		Message:     fmt.Sprintf("%s 下注 ×%d", p.Name, mult),
	})
	if g.allBetDone() {
		g.settle(events)
	}
	return nil
}

func (g *Game) allGrabDone() bool {
	for _, p := range g.Players {
		if p.GrabMult == GrabUnset {
			return false
		}
	}
	return true
}

func (g *Game) allBetDone() bool {
	for i, p := range g.Players {
		if i == g.BankerIndex {
			continue
		}
		if p.BetMult == BetUnset {
			return false
		}
	}
	return true
}

func (g *Game) resolveBanker(events *[]GameEvent) {
	maxGrab := -1
	candidates := []int{}
	for i, p := range g.Players {
		if p.GrabMult > maxGrab {
			maxGrab = p.GrabMult
			candidates = []int{i}
		} else if p.GrabMult == maxGrab {
			candidates = append(candidates, i)
		}
	}
	if maxGrab <= 0 {
		g.BankerIndex = candidates[rand.Intn(len(candidates))]
	} else {
		g.BankerIndex = candidates[rand.Intn(len(candidates))]
	}
	banker := &g.Players[g.BankerIndex]
	*events = append(*events, GameEvent{
		Type:        "banker_set",
		PlayerIndex: g.BankerIndex,
		PlayerName:  banker.Name,
		GrabMult:    banker.GrabMult,
		Message:     fmt.Sprintf("%s 成为庄家", banker.Name),
	})
	g.Phase = PhaseBetting
	g.resetPhaseTimer()
	g.Message = "闲家选择下注倍数"
}

func (g *Game) settle(events *[]GameEvent) {
	for i := range g.Players {
		res := AnalyzeHand(g.Players[i].Hand)
		g.Players[i].HandType = res.Type
		g.Players[i].HandLabel = res.Label
		g.Players[i].HandMultiplier = res.Multiplier
	}

	banker := &g.Players[g.BankerIndex]
	bankerRes := AnalyzeHand(banker.Hand)
	bankerDelta := 0

	for i := range g.Players {
		if i == g.BankerIndex {
			continue
		}
		p := &g.Players[i]
		playerRes := AnalyzeHand(p.Hand)
		cmp := CompareHands(playerRes, bankerRes)
		winMult := playerRes.Multiplier
		if cmp < 0 {
			winMult = bankerRes.Multiplier
		}
		amount := g.BaseAnte * p.BetMult * winMult
		if cmp > 0 {
			p.Chips += amount
			p.RoundDelta = amount
			bankerDelta -= amount
			*events = append(*events, GameEvent{
				Type:        "settle",
				PlayerIndex: i,
				PlayerName:  p.Name,
				TargetIndex: g.BankerIndex,
				TargetName:  banker.Name,
				Amount:      amount,
				HandType:    playerRes.Type,
				HandLabel:   playerRes.Label,
				Multiplier:  winMult,
				Message:     formatSettle(p.Name, amount, true),
			})
		} else if cmp < 0 {
			p.Chips -= amount
			p.RoundDelta = -amount
			bankerDelta += amount
			*events = append(*events, GameEvent{
				Type:        "settle",
				PlayerIndex: i,
				PlayerName:  p.Name,
				TargetIndex: g.BankerIndex,
				TargetName:  banker.Name,
				Amount:      amount,
				HandType:    playerRes.Type,
				HandLabel:   playerRes.Label,
				Multiplier:  winMult,
				Message:     formatSettle(p.Name, amount, false),
			})
		} else {
			p.RoundDelta = 0
			*events = append(*events, GameEvent{
				Type:        "settle",
				PlayerIndex: i,
				PlayerName:  p.Name,
				TargetIndex: g.BankerIndex,
				TargetName:  banker.Name,
				HandType:    playerRes.Type,
				HandLabel:   playerRes.Label,
				Message:     fmt.Sprintf("%s 与庄家平局", p.Name),
			})
		}
	}
	banker.RoundDelta = bankerDelta
	banker.Chips += bankerDelta

	*events = append(*events, GameEvent{
		Type:        "game_over",
		PlayerIndex: g.BankerIndex,
		PlayerName:  banker.Name,
		Amount:      bankerDelta,
		Message:     "本局结算完成",
	})
	g.Phase = PhaseFinished
	g.TurnDeadline = time.Time{}
	g.TurnDeadlineUnix = 0
	g.Message = "本局结束"
}

func (g *Game) PublicViewForSeat(seat int, events []GameEvent) PublicState {
	players := make([]PlayerPublic, len(g.Players))
	reveal := g.Phase == PhaseFinished
	for i, p := range g.Players {
		pub := PlayerPublic{
			Index:          p.Index,
			Name:           p.Name,
			IsAI:           p.IsAI,
			Chips:          p.Chips,
			GrabMult:       p.GrabMult,
			BetMult:        p.BetMult,
			HandType:       p.HandType,
			HandLabel:      p.HandLabel,
			HandMultiplier: p.HandMultiplier,
			RoundDelta:     p.RoundDelta,
			CardCount:      len(p.Hand),
			GrabDone:       p.GrabMult != GrabUnset,
			BetDone:        i == g.BankerIndex || p.BetMult != BetUnset,
		}
		if reveal {
			pub.Hand = append([]card.Card(nil), p.Hand...)
			layout := LayoutForHand(p.Hand)
			pub.HandLayout = &layout
		}
		players[i] = pub
	}
	var myHand []card.Card
	var myHandLabel string
	var myHandType string
	var myHandMultiplier int
	var myHandLayout *HandLayout
	if seat >= 0 && seat < len(g.Players) && len(g.Players[seat].Hand) > 0 {
		myHand = append([]card.Card(nil), g.Players[seat].Hand...)
		res := AnalyzeHand(myHand)
		myHandLabel = res.Label
		myHandType = res.Type
		myHandMultiplier = res.Multiplier
		layout := res.BuildLayout(myHand)
		myHandLayout = &layout
	}
	if events == nil {
		events = []GameEvent{}
	}
	grabOpts := make([]int, MaxGrabMult+1)
	for i := range grabOpts {
		grabOpts[i] = i
	}
	return PublicState{
		ID:               g.ID,
		Phase:            g.Phase,
		Players:          players,
		HumanPlayer:      seat,
		BankerIndex:      g.BankerIndex,
		BaseAnte:         g.BaseAnte,
		Message:          g.Message,
		MyHand:           myHand,
		MyHandLabel:      myHandLabel,
		MyHandType:       myHandType,
		MyHandMultiplier: myHandMultiplier,
		MyHandLayout:     myHandLayout,
		HandMultipliers:  HandMultipliersTable(),
		GrabOptions:      grabOpts,
		BetOptions:       append([]int(nil), BetOptions...),
		TurnDeadlineUnix: g.TurnDeadlineUnix,
		Events:           events,
	}
}
