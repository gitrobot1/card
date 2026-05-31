package service

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/time/card/backend/internal/game/zhajinhua"
)

type zhajinhuaBinding struct {
	game  *zhajinhua.Game
	seats map[uint64]int
}

type ZhajinhuaService struct {
	mu    sync.RWMutex
	games map[string]*zhajinhuaBinding
}

func NewZhajinhuaService() *ZhajinhuaService {
	return &ZhajinhuaService{games: make(map[string]*zhajinhuaBinding)}
}

func (s *ZhajinhuaService) CreateGame(userID uint64, humanName string, botCount int) (zhajinhua.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, err := zhajinhua.NewSoloGame(uuid.NewString(), humanName, botCount)
	if err != nil {
		return zhajinhua.PublicState{}, err
	}
	s.games[g.ID] = &zhajinhuaBinding{game: g, seats: map[uint64]int{userID: 0}}
	return g.PublicViewForSeat(0, nil), nil
}

func (s *ZhajinhuaService) CreateOnlineGame(userIDs []uint64, names []string) (string, zhajinhua.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, err := zhajinhua.NewOnlineGame(uuid.NewString(), names)
	if err != nil {
		return "", zhajinhua.PublicState{}, err
	}
	seats := make(map[uint64]int, len(userIDs))
	for i, uid := range userIDs {
		seats[uid] = i
	}
	id := g.ID
	s.games[id] = &zhajinhuaBinding{game: g, seats: seats}
	return id, g.PublicViewForSeat(0, nil), nil
}

func (s *ZhajinhuaService) GetState(gameID string, userID uint64) (zhajinhua.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return zhajinhua.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return zhajinhua.PublicState{}, err
	}
	events := s.processTimeouts(b.game, nil)
	return s.finalize(b, seat, events), nil
}

func (s *ZhajinhuaService) Look(gameID string, userID uint64) (zhajinhua.PublicState, error) {
	return s.act(gameID, userID, func(g *zhajinhua.Game, seat int, ev *[]zhajinhua.GameEvent) error {
		return g.Look(seat, ev)
	})
}

func (s *ZhajinhuaService) Fold(gameID string, userID uint64) (zhajinhua.PublicState, error) {
	return s.act(gameID, userID, func(g *zhajinhua.Game, seat int, ev *[]zhajinhua.GameEvent) error {
		return g.Fold(seat, ev)
	})
}

func (s *ZhajinhuaService) Follow(gameID string, userID uint64) (zhajinhua.PublicState, error) {
	return s.act(gameID, userID, func(g *zhajinhua.Game, seat int, ev *[]zhajinhua.GameEvent) error {
		return g.Follow(seat, ev)
	})
}

func (s *ZhajinhuaService) Raise(gameID string, userID uint64, amount int) (zhajinhua.PublicState, error) {
	return s.act(gameID, userID, func(g *zhajinhua.Game, seat int, ev *[]zhajinhua.GameEvent) error {
		return g.Raise(seat, amount, ev)
	})
}

func (s *ZhajinhuaService) Compare(gameID string, userID uint64, target int) (zhajinhua.PublicState, error) {
	return s.act(gameID, userID, func(g *zhajinhua.Game, seat int, ev *[]zhajinhua.GameEvent) error {
		return g.Compare(seat, target, ev)
	})
}

func (s *ZhajinhuaService) Tick(gameID string, userID uint64) (zhajinhua.PublicState, error) {
	return s.GetState(gameID, userID)
}

type zhActFn func(g *zhajinhua.Game, seat int, ev *[]zhajinhua.GameEvent) error

func (s *ZhajinhuaService) act(gameID string, userID uint64, fn zhActFn) (zhajinhua.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return zhajinhua.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return zhajinhua.PublicState{}, err
	}
	var events []zhajinhua.GameEvent
	if err := fn(b.game, seat, &events); err != nil {
		return zhajinhua.PublicState{}, err
	}
	return s.finalize(b, seat, events), nil
}

func (s *ZhajinhuaService) seatLocked(b *zhajinhuaBinding, userID uint64) (int, error) {
	seat, ok := b.seats[userID]
	if !ok {
		return 0, ErrNotInGame
	}
	return seat, nil
}

func (s *ZhajinhuaService) processTimeouts(g *zhajinhua.Game, events []zhajinhua.GameEvent) []zhajinhua.GameEvent {
	if events == nil {
		events = []zhajinhua.GameEvent{}
	}
	for g.IsHumanTurn() && g.IsTurnExpired() && !g.IsFinished() {
		if err := g.ApplyHumanTimeout(&events); err != nil {
			break
		}
	}
	return events
}

func (s *ZhajinhuaService) finalize(b *zhajinhuaBinding, seat int, events []zhajinhua.GameEvent) zhajinhua.PublicState {
	if events == nil {
		events = []zhajinhua.GameEvent{}
	}
	g := b.game
	if g.HasAI() {
		zhajinhua.RunAITurns(g, &events)
	}
	return g.PublicViewForSeat(seat, events)
}

func (s *ZhajinhuaService) SeatForUser(gameID string, userID uint64) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.games[gameID]
	if !ok {
		return 0, ErrGameNotFound
	}
	return s.seatLocked(b, userID)
}

var ErrInvalidBotCount = errors.New("invalid bot count")
