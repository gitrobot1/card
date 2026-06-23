package mode

// Card kind strings for targeting (mirror engine card kinds; no engine import).
const (
	TargetSha      = "sha"
	TargetGuohe    = "guohe"
	TargetTannang  = "tannang"
	TargetJuedou   = "juedou"
	TargetLebu     = "lebu"
	TargetBingliang = "bingliang"
	TargetHuogong  = "huogong"
	TargetTiesuo   = "tiesuo"
)

// TargetContext extends Context with combat/targeting queries for play validation.
type TargetContext interface {
	Context
	CanAttack(from, to int) bool
	HasTakeableCard(target int) bool
	CanBingliangTarget(from, to int) bool
	TargetBlocked(target int, cardKind string) bool
	PlayerHP(seat int) (hp, maxHP int)
	HandCount(seat int) int
	// HasJudgeKind 返回 target 判定区是否有 kind 类型的延时锦囊
	HasJudgeKind(target int, kind string) bool
	// LimuActive 返回 source 的立牧是否生效（判定区有牌且有立牧技能）
	LimuActive(source int) bool
	// TrickIgnoresDistance 返回 source 的锦囊是否无视距离（奇才等）
	TrickIgnoresDistance(source int, trickKind string) bool
}

// ValidPlayTargets returns legal target seats for cardKind from source.
// 所有牌均可对任何存活角色使用（包括队友）。
func ValidPlayTargets(ctx TargetContext, source int, cardKind string) []int {
	out := make([]int, 0, ctx.PlayerCount())
	for i := 0; i < ctx.PlayerCount(); i++ {
		if ctx.AliveHP(i) > 0 && IsValidPlayTarget(ctx, source, i, cardKind) {
			out = append(out, i)
		}
	}
	return out
}

// IsValidPlayTarget reports whether source may choose target for cardKind.
func IsValidPlayTarget(ctx TargetContext, source, target int, cardKind string) bool {
	if ctx.AliveHP(target) <= 0 {
		return false
	}
	// 铁索连环/顺手牵羊/过河拆桥：任意存活角色均可为目标（含队友）
	if cardKind == TargetTiesuo || cardKind == TargetGuohe || cardKind == TargetTannang {
		return true
	}
	if source == target {
		return false
	}
	if ctx.TargetBlocked(target, cardKind) {
		return false
	}
	// 延时锦囊：目标判定区已有同名牌则不可选
	if isDelayTrick(cardKind) && ctx.HasJudgeKind(target, cardKind) {
		return false
	}
	// 立牧生效时，所有牌都需要检查攻击范围（但距离计算已被忽略）
	limuActive := ctx.LimuActive(source)
	// 奇才：锦囊无距离限制，顺手牵羊/过河拆桥直接通过
	trickIgnoresDist := ctx.TrickIgnoresDistance(source, cardKind)
	switch cardKind {
	case TargetSha:
		return ctx.CanAttack(source, target)
	case TargetGuohe, TargetTannang:
		if trickIgnoresDist {
			return ctx.HasTakeableCard(target)
		}
		if limuActive {
			if !ctx.CanAttack(source, target) {
				return false
			}
		}
		return ctx.HasTakeableCard(target)
	case TargetBingliang:
		return ctx.CanBingliangTarget(source, target)
	case TargetJuedou, TargetLebu:
		// 注意：立牧只影响"杀"的攻击范围判断，不影响决斗和乐不思蜀
		return true
	case TargetHuogong:
		if limuActive {
			if !ctx.CanAttack(source, target) {
				return false
			}
		}
		return ctx.HandCount(target) > 0
	case TargetTiesuo:
		if limuActive {
			return ctx.CanAttack(source, target)
		}
		return true
	default:
		if needsOpponentTarget(cardKind) {
			if limuActive {
				return ctx.CanAttack(source, target)
			}
			return true
		}
		return false
	}
}

func needsOpponentTarget(kind string) bool {
	switch kind {
	case TargetGuohe, TargetTannang, TargetJuedou, TargetLebu, TargetBingliang, TargetHuogong, TargetTiesuo:
		return true
	default:
		return false
	}
}

func isDelayTrick(kind string) bool {
	switch kind {
	case TargetLebu, TargetBingliang, "shandian":
		return true
	default:
		return false
	}
}

// PickAITarget chooses a seat to target for cardKind; falls back to DefaultEnemy.
func PickAITarget(ctx TargetContext, source int, cardKind string) int {
	valid := ValidPlayTargets(ctx, source, cardKind)
	if Is3v3(ctx) {
		enemyCmd := CommanderSeat3v3(1 - TeamOf3v3(source))
		for _, t := range valid {
			if t == enemyCmd {
				return enemyCmd
			}
		}
	}
	if IsIdentity(ctx) {
		if ic, ok := ctx.(IdentityContext); ok {
			enemy := DefaultEnemyIdentity(ic, source)
			for _, t := range valid {
				if t == enemy {
					return enemy
				}
			}
		}
	}
	if len(valid) == 0 {
		return DefaultEnemy(ctx, source)
	}
	if cardKind != TargetSha {
		return valid[0]
	}
	var wounded, attackable []int
	for _, t := range valid {
		attackable = append(attackable, t)
		hp, maxHP := ctx.PlayerHP(t)
		if hp < maxHP {
			wounded = append(wounded, t)
		}
	}
	if len(wounded) > 0 {
		return wounded[0]
	}
	if len(attackable) > 0 {
		return attackable[0]
	}
	return valid[0]
}

// CheckWinAfterDamage runs team elimination win check when the mode uses teams.
func CheckWinAfterDamage(ctx Context, check TeamElimination, events *[]TeamEvent) bool {
	if Is2v2(ctx) || Is3pDdz(ctx) {
		return CheckTeamElimination(check, events)
	}
	return false
}
