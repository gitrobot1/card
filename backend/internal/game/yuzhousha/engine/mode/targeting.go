package mode

// Card kind strings for targeting (mirror engine card kinds; no engine import).
const (
	TargetSha      = "sha"
	TargetGuohe    = "guohe"
	TargetTannang  = "tannang"
	TargetJuedou   = "juedou"
	TargetLebu     = "lebu"
	TargetBingliang = "bingliang"
)

// TargetContext extends Context with combat/targeting queries for play validation.
type TargetContext interface {
	Context
	CanAttack(from, to int) bool
	HasTakeableCard(target int) bool
	CanBingliangTarget(from, to int) bool
	TargetBlocked(target int, cardKind string) bool
	PlayerHP(seat int) (hp, maxHP int)
}

// ValidPlayTargets returns enemy seats that are legal targets for cardKind from source.
func ValidPlayTargets(ctx TargetContext, source int, cardKind string) []int {
	candidates := EnemiesOf(ctx, source)
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
	if !IsEnemy(ctx, source, target) || ctx.AliveHP(target) <= 0 {
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
	default:
		if needsOpponentTarget(cardKind) {
			return true
		}
		return false
	}
}

func needsOpponentTarget(kind string) bool {
	switch kind {
	case TargetGuohe, TargetTannang, TargetJuedou, TargetLebu, TargetBingliang:
		return true
	default:
		return false
	}
}

// PickAITarget chooses a seat to target for cardKind; falls back to DefaultEnemy.
func PickAITarget(ctx TargetContext, source int, cardKind string) int {
	valid := ValidPlayTargets(ctx, source, cardKind)
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
