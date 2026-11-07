package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// BondTerms reads the stored treasury bond series terms (D-97);
// *valuation/postgres.Store implements it.
type BondTerms interface {
	// AllBondTerms returns every stored series' terms by ISIN.
	AllBondTerms(ctx context.Context) (map[ledger.ISIN]pl.Terms, error)
}

// AssumedBondFees is the check of treasury bond fees carried forward past
// the reviewed fee table (D-97): each series an account holds on the day
// it is valued through whose early-redemption fee is assumed
// (pl.FeesToConfirm) is a condition, its subject the series' ISIN, once
// however many accounts hold it, its detail what the owner is asked
// (pl.AssumedFeeDetail). The owner confirming or overriding the fee, or the
// series no longer held, clears it, so the item resolves itself.
func AssumedBondFees(l Ledger, h Holdings, b BondTerms) Check {
	return Check{Kind: attention.KindBondFeeAssumed, Find: func(ctx context.Context) ([]attention.Condition, error) {
		stored, err := b.AllBondTerms(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading bond terms: %w", err)
		}
		held, err := heldISINs(ctx, l, h)
		if err != nil {
			return nil, err
		}
		terms := make([]pl.Terms, 0, len(stored))
		isinOf := map[pl.Series]ledger.ISIN{}
		for isin := range held {
			if t, ok := stored[isin]; ok {
				terms = append(terms, t)
				isinOf[t.Series] = isin
			}
		}
		confirm := pl.FeesToConfirm(terms)
		out := make([]attention.Condition, 0, len(confirm))
		for i := range confirm {
			out = append(out, attention.Condition{
				Kind: attention.KindBondFeeAssumed, Subject: isinOf[confirm[i].Series].String(), Detail: pl.AssumedFeeDetail(&confirm[i]),
			})
		}
		return out, nil
	}}
}

// heldISINs returns the ISINs every account holds on the day it is valued
// through, as its stored valuations of that day say; an account never
// valued holds none yet.
func heldISINs(ctx context.Context, l Ledger, h Holdings) (map[ledger.ISIN]bool, error) {
	accounts, err := l.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading accounts: %w", err)
	}
	out := map[ledger.ISIN]bool{}
	for i := range accounts {
		id := accounts[i].ID
		through, valued, err := h.ValuedThrough(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("loading the valued-through date of %s: %w", id, err)
		}
		if !valued {
			continue
		}
		rows, err := h.Valuations(ctx, id, through, through)
		if err != nil {
			return nil, fmt.Errorf("loading the valuations of %s: %w", id, err)
		}
		for j := range rows {
			out[rows[j].Instrument.ISIN] = true
		}
	}
	return out, nil
}
