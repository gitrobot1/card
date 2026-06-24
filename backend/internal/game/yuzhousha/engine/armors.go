package engine

import (
	"fmt"

	"github.com/time/card/backend/internal/game/yuzhousha/skill"
)

// ============================================================================
// 防具技能
// ============================================================================

// ============================================================================
// 仁王盾（Renwang）：黑色【杀】对你无效
// 参考 noname: renwang_skill, trigger: { target: "shaBegin" }, content: trigger.cancel()
// ============================================================================

// hasRenwangArmor 检查玩家是否装备仁王盾。
func (g *Game) hasRenwangArmor(seat int) bool {
	return g.Players[seat].Armor != nil && g.Players[seat].Armor.Kind == CardArmorRenwang
}

// renwangBlocksSha 检查仁王盾是否阻挡此杀（黑色杀无效）。
// 参考 noname: filter → event.card.name == "sha" && get.color(event.card) == "black"
func renwangBlocksSha(shaCard Card) bool {
	if !isSha(shaCard.Kind) {
		return false
	}
	return skill.IsBlackSuit(shaCard.Suit)
}

// ============================================================================
// 白银狮子（Baiyin）：伤害值>1时锁定为1，失去时若已受伤则回1血
// 参考 noname: baiyin_skill + baiyin_skill_lose
// ============================================================================

// hasBaiyinArmor 检查玩家是否装备白银狮子。
func (g *Game) hasBaiyinArmor(seat int) bool {
	return g.Players[seat].Armor != nil && g.Players[seat].Armor.Kind == CardArmorBaiyin
}

// baiyinReduceDamage 白银狮子：伤害值>1时锁定为1。
// 参考 noname: baiyin_skill, trigger: { player: "damageBegin4" }, content: trigger.num = 1
func (g *Game) baiyinReduceDamage(target int, damage *int) {
	if g.hasBaiyinArmor(target) && *damage > 1 {
		*damage = 1
	}
}

// handleBaiyinLose 白银狮子失去时，若已受伤则回复1点体力。
// 参考 noname: onLose → addTempSkill("baiyin_skill_lose") → player.recover()
// 调用时机：装备被替换、被拆、被顺、被弃置时。
func (g *Game) handleBaiyinLose(seat int, events *[]GameEvent) {
	if !g.hasBaiyinArmor(seat) {
		return
	}
	p := &g.Players[seat]
	if p.HP >= p.MaxHP {
		return // 满血不回复
	}
	p.HP++
	g.syncCounts()
	msg := fmt.Sprintf("【白银狮子】%s 失去防具，回复 1 点体力", p.Name)
	*events = append(*events, GameEvent{
		Type:        "baiyin_recover",
		PlayerIndex: seat,
		Heal:        1,
		Message:     msg,
	})
}
