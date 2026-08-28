// Package cassor exposes the canonical portable skill as embedded installer assets.
package cassor

import "embed"

// Files contains the canonical portable Cassor skill package.
//
//go:embed all:SKILL.md all:references
var Files embed.FS
