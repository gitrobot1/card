package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

var (
	ErrNotInGame           = errors.New("not in game")
	ErrWrongPhase          = errors.New("wrong phase")
	ErrNotYourTurn         = errors.New("not your turn")
	ErrInvalidCard         = errors.New("invalid card")
	ErrInvalidDiscardCount = errors.New("invalid discard count")
	ErrInvalidTarget       = errors.New("invalid target")
	ErrAlreadyActed        = errors.New("already acted this turn")
	ErrGameOver            = errors.New("game over")
	ErrPendingCombat       = errors.New("pending combat response")
	ErrNoPendingCombat     = errors.New("no pending combat")
)

type Game struct {
	ID               string
	Phase            string
	TurnStep         string
	CurrentTurn      int
	HumanPlayer      int
	Players          []Player
	Pending          *PendingCombat
	damageAftermath  *DamageAftermath
	dyingContext     *DyingContext
	leijiSavedPending *PendingCombat
	leijiShanSeat    int
	pendingDamageResume *DamageResume
	Message          string
	WinnerIndex      *int
	DrawPile         []Card
	DiscardPile      []Card
	TurnDeadline     time.Time
	TurnDeadlineUnix int64
	Mode             string
	LandlordSeat     int
	LordSeat         int
	Identities       []string
	RoleRevealed     []bool
	WinnerTeam       *int
	gameOverStats    *GameOverStats // 游戏结束统计（预留）
	testRand         *rand.Rand // sim/tests: fixed shuffle source
	takeWindow       *takeWindowState
	discardWindow    *discardWindowState
	wuguPicked       map[int]bool // 五谷丰登：已选过牌的玩家
}

type PublicState struct {
	ID               string         `json:"id"`
	Phase            string         `json:"phase"`
	TurnStep         string         `json:"turn_step"`
	CurrentTurn      int            `json:"current_turn"`
	HumanPlayer      int            `json:"human_player"`
	Players          []PlayerPublic `json:"players"`
	Pending          *PendingCombat `json:"pending,omitempty"`
	Message          string         `json:"message"`
	WinnerIndex      *int           `json:"winner_index,omitempty"`
	WinnerTeam       *int           `json:"winner_team,omitempty"`
	Mode             string         `json:"mode,omitempty"`
	LandlordSeat     int            `json:"landlord_seat,omitempty"`
	LordSeat         int            `json:"lord_seat,omitempty"`
	LayoutKey        string         `json:"layout_key,omitempty"`
	SeatMap          []mode.SeatSlot `json:"seat_map,omitempty"`
	DrawCount        int            `json:"draw_count"`
	DiscardCount     int            `json:"discard_count"`
	MyHand           []Card         `json:"my_hand,omitempty"`
	TurnDeadlineUnix int64          `json:"turn_deadline_unix"`
	Events           []GameEvent    `json:"events"`
	ActivatableSkills []SkillMeta   `json:"activatable_skills,omitempty"`
	GameOverStats    *GameOverStats `json:"game_over_stats,omitempty"` // 游戏结束时填充
}

func (g *Game) HasAI() bool {
	for _, p := range g.Players {
		if p.IsAI {
			return true
		}
	}
	return false
}

func (g *Game) IsFinished() bool {
	return g.Phase == PhaseFinished
}

func (g *Game) setupDeck() {
	profile := mode.DeckProfileFor(g.Mode)
	deck := g.shuffleCards(NewDeckForMode(g.Mode))
	handSize := profile.InitialHandSize
	if handSize <= 0 {
		handSize = InitialHandSize
	}
	for i := range g.Players {
		g.Players[i].Hand = append([]Card(nil), deck[:handSize]...)
		deck = deck[handSize:]
	}
	g.DrawPile = deck
	g.DiscardPile = nil
	g.syncCounts()
}

func (g *Game) syncCounts() {
	for i := range g.Players {
		g.Players[i].HandCount = len(g.Players[i].Hand)
	}
}

func cloneCardPtr(card *Card) *Card {
	if card == nil {
		return nil
	}
	cp := *card
	return &cp
}

func (g *Game) resetTimer() {
	g.TurnDeadline = time.Now().Add(TurnTimeoutSec * time.Second)
	g.TurnDeadlineUnix = g.TurnDeadline.Unix()
}

func (g *Game) attackRange(seat int) int {
	if seat < 0 || seat >= len(g.Players) || g.Players[seat].Weapon == nil {
		return 1
	}
	return weaponRange(g.Players[seat].Weapon.Kind)
}

func (g *Game) distanceBetween(from, to int) int {
	dist := g.seatDistance(from, to)
	if from >= 0 && from < len(g.Players) && g.Players[from].MinusHorse != nil {
		dist--
	}
	if to >= 0 && to < len(g.Players) && g.Players[to].PlusHorse != nil {
		dist++
	}
	if dist < 1 {
		dist = 1
	}
	dist += g.skillDistanceDelta(from, to)
	if dist < 1 {
		return 1
	}
	return dist
}

func (g *Game) canAttack(from, to int) bool {
	// 立牧：判定区有牌时，将距离视为1（最小合法距离）
	// 忽略座次距离、马匹修正、技能距离修正，只受武器攻击范围限制
	if g.hasSkill(from, SkillLimu) && len(g.Players[from].JudgeArea) > 0 {
		// 严格判断：距离固定为1，不受任何修正影响
		return 1 <= g.attackRange(from)
	}
	return g.distanceBetween(from, to) <= g.attackRange(from)
}

func weaponRange(kind string) int {
	switch kind {
	case CardWeapon1:
		return 1
	case CardWeapon2:
		return 2
	case CardWeapon3:
		return 3
	case CardWeapon4:
		return 4
	case CardWeapon5:
		return 5
	case CardWeapon6:
		return 2
	case CardWeapon7:
		return 4
	case CardWeapon8:
		return 2
	case CardWeapon9:
		return 3
	default:
		return 1
	}
}

func (g *Game) findCard(seat int, cardID string) (int, Card, bool) {
	for i, c := range g.Players[seat].Hand {
		if c.ID == cardID {
			return i, c, true
		}
	}
	return -1, Card{}, false
}

// findCardInHandOrEquip 在手牌和装备区中查找牌
func (g *Game) findCardInHandOrEquip(seat int, cardID string) (zone string, idx int, card Card, ok bool) {
	// 先查找手牌
	if idx, card, ok := g.findCard(seat, cardID); ok {
		return string(ZoneHand), idx, card, true
	}
	// 再查找装备区
	for _, equip := range []struct {
		zone string
		card *Card
	}{
		{string(EquipWeapon), g.Players[seat].Weapon},
		{string(EquipArmor), g.Players[seat].Armor},
		{string(EquipPlusHorse), g.Players[seat].PlusHorse},
		{string(EquipMinusHorse), g.Players[seat].MinusHorse},
	} {
		if equip.card != nil && equip.card.ID == cardID {
			return equip.zone, -1, *equip.card, true
		}
	}
	return "", -1, Card{}, false
}

// findCardZone 查找牌所在的 zone（用于破军批量选牌）
func (g *Game) findCardZone(seat int, cardID string) ZoneID {
	zone, _, _, ok := g.findCardInHandOrEquip(seat, cardID)
	if !ok {
		// 查找判定区
		for _, jc := range g.Players[seat].JudgeArea {
			if jc.ID == cardID {
				return ZoneJudge
			}
		}
		return ZoneHand // fallback
	}
	return ZoneID(zone)
}

func (g *Game) removeHandCard(seat, idx int, events *[]GameEvent) Card {
	p := &g.Players[seat]
	c := p.Hand[idx]
	p.Hand = append(p.Hand[:idx], p.Hand[idx+1:]...)
	g.syncCounts()
	g.runHandEmptyHooks(seat, events)
	return c
}
func (g *Game) finishGame(winner int, events *[]GameEvent) {
	g.Phase = PhaseFinished
	g.TurnStep = ""
	g.Pending = nil
	g.WinnerIndex = &winner
	g.Message = fmt.Sprintf("%s 获胜", g.Players[winner].Name)
	*events = append(*events, GameEvent{
		Type:        "game_over",
		PlayerIndex: winner,
		Message:     g.Message,
	})

	// 初始化游戏结束统计（预留，目前只记录基础信息）
	g.buildGameOverStats(winner, -1, "damage")
}

// buildGameOverStats 构建游戏结束统计信息（预留接口，暂未填充详细数据）。
func (g *Game) buildGameOverStats(winnerIndex, winnerTeam int, reason string) {
	// 目前只在 PublicViewForSeat 中构建，这里预留 future use
}

// buildGameOverStatsForView 构建用于前端展示的游戏结束统计。
func (g *Game) buildGameOverStatsForView() *GameOverStats {
	if g.WinnerIndex == nil {
		return nil
	}

	// 如果已经构建过，直接返回
	if g.gameOverStats != nil {
		return g.gameOverStats
	}

	winner := *g.WinnerIndex
	stats := &GameOverStats{
		WinnerIndex: winner,
		Reason:      "damage", // TODO: 从实际游戏过程记录失败原因
		PlayerStats: make([]PlayerGameStats, len(g.Players)),
	}
	if g.WinnerTeam != nil {
		stats.WinnerTeam = *g.WinnerTeam
	}

	for i := range g.Players {
		stats.PlayerStats[i] = PlayerGameStats{
			Seat:        i,
			Name:         g.Players[i].Name,
			CharacterID:  g.Players[i].Character.ID,
			IsWinner:     i == winner,
			// TODO: 填充详细统计数据
			// - DamageDealt: 需要跟踪每次伤害的来源
			// - DamageTaken: 需要跟踪每次伤害的目标
			// - HealDone: 需要跟踪治疗来源
			// - KillCount: 需要跟踪击杀关系
			// - SurvivalRank: 需要记录阵亡顺序
		}
	}

	// 存储起来，避免重复构建
	g.gameOverStats = stats
	return stats
}

func (g *Game) IsHumanPending() bool {
	seat := g.PendingActorSeat()
	if seat < 0 || seat >= len(g.Players) {
		return false
	}
	return !g.Players[seat].IsAI
}

func (g *Game) IsPhaseExpired() bool {
	if g.TurnDeadline.IsZero() || g.IsFinished() {
		return false
	}
	return time.Now().After(g.TurnDeadline)
}

func (g *Game) ApplyHumanTimeout(events *[]GameEvent) error {
	if !g.IsHumanPending() || !g.IsPhaseExpired() {
		return nil
	}
	seat := g.PendingActorSeat()
	if seat < 0 || seat >= len(g.Players) {
		return nil
	}
	if g.Phase == PhaseResponse {
		if g.Pending != nil && g.Pending.ResponseMode == ResponseModePeekDeck {
			return g.autoFinishPeekDeck(seat, events)
		}
		if g.Pending != nil && g.Pending.TieqiPending && g.Pending.SourceIndex == seat {
			return g.SkipTieqi(seat, events)
		}
		return g.PassResponse(seat, events)
	}
	if g.Phase == PhasePlaying && g.TurnStep == StepPrepare {
		return g.PassPrepare(seat, events)
	}
	if g.Phase == PhasePlaying && g.TurnStep == StepDraw && g.isDrawPhaseChoicePending(seat) {
		return g.PassDrawPhase(seat, events)
	}
	if g.Phase == PhasePlaying && g.TurnStep == StepDiscard {
		g.autoDiscard(seat, events)
		return g.endTurn(events)
	}
	return g.EndPlay(seat, events)
}

func (g *Game) PublicViewForSeat(seat int, events []GameEvent) PublicState {
	players := make([]PlayerPublic, len(g.Players))
	for i, p := range g.Players {
		pub := PlayerPublic{
			Index:           p.Index,
			Name:            p.Name,
			IsAI:            p.IsAI,
			Character:       p.Character,
			HP:              p.HP,
			MaxHP:           p.MaxHP,
			HandCount:       len(p.Hand),
			Team:            g.teamOf(i),
			ShaUsedThisTurn:      p.ShaUsedThisTurn,
			ShaExtraUsedThisTurn: p.ShaExtraUsedThisTurn,
			SkipPlay:        p.SkipPlay,
			SkipDraw:        p.SkipDraw,
			Drunk:           p.Drunk,
			Weapon:          cloneCardPtr(p.Weapon),
			Armor:           cloneCardPtr(p.Armor),
			PlusHorse:       cloneCardPtr(p.PlusHorse),
			MinusHorse:      cloneCardPtr(p.MinusHorse),
			JudgeArea:       append([]Card(nil), p.JudgeArea...),
			CampCards:       append([]Card(nil), p.CampCards...),
		}
		if len(p.SkillCounters) > 0 {
			pub.SkillCounters = make(map[string]int, len(p.SkillCounters))
			for k, v := range p.SkillCounters {
				pub.SkillCounters[k] = v
			}
		}
		if i == seat || g.IsFinished() {
			pub.Hand = append([]Card(nil), p.Hand...)
		}
		// 破军 TakeWindow 期间，将目标手牌暴露给 Actor 用于选牌
		if g.Pending != nil && g.Pending.WindowKind == WindowKindTake &&
			g.Pending.ResponseMode == ResponseModeSkillPojun &&
			g.Pending.ActorSeat == seat && i == g.Pending.SubjectSeat {
			pub.Hand = append([]Card(nil), p.Hand...)
		}
		if g.isIdentity() && i < len(g.Identities) {
			revealed := g.IdentityRevealed(i)
			pub.IdentityRevealed = revealed
			if i == seat || revealed || g.Identities[i] == mode.RoleLord {
				pub.Identity = g.Identities[i]
			}
		}
		players[i] = pub
	}
	var myHand []Card
	if seat >= 0 && seat < len(g.Players) {
		myHand = append([]Card(nil), g.Players[seat].Hand...)
	}
	if events == nil {
		events = []GameEvent{}
	}
	var pending *PendingCombat
	if g.Pending != nil {
		pc := *g.Pending
		FillPendingRoles(&pc)
		if pc.ResponseMode == ResponseModeSkillFanjianSuit && seat == pc.TargetIndex {
			pc.Card = Card{}
		}
		pending = &pc
	}
	var layoutKey string
	var seatMap []mode.SeatSlot
	if meta, ok := mode.Lookup(g.Mode); ok {
		layoutKey = meta.LayoutKey
		if len(meta.SeatMap) > 0 {
			seatMap = append([]mode.SeatSlot(nil), meta.SeatMap...)
		}
	}
	// 构建游戏结束统计（仅在游戏结束时）
	var gameOverStats *GameOverStats
	if g.IsFinished() && g.WinnerIndex != nil {
		gameOverStats = g.buildGameOverStatsForView()
	}

	return PublicState{
		ID:               g.ID,
		Phase:            g.Phase,
		TurnStep:         g.TurnStep,
		CurrentTurn:      g.CurrentTurn,
		HumanPlayer:      seat,
		Mode:             g.Mode,
		LandlordSeat:     g.LandlordSeat,
		LordSeat:         g.LordSeat,
		LayoutKey:        layoutKey,
		SeatMap:          seatMap,
		Players:          players,
		Pending:          pending,
		Message:          g.Message,
		WinnerIndex:      g.WinnerIndex,
		WinnerTeam:       g.WinnerTeam,
		DrawCount:        len(g.DrawPile),
		DiscardCount:     len(g.DiscardPile),
		MyHand:           myHand,
		TurnDeadlineUnix: g.TurnDeadlineUnix,
		Events:           events,
		ActivatableSkills: g.ListActivatableSkills(seat),
		GameOverStats:    gameOverStats,
	}
}
