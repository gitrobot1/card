package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterQixiUsed       = "qixi_used_play"
	ResponseModeSkillQixi = "skill_qixi"
)

func (g *Game) hasBlackHandCard(seat int) bool {
	for _, c := range g.Players[seat].Hand {
		if skill.IsBlackSuit(c.Suit) {
			return true
		}
	}
	return false
}

func (g *Game) opponentHasHandCard(seat int) bool {
	opp := g.firstEnemyWhere(seat, func(e int) bool { return len(g.Players[e].Hand) > 0 })
	return len(g.Players[opp].Hand) > 0
}

func (g *Game) ActivateQixi(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillQixi) || g.getSkillCounter(seat, counterQixiUsed) > 0 {
		return ErrWrongPhase
	}
	opp := g.firstEnemyWhere(seat, func(e int) bool { return len(g.Players[e].Hand) > 0 })
	if len(g.Players[opp].Hand) == 0 {
		return ErrInvalidTarget
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || !skill.IsBlackSuit(cardObj.Suit) {
		return ErrInvalidCard
	}
	discarded := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, discarded)
	g.syncCounts()
	g.runCardsDiscardedHooks(seat, "cost", []Card{discarded}, events)
	g.setSkillCounter(seat, counterQixiUsed, 1)

	msg := fmt.Sprintf("%s 发动【奇袭】，请选择获得 %s 的一张手牌", g.Players[seat].Name, g.Players[opp].Name)
	actor := seat
	return g.OpenTakeWindow(TakeWindowConfig{
		SkillID:         skill.IDQixi,
		ResponseMode:    ResponseModeSkillQixi,
		ActorSeat:       seat,
		SubjectSeat:     opp,
		OriginSeat:      seat,
		MaxTake:         1,
		AllowedZones:    []ZoneID{ZoneHand},
		Destination:     TakeDestination{Zone: ZoneHand, Seat: seat},
		Message:         msg,
		EventType:       "qixi_take",
		SkillEventLabel: "奇袭",
		OnComplete: func(g *Game, events *[]GameEvent) error {
			return g.finishQixi(actor, events)
		},
	}, events)
}

// QixiTakeFrom 奇袭拿牌（TakeWindow 薄封装）。
func (g *Game) QixiTakeFrom(seat int, cardID string, events *[]GameEvent) error {
	return g.TakeOne(seat, ZoneHand, cardID, events)
}

func (g *Game) finishQixi(seat int, events *[]GameEvent) error {
	g.Pending = nil
	g.Phase = PhasePlaying
	g.TurnStep = StepPlay
	g.CurrentTurn = seat
	g.Message = fmt.Sprintf("%s 继续出牌", g.Players[seat].Name)
	g.resetTimer()
	return nil
}

func (g *Game) aiPickHandTakeTarget(target int) (zone, cardID string) {
	if len(g.Players[target].Hand) > 0 {
		return "hand", g.Players[target].Hand[0].ID
	}
	return "", ""
}
