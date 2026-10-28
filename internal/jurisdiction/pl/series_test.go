package pl_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

func TestParseSeries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		code    string
		want    pl.Series
		wantErr error
	}{
		{name: "EDO0536 is EDO maturing May 2036", code: "EDO0536", want: pl.Series{Type: pl.EDO, MaturityYear: 2036, MaturityMonth: time.May}},
		{name: "COI", code: "COI0129", want: pl.Series{Type: pl.COI, MaturityYear: 2029, MaturityMonth: time.January}},
		{name: "ROR", code: "ROR1226", want: pl.Series{Type: pl.ROR, MaturityYear: 2026, MaturityMonth: time.December}},
		{name: "DOR", code: "DOR0227", want: pl.Series{Type: pl.DOR, MaturityYear: 2027, MaturityMonth: time.February}},
		{name: "TOS", code: "TOS0227", want: pl.Series{Type: pl.TOS, MaturityYear: 2027, MaturityMonth: time.February}},
		{name: "OTS", code: "OTS0825", want: pl.Series{Type: pl.OTS, MaturityYear: 2025, MaturityMonth: time.August}},
		{name: "ROS", code: "ROS0331", want: pl.Series{Type: pl.ROS, MaturityYear: 2031, MaturityMonth: time.March}},
		{name: "ROD", code: "ROD0337", want: pl.Series{Type: pl.ROD, MaturityYear: 2037, MaturityMonth: time.March}},
		{name: "too short", code: "EDO536", wantErr: pl.ErrInvalidSeries},
		{name: "unknown type", code: "XYZ0536", wantErr: pl.ErrInvalidSeries},
		{name: "lower case", code: "edo0536", wantErr: pl.ErrInvalidSeries},
		{name: "month 13", code: "EDO1336", wantErr: pl.ErrInvalidSeries},
		{name: "month 00", code: "EDO0036", wantErr: pl.ErrInvalidSeries},
		{name: "sign in year", code: "EDO05+6", wantErr: pl.ErrInvalidSeries},
		{name: "empty", code: "", wantErr: pl.ErrInvalidSeries},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.ParseSeries(tt.code)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseSeries(%q) error (-want +got):\n%s", tt.code, diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseSeries(%q) (-want +got):\n%s", tt.code, diff)
			}
			if got.String() != tt.code {
				t.Errorf("String() = %q, want %q", got.String(), tt.code)
			}
		})
	}
}

func TestBondTypeStructure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		typ  pl.BondType
		want pl.Structure
	}{
		{pl.OTS, pl.Structure{TenorMonths: 3, PeriodMonths: 3, Indexation: pl.Fixed, Payout: pl.Capitalized}},
		{pl.ROR, pl.Structure{TenorMonths: 12, PeriodMonths: 1, Indexation: pl.ReferenceRate, Payout: pl.Coupon}},
		{pl.DOR, pl.Structure{TenorMonths: 24, PeriodMonths: 1, Indexation: pl.ReferenceRate, Payout: pl.Coupon}},
		{pl.TOS, pl.Structure{TenorMonths: 36, PeriodMonths: 12, Indexation: pl.Fixed, Payout: pl.Capitalized}},
		{pl.COI, pl.Structure{TenorMonths: 48, PeriodMonths: 12, Indexation: pl.Inflation, Payout: pl.Coupon}},
		{pl.EDO, pl.Structure{TenorMonths: 120, PeriodMonths: 12, Indexation: pl.Inflation, Payout: pl.Capitalized}},
		{pl.ROS, pl.Structure{TenorMonths: 72, PeriodMonths: 12, Indexation: pl.Inflation, Payout: pl.Capitalized}},
		{pl.ROD, pl.Structure{TenorMonths: 144, PeriodMonths: 12, Indexation: pl.Inflation, Payout: pl.Capitalized}},
	}
	for _, tt := range tests {
		t.Run(tt.typ.String(), func(t *testing.T) {
			t.Parallel()
			got, err := tt.typ.Structure()
			if err != nil {
				t.Fatalf("Structure(): %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()
		_, err := pl.BondType(99).Structure()
		if diff := cmp.Diff(pl.ErrUnknownBondType, err, cmpopts.EquateErrors()); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
		if got := pl.BondType(99).String(); got != "BondType(99)" {
			t.Errorf("String() = %q", got)
		}
	})
}
