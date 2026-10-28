package pl_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

func TestMaturity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		series    string
		purchased string
		want      string
		wantErr   error
	}{
		{name: "EDO bought 2026-05-12 matures 10 years later", series: "EDO0536", purchased: "2026-05-12", want: "2036-05-12"},
		{name: "OTS term is 3 months", series: "OTS0825", purchased: "2025-05-05", want: "2025-08-05"},
		{name: "leap-day TOS matures on Feb 28", series: "TOS0227", purchased: "2024-02-29", want: "2027-02-28"},
		{name: "ROR from Jan 31", series: "ROR0126", purchased: "2025-01-31", want: "2026-01-31"},
		{name: "bought outside the series' sale month", series: "EDO0536", purchased: "2026-06-01", wantErr: pl.ErrSeriesMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.Maturity(series(t, tt.series), date(t, tt.purchased))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(date(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestPeriodStart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		typ       pl.BondType
		purchased string
		n         int
		want      string
	}{
		{name: "period 1 starts on the purchase date", typ: pl.EDO, purchased: "2024-03-15", n: 1, want: "2024-03-15"},
		{name: "yearly period 2", typ: pl.EDO, purchased: "2024-03-15", n: 2, want: "2025-03-15"},
		{name: "monthly period from Jan 31 ends Feb 28", typ: pl.ROR, purchased: "2025-01-31", n: 2, want: "2025-02-28"},
		{name: "monthly period 3 goes back to the 31st", typ: pl.ROR, purchased: "2025-01-31", n: 3, want: "2025-03-31"},
		{name: "leap-day purchase, next year Feb 28", typ: pl.TOS, purchased: "2024-02-29", n: 2, want: "2025-02-28"},
		{name: "leap-day purchase, 4 years later Feb 29", typ: pl.EDO, purchased: "2024-02-29", n: 5, want: "2028-02-29"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.PeriodStart(tt.typ, date(t, tt.purchased), tt.n)
			if err != nil {
				t.Fatalf("PeriodStart: %v", err)
			}
			if diff := cmp.Diff(date(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestCPIMonth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		start     string
		wantYear  int
		wantMonth time.Month
	}{
		{name: "period from March uses January CPI", start: "2025-03-15", wantYear: 2025, wantMonth: time.January},
		{name: "period from January uses November CPI of the year before", start: "2025-01-10", wantYear: 2024, wantMonth: time.November},
		{name: "period from February uses December CPI", start: "2026-02-01", wantYear: 2025, wantMonth: time.December},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			y, m := pl.CPIMonth(date(t, tt.start))
			if y != tt.wantYear || m != tt.wantMonth {
				t.Errorf("CPIMonth(%s) = %d-%s, want %d-%s", tt.start, y, m, tt.wantYear, tt.wantMonth)
			}
		})
	}
}
