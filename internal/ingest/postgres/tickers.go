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

// errStoredEntry reports a stored ticker entry that doesn't read back.
var errStoredEntry = errors.New("stored ticker entry")

// ConfirmedTickers returns the ISIN the owner confirmed for each ticker of
// instance.
func (s *Store) ConfirmedTickers(ctx context.Context, instance string) (map[string]ledger.ISIN, error) {
	rows, err := s.pool.Query(ctx, `SELECT symbol, isin FROM ingest.tickers WHERE instance = $1 AND isin IS NOT NULL`, instance)
	if err != nil {
		return nil, fmt.Errorf("loading the confirmed tickers of %s: %w", instance, err)
	}
	out := map[string]ledger.ISIN{}
	var symbol, code string
	_, err = pgx.ForEachRow(rows, []any{&symbol, &code}, func() error {
		isin, parseErr := ledger.ParseISIN(code)
		if parseErr != nil {
			return fmt.Errorf("ticker %s: %w", symbol, parseErr)
		}
		out[symbol] = isin
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading the confirmed tickers of %s: %w", instance, err)
	}
	return out, nil
}

// HoldTickers records instance's tickers and the entries waiting for their
// ISIN in one transaction, and returns the tickers it hadn't seen.
func (s *Store) HoldTickers(ctx context.Context, instance string, tickers []ingest.Ticker, waiting []ingest.TickerEntry) ([]ingest.Ticker, error) {
	var added []ingest.Ticker
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, t := range tickers {
			tag, err := tx.Exec(ctx, `INSERT INTO ingest.tickers (instance, symbol, code, country, name)
				VALUES ($1, $2, $3, $4, $5) ON CONFLICT (instance, symbol) DO NOTHING`,
				instance, t.Symbol, t.Code, t.Country, t.Name)
			if err != nil {
				return fmt.Errorf("storing ticker %s: %w", t.Symbol, err)
			}
			if tag.RowsAffected() == 1 {
				added = append(added, t)
			}
		}
		for i := range waiting {
			if err := holdEntry(ctx, tx, instance, &waiting[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("holding the ticker entries of %s: %w", instance, err)
	}
	return added, nil
}

// holdEntry stores one waiting entry, replacing an earlier one of its key.
func holdEntry(ctx context.Context, tx pgx.Tx, instance string, e *ingest.TickerEntry) error {
	charge, chargeCurrency := "", ""
	if !e.Charge.IsZero() {
		charge, chargeCurrency = e.Charge.Amount().String(), e.Charge.Currency().String()
	}
	var opened *string
	if !e.Opened.IsZero() {
		day := e.Opened.String()
		opened = &day
	}
	_, err := tx.Exec(ctx, `INSERT INTO ingest.ticker_entries (instance, key_source, external_id, symbol, account_id, kind,
			date, quantity, amount, currency, charge, charge_currency, description, ref, lot_opened)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (instance, key_source, external_id) DO UPDATE SET symbol = EXCLUDED.symbol,
			account_id = EXCLUDED.account_id, kind = EXCLUDED.kind, date = EXCLUDED.date, quantity = EXCLUDED.quantity,
			amount = EXCLUDED.amount, currency = EXCLUDED.currency, charge = EXCLUDED.charge,
			charge_currency = EXCLUDED.charge_currency, description = EXCLUDED.description, ref = EXCLUDED.ref,
			lot_opened = EXCLUDED.lot_opened`,
		instance, e.Header.Key.Source, e.Header.Key.ExternalID, e.Symbol, string(e.Header.Account), e.Kind.String(),
		e.Header.Date.String(), e.Quantity.String(), e.Amount.Amount().String(), e.Amount.Currency().String(),
		charge, chargeCurrency, e.Description, e.Ref, opened)
	if err != nil {
		return fmt.Errorf("holding %s: %w", e.Header.Key, err)
	}
	return nil
}

// SaveProposals records the ISIN proposed for tickers of instance.
func (s *Store) SaveProposals(ctx context.Context, instance string, proposals []ingest.ISINProposal) error {
	for _, p := range proposals {
		_, err := s.pool.Exec(ctx, `UPDATE ingest.tickers SET proposed_isin = $3, proposed_name = $4
			WHERE instance = $1 AND symbol = $2`, instance, p.Symbol, p.ISIN.String(), p.Name)
		if err != nil {
			return fmt.Errorf("saving the proposal for %s of %s: %w", p.Symbol, instance, err)
		}
	}
	return nil
}

// PendingTickers returns every ticker without a confirmation, by instance
// and symbol, with its proposal and how many entries wait for it.
func (s *Store) PendingTickers(ctx context.Context) ([]ingest.PendingTicker, error) {
	rows, err := s.pool.Query(ctx, `SELECT t.instance, t.symbol, t.code, t.country, t.name, coalesce(t.proposed_isin, ''),
			t.proposed_name, (SELECT count(*) FROM ingest.ticker_entries e WHERE e.instance = t.instance AND e.symbol = t.symbol)
		FROM ingest.tickers t WHERE t.isin IS NULL ORDER BY t.instance COLLATE "C", t.symbol COLLATE "C"`)
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
		&p.Proposal.Name, &waiting); err != nil {
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

// ConfirmTicker records the owner's ISIN for a ticker, once, and returns the
// entries still waiting for it, in key order.
func (s *Store) ConfirmTicker(ctx context.Context, c ingest.TickerConfirmation) ([]ingest.TickerEntry, error) {
	var out []ingest.TickerEntry
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var current *string
		err := tx.QueryRow(ctx, `SELECT isin FROM ingest.tickers WHERE instance = $1 AND symbol = $2 FOR UPDATE`,
			c.Instance, c.Symbol).Scan(&current)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return app.ErrUnknownTicker
		case err != nil:
			return fmt.Errorf("loading the ticker: %w", err)
		case current != nil && *current != c.ISIN.String():
			return app.ErrTickerConfirmed
		case current == nil:
			if _, execErr := tx.Exec(ctx, `UPDATE ingest.tickers SET isin = $3, confirmed_at = now()
				WHERE instance = $1 AND symbol = $2`, c.Instance, c.Symbol, c.ISIN.String()); execErr != nil {
				return fmt.Errorf("confirming the ticker: %w", execErr)
			}
		}
		out, err = waitingEntries(ctx, tx, c.Instance, c.Symbol)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("confirming %s of %s: %w", c.Symbol, c.Instance, err)
	}
	return out, nil
}

// waitingEntries returns the entries of instance's ticker symbol, in key
// order.
func waitingEntries(ctx context.Context, tx pgx.Tx, instance, symbol string) ([]ingest.TickerEntry, error) {
	rows, err := tx.Query(ctx, `SELECT key_source, external_id, account_id, kind, date::text, quantity, amount, currency,
			charge, charge_currency, description, ref, coalesce(lot_opened::text, '')
		FROM ingest.ticker_entries WHERE instance = $1 AND symbol = $2
		ORDER BY key_source COLLATE "C", external_id COLLATE "C"`, instance, symbol)
	if err != nil {
		return nil, fmt.Errorf("loading the waiting entries: %w", err)
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ingest.TickerEntry, error) {
		var r entryRow
		if scanErr := row.Scan(&r.source, &r.id, &r.account, &r.kind, &r.date, &r.quantity, &r.amount, &r.currency,
			&r.charge, &r.chargeCurrency, &r.description, &r.ref, &r.opened); scanErr != nil {
			return ingest.TickerEntry{}, fmt.Errorf("scanning a waiting entry: %w", scanErr)
		}
		return r.entry(symbol)
	})
	if err != nil {
		return nil, fmt.Errorf("loading the waiting entries: %w", err)
	}
	return out, nil
}

// entryRow is one stored ticker entry's columns.
type entryRow struct {
	source, id, account, kind, date, quantity, amount, currency string
	charge, chargeCurrency, description, ref, opened            string
}

// entry returns the ticker entry of symbol r stores.
func (r *entryRow) entry(symbol string) (ingest.TickerEntry, error) {
	kinds := map[string]ledger.TransactionKind{
		"buy": ledger.KindBuy, "sell": ledger.KindSell, "dividend": ledger.KindDividend, "opening-lot": ledger.KindOpeningLot,
	}
	kind, ok := kinds[r.kind]
	on, dateErr := kernel.ParseDate(r.date)
	qty, qtyErr := kernel.ParseQuantity(r.quantity)
	cur, curErr := kernel.ParseCurrency(r.currency)
	amount, amountErr := kernel.ParseMoney(r.amount, cur)
	if err := errors.Join(dateErr, qtyErr, curErr, amountErr); !ok || err != nil {
		return ingest.TickerEntry{}, fmt.Errorf("%w %s/%s: %w", errStoredEntry, r.source, r.id, err)
	}
	e := ingest.TickerEntry{
		Symbol: symbol, Kind: kind, Quantity: qty, Amount: amount, Description: r.description, Ref: r.ref,
		Header: ledger.TransactionHeader{
			Key: ledger.TransactionKey{Source: r.source, ExternalID: r.id}, Account: ledger.AccountID(r.account), Date: on,
		},
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

// BookedTickerEntries drops the waiting entries of instance with keys.
func (s *Store) BookedTickerEntries(ctx context.Context, instance string, keys []ledger.TransactionKey) error {
	sources := make([]string, len(keys))
	ids := make([]string, len(keys))
	for i, k := range keys {
		sources[i], ids[i] = k.Source, k.ExternalID
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM ingest.ticker_entries e USING unnest($2::text[], $3::text[]) AS k(source, id)
		WHERE e.instance = $1 AND e.key_source = k.source AND e.external_id = k.id`, instance, sources, ids)
	if err != nil {
		return fmt.Errorf("dropping the booked ticker entries of %s: %w", instance, err)
	}
	return nil
}
