package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Holdings reads what the owner holds; ledger/app.Ledger implements it.
type Holdings interface {
	// Accounts returns every account.
	Accounts(ctx context.Context) ([]ledger.Account, error)
	// HoldingsOn returns account's non-zero holdings at the end of on.
	HoldingsOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Holding, error)
}

// BondSeries tells which instruments are Polish retail treasury bonds,
// which have no market price; the valuation store implements it.
type BondSeries interface {
	// BondTerms returns the series terms of the bond with isin, and false
	// when isin is no bond with stored terms.
	BondTerms(ctx context.Context, isin ledger.ISIN) (pl.Terms, bool, error)
}

// ModelTargets reads the model portfolio's targets while it is on (D-52);
// portfolio/app.Models implements it.
type ModelTargets interface {
	// MarketDataTargets returns the model portfolio's instrument targets
	// and its index targets by name, none while it is off.
	MarketDataTargets(ctx context.Context) ([]ledger.Instrument, []string, error)
}

// PricingScope works out what market-data connectors are asked to price
// (D-93): every instrument held today in any account, other than treasury
// bonds with stored terms, and the model portfolio's targets while it is
// on. A newly bought instrument is in it from the sync after its purchase
// is booked, with no configuration.
type PricingScope struct {
	Holdings Holdings
	Bonds    BondSeries
	// Model reads the model portfolio's targets; nil has none.
	Model ModelTargets
	// Today returns the day whose holdings count.
	Today func() kernel.Date
}

// MarketDataScope returns the scope (ingest.NewMarketDataScope).
func (p *PricingScope) MarketDataScope(ctx context.Context) (ingest.MarketDataScope, error) {
	held, err := p.held(ctx)
	if err != nil {
		return ingest.MarketDataScope{}, err
	}
	var indices []string
	if p.Model != nil {
		targets, ix, err := p.Model.MarketDataTargets(ctx)
		if err != nil {
			return ingest.MarketDataScope{}, fmt.Errorf("model portfolio targets: %w", err)
		}
		held, indices = append(held, targets...), ix
	}
	return ingest.NewMarketDataScope(held, indices), nil
}

// held returns the instruments held today in any account, bonds with
// stored terms left out.
func (p *PricingScope) held(ctx context.Context) ([]ledger.Instrument, error) {
	accounts, err := p.Holdings.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("accounts: %w", err)
	}
	today := p.Today()
	bonds := map[ledger.ISIN]bool{}
	var out []ledger.Instrument
	for i := range accounts {
		holdings, err := p.Holdings.HoldingsOn(ctx, accounts[i].ID, today)
		if err != nil {
			return nil, fmt.Errorf("holdings of %s: %w", accounts[i].ID, err)
		}
		for j := range holdings {
			in := holdings[j].Instrument
			bond, err := p.isBond(ctx, bonds, in.ISIN)
			if err != nil {
				return nil, err
			}
			if !bond {
				out = append(out, in)
			}
		}
	}
	return out, nil
}

// isBond reports whether isin has stored bond terms, asking Bonds once per
// ISIN and remembering the answer in known.
func (p *PricingScope) isBond(ctx context.Context, known map[ledger.ISIN]bool, isin ledger.ISIN) (bool, error) {
	if bond, ok := known[isin]; ok {
		return bond, nil
	}
	_, bond, err := p.Bonds.BondTerms(ctx, isin)
	if err != nil {
		return false, fmt.Errorf("bond terms of %s: %w", isin, err)
	}
	known[isin] = bond
	return bond, nil
}
