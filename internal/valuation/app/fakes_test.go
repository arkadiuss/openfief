package app_test

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

var errStoreDown = errors.New("synthetic store failure")

// Synthetic accounts and ISINs with valid check digits; nothing real is
// implied.
const (
	acc1 ledger.AccountID = "acc-1"
	acc2 ledger.AccountID = "acc-2"

	etfISIN   = "IETEST000026"
	otherISIN = "PLTEST000010"
	bondISIN  = "PL0000000014"
)

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

func money(t *testing.T, amount, code string) kernel.Money {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	return kernel.NewMoney(dec(t, amount), c)
}

func instrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

// buy returns version 1 of a buy of quantity units of in for cost in
// account on day, with no fee.
func buy(t *testing.T, account ledger.AccountID, id, day string, in ledger.Instrument, quantity, cost string) ledger.VersionedTransaction {
	t.Helper()
	c := money(t, cost, "PLN")
	if in.ISIN.String() == etfISIN {
		c = money(t, cost, "USD")
	}
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: account, Date: date(t, day)}
	tx, err := ledger.NewBuy(h, in, kernel.NewQuantity(dec(t, quantity)), c, kernel.NewMoney(kernel.Decimal{}, c.Currency()))
	if err != nil {
		t.Fatalf("NewBuy: %v", err)
	}
	return ledger.VersionedTransaction{Transaction: tx, Version: 1}
}

// fakeLedger is the ledger's accounts and logs in memory.
type fakeLedger struct {
	logs map[ledger.AccountID][]ledger.VersionedTransaction
	err  error
}

func (f *fakeLedger) Log(_ context.Context, account ledger.AccountID) ([]ledger.VersionedTransaction, error) {
	return f.logs[account], f.err
}

func (f *fakeLedger) Accounts(context.Context) ([]ledger.Account, error) {
	out := make([]ledger.Account, 0, len(f.logs))
	for id := range f.logs {
		out = append(out, ledger.Account{ID: id})
	}
	slices.SortFunc(out, func(a, b ledger.Account) int { return cmp.Compare(a.ID, b.ID) })
	return out, f.err
}

// fakeStore keeps inputs and valuations in memory.
type fakeStore struct {
	prices   []valuation.Price
	rates    []valuation.FXRate
	readings []valuation.IndexReading
	terms    map[ledger.ISIN]pl.Terms
	wrappers map[ledger.AccountID]pl.TaxWrapper
	rows     map[ledger.AccountID][]valuation.HoldingValuation
	through  map[ledger.AccountID]kernel.Date
	err      error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		terms: map[ledger.ISIN]pl.Terms{}, wrappers: map[ledger.AccountID]pl.TaxWrapper{},
		rows: map[ledger.AccountID][]valuation.HoldingValuation{}, through: map[ledger.AccountID]kernel.Date{},
	}
}

func (f *fakeStore) Prices(_ context.Context, instrument string, through kernel.Date) ([]valuation.Price, error) {
	var out []valuation.Price
	for _, p := range f.prices {
		if p.Instrument == instrument && !p.Date.After(through) {
			out = append(out, p)
		}
	}
	return out, f.err
}

func (f *fakeStore) FXRates(_ context.Context, c kernel.Currency, through kernel.Date) ([]valuation.FXRate, error) {
	var out []valuation.FXRate
	for _, r := range f.rates {
		if r.Currency.Equal(c) && !r.Date.After(through) {
			out = append(out, r)
		}
	}
	return out, f.err
}

func (f *fakeStore) IndexReadings(context.Context) ([]valuation.IndexReading, error) {
	return f.readings, f.err
}

func (f *fakeStore) BondTerms(_ context.Context, isin ledger.ISIN) (pl.Terms, bool, error) {
	t, ok := f.terms[isin]
	return t, ok, f.err
}

func (f *fakeStore) TaxWrapper(_ context.Context, account ledger.AccountID) (pl.TaxWrapper, error) {
	if w, ok := f.wrappers[account]; ok {
		return w, f.err
	}
	return pl.Taxable, f.err
}

func (f *fakeStore) ValuedThrough(_ context.Context, account ledger.AccountID) (kernel.Date, bool, error) {
	d, ok := f.through[account]
	return d, ok, f.err
}

func (f *fakeStore) ReplaceValuations(_ context.Context, account ledger.AccountID, from kernel.Date, rows []valuation.HoldingValuation, through kernel.Date) error {
	if f.err != nil {
		return f.err
	}
	kept := slices.DeleteFunc(slices.Clone(f.rows[account]), func(r valuation.HoldingValuation) bool { return !r.Date.Before(from) })
	f.rows[account] = append(kept, rows...)
	f.through[account] = through
	return nil
}

// recorded is one call to fakeRecorder.
type recorded struct {
	what   string
	revals []valuation.Revaluation
}

// fakeRecorder records what it was asked to store and enqueue.
type fakeRecorder struct {
	calls   []recorded
	changed bool
	err     error
}

func (f *fakeRecorder) SavePrices(_ context.Context, _ []valuation.Price, revals []valuation.Revaluation) error {
	return f.record("prices", revals)
}

func (f *fakeRecorder) SaveFXRates(_ context.Context, _ []valuation.FXRate, revals []valuation.Revaluation) error {
	return f.record("fx", revals)
}

func (f *fakeRecorder) SaveIndexReadings(_ context.Context, _ []valuation.IndexReading, revals []valuation.Revaluation) error {
	return f.record("index", revals)
}

func (f *fakeRecorder) SaveBondTerms(_ context.Context, _ ledger.ISIN, _ *pl.Terms, revals []valuation.Revaluation) error {
	return f.record("terms", revals)
}

func (f *fakeRecorder) SaveTaxWrapper(_ context.Context, _ ledger.AccountID, _ pl.TaxWrapper, revals []valuation.Revaluation) error {
	return f.record("wrapper", revals)
}

func (f *fakeRecorder) SaveBondBasis(_ context.Context, basis valuation.Basis, revals []valuation.Revaluation) (bool, error) {
	return f.changed, f.record("basis "+basis.String(), revals)
}

func (f *fakeRecorder) Enqueue(_ context.Context, revals []valuation.Revaluation) error {
	return f.record("enqueue", revals)
}

func (f *fakeRecorder) record(what string, revals []valuation.Revaluation) error {
	f.calls = append(f.calls, recorded{what: what, revals: revals})
	return f.err
}
