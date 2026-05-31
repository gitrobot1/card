package service

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/time/card/backend/internal/game"
	"github.com/time/card/backend/internal/game/doudizhu"
)

var ErrGameNotFound = errors.New("game not found")
var ErrNotInGame = errors.New("not in game")

type gameBinding struct {
	game  *doudizhu.Game
	seats map[uint64]int
}

type DouDizhuService struct {
	mu    sync.RWMutex
	games map[string]*gameBinding
}

func NewDouDizhuService() *DouDizhuService {
	return &DouDizhuService{
		games: make(map[string]*gameBinding),
	}
}

func (s *DouDizhuService) Catalog() []game.Meta {
	return game.Catalog()
}

func (s *DouDizhuService) CreateGame(userID uint64, humanName string) doudizhu.PublicState {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	g := doudizhu.NewGame(id, humanName)
	s.games[id] = &gameBinding{
		game:  g,
		seats: map[uint64]int{userID: 0},
	}
	return s.finalizeBinding(s.games[id], 0, nil)
}

func (s *DouDizhuService) CreateOnlineGame(userIDs [3]uint64, names [3]string) (string, doudizhu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	g := doudizhu.NewOnlineGame(id, names)
	seats := map[uint64]int{
		userIDs[0]: 0,
		userIDs[1]: 1,
		userIDs[2]: 2,
	}
	s.games[id] = &gameBinding{game: g, seats: seats}
	return id, s.finalizeBinding(s.games[id], 0, nil), nil
}

func (s *DouDizhuService) SeatForUser(gameID string, userID uint64) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	binding, ok := s.games[gameID]
	if !ok {
		return 0, ErrGameNotFound
	}
	seat, ok := binding.seats[userID]
	if !ok {
		return 0, ErrNotInGame
	}
	return seat, nil
}

func (s *DouDizhuService) GetState(gameID string, userID uint64) (doudizhu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	binding, ok := s.games[gameID]
	if !ok {
		return doudizhu.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatForUserLocked(binding, userID)
	if err != nil {
		return doudizhu.PublicState{}, err
	}
	events := s.processTimeouts(binding.game, nil)
	return s.finalizeBinding(binding, seat, events), nil
}

func (s *DouDizhuService) Tick(gameID string, userID uint64) (doudizhu.PublicState, error) {
	return s.GetState(gameID, userID)
}

func (s *DouDizhuService) Call(gameID string, userID uint64, want bool) (doudizhu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	binding, ok := s.games[gameID]
	if !ok {
		return doudizhu.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatForUserLocked(binding, userID)
	if err != nil {
		return doudizhu.PublicState{}, err
	}

	events := s.processTimeouts(binding.game, nil)
	if err := binding.game.CallLandlord(seat, want); err != nil {
		return doudizhu.PublicState{}, err
	}
	doudizhu.AppendCallEventPublic(&events, seat, binding.game.Players[seat].Name, want)
	return s.finalizeBinding(binding, seat, events), nil
}

func (s *DouDizhuService) Play(gameID string, userID uint64, cardIDs []string) (doudizhu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	binding, ok := s.games[gameID]
	if !ok {
		return doudizhu.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatForUserLocked(binding, userID)
	if err != nil {
		return doudizhu.PublicState{}, err
	}

	events := s.processTimeouts(binding.game, nil)
	record, err := binding.game.Play(seat, cardIDs)
	if err != nil {
		return doudizhu.PublicState{}, err
	}
	doudizhu.AppendPlayEventPublic(&events, record)
	return s.finalizeBinding(binding, seat, events), nil
}

func (s *DouDizhuService) Pass(gameID string, userID uint64) (doudizhu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	binding, ok := s.games[gameID]
	if !ok {
		return doudizhu.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatForUserLocked(binding, userID)
	if err != nil {
		return doudizhu.PublicState{}, err
	}

	events := s.processTimeouts(binding.game, nil)
	if err := binding.game.Pass(seat); err != nil {
		return doudizhu.PublicState{}, err
	}
	doudizhu.AppendPassEventPublic(&events, seat, binding.game.Players[seat].Name)
	return s.finalizeBinding(binding, seat, events), nil
}

func (s *DouDizhuService) Hint(gameID string, userID uint64) (*doudizhu.HintResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	binding, ok := s.games[gameID]
	if !ok {
		return nil, ErrGameNotFound
	}
	seat, err := s.seatForUserLocked(binding, userID)
	if err != nil {
		return nil, err
	}
	if binding.game.Online {
		return nil, doudizhu.ErrWrongPhase
	}
	return binding.game.Hint(seat)
}

func (s *DouDizhuService) IsGameFinished(gameID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	binding, ok := s.games[gameID]
	if !ok {
		return false
	}
	return binding.game.IsFinished()
}

func (s *DouDizhuService) seatForUserLocked(binding *gameBinding, userID uint64) (int, error) {
	seat, ok := binding.seats[userID]
	if !ok {
		return 0, ErrNotInGame
	}
	return seat, nil
}

func (s *DouDizhuService) processTimeouts(g *doudizhu.Game, events []doudizhu.GameEvent) []doudizhu.GameEvent {
	if events == nil {
		events = []doudizhu.GameEvent{}
	}
	for g.IsHumanTurn() && g.IsTurnExpired() && !g.IsFinished() {
		if err := g.ApplyHumanTimeout(&events); err != nil {
			break
		}
	}
	return events
}

func (s *DouDizhuService) finalizeBinding(binding *gameBinding, seatIndex int, events []doudizhu.GameEvent) doudizhu.PublicState {
	if events == nil {
		events = []doudizhu.GameEvent{}
	}
	g := binding.game
	if g.HasAI() {
		doudizhu.RunAITurns(g, &events)
	}
	if g.IsFinished() && g.WinnerIndex != nil && !hasGameOverEvent(events) {
		winner := *g.WinnerIndex
		doudizhu.AppendGameOverEventPublic(&events, winner, g.Players[winner].Name)
	}
	return g.PublicViewForSeat(seatIndex, events)
}

func hasGameOverEvent(events []doudizhu.GameEvent) bool {
	for _, event := range events {
		if event.Type == doudizhu.EventGameOver {
			return true
		}
	}
	return false
}
