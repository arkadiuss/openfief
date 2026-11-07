package pl_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
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
// sides of each change the table records and at its edges, and carries the
// latest reviewed fee forward, as assumed, past the reviewed month.
func TestEarlyRedemptionFee(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, series, want string
		source             pl.FeeSource
		letter             string
		wantErr            error
	}{
		{name: "OTS has none", series: "OTS0118", want: "0.00", source: pl.FeeFromTable},
		{name: "OTS sold in the last reviewed month", series: "OTS0227", want: "0.00", source: pl.FeeFromTable},
		{name: "ROR", series: "ROR1127", want: "0.50", source: pl.FeeFromTable},
		{name: "DOR", series: "DOR0624", want: "0.70", source: pl.FeeFromTable},
		{name: "TOS before September 2024", series: "TOS0827", want: "0.70", source: pl.FeeFromTable},
		{name: "TOS from September 2024", series: "TOS0927", want: "1.00", source: pl.FeeFromTable},
		{name: "COI in 2014", series: "COI0118", want: "1.00", source: pl.FeeFromTable},
		{name: "COI the month before the 2015 change", series: "COI0319", want: "1.00", source: pl.FeeFromTable},
		{name: "COI from April 2015", series: "COI0419", want: "0.70", source: pl.FeeFromTable},
		{name: "COI the month before the 2024 change", series: "COI0828", want: "0.70", source: pl.FeeFromTable},
		{name: "COI from September 2024", series: "COI0928", want: "2.00", source: pl.FeeFromTable},
		{name: "EDO before September 2024", series: "EDO0834", want: "2.00", source: pl.FeeFromTable},
		{name: "EDO from September 2024", series: "EDO0934", want: "3.00", source: pl.FeeFromTable, letter: "list emisyjny nr 96/2024 (EDO0934)"},
		{name: "EDO in the last reviewed month", series: "EDO1136", want: "3.00", source: pl.FeeFromTable},
		{name: "ROS before September 2024", series: "ROS0830", want: "0.70", source: pl.FeeFromTable},
		{name: "ROS from September 2024", series: "ROS0930", want: "2.00", source: pl.FeeFromTable},
		{name: "ROD before September 2024", series: "ROD0836", want: "2.00", source: pl.FeeFromTable},
		{name: "ROD from September 2024", series: "ROD1138", want: "3.00", source: pl.FeeFromTable},
		{name: "after the reviewed month: assumed", series: "EDO1236", want: "3.00", source: pl.FeeAssumed, letter: "list emisyjny nr 96/2024 (EDO0934)"},
		{name: "a year after the reviewed month: assumed", series: "EDO1137", want: "3.00", source: pl.FeeAssumed},
		{name: "OTS after the reviewed month: assumed", series: "OTS0327", want: "0.00", source: pl.FeeAssumed},
		{name: "before the type's first row", series: "ROR0522", wantErr: pl.ErrNoReviewedFee},
		{name: "EDO before 2014", series: "EDO1223", wantErr: pl.ErrNoReviewedFee},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := series(t, tt.series)
			got, err := pl.EarlyRedemptionFee(s)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("EarlyRedemptionFee(%s) error (-want +got):\n%s", tt.series, diff)
			}
			if tt.wantErr == nil {
				for _, problem := range tableFeeProblems(&got, s, plnMoney(t, tt.want), tt.source, tt.letter) {
					t.Errorf("EarlyRedemptionFee(%s): %s", tt.series, problem)
				}
			}
		})
	}
}

// tableFeeProblems returns what is wrong with got, series s's fee from the
// table: its fee, source and letter, and its row's fee and type.
func tableFeeProblems(got *pl.TableFee, s pl.Series, fee kernel.Money, source pl.FeeSource, letter string) []string {
	var out []string
	if !got.Fee.Equal(fee) || got.Source != source {
		out = append(out, fmt.Sprintf("%s %s, want %s %s", got.Fee, got.Source, fee, source))
	}
	if !got.Fee.Equal(got.Row.Fee) || got.Row.Type != s.Type || got.Row.Letter == "" {
		out = append(out, fmt.Sprintf("row %+v doesn't carry the fee", got.Row))
	}
	if letter != "" && got.Row.Letter != letter {
		out = append(out, fmt.Sprintf("letter %q, want %q", got.Row.Letter, letter))
	}
	return out
}

// terms is series code's terms with fee in PLN from source.
func terms(t *testing.T, code, fee string, source pl.FeeSource) pl.Terms {
	t.Helper()
	return pl.Terms{Series: series(t, code), EarlyRedemptionFee: plnMoney(t, fee), FeeSource: source}
}

// TestFeesToConfirm keeps the held series whose fee is assumed, in series
// order, and drops the table's and the owner's.
func TestFeesToConfirm(t *testing.T) {
	t.Parallel()
	held := []pl.Terms{
		terms(t, "ROD1238", "3.00", pl.FeeAssumed),
		terms(t, "EDO1136", "3.00", pl.FeeFromTable),
		terms(t, "EDO1137", "3.00", pl.FeeAssumed),
		terms(t, "COI1230", "2.50", pl.FeeEntered),
	}
	var got []string
	for _, tt := range pl.FeesToConfirm(held) {
		got = append(got, tt.Series.String())
	}
	if diff := cmp.Diff([]string{"EDO1137", "ROD1238"}, got); diff != "" {
		t.Errorf("FeesToConfirm (-want +got):\n%s", diff)
	}
	if got := pl.FeesToConfirm(nil); got != nil {
		t.Errorf("FeesToConfirm(nil) = %v, want none", got)
	}
}

// TestConfirmFee keeps an assumed fee as the owner's and refuses any
// other.
func TestConfirmFee(t *testing.T) {
	t.Parallel()
	assumed := terms(t, "EDO1137", "3.00", pl.FeeAssumed)
	confirmed, err := assumed.ConfirmFee()
	if err != nil || confirmed.FeeSource != pl.FeeEntered || !confirmed.EarlyRedemptionFee.Equal(assumed.EarlyRedemptionFee) {
		t.Errorf("ConfirmFee = %+v, %v; want the same fee, entered", confirmed, err)
	}
	for _, src := range []pl.FeeSource{pl.FeeEntered, pl.FeeFromTable} {
		tt := terms(t, "EDO1136", "3.00", src)
		if _, err := tt.ConfirmFee(); !errors.Is(err, pl.ErrFeeNotAssumed) {
			t.Errorf("ConfirmFee of a %s fee: %v, want ErrFeeNotAssumed", src, err)
		}
	}
}

// TestOverrideFee makes any fee the owner's and refuses a negative or
// non-PLN one.
func TestOverrideFee(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, amount, currency string
		wantErr                error
	}{
		{name: "a new fee", amount: "2.50", currency: "PLN"},
		{name: "no fee", amount: "0", currency: "PLN"},
		{name: "negative", amount: "-1", currency: "PLN", wantErr: pl.ErrInvalidTerms},
		{name: "not PLN", amount: "1", currency: "EUR", wantErr: pl.ErrInvalidTerms},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assumed := terms(t, "EDO1137", "3.00", pl.FeeAssumed)
			cur, err := kernel.ParseCurrency(tt.currency)
			if err != nil {
				t.Fatal(err)
			}
			fee, err := kernel.ParseMoney(tt.amount, cur)
			if err != nil {
				t.Fatal(err)
			}
			got, err := assumed.OverrideFee(fee)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("OverrideFee error (-want +got):\n%s", diff)
			}
			if err == nil && (got.FeeSource != pl.FeeEntered || !got.EarlyRedemptionFee.Equal(fee) || got.Series != assumed.Series) {
				t.Errorf("OverrideFee = %+v, want %s entered", got, fee)
			}
		})
	}
}

// TestAssumedFeeDetail names the series, the fee and the letter it was
// carried from, and leaves out a letter whose fee isn't the one stored.
func TestAssumedFeeDetail(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, series, fee, want string
	}{
		{
			name: "the table's latest EDO fee", series: "EDO1137", fee: "3.00",
			want: "confirm the early-redemption fee for EDO1137: 3.00 PLN assumed from the latest reviewed letter, list emisyjny nr 96/2024 (EDO0934)",
		},
		{
			name: "a fee the table no longer has", series: "EDO1137", fee: "2.00",
			want: "confirm the early-redemption fee for EDO1137: 2.00 PLN assumed from the latest reviewed letter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := terms(t, tt.series, tt.fee, pl.FeeAssumed)
			if d := pl.AssumedFeeDetail(&got); d != tt.want {
				t.Errorf("AssumedFeeDetail = %q, want %q", d, tt.want)
			}
		})
	}
}

// TestParseFeeSource reads back every source's name and refuses others.
func TestParseFeeSource(t *testing.T) {
	t.Parallel()
	for _, s := range []pl.FeeSource{pl.FeeEntered, pl.FeeFromTable, pl.FeeAssumed} {
		got, err := pl.ParseFeeSource(s.String())
		if err != nil || got != s {
			t.Errorf("ParseFeeSource(%q) = %v, %v", s, got, err)
		}
	}
	if _, err := pl.ParseFeeSource("guessed"); !errors.Is(err, pl.ErrUnknownFeeSource) {
		t.Errorf("ParseFeeSource(guessed): %v, want ErrUnknownFeeSource", err)
	}
	if got := pl.FeeSource(9).String(); got != "FeeSource(9)" {
		t.Errorf("FeeSource(9).String() = %q", got)
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
