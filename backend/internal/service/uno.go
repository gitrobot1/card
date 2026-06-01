package service

import (
	"sync"

	"github.com/google/uuid"
	"github.com/time/card/backend/internal/game/uno"
)

type unoBinding struct {
	game  *uno.Game
	seats map[uint64]int
}

type UnoService struct {
	mu    sync.RWMutex
	games map[string]*unoBinding
}

func NewUnoService() *UnoService {
	return &UnoService{games: make(map[string]*unoBinding)}
}

func (s *UnoService) CreateGame(userID uint64, humanName string, botCount int) (uno.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, err := uno.NewSoloGame(uuid.NewString(), humanName, botCount)
	if err != nil {
		return uno.PublicState{}, err
	}
	s.games[g.ID] = &unoBinding{game: g, seats: map[uint64]int{userID: 0}}
	return g.PublicViewForSeat(0, nil), nil
}

func (s *UnoService) CreateOnlineGame(userIDs []uint64, names []string) (string, uno.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, err := uno.NewOnlineGame(uuid.NewString(), names)
	if err != nil {
		return "", uno.PublicState{}, err
	}
	seats := make(map[uint64]int, len(userIDs))
	for i, uid := range userIDs {
		seats[uid] = i
	}
	id := g.ID
	s.games[id] = &unoBinding{game: g, seats: seats}
	return id, g.PublicViewForSeat(0, nil), nil
}

func (s *UnoService) GetState(gameID string, userID uint64) (uno.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return uno.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return uno.PublicState{}, err
	}
	events := s.processTimeouts(b.game, nil)
	return s.finalize(b, seat, events), nil
}

func (s *UnoService) Play(gameID string, userID uint64, cardID string, color uno.Color) (uno.PublicState, error) {
	return s.act(gameID, userID, func(g *uno.Game, seat int, ev *[]uno.GameEvent) error {
		return g.Play(seat, cardID, color, ev)
	})
}

func (s *UnoService) Draw(gameID string, userID uint64) (uno.PublicState, error) {
	return s.act(gameID, userID, func(g *uno.Game, seat int, ev *[]uno.GameEvent) error {
		return g.Draw(seat, ev)
	})
}

func (s *UnoService) VoteEnd(gameID string, userID uint64) (uno.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return uno.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return uno.PublicState{}, err
	}
	var events []uno.GameEvent
	if err := b.game.VoteEnd(seat, &events); err != nil {
		return uno.PublicState{}, err
	}
	if b.game.HasAI() {
		uno.RunAIVoteEnd(b.game, &events)
	}
	return b.game.PublicViewForSeat(seat, events), nil
}

func (s *UnoService) RollFirst(gameID string, userID uint64) (uno.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return uno.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return uno.PublicState{}, err
	}
	g := b.game
	if g.Phase != uno.PhaseRollForFirst {
		return g.PublicViewForSeat(seat, nil), nil
	}
	var events []uno.GameEvent
	if err := g.RollRound(&events); err != nil {
		return uno.PublicState{}, err
	}
	return s.finalize(b, seat, events), nil
}

func (s *UnoService) Tick(gameID string, userID uint64) (uno.PublicState, error) {
	return s.GetState(gameID, userID)
}

type unoActFn func(g *uno.Game, seat int, ev *[]uno.GameEvent) error

func (s *UnoService) act(gameID string, userID uint64, fn unoActFn) (uno.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return uno.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return uno.PublicState{}, err
	}
	var events []uno.GameEvent
	if err := fn(b.game, seat, &events); err != nil {
		return uno.PublicState{}, err
	}
	return s.finalize(b, seat, events), nil
}

func (s *UnoService) seatLocked(b *unoBinding, userID uint64) (int, error) {
	seat, ok := b.seats[userID]
	if !ok {
		return 0, ErrNotInGame
	}
	return seat, nil
}

func (s *UnoService) processTimeouts(g *uno.Game, events []uno.GameEvent) []uno.GameEvent {
	if events == nil {
		events = []uno.GameEvent{}
	}
	for g.IsHumanTurn() && g.IsTurnExpired() && !g.IsFinished() {
		if err := g.ApplyHumanTimeout(&events); err != nil {
			break
		}
	}
	return events
}

func (s *UnoService) finalize(b *unoBinding, seat int, events []uno.GameEvent) uno.PublicState {
	if events == nil {
		events = []uno.GameEvent{}
	}
	g := b.game
	if g.HasAI() {
		uno.RunAITurns(g, &events)
	}
	return g.PublicViewForSeat(seat, events)
}
