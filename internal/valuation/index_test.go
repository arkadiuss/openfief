package valuation_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/valuation"
)

func TestIndexKindNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    valuation.IndexKind
		wantErr error
	}{
		{in: "cpi", want: valuation.CPI},
		{in: "reference_rate", want: valuation.ReferenceRate},
		{in: "CPI", wantErr: valuation.ErrUnknownIndexKind},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.ParseIndexKind(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err == nil && (got != tt.want || got.String() != tt.in) {
				t.Errorf("ParseIndexKind(%q) = %v", tt.in, got)
			}
		})
	}
	if got := valuation.IndexKind(7).String(); got != "IndexKind(7)" {
		t.Errorf("IndexKind(7).String() = %q", got)
	}
}

func TestLotIndices(t *testing.T) {
	t.Parallel()
	readings := []valuation.IndexReading{
		{Kind: valuation.CPI, Year: 2025, Month: time.January, Value: dec(t, "0.049"), Source: "gus"},
		{Kind: valuation.CPI, Year: 2026, Month: time.January, Value: dec(t, "0.029"), Source: "gus"},
		{Kind: valuation.ReferenceRate, Year: 2026, Month: time.April, Value: dec(t, "0.0575"), Source: "nbp"},
		{Kind: valuation.CPI, Year: 2026, Month: time.February, Value: dec(t, "0.5"), Source: "gus"},
	}
	tests := []struct {
		name      string
		series    string
		purchased string
		on        string
		want      pl.Indices
	}{
		{name: "first period needs no reading", series: "EDO0334", purchased: "2024-03-15", on: "2025-03-14", want: pl.Indices{}},
		{name: "second EDO period takes January CPI", series: "EDO0334", purchased: "2024-03-15", on: "2025-03-15", want: pl.Indices{2: dec(t, "0.049")}},
		{name: "third period adds the next January", series: "EDO0334", purchased: "2024-03-15", on: "2026-06-01", want: pl.Indices{2: dec(t, "0.049"), 3: dec(t, "0.029")}},
		{name: "fixed-rate TOS needs none", series: "TOS0327", purchased: "2024-03-15", on: "2026-06-01", want: pl.Indices{}},
		{name: "ROR takes the reading of the month the period starts", series: "ROR0327", purchased: "2026-03-10", on: "2026-04-10", want: pl.Indices{2: dec(t, "0.0575")}},
		{name: "a reading not supplied is left out", series: "ROR0327", purchased: "2026-03-10", on: "2026-05-10", want: pl.Indices{2: dec(t, "0.0575")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := pl.ParseSeries(tt.series)
			if err != nil {
				t.Fatal(err)
			}
			got, err := valuation.LotIndices(s, date(t, tt.purchased), date(t, tt.on), readings)
			if err != nil {
				t.Fatalf("LotIndices: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
	if _, err := valuation.LotIndices(pl.Series{}, date(t, "2026-01-01"), date(t, "2026-01-01"), nil); !cmp.Equal(pl.ErrUnknownBondType, err, cmpopts.EquateErrors()) {
		t.Errorf("unknown bond type: error = %v", err)
	}
}
