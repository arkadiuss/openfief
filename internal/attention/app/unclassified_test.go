package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// The synthetic ISINs the unclassified-instrument check sees: two funds and a
// treasury bond.
const (
	fundA = "PLTEST000010"
	fundB = "PLTEST000028"
	bond  = "PLEDO0536000"
)

// fakeHoldings is app.Holdings over fixed valuations and classes; an account
// without rows has never been revalued.
type fakeHoldings struct {
	rows    map[ledger.AccountID][]valuation.HoldingValuation
	classes map[string]jurisdiction.ExemptionClass
}

func (f *fakeHoldings) ValuedThrough(_ context.Context, a ledger.AccountID) (kernel.Date, bool, error) {
	rows := f.rows[a]
	if len(rows) == 0 {
		return kernel.Date{}, false, nil
	}
	return rows[len(rows)-1].Date, true, nil
}

func (f *fakeHoldings) Valuations(_ context.Context, a ledger.AccountID, from, to kernel.Date) ([]valuation.HoldingValuation, error) {
	var out []valuation.HoldingValuation
	rows := f.rows[a]
	for i := range rows {
		if !rows[i].Date.Before(from) && !rows[i].Date.After(to) {
			out = append(out, rows[i])
		}
	}
	return out, nil
}

func (f *fakeHoldings) ExemptionClasses(context.Context) (map[string]jurisdiction.ExemptionClass, error) {
	return f.classes, nil
}

// held returns a valued row of isin in account on date, priced at market,
// or at its accrued value when accrued.
func held(t *testing.T, account ledger.AccountID, isin, date string, accrued bool) valuation.HoldingValuation {
	t.Helper()
	parsed, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	basis := valuation.Market
	if accrued {
		basis = valuation.GrossAccrued
	}
	return valuation.HoldingValuation{
		Account: account, Instrument: ledger.Instrument{ISIN: parsed}, Date: day(t, date), Basis: basis,
		Value: pln(t, "100"), PLN: pln(t, "100"),
	}
}

// TestUnclassifiedInstruments checks D-148's check: an instrument priced at
// market held in an OKI account in the tax year it is valued through, which
// the pack can't classify and the owner hasn't, is one condition however
// many accounts hold it; a bond, a taxable account's holding, an earlier
// year's holding, an account never revalued and an instrument the owner
// classified are none.
func TestUnclassifiedInstruments(t *testing.T) {
	t.Parallel()
	const oki1, oki2, taxable, fresh = ledger.AccountID("oki-1"), ledger.AccountID("oki-2"), ledger.AccountID("taxable"), ledger.AccountID("fresh")
	detail := "the jurisdiction pack can't tell this instrument's exemption class for oki's yearly tax"
	tests := []struct {
		name    string
		classes map[string]jurisdiction.ExemptionClass
		want    []attention.Condition
	}{
		{
			name: "unclassified funds in OKI accounts, once each",
			want: []attention.Condition{
				{Kind: attention.KindInstrumentUnclassified, Subject: fundA, Detail: detail},
				{Kind: attention.KindInstrumentUnclassified, Subject: fundB, Detail: detail},
			},
		},
		{
			name:    "the owner classified one",
			classes: map[string]jurisdiction.ExemptionClass{fundA: pl.Investment},
			want:    []attention.Condition{{Kind: attention.KindInstrumentUnclassified, Subject: fundB, Detail: detail}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &fakeLedger{accounts: []ledger.Account{{ID: fresh}, {ID: oki1}, {ID: oki2}, {ID: taxable}}}
			v := &fakeValues{wrappers: map[ledger.AccountID]jurisdiction.Wrapper{fresh: "oki", oki1: "oki", oki2: "oki", taxable: "taxable"}}
			h := &fakeHoldings{classes: tt.classes, rows: map[ledger.AccountID][]valuation.HoldingValuation{
				oki1: {
					held(t, oki1, fundB, "2027-12-30", false), held(t, oki1, bond, "2028-01-02", true),
					held(t, oki1, fundA, "2028-01-03", false),
				},
				oki2:    {held(t, oki2, fundA, "2028-03-01", false), held(t, oki2, fundB, "2028-03-02", false)},
				taxable: {held(t, taxable, "PLTEST000036", "2028-03-01", false)},
			}}
			c := app.UnclassifiedInstruments(l, v, h, pl.NewPack())
			got, err := c.Find(t.Context())
			if err != nil {
				t.Fatalf("Find: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Find (-want +got):\n%s", diff)
			}
		})
	}
}

// TestUnclassifiedResolvesItself checks that the item resolves once the owner
// classifies the instrument (D-148): the next sweep clears it.
func TestUnclassifiedResolvesItself(t *testing.T) {
	t.Parallel()
	const oki = ledger.AccountID("oki-1")
	h := &fakeHoldings{rows: map[ledger.AccountID][]valuation.HoldingValuation{oki: {held(t, oki, fundA, "2028-03-02", false)}}}
	check := app.UnclassifiedInstruments(&fakeLedger{accounts: []ledger.Account{{ID: oki}}},
		&fakeValues{wrappers: map[ledger.AccountID]jurisdiction.Wrapper{oki: "oki"}}, h, pl.NewPack())
	items := app.New(newMemStore(), fakeOwners{}, clock(), check)
	if err := items.Sweep(t.Context()); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if diff := cmp.Diff(map[string]string{fundA: "open"}, states(t, items)); diff != "" {
		t.Fatalf("before the owner classified it (-want +got):\n%s", diff)
	}
	h.classes = map[string]jurisdiction.ExemptionClass{fundA: pl.Savings}
	if err := items.Sweep(t.Context()); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if diff := cmp.Diff(map[string]string{fundA: "resolved"}, states(t, items)); diff != "" {
		t.Errorf("after the owner classified it (-want +got):\n%s", diff)
	}
}
