package service

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

const MaxDouDizhuRoomPlayers = 3

var (
	ErrRoomNotFound   = errors.New("room not found")
	ErrRoomFull       = errors.New("room is full")
	ErrNotInRoom      = errors.New("not in room")
	ErrRoomNotWaiting = errors.New("room is not accepting players")
)

type RoomPlayer struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Ready    bool   `json:"ready"`
}

type DouDizhuRoom struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	GameID string       `json:"game_id,omitempty"`
	Players []RoomPlayer `json:"players"`
}

type DouDizhuRoomService struct {
	mu    sync.RWMutex
	rooms map[string]*douDizhuRoomState
}

type douDizhuRoomState struct {
	room      DouDizhuRoom
	userSeats map[uint64]int
}

func NewDouDizhuRoomService() *DouDizhuRoomService {
	return &DouDizhuRoomService{
		rooms: make(map[string]*douDizhuRoomState),
	}
}

func (s *DouDizhuRoomService) Join(userID uint64, username string) (DouDizhuRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing := s.findRoomByUserLocked(userID); existing != nil {
		return existing.room, nil
	}

	for _, state := range s.rooms {
		if state.room.Status != "waiting" {
			continue
		}
		if len(state.room.Players) >= MaxDouDizhuRoomPlayers {
			continue
		}
		return s.addPlayerLocked(state, userID, username)
	}

	id := uuid.NewString()
	state := &douDizhuRoomState{
		room: DouDizhuRoom{
			ID:      id,
			Status:  "waiting",
			Players: []RoomPlayer{},
		},
		userSeats: make(map[uint64]int),
	}
	s.rooms[id] = state
	return s.addPlayerLocked(state, userID, username)
}

func (s *DouDizhuRoomService) JoinRoom(roomID string, userID uint64, username string) (DouDizhuRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing := s.findRoomByUserLocked(userID); existing != nil {
		if existing.room.ID == roomID {
			return existing.room, nil
		}
		return DouDizhuRoom{}, ErrNotInRoom
	}

	state, ok := s.rooms[roomID]
	if !ok {
		return DouDizhuRoom{}, ErrRoomNotFound
	}
	if state.room.Status != "waiting" {
		return DouDizhuRoom{}, ErrRoomNotWaiting
	}
	if len(state.room.Players) >= MaxDouDizhuRoomPlayers {
		return DouDizhuRoom{}, ErrRoomFull
	}
	return s.addPlayerLocked(state, userID, username)
}

func (s *DouDizhuRoomService) Get(roomID string, userID uint64) (DouDizhuRoom, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.rooms[roomID]
	if !ok {
		return DouDizhuRoom{}, ErrRoomNotFound
	}
	if _, ok := state.userSeats[userID]; !ok {
		return DouDizhuRoom{}, ErrNotInRoom
	}
	return state.room, nil
}

func (s *DouDizhuRoomService) Leave(roomID string, userID uint64) (DouDizhuRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.rooms[roomID]
	if !ok {
		return DouDizhuRoom{}, ErrRoomNotFound
	}
	seat, ok := state.userSeats[userID]
	if !ok {
		return DouDizhuRoom{}, ErrNotInRoom
	}

	players := make([]RoomPlayer, 0, len(state.room.Players)-1)
	for i, player := range state.room.Players {
		if i == seat {
			continue
		}
		players = append(players, player)
	}
	state.room.Players = players
	delete(state.userSeats, userID)
	s.reindexSeatsLocked(state)

	if len(state.room.Players) == 0 {
		delete(s.rooms, roomID)
		return DouDizhuRoom{}, nil
	}

	if state.room.Status == "waiting" {
		for i := range state.room.Players {
			state.room.Players[i].Ready = false
		}
	}

	return state.room, nil
}

func (s *DouDizhuRoomService) SetReady(roomID string, userID uint64, ready bool) (DouDizhuRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.rooms[roomID]
	if !ok {
		return DouDizhuRoom{}, false, ErrRoomNotFound
	}
	seat, ok := state.userSeats[userID]
	if !ok {
		return DouDizhuRoom{}, false, ErrNotInRoom
	}
	if state.room.Status != "waiting" {
		return state.room, false, ErrRoomNotWaiting
	}

	state.room.Players[seat].Ready = ready
	allReady := len(state.room.Players) == MaxDouDizhuRoomPlayers
	if allReady {
		for _, player := range state.room.Players {
			if !player.Ready {
				allReady = false
				break
			}
		}
	}
	return state.room, allReady, nil
}

// SetReadyForNext 局结束后准备下一局：首次请求时 playing→waiting，后续只累计准备状态。
func (s *DouDizhuRoomService) SetReadyForNext(roomID string, userID uint64, ready bool) (DouDizhuRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.rooms[roomID]
	if !ok {
		return DouDizhuRoom{}, false, ErrRoomNotFound
	}
	seat, ok := state.userSeats[userID]
	if !ok {
		return DouDizhuRoom{}, false, ErrNotInRoom
	}

	if state.room.Status == "playing" {
		state.room.Status = "waiting"
		state.room.GameID = ""
		for i := range state.room.Players {
			state.room.Players[i].Ready = false
		}
	}
	if state.room.Status != "waiting" {
		return state.room, false, ErrRoomNotWaiting
	}

	state.room.Players[seat].Ready = ready
	allReady := len(state.room.Players) == MaxDouDizhuRoomPlayers
	if allReady {
		for _, player := range state.room.Players {
			if !player.Ready {
				allReady = false
				break
			}
		}
	}
	return state.room, allReady, nil
}

func (s *DouDizhuRoomService) MarkPlaying(roomID, gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}
	state.room.Status = "playing"
	state.room.GameID = gameID
	return nil
}

func (s *DouDizhuRoomService) ResetAfterGame(roomID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}
	state.room.Status = "waiting"
	state.room.GameID = ""
	for i := range state.room.Players {
		state.room.Players[i].Ready = false
	}
	return nil
}

func (s *DouDizhuRoomService) PlayersForGame(roomID string) ([3]string, [3]uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.rooms[roomID]
	if !ok {
		return [3]string{}, [3]uint64{}, ErrRoomNotFound
	}
	if len(state.room.Players) != MaxDouDizhuRoomPlayers {
		return [3]string{}, [3]uint64{}, errors.New("room players incomplete")
	}

	names := [3]string{}
	ids := [3]uint64{}
	for seat, player := range state.room.Players {
		names[seat] = player.Username
		ids[seat] = player.UserID
	}
	return names, ids, nil
}

func (s *DouDizhuRoomService) addPlayerLocked(state *douDizhuRoomState, userID uint64, username string) (DouDizhuRoom, error) {
	if len(state.room.Players) >= MaxDouDizhuRoomPlayers {
		return DouDizhuRoom{}, ErrRoomFull
	}
	seat := len(state.room.Players)
	state.room.Players = append(state.room.Players, RoomPlayer{
		UserID:   userID,
		Username: username,
		Ready:    false,
	})
	state.userSeats[userID] = seat
	return state.room, nil
}

func (s *DouDizhuRoomService) reindexSeatsLocked(state *douDizhuRoomState) {
	state.userSeats = make(map[uint64]int, len(state.room.Players))
	for i, player := range state.room.Players {
		state.userSeats[player.UserID] = i
	}
}

func (s *DouDizhuRoomService) findRoomByUserLocked(userID uint64) *douDizhuRoomState {
	for _, state := range s.rooms {
		if _, ok := state.userSeats[userID]; ok {
			return state
		}
	}
	return nil
}
