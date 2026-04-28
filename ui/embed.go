package ui

import "embed"

// Files contains the compiled React SPA served by the orchestrator.
// The embed path resolves relative to this file, so "dist" maps to ui/dist/.
//
//go:embed all:dist
var Files embed.FS
