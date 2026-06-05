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
}

// ValidPlayTargets returns legal target seats for cardKind from source.
func ValidPlayTargets(ctx TargetContext, source int, cardKind string) []int {
	if cardKind == TargetTiesuo {
		out := make([]int, 0, 2)
		if IsValidPlayTarget(ctx, source, source, cardKind) {
			out = append(out, source)
		}
		var candidates []int
		if IsIdentity(ctx) {
			candidates = IdentityPlayTargets(ctx, source)
		} else {
			candidates = EnemiesOf(ctx, source)
		}
		for _, t := range candidates {
			if IsValidPlayTarget(ctx, source, t, cardKind) {
				out = append(out, t)
			}
		}
		return out
	}
	var candidates []int
	if IsIdentity(ctx) {
		candidates = IdentityPlayTargets(ctx, source)
	} else {
		candidates = EnemiesOf(ctx, source)
	}
	out := make([]int, 0, len(candidates))
	for _, t := range candidates {
		if IsValidPlayTarget(ctx, source, t, cardKind) {
			out = append(out, t)
		}
	}
	return out
}

// IsValidPlayTarget reports whether source may choose target for cardKind.
func IsValidPlayTarget(ctx TargetContext, source, target int, cardKind string) bool {
	if ctx.AliveHP(target) <= 0 {
		return false
	}
	if cardKind == TargetTiesuo && source == target {
		return true
	}
	if source == target {
		return false
	}
	if !IsIdentity(ctx) && !IsEnemy(ctx, source, target) {
		return false
	}
	if ctx.TargetBlocked(target, cardKind) {
		return false
	}
	switch cardKind {
	case TargetSha:
		return ctx.CanAttack(source, target)
	case TargetGuohe, TargetTannang:
		return ctx.HasTakeableCard(target)
	case TargetBingliang:
		return ctx.CanBingliangTarget(source, target)
	case TargetJuedou, TargetLebu:
		return true
	case TargetHuogong:
		return ctx.HandCount(target) > 0
	case TargetTiesuo:
		return true
	default:
		if needsOpponentTarget(cardKind) {
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
