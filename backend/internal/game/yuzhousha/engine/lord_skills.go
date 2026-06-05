package engine

import "github.com/time/card/backend/internal/game/yuzhousha/engine/mode"

// lordSkillsActive reports whether lord skills (激将/护驾/救援) take effect in this mode.
func lordSkillsActive(modeID string) bool {
	return mode.LordSkillsActive(modeID)
}
