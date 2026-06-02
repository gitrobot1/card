package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/time/card/backend/internal/game/douniu"
	"github.com/time/card/backend/internal/service"
	cardws "github.com/time/card/backend/internal/ws"
)

type DouNiuHandler struct {
	Games *service.DouNiuService
	Rooms *service.DouNiuRoomService
	Hub   *cardws.DouNiuHub
}

type douniuStartRequest struct {
	BotCount       int    `json:"bot_count"`
	PreviousGameID string `json:"previous_game_id,omitempty"`
}

type douniuMultRequest struct {
	Multiplier int `json:"multiplier"`
}

type douniuReadyRequest struct {
	Ready bool `json:"ready"`
}

func (h *DouNiuHandler) Start(c *gin.Context) {
	var req douniuStartRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.BotCount < 1 {
		req.BotCount = 1
	}
	if req.BotCount > 7 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "电脑数量最多 7 个"})
		return
	}
	userID, username := currentUser(c)
	state, err := h.Games.CreateGame(userID, username, req.BotCount, req.PreviousGameID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *DouNiuHandler) GetState(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.GetState(c.Param("gameId"), userID)
	if err != nil {
		writeDouNiuError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *DouNiuHandler) GrabBanker(c *gin.Context) {
	var req douniuMultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.GrabBanker(c.Param("gameId"), userID, req.Multiplier)
	if err != nil {
		writeDouNiuError(c, err)
		return
	}
	h.broadcastGame(c.Param("gameId"), state.Events, userID)
	c.JSON(http.StatusOK, state)
}

func (h *DouNiuHandler) PlaceBet(c *gin.Context) {
	var req douniuMultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.PlaceBet(c.Param("gameId"), userID, req.Multiplier)
	if err != nil {
		writeDouNiuError(c, err)
		return
	}
	h.broadcastGame(c.Param("gameId"), state.Events, userID)
	c.JSON(http.StatusOK, state)
}

func (h *DouNiuHandler) Tick(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.Tick(c.Param("gameId"), userID)
	if err != nil {
		writeDouNiuError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *DouNiuHandler) JoinRoom(c *gin.Context) {
	userID, username := currentUser(c)
	var req joinRoomRequest
	_ = c.ShouldBindJSON(&req)
	var room service.DouNiuRoom
	var err error
	if req.RoomID != "" {
		room, err = h.Rooms.JoinRoom(req.RoomID, userID, username)
	} else {
		room, err = h.Rooms.Join(userID, username)
	}
	if err != nil {
		writeRoomError(c, err)
		return
	}
	h.broadcastRoom(room)
	c.JSON(http.StatusOK, room)
}

func (h *DouNiuHandler) GetRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Get(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *DouNiuHandler) LeaveRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Leave(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	if room.ID != "" {
		h.broadcastRoom(room)
	}
	c.JSON(http.StatusOK, room)
}

func (h *DouNiuHandler) ReadyRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	var req douniuReadyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Ready = true
	}
	room, _, err := h.Rooms.SetReady(c.Param("roomId"), userID, req.Ready)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	h.broadcastRoom(room)
	c.JSON(http.StatusOK, room)
}

func (h *DouNiuHandler) StartRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Start(c.Param("roomId"), userID, h.Games)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	h.broadcastRoom(room)
	c.JSON(http.StatusOK, room)
}

func (h *DouNiuHandler) ReadyNext(c *gin.Context) {
	userID, _ := currentUser(c)
	roomID := c.Param("roomId")

	var req douniuReadyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	roomBefore, err := h.Rooms.Get(roomID, userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	oldGameID := roomBefore.GameID

	room, allReady, err := h.Rooms.SetReadyForNext(roomID, userID, req.Ready)
	if err != nil {
		writeRoomError(c, err)
		return
	}

	if allReady {
		userIDs, names, err := h.Rooms.PlayersForGame(roomID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		gameID, _, err := h.Games.CreateOnlineGame(userIDs, names, oldGameID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.Rooms.SetPlaying(roomID, gameID)
		room, _ = h.Rooms.Get(roomID, userID)
	}

	h.broadcastRoom(room)
	c.JSON(http.StatusOK, room)
}

func writeDouNiuError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "对局不存在或已过期，请重新开局"})
	case errors.Is(err, service.ErrNotInGame):
		c.JSON(http.StatusForbidden, gin.H{"error": "你不在这局游戏中"})
	case errors.Is(err, douniu.ErrWrongPhase):
		c.JSON(http.StatusConflict, gin.H{"error": "当前阶段不能执行此操作"})
	case errors.Is(err, douniu.ErrInvalidGrab):
		c.JSON(http.StatusBadRequest, gin.H{"error": "抢庄倍数无效"})
	case errors.Is(err, douniu.ErrInvalidBet):
		c.JSON(http.StatusBadRequest, gin.H{"error": "下注倍数无效"})
	case errors.Is(err, douniu.ErrAlreadyActed):
		c.JSON(http.StatusConflict, gin.H{"error": "你已经操作过了"})
	case errors.Is(err, douniu.ErrGameOver):
		c.JSON(http.StatusConflict, gin.H{"error": "本局已结束"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
