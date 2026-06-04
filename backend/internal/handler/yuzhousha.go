package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/service"
)

type YuzhoushaHandler struct {
	Games *service.YuzhoushaService
}

type yzsPlayRequest struct {
	CardID       string `json:"card_id"`
	TargetIndex  int    `json:"target_index"`
	TargetZone   string `json:"target_zone"`
	TargetCardID string `json:"target_card_id"`
}

type yzsRespondRequest struct {
	CardID string `json:"card_id"`
}

type yzsDiscardRequest struct {
	CardIDs []string `json:"card_ids"`
}

type yzsStartRequest struct {
	CharacterID string `json:"character_id"`
	Mode        string `json:"mode"`
}

type yzsSkillRequest struct {
	SkillID      string   `json:"skill_id"`
	TargetIndex  int      `json:"target_index"`
	CardIDs      []string `json:"card_ids"`
	TargetZone   string   `json:"target_zone"`
	TargetCardID string   `json:"target_card_id"`
}

func (h *YuzhoushaHandler) Heroes(c *gin.Context) {
	q := engine.HeroesQuery{
		Mode:    c.Query("mode"),
		Kingdom: c.Query("kingdom"),
		Pack:    c.Query("pack"),
		Page:    parseQueryInt(c.Query("page"), 1),
		PageSize: parseQueryInt(c.Query("page_size"), 20),
	}
	c.JSON(http.StatusOK, h.Games.ListHeroes(q))
}

func parseQueryInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func (h *YuzhoushaHandler) Modes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"modes": h.Games.ModesCatalog()})
}

func (h *YuzhoushaHandler) Packs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"packs": h.Games.PacksCatalog()})
}

func (h *YuzhoushaHandler) Start(c *gin.Context) {
	var req yzsStartRequest
	_ = c.ShouldBindJSON(&req)
	userID, username := currentUser(c)
	state, err := h.Games.CreateSolo(userID, username, req.CharacterID, req.Mode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) UseSkill(c *gin.Context) {
	var req yzsSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SkillID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.UseSkill(c.Param("gameId"), userID, engine.UseSkillRequest{
		SkillID:      req.SkillID,
		TargetIndex:  req.TargetIndex,
		CardIDs:      req.CardIDs,
		TargetZone:   req.TargetZone,
		TargetCardID: req.TargetCardID,
	})
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) GetState(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.GetState(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) PlayCard(c *gin.Context) {
	var req yzsPlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.CardID == "" && req.TargetZone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.PlayCard(c.Param("gameId"), userID, req.CardID, engine.PlayTarget{
		SeatIndex: req.TargetIndex,
		Zone:      req.TargetZone,
		CardID:    req.TargetCardID,
	})
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) RespondShan(c *gin.Context) {
	h.RespondCard(c)
}

func (h *YuzhoushaHandler) RespondCard(c *gin.Context) {
	var req yzsRespondRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.CardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.RespondCard(c.Param("gameId"), userID, req.CardID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) PassResponse(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.PassResponse(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) BaguaJudge(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.TryBaguaJudge(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) EndPlay(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.EndPlay(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) DiscardCard(c *gin.Context) {
	var req yzsDiscardRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.CardIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.DiscardCards(c.Param("gameId"), userID, req.CardIDs)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

type yzsPeekDeckRequest struct {
	TopCardIDs    []string `json:"top_card_ids"`
	BottomCardIDs []string `json:"bottom_card_ids"`
}

type yzsGuanxingRequest = yzsPeekDeckRequest

func (h *YuzhoushaHandler) PassPrepare(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.PassPrepare(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) PassDraw(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.PassDraw(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) FinishPeekDeck(c *gin.Context) {
	var req yzsPeekDeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.FinishPeekDeck(c.Param("gameId"), userID, engine.PeekDeckRequest{
		TopCardIDs:    req.TopCardIDs,
		BottomCardIDs: req.BottomCardIDs,
	})
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *YuzhoushaHandler) FinishGuanxing(c *gin.Context) {
	h.FinishPeekDeck(c)
}

func (h *YuzhoushaHandler) Tick(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.Tick(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func writeYuzhoushaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "对局不存在或已过期"})
	case errors.Is(err, service.ErrNotInGame):
		c.JSON(http.StatusForbidden, gin.H{"error": "你不在这局游戏中"})
	case errors.Is(err, engine.ErrWrongPhase):
		c.JSON(http.StatusConflict, gin.H{"error": "当前阶段不能执行此操作"})
	case errors.Is(err, engine.ErrNotYourTurn):
		c.JSON(http.StatusConflict, gin.H{"error": "还没轮到你"})
	case errors.Is(err, engine.ErrInvalidCard):
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能出这张牌"})
	case errors.Is(err, engine.ErrInvalidDiscardCount):
		c.JSON(http.StatusBadRequest, gin.H{"error": "弃牌数量不正确，需一次弃掉全部应弃的牌"})
	case errors.Is(err, engine.ErrInvalidTarget):
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效目标"})
	case errors.Is(err, engine.ErrAlreadyActed):
		c.JSON(http.StatusConflict, gin.H{"error": "本回合已出过杀"})
	case errors.Is(err, engine.ErrPendingCombat):
		c.JSON(http.StatusConflict, gin.H{"error": "请先响应杀"})
	case errors.Is(err, engine.ErrNoPendingCombat):
		c.JSON(http.StatusConflict, gin.H{"error": "当前没有待响应的杀"})
	case errors.Is(err, engine.ErrGameOver):
		c.JSON(http.StatusConflict, gin.H{"error": "本局已结束"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
