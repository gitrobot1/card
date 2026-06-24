package service

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
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

func (s *YuzhoushaService) CreateOnlineGame(roomMode string, userIDs []uint64, names []string, charIDs []string) (string, error) {
	n := len(userIDs)
	if n != len(names) || n != len(charIDs) || n < 2 {
		return "", fmt.Errorf("online game requires matching player lists")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	var (
		g   *engine.Game
		err error
	)
	switch mode.NormalizeID(roomMode) {
	case engine.Mode1v1:
		if n != 2 {
			return "", fmt.Errorf("online 1v1 requires exactly 2 players")
		}
		var nameArr [2]string
		var charArr [2]string
		copy(nameArr[:], names)
		copy(charArr[:], charIDs)
		g, err = engine.NewOnline1v1(id, nameArr, charArr)
	case engine.Mode2v2:
		if n != 4 {
			return "", fmt.Errorf("online 2v2 requires exactly 4 players")
		}
		var nameArr [4]string
		var charArr [4]string
		copy(nameArr[:], names)
		copy(charArr[:], charIDs)
		g, err = engine.NewOnline2v2(id, nameArr, charArr)
	case engine.Mode3pChain:
		if n != 3 {
			return "", fmt.Errorf("online 3p chain requires exactly 3 players")
		}
		var nameArr [3]string
		var charArr [3]string
		copy(nameArr[:], names)
		copy(charArr[:], charIDs)
		g, err = engine.NewOnline3pChain(id, nameArr, charArr)
	case engine.Mode3v3:
		if n != 6 {
			return "", fmt.Errorf("online 3v3 requires exactly 6 players")
		}
		var nameArr [6]string
		var charArr [6]string
		copy(nameArr[:], names)
		copy(charArr[:], charIDs)
		g, err = engine.NewOnline3v3(id, nameArr, charArr)
	case engine.ModeIdentity5:
		if n != 5 {
			return "", fmt.Errorf("online identity_5 requires exactly 5 players")
		}
		var nameArr [5]string
		var charArr [5]string
		copy(nameArr[:], names)
		copy(charArr[:], charIDs)
		g, err = engine.NewOnlineIdentity5(id, nameArr, charArr)
	case engine.ModeIdentity8:
		if n != 8 {
			return "", fmt.Errorf("online identity_8 requires exactly 8 players")
		}
		var nameArr [8]string
		var charArr [8]string
		copy(nameArr[:], names)
		copy(charArr[:], charIDs)
		g, err = engine.NewOnlineIdentity8(id, nameArr, charArr)
	default:
		return "", ErrOnlineModeUnknown
	}
	if err != nil {
		return "", err
	}
	seats := make(map[uint64]int, n)
	for i, uid := range userIDs {
		seats[uid] = i
	}
	s.games[g.ID] = &yuzhoushaBinding{game: g, seats: seats}
	return g.ID, nil
}

func (s *YuzhoushaService) MemberUserIDs(gameID string) ([]uint64, error) {
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

func (s *YuzhoushaService) SnapshotForUser(gameID string, userID uint64, events []engine.GameEvent) (engine.PublicState, error) {
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
	if events == nil {
		events = []engine.GameEvent{}
	}
	events = s.processTimeouts(b.game, events)
	return s.finalize(b.game, seat, events), nil
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

// PlayZhangbaSha 丈八蛇矛：两张手牌当杀使用。
func (s *YuzhoushaService) PlayZhangbaSha(gameID string, userID uint64, card1ID, card2ID string, targetIndex int) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.TryZhangbaSha(seat, targetIndex, []string{card1ID, card2ID}, ev)
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

func (s *YuzhoushaService) PassAllWuxiek(gameID string, userID uint64) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.PassAllWuxiek(seat, ev)
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

func (s *YuzhoushaService) RespondDiscardCards(gameID string, userID uint64, cardIDs []string) (engine.PublicState, error) {
	return s.act(gameID, userID, func(g *engine.Game, seat int, ev *[]engine.GameEvent) error {
		return g.RespondDiscardCards(seat, cardIDs, ev)
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
	engine.LogGameState(g, "finalize BEFORE AI")
	if g.HasAI() {
		engine.RunAIActions(g, &events)
	}
	engine.LogGameState(g, "finalize AFTER AI")
	return g.PublicViewForSeat(seat, events)
}
