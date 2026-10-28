// Package ui is the built-in UI (D-3): a view plugin that serves the owner's
// browser as its backend (backend for frontend). The browser talks only to
// the UI. The UI authenticates the owner through the core's AuthService,
// presenting its own projected ServiceAccount token (D-27), and keeps the
// session token the core issues in an HttpOnly cookie. It reads the record
// only through the public read API (D-6), with that session token as the
// bearer token, and renders server-side HTML with html/template and htmx.
//
// Pages: the login (passkey, D-26, and the demo's read-only sessions), the
// first-passkey enrollment page (D-60), the overview with the allocation
// and value-over-time charts (D-30) and every account's valuation status
// (D-36), an account's holdings, lots, cash and paged transactions, the
// flagged balance assertions, and the owner's settings: passkeys, script
// tokens (D-27) and enrollment links for another hostname (D-60).
//
// Every response carries a strict Content-Security-Policy: scripts and
// styles load only from the UI itself, nothing inline. Every state-changing
// request must come from the UI's own origin (Sec-Fetch-Site or Origin).
package ui
