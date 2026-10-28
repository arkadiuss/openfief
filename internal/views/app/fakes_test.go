package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// Synthetic ISINs with valid check digits; no real instrument is implied.
const (
	isinPL = "PLTEST000010"
	isinIE = "IETEST000026"
)

var errStore = errors.New("store unavailable")

// fakeLedger answers the ledger queries from fixed data, like
// ledgerapp.Ledger: an unknown account is ledgerapp.ErrUnknownAccount.
type fakeLedger struct {
	sleeves  []portfolio.Sleeve
	accounts []ledger.Account
	txs      map[ledger.AccountID][]ledger.Transaction
	holdings map[ledger.AccountID][]ledger.Holding
	cash     map[ledger.AccountID][]kernel.Money
	checks   map[ledger.AccountID][]ledgerapp.AssertionCheck
	fail     error
}

func (f *fakeLedger) Sleeves(context.Context) ([]portfolio.Sleeve, error) { return f.sleeves, f.fail }

func (f *fakeLedger) Accounts(context.Context) ([]ledger.Account, error) { return f.accounts, f.fail }

func (f *fakeLedger) Account(_ context.Context, id ledger.AccountID) (ledger.Account, error) {
	if f.fail != nil {
		return ledger.Account{}, f.fail
	}
	for _, a := range f.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return ledger.Account{}, fmt.Errorf("%w: %s", ledgerapp.ErrUnknownAccount, id)
}

func (f *fakeLedger) CurrentTransactions(ctx context.Context, id ledger.AccountID) ([]ledger.Transaction, error) {
	_, err := f.Account(ctx, id)
	return f.txs[id], err
}

func (f *fakeLedger) HoldingsOn(ctx context.Context, id ledger.AccountID, _ kernel.Date) ([]ledger.Holding, error) {
	_, err := f.Account(ctx, id)
	return f.holdings[id], err
}

func (f *fakeLedger) LotsOn(ctx context.Context, id ledger.AccountID, _ kernel.Date) ([]ledger.Lot, error) {
	_, err := f.Account(ctx, id)
	return nil, err
}

func (f *fakeLedger) CashOn(ctx context.Context, id ledger.AccountID, _ kernel.Date) ([]kernel.Money, error) {
	_, err := f.Account(ctx, id)
	return f.cash[id], err
}

func (f *fakeLedger) AssertionChecks(ctx context.Context, id ledger.AccountID) ([]ledgerapp.AssertionCheck, error) {
	_, err := f.Account(ctx, id)
	return f.checks[id], err
}

// fakeValuations returns the stored valuations of an account dated in the
// range asked for.
type fakeValuations struct {
	vals []views.HoldingValuation
	fail error
}

func (f *fakeValuations) HoldingValuations(_ context.Context, id ledger.AccountID, from, to kernel.Date) ([]views.HoldingValuation, error) {
	var out []views.HoldingValuation
	for i := range f.vals {
		if v := &f.vals[i]; v.Account == id && !v.Date.Before(from) && !v.Date.After(to) {
			out = append(out, *v)
		}
	}
	return out, f.fail
}

// fakeStatuses returns a fixed status per account; none is a never-valued
// account.
type fakeStatuses map[ledger.AccountID]app.ValuationStatus

func (f fakeStatuses) ValuationStatus(_ context.Context, id ledger.AccountID) (app.ValuationStatus, error) {
	return f[id], nil
}

// fakeRates returns fixed FX rates.
type fakeRates []valuation.FXRate

func (f fakeRates) LatestRates(context.Context, kernel.Date) ([]valuation.FXRate, error) {
	return f, nil
}

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

func cur(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	return c
}

func money(t *testing.T, amount, code string) kernel.Money {
	t.Helper()
	return kernel.NewMoney(dec(t, amount), cur(t, code))
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func instrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

// fixture is a synthetic portfolio on 2026-03-02: acc-1 at broker-a in core
// holds 10 EDO0334 bonds (valued 1050 PLN), 3 ETF1 (valued 30 USD, 120 PLN)
// and 2 ETF2 (not valued), with 100 USD and 50 PLN cash; acc-2 at broker-b in
// offense holds 100 EUR cash. USD is 4.00 PLN, EUR 4.30 PLN.
type fixture struct {
	ledger     *fakeLedger
	valuations *fakeValuations
	statuses   fakeStatuses
	rates      fakeRates
	on         kernel.Date
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	on := date(t, "2026-03-02")
	bond, etf1, etf2 := instrument(t, isinPL, "EDO0334"), instrument(t, isinIE, "ETF1"), instrument(t, isinIE, "ETF2")
	units := func(n int64) kernel.Quantity { return kernel.NewQuantity(kernel.NewDecimal(n, 0)) }
	value := func(in ledger.Instrument, day, v, c, pln string) views.HoldingValuation {
		return views.HoldingValuation{Account: "acc-1", Instrument: in, Date: date(t, day), Basis: valuation.Market, Source: "test", Value: money(t, v, c), Reporting: money(t, pln, "PLN")}
	}
	return &fixture{
		on: on,
		ledger: &fakeLedger{
			sleeves: []portfolio.Sleeve{{ID: "core", Name: "Core", Role: portfolio.RoleDefensive}, {ID: "offense", Name: "Offense", Role: portfolio.RoleOffensive}},
			accounts: []ledger.Account{
				{ID: "acc-1", Institution: "broker-a", Name: "Synthetic A", Sleeve: "core"},
				{ID: "acc-2", Institution: "broker-b", Name: "Synthetic B", Sleeve: "offense"},
			},
			holdings: map[ledger.AccountID][]ledger.Holding{"acc-1": {
				{Account: "acc-1", Instrument: bond, Date: on, Quantity: units(10)},
				{Account: "acc-1", Instrument: etf1, Date: on, Quantity: units(3)},
				{Account: "acc-1", Instrument: etf2, Date: on, Quantity: units(2)},
			}},
			cash: map[ledger.AccountID][]kernel.Money{
				"acc-1": {money(t, "50", "PLN"), money(t, "100", "USD")},
				"acc-2": {money(t, "100", "EUR")},
			},
		},
		valuations: &fakeValuations{vals: []views.HoldingValuation{
			value(bond, "2026-03-01", "1049.00", "PLN", "1049.00"),
			value(etf1, "2026-03-01", "29", "USD", "116.00"),
			value(bond, "2026-03-02", "1050.00", "PLN", "1050.00"),
			value(etf1, "2026-03-02", "30", "USD", "120.00"),
		}},
		statuses: fakeStatuses{"acc-1": {ValuedThrough: on, Pending: true}},
		rates: fakeRates{
			{Currency: cur(t, "USD"), Date: on, Rate: dec(t, "4.00"), Source: "test"},
			{Currency: cur(t, "EUR"), Date: on.AddDays(-3), Rate: dec(t, "4.30"), Source: "test"},
		},
	}
}

func (f *fixture) views(t *testing.T) *app.Views {
	t.Helper()
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"))
	if err != nil {
		t.Fatalf("app.New(PLN): %v", err)
	}
	return v
}
