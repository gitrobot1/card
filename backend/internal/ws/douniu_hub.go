package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/time/card/backend/internal/game/douniu"
	"github.com/time/card/backend/internal/service"
)

type DouNiuHub struct {
	mu      sync.RWMutex
	rooms   map[string]map[uint64]*websocket.Conn
	games   map[string]map[uint64]*websocket.Conn
	writeMu map[*websocket.Conn]*sync.Mutex
}

func NewDouNiuHub() *DouNiuHub {
	return &DouNiuHub{
		rooms:   make(map[string]map[uint64]*websocket.Conn),
		games:   make(map[string]map[uint64]*websocket.Conn),
		writeMu: make(map[*websocket.Conn]*sync.Mutex),
	}
}

type douniuRoomMessage struct {
	Type string            `json:"type"`
	Room service.DouNiuRoom `json:"room,omitempty"`
}

type douniuGameMessage struct {
	Type  string             `json:"type"`
	State douniu.PublicState `json:"state,omitempty"`
}

func (h *DouNiuHub) connLock(conn *websocket.Conn) *sync.Mutex {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m, ok := h.writeMu[conn]; ok {
		return m
	}
	m := &sync.Mutex{}
	h.writeMu[conn] = m
	return m
}

func (h *DouNiuHub) writeJSON(conn *websocket.Conn, payload any) error {
	h.connLock(conn).Lock()
	defer h.connLock(conn).Unlock()
	return conn.WriteJSON(payload)
}

func (h *DouNiuHub) RegisterRoom(roomID string, userID uint64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[uint64]*websocket.Conn)
	}
	if old, ok := h.rooms[roomID][userID]; ok && old != conn {
		_ = old.Close()
		delete(h.writeMu, old)
	}
	h.rooms[roomID][userID] = conn
}

func (h *DouNiuHub) RegisterGame(gameID string, userID uint64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.games[gameID] == nil {
		h.games[gameID] = make(map[uint64]*websocket.Conn)
	}
	if old, ok := h.games[gameID][userID]; ok && old != conn {
		_ = old.Close()
		delete(h.writeMu, old)
	}
	h.games[gameID][userID] = conn
}

func (h *DouNiuHub) UnregisterRoom(roomID string, userID uint64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	group, ok := h.rooms[roomID]
	if !ok {
		return
	}
	if group[userID] == conn {
		delete(group, userID)
		delete(h.writeMu, conn)
	}
	if len(group) == 0 {
		delete(h.rooms, roomID)
	}
}

func (h *DouNiuHub) UnregisterGame(gameID string, userID uint64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	group, ok := h.games[gameID]
	if !ok {
		return
	}
	if group[userID] == conn {
		delete(group, userID)
		delete(h.writeMu, conn)
	}
	if len(group) == 0 {
		delete(h.games, gameID)
	}
}

func (h *DouNiuHub) BroadcastRoom(room service.DouNiuRoom) {
	h.mu.RLock()
	group := h.rooms[room.ID]
	conns := make([]*websocket.Conn, 0, len(group))
	for _, conn := range group {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	msg := douniuRoomMessage{Type: "room", Room: room}
	for _, conn := range conns {
		_ = h.writeJSON(conn, msg)
	}
}

func (h *DouNiuHub) BroadcastGame(games *service.DouNiuService, gameID string, events []douniu.GameEvent, excludeUserID uint64) {
	members, err := games.MemberUserIDs(gameID)
	if err != nil {
		return
	}

	h.mu.RLock()
	group := h.games[gameID]
	targets := make(map[uint64]*websocket.Conn, len(members))
	for _, uid := range members {
		if uid == excludeUserID {
			continue
		}
		if conn, ok := group[uid]; ok {
			targets[uid] = conn
		}
	}
	h.mu.RUnlock()

	for uid, conn := range targets {
		state, err := games.SnapshotForUser(gameID, uid, events)
		if err != nil {
			continue
		}
		_ = h.writeJSON(conn, douniuGameMessage{Type: "game_state", State: state})
	}
}

func (h *DouNiuHub) SendGameState(conn *websocket.Conn, state douniu.PublicState) {
	_ = h.writeJSON(conn, douniuGameMessage{Type: "game_state", State: state})
}

func (h *DouNiuHub) SendRoom(conn *websocket.Conn, room service.DouNiuRoom) {
	_ = h.writeJSON(conn, douniuRoomMessage{Type: "room", Room: room})
}

func (h *DouNiuHub) SendError(conn *websocket.Conn, message string) {
	_ = h.writeJSON(conn, map[string]string{"type": "error", "error": message})
}

func EncodeRoom(room service.DouNiuRoom) ([]byte, error) {
	return json.Marshal(douniuRoomMessage{Type: "room", Room: room})
}
