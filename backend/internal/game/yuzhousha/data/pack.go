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

// SPHeroesJSON is the SP hero pack (see heroes/sp.json).
//
//go:embed heroes/sp.json
var SPHeroesJSON []byte

// ShenHeroesJSON is the shen hero pack (see heroes/shen.json).
//
//go:embed heroes/shen.json
var ShenHeroesJSON []byte

// SPPackManifestJSON describes the SP content pack.
//
//go:embed packs/sp.json
var SPPackManifestJSON []byte

// ShenPackManifestJSON describes the shen content pack.
//
//go:embed packs/shen.json
var ShenPackManifestJSON []byte
