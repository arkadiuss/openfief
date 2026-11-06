package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// ErrUnknownInstrument reports an ISIN that no account the caller may view
// posts to and that has no stored price.
var ErrUnknownInstrument = errors.New("unknown instrument")

// InstrumentPrices reads an instrument's stored prices from every source
// and which source each listing is valued at (D-94);
// *valuation/postgres.Store implements it.
type InstrumentPrices interface {
	// PriceHistory returns every stored price of the listings of isin dated
	// in the range from through to, in listing, date and source order.
	PriceHistory(ctx context.Context, isin string, from, to kernel.Date) ([]valuation.Price, error)
	// LatestPrices returns the latest stored price of each listing of isin
	// from each source, in listing then source order.
	LatestPrices(ctx context.Context, isin string) ([]valuation.Price, error)
	// SourceOf returns the source of series; ok is false when none.
	SourceOf(ctx context.Context, series valuation.Series) (valuation.SeriesSource, bool, error)
}

// WithInstrumentPrices reads instruments' prices from p; without it an
// instrument has no price history and no price source.
func WithInstrumentPrices(p InstrumentPrices) Option {
	return func(v *Views) { v.prices = p }
}

// Listing is one listing of an instrument with the source it is valued at
// and that source's latest price; Latest is nil when none.
type Listing struct {
	Instrument ledger.Instrument
	Source     string
	Latest     *valuation.Price
}

// Instrument is an instrument across the accounts the caller may view
// (D-195).
type Instrument struct {
	ISIN       string
	Listings   []Listing
	AssetClass views.AssetClass
	Classes    []views.InstrumentClass
}

// viewableTransactions returns the current transactions of every account
// the caller may view that post to isin, in account order.
func (v *Views) viewableTransactions(ctx context.Context, isin string) ([]ledger.Transaction, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	var out []ledger.Transaction
	for i := range accounts {
		txs, txErr := v.ledger.CurrentTransactions(ctx, accounts[i].ID)
		if txErr != nil {
			return nil, accountErr("loading transactions", txErr)
		}
		out = append(out, views.InstrumentTransactions(txs, isin)...)
	}
	return out, nil
}

// Instrument returns the instrument with ISIN isin (D-195): its listings,
// those the caller's viewable accounts post to and those priced, each with
// its price source and latest price, its asset class, and its exemption
// class in each wrapper whose yearly tax tells assets apart (D-127, D-132).
// An ISIN no viewable account posts to and with no stored price is
// ErrUnknownInstrument.
func (v *Views) Instrument(ctx context.Context, isin string) (Instrument, error) {
	listings, latest, err := v.instrumentListings(ctx, isin)
	if err != nil {
		return Instrument{}, err
	}
	out := Instrument{ISIN: isin, AssetClass: views.AssetClassOf(listings[0])}
	var priced valuation.Price
	for _, in := range listings {
		l, listErr := v.listing(ctx, in, latest)
		if listErr != nil {
			return Instrument{}, listErr
		}
		if l.Latest != nil {
			priced = *l.Latest
		}
		out.Listings = append(out.Listings, l)
	}
	if out.Classes, err = v.instrumentClasses(ctx, listings[0], &priced); err != nil {
		return Instrument{}, err
	}
	return out, nil
}

// instrumentListings returns the listings of isin that the caller's
// viewable accounts post to, then those only priced, with the latest price
// of each listing from each source; ErrUnknownInstrument when there are
// none.
func (v *Views) instrumentListings(ctx context.Context, isin string) ([]ledger.Instrument, []valuation.Price, error) {
	txs, err := v.viewableTransactions(ctx, isin)
	if err != nil {
		return nil, nil, err
	}
	listings := views.InstrumentListings(txs, isin)
	latest, err := v.latestPrices(ctx, isin)
	if err != nil {
		return nil, nil, err
	}
	for i := range latest {
		if in, ok := views.ListingOf(latest[i].Instrument); ok && !slices.Contains(listings, in) {
			listings = append(listings, in)
		}
	}
	if len(listings) == 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnknownInstrument, isin)
	}
	return listings, latest, nil
}

// listing returns in with the source its prices come from and that
// source's latest price among latest.
func (v *Views) listing(ctx context.Context, in ledger.Instrument, latest []valuation.Price) (Listing, error) {
	src, err := v.sourceOf(ctx, in)
	if err != nil {
		return Listing{}, err
	}
	l := Listing{Instrument: in, Source: src}
	for i := range latest {
		if p := latest[i]; src != "" && p.Instrument == in.String() && p.Source == src {
			l.Latest = &p
		}
	}
	return l, nil
}

// latestPrices returns, for each listing of isin and each source, the
// latest stored price; none without the port.
func (v *Views) latestPrices(ctx context.Context, isin string) ([]valuation.Price, error) {
	if v.prices == nil {
		return nil, nil
	}
	out, err := v.prices.LatestPrices(ctx, isin)
	if err != nil {
		return nil, fmt.Errorf("loading the latest prices of %s: %w", isin, err)
	}
	return out, nil
}

// sourceOf returns the source in's prices come from; "" when none.
func (v *Views) sourceOf(ctx context.Context, in ledger.Instrument) (string, error) {
	if v.prices == nil {
		return "", nil
	}
	src, found, err := v.prices.SourceOf(ctx, valuation.PriceSeries(in.String()))
	if err != nil {
		return "", fmt.Errorf("loading the price source of %s: %w", in, err)
	}
	if !found {
		return "", nil
	}
	return src.Source, nil
}

// instrumentClasses returns in's exemption classes; none without the pack
// or the owner's classes.
func (v *Views) instrumentClasses(ctx context.Context, in ledger.Instrument, priced *valuation.Price) ([]views.InstrumentClass, error) {
	if v.pack == nil || v.terms == nil {
		return nil, nil
	}
	owner, err := v.terms.ExemptionClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading exemption classes: %w", err)
	}
	asset := views.AssetOf(in, priced)
	return views.InstrumentClasses(v.pack, &asset, owner), nil
}

// InstrumentPrices returns isin's stored prices in the range from through
// to, every source's, each marked when its source is the one its listing is
// valued at (D-94, D-195).
func (v *Views) InstrumentPrices(ctx context.Context, isin string, from, to kernel.Date) ([]views.InstrumentPrice, error) {
	if err := views.ValidateRange(from, to); err != nil {
		return nil, fmt.Errorf("price history: %w", err)
	}
	if v.prices == nil {
		return nil, nil
	}
	prices, err := v.prices.PriceHistory(ctx, isin, from, to)
	if err != nil {
		return nil, fmt.Errorf("loading prices of %s: %w", isin, err)
	}
	sources := map[string]string{}
	for i := range prices {
		if _, seen := sources[prices[i].Instrument]; seen {
			continue
		}
		src, ok, srcErr := v.prices.SourceOf(ctx, valuation.PriceSeries(prices[i].Instrument))
		if srcErr != nil {
			return nil, fmt.Errorf("loading the price source of %s: %w", prices[i].Instrument, srcErr)
		}
		sources[prices[i].Instrument] = ""
		if ok {
			sources[prices[i].Instrument] = src.Source
		}
	}
	return views.MarkPrimary(prices, sources), nil
}

// InstrumentHoldings returns the holdings of isin at the end of on in every
// account the caller may view, each with its open lots, in account then
// symbol order (D-195).
func (v *Views) InstrumentHoldings(ctx context.Context, isin string, on kernel.Date) ([]views.InstrumentHolding, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	var out []views.InstrumentHolding
	for i := range accounts {
		id := accounts[i].ID
		holdings, hErr := v.ledger.HoldingsOn(ctx, id, on)
		if hErr != nil {
			return nil, accountErr("loading holdings", hErr)
		}
		lots, lErr := v.ledger.LotsOn(ctx, id, on)
		if lErr != nil {
			return nil, accountErr("loading lots", lErr)
		}
		out = append(out, views.HoldingsOf(id, holdings, lots, isin)...)
	}
	return out, nil
}

// InstrumentTransactions returns one page of the current transactions that
// post to isin in the accounts the caller may view, in date then key
// order, after the cursor after, or from the first when after is nil, each
// with its origin (D-58, D-195).
func (v *Views) InstrumentTransactions(ctx context.Context, isin string, after *views.TransactionCursor, size int) (views.TransactionPage, error) {
	txs, err := v.viewableTransactions(ctx, isin)
	if err != nil {
		return views.TransactionPage{}, err
	}
	page, err := views.PageTransactions(txs, after, size)
	if err != nil {
		return views.TransactionPage{}, fmt.Errorf("paging transactions: %w", err)
	}
	page.Origins = v.installed(ctx).Origins(page.Transactions)
	return page, nil
}
