package engine

import (
	"fmt"
)

// 注意：这是一个示例文件，展示如何使用 applyHPLossWithHook 实现【蛊惑】技能
// 完整的【蛊惑】技能实现需要更多的逻辑和交互

const (
	// GuhuoDamageAmount 蛊惑造成的血量流失量
	GuhuoDamageAmount = 1
)

// ExampleGuhuoEffect 示例：【蛊惑】技能效果
// 当【蛊惑】生效时，目标角色流失 1 点体力
func (g *Game) ExampleGuhuoEffect(source, target int, events *[]GameEvent) {
	// 使用 applyHPLossWithHook 触发血量流失钩子
	g.applyHPLossWithHook(target, GuhuoDamageAmount, "skill", source, "guhuo", events)
	
	// 记录事件
	*events = append(*events, GameEvent{
		Type:        "skill_effect",
		PlayerIndex: source,
		TargetIndex: target,
		Damage:      GuhuoDamageAmount,
		SkillID:     "guhuo",
		Message:     fmt.Sprintf("%s 的【蛊惑】生效，%s 流失 1 点体力", 
			g.Players[source].Name, g.Players[target].Name),
	})
	
	// 后续处理...
	// 例如：目标角色需要进行判定或其他操作
}

