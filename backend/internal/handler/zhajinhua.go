package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/time/card/backend/internal/game/zhajinhua"
	"github.com/time/card/backend/internal/service"
)

type ZhajinhuaHandler struct {
	Games *service.ZhajinhuaService
	Rooms *service.ZhajinhuaRoomService
}

type zhajinhuaStartRequest struct {
	BotCount int `json:"bot_count"`
}

type zhajinhuaRaiseRequest struct {
	Amount int `json:"amount"`
}

type zhajinhuaCompareRequest struct {
	TargetIndex int `json:"target_index"`
}

type zhajinhuaReadyRequest struct {
	Ready bool `json:"ready"`
}

func (h *ZhajinhuaHandler) Start(c *gin.Context) {
	var req zhajinhuaStartRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.BotCount < 1 {
		req.BotCount = 1
	}
	if req.BotCount > 7 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "电脑数量最多 7 个"})
		return
	}
	userID, username := currentUser(c)
	state, err := h.Games.CreateGame(userID, username, req.BotCount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *ZhajinhuaHandler) GetState(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.GetState(c.Param("gameId"), userID)
	if err != nil {
		writeZhajinhuaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *ZhajinhuaHandler) Look(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.Look(c.Param("gameId"), userID)
	if err != nil {
		writeZhajinhuaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *ZhajinhuaHandler) Fold(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.Fold(c.Param("gameId"), userID)
	if err != nil {
		writeZhajinhuaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *ZhajinhuaHandler) Follow(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.Follow(c.Param("gameId"), userID)
	if err != nil {
		writeZhajinhuaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *ZhajinhuaHandler) Raise(c *gin.Context) {
	var req zhajinhuaRaiseRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid raise amount"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.Raise(c.Param("gameId"), userID, req.Amount)
	if err != nil {
		writeZhajinhuaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *ZhajinhuaHandler) Compare(c *gin.Context) {
	var req zhajinhuaCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.Compare(c.Param("gameId"), userID, req.TargetIndex)
	if err != nil {
		writeZhajinhuaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *ZhajinhuaHandler) Tick(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.Tick(c.Param("gameId"), userID)
	if err != nil {
		writeZhajinhuaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *ZhajinhuaHandler) JoinRoom(c *gin.Context) {
	userID, username := currentUser(c)
	var req joinRoomRequest
	_ = c.ShouldBindJSON(&req)
	var room service.ZhajinhuaRoom
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
	c.JSON(http.StatusOK, room)
}

func (h *ZhajinhuaHandler) GetRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Get(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *ZhajinhuaHandler) LeaveRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Leave(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *ZhajinhuaHandler) ReadyRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	var req zhajinhuaReadyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Ready = true
	}
	room, _, err := h.Rooms.SetReady(c.Param("roomId"), userID, req.Ready)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *ZhajinhuaHandler) StartRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Start(c.Param("roomId"), userID, h.Games)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *ZhajinhuaHandler) ReadyNext(c *gin.Context) {
	userID, _ := currentUser(c)
	roomID := c.Param("roomId")

	var req zhajinhuaReadyRequest
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
		userIDs, names, err := h.Rooms.PlayersForGame(roomID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		gameID, _, err := h.Games.CreateOnlineGame(userIDs, names)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.Rooms.SetPlaying(roomID, gameID)
		room, _ = h.Rooms.Get(roomID, userID)
	}

	c.JSON(http.StatusOK, room)
}

func writeZhajinhuaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
	case errors.Is(err, service.ErrNotInGame):
		c.JSON(http.StatusForbidden, gin.H{"error": "你不在这局游戏中"})
	case errors.Is(err, zhajinhua.ErrNotYourTurn):
		c.JSON(http.StatusConflict, gin.H{"error": "还没轮到你"})
	case errors.Is(err, zhajinhua.ErrInvalidAction):
		c.JSON(http.StatusBadRequest, gin.H{"error": "操作不允许"})
	case errors.Is(err, zhajinhua.ErrCompareNeedLook):
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先点击「看牌」后再比牌"})
	case errors.Is(err, zhajinhua.ErrCompareTargetNeedLook):
		c.JSON(http.StatusBadRequest, gin.H{"error": "对方尚未看牌，无法比牌"})
	case errors.Is(err, zhajinhua.ErrInsufficientChips):
		c.JSON(http.StatusBadRequest, gin.H{"error": "筹码不足"})
	case errors.Is(err, zhajinhua.ErrTargetInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "比牌目标无效"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
