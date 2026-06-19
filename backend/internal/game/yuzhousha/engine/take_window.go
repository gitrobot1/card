package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

type ZoneID string

const (
	ZoneHand       ZoneID = "hand"
	ZoneWeapon     ZoneID = "weapon"
	ZoneArmor      ZoneID = "armor"
	ZonePlusHorse  ZoneID = "plus_horse"
	ZoneMinusHorse ZoneID = "minus_horse"
	ZoneJudge      ZoneID = "judge"
	ZoneCamp       ZoneID = "camp"
	ZoneDiscard    ZoneID = "discard"
	ZoneVoid       ZoneID = "void"
)

type TakeDestination struct {
	Zone ZoneID
	Seat int
}

type TakeWindowConfig struct {
	SkillID      string
	ResponseMode string
	ActorSeat    int
	SubjectSeat  int
	OriginSeat   int

	MaxTake      int
	MinTake      int
	AllowedZones []ZoneID

	Destination TakeDestination

	OnEachTake func(g *Game, card Card, label string, events *[]GameEvent) error
	OnComplete func(g *Game, events *[]GameEvent) error

	Message          string
	EventType        string
	SkillEventLabel  string
	PassClosesWindow bool
	// PickTarget 自定义 AI/自动选牌；nil 时用 aiPickTakeTarget。
	PickTarget func(g *Game, actor, subject int) (zone, cardID string, ok bool)
}

type takeWindowState struct {
	cfg   TakeWindowConfig
	taken int
}

func (tw *takeWindowState) remaining() int {
	if tw.cfg.MaxTake <= 0 {
		return 0
	}
	return tw.cfg.MaxTake - tw.taken
}

func zoneIDToPlayZone(z ZoneID) string {
	return string(z)
}

func (g *Game) OpenTakeWindow(cfg TakeWindowConfig, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if err := g.normalizeTakeWindowConfig(&cfg); err != nil {
		return err
	}
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:  cfg.SubjectSeat,
		TargetIndex:  cfg.ActorSeat,
		ReturnIndex:  cfg.ActorSeat,
		ResponseMode: cfg.ResponseMode,
		SkillID:      cfg.SkillID,
		WindowKind:   WindowKindTake,
		ActorSeat:    cfg.ActorSeat,
		SubjectSeat:  cfg.SubjectSeat,
		OriginSeat:   cfg.OriginSeat,
	}
	g.applyTakeWindowLegacyCounters(cfg)
	return g.attachTakeWindow(cfg, events)
}

// OpenTakeWindowOnPending 在已有 Pending（如杀链）上挂载 TakeWindow，保留 Card/RequiredKind 等字段。
func (g *Game) OpenTakeWindowOnPending(cfg TakeWindowConfig, events *[]GameEvent) error {
	if g.Pending == nil {
		return ErrWrongPhase
	}
	if err := g.normalizeTakeWindowConfig(&cfg); err != nil {
		return err
	}
	p := g.Pending
	p.ResponseMode = cfg.ResponseMode
	p.SkillID = cfg.SkillID
	p.WindowKind = WindowKindTake
	p.ActorSeat = cfg.ActorSeat
	p.SubjectSeat = cfg.SubjectSeat
	p.OriginSeat = cfg.OriginSeat
	FillPendingRoles(p)
	return g.attachTakeWindow(cfg, events)
}

func (g *Game) normalizeTakeWindowConfig(cfg *TakeWindowConfig) error {
	if cfg.MaxTake <= 0 {
		return ErrInvalidTarget
	}
	if cfg.ActorSeat < 0 || cfg.SubjectSeat < 0 {
		return ErrInvalidTarget
	}
	if cfg.EventType == "" {
		cfg.EventType = "take_window"
	}
	if cfg.Destination.Zone == "" {
		cfg.Destination.Zone = ZoneHand
	}
	if cfg.Destination.Seat < 0 {
		cfg.Destination.Seat = cfg.ActorSeat
	}
	return nil
}

func (g *Game) applyTakeWindowLegacyCounters(cfg TakeWindowConfig) {
	switch cfg.ResponseMode {
	case ResponseModeSkillFankui:
		g.Pending.FankuiRemaining = cfg.MaxTake
	case ResponseModeSkillTuxi:
		g.Pending.TuxiRemaining = cfg.MaxTake
	}
	FillPendingRoles(g.Pending)
}

func (g *Game) attachTakeWindow(cfg TakeWindowConfig, events *[]GameEvent) error {
	g.takeWindow = &takeWindowState{cfg: cfg}
	if cfg.Message != "" {
		g.Message = cfg.Message
	}
	if cfg.SkillID != "" && events != nil && cfg.Message != "" {
		g.appendSkillEvent(events, cfg.SkillID, cfg.ActorSeat, cfg.SubjectSeat, g.Message)
	}
	g.resetTimer()
	return nil
}

func (g *Game) syncLegacyTakeRemaining() {
	if g.takeWindow == nil || g.Pending == nil {
		return
	}
	switch g.Pending.ResponseMode {
	case ResponseModeSkillFankui:
		g.Pending.FankuiRemaining = g.takeWindow.remaining()
	case ResponseModeSkillTuxi:
		g.Pending.TuxiRemaining = g.takeWindow.remaining()
	}
}

func (g *Game) validateTakeActor(actor int) error {
	if g.Phase != PhaseResponse || g.Pending == nil || g.takeWindow == nil {
		return ErrWrongPhase
	}
	if g.Pending.WindowKind != WindowKindTake && g.Pending.ResponseMode != ResponseModeSkillFankui &&
		g.Pending.ResponseMode != ResponseModeSkillTuxi &&
		g.Pending.ResponseMode != ResponseModeSkillPojun {
		return ErrWrongPhase
	}
	if !g.IsActorSeat(actor) {
		return ErrNotYourTurn
	}
	if g.takeWindow.remaining() <= 0 {
		return ErrWrongPhase
	}
	return nil
}

func (g *Game) HasTakeableInWindow(subject int, zones []ZoneID) bool {
	if len(zones) == 0 {
		return g.hasTakeableCard(subject)
	}
	p := &g.Players[subject]
	for _, z := range zones {
		switch z {
		case ZoneHand:
			if len(p.Hand) > 0 {
				return true
			}
		case ZoneWeapon:
			if p.Weapon != nil {
				return true
			}
		case ZoneArmor:
			if p.Armor != nil {
				return true
			}
		case ZonePlusHorse:
			if p.PlusHorse != nil {
				return true
			}
		case ZoneMinusHorse:
			if p.MinusHorse != nil {
				return true
			}
		case ZoneJudge:
			if len(p.JudgeArea) > 0 {
				return true
			}
		}
	}
	return false
}

func (g *Game) zoneAllowed(zone string, allowed []ZoneID) bool {
	if len(allowed) == 0 {
		return true
	}
	if zone == "" {
		zone = "hand"
	}
	for _, z := range allowed {
		if string(z) == zone {
			return true
		}
	}
	return false
}

func (g *Game) TakeOne(actor int, zone ZoneID, cardID string, events *[]GameEvent) error {
	if err := g.validateTakeActor(actor); err != nil {
		return err
	}
	tw := g.takeWindow
	subject := g.Pending.SubjectSeat
	playZone := zoneIDToPlayZone(zone)
	if playZone == "" {
		playZone = "hand"
	}
	if !g.zoneAllowed(playZone, tw.cfg.AllowedZones) {
		return ErrInvalidTarget
	}
	spec := PlayTarget{SeatIndex: subject, Zone: playZone, CardID: cardID}
	card, label, ok := g.takeTargetCard(subject, spec, events)
	if !ok {
		if tw.cfg.PassClosesWindow {
			return g.finishTakeWindow(events)
		}
		return ErrInvalidTarget
	}
	if err := g.placeTakenCard(tw.cfg.Destination, card, events); err != nil {
		return err
	}
	actorSeat := tw.cfg.ActorSeat
	subjectSeat := tw.cfg.SubjectSeat
	var msg string
	if tw.cfg.SkillEventLabel != "" {
		msg = fmt.Sprintf("%s 发动【%s】，获得 %s 的%s",
			g.Players[actorSeat].Name, tw.cfg.SkillEventLabel, g.Players[subjectSeat].Name, label)
	} else {
		msg = fmt.Sprintf("%s 获得 %s 的%s", g.Players[actorSeat].Name, g.Players[subjectSeat].Name, label)
	}
	if tw.cfg.OnEachTake != nil {
		if err := tw.cfg.OnEachTake(g, card, label, events); err != nil {
			return err
		}
	} else if events != nil {
		if tw.cfg.SkillID != "" {
			g.appendSkillEvent(events, tw.cfg.SkillID, actorSeat, subjectSeat, msg)
		}
		*events = append(*events, GameEvent{
			Type:        tw.cfg.EventType,
			PlayerIndex: actorSeat,
			TargetIndex: subjectSeat,
			Card:        &card,
			Message:     msg,
		})
	}
	tw.taken++
	g.syncLegacyTakeRemaining()
	if tw.remaining() > 0 && g.HasTakeableInWindow(subject, tw.cfg.AllowedZones) {
		switch g.Pending.ResponseMode {
		case ResponseModeSkillPojun:
			g.Message = fmt.Sprintf("%s 继续发动【破军】", g.Players[actorSeat].Name)
		case ResponseModeSkillFankui, ResponseModeSkillTuxi:
			if tw.cfg.SkillEventLabel != "" {
				g.Message = fmt.Sprintf("%s 可继续发动【%s】", g.Players[actorSeat].Name, tw.cfg.SkillEventLabel)
			} else {
				g.Message = fmt.Sprintf("%s 继续选择目标牌", g.Players[actorSeat].Name)
			}
		default:
			if tw.cfg.SkillEventLabel != "" {
				g.Message = fmt.Sprintf("%s 可继续发动【%s】", g.Players[actorSeat].Name, tw.cfg.SkillEventLabel)
			} else {
				g.Message = fmt.Sprintf("%s 继续选择目标牌", g.Players[actorSeat].Name)
			}
		}
		g.resetTimer()
		return nil
	}
	return g.finishTakeWindow(events)
}

func (g *Game) PassTake(actor int, events *[]GameEvent) error {
	if err := g.validateTakeActor(actor); err != nil {
		return err
	}
	tw := g.takeWindow
	if tw.cfg.PassClosesWindow {
		return g.finishTakeWindow(events)
	}
	tw.taken++
	g.syncLegacyTakeRemaining()
	subject := g.Pending.SubjectSeat
	if tw.remaining() > 0 && g.HasTakeableInWindow(subject, tw.cfg.AllowedZones) {
		actorSeat := tw.cfg.ActorSeat
		if tw.cfg.SkillEventLabel != "" {
			g.Message = fmt.Sprintf("%s 跳过【%s】，仍可再发动", g.Players[actorSeat].Name, tw.cfg.SkillEventLabel)
		} else {
			g.Message = fmt.Sprintf("%s 跳过拿牌，仍可继续", g.Players[actorSeat].Name)
		}
		g.resetTimer()
		return nil
	}
	return g.finishTakeWindow(events)
}

func (g *Game) finishTakeWindow(events *[]GameEvent) error {
	tw := g.takeWindow
	onComplete := tw.cfg.OnComplete
	g.takeWindow = nil
	if onComplete != nil {
		return onComplete(g, events)
	}
	g.Pending = nil
	g.Phase = PhasePlaying
	return nil
}

// abandonTakeWindow 放弃剩余拿牌（不触发 OnComplete），用于杀链等非 actor 先响应时关闭重叠窗。
func (g *Game) abandonTakeWindow() {
	g.takeWindow = nil
}

func (g *Game) AutoTakeWindow(actor int, events *[]GameEvent) {
	if g.takeWindow == nil || !g.IsActorSeat(actor) {
		return
	}
	subject := g.Pending.SubjectSeat
	for g.takeWindow != nil && g.takeWindow.remaining() > 0 && g.HasTakeableInWindow(subject, g.takeWindow.cfg.AllowedZones) {
		var zone, cardID string
		if g.takeWindow.cfg.PickTarget != nil {
			var ok bool
			zone, cardID, ok = g.takeWindow.cfg.PickTarget(g, actor, subject)
			if !ok {
				_ = g.PassTake(actor, events)
				return
			}
		} else {
			zone, cardID = aiPickTakeTarget(g, subject)
		}
		if zone == "" {
			_ = g.PassTake(actor, events)
			return
		}
		if err := g.TakeOne(actor, ZoneID(zone), cardID, events); err != nil {
			_ = g.PassTake(actor, events)
			return
		}
	}
}

func (g *Game) autoTakeWindowIfNeeded(events *[]GameEvent) bool {
	if g.takeWindow == nil || g.Pending == nil {
		return false
	}
	seat := g.PendingActorSeat()
	if seat < 0 || !g.Players[seat].IsAI {
		return false
	}
	subject := g.Pending.SubjectSeat
	if !g.HasTakeableInWindow(subject, g.takeWindow.cfg.AllowedZones) {
		_ = g.PassTake(seat, events)
		return true
	}
	if g.Pending.SkillID != "" {
		rt := g.skillRuntime(events)
		if h, ok := skill.Lookup(g.Pending.SkillID); ok && h.CanActivate(rt, seat) {
			if err := h.AIActivate(rt, seat); err != nil {
				_ = g.PassTake(seat, events)
			}
			return true
		}
	}
	g.AutoTakeWindow(seat, events)
	return true
}

func (g *Game) PendingWindowKind() string {
	if g.Pending == nil {
		return ""
	}
	g.ensurePendingRoles()
	return g.Pending.WindowKind
}

func (g *Game) PendingOriginSeat() int {
	if g.Pending == nil {
		return -1
	}
	g.ensurePendingRoles()
	return g.Pending.OriginSeat
}
