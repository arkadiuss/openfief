package app

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// BondSeriesWait is how many days a held treasury bond series waits for
// the registry's ISIN and its terms from the Ministry of Finance's data
// before it raises an attention item (D-214). Until then, data that is
// late or can't be read stays silent.
const BondSeriesWait = 7

// IncompleteBondSeries is the check of held treasury bond series still
// without the registry's ISIN or without terms (D-214): each series some
// account holds today (heldSeries) under its provisional ISIN, or whose
// terms are not stored, and has held for BondSeriesWait days or more, is a
// condition, its subject the ISIN it is held under, once however many
// accounts hold it, its detail what is missing (incompleteDetail). It
// resolves itself once both arrive, the series re-keyed to the registry's
// ISIN with its terms stored, or once no account holds the series.
func IncompleteBondSeries(l Ledger, b BondTerms, today func() kernel.Date) Check {
	return Check{Kind: attention.KindBondSeriesIncomplete, Find: func(ctx context.Context) ([]attention.Condition, error) {
		stored, err := b.AllBondTerms(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading bond terms: %w", err)
		}
		accounts, err := l.Accounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading accounts: %w", err)
		}
		var held []heldSeries
		for i := range accounts {
			txs, err := l.CurrentTransactions(ctx, accounts[i].ID)
			if err != nil {
				return nil, fmt.Errorf("loading the transactions of %s: %w", accounts[i].ID, err)
			}
			held = append(held, bondSeriesHeld(txs, today())...)
		}
		return incompleteSeries(held, stored, today()), nil
	}}
}

// heldSeries is a treasury bond series one account holds: the instrument
// it is held under and the first day it was held.
type heldSeries struct {
	instrument ledger.Instrument
	series     pl.Series
	since      kernel.Date
}

// bondSeriesHeld returns the treasury bond series txs, one account's
// current transactions, hold on today: the instruments whose symbol is a
// series code and whose ISIN is Polish, provisional or the registry's,
// with a quantity above zero from the holding postings dated up to today,
// each with the date of its first such posting.
func bondSeriesHeld(txs []ledger.Transaction, today kernel.Date) []heldSeries {
	type state struct {
		qty   kernel.Quantity
		since kernel.Date
	}
	by := map[ledger.Instrument]*state{}
	for i := range txs {
		t := &txs[i]
		for _, p := range t.Postings {
			in := p.Book.Instrument
			if t.Date.After(today) || p.Book.Kind != ledger.BookHolding || pl.CheckBondISIN(in.ISIN.String()) != nil {
				continue
			}
			if by[in] == nil {
				by[in] = &state{since: t.Date}
			}
			by[in].qty = by[in].qty.Add(p.Quantity)
			by[in].since = earlier(by[in].since, t.Date)
		}
	}
	var out []heldSeries
	for in, st := range by {
		s, err := pl.ParseSeries(in.Symbol)
		if err != nil || st.qty.Sign() <= 0 {
			continue
		}
		out = append(out, heldSeries{instrument: in, series: s, since: st.since})
	}
	return out
}

// earlier returns the earlier of a and b.
func earlier(a, b kernel.Date) kernel.Date {
	if b.Before(a) {
		return b
	}
	return a
}

// incompleteSeries returns the condition of each series of held under its
// provisional ISIN or without terms in stored, held since BondSeriesWait
// days or more before today, by the earliest day any account held it, in
// subject order.
func incompleteSeries(held []heldSeries, stored map[ledger.ISIN]pl.Terms, today kernel.Date) []attention.Condition {
	first := map[ledger.Instrument]heldSeries{}
	for _, h := range held {
		if f, ok := first[h.instrument]; !ok || h.since.Before(f.since) {
			first[h.instrument] = h
		}
	}
	var out []attention.Condition
	for in, h := range first {
		noISIN := pl.IsProvisional(in.ISIN.String(), in.Symbol)
		_, hasTerms := stored[in.ISIN]
		if (!noISIN && hasTerms) || h.since.AddDays(BondSeriesWait).After(today) {
			continue
		}
		out = append(out, attention.Condition{
			Kind: attention.KindBondSeriesIncomplete, Subject: in.ISIN.String(), Detail: incompleteDetail(h.series, h.since, noISIN, !hasTerms),
		})
	}
	slices.SortFunc(out, func(a, b attention.Condition) int { return cmp.Compare(a.Subject, b.Subject) })
	return out
}

// incompleteDetail says what series s, held since since, still lacks: the
// registry's ISIN, its terms or both (D-214), and what the owner can do.
func incompleteDetail(s pl.Series, since kernel.Date, noISIN, noTerms bool) string {
	head := s.String() + ", held since " + since.String() + ", still has no "
	const terms = "; its holding is not valued until they arrive, or until you enter its terms with a lot"
	switch {
	case noISIN && noTerms:
		return head + "ISIN and no series terms: the Ministry of Finance's data doesn't list them yet" + terms
	case noISIN:
		return head + "ISIN: the Ministry of Finance's data doesn't list it yet, so it is held under a provisional one"
	}
	return head + "series terms: the Ministry of Finance's data doesn't list them yet" + terms
}
