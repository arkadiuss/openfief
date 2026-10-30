package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// valuationArgs returns the insertValuation arguments of r: its gross value
// in its own currency and in PLN, or the missing input; its net and
// liquidation values in PLN, or what each lacks.
func valuationArgs(r *valuation.HoldingValuation) []any {
	var value, currency, pln, missing, net, liquidation, netMissing, liquidationMissing *string
	text := func(s string) *string { return &s }
	switch {
	case !r.Valued():
		missing = text(r.Missing.String())
	default:
		value, currency, pln = text(r.Value.Amount().String()), text(r.Value.Currency().String()), text(r.PLN.Amount().String())
		if r.NetValued() {
			net = text(r.Net.Amount().String())
		} else {
			netMissing = text(r.NetMissing.String())
		}
		if r.LiquidationValued() {
			liquidation = text(r.Liquidation.Amount().String())
		} else {
			liquidationMissing = text(r.LiquidationMissing.String())
		}
	}
	return []any{
		string(r.Account), r.Instrument.ISIN.String(), r.Instrument.Symbol, r.Date.String(), r.Basis.String(),
		r.Quantity.String(), value, currency, pln, r.Source, optionalDate(r.PriceDate), optionalDate(r.RateDate),
		missing, r.Detail, net, liquidation, netMissing, liquidationMissing,
	}
}

// Valuations returns account's stored valuations dated on or after from
// and on or before to, in date, instrument and basis order.
func (s *Store) Valuations(ctx context.Context, account ledger.AccountID, from, to kernel.Date) ([]valuation.HoldingValuation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT account_id, isin, symbol, date::text, basis, quantity::text, coalesce(value::text, ''),
		       coalesce(value_currency, ''), coalesce(pln::text, ''), source, coalesce(price_date::text, ''),
		       coalesce(rate_date::text, ''), coalesce(missing, ''), detail,
		       coalesce(net::text, ''), coalesce(liquidation::text, ''), coalesce(net_missing, ''), coalesce(liquidation_missing, '')
		FROM valuation.holding_valuations
		WHERE account_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date, isin, symbol COLLATE "C", basis`, string(account), from.String(), to.String())
	if err != nil {
		return nil, fmt.Errorf("loading the valuations of %s: %w", account, err)
	}
	out, err := pgx.CollectRows(rows, scanValuation)
	if err != nil {
		return nil, fmt.Errorf("loading the valuations of %s: %w", account, err)
	}
	return out, nil
}

// scanValuation reads one row of Valuations.
func scanValuation(row pgx.CollectableRow) (valuation.HoldingValuation, error) {
	var account, source, detail string
	var t valuationText
	if err := row.Scan(&account, &t.isin, &t.symbol, &t.day, &t.basis, &t.quantity, &t.value, &t.currency, &t.pln,
		&source, &t.priceDate, &t.rateDate, &t.missing, &detail, &t.net, &t.liquidation, &t.netMissing, &t.liquidationMissing); err != nil {
		return valuation.HoldingValuation{}, fmt.Errorf("reading a valuation: %w", err)
	}
	v, err := decodeValuation(&t)
	if err != nil {
		return valuation.HoldingValuation{}, fmt.Errorf("%w: valuation of %s: %w", ErrCorrupt, account, err)
	}
	v.Account, v.Source, v.Detail = ledger.AccountID(account), source, detail
	return v, nil
}
