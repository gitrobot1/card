package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/time/card/backend/internal/game/uno"
	"github.com/time/card/backend/internal/service"
)

type UnoHandler struct {
	Games *service.UnoService
	Rooms *service.UnoRoomService
}

type unoStartRequest struct {
	BotCount int `json:"bot_count"`
}

type unoPlayRequest struct {
	CardID string    `json:"card_id"`
	Color  uno.Color `json:"color,omitempty"`
}

func (h *UnoHandler) Start(c *gin.Context) {
	var req unoStartRequest
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

func (h *UnoHandler) GetState(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.GetState(c.Param("gameId"), userID)
	if err != nil {
		writeUnoError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *UnoHandler) Play(c *gin.Context) {
	var req unoPlayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.CardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要出的牌"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.Play(c.Param("gameId"), userID, req.CardID, req.Color)
	if err != nil {
		writeUnoError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *UnoHandler) Draw(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.Draw(c.Param("gameId"), userID)
	if err != nil {
		writeUnoError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *UnoHandler) VoteEnd(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.VoteEnd(c.Param("gameId"), userID)
	if err != nil {
		writeUnoError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *UnoHandler) RollFirst(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.RollFirst(c.Param("gameId"), userID)
	if err != nil {
		writeUnoError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *UnoHandler) Tick(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.Tick(c.Param("gameId"), userID)
	if err != nil {
		writeUnoError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

type unoReadyRequest struct {
	Ready bool `json:"ready"`
}

func (h *UnoHandler) JoinRoom(c *gin.Context) {
	userID, username := currentUser(c)
	var req joinRoomRequest
	_ = c.ShouldBindJSON(&req)
	var room service.UnoRoom
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

func (h *UnoHandler) GetRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Get(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *UnoHandler) LeaveRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Leave(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *UnoHandler) ReadyRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	var req unoReadyRequest
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

func (h *UnoHandler) StartRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Start(c.Param("roomId"), userID, h.Games)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *UnoHandler) ReadyNext(c *gin.Context) {
	userID, _ := currentUser(c)
	roomID := c.Param("roomId")

	var req unoReadyRequest
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

func writeUnoError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "对局不存在或已过期，请重新开局"})
	case errors.Is(err, service.ErrNotInGame):
		c.JSON(http.StatusForbidden, gin.H{"error": "你不在这局游戏中"})
	case errors.Is(err, uno.ErrNotYourTurn):
		c.JSON(http.StatusConflict, gin.H{"error": "还没轮到你"})
	case errors.Is(err, uno.ErrCannotPlay):
		c.JSON(http.StatusBadRequest, gin.H{"error": "这张牌现在不能出"})
	case errors.Is(err, uno.ErrLastCardMustBeBasic):
		c.JSON(http.StatusBadRequest, gin.H{"error": "最后一张必须是数字牌，功能牌不能打出"})
	case errors.Is(err, uno.ErrInvalidCard):
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的牌"})
	case errors.Is(err, uno.ErrInvalidColor):
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择有效颜色"})
	case errors.Is(err, uno.ErrWrongPhase):
		c.JSON(http.StatusConflict, gin.H{"error": "当前阶段不能执行此操作"})
	case errors.Is(err, uno.ErrGameOver):
		c.JSON(http.StatusConflict, gin.H{"error": "本局已结束"})
	case errors.Is(err, uno.ErrCannotVoteEnd):
		c.JSON(http.StatusConflict, gin.H{"error": "当前不能投票结束"})
	case errors.Is(err, uno.ErrAlreadyVoted):
		c.JSON(http.StatusConflict, gin.H{"error": "你已经同意结束了"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
