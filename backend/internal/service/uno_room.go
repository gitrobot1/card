package service

import (
	"sync"

	"github.com/google/uuid"
)

const (
	MinUnoRoomPlayers = 2
	MaxUnoRoomPlayers = 8
)

type UnoRoom struct {
	ID         string       `json:"id"`
	Status     string       `json:"status"`
	GameID     string       `json:"game_id,omitempty"`
	HostUserID uint64       `json:"host_user_id"`
	Players    []RoomPlayer `json:"players"`
}

type UnoRoomService struct {
	mu    sync.RWMutex
	rooms map[string]*unoRoomState
}

type unoRoomState struct {
	room      UnoRoom
	userSeats map[uint64]int
}

func NewUnoRoomService() *UnoRoomService {
	return &UnoRoomService{rooms: make(map[string]*unoRoomState)}
}

func (s *UnoRoomService) Join(userID uint64, username string) (UnoRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.findByUserLocked(userID); existing != nil {
		return existing.room, nil
	}
	for _, st := range s.rooms {
		if st.room.Status == "waiting" && len(st.room.Players) < MaxUnoRoomPlayers {
			return s.addPlayerLocked(st, userID, username)
		}
	}
	id := uuid.NewString()
	st := &unoRoomState{
		room:      UnoRoom{ID: id, Status: "waiting"},
		userSeats: map[uint64]int{},
	}
	s.rooms[id] = st
	return s.addPlayerLocked(st, userID, username)
}

func (s *UnoRoomService) JoinRoom(roomID string, userID uint64, username string) (UnoRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.findByUserLocked(userID); existing != nil {
		if existing.room.ID == roomID {
			return existing.room, nil
		}
		return UnoRoom{}, ErrNotInRoom
	}
	st, ok := s.rooms[roomID]
	if !ok {
		return UnoRoom{}, ErrRoomNotFound
	}
	if st.room.Status != "waiting" {
		return UnoRoom{}, ErrRoomNotWaiting
	}
	if len(st.room.Players) >= MaxUnoRoomPlayers {
		return UnoRoom{}, ErrRoomFull
	}
	return s.addPlayerLocked(st, userID, username)
}

func (s *UnoRoomService) Get(roomID string, userID uint64) (UnoRoom, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return UnoRoom{}, ErrRoomNotFound
	}
	if _, ok := st.userSeats[userID]; !ok {
		return UnoRoom{}, ErrNotInRoom
	}
	return st.room, nil
}

func (s *UnoRoomService) Leave(roomID string, userID uint64) (UnoRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return UnoRoom{}, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return UnoRoom{}, ErrNotInRoom
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
		return UnoRoom{}, nil
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

func (s *UnoRoomService) SetReady(roomID string, userID uint64, ready bool) (UnoRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return UnoRoom{}, false, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return UnoRoom{}, false, ErrNotInRoom
	}
	if st.room.Status != "waiting" {
		return st.room, false, ErrRoomNotWaiting
	}

	st.room.Players[seat].Ready = ready
	allReady := len(st.room.Players) >= MinUnoRoomPlayers
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

func (s *UnoRoomService) Start(roomID string, userID uint64, games *UnoService) (UnoRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return UnoRoom{}, ErrRoomNotFound
	}
	if _, ok := st.userSeats[userID]; !ok {
		return UnoRoom{}, ErrNotInRoom
	}
	if st.room.HostUserID != userID {
		return UnoRoom{}, ErrNotRoomHost
	}
	if st.room.Status != "waiting" {
		return st.room, ErrRoomNotWaiting
	}
	if len(st.room.Players) < MinUnoRoomPlayers {
		return UnoRoom{}, ErrZhajinhuaNeedMorePlayers
	}
	for i := range st.room.Players {
		if st.room.Players[i].UserID == st.room.HostUserID {
			st.room.Players[i].Ready = true
		}
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
		return UnoRoom{}, err
	}
	st.room.Status = "playing"
	st.room.GameID = gameID
	return st.room, nil
}

func (s *UnoRoomService) SetReadyForNext(roomID string, userID uint64, ready bool) (UnoRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return UnoRoom{}, false, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return UnoRoom{}, false, ErrNotInRoom
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
	allReady := len(st.room.Players) >= MinUnoRoomPlayers
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

func (s *UnoRoomService) PlayersForGame(roomID string) ([]uint64, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return nil, nil, ErrRoomNotFound
	}
	if len(st.room.Players) < MinUnoRoomPlayers {
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

func (s *UnoRoomService) SetPlaying(roomID, gameID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.rooms[roomID]; ok {
		st.room.Status = "playing"
		st.room.GameID = gameID
	}
}

func (s *UnoRoomService) addPlayerLocked(st *unoRoomState, userID uint64, username string) (UnoRoom, error) {
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

func (s *UnoRoomService) findByUserLocked(userID uint64) *unoRoomState {
	for _, st := range s.rooms {
		if _, ok := st.userSeats[userID]; ok {
			return st
		}
	}
	return nil
}
