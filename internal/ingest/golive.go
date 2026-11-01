package ingest

import (
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// PlanSync returns the import batches, in order, of a scheduled sync of a
// connector instance bound to the accounts bound, where goLive holds the
// go-live date of each account already seeded (D-23, D-95):
//
//   - no bound account, or every one seeded: one transactions batch, so a
//     seeded account is synced in transactions mode with no manual step;
//   - none seeded: one opening batch, which seeds them;
//   - some seeded and some not, such as a broker account mapped after the
//     others were seeded (D-72): an opening batch that seeds the rest, then
//     a transactions batch for the seeded ones.
func PlanSync(bound []ledger.AccountID, goLive map[ledger.AccountID]kernel.Date) []Mode {
	seeded := 0
	for _, a := range bound {
		if _, ok := goLive[a]; ok {
			seeded++
		}
	}
	switch {
	case seeded == len(bound):
		return []Mode{ModeTransactions}
	case seeded == 0:
		return []Mode{ModeOpening}
	}
	return []Mode{ModeOpening, ModeTransactions}
}

// KeepFromGoLive returns m without what a sync must not book again for an
// account already seeded, whose go-live date goLive holds (D-23, D-95). In
// an opening batch, that is its seeding transactions and opening-lot ticker
// entries: the ledger refuses to seed an account twice, and the opening
// batch seeds the others. In a transactions batch, it is every transaction
// and ticker entry dated on or before its go-live date: the opening lots
// and opening cash already hold their effect. A statement upload in
// transactions mode is kept the same way, so a statement that reaches back
// before an account's start books only what came after it (D-136). Each
// left-out transaction or entry becomes a skipped problem named by its key.
// Transactions and entries of accounts not in goLive, balance assertions
// and the rest of m are kept; ticker entries must have their account.
func KeepFromGoLive(m *Mapping, mode Mode, goLive map[ledger.AccountID]kernel.Date) Mapping {
	out := *m
	out.Transactions = nil
	out.TickerEntries = nil
	out.Problems = slices.Clone(m.Problems)
	for i := range m.Transactions {
		t := &m.Transactions[i]
		if reason, gone := beforeGoLive(t.Account, t.Date, ledger.IsSeeding(t), mode, goLive); gone {
			out.Problems = append(out.Problems, Skipped(t.Key.String(), reason))
			continue
		}
		out.Transactions = append(out.Transactions, *t)
	}
	for i := range m.TickerEntries {
		e := &m.TickerEntries[i]
		if reason, gone := beforeGoLive(e.Header.Account, e.Header.Date, e.Kind == ledger.KindOpeningLot, mode, goLive); gone {
			out.Problems = append(out.Problems, Skipped(e.Header.Key.String(), reason))
			continue
		}
		out.TickerEntries = append(out.TickerEntries, *e)
	}
	return out
}

// beforeGoLive returns why a transaction or ticker entry of account on day,
// seeding it or not, is left out of a batch in mode, and whether it is.
func beforeGoLive(account ledger.AccountID, day kernel.Date, seeding bool, mode Mode, goLive map[ledger.AccountID]kernel.Date) (string, bool) {
	live, seeded := goLive[account]
	switch {
	case !seeded:
	case mode == ModeOpening && seeding:
		return "the account is already seeded", true
	case mode == ModeTransactions && !day.After(live):
		return "a transaction before go-live " + live.String(), true
	}
	return "", false
}
