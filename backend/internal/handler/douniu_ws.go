package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/time/card/backend/internal/game/douniu"
	"github.com/time/card/backend/internal/service"
	cardws "github.com/time/card/backend/internal/ws"
)

var douniuUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type DouNiuWSHandler struct {
	Auth  *service.AuthService
	Games *service.DouNiuService
	Rooms *service.DouNiuRoomService
	Hub   *cardws.DouNiuHub
}

func (h *DouNiuWSHandler) authUserID(c *gin.Context) (uint64, bool) {
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

func (h *DouNiuWSHandler) Room(c *gin.Context) {
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

	conn, err := douniuUpgrader.Upgrade(c.Writer, c.Request, nil)
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

func (h *DouNiuWSHandler) Game(c *gin.Context) {
	userID, ok := h.authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
		return
	}

	gameID := c.Param("gameId")
	state, err := h.Games.SnapshotForUser(gameID, userID, nil)
	if err != nil {
		writeDouNiuError(c, err)
		return
	}

	conn, err := douniuUpgrader.Upgrade(c.Writer, c.Request, nil)
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

func (h *DouNiuHandler) broadcastRoom(room service.DouNiuRoom) {
	if h.Hub != nil {
		h.Hub.BroadcastRoom(room)
	}
}

func (h *DouNiuHandler) broadcastGame(gameID string, events []douniu.GameEvent, actorUserID uint64) {
	if h.Hub == nil || gameID == "" {
		return
	}
	h.Hub.BroadcastGame(h.Games, gameID, events, actorUserID)
}
