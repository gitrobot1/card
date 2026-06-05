package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/service"
)

type YuzhoushaHub struct {
	mu      sync.RWMutex
	rooms   map[string]map[uint64]*websocket.Conn
	games   map[string]map[uint64]*websocket.Conn
	writeMu map[*websocket.Conn]*sync.Mutex
}

func NewYuzhoushaHub() *YuzhoushaHub {
	return &YuzhoushaHub{
		rooms:   make(map[string]map[uint64]*websocket.Conn),
		games:   make(map[string]map[uint64]*websocket.Conn),
		writeMu: make(map[*websocket.Conn]*sync.Mutex),
	}
}

type yuzhoushaRoomMessage struct {
	Type string                `json:"type"`
	Room service.YuzhoushaRoom `json:"room,omitempty"`
}

type yuzhoushaGameMessage struct {
	Type  string             `json:"type"`
	State engine.PublicState `json:"state,omitempty"`
}

func (h *YuzhoushaHub) connLock(conn *websocket.Conn) *sync.Mutex {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m, ok := h.writeMu[conn]; ok {
		return m
	}
	m := &sync.Mutex{}
	h.writeMu[conn] = m
	return m
}

func (h *YuzhoushaHub) writeJSON(conn *websocket.Conn, payload any) error {
	h.connLock(conn).Lock()
	defer h.connLock(conn).Unlock()
	return conn.WriteJSON(payload)
}

func (h *YuzhoushaHub) RegisterRoom(roomID string, userID uint64, conn *websocket.Conn) {
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

func (h *YuzhoushaHub) RegisterGame(gameID string, userID uint64, conn *websocket.Conn) {
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

func (h *YuzhoushaHub) UnregisterRoom(roomID string, userID uint64, conn *websocket.Conn) {
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

func (h *YuzhoushaHub) UnregisterGame(gameID string, userID uint64, conn *websocket.Conn) {
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

func (h *YuzhoushaHub) BroadcastRoom(room service.YuzhoushaRoom) {
	h.mu.RLock()
	group := h.rooms[room.ID]
	conns := make([]*websocket.Conn, 0, len(group))
	for _, conn := range group {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	msg := yuzhoushaRoomMessage{Type: "room", Room: room}
	for _, conn := range conns {
		_ = h.writeJSON(conn, msg)
	}
}

func (h *YuzhoushaHub) BroadcastGame(games *service.YuzhoushaService, gameID string, events []engine.GameEvent, excludeUserID uint64) {
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
		_ = h.writeJSON(conn, yuzhoushaGameMessage{Type: "game_state", State: state})
	}
}

func (h *YuzhoushaHub) SendGameState(conn *websocket.Conn, state engine.PublicState) {
	_ = h.writeJSON(conn, yuzhoushaGameMessage{Type: "game_state", State: state})
}

func (h *YuzhoushaHub) SendRoom(conn *websocket.Conn, room service.YuzhoushaRoom) {
	_ = h.writeJSON(conn, yuzhoushaRoomMessage{Type: "room", Room: room})
}

func (h *YuzhoushaHub) SendError(conn *websocket.Conn, message string) {
	_ = h.writeJSON(conn, map[string]string{"type": "error", "error": message})
}
