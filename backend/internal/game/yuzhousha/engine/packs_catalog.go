package engine

import "github.com/time/card/backend/internal/game/yuzhousha/skill"

// PacksCatalog returns registered content pack manifests.
func PacksCatalog() []skill.PackManifest {
	return skill.AllPacks()
}
