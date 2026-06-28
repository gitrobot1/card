package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const (
	counterShuangxiongActive  = "shuangxiong_active"
	counterShuangxiongRefRed = "shuangxiong_ref_red"
)

func (g *Game) shuangxiongRefIsRed(seat int) bool {
	return g.getSkillCounter(seat, counterShuangxiongRefRed) > 0
}

func (g *Game) shuangxiongHandColorOk(card Card, refRed bool) bool {
	if card.Suit == "" {
		return false
	}
	return isRedSuit(card.Suit) != refRed
}

func (g *Game) hasShuangxiongJuedouCard(seat int) bool {
	if g.getSkillCounter(seat, counterShuangxiongActive) == 0 {
		return false
	}
	refRed := g.shuangxiongRefIsRed(seat)
	for _, c := range g.Players[seat].Hand {
		if g.shuangxiongHandColorOk(c, refRed) {
			return true
		}
	}
	return false
}

func (g *Game) ActivateShuangxiongDraw(seat int, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if !g.isDrawPhaseChoicePending(seat) || !g.hasSkill(seat, SkillShuangxiong) {
		return ErrWrongPhase
	}
	if len(g.DrawPile) == 0 {
		g.refillDrawPile()
	}
	if len(g.DrawPile) == 0 {
		return ErrInvalidCard
	}
	card := g.DrawPile[0]
	g.DrawPile = g.DrawPile[1:]
	g.Players[seat].Hand = append(g.Players[seat].Hand, card)
	g.SyncCounts()

	refRed := isRedSuit(card.Suit)
	g.setSkillCounter(seat, counterDrawChoicePending, 0)
	g.setSkillCounter(seat, counterShuangxiongActive, 1)
	if refRed {
		g.setSkillCounter(seat, counterShuangxiongRefRed, 1)
	} else {
		g.setSkillCounter(seat, counterShuangxiongRefRed, 0)
	}

	msg := fmt.Sprintf("%s 发动【双雄】，亮出 %s 并获得", g.Players[seat].Name, card.Label)
	g.Message = msg
	g.appendSkillEvent(events, skill.IDShuangxiong, seat, seat, msg)
	*events = append(*events, GameEvent{
		Type:        "skill_shuangxiong_draw",
		PlayerIndex: seat,
		Card:        &card,
		Message:     msg,
		SkillID:     skill.IDShuangxiong,
	})
	return g.advanceTurnAfterDraw(seat, events)
}

func (g *Game) ActivateShuangxiongJuedou(seat int, cardID string, events *[]GameEvent) error {
	if g.IsFinished() {
		return ErrGameOver
	}
	if g.Phase != PhasePlaying || g.TurnStep != StepPlay || g.CurrentTurn != seat {
		return ErrWrongPhase
	}
	if !g.hasSkill(seat, SkillShuangxiong) || g.getSkillCounter(seat, counterShuangxiongActive) == 0 {
		return ErrWrongPhase
	}
	target := g.opponentOf(seat)
	if g.runSkillHooks(nil, skill.HookCall{Kind: skill.HookTargetBlocked, Target: target, CardKind: CardJueDou}).Bool {
		return ErrInvalidTarget
	}
	idx, cardObj, ok := g.findCard(seat, cardID)
	if !ok || !g.shuangxiongHandColorOk(cardObj, g.shuangxiongRefIsRed(seat)) {
		return ErrInvalidCard
	}
	played := g.removeHandCard(seat, idx, events)
	g.DiscardPile = append(g.DiscardPile, played)
	g.runCardsDiscardedHooks(seat, "play", []Card{played}, events)

	juedou := Card{
		ID:    played.ID,
		Kind:  CardJueDou,
		Name:  "决斗",
		Suit:  played.Suit,
		Label: played.Label + "（双雄）",
	}
	msg := fmt.Sprintf("%s 发动【双雄】，将 %s 当【决斗】对 %s 使用", g.Players[seat].Name, played.Label, g.Players[target].Name)
	g.appendSkillEvent(events, skill.IDShuangxiong, seat, target, msg)
	*events = append(*events, GameEvent{
		Type:        "skill_shuangxiong_juedou",
		PlayerIndex: seat,
		TargetIndex: target,
		Card:        &juedou,
		Message:     msg,
		SkillID:     skill.IDShuangxiong,
	})
	return g.startWuxiekTrickWindow(seat, target, target, juedou, PlayTarget{SeatIndex: target}, events)
}
