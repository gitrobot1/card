package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// ====== 绝境-高达一号 ======
// 锁定技，你跳过摸牌阶段。你的手牌数始终为4。

// 跳过摸牌阶段
func (g *Game) gundamSkipDrawPhase(seat int) bool {
	return g.hasSkill(seat, skill.IDJuejingGundam)
}

// 手牌数始终为4（在摸牌/出牌/弃牌后自动补牌）
func (g *Game) gundamSyncHandSize(seat int, events *[]GameEvent) {
	if !g.hasSkill(seat, skill.IDJuejingGundam) {
		return
	}
	p := &g.Players[seat]
	targetSize := 4

	if len(p.Hand) < targetSize {
		// 手牌不足，补到4
		need := targetSize - len(p.Hand)
		g.drawCards(seat, need, events)
	} else if len(p.Hand) > targetSize {
		// 手牌超过4，需要弃置（但在弃牌阶段处理）
		// 这里只在出牌阶段结束后自动调整
	}
}

// ====== 斩将 ======
// 准备阶段，如果场上有【青釭剑】，你可以获得之。

func (g *Game) gundamZhanjiang(seat int, events *[]GameEvent) {
	if !g.hasSkill(seat, skill.IDZhanjiang) {
		return
	}
	// 检查场上是否有青釭剑（武器2）
	for i := range g.Players {
		p := &g.Players[i]
		if p.Weapon != nil && p.Weapon.Kind == CardWeapon2 {
			// 获得青釭剑
			sword := *p.Weapon
			g.removeEquipSkill(i, sword.Kind) // TagEquipSkill: 移除装备技能
			p.Weapon = nil
			g.notifyEquipLost(i, sword, "taken", events)
			g.Players[seat].Hand = append(g.Players[seat].Hand, sword)
			g.SyncCounts()
			msg := fmt.Sprintf("%s 发动【斩将】，获得 %s 的【青釭剑】", g.Players[seat].Name, p.Name)
			g.Message = msg
			*events = append(*events, GameEvent{
				Type:        "skill_trigger",
				PlayerIndex: seat,
				TargetIndex: i,
				SkillID:     skill.IDZhanjiang,
				Card:        &sword,
				Message:     msg,
			})
			return
		}
	}
}
