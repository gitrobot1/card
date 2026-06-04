// Package data holds embedded static game data for yuzhousha.
package data

import _ "embed"

// StandardHeroesJSON is the standard hero pack (see heroes/standard.json).
//
//go:embed heroes/standard.json
var StandardHeroesJSON []byte

// StandardSkinsJSON is the standard skin pack (see skins/standard.json).
// Additional skins can be appended without changing hero definitions.
//
//go:embed skins/standard.json
var StandardSkinsJSON []byte

// StandardPackManifestJSON describes the standard content pack.
//
//go:embed packs/standard.json
var StandardPackManifestJSON []byte
