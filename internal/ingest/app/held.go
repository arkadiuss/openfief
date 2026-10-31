package app

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// HeldSource says what the accounts bound to a connector instance hold
// (D-117). HoldingsScope implements it.
type HeldSource interface {
	HeldAccounts(ctx context.Context, accounts []ledger.AccountID) ([]ingest.HeldAccount, []ingest.BondTerms, error)
}

// AccountHoldings reads an account's holdings and open lots on a day;
// ledger/app.Ledger implements it.
type AccountHoldings interface {
	HoldingsOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Holding, error)
	LotsOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Lot, error)
}

// TaxWrappers reads an account's recorded tax wrapper (D-76), and false
// when none was recorded.
type TaxWrappers interface {
	TaxWrapper(ctx context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error)
}

// HoldingsScope works out what a sync tells a connector bound to accounts
// (D-117): for each account, in ID order and each once, its go-live date,
// tax wrapper, holdings and open lots today, and the stored series terms
// of every treasury bond among them. It reads data only; what a connector
// makes of it, such as the coupons a bond paid, is the connector's (the
// core holds no bond logic).
type HoldingsScope struct {
	Ledger  AccountHoldings
	Seeding AccountSeeding
	Bonds   BondSeries
	// Wrappers reads the accounts' tax wrappers; an account without one
	// recorded has Default.
	Wrappers TaxWrappers
	Default  jurisdiction.Wrapper
	// Today returns the day whose holdings count.
	Today func() kernel.Date
}

// HeldAccounts returns what accounts hold today and the terms of their
// bonds, each series once in ISIN order.
func (h *HoldingsScope) HeldAccounts(ctx context.Context, accounts []ledger.AccountID) ([]ingest.HeldAccount, []ingest.BondTerms, error) {
	ids := slices.Clone(accounts)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	today := h.Today()
	var (
		out   []ingest.HeldAccount
		terms []ingest.BondTerms
		known = map[ledger.ISIN]bool{}
	)
	for _, id := range ids {
		a, err := h.held(ctx, id, today)
		if err != nil {
			return nil, nil, err
		}
		for i := range a.Lots {
			if terms, err = h.addTerms(ctx, terms, known, a.Lots[i].Instrument); err != nil {
				return nil, nil, err
			}
		}
		out = append(out, a)
	}
	slices.SortFunc(terms, func(a, b ingest.BondTerms) int {
		return cmp.Compare(a.Instrument.ISIN.String(), b.Instrument.ISIN.String())
	})
	return out, terms, nil
}

// held returns what account id holds on today.
func (h *HoldingsScope) held(ctx context.Context, id ledger.AccountID, today kernel.Date) (ingest.HeldAccount, error) {
	a := ingest.HeldAccount{Account: id, Wrapper: string(h.Default)}
	goLive, seeded, err := h.Seeding.GoLive(ctx, id)
	if err != nil {
		return ingest.HeldAccount{}, fmt.Errorf("go-live of %s: %w", id, err)
	}
	if seeded {
		a.GoLive = goLive
	}
	w, recorded, err := h.Wrappers.TaxWrapper(ctx, id)
	if err != nil {
		return ingest.HeldAccount{}, fmt.Errorf("tax wrapper of %s: %w", id, err)
	}
	if recorded {
		a.Wrapper = string(w)
	}
	if a.Holdings, err = h.Ledger.HoldingsOn(ctx, id, today); err != nil {
		return ingest.HeldAccount{}, fmt.Errorf("holdings of %s: %w", id, err)
	}
	if a.Lots, err = h.Ledger.LotsOn(ctx, id, today); err != nil {
		return ingest.HeldAccount{}, fmt.Errorf("open lots of %s: %w", id, err)
	}
	return a, nil
}

// addTerms appends in's stored bond terms to terms, asking Bonds once per
// ISIN, and returns terms unchanged when in is no bond with terms.
func (h *HoldingsScope) addTerms(ctx context.Context, terms []ingest.BondTerms, known map[ledger.ISIN]bool, in ledger.Instrument) ([]ingest.BondTerms, error) {
	if known[in.ISIN] {
		return terms, nil
	}
	known[in.ISIN] = true
	t, bond, err := h.Bonds.BondTerms(ctx, in.ISIN)
	if err != nil {
		return nil, fmt.Errorf("bond terms of %s: %w", in.ISIN, err)
	}
	if !bond {
		return terms, nil
	}
	return append(terms, ingest.BondTerms{Instrument: in, Terms: t}), nil
}
