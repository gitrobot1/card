package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/time/card/backend/internal/game/doudizhu"
	"github.com/time/card/backend/internal/service"
)

type GameHandler struct {
	DouDizhu *service.DouDizhuService
}

func (h *GameHandler) Catalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"games": h.DouDizhu.Catalog()})
}

func (h *GameHandler) StartDouDizhu(c *gin.Context) {
	userID, username := currentUser(c)
	state := h.DouDizhu.CreateGame(userID, username)
	c.JSON(http.StatusOK, state)
}

func (h *GameHandler) GetDouDizhuState(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.DouDizhu.GetState(c.Param("gameId"), userID)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

type callRequest struct {
	Call bool `json:"call"`
}

func (h *GameHandler) CallDouDizhu(c *gin.Context) {
	var req callRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, _ := currentUser(c)
	state, err := h.DouDizhu.Call(c.Param("gameId"), userID, req.Call)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

type playRequest struct {
	CardIDs []string `json:"card_ids"`
}

func (h *GameHandler) PlayDouDizhu(c *gin.Context) {
	var req playRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, _ := currentUser(c)
	state, err := h.DouDizhu.Play(c.Param("gameId"), userID, req.CardIDs)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *GameHandler) PassDouDizhu(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.DouDizhu.Pass(c.Param("gameId"), userID)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *GameHandler) HintDouDizhu(c *gin.Context) {
	userID, _ := currentUser(c)
	hint, err := h.DouDizhu.Hint(c.Param("gameId"), userID)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, hint)
}

func (h *GameHandler) TickDouDizhu(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.DouDizhu.Tick(c.Param("gameId"), userID)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func writeGameError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
	case errors.Is(err, service.ErrNotInGame):
		c.JSON(http.StatusForbidden, gin.H{"error": "你不在这局游戏中"})
	case errors.Is(err, doudizhu.ErrNotYourTurn):
		c.JSON(http.StatusConflict, gin.H{"error": "还没轮到你"})
	case errors.Is(err, doudizhu.ErrInvalidCards):
		c.JSON(http.StatusBadRequest, gin.H{"error": "出牌不合法"})
	case errors.Is(err, doudizhu.ErrInvalidPattern):
		c.JSON(http.StatusBadRequest, gin.H{"error": "牌型不正确"})
	case errors.Is(err, doudizhu.ErrCannotBeat):
		c.JSON(http.StatusBadRequest, gin.H{"error": "压不过上家"})
	case errors.Is(err, doudizhu.ErrMustPlay):
		c.JSON(http.StatusBadRequest, gin.H{"error": "本轮必须出牌"})
	case errors.Is(err, doudizhu.ErrWrongPhase):
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前阶段不能执行该操作"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
