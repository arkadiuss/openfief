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
// owner's confirmations, which hold for every instance of the same
// institution and which the owner can correct (D-153), and the ticker
// entries and holdings (D-155) waiting for one, or booked with one.
type TickerStore interface {
	// ConfirmedTickers returns the listing the owner's current
	// confirmation settled for each ticker symbol of institution
	// (ingest.ListingOf, D-159, D-170).
	ConfirmedTickers(ctx context.Context, institution string) (map[string]ingest.Listing, error)
	// HoldTickers records the tickers an instance reported, keeping an
	// earlier name, and its rows: those waiting for their ticker's ISIN and
	// those booked, each once by key, or by account, ticker and date for an
	// assertion: holding one again replaces it. It returns the tickers it
	// hadn't seen before.
	HoldTickers(ctx context.Context, h *ingest.TickerHold) ([]ingest.Ticker, error)
	// SaveProposals records the ISIN proposed for tickers of instance, with
	// the currency its listing trades in, replacing an earlier proposal.
	SaveProposals(ctx context.Context, instance string, proposals []ingest.ISINProposal) error
	// PendingTickers returns every ticker its institution has no
	// confirmation of, by instance and symbol.
	PendingTickers(ctx context.Context) ([]ingest.PendingTicker, error)
	// ConfirmedTickerList returns every ticker confirmed at an institution,
	// by institution and symbol.
	ConfirmedTickerList(ctx context.Context) ([]ingest.ConfirmedTicker, error)
	// TickerAccounts returns the accounts of the rows of instance's ticker
	// at every instance of its institution: those waiting, or all with all.
	// A ticker never reported is ErrUnknownTicker.
	TickerAccounts(ctx context.Context, instance, symbol string, all bool) ([]ledger.AccountID, error)
	// ConfirmTicker records the owner's confirmation for the ticker's
	// institution and returns what it books: the listing it settles
	// (ingest.ConfirmListing) and the rows still waiting for the ticker at
	// every instance of the institution. A ticker never reported is
	// ErrUnknownTicker; one confirmed with another ISIN or trading currency
	// ErrTickerConfirmed; one whose listing names no currency on an
	// exchange that quotes in more than one, confirmed without picking one
	// of them, an *ingest.TradingCurrencyChoiceError that records nothing
	// (D-170). Confirming the same ISIN again changes nothing.
	ConfirmTicker(ctx context.Context, c ingest.TickerConfirmation) (ingest.TickerBooking, error)
	// CorrectTicker records the owner's correction of a confirmed ticker's
	// ISIN, or trading currency, for its institution, keeping the earlier
	// confirmation, and returns what it books: the listing it settles, the
	// one it replaces, and every row of the ticker at every instance of the
	// institution (D-153). A ticker never confirmed is
	// ingest.ErrTickerNotConfirmed; a wrong pick an
	// *ingest.TradingCurrencyChoiceError that records nothing.
	CorrectTicker(ctx context.Context, c ingest.TickerConfirmation) (ingest.TickerBooking, error)
	// BookedTickerRows marks rows booked, once the ledger holds them.
	BookedTickerRows(ctx context.Context, rows []ingest.TickerRows) error
}

// AssertionWithdrawer withdraws balance assertions that no longer hold,
// such as those of the instrument a ticker's ISIN was corrected from
// (D-153): their checks are removed, so they are no longer flagged.
type AssertionWithdrawer interface {
	WithdrawAssertions(ctx context.Context, assertions []ledger.BalanceAssertion) error
}

// InstrumentLookup asks a market-data connector instance for the ISIN of
// tickers (D-116). One that can't look instruments up wraps ErrNoEntries.
type InstrumentLookup interface {
	LookupInstruments(ctx context.Context, instance string, tickers []ingest.Ticker) ([]ingest.ISINProposal, error)
}

// bookTickers books m's ticker entries and holdings (D-155) whose ticker's
// ISIN the owner confirmed for institution (D-153), and holds the rest,
// each entry an ingest problem of the batch, until the owner confirms it
// (D-116); what it books is kept too, for a correction. Tickers reported
// for the first time and not confirmed are looked up by the market-data
// connectors, whose proposals are kept for the owner; a lookup that fails
// proposes nothing.
func (c *Connectors) bookTickers(ctx context.Context, instance string, institution ingest.InstitutionID, m *ingest.Mapping) (ingest.Mapping, error) {
	if len(m.TickerEntries) == 0 && len(m.TickerAssertions) == 0 && len(m.Tickers) == 0 {
		return *m, nil
	}
	if c.Tickers == nil {
		return ingest.Mapping{}, ErrNoTickerStore
	}
	confirmed, err := c.Tickers.ConfirmedTickers(ctx, string(institution))
	if err != nil {
		return ingest.Mapping{}, fmt.Errorf("confirmed tickers: %w", err)
	}
	out, waiting := ingest.BookTickerEntries(m, confirmed)
	out, waitingAssertions := ingest.BookTickerAssertions(&out, confirmed)
	added, err := c.Tickers.HoldTickers(ctx, &ingest.TickerHold{
		Institution: string(institution), Tickers: m.Tickers,
		Waiting: ingest.TickerRows{Instance: instance, Entries: waiting, Assertions: waitingAssertions},
		Booked:  ingest.BookedTickerRows(instance, m, confirmed),
	})
	if err != nil {
		return ingest.Mapping{}, fmt.Errorf("holding ticker entries: %w", err)
	}
	added = slices.DeleteFunc(added, func(t ingest.Ticker) bool { _, ok := confirmed[t.Symbol]; return ok })
	if err := c.proposeISINs(ctx, instance, added); err != nil {
		return ingest.Mapping{}, err
	}
	return out, nil
}

// proposeISINs asks each market-data connector instance, the running ones
// whose manifest declares it and whose grant approves that (D-123), in
// name order, for the ISINs of tickers of instance
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

// marketDataInstances returns the running market-data connector
// instances, in name order (marketData).
func (c *Connectors) marketDataInstances(ctx context.Context) ([]string, error) {
	all, err := c.Instances.ConnectorInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("connector instances: %w", err)
	}
	var out []string
	for _, inst := range all {
		ok, err := c.marketData(ctx, inst)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, inst)
		}
	}
	slices.Sort(out)
	return out, nil
}

// marketData reports whether instance is a market-data connector: its
// running version's manifest declares it and its grant approves that
// (D-123). Without Declarations no instance is one.
func (c *Connectors) marketData(ctx context.Context, instance string) (bool, error) {
	if c.Declarations == nil {
		return false, nil
	}
	d, err := c.Declarations.Declarations(ctx, instance)
	if err != nil {
		return false, fmt.Errorf("manifest declarations: %w", err)
	}
	return d.MarketData, nil
}

// Tickers is the owner's service for the tickers that connectors report
// without an ISIN (D-116, D-79): it lists those waiting for the owner, with
// the market-data connector's proposal, and those confirmed, takes the
// owner's confirmation of an ISIN, proposed or typed, which holds for every
// account at the institution and books the rows waiting for it, and the
// owner's correction of a confirmed one, which books every row of the
// ticker anew (D-153).
type Tickers struct {
	Owners Owners
	Store  TickerStore
	Ledger LedgerImporter
	// Assertions withdraws the holdings asserted under the ISIN a
	// correction replaces; without it a correction to another ISIN fails.
	Assertions AssertionWithdrawer
}

// ErrNoAssertionWithdrawer reports a correction to another ISIN by a
// Tickers that can't withdraw the holdings asserted under the old one.
var ErrNoAssertionWithdrawer = errors.New("no way to withdraw balance assertions")

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

// Confirmed returns, to the owner, every ticker whose ISIN is confirmed,
// once per institution (D-153).
func (t *Tickers) Confirmed(ctx context.Context, origin, sessionToken string) ([]ingest.ConfirmedTicker, error) {
	if err := t.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return nil, fmt.Errorf("listing confirmed tickers: %w", err)
	}
	out, err := t.Store.ConfirmedTickerList(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing confirmed tickers: %w", err)
	}
	return out, nil
}

// Confirm records, for the owner, isin as the ISIN of instance's ticker
// symbol, for every account at its institution (D-153), then books every
// row waiting for it into the ledger: entries as transactions, their keys
// scoped to their instance as a delivery's are, and holdings as balance
// assertions (D-155). It asks first whether the owner may change every
// account a waiting row is in (D-163). Waiting opening lots join their
// account's opening on its go-live date (book, D-160), their cost in the
// trading currency of the listing the confirmation settles (D-159), or the
// one the owner picks when the listing names none and its exchange quotes
// in more than one: until the owner picks one, the confirmation is refused
// and they keep waiting (D-170). Confirming again books whatever still
// waits.
func (t *Tickers) Confirm(ctx context.Context, origin, sessionToken string, c ingest.TickerConfirmation) (ingest.ImportSummary, error) {
	if err := t.authorize(ctx, origin, sessionToken, &c, false); err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("confirming a ticker: %w", err)
	}
	b, err := t.Store.ConfirmTicker(ctx, c)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("confirming %s of %s: %w", c.Symbol, c.Instance, err)
	}
	sum, err := t.bookRows(ctx, &b)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("booking %s of %s: %w", c.Symbol, c.Instance, err)
	}
	return sum, nil
}

// Correct records, for the owner, isin, with the trading currency the
// owner picked when its listing names none (D-170), as the corrected ISIN
// of instance's ticker symbol, for every account at its institution
// (D-153), then books every row of the ticker anew with it: its
// transactions are new versions of their keys, which supersede those
// booked under the old ISIN and keep them (D-18), and its holdings are
// asserted of the corrected instrument, those of the old one withdrawn. It
// asks first whether the owner may change every account a row of the
// ticker is in (D-163). A ticker never confirmed is
// ingest.ErrTickerNotConfirmed.
func (t *Tickers) Correct(ctx context.Context, origin, sessionToken string, c ingest.TickerConfirmation) (ingest.ImportSummary, error) {
	if err := t.authorize(ctx, origin, sessionToken, &c, true); err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("correcting a ticker: %w", err)
	}
	b, err := t.Store.CorrectTicker(ctx, c)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("correcting %s of %s: %w", c.Symbol, c.Instance, err)
	}
	sum, err := t.bookRows(ctx, &b)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("booking %s of %s anew: %w", c.Symbol, c.Instance, err)
	}
	return sum, nil
}

// authorize checks that the caller is the owner and may change every
// account of the rows of c's ticker (D-163): those waiting, or all with
// all.
func (t *Tickers) authorize(ctx context.Context, origin, sessionToken string, c *ingest.TickerConfirmation, all bool) error {
	if err := t.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("checking the owner: %w", err)
	}
	accounts, err := t.Store.TickerAccounts(ctx, c.Instance, c.Symbol, all)
	if err != nil {
		return fmt.Errorf("the ticker's accounts: %w", err)
	}
	return mayChange(ctx, t.Owners, origin, sessionToken, accounts...)
}

// bookRows books what b books (ingest.TickerBooking.Ledger) into the
// ledger, withdraws the holdings it replaces, and marks its rows booked.
func (t *Tickers) bookRows(ctx context.Context, b *ingest.TickerBooking) (ingest.ImportSummary, error) {
	l, err := b.Ledger()
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("the ticker's rows: %w", err)
	}
	if len(l.Transactions) == 0 && len(l.Asserted) == 0 {
		return ingest.ImportSummary{}, nil
	}
	if len(l.Replaced) > 0 && t.Assertions == nil {
		return ingest.ImportSummary{}, ErrNoAssertionWithdrawer
	}
	sum, err := t.book(ctx, l.Transactions, l.Asserted)
	if err != nil {
		return ingest.ImportSummary{}, err
	}
	if len(l.Replaced) > 0 {
		if err := t.Assertions.WithdrawAssertions(ctx, l.Replaced); err != nil {
			return ingest.ImportSummary{}, fmt.Errorf("withdrawing the old holdings: %w", err)
		}
	}
	if err := t.Store.BookedTickerRows(ctx, b.Rows); err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("marking the rows booked: %w", err)
	}
	return sum, nil
}

// book imports a ticker's transactions and holdings: those that hold an
// opening lot, which an account seeded from an xStation export's open
// positions waits for (D-136), are added to its seeding one at a time, on
// its go-live date (LedgerImporter.AddOpeningLots, the path hand-entered
// bond lots take today); any other are imported as they are.
func (t *Tickers) book(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	if slices.ContainsFunc(txs, func(tx ledger.Transaction) bool { return tx.Kind == ledger.KindOpeningLot }) {
		sum, err := t.Ledger.AddOpeningLots(ctx, txs, assertions)
		if err != nil {
			return ingest.ImportSummary{}, fmt.Errorf("adding opening lots: %w", err)
		}
		return sum, nil
	}
	sum, err := t.Ledger.Import(ctx, txs, assertions)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("importing: %w", err)
	}
	return sum, nil
}
