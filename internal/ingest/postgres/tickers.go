package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

var _ app.TickerStore = (*Store)(nil)

// errStoredEntry reports a stored ticker entry or assertion that doesn't
// read back.
var errStoredEntry = errors.New("stored ticker entry")

// confirmation is one version of the owner's confirmation of a ticker at an
// institution (D-116, D-153), with what its listing needs.
type confirmation struct {
	version                         int
	isin, country, proposalCurrency string
	picked                          string
}

// selectConfirmation reads a confirmation's columns; callers add FROM and
// WHERE.
const selectConfirmation = `SELECT version, isin, country, proposal_currency, picked`

// dest returns c's scan destinations, in selectConfirmation's order.
func (c *confirmation) dest() []any {
	return []any{&c.version, &c.isin, &c.country, &c.proposalCurrency, &c.picked}
}

// parts returns the ticker of symbol, the proposal and the ISIN c settles
// its listing from (ingest.ListingOf).
func (c *confirmation) parts(symbol string) (ingest.Ticker, ingest.ISINProposal, ledger.ISIN, error) {
	isin, err := ledger.ParseISIN(c.isin)
	if err != nil {
		return ingest.Ticker{}, ingest.ISINProposal{}, ledger.ISIN{}, fmt.Errorf("%w: ticker %s: %w", errStoredEntry, symbol, err)
	}
	proposal := ingest.ISINProposal{Symbol: symbol, ISIN: isin, Currency: c.proposalCurrency}
	return ingest.Ticker{Symbol: symbol, Country: c.country}, proposal, isin, nil
}

// listing returns the listing c settles.
func (c *confirmation) listing(symbol string) (ingest.Listing, error) {
	t, p, isin, err := c.parts(symbol)
	if err != nil {
		return ingest.Listing{}, err
	}
	return ingest.ListingOf(&t, &p, isin, c.picked), nil
}

// code returns the trading currency code of the listing c settles.
func (c *confirmation) code(symbol string) (string, error) {
	t, p, isin, err := c.parts(symbol)
	if err != nil {
		return "", err
	}
	return ingest.ListingCode(&t, &p, isin, c.picked), nil
}

// tickerRow is the ticker row a confirmation or correction names.
type tickerRow struct {
	institution, country, proposed, proposedCurrency string
}

// loadTicker returns instance's ticker symbol, locked, or
// app.ErrUnknownTicker.
func loadTicker(ctx context.Context, tx pgx.Tx, instance, symbol string) (tickerRow, error) {
	var r tickerRow
	err := tx.QueryRow(ctx, `SELECT institution, country, coalesce(proposed_isin, ''), proposed_currency
		FROM ingest.tickers WHERE instance = $1 AND symbol = $2 FOR UPDATE`, instance, symbol).
		Scan(&r.institution, &r.country, &r.proposed, &r.proposedCurrency)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return tickerRow{}, app.ErrUnknownTicker
	case err != nil:
		return tickerRow{}, fmt.Errorf("loading the ticker: %w", err)
	}
	return r, nil
}

// confirmationOf returns what the owner's confirmation of isin, picking
// picked, for r's ticker symbol records: the listing's parts, with the
// proposal's trading currency only when the owner confirmed its ISIN.
func (r *tickerRow) confirmationOf(symbol string, isin ledger.ISIN, picked string) (confirmation, error) {
	c := confirmation{isin: isin.String(), country: r.country, picked: picked}
	if r.proposed == isin.String() {
		if _, _, ok := ingest.ParseTradingCurrency(r.proposedCurrency); ok {
			c.proposalCurrency = r.proposedCurrency
		}
	}
	t := ingest.Ticker{Symbol: symbol, Country: r.country}
	p := ingest.ISINProposal{Symbol: symbol, Currency: c.proposalCurrency}
	if c.proposalCurrency != "" {
		p.ISIN = isin
	}
	if _, err := ingest.ConfirmListing(&t, &p, isin, picked); err != nil {
		return confirmation{}, fmt.Errorf("confirming the ticker: %w", err)
	}
	return c, nil
}

// current returns the current confirmation of symbol at institution, locked,
// and whether there is one.
func current(ctx context.Context, tx pgx.Tx, institution, symbol string) (confirmation, bool, error) {
	var c confirmation
	err := tx.QueryRow(ctx, selectConfirmation+` FROM ingest.ticker_confirmations
		WHERE institution = $1 AND symbol = $2 ORDER BY version DESC LIMIT 1 FOR UPDATE`, institution, symbol).Scan(c.dest()...)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return confirmation{}, false, nil
	case err != nil:
		return confirmation{}, false, fmt.Errorf("loading the confirmation: %w", err)
	}
	return c, true, nil
}

// record stores c as version of symbol's confirmation at institution.
func record(ctx context.Context, tx pgx.Tx, institution, symbol string, c *confirmation) error {
	_, err := tx.Exec(ctx, `INSERT INTO ingest.ticker_confirmations
			(institution, symbol, version, isin, country, proposal_currency, picked)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		institution, symbol, c.version, c.isin, c.country, c.proposalCurrency, c.picked)
	if err != nil {
		return fmt.Errorf("recording the confirmation: %w", err)
	}
	return nil
}

// ConfirmedTickers returns the listing the owner's current confirmation
// settled for each ticker symbol of institution (D-116, D-153, D-159).
func (s *Store) ConfirmedTickers(ctx context.Context, institution string) (map[string]ingest.Listing, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT ON (symbol) symbol, version, isin, country, proposal_currency, picked
		FROM ingest.ticker_confirmations WHERE institution = $1 ORDER BY symbol, version DESC`, institution)
	if err != nil {
		return nil, fmt.Errorf("loading the confirmed tickers of %s: %w", institution, err)
	}
	out := map[string]ingest.Listing{}
	var (
		symbol string
		c      confirmation
	)
	_, err = pgx.ForEachRow(rows, append([]any{&symbol}, c.dest()...), func() error {
		l, listingErr := c.listing(symbol)
		if listingErr != nil {
			return listingErr
		}
		out[symbol] = l
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading the confirmed tickers of %s: %w", institution, err)
	}
	return out, nil
}

// HoldTickers records, in one transaction, the tickers h names for its
// instance at its institution, keeping an earlier name, and its entries
// and assertions: those waiting for their ticker's ISIN, and those booked,
// kept for a correction (D-153). Holding a key again replaces its entry;
// an assertion of the same account, ticker and date replaces the earlier
// one. It returns the tickers it hadn't seen before.
func (s *Store) HoldTickers(ctx context.Context, h *ingest.TickerHold) ([]ingest.Ticker, error) {
	instance := h.Waiting.Instance
	if instance == "" {
		instance = h.Booked.Instance
	}
	var added []ingest.Ticker
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var err error
		if added, err = holdTickerRows(ctx, tx, instance, h); err != nil {
			return err
		}
		if waitErr := holdRows(ctx, tx, instance, &h.Waiting, false); waitErr != nil {
			return waitErr
		}
		return holdRows(ctx, tx, instance, &h.Booked, true)
	})
	if err != nil {
		return nil, fmt.Errorf("holding the ticker entries of %s: %w", instance, err)
	}
	return added, nil
}

// holdTickerRows stores h's tickers of instance at h's institution,
// keeping an earlier one, and returns those it hadn't seen.
func holdTickerRows(ctx context.Context, tx pgx.Tx, instance string, h *ingest.TickerHold) ([]ingest.Ticker, error) {
	var added []ingest.Ticker
	for _, t := range h.Tickers {
		tag, err := tx.Exec(ctx, `INSERT INTO ingest.tickers (instance, symbol, code, country, name, institution)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (instance, symbol) DO NOTHING`,
			instance, t.Symbol, t.Code, t.Country, t.Name, h.Institution)
		if err != nil {
			return nil, fmt.Errorf("storing ticker %s: %w", t.Symbol, err)
		}
		if tag.RowsAffected() == 1 {
			added = append(added, t)
		}
	}
	return added, nil
}

// holdRows stores r's entries and assertions of instance, as booked or
// waiting.
func holdRows(ctx context.Context, tx pgx.Tx, instance string, r *ingest.TickerRows, booked bool) error {
	for i := range r.Entries {
		if err := holdEntry(ctx, tx, instance, &r.Entries[i], booked); err != nil {
			return err
		}
	}
	for i := range r.Assertions {
		if err := holdAssertion(ctx, tx, instance, &r.Assertions[i], booked); err != nil {
			return err
		}
	}
	return nil
}

// holdEntry stores one entry, replacing an earlier one of its key; once
// booked, it stays booked.
func holdEntry(ctx context.Context, tx pgx.Tx, instance string, e *ingest.TickerEntry, booked bool) error {
	charge, chargeCurrency := "", ""
	if !e.Charge.IsZero() {
		charge, chargeCurrency = e.Charge.Amount().String(), e.Charge.Currency().String()
	}
	var opened, openPrice *string
	if !e.Opened.IsZero() {
		day := e.Opened.String()
		opened = &day
	}
	if e.PricedByListing() {
		price := e.OpenPrice.String()
		openPrice = &price
	}
	_, err := tx.Exec(ctx, `INSERT INTO ingest.ticker_entries (instance, key_source, external_id, symbol, account_id, kind,
			date, quantity, amount, currency, charge, charge_currency, description, ref, lot_opened, open_price, booked)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (instance, key_source, external_id) DO UPDATE SET symbol = EXCLUDED.symbol,
			account_id = EXCLUDED.account_id, kind = EXCLUDED.kind, date = EXCLUDED.date, quantity = EXCLUDED.quantity,
			amount = EXCLUDED.amount, currency = EXCLUDED.currency, charge = EXCLUDED.charge,
			charge_currency = EXCLUDED.charge_currency, description = EXCLUDED.description, ref = EXCLUDED.ref,
			lot_opened = EXCLUDED.lot_opened, open_price = EXCLUDED.open_price,
			booked = ingest.ticker_entries.booked OR EXCLUDED.booked`,
		instance, e.Header.Key.Source, e.Header.Key.ExternalID, e.Symbol, string(e.Header.Account), e.Kind.String(),
		e.Header.Date.String(), e.Quantity.String(), e.Amount.Amount().String(), e.Amount.Currency().String(),
		charge, chargeCurrency, e.Description, e.Ref, opened, openPrice, booked)
	if err != nil {
		return fmt.Errorf("holding %s: %w", e.Header.Key, err)
	}
	return nil
}

// holdAssertion stores one ticker assertion, replacing an earlier one of
// its account, ticker and date; once asserted, it stays asserted.
func holdAssertion(ctx context.Context, tx pgx.Tx, instance string, a *ingest.TickerAssertion, asserted bool) error {
	_, err := tx.Exec(ctx, `INSERT INTO ingest.ticker_assertions (instance, symbol, account_id, date, quantity, ref, asserted)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (instance, symbol, account_id, date) DO UPDATE SET quantity = EXCLUDED.quantity, ref = EXCLUDED.ref,
			asserted = ingest.ticker_assertions.asserted OR EXCLUDED.asserted`,
		instance, a.Symbol, string(a.Account), a.Date.String(), a.Quantity.String(), a.Ref, asserted)
	if err != nil {
		return fmt.Errorf("holding the holding of %s on %s: %w", a.Symbol, a.Date, err)
	}
	return nil
}

// SaveProposals records the ISIN proposed for tickers of instance, with the
// currency its listing trades in (D-159); a currency that is no three
// letters is kept as none.
func (s *Store) SaveProposals(ctx context.Context, instance string, proposals []ingest.ISINProposal) error {
	for _, p := range proposals {
		currency := p.Currency
		if _, _, ok := ingest.ParseTradingCurrency(currency); !ok {
			currency = ""
		}
		_, err := s.pool.Exec(ctx, `UPDATE ingest.tickers SET proposed_isin = $3, proposed_name = $4, proposed_currency = $5
			WHERE instance = $1 AND symbol = $2`, instance, p.Symbol, p.ISIN.String(), p.Name, currency)
		if err != nil {
			return fmt.Errorf("saving the proposal for %s of %s: %w", p.Symbol, instance, err)
		}
	}
	return nil
}

// PendingTickers returns every ticker its institution has no confirmation
// of, by instance and symbol, with its proposal and how many entries wait
// for it.
func (s *Store) PendingTickers(ctx context.Context) ([]ingest.PendingTicker, error) {
	rows, err := s.pool.Query(ctx, `SELECT t.instance, t.symbol, t.code, t.country, t.name, coalesce(t.proposed_isin, ''),
			t.proposed_name, t.proposed_currency,
			(SELECT count(*) FROM ingest.ticker_entries e WHERE e.instance = t.instance AND e.symbol = t.symbol AND NOT e.booked)
		FROM ingest.tickers t
		WHERE NOT EXISTS (SELECT 1 FROM ingest.ticker_confirmations c WHERE c.institution = t.institution AND c.symbol = t.symbol)
		ORDER BY t.instance COLLATE "C", t.symbol COLLATE "C"`)
	if err != nil {
		return nil, fmt.Errorf("loading the tickers to confirm: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanPending)
	if err != nil {
		return nil, fmt.Errorf("loading the tickers to confirm: %w", err)
	}
	return out, nil
}

// scanPending reads one row of PendingTickers.
func scanPending(row pgx.CollectableRow) (ingest.PendingTicker, error) {
	var (
		p        ingest.PendingTicker
		proposed string
		waiting  int64
	)
	if err := row.Scan(&p.Instance, &p.Ticker.Symbol, &p.Ticker.Code, &p.Ticker.Country, &p.Ticker.Name, &proposed,
		&p.Proposal.Name, &p.Proposal.Currency, &waiting); err != nil {
		return ingest.PendingTicker{}, fmt.Errorf("scanning a ticker: %w", err)
	}
	p.Waiting = int(waiting)
	if proposed != "" {
		isin, err := ledger.ParseISIN(proposed)
		if err != nil {
			return ingest.PendingTicker{}, fmt.Errorf("ticker %s: %w", p.Ticker.Symbol, err)
		}
		p.Proposal.Symbol, p.Proposal.ISIN = p.Ticker.Symbol, isin
	}
	return p, nil
}

// ConfirmedTickerList returns every ticker confirmed at an institution
// (D-153), by institution and symbol, with its current confirmation, the
// first of its institution's instances that reported it, by name, and how
// many of its entries are booked.
func (s *Store) ConfirmedTickerList(ctx context.Context) ([]ingest.ConfirmedTicker, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT ON (c.institution, c.symbol) c.institution, t.instance, c.symbol, t.code,
			t.name, c.version, c.isin, c.country, c.proposal_currency, c.picked,
			(SELECT count(*) FROM ingest.ticker_entries e JOIN ingest.tickers u ON u.instance = e.instance AND u.symbol = e.symbol
				WHERE u.institution = c.institution AND e.symbol = c.symbol AND e.booked)
		FROM ingest.ticker_confirmations c JOIN ingest.tickers t ON t.institution = c.institution AND t.symbol = c.symbol
		ORDER BY c.institution, c.symbol, c.version DESC, t.instance COLLATE "C"`)
	if err != nil {
		return nil, fmt.Errorf("loading the confirmed tickers: %w", err)
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ingest.ConfirmedTicker, error) {
		var (
			ct     ingest.ConfirmedTicker
			c      confirmation
			booked int64
		)
		if scanErr := row.Scan(&ct.Institution, &ct.Instance, &ct.Ticker.Symbol, &ct.Ticker.Code, &ct.Ticker.Name,
			&c.version, &c.isin, &c.country, &c.proposalCurrency, &c.picked, &booked); scanErr != nil {
			return ingest.ConfirmedTicker{}, fmt.Errorf("scanning a ticker: %w", scanErr)
		}
		ct.Ticker.Country, ct.Booked = c.country, int(booked)
		code, codeErr := c.code(ct.Ticker.Symbol)
		if codeErr != nil {
			return ingest.ConfirmedTicker{}, codeErr
		}
		ct.TradingCurrency = code
		isin, isinErr := ledger.ParseISIN(c.isin)
		if isinErr != nil {
			return ingest.ConfirmedTicker{}, fmt.Errorf("ticker %s: %w", ct.Ticker.Symbol, isinErr)
		}
		ct.ISIN = isin
		return ct, nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading the confirmed tickers: %w", err)
	}
	return out, nil
}

// TickerAccounts returns the accounts of the rows of instance's ticker
// symbol at every instance of its institution: only those waiting for the
// ticker's ISIN, or all of them with all, in ID order. A ticker never
// reported is app.ErrUnknownTicker.
func (s *Store) TickerAccounts(ctx context.Context, instance, symbol string, all bool) ([]ledger.AccountID, error) {
	var institution string
	err := s.pool.QueryRow(ctx, `SELECT institution FROM ingest.tickers WHERE instance = $1 AND symbol = $2`,
		instance, symbol).Scan(&institution)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, fmt.Errorf("the accounts of %s of %s: %w", symbol, instance, app.ErrUnknownTicker)
	case err != nil:
		return nil, fmt.Errorf("the accounts of %s of %s: %w", symbol, instance, err)
	}
	rows, err := s.pool.Query(ctx, `SELECT account_id FROM (
			SELECT e.account_id FROM ingest.ticker_entries e JOIN ingest.tickers t ON t.instance = e.instance AND t.symbol = e.symbol
				WHERE t.institution = $1 AND e.symbol = $2 AND ($3 OR NOT e.booked)
			UNION
			SELECT a.account_id FROM ingest.ticker_assertions a JOIN ingest.tickers t ON t.instance = a.instance AND t.symbol = a.symbol
				WHERE t.institution = $1 AND a.symbol = $2 AND ($3 OR NOT a.asserted)) x
		ORDER BY account_id COLLATE "C"`, institution, symbol, all)
	if err != nil {
		return nil, fmt.Errorf("the accounts of %s of %s: %w", symbol, instance, err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("the accounts of %s of %s: %w", symbol, instance, err)
	}
	out := make([]ledger.AccountID, len(ids))
	for i, id := range ids {
		out[i] = ledger.AccountID(id)
	}
	return out, nil
}

// ConfirmTicker records the owner's ISIN for a ticker at its instance's
// institution, once (D-153), with the trading currency the owner picked
// when the confirmation must pick one (ingest.ConfirmListing, D-170), and
// returns what it books: the listing it settles and the rows still waiting
// for the ticker at every instance of the institution, in instance and key
// order. A confirmation that picks wrongly records nothing, so the rows
// keep waiting; confirming another ISIN, or another pick, is
// app.ErrTickerConfirmed.
func (s *Store) ConfirmTicker(ctx context.Context, c ingest.TickerConfirmation) (ingest.TickerBooking, error) {
	var out ingest.TickerBooking
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var err error
		out, err = confirmIn(ctx, tx, &c)
		return err
	})
	if err != nil {
		return ingest.TickerBooking{}, fmt.Errorf("confirming %s of %s: %w", c.Symbol, c.Instance, err)
	}
	return out, nil
}

// confirmIn is ConfirmTicker's work in tx.
func confirmIn(ctx context.Context, tx pgx.Tx, c *ingest.TickerConfirmation) (ingest.TickerBooking, error) {
	r, err := loadTicker(ctx, tx, c.Instance, c.Symbol)
	if err != nil {
		return ingest.TickerBooking{}, err
	}
	stored, err := confirmed(ctx, tx, &r, c)
	if err != nil {
		return ingest.TickerBooking{}, err
	}
	var out ingest.TickerBooking
	if out.Listing, err = stored.listing(c.Symbol); err != nil {
		return ingest.TickerBooking{}, err
	}
	out.Rows, err = tickerRows(ctx, tx, r.institution, c.Symbol, false)
	return out, err
}

// confirmed returns the current confirmation of c's ticker at r's
// institution, recording c as its first when there is none. Another ISIN,
// or another pick, than the current one is app.ErrTickerConfirmed.
func confirmed(ctx context.Context, tx pgx.Tx, r *tickerRow, c *ingest.TickerConfirmation) (confirmation, error) {
	stored, ok, err := current(ctx, tx, r.institution, c.Symbol)
	switch {
	case err != nil:
		return confirmation{}, err
	case ok && (stored.isin != c.ISIN.String() || (c.Currency != "" && c.Currency != stored.picked)):
		return confirmation{}, app.ErrTickerConfirmed
	case ok:
		return stored, nil
	}
	if stored, err = r.confirmationOf(c.Symbol, c.ISIN, c.Currency); err != nil {
		return confirmation{}, err
	}
	stored.version = 1
	if err := record(ctx, tx, r.institution, c.Symbol, &stored); err != nil {
		return confirmation{}, err
	}
	return stored, nil
}

// CorrectTicker records the owner's correction of a ticker's confirmed
// ISIN, and trading currency, at its instance's institution (D-153) as a
// new version of its confirmation, which keeps the earlier ones, and
// returns what it books: the listing it settles, the listing it replaces
// and every row of the ticker at every instance of the institution, booked
// or waiting, in instance and key order. Correcting to the current ISIN
// and pick records nothing. A ticker never confirmed is
// ingest.ErrTickerNotConfirmed; a wrong pick an
// *ingest.TradingCurrencyChoiceError that records nothing (D-170).
func (s *Store) CorrectTicker(ctx context.Context, c ingest.TickerConfirmation) (ingest.TickerBooking, error) {
	var out ingest.TickerBooking
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var err error
		out, err = correctIn(ctx, tx, &c)
		return err
	})
	if err != nil {
		return ingest.TickerBooking{}, fmt.Errorf("correcting %s of %s: %w", c.Symbol, c.Instance, err)
	}
	return out, nil
}

// correctIn is CorrectTicker's work in tx.
func correctIn(ctx context.Context, tx pgx.Tx, c *ingest.TickerConfirmation) (ingest.TickerBooking, error) {
	r, err := loadTicker(ctx, tx, c.Instance, c.Symbol)
	if err != nil {
		return ingest.TickerBooking{}, err
	}
	previous, ok, err := current(ctx, tx, r.institution, c.Symbol)
	switch {
	case err != nil:
		return ingest.TickerBooking{}, err
	case !ok:
		return ingest.TickerBooking{}, ingest.ErrTickerNotConfirmed
	}
	var out ingest.TickerBooking
	if out.Previous, err = previous.listing(c.Symbol); err != nil {
		return ingest.TickerBooking{}, err
	}
	next := previous
	if previous.isin != c.ISIN.String() || previous.picked != c.Currency {
		if next, err = r.confirmationOf(c.Symbol, c.ISIN, c.Currency); err != nil {
			return ingest.TickerBooking{}, err
		}
		next.version = previous.version + 1
		if recErr := record(ctx, tx, r.institution, c.Symbol, &next); recErr != nil {
			return ingest.TickerBooking{}, recErr
		}
	}
	if out.Listing, err = next.listing(c.Symbol); err != nil {
		return ingest.TickerBooking{}, err
	}
	out.Rows, err = tickerRows(ctx, tx, r.institution, c.Symbol, true)
	return out, err
}

// tickerRows returns the entries and assertions of symbol at every
// instance of institution, by instance: those not booked yet, or all with
// all, entries in key order, assertions in account and date order.
func tickerRows(ctx context.Context, tx pgx.Tx, institution, symbol string, all bool) ([]ingest.TickerRows, error) {
	rows, err := tx.Query(ctx, `SELECT instance FROM ingest.tickers WHERE institution = $1 AND symbol = $2
		ORDER BY instance COLLATE "C"`, institution, symbol)
	if err != nil {
		return nil, fmt.Errorf("loading the instances: %w", err)
	}
	instances, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("loading the instances: %w", err)
	}
	var out []ingest.TickerRows
	for _, instance := range instances {
		r := ingest.TickerRows{Instance: instance}
		if r.Entries, err = tickerEntries(ctx, tx, instance, symbol, all); err != nil {
			return nil, err
		}
		if r.Assertions, err = tickerAssertions(ctx, tx, instance, symbol, all); err != nil {
			return nil, err
		}
		if len(r.Entries) > 0 || len(r.Assertions) > 0 {
			out = append(out, r)
		}
	}
	return out, nil
}

// tickerEntries returns the entries of instance's ticker symbol: those not
// booked yet, or all with all, in key order.
func tickerEntries(ctx context.Context, tx pgx.Tx, instance, symbol string, all bool) ([]ingest.TickerEntry, error) {
	rows, err := tx.Query(ctx, `SELECT key_source, external_id, account_id, kind, date::text, quantity, amount, currency,
			charge, charge_currency, description, ref, coalesce(lot_opened::text, ''), coalesce(open_price, '')
		FROM ingest.ticker_entries WHERE instance = $1 AND symbol = $2 AND ($3 OR NOT booked)
		ORDER BY key_source COLLATE "C", external_id COLLATE "C"`, instance, symbol, all)
	if err != nil {
		return nil, fmt.Errorf("loading the ticker entries: %w", err)
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ingest.TickerEntry, error) {
		var r entryRow
		if scanErr := row.Scan(&r.source, &r.id, &r.account, &r.kind, &r.date, &r.quantity, &r.amount, &r.currency,
			&r.charge, &r.chargeCurrency, &r.description, &r.ref, &r.opened, &r.openPrice); scanErr != nil {
			return ingest.TickerEntry{}, fmt.Errorf("scanning a ticker entry: %w", scanErr)
		}
		return r.entry(symbol)
	})
	if err != nil {
		return nil, fmt.Errorf("loading the ticker entries: %w", err)
	}
	return out, nil
}

// tickerAssertions returns the ticker assertions of instance's ticker
// symbol: those not asserted yet, or all with all, in account and date
// order.
func tickerAssertions(ctx context.Context, tx pgx.Tx, instance, symbol string, all bool) ([]ingest.TickerAssertion, error) {
	rows, err := tx.Query(ctx, `SELECT account_id, date::text, quantity, ref FROM ingest.ticker_assertions
		WHERE instance = $1 AND symbol = $2 AND ($3 OR NOT asserted) ORDER BY account_id COLLATE "C", date`,
		instance, symbol, all)
	if err != nil {
		return nil, fmt.Errorf("loading the ticker assertions: %w", err)
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ingest.TickerAssertion, error) {
		var account, day, quantity, ref string
		if scanErr := row.Scan(&account, &day, &quantity, &ref); scanErr != nil {
			return ingest.TickerAssertion{}, fmt.Errorf("scanning a ticker assertion: %w", scanErr)
		}
		on, dateErr := kernel.ParseDate(day)
		qty, qtyErr := kernel.ParseQuantity(quantity)
		if joined := errors.Join(dateErr, qtyErr); joined != nil {
			return ingest.TickerAssertion{}, fmt.Errorf("%w %s of %s on %s: %w", errStoredEntry, symbol, account, day, joined)
		}
		return ingest.TickerAssertion{Symbol: symbol, Account: ledger.AccountID(account), Date: on, Quantity: qty, Ref: ref}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading the ticker assertions: %w", err)
	}
	return out, nil
}

// entryRow is one stored ticker entry's columns.
type entryRow struct {
	source, id, account, kind, date, quantity, amount, currency string
	charge, chargeCurrency, description, ref, opened, openPrice string
}

// entry returns the ticker entry of symbol r stores.
func (r *entryRow) entry(symbol string) (ingest.TickerEntry, error) {
	kinds := map[string]ledger.TransactionKind{
		"buy": ledger.KindBuy, "sell": ledger.KindSell, "dividend": ledger.KindDividend, "opening-lot": ledger.KindOpeningLot,
	}
	kind, ok := kinds[r.kind]
	on, dateErr := kernel.ParseDate(r.date)
	qty, qtyErr := kernel.ParseQuantity(r.quantity)
	amount, amountErr := r.money()
	if err := errors.Join(dateErr, qtyErr, amountErr); !ok || err != nil {
		return ingest.TickerEntry{}, fmt.Errorf("%w %s/%s: %w", errStoredEntry, r.source, r.id, err)
	}
	e := ingest.TickerEntry{
		Symbol: symbol, Kind: kind, Quantity: qty, Amount: amount, Description: r.description, Ref: r.ref,
		Header: ledger.TransactionHeader{
			Key: ledger.TransactionKey{Source: r.source, ExternalID: r.id}, Account: ledger.AccountID(r.account), Date: on,
		},
	}
	if r.openPrice != "" {
		price, err := kernel.ParseDecimal(r.openPrice)
		if err != nil {
			return ingest.TickerEntry{}, fmt.Errorf("%w %s/%s: %w", errStoredEntry, r.source, r.id, err)
		}
		e.OpenPrice = price
	}
	if r.opened != "" {
		opened, err := kernel.ParseDate(r.opened)
		if err != nil {
			return ingest.TickerEntry{}, fmt.Errorf("%w %s/%s: %w", errStoredEntry, r.source, r.id, err)
		}
		e.Opened = opened
	}
	if r.charge != "" {
		c, err := kernel.ParseCurrency(r.chargeCurrency)
		if err != nil {
			return ingest.TickerEntry{}, fmt.Errorf("%w %s/%s: %w", errStoredEntry, r.source, r.id, err)
		}
		if e.Charge, err = kernel.ParseMoney(r.charge, c); err != nil {
			return ingest.TickerEntry{}, fmt.Errorf("%w %s/%s: %w", errStoredEntry, r.source, r.id, err)
		}
	}
	return e, nil
}

// money returns r's amount, or none for an opening lot priced by its open
// price in its trading currency, stored without a currency (D-159).
func (r *entryRow) money() (kernel.Money, error) {
	if r.currency == "" {
		return kernel.Money{}, nil
	}
	cur, err := kernel.ParseCurrency(r.currency)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("currency: %w", err)
	}
	m, err := kernel.ParseMoney(r.amount, cur)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("amount: %w", err)
	}
	return m, nil
}

// BookedTickerRows marks rows booked: each entry by its instance and key,
// each assertion by its instance, ticker, account and date. A correction
// books them anew (D-153); the owner's confirmation no longer waits for
// them.
func (s *Store) BookedTickerRows(ctx context.Context, rows []ingest.TickerRows) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for i := range rows {
			if err := markBooked(ctx, tx, &rows[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("marking ticker rows booked: %w", err)
	}
	return nil
}

// markBooked marks r's entries booked and its assertions asserted.
func markBooked(ctx context.Context, tx pgx.Tx, r *ingest.TickerRows) error {
	for j := range r.Entries {
		k := r.Entries[j].Header.Key
		if _, err := tx.Exec(ctx, `UPDATE ingest.ticker_entries SET booked = true
			WHERE instance = $1 AND key_source = $2 AND external_id = $3`, r.Instance, k.Source, k.ExternalID); err != nil {
			return fmt.Errorf("marking %s booked: %w", k, err)
		}
	}
	for j := range r.Assertions {
		a := &r.Assertions[j]
		if _, err := tx.Exec(ctx, `UPDATE ingest.ticker_assertions SET asserted = true
			WHERE instance = $1 AND symbol = $2 AND account_id = $3 AND date = $4`,
			r.Instance, a.Symbol, string(a.Account), a.Date.String()); err != nil {
			return fmt.Errorf("marking the holding of %s asserted: %w", a.Symbol, err)
		}
	}
	return nil
}
