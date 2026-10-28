package postgres

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// parseMoney parses a stored numeric's text and currency code.
func parseMoney(amount, currency string) (kernel.Money, error) {
	c, err := kernel.ParseCurrency(currency)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("currency: %w", err)
	}
	m, err := kernel.ParseMoney(amount, c)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("amount: %w", err)
	}
	return m, nil
}

// wrapperNames are the stored names of the tax wrappers.
func wrapperNames() map[pl.TaxWrapper]string {
	return map[pl.TaxWrapper]string{pl.Taxable: "taxable", pl.IKE: "ike", pl.IKZE: "ikze"}
}

// wrapperName returns the stored name of w.
func wrapperName(w pl.TaxWrapper) (string, error) {
	name, ok := wrapperNames()[w]
	if !ok {
		return "", fmt.Errorf("%w: %d", pl.ErrUnknownTaxWrapper, w)
	}
	return name, nil
}

// parseWrapper returns the tax wrapper stored as name.
func parseWrapper(name string) (pl.TaxWrapper, error) {
	for w, n := range wrapperNames() {
		if n == name {
			return w, nil
		}
	}
	return 0, fmt.Errorf("%w: tax wrapper %q", ErrCorrupt, name)
}

// optionalDate returns d's text, or nil (SQL NULL) for the zero date.
func optionalDate(d kernel.Date) *string {
	if d.IsZero() {
		return nil
	}
	s := d.String()
	return &s
}

// parseOptionalDate parses a stored date's text; "" is the zero date.
func parseOptionalDate(s string) (kernel.Date, error) {
	if s == "" {
		return kernel.Date{}, nil
	}
	d, err := kernel.ParseDate(s)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("date: %w", err)
	}
	return d, nil
}

// valuationText is one stored valuation's columns as text; NULL is "".
type valuationText struct {
	isin, symbol, day, basis, quantity, value, currency, pln, priceDate, rateDate, missing string
}

// decodeValuation decodes the columns of a stored valuation, all but its
// account, source and detail.
func decodeValuation(t *valuationText) (valuation.HoldingValuation, error) {
	isin, err1 := ledger.ParseISIN(t.isin)
	day, err2 := kernel.ParseDate(t.day)
	basis, err3 := valuation.ParseBasis(t.basis)
	q, err4 := kernel.ParseQuantity(t.quantity)
	priceDate, err5 := parseOptionalDate(t.priceDate)
	rateDate, err6 := parseOptionalDate(t.rateDate)
	if err := errors.Join(err1, err2, err3, err4, err5, err6); err != nil {
		return valuation.HoldingValuation{}, err
	}
	v := valuation.HoldingValuation{
		Instrument: ledger.Instrument{ISIN: isin, Symbol: t.symbol}, Date: day, Basis: basis, Quantity: q,
		PriceDate: priceDate, RateDate: rateDate,
	}
	if t.missing != "" {
		reason, err := valuation.ParseMissingReason(t.missing)
		if err != nil {
			return valuation.HoldingValuation{}, fmt.Errorf("missing input: %w", err)
		}
		v.Missing = reason
		return v, nil
	}
	value, err1 := parseMoney(t.value, t.currency)
	pln, err2 := parseMoney(t.pln, "PLN")
	if err := errors.Join(err1, err2); err != nil {
		return valuation.HoldingValuation{}, err
	}
	v.Value, v.PLN = value, pln
	return v, nil
}
