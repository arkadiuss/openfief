package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// fakeBondFees keeps one bond's terms, when set, and records the fee saved
// with its revaluations.
type fakeBondFees struct {
	isin   ledger.ISIN
	terms  *pl.Terms
	saved  string
	revals []valuation.Revaluation
}

func (f *fakeBondFees) BondTerms(_ context.Context, isin ledger.ISIN) (pl.Terms, bool, error) {
	if f.terms == nil || isin != f.isin {
		return pl.Terms{}, false, nil
	}
	return *f.terms, true, nil
}

func (f *fakeBondFees) SaveBondFee(_ context.Context, _ ledger.ISIN, fee kernel.Money, source pl.FeeSource, revals []valuation.Revaluation) error {
	f.saved, f.revals = fee.String()+" "+source.String(), revals
	return nil
}

// TestSetEarlyRedemptionFee: the owner confirms an assumed fee, which
// revalues nothing, or overrides any fee, which revalues the accounts that
// held the series when the amount changes (D-97); confirming a fee that
// isn't assumed, a bond without terms, a malformed fee or a read session
// change nothing.
func TestSetEarlyRedemptionFee(t *testing.T) {
	t.Parallel()
	series, err := pl.ParseSeries("EDO0336")
	if err != nil {
		t.Fatal(err)
	}
	fee, err := kernel.ParseMoney("3.00", pl.Nominal().Currency())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, token, isin, fee string
		source                 pl.FeeSource
		wantSaved              string
		wantRevals             []valuation.Revaluation
		wantErr                error
	}{
		{name: "confirm an assumed fee", token: ownerToken, isin: bondISIN, source: pl.FeeAssumed, wantSaved: "3 PLN entered"},
		{
			name: "override a table's fee", token: ownerToken, isin: bondISIN, fee: "2.50", source: pl.FeeFromTable,
			wantSaved: "2.5 PLN entered", wantRevals: revals(t, "acc-1", "2026-03-02"),
		},
		{name: "override with the same amount", token: ownerToken, isin: bondISIN, fee: "3", source: pl.FeeAssumed, wantSaved: "3 PLN entered"},
		{name: "confirm a fee that isn't assumed", token: ownerToken, isin: bondISIN, source: pl.FeeFromTable, wantErr: pl.ErrFeeNotAssumed},
		{name: "a bond without terms", token: ownerToken, isin: otherISIN, source: pl.FeeAssumed, wantErr: app.ErrUnknownBond},
		{name: "a negative fee", token: ownerToken, isin: bondISIN, fee: "-1", source: pl.FeeAssumed, wantErr: app.ErrInvalidInput},
		{name: "a fee that isn't a number", token: ownerToken, isin: bondISIN, fee: "3 zł", source: pl.FeeAssumed, wantErr: app.ErrInvalidInput},
		{name: "a malformed ISIN", token: ownerToken, isin: "EDO0336", source: pl.FeeAssumed, wantErr: app.ErrInvalidInput},
		{name: "a read session", token: "read-session", isin: bondISIN, source: pl.FeeAssumed, wantErr: app.ErrNotOwner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			isin, err := ledger.ParseISIN(bondISIN)
			if err != nil {
				t.Fatal(err)
			}
			store := &fakeBondFees{isin: isin, terms: &pl.Terms{Series: series, EarlyRedemptionFee: fee, FeeSource: tt.source}}
			s := newSettings(t, newFakeStore(), &fakeRecorder{}).WithBondFees(store)
			err = s.SetEarlyRedemptionFee(t.Context(), "o", tt.token, tt.isin, tt.fee)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if store.saved != tt.wantSaved {
				t.Errorf("saved %q, want %q", store.saved, tt.wantSaved)
			}
			if diff := cmp.Diff(tt.wantRevals, store.revals, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("revaluations (-want +got):\n%s", diff)
			}
		})
	}
}
