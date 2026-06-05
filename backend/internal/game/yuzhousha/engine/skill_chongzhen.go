package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

const SkillChongzhen = skill.IDChongzhen

func isChongzhenTriggerKind(kind string) bool {
	switch kind {
	case CardSha, CardGuoHe, CardTanNang, CardJueDou, CardLeBu, CardBingLiang:
		return true
	default:
		return false
	}
}

// notifyBecameTarget 某角色成为牌的目标后（如【冲阵】）。
func (g *Game) notifyBecameTarget(target, source int, card Card, events *[]GameEvent) {
	if target < 0 || target >= len(g.Players) || !isChongzhenTriggerKind(card.Kind) {
		return
	}
	g.tryChongzhen(target, source, card, events)
}

func (g *Game) tryChongzhen(seat int, source int, card Card, events *[]GameEvent) {
	if !g.hasSkill(seat, SkillChongzhen) {
		return
	}
	p := &g.Players[seat]
	var ids []string
	for _, c := range p.Hand {
		if c.Suit == card.Suit {
			ids = append(ids, c.ID)
		}
	}
	if len(ids) == 0 {
		return
	}
	if !p.IsAI {
		return
	}
	count := len(ids)
	if count > 2 {
		count = 2
	}
	ids = ids[:count]
	discarded := make([]Card, 0, count)
	for _, id := range ids {
		idx, _, ok := g.findCard(seat, id)
		if !ok {
			continue
		}
		c := g.removeHandCard(seat, idx, events)
		g.DiscardPile = append(g.DiscardPile, c)
		discarded = append(discarded, c)
	}
	if len(discarded) == 0 {
		return
	}
	g.Message = fmt.Sprintf("%s 发动【冲阵】，弃 %d 张摸 %d 张", p.Name, len(discarded), len(discarded))
	g.appendSkillEvent(events, SkillChongzhen, seat, source, g.Message)
	g.runCardsDiscardedHooks(seat, "chongzhen", discarded, events)
	g.drawSkillCards(seat, SkillChongzhen, len(discarded), g.Message, events)
}
