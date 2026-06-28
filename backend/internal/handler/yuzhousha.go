package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/service"
	cardws "github.com/time/card/backend/internal/ws"
)

type YuzhoushaHandler struct {
	Games *service.YuzhoushaService
	Rooms *service.YuzhoushaRoomService
	Hub   *cardws.YuzhoushaHub
}

type yzsPlayRequest struct {
	CardID               string  `json:"card_id"`
	TargetIndex          int     `json:"target_index"`
	SecondTargetIndex    *int    `json:"second_target_index"`
	TargetZone           string  `json:"target_zone"`
	TargetCardID         string  `json:"target_card_id"`
	ZhangbaSecondCardID  string  `json:"zhangba_second_card_id"`  // 丈八蛇矛第二张牌
	FangtianExtraTargets []int   `json:"fangtian_extra_targets"`  // 方天画戟额外目标列表
}

type yzsRespondRequest struct {
	CardID string `json:"card_id"`
}

type yzsZhangbaRespondRequest struct {
	CardIDs []string `json:"card_ids"` // 丈八蛇矛响应：两张手牌ID
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

type yzsReadyRequest struct {
	Ready bool `json:"ready"`
}

type yzsSetHeroRequest struct {
	CharacterID string `json:"character_id"`
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

func (h *YuzhoushaHandler) JoinRoom(c *gin.Context) {
	userID, username := currentUser(c)
	var req joinRoomRequest
	_ = c.ShouldBindJSON(&req)
	var room service.YuzhoushaRoom
	var err error
	if req.RoomID != "" {
		room, err = h.Rooms.JoinRoom(req.RoomID, userID, username)
	} else {
		room, err = h.Rooms.Join(userID, username, req.Mode)
	}
	if err != nil {
		writeRoomError(c, err)
		return
	}
	h.broadcastRoom(room)
	c.JSON(http.StatusOK, room)
}

func (h *YuzhoushaHandler) GetRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Get(c.Param("roomId"), userID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *YuzhoushaHandler) LeaveRoom(c *gin.Context) {
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

func (h *YuzhoushaHandler) SetHeroRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	var req yzsSetHeroRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.CharacterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择武将"})
		return
	}
	room, err := h.Rooms.SetHero(c.Param("roomId"), userID, req.CharacterID)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	h.broadcastRoom(room)
	c.JSON(http.StatusOK, room)
}

func (h *YuzhoushaHandler) ReadyRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	var req yzsReadyRequest
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

func (h *YuzhoushaHandler) StartRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	room, err := h.Rooms.Start(c.Param("roomId"), userID, h.Games)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	h.broadcastRoom(room)
	c.JSON(http.StatusOK, room)
}

func (h *YuzhoushaHandler) ReadyNextRoom(c *gin.Context) {
	userID, _ := currentUser(c)
	var req yzsReadyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Ready = true
	}
	room, _, err := h.Rooms.SetReadyForNext(c.Param("roomId"), userID, req.Ready)
	if err != nil {
		writeRoomError(c, err)
		return
	}
	h.broadcastRoom(room)
	c.JSON(http.StatusOK, room)
}

func (h *YuzhoushaHandler) UseSkill(c *gin.Context) {
	var req yzsSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	// 允许空 skill_id（用于过河拆桥/顺手牵羊等 TakeWindow 选牌）
	if req.SkillID == "" && req.TargetZone == "" {
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
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) GetState(c *gin.Context) {
	userID, _ := currentUser(c)
	gameID := c.Param("gameId")
	state, err := h.Games.GetState(gameID, userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	if len(state.Events) > 0 {
		h.broadcastGame(gameID, state.Events, userID)
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
	secondSeat := -1
	if req.SecondTargetIndex != nil {
		secondSeat = *req.SecondTargetIndex
	}
	// 丈八蛇矛：两张手牌当杀
	if req.ZhangbaSecondCardID != "" {
		state, err := h.Games.PlayZhangbaSha(c.Param("gameId"), userID, req.CardID, req.ZhangbaSecondCardID, req.TargetIndex)
		if err != nil {
			writeYuzhoushaError(c, err)
			return
		}
		h.writeGameResponse(c, c.Param("gameId"), userID, state)
		return
	}
	state, err := h.Games.PlayCard(c.Param("gameId"), userID, req.CardID, engine.PlayTarget{
		SeatIndex:            req.TargetIndex,
		SecondSeatIndex:      secondSeat,
		Zone:                 req.TargetZone,
		CardID:               req.TargetCardID,
		FangtianExtraTargets: req.FangtianExtraTargets,
	})
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) RespondShan(c *gin.Context) {
	h.RespondCard(c)
}

func (h *YuzhoushaHandler) RespondZhangba(c *gin.Context) {
	var req yzsZhangbaRespondRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.CardIDs) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.RespondZhangbaSha(c.Param("gameId"), userID, req.CardIDs[0], req.CardIDs[1])
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
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
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) PassResponse(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.PassResponse(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) PassAllWuxiek(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.PassAllWuxiek(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) BaguaJudge(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.TryBaguaJudge(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) EndPlay(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.EndPlay(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
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
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) RespondWeaponDiscard(c *gin.Context) {
	var req yzsDiscardRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.CardIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	userID, _ := currentUser(c)
	state, err := h.Games.RespondDiscardCards(c.Param("gameId"), userID, req.CardIDs)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
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
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) PassDraw(c *gin.Context) {
	userID, _ := currentUser(c)
	state, err := h.Games.PassDraw(c.Param("gameId"), userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
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
	h.writeGameResponse(c, c.Param("gameId"), userID, state)
}

func (h *YuzhoushaHandler) FinishGuanxing(c *gin.Context) {
	h.FinishPeekDeck(c)
}

func (h *YuzhoushaHandler) Tick(c *gin.Context) {
	userID, _ := currentUser(c)
	gameID := c.Param("gameId")
	state, err := h.Games.Tick(gameID, userID)
	if err != nil {
		writeYuzhoushaError(c, err)
		return
	}
	if len(state.Events) > 0 {
		h.broadcastGame(gameID, state.Events, userID)
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
