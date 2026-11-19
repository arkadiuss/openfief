// Package answer holds the Ask box's answerers: the six read-only commands
// (what needs me, show, sync status, allocation, performance and value), the
// questions and tool arguments that name them (Parse, Tool), and the cards
// they answer with (Card), read from the read API with the caller's own token.
//
// It is the one place both callers get their answers from: the built-in UI's
// Ask box, which parses a typed question, lets a model call the tools and draws
// the cards as HTML, and the core's MCP server, which offers the same tools
// (Toolbox) and serialises the same cards as JSON (CardResult). Neither knows
// the other: the package imports only the shared kernel and the SDK's generated
// read API contract, never the UI, so the core links none of the page's
// templates, assets or design system.
//
// What only a page knows comes in through small values and interfaces: the read
// API (Reader), the caller's token and number format (Caller), the clock, and
// where a card's links go (Links, empty for MCP). A card holds data, not
// markup. Nothing here changes a record, logs a question or keeps anything
// between calls (GR-12, D-203).
package answer
