package service

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

const (
	MinZhajinhuaRoomPlayers = 2
	MaxZhajinhuaRoomPlayers = 8
)

var (
	ErrZhajinhuaNeedMorePlayers = errors.New("need at least 2 players")
	ErrNotRoomHost              = errors.New("not room host")
	ErrNotAllReady              = errors.New("not all players ready")
)

type ZhajinhuaRoom struct {
	ID         string       `json:"id"`
	Status     string       `json:"status"`
	GameID     string       `json:"game_id,omitempty"`
	HostUserID uint64       `json:"host_user_id"`
	Players    []RoomPlayer `json:"players"`
}

type ZhajinhuaRoomService struct {
	mu    sync.RWMutex
	rooms map[string]*zhajinhuaRoomState
}

type zhajinhuaRoomState struct {
	room      ZhajinhuaRoom
	userSeats map[uint64]int
}

func NewZhajinhuaRoomService() *ZhajinhuaRoomService {
	return &ZhajinhuaRoomService{rooms: make(map[string]*zhajinhuaRoomState)}
}

func (s *ZhajinhuaRoomService) Join(userID uint64, username string) (ZhajinhuaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.findByUserLocked(userID); existing != nil {
		return existing.room, nil
	}
	for _, st := range s.rooms {
		if st.room.Status == "waiting" && len(st.room.Players) < MaxZhajinhuaRoomPlayers {
			return s.addPlayerLocked(st, userID, username)
		}
	}
	id := uuid.NewString()
	st := &zhajinhuaRoomState{
		room:      ZhajinhuaRoom{ID: id, Status: "waiting"},
		userSeats: map[uint64]int{},
	}
	s.rooms[id] = st
	return s.addPlayerLocked(st, userID, username)
}

func (s *ZhajinhuaRoomService) JoinRoom(roomID string, userID uint64, username string) (ZhajinhuaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.findByUserLocked(userID); existing != nil {
		if existing.room.ID == roomID {
			return existing.room, nil
		}
		return ZhajinhuaRoom{}, ErrNotInRoom
	}
	st, ok := s.rooms[roomID]
	if !ok {
		return ZhajinhuaRoom{}, ErrRoomNotFound
	}
	if st.room.Status != "waiting" {
		return ZhajinhuaRoom{}, ErrRoomNotWaiting
	}
	if len(st.room.Players) >= MaxZhajinhuaRoomPlayers {
		return ZhajinhuaRoom{}, ErrRoomFull
	}
	return s.addPlayerLocked(st, userID, username)
}

func (s *ZhajinhuaRoomService) Get(roomID string, userID uint64) (ZhajinhuaRoom, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return ZhajinhuaRoom{}, ErrRoomNotFound
	}
	if _, ok := st.userSeats[userID]; !ok {
		return ZhajinhuaRoom{}, ErrNotInRoom
	}
	return st.room, nil
}

func (s *ZhajinhuaRoomService) Leave(roomID string, userID uint64) (ZhajinhuaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return ZhajinhuaRoom{}, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return ZhajinhuaRoom{}, ErrNotInRoom
	}
	wasHost := st.room.HostUserID == userID
	players := st.room.Players[:seat]
	players = append(players, st.room.Players[seat+1:]...)
	st.room.Players = players
	delete(st.userSeats, userID)
	for uid, oldSeat := range st.userSeats {
		if oldSeat > seat {
			st.userSeats[uid] = oldSeat - 1
		}
	}
	for i := range st.room.Players {
		st.userSeats[st.room.Players[i].UserID] = i
	}
	if len(st.room.Players) == 0 {
		delete(s.rooms, roomID)
		return ZhajinhuaRoom{}, nil
	}
	if wasHost {
		st.room.HostUserID = st.room.Players[0].UserID
	}
	if st.room.Status == "waiting" {
		for i := range st.room.Players {
			st.room.Players[i].Ready = false
		}
	}
	return st.room, nil
}

func (s *ZhajinhuaRoomService) SetReady(roomID string, userID uint64, ready bool) (ZhajinhuaRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return ZhajinhuaRoom{}, false, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return ZhajinhuaRoom{}, false, ErrNotInRoom
	}
	if st.room.Status != "waiting" {
		return st.room, false, ErrRoomNotWaiting
	}

	st.room.Players[seat].Ready = ready
	allReady := len(st.room.Players) >= MinZhajinhuaRoomPlayers
	if allReady {
		for _, p := range st.room.Players {
			if !p.Ready {
				allReady = false
				break
			}
		}
	}
	return st.room, allReady, nil
}

func (s *ZhajinhuaRoomService) Start(roomID string, userID uint64, games *ZhajinhuaService) (ZhajinhuaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return ZhajinhuaRoom{}, ErrRoomNotFound
	}
	if _, ok := st.userSeats[userID]; !ok {
		return ZhajinhuaRoom{}, ErrNotInRoom
	}
	if st.room.HostUserID != userID {
		return ZhajinhuaRoom{}, ErrNotRoomHost
	}
	if st.room.Status != "waiting" {
		return st.room, ErrRoomNotWaiting
	}
	if len(st.room.Players) < MinZhajinhuaRoomPlayers {
		return ZhajinhuaRoom{}, ErrZhajinhuaNeedMorePlayers
	}
	for _, p := range st.room.Players {
		if !p.Ready {
			return st.room, ErrNotAllReady
		}
	}

	userIDs := make([]uint64, len(st.room.Players))
	names := make([]string, len(st.room.Players))
	for i, p := range st.room.Players {
		userIDs[i] = p.UserID
		names[i] = p.Username
	}
	gameID, _, err := games.CreateOnlineGame(userIDs, names)
	if err != nil {
		return ZhajinhuaRoom{}, err
	}
	st.room.Status = "playing"
	st.room.GameID = gameID
	return st.room, nil
}

func (s *ZhajinhuaRoomService) SetReadyForNext(roomID string, userID uint64, ready bool) (ZhajinhuaRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return ZhajinhuaRoom{}, false, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return ZhajinhuaRoom{}, false, ErrNotInRoom
	}

	if st.room.Status == "playing" {
		st.room.Status = "waiting"
		st.room.GameID = ""
		for i := range st.room.Players {
			st.room.Players[i].Ready = false
		}
	}
	if st.room.Status != "waiting" {
		return st.room, false, ErrRoomNotWaiting
	}

	st.room.Players[seat].Ready = ready
	allReady := len(st.room.Players) >= MinZhajinhuaRoomPlayers
	if allReady {
		for _, p := range st.room.Players {
			if !p.Ready {
				allReady = false
				break
			}
		}
	}
	return st.room, allReady, nil
}

func (s *ZhajinhuaRoomService) PlayersForGame(roomID string) ([]uint64, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return nil, nil, ErrRoomNotFound
	}
	if len(st.room.Players) < MinZhajinhuaRoomPlayers {
		return nil, nil, ErrZhajinhuaNeedMorePlayers
	}
	userIDs := make([]uint64, len(st.room.Players))
	names := make([]string, len(st.room.Players))
	for i, p := range st.room.Players {
		userIDs[i] = p.UserID
		names[i] = p.Username
	}
	return userIDs, names, nil
}

func (s *ZhajinhuaRoomService) SetPlaying(roomID, gameID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.rooms[roomID]; ok {
		st.room.Status = "playing"
		st.room.GameID = gameID
	}
}

func (s *ZhajinhuaRoomService) addPlayerLocked(st *zhajinhuaRoomState, userID uint64, username string) (ZhajinhuaRoom, error) {
	for _, p := range st.room.Players {
		if p.UserID == userID {
			return st.room, nil
		}
	}
	seat := len(st.room.Players)
	st.room.Players = append(st.room.Players, RoomPlayer{UserID: userID, Username: username})
	st.userSeats[userID] = seat
	if seat == 0 {
		st.room.HostUserID = userID
	}
	return st.room, nil
}

func (s *ZhajinhuaRoomService) findByUserLocked(userID uint64) *zhajinhuaRoomState {
	for _, st := range s.rooms {
		if _, ok := st.userSeats[userID]; ok {
			return st
		}
	}
	return nil
}
