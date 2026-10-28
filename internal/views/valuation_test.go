package views_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views"
)

// total is a wanted DateTotal: pln in PLN, then amount/currency pairs.
type total struct {
	day, pln   string
	byCurrency [][2]string
}

// totals builds the wanted DateTotals.
func totals(t *testing.T, ts []total) []views.DateTotal {
	t.Helper()
	out := make([]views.DateTotal, 0, len(ts))
	for _, w := range ts {
		dt := views.DateTotal{Date: date(t, w.day), Reporting: money(t, w.pln, "PLN"), ByCurrency: []kernel.Money{}}
		for _, c := range w.byCurrency {
			dt.ByCurrency = append(dt.ByCurrency, money(t, c[0], c[1]))
		}
		out = append(out, dt)
	}
	return out
}

func TestTotalsByDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		vals []val
		want []total
	}{
		{name: "no valuations, no totals", want: []total{}},
		{
			name: "a valuation that lacks an input counts nowhere",
			vals: []val{
				{isinIE, "ETF1", "2026-03-02", "1", "USD", "4.00"},
				{isinIE, "ETF2", "2026-03-02", "missing", "USD", "missing"},
			},
			want: []total{{"2026-03-02", "4.00", [][2]string{{"1", "USD"}}}},
		},
		{
			name: "one date, two currencies, sums exactly",
			vals: []val{
				{isinIE, "ETF1", "2026-03-02", "100.10", "USD", "400.40"},
				{isinIE, "ETF2", "2026-03-02", "0.20", "USD", "0.80"},
				{isinPL, "EDO0334", "2026-03-02", "1050.00", "PLN", "1050.00"},
			},
			want: []total{{"2026-03-02", "1451.20", [][2]string{{"1050.00", "PLN"}, {"100.30", "USD"}}}},
		},
		{
			name: "dates come out in order",
			vals: []val{
				{isinIE, "ETF1", "2026-03-03", "1", "EUR", "4.30"},
				{isinIE, "ETF1", "2026-03-01", "1", "EUR", "4.25"},
			},
			want: []total{
				{"2026-03-01", "4.25", [][2]string{{"1", "EUR"}}},
				{"2026-03-03", "4.30", [][2]string{{"1", "EUR"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := views.TotalsByDate(vals(t, tt.vals), pln(t))
			if err != nil {
				t.Fatalf("TotalsByDate() error = %v", err)
			}
			if diff := cmp.Diff(totals(t, tt.want), got); diff != "" {
				t.Errorf("TotalsByDate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTotalsByDateOtherCurrency(t *testing.T) {
	t.Parallel()
	in := vals(t, []val{{isinIE, "ETF1", "2026-03-02", "1", "USD", "1"}})
	in[0].PLN = money(t, "1", "USD")
	if _, err := views.TotalsByDate(in, pln(t)); !errors.Is(err, views.ErrNotReportingCurrency) {
		t.Errorf("TotalsByDate() error = %v, want %v", err, views.ErrNotReportingCurrency)
	}
}

func TestValidateRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, from, to string
		wantErr        error
	}{
		{name: "one day", from: "2026-01-01", to: "2026-01-01"},
		{name: "the longest range", from: "2016-01-01", to: "2026-02-17"},
		{name: "ends before it starts", from: "2026-01-02", to: "2026-01-01", wantErr: views.ErrInvalidRange},
		{name: "one day too long", from: "2016-01-01", to: "2026-02-18", wantErr: views.ErrInvalidRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := views.ValidateRange(date(t, tt.from), date(t, tt.to))
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateRange(%s, %s) = %v, want %v", tt.from, tt.to, err, tt.wantErr)
			}
		})
	}
}
