// Package design embeds OpenFief's design system (docs/dev/design-system.md)
// for Go programs that serve it: the built-in UI serves these files as they
// are, so design/ stays the single source and nothing is copied out of it.
// The docs site mounts the same folder through Hugo.
package design

import "embed"

// Files holds the design system at its own paths: tokens.css, the
// self-hosted fonts with fonts.css and their licenses, the marks and
// Kasztan's poses.
//
//go:embed tokens.css fonts marks mascot
var Files embed.FS
