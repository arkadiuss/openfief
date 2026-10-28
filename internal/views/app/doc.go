// Package app holds the views context's use cases: the queries the read-only
// /v1 API answers (docs/DESIGN.md D-6), composed from the ledger's queries,
// the stored valuations and their status (D-36), and FX rates (D-20).
//
// The ports it reads through (Ledger, Valuations, ValuationStatuses, FXRates)
// and the TokenVerifier the API authenticates with are declared here, by the
// consumer. Grouping and sums are pure functions in package views.
package app
