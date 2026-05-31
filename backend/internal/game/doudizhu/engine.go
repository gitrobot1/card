package doudizhu

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/time/card/backend/internal/game/card"
)

var (
	ErrNotYourTurn     = errors.New("not your turn")
	ErrInvalidCards    = errors.New("cards not in hand")
	ErrInvalidPattern  = errors.New("invalid pattern")
	ErrCannotBeat      = errors.New("cannot beat previous play")
	ErrMustPlay        = errors.New("must play when you are leader")
	ErrWrongPhase      = errors.New("wrong game phase")
	ErrAlreadyFinished = errors.New("game already finished")
)

type Phase string

const (
	PhaseCalling  Phase = "calling"
	PhasePlaying  Phase = "playing"
	PhaseFinished Phase = "finished"

	// debugDealOneCard 调试结算：每人 1 张，跳过叫地主，出完即胜。排查完改回 false。
	debugDealOneCard = false
)

type PlayerSeat struct {
	Index      int         `json:"index"`
	Name       string      `json:"name"`
	IsHuman    bool        `json:"is_human"`
	IsLandlord bool        `json:"is_landlord"`
	HandCount  int         `json:"hand_count"`
}

type PlayRecord struct {
	PlayerIndex int         `json:"player_index"`
	PlayerName  string      `json:"player_name"`
	Cards       []card.Card `json:"cards"`
	Pattern     PlayType    `json:"pattern"`
}

type RevealedHand struct {
	Index      int         `json:"index"`
	Name       string      `json:"name"`
	IsLandlord bool        `json:"is_landlord"`
	Cards      []card.Card `json:"cards"`
}

type Game struct {
	ID             string         `json:"id"`
	Phase          Phase          `json:"phase"`
	Players        [3]playerState `json:"-"`
	Seats          []PlayerSeat   `json:"players"`
	BottomCards    []card.Card    `json:"bottom_cards"`
	CurrentTurn    int            `json:"current_turn"`
	CallingIndex   int            `json:"calling_index"`
	LastCaller     *int           `json:"last_caller"`
	LastPlay       *PlayRecord    `json:"last_play"`
	LeaderIndex    int            `json:"leader_index"`
	PassCount      int            `json:"pass_count"`
	WinnerIndex    *int           `json:"winner_index"`
	WinnerRole     string         `json:"winner_role"`
	RevealedHands  []RevealedHand `json:"revealed_hands,omitempty"`
	Message        string         `json:"message"`
	HumanPlayer    int            `json:"human_player"`
	Online         bool           `json:"online"`
	TurnDeadline    time.Time      `json:"-"`
	TurnDeadlineAt  int64          `json:"turn_deadline_unix"`
	TurnSecondsLeft int            `json:"turn_seconds_left"`
}

type playerState struct {
	Name       string
	Human      bool
	IsLandlord bool
	Hand       []card.Card
}

type PublicState struct {
	Game
	MyHand []card.Card `json:"my_hand"`
	Events []GameEvent `json:"events"`
}

func NewGame(id, humanName string) *Game {
	g := &Game{
		ID:           id,
		Phase:        PhaseCalling,
		HumanPlayer:  0,
		CallingIndex: 0,
		Message:      "请选择是否抢地主",
		Online:       false,
	}
	g.Players[0] = playerState{Name: humanName, Human: true}
	g.Players[1] = playerState{Name: "电脑甲", Human: false}
	g.Players[2] = playerState{Name: "电脑乙", Human: false}
	g.deal()
	if debugDealOneCard {
		g.Phase = PhasePlaying
		g.CurrentTurn = g.HumanPlayer
		g.LeaderIndex = g.HumanPlayer
		g.Message = "调试模式：每人一张，出完即胜"
	}
	g.resetTurnTimer()
	g.syncSeats()
	return g
}

func NewOnlineGame(id string, names [3]string) *Game {
	g := &Game{
		ID:           id,
		Phase:        PhaseCalling,
		HumanPlayer:  0,
		CallingIndex: 0,
		Message:      "请选择是否抢地主",
		Online:       true,
	}
	for i := 0; i < 3; i++ {
		g.Players[i] = playerState{Name: names[i], Human: true}
	}
	g.deal()
	g.resetTurnTimer()
	g.syncSeats()
	return g
}

func (g *Game) deal() {
	deck := card.ShuffleRandom(card.NewDeck54())
	if debugDealOneCard {
		for i := 0; i < 3; i++ {
			g.Players[i].Hand = []card.Card{deck[i]}
		}
		g.BottomCards = []card.Card{}
		return
	}

	for i := 0; i < 3; i++ {
		g.Players[i].Hand = append([]card.Card(nil), deck[i*17:(i+1)*17]...)
		card.SortByRank(g.Players[i].Hand)
	}
	g.BottomCards = append([]card.Card(nil), deck[51:]...)
}

func (g *Game) PublicView(events []GameEvent) PublicState {
	return g.PublicViewForSeat(g.HumanPlayer, events)
}

func (g *Game) PublicViewForSeat(seatIndex int, events []GameEvent) PublicState {
	g.syncSeats()
	if events == nil {
		events = []GameEvent{}
	}
	if g.BottomCards == nil {
		g.BottomCards = []card.Card{}
	}
	for i := range g.RevealedHands {
		if g.RevealedHands[i].Cards == nil {
			g.RevealedHands[i].Cards = []card.Card{}
		}
	}
	view := PublicState{Game: *g, Events: events}
	if g.Phase == PhaseCalling {
		view.BottomCards = []card.Card{}
	}
	if seatIndex < 0 || seatIndex > 2 {
		seatIndex = 0
	}
	view.HumanPlayer = seatIndex
	view.MyHand = append([]card.Card(nil), g.Players[seatIndex].Hand...)
	view.TurnDeadlineAt = g.TurnDeadlineUnix()
	view.TurnSecondsLeft = g.secondsLeft()
	return view
}

func (g *Game) syncSeats() {
	g.Seats = make([]PlayerSeat, 3)
	for i := 0; i < 3; i++ {
		g.Seats[i] = PlayerSeat{
			Index:      i,
			Name:       g.Players[i].Name,
			IsHuman:    g.Players[i].Human,
			IsLandlord: g.Players[i].IsLandlord,
			HandCount:  len(g.Players[i].Hand),
		}
	}
}

func (g *Game) CallLandlord(playerIndex int, want bool) error {
	if g.Phase != PhaseCalling {
		return ErrWrongPhase
	}
	if playerIndex != g.CallingIndex {
		return ErrNotYourTurn
	}

	if want {
		caller := playerIndex
		g.LastCaller = &caller
	}

	g.CallingIndex = (g.CallingIndex + 1) % 3
	g.resetTurnTimer()
	if g.CallingIndex == 0 {
		g.finishCalling()
	}
	return nil
}

func (g *Game) finishCalling() {
	landlord := 0
	if g.LastCaller != nil {
		landlord = *g.LastCaller
	} else {
		landlord = rand.Intn(3)
	}

	g.Players[landlord].IsLandlord = true
	g.Players[landlord].Hand = append(g.Players[landlord].Hand, g.BottomCards...)
	card.SortByRank(g.Players[landlord].Hand)
	g.Phase = PhasePlaying
	g.CurrentTurn = landlord
	g.LeaderIndex = landlord
	g.Message = fmt.Sprintf("%s 成为地主，开始出牌", g.Players[landlord].Name)
	g.resetTurnTimer()
	g.syncSeats()
}

func (g *Game) Play(playerIndex int, cardIDs []string) (*PlayRecord, error) {
	if g.Phase != PhasePlaying {
		return nil, ErrWrongPhase
	}
	if g.WinnerIndex != nil {
		return nil, ErrAlreadyFinished
	}
	if playerIndex != g.CurrentTurn {
		return nil, ErrNotYourTurn
	}

	selected, err := card.FindByIDs(g.Players[playerIndex].Hand, cardIDs)
	if err != nil {
		return nil, ErrInvalidCards
	}

	pattern, err := AnalyzePattern(selected)
	if err != nil {
		return nil, ErrInvalidPattern
	}

	if g.LastPlay != nil && g.PassCount < 2 {
		prevPattern, err := AnalyzePattern(g.LastPlay.Cards)
		if err != nil {
			return nil, err
		}
		if !CanBeat(pattern, prevPattern) {
			return nil, ErrCannotBeat
		}
	}

	g.Players[playerIndex].Hand = card.RemoveCards(g.Players[playerIndex].Hand, selected)
	record := &PlayRecord{
		PlayerIndex: playerIndex,
		PlayerName:  g.Players[playerIndex].Name,
		Cards:       selected,
		Pattern:     pattern.Type,
	}
	g.LastPlay = record
	g.PassCount = 0
	g.LeaderIndex = playerIndex
	g.Message = fmt.Sprintf("%s 出牌", g.Players[playerIndex].Name)

	if len(g.Players[playerIndex].Hand) == 0 {
		g.finish(playerIndex)
		return record, nil
	}

	g.advanceTurn()
	return record, nil
}

func (g *Game) Pass(playerIndex int) error {
	if g.Phase != PhasePlaying {
		return ErrWrongPhase
	}
	if g.IsFinished() {
		return ErrAlreadyFinished
	}
	if playerIndex != g.CurrentTurn {
		return ErrNotYourTurn
	}
	if g.LastPlay == nil || g.LastPlay.PlayerIndex == playerIndex {
		return ErrMustPlay
	}

	g.PassCount++
	g.Message = fmt.Sprintf("%s 不出", g.Players[playerIndex].Name)
	if g.PassCount >= 2 {
		g.LastPlay = nil
		g.PassCount = 0
		g.CurrentTurn = g.LeaderIndex
		g.Message = fmt.Sprintf("新一轮，%s 先出", g.Players[g.LeaderIndex].Name)
		g.resetTurnTimer()
		g.syncSeats()
		return nil
	}

	g.advanceTurn()
	return nil
}

func (g *Game) advanceTurn() {
	g.CurrentTurn = (g.CurrentTurn + 1) % 3
	g.resetTurnTimer()
	g.syncSeats()
}

func (g *Game) finish(winner int) {
	g.Phase = PhaseFinished
	g.WinnerIndex = &winner
	g.TurnDeadline = time.Time{}
	if g.Players[winner].IsLandlord {
		g.WinnerRole = "landlord"
		g.Message = fmt.Sprintf("%s 作为地主获胜", g.Players[winner].Name)
	} else {
		g.WinnerRole = "farmer"
		g.Message = fmt.Sprintf("%s 作为农民获胜", g.Players[winner].Name)
	}
	g.syncSeats()
	g.revealAllHands()
}

func (g *Game) revealAllHands() {
	g.RevealedHands = make([]RevealedHand, 3)
	for i := 0; i < 3; i++ {
		g.RevealedHands[i] = RevealedHand{
			Index:      i,
			Name:       g.Players[i].Name,
			IsLandlord: g.Players[i].IsLandlord,
			Cards:      append([]card.Card(nil), g.Players[i].Hand...),
		}
	}
}

func (g *Game) IsHumanTurn() bool {
	if g.Phase == PhaseFinished {
		return false
	}
	idx := g.CallingIndex
	if g.Phase == PhasePlaying {
		idx = g.CurrentTurn
	}
	return g.Players[idx].Human
}

func (g *Game) HasAI() bool {
	for i := 0; i < 3; i++ {
		if !g.Players[i].Human {
			return true
		}
	}
	return false
}

func (g *Game) IsFinished() bool {
	return g.Phase == PhaseFinished
}
