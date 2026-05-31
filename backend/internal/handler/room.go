package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/time/card/backend/internal/middleware"
	"github.com/time/card/backend/internal/service"
)

type RoomHandler struct {
	Rooms    *service.DouDizhuRoomService
	DouDizhu *service.DouDizhuService
}

type joinRoomRequest struct {
	RoomID string `json:"room_id"`
}

type readyRequest struct {
	Ready bool `json:"ready"`
}

func (h *RoomHandler) Join(c *gin.Context) {
	userID, username := currentUser(c)
	var req joinRoomRequest
	_ = c.ShouldBindJSON(&req)

	var (
		room service.DouDizhuRoom
		err  error
	)
	if req.RoomID != "" {
		room, err = h.Rooms.JoinRoom(req.RoomID, userID, username)
	} else {
		room, err = h.Rooms.Join(userID, username)
	}
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *RoomHandler) Get(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Get(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *RoomHandler) Leave(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Leave(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	if room.ID == "" {
		c.JSON(http.StatusOK, gin.H{"left": true})
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *RoomHandler) Ready(c *gin.Context) {
	userID, _ := currentUser(c)
	roomID := c.Param("roomId")

	var req readyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	room, allReady, err := h.Rooms.SetReady(roomID, userID, req.Ready)
	if err != nil {
		writeRoomError(c, err)
		return
	}

	if allReady {
		names, userIDs, err := h.Rooms.PlayersForGame(roomID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		gameID, _, err := h.DouDizhu.CreateOnlineGame(userIDs, names)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := h.Rooms.MarkPlaying(roomID, gameID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		room, _ = h.Rooms.Get(roomID, userID)
	}

	c.JSON(http.StatusOK, room)
}

func (h *RoomHandler) Next(c *gin.Context) {
	userID, _ := currentUser(c)
	roomID := c.Param("roomId")

	var req readyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	room, allReady, err := h.Rooms.SetReadyForNext(roomID, userID, req.Ready)
	if err != nil {
		writeRoomError(c, err)
		return
	}

	if allReady {
		names, userIDs, err := h.Rooms.PlayersForGame(roomID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		gameID, _, err := h.DouDizhu.CreateOnlineGame(userIDs, names)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := h.Rooms.MarkPlaying(roomID, gameID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		room, _ = h.Rooms.Get(roomID, userID)
	}

	c.JSON(http.StatusOK, room)
}

func currentUser(c *gin.Context) (uint64, string) {
	rawID, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := rawID.(uint64)
	username, _ := c.Get(middleware.ContextUsernameKey)
	name, _ := username.(string)
	return userID, name
}

func writeRoomError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrRoomNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "房间不存在"})
	case errors.Is(err, service.ErrRoomFull):
		c.JSON(http.StatusConflict, gin.H{"error": "人满了，请稍后再试"})
	case errors.Is(err, service.ErrNotInRoom):
		c.JSON(http.StatusForbidden, gin.H{"error": "你不在该房间中"})
	case errors.Is(err, service.ErrRoomNotWaiting):
		c.JSON(http.StatusConflict, gin.H{"error": "房间已开始或不可加入"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
