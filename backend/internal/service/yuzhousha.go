package service

import (
	"sync"

	"github.com/google/uuid"
	"github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

type yuzhoushaBinding struct {
	game  *engine.Game
	seats map[uint64]int
}

type YuzhoushaService struct {
	mu    sync.RWMutex
	games map[string]*yuzhoushaBinding
}

func NewYuzhoushaService() *YuzhoushaService {
	return &YuzhoushaService{games: make(map[string]*yuzhoushaBinding)}
}

func (s *YuzhoushaService) ModesCatalog() []engine.ModeMeta {
	return engine.ModesCatalog()
}

func (s *YuzhoushaService) CreateSolo(userID uint64, humanName, characterID, mode string) (engine.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, err := engine.NewSolo(uuid.NewString(), humanName, characterID, mode)
	if err != nil {
		return engine.PublicState{}, err
	}
	s.games[g.ID] = &yuzhoushaBinding{game: g, seats: map[uint64]int{userID: 0}}
	return s.finalize(g, 0, nil), nil
}

func (s *YuzhoushaService) ListHeroes(q engine.HeroesQuery) engine.HeroesPage {
	return engine.ListHeroes(q)
}

func (s *YuzhoushaService) PacksCatalog() []skill.PackManifest {
	return engine.PacksCatalog()
}

func (s *YuzhoushaService) HeroesCatalog() []engine.Character {
	return engine.HeroesCatalog()
}

func (s *YuzhoushaService) UseSkill(gameID string, userID uint64, req engine.UseSkillRequest) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.UseSkill(seat, req, ev)
	})
}

func (s *YuzhoushaService) GetState(gameID string, userID uint64) (engine.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return engine.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return engine.PublicState{}, err
	}
	events := s.processTimeouts(b.game, nil)
	return s.finalize(b.game, seat, events), nil
}

func (s *YuzhoushaService) PlayCard(gameID string, userID uint64, cardID string, target engine.PlayTarget) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.PlayCardWithTarget(seat, cardID, target, ev)
	})
}

func (s *YuzhoushaService) RespondShan(gameID string, userID uint64, cardID string) (engine.PublicState, error) {
	return s.RespondCard(gameID, userID, cardID)
}

func (s *YuzhoushaService) RespondCard(gameID string, userID uint64, cardID string) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.RespondCard(seat, cardID, ev)
	})
}

func (s *YuzhoushaService) PassResponse(gameID string, userID uint64) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.PassResponse(seat, ev)
	})
}

func (s *YuzhoushaService) TryBaguaJudge(gameID string, userID uint64) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.TryBaguaJudge(seat, ev)
	})
}

func (s *YuzhoushaService) EndPlay(gameID string, userID uint64) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.EndPlay(seat, ev)
	})
}

func (s *YuzhoushaService) DiscardCards(gameID string, userID uint64, cardIDs []string) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.DiscardCards(seat, cardIDs, ev)
	})
}

func (s *YuzhoushaService) PassPrepare(gameID string, userID uint64) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.PassPrepare(seat, ev)
	})
}

func (s *YuzhoushaService) PassDraw(gameID string, userID uint64) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.PassDrawPhase(seat, ev)
	})
}

func (s *YuzhoushaService) FinishPeekDeck(gameID string, userID uint64, req engine.PeekDeckRequest) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.FinishPeekDeck(seat, req, ev)
	})
}

func (s *YuzhoushaService) FinishGuanxing(gameID string, userID uint64, req engine.GuanxingRequest) (engine.PublicState, error) {
	return s.FinishPeekDeck(gameID, userID, req)
}

func (s *YuzhoushaService) Tick(gameID string, userID uint64) (engine.PublicState, error) {
	return s.GetState(gameID, userID)
}

type yzsActFn func(g *engine.Game, seat int, ev *[]engine.GameEvent) error

func (s *YuzhoushaService) act(gameID string, userID uint64, fn yzsActFn) (engine.PublicState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.games[gameID]
	if !ok {
		return engine.PublicState{}, ErrGameNotFound
	}
	seat, err := s.seatLocked(b, userID)
	if err != nil {
		return engine.PublicState{}, err
	}
	var events []engine.GameEvent
	if err := fn(b.game, seat, &events); err != nil {
		return engine.PublicState{}, err
	}
	return s.finalize(b.game, seat, events), nil
}

func (s *YuzhoushaService) seatLocked(b *yuzhoushaBinding, userID uint64) (int, error) {
	seat, ok := b.seats[userID]
	if !ok {
		return 0, ErrNotInGame
	}
	return seat, nil
}

func (s *YuzhoushaService) processTimeouts(g *engine.Game, events []engine.GameEvent) []engine.GameEvent {
	if events == nil {
		events = []engine.GameEvent{}
	}
	for g.IsHumanPending() && g.IsPhaseExpired() && !g.IsFinished() {
		if err := g.ApplyHumanTimeout(&events); err != nil {
			break
		}
	}
	return events
}

func (s *YuzhoushaService) finalize(g *engine.Game, seat int, events []engine.GameEvent) engine.PublicState {
	if events == nil {
		events = []engine.GameEvent{}
	}
	if g.HasAI() {
		engine.RunAIActions(g, &events)
	}
	return g.PublicViewForSeat(seat, events)
}
