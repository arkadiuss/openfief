package pl_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

func TestSaleMonth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, series, want string
	}{
		{name: "EDO: ten years before maturity", series: "EDO1136", want: "2026-11-01"},
		{name: "OTS: three months, across a year", series: "OTS0227", want: "2026-11-01"},
		{name: "ROR: one year", series: "ROR0124", want: "2023-01-01"},
		{name: "DOR: two years", series: "DOR0628", want: "2026-06-01"},
		{name: "TOS: three years", series: "TOS0927", want: "2024-09-01"},
		{name: "COI: four years", series: "COI0928", want: "2024-09-01"},
		{name: "ROS: six years", series: "ROS0930", want: "2024-09-01"},
		{name: "ROD: twelve years", series: "ROD0936", want: "2024-09-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.SaleMonth(series(t, tt.series))
			if err != nil {
				t.Fatalf("SaleMonth: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("SaleMonth(%s) = %s, want %s", tt.series, got, tt.want)
			}
		})
	}
}

// TestEarlyRedemptionFee looks fees up by each series' sale month, on both
// sides of each change the table records and at its edges.
func TestEarlyRedemptionFee(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, series, want string
		wantErr            error
	}{
		{name: "OTS has none", series: "OTS0118", want: "0.00"},
		{name: "OTS sold in the last reviewed month", series: "OTS0227", want: "0.00"},
		{name: "ROR", series: "ROR1127", want: "0.50"},
		{name: "DOR", series: "DOR0624", want: "0.70"},
		{name: "TOS before September 2024", series: "TOS0827", want: "0.70"},
		{name: "TOS from September 2024", series: "TOS0927", want: "1.00"},
		{name: "COI in 2014", series: "COI0118", want: "1.00"},
		{name: "COI the month before the 2015 change", series: "COI0319", want: "1.00"},
		{name: "COI from April 2015", series: "COI0419", want: "0.70"},
		{name: "COI the month before the 2024 change", series: "COI0828", want: "0.70"},
		{name: "COI from September 2024", series: "COI0928", want: "2.00"},
		{name: "EDO before September 2024", series: "EDO0834", want: "2.00"},
		{name: "EDO from September 2024", series: "EDO0934", want: "3.00"},
		{name: "ROS before September 2024", series: "ROS0830", want: "0.70"},
		{name: "ROS from September 2024", series: "ROS0930", want: "2.00"},
		{name: "ROD before September 2024", series: "ROD0836", want: "2.00"},
		{name: "ROD from September 2024", series: "ROD1138", want: "3.00"},
		{name: "before the type's first row", series: "ROR0522", wantErr: pl.ErrNoReviewedFee},
		{name: "EDO before 2014", series: "EDO1223", wantErr: pl.ErrNoReviewedFee},
		{name: "after the reviewed month", series: "EDO1236", wantErr: pl.ErrNoReviewedFee},
		{name: "OTS after the reviewed month", series: "OTS0327", wantErr: pl.ErrNoReviewedFee},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.EarlyRedemptionFee(series(t, tt.series))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("EarlyRedemptionFee(%s) error (-want +got):\n%s", tt.series, diff)
			}
			if tt.wantErr != nil {
				return
			}
			if want := plnMoney(t, tt.want); !got.Equal(want) {
				t.Errorf("EarlyRedemptionFee(%s) = %s, want %s", tt.series, got, want)
			}
		})
	}
}

// TestEarlyRedemptionFeesTable checks the table's shape: PLN fees of zero
// or more, rows of each type in order of their first day, every row on the
// first of a month no later than the reviewed month, each naming its letter,
// and every bond type present.
func TestEarlyRedemptionFeesTable(t *testing.T) {
	t.Parallel()
	last := map[pl.BondType]pl.FeeRow{}
	for _, r := range pl.EarlyRedemptionFees() {
		for _, problem := range feeRowProblems(&r, last) {
			t.Errorf("%s from %s: %s", r.Type, r.From, problem)
		}
		last[r.Type] = r
	}
	for _, typ := range []pl.BondType{pl.OTS, pl.ROR, pl.DOR, pl.TOS, pl.COI, pl.EDO, pl.ROS, pl.ROD} {
		if _, ok := last[typ]; !ok {
			t.Errorf("no row for %s", typ)
		}
	}
}

// feeRowProblems returns what is wrong with row r of the fee table, given
// the last row read of each type.
func feeRowProblems(r *pl.FeeRow, last map[pl.BondType]pl.FeeRow) []string {
	var out []string
	if !r.Fee.Currency().Equal(pl.Nominal().Currency()) || r.Fee.Sign() < 0 {
		out = append(out, "the fee is not a PLN amount of zero or more")
	}
	if r.From.IsZero() || r.From.Day() != 1 || r.From.After(pl.FeesReviewedThrough()) {
		out = append(out, "not the first of a month up to the reviewed month")
	}
	if r.Letter == "" {
		out = append(out, "names no issue letter")
	}
	if prev, ok := last[r.Type]; ok && !prev.From.Before(r.From) {
		out = append(out, "follows a row from "+prev.From.String())
	}
	return out
}
