// Package engine 宇宙杀对局引擎（1v1 / 2v2）：回合、出牌、响应、判定、AI。
//
// 模式与开发规范见 ../dev-guide.md；模式元数据见 engine/mode/。
//
// # 文件职责
//
//	model.go            数据类型
//	constants.go        阶段、牌种、响应模式
//	deck.go             牌堆
//	game.go             构造、基础工具、公开视图
//	turn.go             回合流转
//	phase_prepare.go    准备阶段 + 通用 PeekDeck 流程
//	play.go             主动出牌
//	response.go         响应与伤害结算
//	judge.go            判定区
//	tricks.go           延时锦囊与五谷/闪电/兵粮
//	weapons.go          武器特效
//	skill_runtime.go    skill.Runtime 适配
//	skill_hooks.go    统一 runSkillHooks
//	skill_damage.go   伤害后续链（continueAfterDamage）
//	skill_judge.go    判定 + 鬼才窗口
//	skill_fankui.go   反馈 + 铁骑/八卦/闪电判定结算
//	skill_luoshen.go  洛神连续判定
//	skill_register.go / skill_register_wei.go  复杂技注册
//	skill_actions.go    仁德/激将状态机
//	ai.go               AI
//
// 技能声明式框架见 ../skill/。
package engine
