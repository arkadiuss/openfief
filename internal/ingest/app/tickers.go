package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Errors of the owner's ticker confirmations.
var (
	// ErrUnknownTicker reports a confirmation of a ticker no connector
	// instance reported.
	ErrUnknownTicker = errors.New("no such ticker")
	// ErrTickerConfirmed reports a confirmation of another ISIN for a
	// ticker already confirmed.
	ErrTickerConfirmed = errors.New("the ticker's ISIN is already confirmed")
	// ErrNoTickerStore reports a delivery with ticker entries to a
	// Connectors that has nowhere to keep them.
	ErrNoTickerStore = errors.New("no store for ticker entries")
)

// TickerStore keeps what D-116 needs: the tickers connector instances
// reported without an ISIN, the market-data connector's proposals, the
// owner's confirmations, and the ticker entries waiting for one.
type TickerStore interface {
	// ConfirmedTickers returns the listing the owner's confirmation settled
	// for each ticker symbol of instance (ingest.ListingOf, D-159).
	ConfirmedTickers(ctx context.Context, instance string) (map[string]ingest.Listing, error)
	// HoldTickers records the tickers instance reported, keeping an earlier
	// name, and the entries waiting for their ticker's ISIN, each once by
	// key: holding a key again replaces its entry. It returns the tickers it
	// hadn't seen before.
	HoldTickers(ctx context.Context, instance string, tickers []ingest.Ticker, waiting []ingest.TickerEntry) ([]ingest.Ticker, error)
	// SaveProposals records the ISIN proposed for tickers of instance, with
	// the currency its listing trades in, replacing an earlier proposal.
	SaveProposals(ctx context.Context, instance string, proposals []ingest.ISINProposal) error
	// PendingTickers returns every ticker without a confirmation, by
	// instance and symbol.
	PendingTickers(ctx context.Context) ([]ingest.PendingTicker, error)
	// ConfirmTicker records the owner's confirmation and returns the
	// entries still waiting for the ticker, with the listing the
	// confirmation settles (ingest.ListingOf). A ticker never reported is
	// ErrUnknownTicker; one confirmed with another ISIN ErrTickerConfirmed.
	// Confirming the same ISIN again changes nothing.
	ConfirmTicker(ctx context.Context, c ingest.TickerConfirmation) ([]ingest.TickerEntry, ingest.Listing, error)
	// BookedTickerEntries drops the waiting entries of instance with keys,
	// once the ledger holds them.
	BookedTickerEntries(ctx context.Context, instance string, keys []ledger.TransactionKey) error
}

// InstrumentLookup asks a market-data connector instance for the ISIN of
// tickers (D-116). One that can't look instruments up wraps ErrNoEntries.
type InstrumentLookup interface {
	LookupInstruments(ctx context.Context, instance string, tickers []ingest.Ticker) ([]ingest.ISINProposal, error)
}

// bookTickers books m's ticker entries whose ticker's ISIN the owner
// confirmed, and holds the rest, each an ingest problem of the batch,
// until the owner confirms it (D-116). Tickers reported for the first time
// are looked up by the market-data connectors, whose proposals are kept
// for the owner; a lookup that fails proposes nothing.
func (c *Connectors) bookTickers(ctx context.Context, instance string, m *ingest.Mapping) (ingest.Mapping, error) {
	if len(m.TickerEntries) == 0 && len(m.Tickers) == 0 {
		return *m, nil
	}
	if c.Tickers == nil {
		return ingest.Mapping{}, ErrNoTickerStore
	}
	confirmed, err := c.Tickers.ConfirmedTickers(ctx, instance)
	if err != nil {
		return ingest.Mapping{}, fmt.Errorf("confirmed tickers: %w", err)
	}
	out, waiting := ingest.BookTickerEntries(m, confirmed)
	var unconfirmed []ingest.Ticker
	for _, t := range m.Tickers {
		if _, ok := confirmed[t.Symbol]; !ok {
			unconfirmed = append(unconfirmed, t)
		}
	}
	added, err := c.Tickers.HoldTickers(ctx, instance, unconfirmed, waiting)
	if err != nil {
		return ingest.Mapping{}, fmt.Errorf("holding ticker entries: %w", err)
	}
	if err := c.proposeISINs(ctx, instance, added); err != nil {
		return ingest.Mapping{}, err
	}
	return out, nil
}

// proposeISINs asks each market-data connector instance, the running ones
// bound to no account, in name order, for the ISINs of tickers of instance
// it hasn't been proposed yet, and keeps the first proposal for each. An
// instance that can't look instruments up or fails is passed over.
func (c *Connectors) proposeISINs(ctx context.Context, instance string, tickers []ingest.Ticker) error {
	if len(tickers) == 0 || c.Lookup == nil || c.Instances == nil {
		return nil
	}
	sources, err := c.marketDataInstances(ctx)
	if err != nil {
		return err
	}
	open := slices.Clone(tickers)
	var proposals []ingest.ISINProposal
	for _, source := range sources {
		if len(open) == 0 {
			break
		}
		found, err := c.Lookup.LookupInstruments(ctx, source, open)
		if err != nil {
			continue
		}
		proposals = append(proposals, found...)
		open = slices.DeleteFunc(open, func(t ingest.Ticker) bool {
			return slices.ContainsFunc(found, func(p ingest.ISINProposal) bool { return p.Symbol == t.Symbol })
		})
	}
	if len(proposals) == 0 {
		return nil
	}
	if err := c.Tickers.SaveProposals(ctx, instance, proposals); err != nil {
		return fmt.Errorf("saving ISIN proposals: %w", err)
	}
	return nil
}

// marketDataInstances returns the running connector instances bound to no
// account, in name order: those that can deliver market data only.
func (c *Connectors) marketDataInstances(ctx context.Context) ([]string, error) {
	all, err := c.Instances.ConnectorInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("connector instances: %w", err)
	}
	var out []string
	for _, inst := range all {
		bound, err := c.Bindings.BoundAccounts(ctx, inst)
		if err != nil {
			return nil, fmt.Errorf("connector bindings: %w", err)
		}
		if len(bound) == 0 {
			out = append(out, inst)
		}
	}
	slices.Sort(out)
	return out, nil
}

// Tickers is the owner's service for the tickers that connectors report
// without an ISIN (D-116, D-79): it lists those waiting for the owner, with
// the market-data connector's proposal, and takes the owner's confirmation
// of an ISIN, proposed or typed, after which the entries waiting for it are
// booked.
type Tickers struct {
	Owners Owners
	Store  TickerStore
	Ledger LedgerImporter
}

// Pending returns, to the owner, every ticker waiting for its ISIN.
func (t *Tickers) Pending(ctx context.Context, origin, sessionToken string) ([]ingest.PendingTicker, error) {
	if err := t.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return nil, fmt.Errorf("listing tickers: %w", err)
	}
	out, err := t.Store.PendingTickers(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing tickers: %w", err)
	}
	return out, nil
}

// Confirm records, for the owner, isin as the ISIN of instance's ticker
// symbol, then books every entry waiting for it into the ledger, its keys
// scoped to instance as a delivery's are, and returns what was imported.
// Waiting opening lots join their account's opening on its go-live date
// (book, D-160), their cost in the trading currency of the listing the
// confirmation settles (D-159). Confirming again books whatever still
// waits.
func (t *Tickers) Confirm(ctx context.Context, origin, sessionToken string, c ingest.TickerConfirmation) (ingest.ImportSummary, error) {
	if err := t.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("confirming a ticker: %w", err)
	}
	waiting, listing, err := t.Store.ConfirmTicker(ctx, c)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("confirming %s of %s: %w", c.Symbol, c.Instance, err)
	}
	if len(waiting) == 0 {
		return ingest.ImportSummary{}, nil
	}
	txs := make([]ledger.Transaction, 0, len(waiting))
	keys := make([]ledger.TransactionKey, 0, len(waiting))
	for i := range waiting {
		tx, txErr := waiting[i].Transaction(listing)
		if txErr != nil {
			return ingest.ImportSummary{}, fmt.Errorf("booking %s: %w", waiting[i].Header.Key, txErr)
		}
		txs = append(txs, tx)
		keys = append(keys, waiting[i].Header.Key)
	}
	sum, err := t.book(ctx, ingest.ScopeToConnector(c.Instance, txs))
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("booking %s of %s: %w", c.Symbol, c.Instance, err)
	}
	if err := t.Store.BookedTickerEntries(ctx, c.Instance, keys); err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("booking %s of %s: %w", c.Symbol, c.Instance, err)
	}
	return sum, nil
}

// book imports a confirmed ticker's transactions: those that hold an
// opening lot, which an account seeded from an xStation export's open
// positions waits for (D-136), are added to its seeding one at a time, on
// its go-live date (LedgerImporter.AddOpeningLots, the path hand-entered
// bond lots take today); any other are imported as they are.
func (t *Tickers) book(ctx context.Context, txs []ledger.Transaction) (ingest.ImportSummary, error) {
	if slices.ContainsFunc(txs, func(tx ledger.Transaction) bool { return tx.Kind == ledger.KindOpeningLot }) {
		sum, err := t.Ledger.AddOpeningLots(ctx, txs, nil)
		if err != nil {
			return ingest.ImportSummary{}, fmt.Errorf("adding opening lots: %w", err)
		}
		return sum, nil
	}
	sum, err := t.Ledger.Import(ctx, txs, nil)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("importing: %w", err)
	}
	return sum, nil
}
