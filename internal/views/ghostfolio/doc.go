// Package ghostfolio is the Ghostfolio view (D-3, D-4, D-64): it reads the
// record through the read API and keeps the Ghostfolio accounts it owns
// equal to it, with a reconcile loop.
//
// Activity, account and asset profile are Ghostfolio's words, used only in
// this adapter: an activity is Ghostfolio's projection of a transaction.
// Project turns the record into the desired activities, Plan compares them
// with what the owned accounts hold, and Syncer applies the plan through
// Client, Ghostfolio's HTTP API. Project and Plan are pure.
package ghostfolio
