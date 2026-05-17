// Package ui embeds the HyperFleet browser dashboard HTML asset.
package ui

import "embed"

// StaticFS holds the embedded static assets for the dashboard.
//
//go:embed static/index.html
var StaticFS embed.FS
