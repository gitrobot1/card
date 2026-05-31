package uno

import (
	"errors"
	"fmt"
	"time"
)

const (
	MinPlayers     = 2
	MaxPlayers     = 8
	InitialHand    = 5
	TurnTimeoutSec = 20
)

var (
	ErrNotYourTurn      = errors.New("not your turn")
	ErrWrongPhase       = errors.New("wrong phase")
	ErrInvalidCard      = errors.New("invalid card")
	ErrCannotPlay       = errors.New("cannot play this card")
	ErrLastCardMustBeBasic = errors.New("last card must be a basic card")
	ErrGameOver         = errors.New("game over")
	ErrInvalidColor     = errors.New("invalid color choice")
)

type Phase string

const (
	PhasePlaying  Phase = "playing"
	PhaseFinished Phase = "finished"
)

type Player struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	IsAI      bool   `json:"is_ai"`
	Hand      []Card `json:"hand,omitempty"`
	HandCount int    `json:"hand_count"`
}

type Game struct {
	ID            string    `json:"id"`
	Phase         Phase     `json:"phase"`
	Players       []Player  `json:"players"`
	HumanPlayer   int       `json:"human_player"`
	CurrentTurn   int       `json:"current_turn"`
	Direction     int       `json:"direction"`
	CurrentColor  Color     `json:"current_color"`
	TopCard       Card      `json:"top_card"`
	DrawCount     int       `json:"draw_count"`
	DiscardCount  int       `json:"discard_count"`
	WinnerIndex   *int      `json:"winner_index,omitempty"`
	Message       string    `json:"message"`
	PendingDrawPenalty int  `json:"pending_draw_penalty"`
	DrawStackWild4Only bool `json:"draw_stack_wild4_only"`
	MustPlayAfterStack bool `json:"must_play_after_stack"`
	TurnDeadlineUnix int64  `json:"turn_deadline_unix"`
	TurnDeadline  time.Time `json:"-"`
	drawPile    []Card
	discardPile []Card
}

type PublicState struct {
	Game
	MyHand []Card     `json:"my_hand"`
	Events []GameEvent `json:"events"`
}

func NewSoloGame(id, humanName string, botCount int) (*Game, error) {
	if botCount < 1 || botCount > MaxPlayers-1 {
		return nil, fmt.Errorf("bot count must be 1-%d", MaxPlayers-1)
	}
	names := make([]string, botCount+1)
	names[0] = humanName
	for i := 1; i <= botCount; i++ {
		names[i] = fmt.Sprintf("电脑%d", i)
	}
	isAI := make([]bool, len(names))
	for i := 1; i < len(names); i++ {
		isAI[i] = true
	}
	return newGame(id, names, isAI, 0)
}

func newGame(id string, names []string, isAI []bool, humanSeat int) (*Game, error) {
	if len(names) < MinPlayers || len(names) > MaxPlayers {
		return nil, fmt.Errorf("player count must be %d-%d", MinPlayers, MaxPlayers)
	}
	g := &Game{
		ID:          id,
		Phase:       PhasePlaying,
		HumanPlayer: humanSeat,
		Direction:   1,
	}
	for i, name := range names {
		g.Players = append(g.Players, Player{Index: i, Name: name, IsAI: isAI[i]})
	}
	g.deal()
	g.resetTurnTimer()
	g.Message = fmt.Sprintf("%s 出牌", g.Players[g.CurrentTurn].Name)
	return g, nil
}

func (g *Game) deal() {
	deck := ShuffleDeck(NewDeck108())
	for i := range g.Players {
		g.Players[i].Hand = append([]Card(nil), deck[:InitialHand]...)
		deck = deck[InitialHand:]
	}
	g.drawPile = deck
	for {
		if len(g.drawPile) == 0 {
			g.refillDrawPile()
		}
		top := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		if IsActionValue(top.Value) && !IsWildCard(top) {
			g.discardPile = append(g.discardPile, top)
			continue
		}
		g.discardPile = append(g.discardPile, top)
		g.TopCard = top
		if IsWildCard(top) {
			g.CurrentColor = ColorRed
		} else {
			g.CurrentColor = top.Color
		}
		break
	}
	g.syncCounts()
	g.CurrentTurn = 0
}

func (g *Game) syncCounts() {
	for i := range g.Players {
		g.Players[i].HandCount = len(g.Players[i].Hand)
	}
	g.DrawCount = len(g.drawPile)
	g.DiscardCount = len(g.discardPile)
}

func (g *Game) refillDrawPile() {
	if len(g.discardPile) <= 1 {
		return
	}
	top := g.discardPile[len(g.discardPile)-1]
	rest := append([]Card(nil), g.discardPile[:len(g.discardPile)-1]...)
	g.drawPile = ShuffleDeck(rest)
	g.discardPile = []Card{top}
}

func (g *Game) drawCards(seat int, n int) {
	for i := 0; i < n; i++ {
		if len(g.drawPile) == 0 {
			g.refillDrawPile()
		}
		if len(g.drawPile) == 0 {
			break
		}
		c := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.Players[seat].Hand = append(g.Players[seat].Hand, c)
	}
	g.syncCounts()
}

func (g *Game) ensureTurn(seat int) error {
	if g.Phase != PhasePlaying {
		return ErrGameOver
	}
	if g.CurrentTurn != seat {
		return ErrNotYourTurn
	}
	return nil
}

func (g *Game) CanPlayCard(seat int, card Card) bool {
	return g.canPlayCard(seat, card)
}

func (g *Game) hasDrawStack() bool {
	return g.PendingDrawPenalty > 0
}

func (g *Game) canPlayCard(seat int, card Card) bool {
	if len(g.Players[seat].Hand) == 1 && IsActionValue(card.Value) {
		return false
	}
	if g.hasDrawStack() {
		v := Value(card.Value)
		if v == ValueWild4 {
			return true
		}
		if v == ValueDraw2 && !g.DrawStackWild4Only {
			return true
		}
		return false
	}
	if IsWildCard(card) {
		return true
	}
	if card.Color == g.CurrentColor {
		return true
	}
	return card.Value == g.TopCard.Value
}

func (g *Game) Play(seat int, cardID string, chosen Color, events *[]GameEvent) error {
	if err := g.ensureTurn(seat); err != nil {
		return err
	}
	idx, card := g.findInHand(seat, cardID)
	if idx < 0 {
		return ErrInvalidCard
	}
	if len(g.Players[seat].Hand) == 1 && IsActionValue(card.Value) {
		return ErrLastCardMustBeBasic
	}
	if !g.canPlayCard(seat, card) {
		return ErrCannotPlay
	}
	pickColor := chosen
	if IsWildCard(card) {
		if pickColor == "" || pickColor == ColorWild {
			return ErrInvalidColor
		}
		valid := false
		for _, c := range PlayColors {
			if c == pickColor {
				valid = true
				break
			}
		}
		if !valid {
			return ErrInvalidColor
		}
	}
	g.removeFromHand(seat, idx)
	g.discardPile = append(g.discardPile, card)
	g.TopCard = card
	if IsWildCard(card) {
		g.CurrentColor = pickColor
	} else {
		g.CurrentColor = card.Color
	}
	p := &g.Players[seat]
	msg := fmt.Sprintf("%s 打出 %s", p.Name, card.Label)
	appendEvent(events, GameEvent{
		Type: EventPlay, PlayerIndex: seat, PlayerName: p.Name, Card: &card, Color: g.CurrentColor, Message: msg,
	})
	g.Message = msg
	g.MustPlayAfterStack = false

	if len(p.Hand) == 0 {
		g.finishWinner(seat, events)
		return nil
	}

	g.resolveTurnAfterPlay(card, events)
	return nil
}

func (g *Game) Draw(seat int, events *[]GameEvent) error {
	if err := g.ensureTurn(seat); err != nil {
		return err
	}
	if g.hasDrawStack() {
		return g.acceptDrawStack(seat, events)
	}
	if g.MustPlayAfterStack && len(g.PlayableCards(seat)) > 0 {
		return ErrCannotPlay
	}
	g.drawCards(seat, 1)
	p := &g.Players[seat]
	drawn := p.Hand[len(p.Hand)-1]
	msg := fmt.Sprintf("%s 摸牌", p.Name)
	appendEvent(events, GameEvent{Type: EventDraw, PlayerIndex: seat, PlayerName: p.Name, Card: &drawn, Amount: 1, Message: msg})
	g.MustPlayAfterStack = false
	g.advanceTurn(events)
	return nil
}

func (g *Game) acceptDrawStack(seat int, events *[]GameEvent) error {
	n := g.PendingDrawPenalty
	before := len(g.Players[seat].Hand)
	g.drawCards(seat, n)
	p := &g.Players[seat]
	msg := fmt.Sprintf("%s 摸 %d 张", p.Name, n)
	ev := GameEvent{
		Type: EventDraw, PlayerIndex: seat, PlayerName: p.Name, Amount: n, Message: msg,
	}
	if len(p.Hand) > before {
		last := p.Hand[len(p.Hand)-1]
		ev.Card = &last
	}
	appendEvent(events, ev)
	g.PendingDrawPenalty = 0
	g.DrawStackWild4Only = false
	g.MustPlayAfterStack = true
	g.resetTurnTimer()
	g.Message = fmt.Sprintf("%s 摸完罚牌，请出牌；无牌可出则摸牌", p.Name)
	return nil
}

func (g *Game) setStackTurnMessage() {
	if !g.hasDrawStack() {
		return
	}
	p := g.Players[g.CurrentTurn]
	if g.DrawStackWild4Only {
		g.Message = fmt.Sprintf("%s 需摸 %d 张或出 +4", p.Name, g.PendingDrawPenalty)
	} else {
		g.Message = fmt.Sprintf("%s 需摸 %d 张或叠 +2/+4", p.Name, g.PendingDrawPenalty)
	}
}

func (g *Game) resolveTurnAfterPlay(card Card, events *[]GameEvent) {
	n := len(g.Players)
	switch Value(card.Value) {
	case ValueSkip:
		victim := g.nextSeat(g.CurrentTurn)
		g.drawCards(victim, 1)
		p := g.Players[victim]
		drawn := p.Hand[len(p.Hand)-1]
		appendEvent(events, GameEvent{
			Type: EventDraw, PlayerIndex: victim, PlayerName: p.Name, Card: &drawn, Amount: 1,
			Message: fmt.Sprintf("%s 被跳过，摸 1 张", p.Name),
		})
		g.CurrentTurn = g.nextSeat(victim)
	case ValueReverse:
		if n == 2 {
			g.CurrentTurn = g.nextSeat(g.CurrentTurn)
		} else {
			g.Direction *= -1
			g.CurrentTurn = g.nextSeat(g.CurrentTurn)
		}
	case ValueDraw2:
		g.PendingDrawPenalty += 2
		g.CurrentTurn = g.nextSeat(g.CurrentTurn)
	case ValueWild4:
		g.PendingDrawPenalty += 4
		g.DrawStackWild4Only = true
		g.CurrentTurn = g.nextSeat(g.CurrentTurn)
	default:
		g.CurrentTurn = g.nextSeat(g.CurrentTurn)
	}
	g.resetTurnTimer()
	if g.hasDrawStack() {
		g.setStackTurnMessage()
	} else {
		g.Message = fmt.Sprintf("%s 出牌", g.Players[g.CurrentTurn].Name)
	}
}

func (g *Game) nextSeat(seat int) int {
	n := len(g.Players)
	return (seat + g.Direction + n) % n
}

func (g *Game) advanceTurn(events *[]GameEvent) {
	g.CurrentTurn = g.nextSeat(g.CurrentTurn)
	g.resetTurnTimer()
	g.Message = fmt.Sprintf("%s 出牌", g.Players[g.CurrentTurn].Name)
}

func (g *Game) finishWinner(seat int, events *[]GameEvent) {
	g.Phase = PhaseFinished
	g.WinnerIndex = &seat
	p := g.Players[seat]
	msg := fmt.Sprintf("%s 获胜！", p.Name)
	appendEvent(events, GameEvent{Type: EventGameOver, PlayerIndex: seat, PlayerName: p.Name, Message: msg})
	g.Message = msg
}

func (g *Game) findInHand(seat int, cardID string) (int, Card) {
	for i, c := range g.Players[seat].Hand {
		if c.ID == cardID {
			return i, c
		}
	}
	return -1, Card{}
}

func (g *Game) removeFromHand(seat, idx int) {
	h := g.Players[seat].Hand
	g.Players[seat].Hand = append(h[:idx], h[idx+1:]...)
	g.syncCounts()
}

func (g *Game) resetTurnTimer() {
	g.TurnDeadline = time.Now().Add(TurnTimeoutSec * time.Second)
	g.TurnDeadlineUnix = g.TurnDeadline.Unix()
}

func (g *Game) IsTurnExpired() bool {
	if g.TurnDeadline.IsZero() || g.Phase != PhasePlaying {
		return false
	}
	return time.Now().After(g.TurnDeadline)
}

func (g *Game) IsHumanTurn() bool {
	if g.Phase != PhasePlaying {
		return false
	}
	if g.CurrentTurn < 0 || g.CurrentTurn >= len(g.Players) {
		return false
	}
	return !g.Players[g.CurrentTurn].IsAI
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

func (g *Game) ApplyHumanTimeout(events *[]GameEvent) error {
	if !g.IsHumanTurn() || !g.IsTurnExpired() {
		return nil
	}
	if g.hasDrawStack() {
		return g.acceptDrawStack(g.CurrentTurn, events)
	}
	playable := g.PlayableCards(g.CurrentTurn)
	if len(playable) > 0 {
		card := pickAICard(g, g.CurrentTurn, playable)
		color := g.CurrentColor
		if IsWildCard(card) {
			color = pickAIColor(g, g.CurrentTurn)
		}
		return g.Play(g.CurrentTurn, card.ID, color, events)
	}
	return g.Draw(g.CurrentTurn, events)
}

func (g *Game) PlayableCards(seat int) []Card {
	var out []Card
	for _, c := range g.Players[seat].Hand {
		if g.canPlayCard(seat, c) {
			out = append(out, c)
		}
	}
	return out
}

func (g *Game) PublicViewForSeat(seat int, events []GameEvent) PublicState {
	players := make([]Player, len(g.Players))
	for i, p := range g.Players {
		players[i] = Player{
			Index: p.Index, Name: p.Name, IsAI: p.IsAI, HandCount: len(p.Hand),
		}
	}
	var myHand []Card
	if seat >= 0 && seat < len(g.Players) {
		myHand = append([]Card(nil), g.Players[seat].Hand...)
	}
	return PublicState{
		Game: Game{
			ID: g.ID, Phase: g.Phase, Players: players, HumanPlayer: seat,
			CurrentTurn: g.CurrentTurn, Direction: g.Direction, CurrentColor: g.CurrentColor,
			TopCard: g.TopCard, DrawCount: g.DrawCount, DiscardCount: g.DiscardCount,
			WinnerIndex: g.WinnerIndex, Message: g.Message,
			PendingDrawPenalty: g.PendingDrawPenalty,
			DrawStackWild4Only: g.DrawStackWild4Only,
			MustPlayAfterStack: g.MustPlayAfterStack,
			TurnDeadlineUnix: g.TurnDeadlineUnix,
		},
		MyHand: myHand,
		Events: filterEventsForSeat(events, seat),
	}
}
