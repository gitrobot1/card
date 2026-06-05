package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/service"
	cardws "github.com/time/card/backend/internal/ws"
)

var yuzhoushaUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type YuzhoushaWSHandler struct {
	Auth  *service.AuthService
	Games *service.YuzhoushaService
	Rooms *service.YuzhoushaRoomService
	Hub   *cardws.YuzhoushaHub
}

func (h *YuzhoushaWSHandler) authUserID(c *gin.Context) (uint64, bool) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		header := c.GetHeader("Authorization")
		if parts := strings.SplitN(header, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token = parts[1]
		}
	}
	if token == "" {
		return 0, false
	}
	claims, err := h.Auth.ParseToken(token)
	if err != nil {
		return 0, false
	}
	return claims.UserID, true
}

func (h *YuzhoushaWSHandler) Room(c *gin.Context) {
	userID, ok := h.authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
		return
	}

	roomID := c.Param("roomId")
	if _, err := h.Rooms.Get(roomID, userID); err != nil {
		writeRoomError(c, err)
		return
	}

	conn, err := yuzhoushaUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.Hub.RegisterRoom(roomID, userID, conn)
	defer h.Hub.UnregisterRoom(roomID, userID, conn)

	if room, err := h.Rooms.Get(roomID, userID); err == nil {
		h.Hub.SendRoom(conn, room)
	}

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *YuzhoushaWSHandler) Game(c *gin.Context) {
	userID, ok := h.authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
		return
	}

	gameID := c.Param("gameId")
	state, err := h.Games.SnapshotForUser(gameID, userID, nil)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}

	conn, err := yuzhoushaUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.Hub.RegisterGame(gameID, userID, conn)
	defer h.Hub.UnregisterGame(gameID, userID, conn)

	h.Hub.SendGameState(conn, state)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *YuzhoushaHandler) broadcastRoom(room service.YuzhoushaRoom) {
	if h.Hub != nil {
		h.Hub.BroadcastRoom(room)
	}
}

func (h *YuzhoushaHandler) broadcastGame(gameID string, events []engine.GameEvent, actorUserID uint64) {
	if h.Hub == nil || gameID == "" {
		return
	}
	h.Hub.BroadcastGame(h.Games, gameID, events, actorUserID)
}

func (h *YuzhoushaHandler) writeGameResponse(c *gin.Context, gameID string, userID uint64, state engine.PublicState) {
	h.broadcastGame(gameID, state.Events, userID)
	c.JSON(http.StatusOK, state)
}
