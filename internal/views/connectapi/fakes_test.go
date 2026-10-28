package connectapi_test

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
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// Synthetic ISINs with valid check digits; no real instrument is implied.
const (
	isinPL = "PLTEST000010"
	isinIE = "IETEST000026"
)

// The bearer tokens the fake verifier knows.
const (
	goodToken   = "synthetic-good-token"
	brokenToken = "synthetic-unverifiable-token"
)

var errStore = errors.New("store unavailable")

// fakeVerifier accepts goodToken, cannot check brokenToken and rejects
// anything else.
type fakeVerifier struct{}

func (fakeVerifier) VerifyToken(_ context.Context, token string) error {
	switch token {
	case goodToken:
		return nil
	case brokenToken:
		return errStore
	}
	return fmt.Errorf("%w: unknown", app.ErrInvalidToken)
}

// fakeLedger answers the ledger queries from fixed data, like
// ledgerapp.Ledger: an unknown account is ledgerapp.ErrUnknownAccount.
type fakeLedger struct {
	accounts []ledger.Account
	txs      []ledger.Transaction
	holdings []ledger.Holding
	lots     []ledger.Lot
	cash     []kernel.Money
	checks   []ledgerapp.AssertionCheck
	fail     error
}

func (f *fakeLedger) Sleeves(context.Context) ([]portfolio.Sleeve, error) {
	return []portfolio.Sleeve{{ID: "core", Name: "Core", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive}}, f.fail
}

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
	return f.txs, err
}

func (f *fakeLedger) HoldingsOn(ctx context.Context, id ledger.AccountID, _ kernel.Date) ([]ledger.Holding, error) {
	_, err := f.Account(ctx, id)
	return f.holdings, err
}

func (f *fakeLedger) LotsOn(ctx context.Context, id ledger.AccountID, _ kernel.Date) ([]ledger.Lot, error) {
	_, err := f.Account(ctx, id)
	return f.lots, err
}

func (f *fakeLedger) CashOn(ctx context.Context, id ledger.AccountID, _ kernel.Date) ([]kernel.Money, error) {
	_, err := f.Account(ctx, id)
	return f.cash, err
}

func (f *fakeLedger) AssertionChecks(ctx context.Context, id ledger.AccountID) ([]ledgerapp.AssertionCheck, error) {
	_, err := f.Account(ctx, id)
	return f.checks, err
}

// fakeValuations returns every stored valuation, whatever is asked.
type fakeValuations []valuation.HoldingValuation

func (f fakeValuations) Valuations(context.Context, ledger.AccountID, kernel.Date, kernel.Date) ([]valuation.HoldingValuation, error) {
	return f, nil
}

// fakeStatuses reports every account valued through 2026-03-02 with a
// revaluation running.
type fakeStatuses struct{ through kernel.Date }

func (f fakeStatuses) Status(context.Context, ledger.AccountID) (valuationapp.ValuationStatus, error) {
	return valuationapp.ValuationStatus{ValuedThrough: f.through, Running: true}, nil
}

// fakeRates returns a currency's fixed FX rates up to a date.
type fakeRates []valuation.FXRate

func (f fakeRates) FXRates(_ context.Context, c kernel.Currency, through kernel.Date) ([]valuation.FXRate, error) {
	var out []valuation.FXRate
	for _, r := range f {
		if r.Currency.Equal(c) && !r.Date.After(through) {
			out = append(out, r)
		}
	}
	return out, nil
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

// newLedger returns a synthetic account acc-1 on 2026-03-02 with a buy of 3
// ETF1, an opening lot of 10 EDO0334 bonds, 100 USD cash, a matched and a
// flagged cash assertion.
func newLedger(t *testing.T) *fakeLedger {
	t.Helper()
	on := date(t, "2026-03-02")
	bond, etf := instrument(t, isinPL, "EDO0334"), instrument(t, isinIE, "ETF1")
	usd := cur(t, "USD")
	buy := ledger.Transaction{
		Key: ledger.TransactionKey{Source: "test", ExternalID: "buy-1"}, Account: "acc-1", Kind: ledger.KindBuy, Date: on,
		Description: "Synthetic buy", Postings: []ledger.Posting{
			{Book: ledger.HoldingBook(etf), Amount: money(t, "90.00", "USD"), Quantity: kernel.NewQuantity(dec(t, "3"))},
			{Book: ledger.CashBook(usd), Amount: money(t, "-90.00", "USD")},
		},
	}
	opening := ledger.Transaction{
		Key: ledger.TransactionKey{Source: "test", ExternalID: "open-1"}, Account: "acc-1", Kind: ledger.KindOpeningLot, Date: on,
		LotOpened: date(t, "2025-05-01"), Postings: []ledger.Posting{
			{Book: ledger.HoldingBook(bond), Amount: money(t, "1000.00", "PLN"), Quantity: kernel.NewQuantity(dec(t, "10"))},
			{Book: ledger.EquityBook(cur(t, "PLN")), Amount: money(t, "-1000.00", "PLN")},
		},
	}
	return &fakeLedger{
		accounts: []ledger.Account{{ID: "acc-1", Institution: "broker-a", Name: "Synthetic A", Sleeve: "core"}},
		txs:      []ledger.Transaction{buy, opening},
		holdings: []ledger.Holding{
			{Account: "acc-1", Instrument: bond, Date: on, Quantity: kernel.NewQuantity(dec(t, "10"))},
			{Account: "acc-1", Instrument: etf, Date: on, Quantity: kernel.NewQuantity(dec(t, "3"))},
		},
		lots: []ledger.Lot{{Instrument: bond, Opened: date(t, "2025-05-01"), Quantity: kernel.NewQuantity(dec(t, "10")), Cost: money(t, "1000.00", "PLN"), Opening: true, Origin: opening.Key}},
		cash: []kernel.Money{money(t, "100.00", "USD")},
		checks: []ledgerapp.AssertionCheck{
			{Assertion: ledger.CashAssertion("acc-1", on, money(t, "100.00", "USD")), Status: ledgerapp.AssertionMatched, Derived: dec(t, "100.00")},
			{Assertion: ledger.HoldingAssertion("acc-1", on, etf, kernel.NewQuantity(dec(t, "4"))), Status: ledgerapp.AssertionFlagged, Derived: dec(t, "3")},
		},
	}
}

// newViews returns the views over l with a bond and an ETF valued on
// 2026-03-02 and USD at 4.00 PLN.
func newViews(t *testing.T, l *fakeLedger) *app.Views {
	t.Helper()
	on := date(t, "2026-03-02")
	vals := fakeValuations{
		{Account: "acc-1", Instrument: instrument(t, isinPL, "EDO0334"), Date: on, Basis: valuation.NetIfRedeemed, Source: "statutory", Value: money(t, "1043.21", "PLN"), PLN: money(t, "1043.21", "PLN")},
		{Account: "acc-1", Instrument: instrument(t, isinIE, "ETF1"), Date: on, Basis: valuation.Market, Source: "test-prices", Value: money(t, "31.5", "USD"), PLN: money(t, "126.00", "PLN")},
	}
	rates := fakeRates{{Currency: cur(t, "USD"), Date: on, Rate: dec(t, "4.00"), Source: "test"}}
	v, err := app.New(l, vals, fakeStatuses{through: on}, rates, cur(t, "PLN"))
	if err != nil {
		t.Fatalf("app.New(PLN): %v", err)
	}
	return v
}
