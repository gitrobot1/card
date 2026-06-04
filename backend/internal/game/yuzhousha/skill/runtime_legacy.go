package skill

// LegacyRuntime 标记仍挂在完整 Runtime 上的技能专用方法。
// 新技能应优先使用 EnemiesOf / AlliesOf / DrawSkillCards 等通用能力，
// 或通过 Decl 的 hook 字段（DrawCountBonus、OnTurnEnd、OnHandEmpty）声明行为。
//
// 下列方法保留以兼容现有 engine 实现，后续迁移完成后可逐步移除：
//   - TuxiTakeFrom, FankuiTakeFrom, QixiTakeFrom, ActivateLuoyi, …
type LegacyRuntime interface {
	Runtime
}
