package app

import (
	"context"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// SyncChange is what one sync the core started wrote, for its audit entry
// (D-211): counts and the records they were in, never an amount (GR-12). Each
// count is of what was new or differed from what was stored, so a sync that
// brought nothing new has an empty one.
type SyncChange struct {
	Instance string
	Batch    ingest.BatchID
	// Transactions counts the transaction versions the ledger stored.
	Transactions int
	// Prices and BondTerms count the prices and the bond terms recorded.
	Prices, BondTerms int
	// Accounts are the accounts a new version of a transaction was stored in,
	// in ID order.
	Accounts []ledger.AccountID
}

// Empty reports whether the sync changed no state.
func (c *SyncChange) Empty() bool {
	return c.Transactions == 0 && c.Prices == 0 && c.BondTerms == 0
}

// SyncAuditor keeps a sync's writes together with its audit entry (D-211).
// cmd/core implements it: it runs write in one database transaction and, only
// when write reports a change, enqueues the entry of the sync as the system in
// that transaction, so the writes and the entry commit or roll back together.
// A sync that changed nothing leaves no entry, only its attempt record
// (D-198). An error of write is returned as it is.
type SyncAuditor interface {
	AuditedSync(ctx context.Context, instance string, write func(ctx context.Context) (SyncChange, error)) error
}

// changeOf returns what a delivery wrote: the transaction versions imported,
// by account, and tally's prices and bond terms.
func changeOf(instance string, id ingest.BatchID, imported *ingest.ImportSummary, tally marketTally) SyncChange {
	c := SyncChange{
		Instance: instance, Batch: id, Transactions: imported.New + imported.NewVersions,
		Prices: tally.prices, BondTerms: tally.terms,
	}
	for _, a := range imported.Accounts {
		if a.New+a.NewVersions > 0 {
			c.Accounts = append(c.Accounts, a.Account)
		}
	}
	return c
}
