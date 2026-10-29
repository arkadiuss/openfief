package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Synthetic instruments: test ISINs with valid check digits.
const (
	etfIE = "IETEST000026/ETF1"
	etfPL = "PLTEST000010/TST"
	owner = "ofsess_owner"
)

var errStoreDown = errors.New("store down")

// fakeStore keeps one model in memory.
type fakeStore struct {
	model portfolio.Model
	saves int
	err   error
}

func (f *fakeStore) Model(context.Context) (portfolio.Model, error) { return f.model, f.err }

func (f *fakeStore) SaveModel(_ context.Context, m portfolio.Model) error {
	if f.err != nil {
		return f.err
	}
	f.saves++
	f.model = m
	return nil
}

// fakeMarket serves stored history from maps, counting reads.
type fakeMarket struct {
	prices map[string][]valuation.Price
	levels map[string][]valuation.IndexLevel
	rates  map[kernel.Currency][]valuation.FXRate
	reads  int
	err    error
}

func (f *fakeMarket) Prices(_ context.Context, instrument string, through kernel.Date) ([]valuation.Price, error) {
	f.reads++
	var out []valuation.Price
	for _, p := range f.prices[instrument] {
		if !p.Date.After(through) {
			out = append(out, p)
		}
	}
	return out, f.err
}

func (f *fakeMarket) IndexLevels(_ context.Context, index string, through kernel.Date) ([]valuation.IndexLevel, error) {
	f.reads++
	var out []valuation.IndexLevel
	for _, l := range f.levels[index] {
		if !l.Date.After(through) {
			out = append(out, l)
		}
	}
	return out, f.err
}

func (f *fakeMarket) FXRates(_ context.Context, c kernel.Currency, _ kernel.Date) ([]valuation.FXRate, error) {
	f.reads++
	return f.rates[c], f.err
}

// fakeOwners accepts the token owner and answers err for any other.
type fakeOwners struct {
	err error
}

func (f fakeOwners) CheckOwner(_ context.Context, _, token string) error {
	if token == owner {
		return nil
	}
	return f.err
}

// mustDec parses a decimal or fails the test.
func mustDec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// mustDate parses a date or fails the test.
func mustDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// mustCurrency parses a currency or fails the test.
func mustCurrency(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// newModels returns the use cases over the fakes, reporting in PLN, with
// owners refusing other tokens with ErrNoSession.
func newModels(t *testing.T, store *fakeStore, market *fakeMarket) *app.Models {
	t.Helper()
	m, err := app.New(store, market, fakeOwners{err: app.ErrNoSession}, mustCurrency(t, "PLN"))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
