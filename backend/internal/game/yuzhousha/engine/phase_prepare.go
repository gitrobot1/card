package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

func (g *Game) alivePlayerCount() int {
	n := 0
	for i := range g.Players {
		if g.Players[i].HP > 0 {
			n++
		}
	}
	if n < 1 {
		return 1
	}
	return n
}

func (g *Game) shouldEnterPreparePhase(seat int) bool {
	rt := g.skillRuntime(nil)
	for _, h := range g.playerSkillHandlers(seat) {
		if h.OffersPreparePhase(rt, seat) {
			return true
		}
	}
	return false
}

func (g *Game) enterPreparePhase(seat int, events *[]GameEvent) bool {
	if !g.shouldEnterPreparePhase(seat) {
		return false
	}
	g.TurnStep = StepPrepare
	g.Pending = nil
	g.Message = fmt.Sprintf("%s 准备阶段", g.Players[seat].Name)
	g.resetTimer()
	*events = append(*events, GameEvent{
		Type:        "prepare_phase",
		PlayerIndex: seat,
		Message:     g.Message,
	})
	return true
}

func (g *Game) PassPrepare(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	return g.continueAfterPrepare(seat, events)
}

func (g *Game) peekCountForSkill(seat int, skillID string) int {
	h, ok := skill.Lookup(skillID)
	if !ok {
		return 0
	}
	return skill.PeekCountFor(g.skillRuntime(nil), seat, h)
}

func (g *Game) StartPeekDeck(seat int, skillID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	h, ok := skill.Lookup(skillID)
	if !ok || h.PeekDeckConfig() == nil || !g.hasSkill(seat, skillID) {
		return ErrInvalidCard
	}
	count := skill.PeekCountFor(g.skillRuntime(nil), seat, h)
	if count == 0 {
		return ErrWrongPhase
	}
	revealed := make([]Card, 0, count)
	for i := 0; i < count; i++ {
		c := g.DrawPile[0]
		g.DrawPile = g.DrawPile[1:]
		revealed = append(revealed, c)
	}
	g.syncCounts()

	meta := h.Meta()
	g.Phase = PhaseResponse
	g.Pending = &PendingCombat{
		SourceIndex:   seat,
		TargetIndex:   seat,
		ReturnIndex:   seat,
		ResponseMode:  ResponseModePeekDeck,
		RevealedCards: revealed,
		SkillID:       skillID,
	}
	g.Message = fmt.Sprintf("%s 发动【%s】，请分配 %d 张牌至牌堆顶/底", g.Players[seat].Name, meta.Name, len(revealed))
	g.resetTimer()
	g.appendSkillEvent(events, skillID, seat, seat, g.Message)
	*events = append(*events, GameEvent{
		Type:        "peek_deck_reveal",
		PlayerIndex: seat,
		Amount:      len(revealed),
		SkillID:     skillID,
		Message:     g.Message,
	})
	for i := range revealed {
		c := revealed[i]
		*events = append(*events, GameEvent{
			Type:        "peek_deck_show",
			PlayerIndex: seat,
			Card:        &c,
			SkillID:     skillID,
			Message:     fmt.Sprintf("【%s】 %s", meta.Name, c.Label),
		})
	}
	return nil
}

type PeekDeckRequest struct {
	TopCardIDs    []string
	BottomCardIDs []string
}

// GuanxingRequest 兼容旧调用方。
type GuanxingRequest = PeekDeckRequest

func (g *Game) FinishPeekDeck(seat int, req PeekDeckRequest, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModePeekDeck {
		return ErrWrongPhase
	}
	if g.Pending.TargetIndex != seat {
		return ErrNotYourTurn
	}
	revealed := g.Pending.RevealedCards
	if err := validatePeekDeckPartition(revealed, req.TopCardIDs, req.BottomCardIDs); err != nil {
		return err
	}
	topCards := orderCardsByIDs(revealed, req.TopCardIDs)
	bottomCards := orderCardsByIDs(revealed, req.BottomCardIDs)

	g.DrawPile = append(topCards, g.DrawPile...)
	g.DrawPile = append(g.DrawPile, bottomCards...)
	g.syncCounts()

	skillID := g.Pending.SkillID
	skillName := skillID
	if h, ok := skill.Lookup(skillID); ok && h.Meta().Name != "" {
		skillName = h.Meta().Name
	}

	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPrepare

	msg := fmt.Sprintf("%s 完成【%s】", g.Players[seat].Name, skillName)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "peek_deck_finish",
		PlayerIndex: seat,
		Amount:      len(topCards),
		SkillID:     skillID,
		Message:     msg,
	})
	return g.continueAfterPrepare(seat, events)
}

func (g *Game) FinishGuanxing(seat int, req GuanxingRequest, events *[]GameEvent) error {
	return g.FinishPeekDeck(seat, req, events)
}

func (g *Game) StartGuanxing(seat int, events *[]GameEvent) error {
	return g.StartPeekDeck(seat, skill.IDGuanxing, events)
}

func validatePeekDeckPartition(revealed []Card, topIDs, bottomIDs []string) error {
	if len(topIDs)+len(bottomIDs) != len(revealed) {
		return ErrInvalidCard
	}
	seen := make(map[string]struct{}, len(revealed))
	for _, id := range append(append([]string{}, topIDs...), bottomIDs...) {
		if id == "" {
			return ErrInvalidCard
		}
		if _, dup := seen[id]; dup {
			return ErrInvalidCard
		}
		found := false
		for _, c := range revealed {
			if c.ID == id {
				found = true
				break
			}
		}
		if !found {
			return ErrInvalidCard
		}
		seen[id] = struct{}{}
	}
	return nil
}

func orderCardsByIDs(revealed []Card, ids []string) []Card {
	out := make([]Card, 0, len(ids))
	for _, id := range ids {
		for _, c := range revealed {
			if c.ID == id {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

func (g *Game) continueAfterPrepare(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	g.TurnStep = StepDraw
	g.Pending = nil
	if g.processLightningAtTurnStart(seat, events) {
		return nil
	}
	return g.resumeBeginTurnAfterLightning(seat, events)
}

func (g *Game) autoFinishPeekDeck(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModePeekDeck {
		return ErrWrongPhase
	}
	revealed := g.Pending.RevealedCards
	salt := seat*31 + g.CurrentTurn*17 + len(revealed)
	top, bottom := randomPeekPartition(revealed, salt)
	if err := validatePeekDeckPartition(revealed, top, bottom); err == nil {
		return g.FinishPeekDeck(seat, PeekDeckRequest{TopCardIDs: top, BottomCardIDs: bottom}, events)
	}
	topCards, bottomCards := splitPeekCardsByIndex(revealed, salt)
	return g.applyPeekDeckSplit(seat, topCards, bottomCards, events)
}

func splitPeekCardsByIndex(revealed []Card, salt int) (top, bottom []Card) {
	if len(revealed) == 0 {
		return nil, nil
	}
	for i, c := range revealed {
		if (i+salt)%2 == 0 {
			top = append(top, c)
		} else {
			bottom = append(bottom, c)
		}
	}
	if len(top) == 0 && len(bottom) > 0 {
		top = append(top, bottom[0])
		bottom = bottom[1:]
	}
	if len(bottom) == 0 && len(top) > 1 {
		bottom = append(bottom, top[len(top)-1])
		top = top[:len(top)-1]
	}
	return top, bottom
}

func (g *Game) applyPeekDeckSplit(seat int, topCards, bottomCards []Card, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhaseResponse || g.Pending == nil || g.Pending.ResponseMode != ResponseModePeekDeck {
		return ErrWrongPhase
	}
	if g.Pending.TargetIndex != seat {
		return ErrNotYourTurn
	}
	g.DrawPile = append(topCards, g.DrawPile...)
	g.DrawPile = append(g.DrawPile, bottomCards...)
	g.syncCounts()

	skillID := g.Pending.SkillID
	skillName := skillID
	if h, ok := skill.Lookup(skillID); ok && h.Meta().Name != "" {
		skillName = h.Meta().Name
	}

	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPrepare

	msg := fmt.Sprintf("%s 完成【%s】", g.Players[seat].Name, skillName)
	g.Message = msg
	*events = append(*events, GameEvent{
		Type:        "peek_deck_finish",
		PlayerIndex: seat,
		Amount:      len(topCards),
		SkillID:     skillID,
		Message:     msg,
	})
	return g.continueAfterPrepare(seat, events)
}

func (g *Game) aiPartitionPeekDeck(seat int, revealed []Card) (topIDs, bottomIDs []string) {
	if g.Pending == nil {
		return nil, nil
	}
	skillID := g.Pending.SkillID
	h, ok := skill.Lookup(skillID)
	if !ok {
		return defaultAIPeekAllTop(revealed)
	}
	cfg := h.PeekDeckConfig()
	if cfg == nil || cfg.AIPartition == nil {
		return defaultAIPeekAllTop(revealed)
	}
	views := make([]skill.PeekCardView, len(revealed))
	for i, c := range revealed {
		views[i] = skill.PeekCardView{ID: c.ID, Kind: c.Kind}
	}
	return cfg.AIPartition(g.skillRuntime(nil), seat, views)
}

func (g *Game) finishPeekDeckAsAI(seat int, events *[]GameEvent) error {
	if g.Pending == nil || g.Pending.ResponseMode != ResponseModePeekDeck {
		return ErrWrongPhase
	}
	revealed := g.Pending.RevealedCards
	top, bottom := g.aiPartitionPeekDeck(seat, revealed)
	salt := seat*31 + g.CurrentTurn*17 + len(revealed)
	if len(top)+len(bottom) != len(revealed) {
		top, bottom = randomPeekPartition(revealed, salt)
	}
	if err := validatePeekDeckPartition(revealed, top, bottom); err == nil {
		return g.FinishPeekDeck(seat, PeekDeckRequest{TopCardIDs: top, BottomCardIDs: bottom}, events)
	}
	topCards, bottomCards := splitPeekCardsByIndex(revealed, salt)
	return g.applyPeekDeckSplit(seat, topCards, bottomCards, events)
}

func defaultAIPeekAllTop(revealed []Card) (topIDs, bottomIDs []string) {
	for _, c := range revealed {
		topIDs = append(topIDs, c.ID)
	}
	return topIDs, nil
}

// randomPeekPartition 将亮出牌伪随机分配至顶/底（sim 与人类强交互兜底，保证合法分区）。
func randomPeekPartition(revealed []Card, salt int) (topIDs, bottomIDs []string) {
	if len(revealed) == 0 {
		return nil, nil
	}
	for i, c := range revealed {
		if (i+salt)%2 == 0 {
			topIDs = append(topIDs, c.ID)
		} else {
			bottomIDs = append(bottomIDs, c.ID)
		}
	}
	if len(topIDs) == 0 {
		topIDs = append(topIDs, bottomIDs[0])
		bottomIDs = bottomIDs[1:]
	}
	if len(bottomIDs) == 0 && len(topIDs) > 1 {
		bottomIDs = append(bottomIDs, topIDs[len(topIDs)-1])
		topIDs = topIDs[:len(topIDs)-1]
	}
	return topIDs, bottomIDs
}

func (g *Game) runAIPreparePhase(seat int, events *[]GameEvent) {
	if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
		return
	}
	for attempt := 0; attempt < 8; attempt++ {
		if !g.runAIActiveSkills(seat, events) {
			break
		}
		if g.Phase == PhaseResponse && g.Pending != nil && g.Pending.ResponseMode == ResponseModePeekDeck {
			_ = g.finishPeekDeckAsAI(seat, events)
		}
		if g.Phase != PhasePlaying || g.TurnStep != StepPrepare || g.CurrentTurn != seat {
			return
		}
	}
	_ = g.PassPrepare(seat, events)
}

func (g *Game) isPeekDeckPending() bool {
	return g.Pending != nil && g.Pending.ResponseMode == ResponseModePeekDeck
}

func (g *Game) peekDeckSkillID() string {
	if g.Pending == nil {
		return ""
	}
	return g.Pending.SkillID
}
