// Package skill 提供宇宙杀技能框架：注册表、武将目录、可复用的声明式被动技。
// 复杂状态机（仁德交牌、激将 pending）由 engine 实现并通过 Register 挂载。
//
// 完整开发规范（六阶段架构、2v2 测试、前端 pending/animation 注册表）见同目录：
//
//	dev-guide.md
//
// # 开发原则
//
// 不做「抄技能描述」，先补引擎通用能力，武将技能尽量挂 hook / Runtime 接口。
// 缺机制先做机制，缺武将再填 Decl。2v2 敌友判定用 Runtime.EnemiesOf/AlliesOf，勿写死 OpponentOf。
//
// # 能力矩阵（1v1 / 2v2 共用引擎）
//
//	类型          代表技能           引擎能力                         状态
//	受伤触发      奸雄/反馈/刚烈     伤害链、可选响应窗口              部分（continueAfterDamage 链；濒死救援已接）
//	判定          洛神/鬼才/铁骑     判定区、改判、连续判定            部分（startJudge/鬼才/洛神/铁骑；LeBu/BingLiang 判定流程已有）
//	拿牌/弃牌     突袭/遗计/连营     选目标区牌、弃牌后触发            部分（takeTargetCard/反馈；HookCardsDiscarded 新增；主动拿牌窗口待泛化）
//	阶段技        观星/制衡/洛神     牌堆顶操作、多步 UI               部分（PreparePhase + PeekDeck；制衡待做）
//	转化/当牌     倾国/急救/青囊     黑当闪、弃牌当桃等                部分（HookCardPlaysAs；急救/青囊待做）
//	主公技        护驾/救援/激将     1v1 标记不可用、2v2 可用           部分（Meta.InactiveIn1v1；激将逻辑保留）
//
// # 新增技能指引（阶段 5 注册表优先）
//
//  1. 查 dev-guide.md（上级目录）决策树
//  2. 简单被动 → catalog_skills.go（勿在 engine/skill_register_*.go 重复 Register）
//  3. 新 response_mode → engine 开窗 + frontend pending/handlers.ts
//
//   - 简单被动（改杀次数、牌当牌）：catalog_skills.go 用 Decl + hook 字段
//   - 摸牌/结束/空手 hook：DrawCountBonus / OnTurnEnd / OnHandEmpty
//   - 花色/完杀/帷幕/绝情/无双/克己/激昂：EffectiveSuit / BlocksPeachUse / BlocksTrickTarget /
//     DamageAsHPLoss / ExtraResponsesNeeded / SkipsDiscardPhase / OnCardResolved
//   - 准备阶段看牌堆顶（观星等）：catalog_peek.go 挂载 PeekDeck + PreparePhase
//   - 受伤后可选（反馈类）：OnDamageDealt 或 engine 在 continueAfterDamage 排队
//   - 弃牌后（连营等）：OnCardsDiscarded
//   - 判定后（洛神等）：OnJudgeResult
//   - 复杂主动/主公技：engine/skill_register*.go + 专用状态机文件
//   - 新武将：data/heroes/*.json 按扩展包增加数据，由 load_heroes.go 加载
//   - 新皮肤：data/skins/*.json，id 格式 hero_id:skin_key，由 skins.go 加载
//   - 扩展包清单：data/packs/*.json，关联 hero_pack / skin_pack
//
// # Hook 一览（engine.runSkillHooks）
//
//	HookTargetBlocked / HookDistanceDelta / HookTrickIgnoresDistance
//	HookInstantTrickUsed / HookCardPlaysAs / HookUnlimitedSha
//	HookDamageDealt / HookJudgeResult / HookCardsDiscarded
package skill
