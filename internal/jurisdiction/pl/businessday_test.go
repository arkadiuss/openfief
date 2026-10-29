package pl_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

func TestIsBusinessDay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		day     string
		want    bool
		wantErr error
	}{
		{name: "an ordinary Thursday", day: "2026-10-29", want: true},
		{name: "a Saturday", day: "2026-10-31", want: false},
		{name: "a Sunday", day: "2026-11-01", want: false},
		{name: "New Year", day: "2026-01-01", want: false},
		{name: "Epiphany since 2011", day: "2026-01-06", want: false},
		{name: "6 January was a working day in 2010", day: "2010-01-06", want: true},
		{name: "Easter Monday 2025", day: "2025-04-21", want: false},
		{name: "Easter Monday 2026", day: "2026-04-06", want: false},
		{name: "Easter Monday 2008, the earliest in the range", day: "2008-03-24", want: false},
		{name: "Easter Monday 2038, the latest possible", day: "2038-04-26", want: false},
		{name: "the Tuesday after Easter is a working day", day: "2025-04-22", want: true},
		{name: "Labour Day", day: "2026-05-01", want: false},
		{name: "Constitution Day", day: "2025-05-03", want: false},
		{name: "Corpus Christi 2025", day: "2025-06-19", want: false},
		{name: "Corpus Christi 2026", day: "2026-06-04", want: false},
		{name: "Assumption", day: "2025-08-15", want: false},
		{name: "Independence Day", day: "2026-11-11", want: false},
		{name: "Christmas Eve was a working day in 2024", day: "2024-12-24", want: true},
		{name: "Christmas Eve is a day off from 2025", day: "2025-12-24", want: false},
		{name: "Christmas", day: "2025-12-25", want: false},
		{name: "Boxing Day", day: "2025-12-26", want: false},
		{name: "before the calendar", day: "1990-06-01", wantErr: pl.ErrOutsideCalendar},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.IsBusinessDay(date(t, tt.day))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if got != tt.want {
				t.Errorf("IsBusinessDay(%s) = %t, want %t", tt.day, got, tt.want)
			}
		})
	}
}

func TestReferenceRateDay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		year    int
		month   time.Month
		want    string
		wantErr error
	}{
		// Counting back from 30 April, Easter Monday (21 April) is skipped.
		{name: "May 2025 skips Easter Monday", year: 2025, month: time.May, want: "2025-04-16"},
		{name: "June 2025: ten weekdays back", year: 2025, month: time.June, want: "2025-05-19"},
		// Corpus Christi, Thursday 19 June 2025, is skipped.
		{name: "July 2025 skips Corpus Christi", year: 2025, month: time.July, want: "2025-06-16"},
		// 24, 25 and 26 December are all days off from 2025.
		{name: "January 2026 skips three Christmas days", year: 2026, month: time.January, want: "2025-12-15"},
		{name: "January 2025: Christmas Eve 2024 still counts", year: 2025, month: time.January, want: "2024-12-16"},
		{name: "November 2026", year: 2026, month: time.November, want: "2026-10-19"},
		{name: "before the calendar", year: 1991, month: time.January, wantErr: pl.ErrOutsideCalendar},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pl.ReferenceRateDay(tt.year, tt.month)
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

func TestRateInForce(t *testing.T) {
	t.Parallel()
	// Synthetic changes, deliberately out of order.
	changes := []pl.RateChange{
		{From: date(t, "2025-07-03"), Rate: dec(t, "0.05")},
		{From: date(t, "2025-05-08"), Rate: dec(t, "0.0525")},
		{From: date(t, "2023-10-05"), Rate: dec(t, "0.0575")},
	}
	tests := []struct {
		name   string
		day    string
		want   string
		wantOK bool
	}{
		{name: "before the first change", day: "2023-10-04", wantOK: false},
		{name: "on the day a change takes effect", day: "2025-05-08", want: "0.0525", wantOK: true},
		{name: "the day before a change", day: "2025-07-02", want: "0.0525", wantOK: true},
		{name: "after the last change", day: "2026-10-19", want: "0.05", wantOK: true},
		{name: "between two older changes", day: "2024-06-14", want: "0.0575", wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := pl.RateInForce(changes, date(t, tt.day))
			if ok != tt.wantOK {
				t.Fatalf("RateInForce(%s) ok = %t, want %t", tt.day, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if diff := cmp.Diff(dec(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
