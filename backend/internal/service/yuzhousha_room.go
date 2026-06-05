package service

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

var (
	ErrHeroNotSelected   = errors.New("hero not selected")
	ErrDuplicateHero     = errors.New("duplicate hero in room")
	ErrOnlineModeUnknown = errors.New("online mode not supported")
)

type YuzhoushaRoomPlayer struct {
	UserID      uint64 `json:"user_id"`
	Username    string `json:"username"`
	Ready       bool   `json:"ready"`
	CharacterID string `json:"character_id,omitempty"`
}

type YuzhoushaRoom struct {
	ID         string                `json:"id"`
	Mode       string                `json:"mode"`
	Status     string                `json:"status"`
	GameID     string                `json:"game_id,omitempty"`
	HostUserID uint64                `json:"host_user_id"`
	Players    []YuzhoushaRoomPlayer `json:"players"`
}

type YuzhoushaRoomService struct {
	mu    sync.RWMutex
	rooms map[string]*yuzhoushaRoomState
}

type yuzhoushaRoomState struct {
	room      YuzhoushaRoom
	userSeats map[uint64]int
}

func NewYuzhoushaRoomService() *YuzhoushaRoomService {
	return &YuzhoushaRoomService{rooms: make(map[string]*yuzhoushaRoomState)}
}

func yzsRoomPlayerCount(roomMode string) int {
	if m, ok := mode.Lookup(roomMode); ok && m.PlayerCount > 0 {
		return m.PlayerCount
	}
	return 2
}

func normalizeYzsRoomMode(roomMode string) string {
	return mode.NormalizeID(roomMode)
}

func (s *YuzhoushaRoomService) Join(userID uint64, username string, roomMode string) (YuzhoushaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.findByUserLocked(userID); existing != nil {
		return existing.room, nil
	}
	normalized := normalizeYzsRoomMode(roomMode)
	maxPlayers := yzsRoomPlayerCount(normalized)
	for _, st := range s.rooms {
		if st.room.Status == "waiting" && st.room.Mode == normalized && len(st.room.Players) < maxPlayers {
			return s.addPlayerLocked(st, userID, username)
		}
	}
	id := uuid.NewString()
	st := &yuzhoushaRoomState{
		room: YuzhoushaRoom{
			ID:     id,
			Mode:   normalized,
			Status: "waiting",
		},
		userSeats: map[uint64]int{},
	}
	s.rooms[id] = st
	return s.addPlayerLocked(st, userID, username)
}

func (s *YuzhoushaRoomService) JoinRoom(roomID string, userID uint64, username string) (YuzhoushaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing := s.findByUserLocked(userID); existing != nil {
		if existing.room.ID == roomID {
			return existing.room, nil
		}
		return YuzhoushaRoom{}, ErrNotInRoom
	}
	st, ok := s.rooms[roomID]
	if !ok {
		return YuzhoushaRoom{}, ErrRoomNotFound
	}
	if st.room.Status != "waiting" {
		return YuzhoushaRoom{}, ErrRoomNotWaiting
	}
	if len(st.room.Players) >= yzsRoomPlayerCount(st.room.Mode) {
		return YuzhoushaRoom{}, ErrRoomFull
	}
	return s.addPlayerLocked(st, userID, username)
}

func (s *YuzhoushaRoomService) Get(roomID string, userID uint64) (YuzhoushaRoom, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return YuzhoushaRoom{}, ErrRoomNotFound
	}
	if _, ok := st.userSeats[userID]; !ok {
		return YuzhoushaRoom{}, ErrNotInRoom
	}
	return st.room, nil
}

func (s *YuzhoushaRoomService) Leave(roomID string, userID uint64) (YuzhoushaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return YuzhoushaRoom{}, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return YuzhoushaRoom{}, ErrNotInRoom
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
		return YuzhoushaRoom{}, nil
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

func (s *YuzhoushaRoomService) SetHero(roomID string, userID uint64, characterID string) (YuzhoushaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return YuzhoushaRoom{}, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return YuzhoushaRoom{}, ErrNotInRoom
	}
	if st.room.Status != "waiting" {
		return YuzhoushaRoom{}, ErrRoomNotWaiting
	}
	if characterID == "" {
		return YuzhoushaRoom{}, ErrHeroNotSelected
	}
	for i, p := range st.room.Players {
		if i != seat && p.CharacterID == characterID {
			return YuzhoushaRoom{}, ErrDuplicateHero
		}
	}
	st.room.Players[seat].CharacterID = characterID
	st.room.Players[seat].Ready = false
	return st.room, nil
}

func (s *YuzhoushaRoomService) SetReady(roomID string, userID uint64, ready bool) (YuzhoushaRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return YuzhoushaRoom{}, false, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return YuzhoushaRoom{}, false, ErrNotInRoom
	}
	if st.room.Status != "waiting" {
		return st.room, false, ErrRoomNotWaiting
	}
	if ready && st.room.Players[seat].CharacterID == "" {
		return st.room, false, ErrHeroNotSelected
	}
	st.room.Players[seat].Ready = ready
	return st.room, s.allPlayersReadyLocked(st), nil
}

func (s *YuzhoushaRoomService) Start(roomID string, userID uint64, games *YuzhoushaService) (YuzhoushaRoom, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return YuzhoushaRoom{}, ErrRoomNotFound
	}
	if _, ok := st.userSeats[userID]; !ok {
		return YuzhoushaRoom{}, ErrNotInRoom
	}
	if st.room.HostUserID != userID {
		return YuzhoushaRoom{}, ErrNotRoomHost
	}
	if st.room.Status != "waiting" {
		return st.room, ErrRoomNotWaiting
	}
	required := yzsRoomPlayerCount(st.room.Mode)
	if len(st.room.Players) < required {
		return st.room, ErrZhajinhuaNeedMorePlayers
	}
	for i := range st.room.Players {
		if st.room.Players[i].UserID == st.room.HostUserID {
			st.room.Players[i].Ready = true
		}
	}
	if !s.allPlayersReadyLocked(st) {
		return st.room, ErrNotAllReady
	}
	seenHeroes := map[string]bool{}
	userIDs := make([]uint64, len(st.room.Players))
	names := make([]string, len(st.room.Players))
	charIDs := make([]string, len(st.room.Players))
	for i, p := range st.room.Players {
		if p.CharacterID == "" {
			return st.room, ErrHeroNotSelected
		}
		if seenHeroes[p.CharacterID] {
			return st.room, ErrDuplicateHero
		}
		seenHeroes[p.CharacterID] = true
		userIDs[i] = p.UserID
		names[i] = p.Username
		charIDs[i] = p.CharacterID
	}
	gameID, err := games.CreateOnlineGame(st.room.Mode, userIDs, names, charIDs)
	if err != nil {
		return YuzhoushaRoom{}, err
	}
	st.room.Status = "playing"
	st.room.GameID = gameID
	for i := range st.room.Players {
		st.room.Players[i].Ready = false
	}
	return st.room, nil
}

func (s *YuzhoushaRoomService) SetReadyForNext(roomID string, userID uint64, ready bool) (YuzhoushaRoom, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.rooms[roomID]
	if !ok {
		return YuzhoushaRoom{}, false, ErrRoomNotFound
	}
	seat, ok := st.userSeats[userID]
	if !ok {
		return YuzhoushaRoom{}, false, ErrNotInRoom
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
	return st.room, s.allPlayersReadyLocked(st), nil
}

func (s *YuzhoushaRoomService) allPlayersReadyLocked(st *yuzhoushaRoomState) bool {
	required := yzsRoomPlayerCount(st.room.Mode)
	if len(st.room.Players) < required {
		return false
	}
	for _, p := range st.room.Players {
		if p.UserID != st.room.HostUserID && !p.Ready {
			return false
		}
		if p.CharacterID == "" {
			return false
		}
	}
	return true
}

func (s *YuzhoushaRoomService) addPlayerLocked(st *yuzhoushaRoomState, userID uint64, username string) (YuzhoushaRoom, error) {
	for _, p := range st.room.Players {
		if p.UserID == userID {
			return st.room, nil
		}
	}
	maxPlayers := yzsRoomPlayerCount(st.room.Mode)
	if len(st.room.Players) >= maxPlayers {
		return YuzhoushaRoom{}, ErrRoomFull
	}
	seat := len(st.room.Players)
	st.room.Players = append(st.room.Players, YuzhoushaRoomPlayer{UserID: userID, Username: username})
	st.userSeats[userID] = seat
	if seat == 0 {
		st.room.HostUserID = userID
	}
	return st.room, nil
}

func (s *YuzhoushaRoomService) findByUserLocked(userID uint64) *yuzhoushaRoomState {
	for _, st := range s.rooms {
		if _, ok := st.userSeats[userID]; ok {
			return st
		}
	}
	return nil
}
