package service

import (
	"sync"

	"github.com/google/uuid"
)

const (
	MinDouNiuRoomPlayers = 2
	MaxDouNiuRoomPlayers = 8
)

type DouNiuRoom struct {
	ID         string       `json:"id"`
	Status     string       `json:"status"`
	GameID     string       `json:"game_id,omitempty"`
	HostUserID uint64       `json:"host_user_id"`
	Players    []RoomPlayer `json:"players"`
}

type DouNiuRoomService struct {
	mu    sync.RWMutex
	rooms map[string]*douniuRoomState
}

type douniuRoomState struct {
	room      DouNiuRoom
	userSeats map[uint64]int
}

func NewDouNiuRoomService() *DouNiuRoomService {
	return &DouNiuRoomService{rooms: make(map[string]*douniuRoomState)}
}

func (s *DouNiuRoomService) Join(userID uint64, username string) (DouNiuRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.findByUserLocked(userID); existing != nil {
		return existing.room, nil
	}
	for _, st := range s.rooms {
		if st.room.Status == "waiting" && len(st.room.Players) < MaxDouNiuRoomPlayers {
			return s.addPlayerLocked(st, userID, username)
		}
	}
	id := uuid.NewString()
	st := &douniuRoomState{
		room:      DouNiuRoom{ID: id, Status: "waiting"},
		userSeats: map[uint64]int{},
	}
	s.rooms[id] = st
	return s.addPlayerLocked(st, userID, username)
}

func (s *DouNiuRoomService) JoinRoom(roomID string, userID uint64, username string) (DouNiuRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.findByUserLocked(userID); existing != nil {
		if existing.room.ID == roomID {
			return existing.room, nil
		}
		return DouNiuRoom{}, ErrNotInRoom
	}
	st, ok := s.rooms[roomID]
	if !ok {
		return DouNiuRoom{}, ErrRoomNotFound
	}
	if st.room.Status != "waiting" {
		return DouNiuRoom{}, ErrRoomNotWaiting
	}
	if len(st.room.Players) >= MaxDouNiuRoomPlayers {
		return DouNiuRoom{}, ErrRoomFull
	}
	return s.addPlayerLocked(st, userID, username)
}

func (s *DouNiuRoomService) Get(roomID string, userID uint64) (DouNiuRoom, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return DouNiuRoom{}, ErrRoomNotFound
	}
	if _, ok := st.userSeats[userID]; !ok {
		return DouNiuRoom{}, ErrNotInRoom
	}
	return st.room, nil
}

func (s *DouNiuRoomService) Leave(roomID string, userID uint64) (DouNiuRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return DouNiuRoom{}, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return DouNiuRoom{}, ErrNotInRoom
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
		return DouNiuRoom{}, nil
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

func (s *DouNiuRoomService) SetReady(roomID string, userID uint64, ready bool) (DouNiuRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return DouNiuRoom{}, false, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return DouNiuRoom{}, false, ErrNotInRoom
	}
	if st.room.Status != "waiting" {
		return st.room, false, ErrRoomNotWaiting
	}

	st.room.Players[seat].Ready = ready
	allReady := len(st.room.Players) >= MinDouNiuRoomPlayers
	if allReady {
		for _, p := range st.room.Players {
			if p.UserID != st.room.HostUserID && !p.Ready {
				allReady = false
				break
			}
		}
	} else {
		allReady = false
	}
	return st.room, allReady, nil
}

func (s *DouNiuRoomService) Start(roomID string, userID uint64, games *DouNiuService) (DouNiuRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return DouNiuRoom{}, ErrRoomNotFound
	}
	if _, ok := st.userSeats[userID]; !ok {
		return DouNiuRoom{}, ErrNotInRoom
	}
	if st.room.HostUserID != userID {
		return DouNiuRoom{}, ErrNotRoomHost
	}
	if st.room.Status != "waiting" {
		return st.room, ErrRoomNotWaiting
	}
	if len(st.room.Players) < MinDouNiuRoomPlayers {
		return DouNiuRoom{}, ErrZhajinhuaNeedMorePlayers
	}
	for i := range st.room.Players {
		if st.room.Players[i].UserID == st.room.HostUserID {
			st.room.Players[i].Ready = true
		}
	}
	for _, p := range st.room.Players {
		if p.UserID != st.room.HostUserID && !p.Ready {
			return st.room, ErrNotAllReady
		}
	}

	userIDs := make([]uint64, len(st.room.Players))
	names := make([]string, len(st.room.Players))
	for i, p := range st.room.Players {
		userIDs[i] = p.UserID
		names[i] = p.Username
	}
	gameID, _, err := games.CreateOnlineGame(userIDs, names, "")
	if err != nil {
		return DouNiuRoom{}, err
	}
	st.room.Status = "playing"
	st.room.GameID = gameID
	for i := range st.room.Players {
		st.room.Players[i].Ready = false
	}
	return st.room, nil
}

func (s *DouNiuRoomService) SetReadyForNext(roomID string, userID uint64, ready bool) (DouNiuRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return DouNiuRoom{}, false, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return DouNiuRoom{}, false, ErrNotInRoom
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
	allReady := len(st.room.Players) >= MinDouNiuRoomPlayers
	if allReady {
		for _, p := range st.room.Players {
			if p.UserID != st.room.HostUserID && !p.Ready {
				allReady = false
				break
			}
		}
	} else {
		allReady = false
	}
	return st.room, allReady, nil
}

func (s *DouNiuRoomService) PlayersForGame(roomID string) ([]uint64, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st, ok := s.rooms[roomID]
	if !ok {
		return nil, nil, ErrRoomNotFound
	}
	if len(st.room.Players) < MinDouNiuRoomPlayers {
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

func (s *DouNiuRoomService) SetPlaying(roomID, gameID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.rooms[roomID]; ok {
		st.room.Status = "playing"
		st.room.GameID = gameID
	}
}

func (s *DouNiuRoomService) addPlayerLocked(st *douniuRoomState, userID uint64, username string) (DouNiuRoom, error) {
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

func (s *DouNiuRoomService) findByUserLocked(userID uint64) *douniuRoomState {
	for _, st := range s.rooms {
		if _, ok := st.userSeats[userID]; ok {
			return st
		}
	}
	return nil
}
