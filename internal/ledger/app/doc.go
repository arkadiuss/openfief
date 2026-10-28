// Package app holds the ledger context's use cases: registering sleeves and
// accounts, importing transactions and balance assertions into the
// append-only, versioned transaction log (docs/DESIGN.md D-18), and the
// queries the read API answers from it.
//
// The ports it needs, such as Store, are declared here, by the consumer, and
// implemented by adapters such as internal/ledger/postgres. Domain rules stay
// in package ledger; this package loads what they need, calls them and
// stores the result.
package app
