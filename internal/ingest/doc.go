// Package ingest is the ingest bounded context: institutions, import
// batches, and the mapping of what a connector reads from an institution into
// ledger transactions, balance assertions and ingest problems (docs/DESIGN.md
// D-7, D-18, D-23).
//
// A connector's adapter parses the institution's own format and translates it
// here; broker vocabulary never leaves the adapter. A mapping is not an
// all-or-nothing affair: a row the mapping cannot or will not turn into ledger
// data becomes an ingest problem with its reason, and the rest is imported.
//
// Everything here is a pure function over values.
package ingest
