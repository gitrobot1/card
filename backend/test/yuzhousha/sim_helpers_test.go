package engine_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	engine "github.com/time/card/backend/internal/game/yuzhousha/engine"
	"github.com/time/card/backend/internal/game/yuzhousha/engine/mode"
)

const (
	defaultSimMaxSteps   = 8000
	identitySimMaxSteps  = 20000
	identity8SimMaxSteps = 30000
)

func expectedDeckSizeFor(g *engine.Game) int {
	return mode.DeckProfileFor(g.Mode).TotalCards()
}

func simTraceEnabled() bool {
	return os.Getenv("CARD_SIM_TRACE") == "1"
}

func enableAllAI(g *engine.Game) {
	for i := range g.Players {
		g.Players[i].IsAI = true
	}
}

func gameFingerprint(g *engine.Game) string {
	pending := ""
	if g.Pending != nil {
		pending = g.Pending.ResponseMode + "@s" + fmt.Sprint(g.Pending.SourceIndex)
		if g.Pending.GanglieIndex > 0 {
			pending += fmt.Sprintf(":g%d", g.Pending.GanglieIndex)
		}
		if g.Pending.BaguaUsed {
			pending += ":bagua"
		}
	}
	seats := make([]string, len(g.Players))
	for i, p := range g.Players {
		seats[i] = fmt.Sprintf("%d:%d:%d:%d", i, len(p.Hand), len(p.JudgeArea), p.HP)
	}
	return fmt.Sprintf("%s|%s|%d|%s|%s",
		g.Phase, g.TurnStep, g.CurrentTurn, pending, strings.Join(seats, ","))
}

func countCardsInPlay(g *engine.Game) int {
	n := len(g.DrawPile) + len(g.DiscardPile)
	for _, p := range g.Players {
		n += len(p.Hand) + len(p.JudgeArea)
		for _, slot := range []*engine.Card{p.Weapon, p.Armor, p.PlusHorse, p.MinusHorse} {
			if slot != nil {
				n++
			}
		}
	}
	if g.Pending != nil {
		n += len(g.Pending.RevealedCards)
		if g.Pending.JudgeCard.ID != "" {
			n++
		}
	}
	return n
}

func assertGameInvariants(t *testing.T, g *engine.Game) {
	t.Helper()
	for i, p := range g.Players {
		if p.HP < 0 || p.HP > p.MaxHP {
			t.Fatalf("player %d hp out of range: %d/%d", i, p.HP, p.MaxHP)
		}
	}
	total := countCardsInPlay(g)
	want := expectedDeckSizeFor(g)
	if total != want {
		t.Fatalf("card conservation: expected %d cards in play, got %d (phase=%s step=%s pending=%v)",
			want, total, g.Phase, g.TurnStep, g.Pending)
	}
}

func responseActor(g *engine.Game) (int, bool) {
	if g.Pending == nil {
		return -1, false
	}
	p := g.Pending
	switch p.ResponseMode {
	case engine.ResponseModeGuanYuFollow, engine.ResponseModeQilinBow:
		return p.TargetIndex, true
	case engine.ResponseModeSkillGanglieChoice:
		return p.TargetIndex, true
	case engine.ResponseModeSkillGuicai, engine.ResponseModeSkillFankui,
		engine.ResponseModeSkillJianxiong, engine.ResponseModeSkillGanglieOffer,
		engine.ResponseModeSkillYijiOffer, engine.ResponseModeSkillYijiGive:
		return p.TargetIndex, true
	case engine.ResponseModePeekDeck:
		return p.TargetIndex, true
	case engine.ResponseModeWuguPick:
		return p.WuguPickSeat, true
	case engine.ResponseModeDying:
		return p.SourceIndex, true
	default:
		return p.TargetIndex, true
	}
}

func forceProgress(g *engine.Game, events *[]engine.GameEvent) error {
	if g.IsFinished() {
		return nil
	}
	if g.Phase == engine.PhaseResponse && g.Pending != nil {
		switch g.Pending.ResponseMode {
		case engine.ResponseModePeekDeck:
			seat := g.Pending.TargetIndex
			return g.FinishPeekDeckForSim(seat, events)
		case engine.ResponseModeWuxiekTrick, engine.ResponseModeWuxiekLebu,
			engine.ResponseModeWuxiekBingliang, engine.ResponseModeWuxiekShandian:
			seat := g.Pending.TargetIndex
			return g.PassResponse(seat, events)
		case engine.ResponseModeWuguPick:
			return g.AutoPickWuguForSim(events)
		case engine.ResponseModeSkillFankui:
			seat, ok := responseActor(g)
			if !ok {
				return fmt.Errorf("no fankui actor")
			}
			return g.PassFankui(seat, events)
		}
		seat, ok := responseActor(g)
		if !ok {
			return fmt.Errorf("no response actor")
		}
		return g.PassResponse(seat, events)
	}
	if g.Phase == engine.PhaseResponse && g.Pending == nil {
		g.Phase = engine.PhasePlaying
		if g.TurnStep == "" {
			g.TurnStep = engine.StepPlay
		}
		return nil
	}
	if g.Phase == engine.PhasePlaying {
		seat := g.CurrentTurn
		if seat >= 0 && seat < len(g.Players) && g.Players[seat].HP <= 0 {
			return g.ForceSkipDeadTurnForTest(events)
		}
		switch g.TurnStep {
		case engine.StepPrepare:
			return g.PassPrepare(seat, events)
		case engine.StepDraw:
			return g.PassDrawPhase(seat, events)
		case engine.StepPlay:
			return g.EndPlay(seat, events)
		case engine.StepDiscard:
			g.AutoDiscardForSim(seat, events)
			return g.EndTurnForSim(events)
		}
	}
	return fmt.Errorf("cannot force progress: phase=%s step=%s", g.Phase, g.TurnStep)
}

type simResult struct {
	steps    int
	finished bool
	stuck    bool // fingerprint 不变或 forceProgress 失败
	timeout  bool // 达到 maxSteps 仍未结束
}

func appendEventTrail(trail []engine.GameEvent, events []engine.GameEvent) []engine.GameEvent {
	if len(events) == 0 {
		return trail
	}
	trail = append(trail, events...)
	const maxTrail = 25
	if len(trail) > maxTrail {
		trail = trail[len(trail)-maxTrail:]
	}
	return trail
}

func runAISimulation(t *testing.T, g *engine.Game, maxSteps int) simRun {
	t.Helper()
	enableAllAI(g)
	steps := 0
	var trail []engine.GameEvent
	var stuckAtFP string
	var forceErr string

	for !g.IsFinished() && steps < maxSteps {
		before := gameFingerprint(g)
		var events []engine.GameEvent
		if engine.RunAIActionStep(g, &events) {
			if simTraceEnabled() {
				trail = appendEventTrail(trail, events)
			}
			if gameFingerprint(g) != before {
				steps++
				continue
			}
		}
		if err := forceProgress(g, &events); err != nil {
			forceErr = err.Error()
			if simTraceEnabled() {
				trail = appendEventTrail(trail, events)
			}
			return simRun{
				result:     simResult{steps: steps, stuck: true},
				lastEvents: trail,
				stuckAtFP:  before,
				forceErr:   forceErr,
			}
		}
		if simTraceEnabled() {
			trail = appendEventTrail(trail, events)
		}
		if gameFingerprint(g) == before {
			return simRun{
				result:     simResult{steps: steps, stuck: true},
				lastEvents: trail,
				stuckAtFP:  before,
			}
		}
		steps++
	}

	return simRun{
		result: simResult{
			steps:    steps,
			finished: g.IsFinished(),
			timeout:  !g.IsFinished(),
		},
		lastEvents: trail,
		stuckAtFP:  stuckAtFP,
	}
}
