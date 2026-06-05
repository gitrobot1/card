package mode

import "fmt"

const (
	SoloIdentity5 = "identity_5"
	SoloIdentity8 = "identity_8"
)

// IsIdentityMode reports standard hidden-role identity modes (5p / 8p / future Np).
func IsIdentityMode(modeID string) bool {
	switch modeID {
	case SoloIdentity5, SoloIdentity8:
		return true
	default:
		return false
	}
}

// LordSkillsActive reports whether lord skills (激将/护驾/救援) take effect in this mode.
func LordSkillsActive(modeID string) bool {
	return modeID == Solo2v2 || IsIdentityMode(modeID)
}

const (
	RoleLord     = "lord"
	RoleLoyalist = "loyalist"
	RoleSpy      = "spy"
	RoleRebel    = "rebel"
)

const (
	IdentityTeamLordFaction = 0
	IdentityTeamRebel       = 1
	IdentityTeamSpy         = 2
)

func IsIdentity(ctx Context) bool {
	return IsIdentityMode(ctx.ModeID())
}

// IdentityContext exposes hidden-role state for identity mode rules.
type IdentityContext interface {
	Context
	IdentityLordSeat() int
	IdentityOf(seat int) string
	IdentityRevealed(seat int) bool
}

func IsLordFaction(role string) bool {
	return role == RoleLord || role == RoleLoyalist
}

func IdentityTeamOf(role string) int {
	switch role {
	case RoleLord, RoleLoyalist:
		return IdentityTeamLordFaction
	case RoleRebel:
		return IdentityTeamRebel
	case RoleSpy:
		return IdentityTeamSpy
	default:
		return IdentityTeamLordFaction
	}
}

func RoleLabel(role string) string {
	switch role {
	case RoleLord:
		return "主公"
	case RoleLoyalist:
		return "忠臣"
	case RoleSpy:
		return "内奸"
	case RoleRebel:
		return "反贼"
	default:
		return role
	}
}

func ValidateIdentity5Roles(roles []string) error {
	if len(roles) != 5 {
		return fmt.Errorf("identity_5 requires 5 roles, got %d", len(roles))
	}
	counts := map[string]int{}
	for _, r := range roles {
		switch r {
		case RoleLord, RoleLoyalist, RoleSpy, RoleRebel:
			counts[r]++
		default:
			return fmt.Errorf("unknown identity role: %s", r)
		}
	}
	if counts[RoleLord] != 1 || counts[RoleLoyalist] != 1 || counts[RoleSpy] != 1 || counts[RoleRebel] != 2 {
		return fmt.Errorf("identity_5 needs 1 lord, 1 loyalist, 1 spy, 2 rebels; got lord=%d loyalist=%d spy=%d rebel=%d",
			counts[RoleLord], counts[RoleLoyalist], counts[RoleSpy], counts[RoleRebel])
	}
	return nil
}

func ValidateIdentity8Roles(roles []string) error {
	if len(roles) != 8 {
		return fmt.Errorf("identity_8 requires 8 roles, got %d", len(roles))
	}
	counts := map[string]int{}
	for _, r := range roles {
		switch r {
		case RoleLord, RoleLoyalist, RoleSpy, RoleRebel:
			counts[r]++
		default:
			return fmt.Errorf("unknown identity role: %s", r)
		}
	}
	if counts[RoleLord] != 1 || counts[RoleLoyalist] != 2 || counts[RoleSpy] != 1 || counts[RoleRebel] != 4 {
		return fmt.Errorf("identity_8 needs 1 lord, 2 loyalists, 1 spy, 4 rebels; got lord=%d loyalist=%d spy=%d rebel=%d",
			counts[RoleLord], counts[RoleLoyalist], counts[RoleSpy], counts[RoleRebel])
	}
	return nil
}

// EvaluateIdentityWin checks win conditions after victim has died (already at 0 HP).
// killer is the damage source seat when known; use -1 if unknown.
func EvaluateIdentityWin(ctx IdentityContext, victim, killer int) (finished bool, winnerTeam int, message string) {
	if !IsIdentity(ctx) {
		return false, -1, ""
	}
	lord := ctx.IdentityLordSeat()
	if victim == lord {
		if duel, _ := identitySpyDuel(ctx, lord); duel {
			return true, IdentityTeamSpy, "主公在内奸对局中阵亡，内奸获胜"
		}
		return true, IdentityTeamRebel, "主公阵亡，反贼获胜"
	}

	rebelsAlive := 0
	spiesAlive := 0
	aliveCount := 0
	for seat := 0; seat < ctx.PlayerCount(); seat++ {
		if ctx.AliveHP(seat) <= 0 {
			continue
		}
		aliveCount++
		switch ctx.IdentityOf(seat) {
		case RoleRebel:
			rebelsAlive++
		case RoleSpy:
			spiesAlive++
		}
	}
	if spiesAlive == 1 && aliveCount == 1 {
		return true, IdentityTeamSpy, "内奸独自存活，内奸获胜"
	}
	if rebelsAlive == 0 && spiesAlive == 0 {
		return true, IdentityTeamLordFaction, "反贼与内奸全灭，主公阵营获胜"
	}
	return false, -1, ""
}

func IdentityPlayTargets(ctx Context, source int) []int {
	if !IsIdentity(ctx) {
		return nil
	}
	out := make([]int, 0, ctx.PlayerCount()-1)
	for i := 0; i < ctx.PlayerCount(); i++ {
		if i != source && ctx.AliveHP(i) > 0 {
			out = append(out, i)
		}
	}
	return out
}

func DefaultEnemyIdentity(ctx IdentityContext, seat int) int {
	lord := ctx.IdentityLordSeat()
	role := ctx.IdentityOf(seat)
	n := ctx.PlayerCount()

	preferLord := role == RoleRebel || role == RoleSpy
	if preferLord && ctx.AliveHP(lord) > 0 {
		return lord
	}

	if IsLordFaction(role) {
		for i := 1; i < n; i++ {
			next := (seat + i) % n
			if next == seat || ctx.AliveHP(next) <= 0 {
				continue
			}
			if ctx.IdentityOf(next) == RoleRebel {
				return next
			}
		}
		for i := 1; i < n; i++ {
			next := (seat + i) % n
			if next == seat || ctx.AliveHP(next) <= 0 {
				continue
			}
			if ctx.IdentityOf(next) == RoleSpy {
				return next
			}
		}
	}

	for i := 1; i < n; i++ {
		next := (seat + i) % n
		if next != seat && ctx.AliveHP(next) > 0 {
			return next
		}
	}
	return (seat + 1) % n
}

// identitySpyDuel reports whether only the spy remains alive besides the lord.
func identitySpyDuel(ctx IdentityContext, lord int) (active bool, spySeat int) {
	otherAlive := 0
	spySeat = -1
	for seat := 0; seat < ctx.PlayerCount(); seat++ {
		if seat == lord {
			continue
		}
		if ctx.AliveHP(seat) <= 0 {
			continue
		}
		otherAlive++
		if ctx.IdentityOf(seat) == RoleSpy {
			spySeat = seat
		}
	}
	return otherAlive == 1 && spySeat >= 0, spySeat
}
