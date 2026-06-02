package service

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/time/card/backend/internal/game/douniu"
)

type douniuBinding struct {
	game  *douniu.Game
	seats map[uint64]int
}

type DouNiuService struct {
	mu    sync.RWMutex
	games map[string]*douniuBinding
}

func NewDouNiuService() *DouNiuService {
	return &DouNiuService{games: make(map[string]*douniuBinding)}
}

func (s *DouNiuService) CreateGame(userID uint64, humanName string, botCount int, prevGameID string) (douniu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chips := s.chipsForSoloRematchLocked(prevGameID, userID, humanName, botCount)
	g, err := douniu.NewSoloGame(uuid.NewString(), humanName, botCount, chips)
	if err != nil {
		return douniu.PublicState{}, err
	}
	s.games[g.ID] = &douniuBinding{game: g, seats: map[uint64]int{userID: 0}}
	return g.PublicViewForSeat(0, nil), nil
}

func (s *DouNiuService) CreateOnlineGame(userIDs []uint64, names []string, prevGameID string) (string, douniu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chips := s.chipsForOnlineRematchLocked(prevGameID, names)
	g, err := douniu.NewOnlineGame(uuid.NewString(), names, chips)
	if err != nil {
		return "", douniu.PublicState{}, err
	}
	seats := make(map[uint64]int, len(userIDs))
	for i, uid := range userIDs {
		seats[uid] = i
	}
	id := g.ID
	s.games[id] = &douniuBinding{game: g, seats: seats}
	return id, g.PublicViewForSeat(0, nil), nil
}

func (s *DouNiuService) chipsForSoloRematchLocked(prevGameID string, userID uint64, humanName string, botCount int) []int {
	if prevGameID == "" {
		return nil
	}
	b, ok := s.games[prevGameID]
	if !ok {
		return nil
	}
	if _, ok := b.seats[userID]; !ok || !b.game.IsFinished() {
		return nil
	}
	names := make([]string, botCount+1)
	names[0] = humanName
	for i := 1; i <= botCount; i++ {
		names[i] = fmt.Sprintf("电脑%d", i)
	}
	return douniu.CarryChipsFromGame(b.game, names)
}

func (s *DouNiuService) chipsForOnlineRematchLocked(prevGameID string, names []string) []int {
	if prevGameID == "" {
		return nil
	}
	b, ok := s.games[prevGameID]
	if !ok || !b.game.IsFinished() {
		return nil
	}
	return douniu.CarryChipsFromGame(b.game, names)
}

func (s *DouNiuService) GetState(gameID string, userID uint64) (douniu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return douniu.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return douniu.PublicState{}, err
	}
	events := s.processTimeouts(b.game, nil)
	return s.finalize(b.game, seat, events), nil
}

func (s *DouNiuService) GrabBanker(gameID string, userID uint64, mult int) (douniu.PublicState, error) {
	return s.act(gameID, userID, func(g *douniu.Game, seat int, ev *[]douniu.GameEvent) error {
		return g.GrabBanker(seat, mult, ev)
	})
}

func (s *DouNiuService) PlaceBet(gameID string, userID uint64, mult int) (douniu.PublicState, error) {
	return s.act(gameID, userID, func(g *douniu.Game, seat int, ev *[]douniu.GameEvent) error {
		return g.PlaceBet(seat, mult, ev)
	})
}

func (s *DouNiuService) Tick(gameID string, userID uint64) (douniu.PublicState, error) {
	return s.GetState(gameID, userID)
}

func (s *DouNiuService) MemberUserIDs(gameID string) ([]uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.games[gameID]
	if !ok {
		return nil, ErrGameNotFound
	}
	out := make([]uint64, 0, len(b.seats))
	for uid := range b.seats {
		out = append(out, uid)
	}
	return out, nil
}

func (s *DouNiuService) SnapshotForUser(gameID string, userID uint64, events []douniu.GameEvent) (douniu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return douniu.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return douniu.PublicState{}, err
	}
	if events == nil {
		events = []douniu.GameEvent{}
	}
	events = s.processTimeouts(b.game, events)
	return s.finalize(b.game, seat, events), nil
}

type douniuActFn func(g *douniu.Game, seat int, ev *[]douniu.GameEvent) error

func (s *DouNiuService) act(gameID string, userID uint64, fn douniuActFn) (douniu.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return douniu.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return douniu.PublicState{}, err
	}
	var events []douniu.GameEvent
	if err := fn(b.game, seat, &events); err != nil {
		return douniu.PublicState{}, err
	}
	return s.finalize(b.game, seat, events), nil
}

func (s *DouNiuService) seatLocked(b *douniuBinding, userID uint64) (int, error) {
	seat, ok := b.seats[userID]
	if !ok {
		return 0, ErrNotInGame
	}
	return seat, nil
}

func (s *DouNiuService) processTimeouts(g *douniu.Game, events []douniu.GameEvent) []douniu.GameEvent {
	if events == nil {
		events = []douniu.GameEvent{}
	}
	for g.IsHumanPending() && g.IsPhaseExpired() && !g.IsFinished() {
		if err := g.ApplyHumanTimeout(&events); err != nil {
			break
		}
	}
	return events
}

func (s *DouNiuService) finalize(g *douniu.Game, seat int, events []douniu.GameEvent) douniu.PublicState {
	if events == nil {
		events = []douniu.GameEvent{}
	}
	if g.HasAI() {
		douniu.RunAIActions(g, &events)
	}
	return g.PublicViewForSeat(seat, events)
}
