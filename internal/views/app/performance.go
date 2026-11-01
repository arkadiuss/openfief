package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// Performance is a time-weighted return over the days it covers, From to
// To, the portfolio's or one account's, with whether a revaluation that
// could change it is pending or running.
type Performance struct {
	From, To kernel.Date
	// Start is the charts' default start (D-138), which is also where
	// tracking began: zero when nothing is recorded.
	Start  kernel.Date
	Series views.PerformanceSeries
	// Pending reports a revaluation of some account that is pending or
	// running: its stored valuations may change.
	Pending bool
}

// Performance returns the portfolio's time-weighted return (D-61) per day,
// with its value including cash, in the reporting currency, over the days
// from through to, both inclusive. A zero from means the charts' default
// start (D-138, views.ChartStart): the latest day an account was seeded
// from a starting balance, or the first transaction's date when none was.
// A from before the first transaction starts there; the index is 100 on
// the first day. A zero to means the latest valued-through date of any
// account. The series stops at the last day all accounts are valued for
// and every rate it needs is stored, naming the gap, so an account valued
// less far than the others shows as a gap; a portfolio with no
// transactions has no days.
func (v *Views) Performance(ctx context.Context, from, to kernel.Date) (Performance, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return Performance{}, err
	}
	rec, err := v.recorded(ctx, accounts)
	if err != nil {
		return Performance{}, err
	}
	links, err := v.ledger.TransferLinks(ctx)
	if err != nil {
		return Performance{}, fmt.Errorf("loading transfer links: %w", err)
	}
	return v.performance(ctx, accounts, &rec, rec.start, from, to, links)
}

// AccountPerformance returns one account's time-weighted return as
// Performance does the portfolio's, except that its default start is the
// account's own first day (D-138) and every deposit and withdrawal is a
// cash flow of the account, a transfer from or to another of the owner's
// accounts included. An unknown account fails like Account.
func (v *Views) AccountPerformance(ctx context.Context, id ledger.AccountID, from, to kernel.Date) (Performance, error) {
	acc, err := v.Account(ctx, id)
	if err != nil {
		return Performance{}, err
	}
	accounts := []ledger.Account{acc}
	rec, err := v.recorded(ctx, accounts)
	if err != nil {
		return Performance{}, err
	}
	return v.performance(ctx, accounts, &rec, rec.first, from, to, nil)
}

// performance computes the time-weighted return of accounts, whose record
// is rec, from the start chosen as from over def (views.ChosenStart)
// through to, leaving out the cash flows of the confirmed transfers in
// links.
func (v *Views) performance(ctx context.Context, accounts []ledger.Account, rec *record, def, from, to kernel.Date, links []ledger.TransferLink) (Performance, error) {
	if len(rec.txs) == 0 {
		return Performance{From: from, To: to}, nil
	}
	from = views.ChosenStart(from, rec.first, def)
	if to.IsZero() {
		// When nothing is valued that late, the series is the first day, unvalued.
		to = later(rec.latest, from)
	}
	if rangeErr := views.ValidateRange(from, to); rangeErr != nil {
		return Performance{}, fmt.Errorf("performance: %w", rangeErr)
	}
	in := &views.PerformanceInput{From: from, To: to, ValuedThrough: rec.through, Transactions: rec.txs, Reporting: v.reporting, Links: links}
	var err error
	if in.Valuations, err = v.valuationsFor(ctx, accounts, from, to); err != nil {
		return Performance{}, err
	}
	if in.Rates, err = v.cashRates(ctx, rec.txs, to); err != nil {
		return Performance{}, err
	}
	out := Performance{From: from, To: to, Start: def, Pending: rec.pending}
	if out.Series, err = views.Performance(in); err != nil {
		return Performance{}, fmt.Errorf("computing performance: %w", err)
	}
	return out, nil
}

// record is what the portfolio has recorded, across the accounts that have
// transactions: their current transactions, the first date of any, the
// charts' default start (views.ChartStart), the earliest and the latest
// date an account is valued through, and whether a revaluation is pending
// or running.
type record struct {
	txs                           []ledger.Transaction
	first, start, through, latest kernel.Date
	pending                       bool
}

// recorded loads the record of accounts. An account with transactions that
// was never valued counts as valued through the day before its first one.
func (v *Views) recorded(ctx context.Context, accounts []ledger.Account) (record, error) {
	var (
		rec       record
		histories []views.AccountHistory
	)
	for i := range accounts {
		txs, err := v.ledger.CurrentTransactions(ctx, accounts[i].ID)
		if err != nil {
			return record{}, fmt.Errorf("loading transactions: %w", err)
		}
		if len(txs) == 0 {
			continue
		}
		s, err := v.statuses.Status(ctx, accounts[i].ID)
		if err != nil {
			return record{}, fmt.Errorf("loading valuation status: %w", err)
		}
		h := views.HistoryOf(txs)
		histories = append(histories, h)
		through := s.ValuedThrough
		if through.IsZero() {
			through = h.First.AddDays(-1)
		}
		if len(rec.txs) == 0 {
			rec.first, rec.through, rec.latest = h.First, through, through
		}
		rec.first, rec.through, rec.latest = earlier(rec.first, h.First), earlier(rec.through, through), later(rec.latest, through)
		rec.pending = rec.pending || s.Pending || s.Running
		rec.txs = append(rec.txs, txs...)
	}
	rec.start, _ = views.ChartStart(histories)
	return rec, nil
}

// earlier returns the earlier of a and b.
func earlier(a, b kernel.Date) kernel.Date {
	if b.Before(a) {
		return b
	}
	return a
}

// later returns the later of a and b.
func later(a, b kernel.Date) kernel.Date {
	if b.After(a) {
		return b
	}
	return a
}

// firstDate returns the earliest date in txs, which is not empty.
func firstDate(txs []ledger.Transaction) kernel.Date {
	first := txs[0].Date
	for i := range txs {
		if txs[i].Date.Before(first) {
			first = txs[i].Date
		}
	}
	return first
}

// valuationsFor loads every account's stored valuations dated from to to.
func (v *Views) valuationsFor(ctx context.Context, accounts []ledger.Account, from, to kernel.Date) ([]valuation.HoldingValuation, error) {
	var all []valuation.HoldingValuation
	for i := range accounts {
		vals, err := v.valuations.Valuations(ctx, accounts[i].ID, from, to)
		if err != nil {
			return nil, fmt.Errorf("loading valuations: %w", err)
		}
		all = append(all, vals...)
	}
	return all, nil
}

// cashRates loads the FX rates dated through to of every currency other than
// the reporting one that the cash of txs moves in.
func (v *Views) cashRates(ctx context.Context, txs []ledger.Transaction, to kernel.Date) ([]valuation.FXRate, error) {
	seen := map[kernel.Currency]bool{v.reporting: true}
	var out []valuation.FXRate
	for i := range txs {
		for _, p := range txs[i].Postings {
			c := p.Amount.Currency()
			if p.Book.Kind != ledger.BookCash || seen[c] {
				continue
			}
			seen[c] = true
			rates, err := v.rates.FXRates(ctx, c, to)
			if err != nil {
				return nil, fmt.Errorf("loading FX rates: %w", err)
			}
			out = append(out, rates...)
		}
	}
	return out, nil
}
