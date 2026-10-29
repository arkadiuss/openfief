package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// UnsupportedReportingCurrencyError reports a reporting currency (D-62) the
// backtest cannot convert to: FX rates are quoted against PLN only (D-20),
// and cross rates are undecided (T-12), as in the read API.
type UnsupportedReportingCurrencyError struct {
	Currency kernel.Currency
}

// Error names the currency.
func (e *UnsupportedReportingCurrencyError) Error() string {
	return fmt.Sprintf("reporting currency %s is not supported: only %s", e.Currency, valuation.PLN())
}

// Models runs the model portfolio's use cases: the owner's entry (D-52)
// and its report for the read API (D-53, D-54).
type Models struct {
	store     ModelStore
	market    MarketData
	owners    Owners
	reporting kernel.Currency
}

// New returns the use cases over the ports, with the backtest in the
// reporting currency reporting (D-62): PLN, or an
// *UnsupportedReportingCurrencyError.
func New(store ModelStore, market MarketData, owners Owners, reporting kernel.Currency) (*Models, error) {
	if !reporting.Equal(valuation.PLN()) {
		return nil, &UnsupportedReportingCurrencyError{Currency: reporting}
	}
	return &Models{store: store, market: market, owners: owners, reporting: reporting}, nil
}

// Settings returns the stored model portfolio, on or off, to the owner: a
// session of owner scope on origin.
func (m *Models) Settings(ctx context.Context, origin, sessionToken string) (portfolio.Model, error) {
	if err := m.checkOwner(ctx, origin, sessionToken); err != nil {
		return portfolio.Model{}, err
	}
	model, err := m.store.Model(ctx)
	if err != nil {
		return portfolio.Model{}, fmt.Errorf("loading the model portfolio: %w", err)
	}
	return model, nil
}

// Save replaces the model portfolio with model for the owner: a session of
// owner scope on origin. A model that breaks a rule of Model.Validate, or
// names an instrument that is not "ISIN/symbol" with a valid ISIN, is an
// error wrapping portfolio.ErrInvalidModel and nothing is stored.
func (m *Models) Save(ctx context.Context, origin, sessionToken string, model portfolio.Model) (portfolio.Model, error) {
	if err := m.checkOwner(ctx, origin, sessionToken); err != nil {
		return portfolio.Model{}, err
	}
	if err := CheckModel(&model); err != nil {
		return portfolio.Model{}, err
	}
	if err := m.store.SaveModel(ctx, model); err != nil {
		return portfolio.Model{}, fmt.Errorf("saving the model portfolio: %w", err)
	}
	return model, nil
}

// CheckModel checks model's rules (Model.Validate) and that each instrument
// target names an instrument as the ledger writes one: "ISIN/symbol" with
// a valid ISIN and a symbol.
func CheckModel(model *portfolio.Model) error {
	if err := model.Validate(); err != nil {
		return fmt.Errorf("checking the model portfolio: %w", err)
	}
	for i := range model.Targets {
		t := &model.Targets[i]
		if t.Kind != portfolio.TargetInstrument {
			continue
		}
		if _, err := ParseInstrument(t.Name); err != nil {
			return fmt.Errorf("checking the model portfolio: %w: %w", portfolio.ErrInvalidModel, err)
		}
	}
	return nil
}

// ParseInstrument reads an instrument written "ISIN/symbol", as
// ledger.Instrument.String writes it.
func ParseInstrument(name string) (ledger.Instrument, error) {
	code, symbol, _ := strings.Cut(name, "/")
	isin, err := ledger.ParseISIN(code)
	if err != nil {
		return ledger.Instrument{}, fmt.Errorf("instrument %q: %w", name, err)
	}
	in := ledger.Instrument{ISIN: isin, Symbol: symbol}
	if err := in.Validate(); err != nil {
		return ledger.Instrument{}, fmt.Errorf("instrument %q: %w", name, err)
	}
	return in, nil
}

// checkOwner checks that sessionToken is the owner's on origin.
func (m *Models) checkOwner(ctx context.Context, origin, sessionToken string) error {
	if err := m.owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("checking the session: %w", err)
	}
	return nil
}
