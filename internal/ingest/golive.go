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
// an opening batch, that is its seeding transactions: the ledger refuses
// to seed an account twice, and the opening batch seeds the others. In a
// transactions batch, it is every transaction dated on or before its
// go-live date: the opening lots and opening cash already hold their
// effect. Each left-out transaction becomes a skipped problem named by its
// key. Transactions of accounts not in goLive, balance assertions and the
// rest of m are kept.
func KeepFromGoLive(m *Mapping, mode Mode, goLive map[ledger.AccountID]kernel.Date) Mapping {
	out := *m
	out.Transactions = nil
	out.Problems = slices.Clone(m.Problems)
	for i := range m.Transactions {
		t := &m.Transactions[i]
		day, seeded := goLive[t.Account]
		switch {
		case !seeded:
		case mode == ModeOpening && ledger.IsSeeding(t):
			out.Problems = append(out.Problems, Skipped(t.Key.String(), "the account is already seeded"))
			continue
		case mode == ModeTransactions && !t.Date.After(day):
			out.Problems = append(out.Problems, Skipped(t.Key.String(), "a transaction before go-live "+day.String()))
			continue
		}
		out.Transactions = append(out.Transactions, *t)
	}
	return out
}
