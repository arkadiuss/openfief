// Package views holds the pure rules behind the read API (docs/DESIGN.md
// D-6): paging transactions, totalling valuations per date, and the
// allocation the built-in UI charts (D-30), grouped by asset class, currency,
// institution or sleeve.
//
// The use cases that load what these rules need are in internal/views/app;
// the HTTP adapter that serves them as the read-only /v1 API is
// internal/views/connectapi.
package views
