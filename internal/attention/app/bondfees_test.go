package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// The synthetic bond ISINs the assumed-fee check sees.
const (
	edoLater = "PLEDO1137006"
	edoTable = "PLEDO1136008"
	rodLater = "PLROD1238004"
)

// fakeBondTerms is app.BondTerms over fixed terms by ISIN.
type fakeBondTerms struct{ terms map[ledger.ISIN]pl.Terms }

func (f *fakeBondTerms) AllBondTerms(context.Context) (map[ledger.ISIN]pl.Terms, error) {
	return f.terms, nil
}

// bondTerms is series code's terms with a 3.00 PLN fee from source, keyed
// by isin.
func bondTerms(t *testing.T, terms map[ledger.ISIN]pl.Terms, isin, code string, source pl.FeeSource) {
	t.Helper()
	parsed, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	s, err := pl.ParseSeries(code)
	if err != nil {
		t.Fatal(err)
	}
	terms[parsed] = pl.Terms{Series: s, EarlyRedemptionFee: pln(t, "3.00"), FeeSource: source}
}

// TestAssumedBondFees checks D-97's check: each held series whose fee is
// assumed is one condition, its subject the ISIN, however many accounts
// hold it; a series with the table's fee, one no longer held and one
// without stored terms are none.
func TestAssumedBondFees(t *testing.T) {
	t.Parallel()
	const a, b = ledger.AccountID("bonds-1"), ledger.AccountID("bonds-2")
	terms := map[ledger.ISIN]pl.Terms{}
	bondTerms(t, terms, edoLater, "EDO1137", pl.FeeAssumed)
	bondTerms(t, terms, edoTable, "EDO1136", pl.FeeFromTable)
	bondTerms(t, terms, rodLater, "ROD1238", pl.FeeAssumed)
	h := &fakeHoldings{rows: map[ledger.AccountID][]valuation.HoldingValuation{
		a: {held(t, a, rodLater, "2027-11-30", true), held(t, a, edoLater, "2027-12-01", true), held(t, a, edoTable, "2027-12-01", true)},
		b: {held(t, b, edoLater, "2027-12-01", true), held(t, b, fundA, "2027-12-01", false)},
	}}
	l := &fakeLedger{accounts: []ledger.Account{{ID: a}, {ID: b}, {ID: "never-valued"}}}
	got, err := app.AssumedBondFees(l, h, &fakeBondTerms{terms: terms}).Find(t.Context())
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	want := []attention.Condition{{
		Kind: attention.KindBondFeeAssumed, Subject: edoLater,
		Detail: "confirm the early-redemption fee for EDO1137: 3.00 PLN assumed from the latest reviewed letter, list emisyjny nr 96/2024 (EDO0934)",
	}}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Find (-want +got):\n%s", diff)
	}
}

// TestAssumedBondFeeResolvesItself: the item resolves once the owner
// confirms the fee, which makes it the owner's (D-97).
func TestAssumedBondFeeResolvesItself(t *testing.T) {
	t.Parallel()
	const a = ledger.AccountID("bonds-1")
	terms := map[ledger.ISIN]pl.Terms{}
	bondTerms(t, terms, edoLater, "EDO1137", pl.FeeAssumed)
	h := &fakeHoldings{rows: map[ledger.AccountID][]valuation.HoldingValuation{a: {held(t, a, edoLater, "2027-12-01", true)}}}
	stored := &fakeBondTerms{terms: terms}
	items := app.New(newMemStore(), fakeOwners{}, clock(), app.AssumedBondFees(&fakeLedger{accounts: []ledger.Account{{ID: a}}}, h, stored))
	if err := items.Sweep(t.Context()); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if diff := cmp.Diff(map[string]string{edoLater: "open"}, states(t, items)); diff != "" {
		t.Fatalf("before the owner confirmed it (-want +got):\n%s", diff)
	}
	bondTerms(t, terms, edoLater, "EDO1137", pl.FeeEntered)
	if err := items.Sweep(t.Context()); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if diff := cmp.Diff(map[string]string{edoLater: "resolved"}, states(t, items)); diff != "" {
		t.Errorf("after the owner confirmed it (-want +got):\n%s", diff)
	}
}
