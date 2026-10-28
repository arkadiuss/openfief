// Package ledger is the ledger bounded context: accounts, ledger accounts,
// transactions and their postings, instruments, holdings, lots and balance
// assertions (docs/DESIGN.md D-18, D-19, D-20, D-23).
//
// The transaction log is the source of truth. It is append-only and
// versioned: DecideImport says whether an incoming transaction is unchanged,
// new, or a new version that supersedes the current one, and the old version
// is never deleted. Holdings, lots and the balances that balance assertions
// are checked against are derived only from the current versions.
//
// Every transaction is double-entry: its postings sum to zero in every
// currency. A posting's sign is from the ledger account's side: positive
// means cash received, units acquired, a fee paid, tax withheld or currency
// bought; negative means the opposite. Every amount keeps its original
// currency; conversion to PLN belongs to valuation, not here.
//
// Everything here is a pure function over values.
package ledger
