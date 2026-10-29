// Package treasurybonds is the Polish retail treasury bond connector's
// logic (D-9, D-49): it maps the bond lots the owner enters by hand in the
// built-in UI, which the core hands to the connector, into ledger
// transactions and the series terms valuation needs. Neither the registry
// nor the bank that sells the bonds offers an API (R-5), so manual entry is
// this connector's ingest path; the plugin plugins/treasury-bonds serves it
// to the core.
//
// A lot is what D-9 names (series, quantity, purchase date) plus its
// series' terms from the issue letter. Each lot becomes a deposit of its
// cost and the buy of its bonds, on its purchase date, in the one ledger
// account the connector instance is configured for. The bonds' value comes
// from the valuation context's bond rules (jurisdiction/pl), never from a
// price.
package treasurybonds
