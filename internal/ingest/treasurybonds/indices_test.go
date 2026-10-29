package treasurybonds_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// decimal parses a decimal or fails the test.
func decimal(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

// synthetic reference rate changes, deliberately out of order.
func changes(t *testing.T) []pl.RateChange {
	t.Helper()
	return []pl.RateChange{
		{From: mustDate(t, "2025-07-03"), Rate: decimal(t, "0.05")},
		{From: mustDate(t, "2023-10-05"), Rate: decimal(t, "0.0575")},
		{From: mustDate(t, "2025-05-08"), Rate: decimal(t, "0.0525")},
	}
}

func TestIndexWatermark(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{name: "a first sync", in: "", want: ""},
		{name: "CPI only", in: "cpi=2026-09", want: "cpi=2026-09"},
		{name: "reference rate only", in: "reference-rate=2026-11", want: "reference-rate=2026-11"},
		{name: "both", in: "cpi=2026-09;reference-rate=2026-11", want: "cpi=2026-09;reference-rate=2026-11"},
		{name: "both, the other way round", in: "reference-rate=2026-11;cpi=2026-09", want: "cpi=2026-09;reference-rate=2026-11"},
		{name: "another connector's date", in: "2026-10-28", wantErr: treasurybonds.ErrBadWatermark},
		{name: "no month", in: "cpi", wantErr: treasurybonds.ErrBadWatermark},
		{name: "month 13", in: "cpi=2026-13", wantErr: treasurybonds.ErrBadWatermark},
		{name: "unknown index", in: "wibor=2026-01", wantErr: treasurybonds.ErrBadWatermark},
		{name: "CPI twice", in: "cpi=2026-01;cpi=2026-02", wantErr: treasurybonds.ErrBadWatermark},
		{name: "an empty part", in: "cpi=2026-01;", wantErr: treasurybonds.ErrBadWatermark},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w, err := treasurybonds.ParseIndexWatermark(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if got := w.String(); err == nil && got != tt.want {
				t.Errorf("String = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultIndexStart(t *testing.T) {
	t.Parallel()
	if got := treasurybonds.DefaultIndexStart(mustDate(t, "2026-10-29")); !got.Equal(mustDate(t, "2013-01-01")) {
		t.Errorf("DefaultIndexStart = %s, want 2013-01-01", got)
	}
}

// refs is reference-rate readings for consecutive months from year-month.
func refs(t *testing.T, year int, month time.Month, values ...string) []valuation.IndexReading {
	t.Helper()
	out := make([]valuation.IndexReading, 0, len(values))
	for i, v := range values {
		m := time.Date(year, month+time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		out = append(out, valuation.IndexReading{Kind: valuation.ReferenceRate, Year: m.Year(), Month: m.Month(), Value: decimal(t, v), Source: "nbp-ref"})
	}
	return out
}

func TestReferenceReadings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		from  string
		today string
		want  []valuation.IndexReading
	}{
		// May 2025's day is 16 April, June's 19 May, July's 16 June and
		// August's 18 July; the changes of 8 May and 3 July land in June and
		// August.
		{name: "each month takes the rate in force on its day", from: "2025-05-20", today: "2025-07-19", want: refs(t, 2025, time.May, "0.0575", "0.0525", "0.0525", "0.05")},
		{name: "a month whose day is today waits", from: "2025-05-01", today: "2025-07-18", want: refs(t, 2025, time.May, "0.0575", "0.0525", "0.0525")},
		// September 2023's day is 18 August and October's 15 September,
		// before the first change of 5 October; November's is 17 October.
		{name: "months before the first change have none", from: "2023-09-01", today: "2023-10-20", want: refs(t, 2023, time.November, "0.0575")},
		{name: "no month due yet", from: "2025-08-01", today: "2025-07-18", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := treasurybonds.ReferenceReadings(changes(t), mustDate(t, tt.from), mustDate(t, tt.today))
			if err != nil {
				t.Fatalf("ReferenceReadings: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
	if _, err := treasurybonds.ReferenceReadings(changes(t), mustDate(t, "1991-01-01"), mustDate(t, "2025-07-18")); !errors.Is(err, pl.ErrOutsideCalendar) {
		t.Errorf("a month before the holiday calendar: %v, want ErrOutsideCalendar", err)
	}
}

// fakeCPI serves synthetic CPI of 3% for every month up to September 2026,
// and records the years asked for.
type fakeCPI struct {
	mu    sync.Mutex
	years []int
	err   error
}

func (f *fakeCPI) CPI(_ context.Context, year int) ([]valuation.IndexReading, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.years = append(f.years, year)
	if f.err != nil {
		return nil, f.err
	}
	months := 12
	switch {
	case year == 2026:
		months = 9
	case year > 2026:
		months = 0
	}
	out := make([]valuation.IndexReading, 0, months)
	for m := time.January; int(m) <= months; m++ {
		out = append(out, valuation.IndexReading{Kind: valuation.CPI, Year: year, Month: m, Value: kernel.NewDecimal(3, 2), Source: "gus-cpi"})
	}
	return out, nil
}

// fakeRates serves fixed changes and counts the requests.
type fakeRates struct {
	changes []pl.RateChange
	calls   int
	err     error
}

func (f *fakeRates) ReferenceRates(context.Context) ([]pl.RateChange, error) {
	f.calls++
	return f.changes, f.err
}

// span names the first and last month of the readings of one kind.
func span(readings []valuation.IndexReading, kind valuation.IndexKind) string {
	var first, last string
	n := 0
	for _, r := range readings {
		if r.Kind != kind {
			continue
		}
		m := fmt.Sprintf("%d-%02d", r.Year, int(r.Month))
		if first == "" {
			first = m
		}
		last = m
		n++
	}
	return fmt.Sprintf("%d from %s to %s", n, first, last)
}

func TestIndexSyncReadings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		start         string
		watermark     string
		today         string
		wantCPI       string
		wantRates     string
		wantWatermark string
		wantYears     []int
		wantNBPCalls  int
	}{
		{
			name: "a first sync delivers both from the start", start: "2024-11-01", today: "2026-10-29",
			wantCPI: "23 from 2024-11 to 2026-09", wantRates: "25 from 2024-11 to 2026-11",
			wantWatermark: "cpi=2026-09;reference-rate=2026-11", wantYears: []int{2024, 2025, 2026}, wantNBPCalls: 1,
		},
		{
			name: "nothing new: NBP isn't asked", start: "2024-11-01", watermark: "cpi=2026-09;reference-rate=2026-11", today: "2026-11-05",
			wantCPI: "0 from  to ", wantRates: "0 from  to ",
			wantWatermark: "cpi=2026-09;reference-rate=2026-11", wantYears: []int{2026}, wantNBPCalls: 0,
		},
		{
			name: "CPI history arrives four years a sync", start: "2013-01-01", today: "2026-10-29",
			wantCPI: "48 from 2013-01 to 2016-12", wantRates: "37 from 2023-11 to 2026-11",
			wantWatermark: "cpi=2016-12;reference-rate=2026-11", wantYears: []int{2013, 2014, 2015, 2016}, wantNBPCalls: 1,
		},
		{
			name: "a new year continues after December", start: "2013-01-01", watermark: "cpi=2025-12;reference-rate=2026-11", today: "2026-11-20",
			wantCPI: "9 from 2026-01 to 2026-09", wantRates: "1 from 2026-12 to 2026-12",
			wantWatermark: "cpi=2026-09;reference-rate=2026-12", wantYears: []int{2026}, wantNBPCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cpi, rates := &fakeCPI{}, &fakeRates{changes: changes(t)}
			s := treasurybonds.IndexSync{CPI: cpi, Rates: rates, Start: mustDate(t, tt.start)}
			got, watermark, err := s.Readings(context.Background(), tt.watermark, mustDate(t, tt.today))
			if err != nil {
				t.Fatalf("Readings: %v", err)
			}
			type result struct {
				CPI, Rates, Watermark string
				Years                 []int
				NBPCalls              int
			}
			want := result{CPI: tt.wantCPI, Rates: tt.wantRates, Watermark: tt.wantWatermark, Years: tt.wantYears, NBPCalls: tt.wantNBPCalls}
			gotResult := result{
				CPI: span(got, valuation.CPI), Rates: span(got, valuation.ReferenceRate),
				Watermark: watermark, Years: cpi.years, NBPCalls: rates.calls,
			}
			if diff := cmp.Diff(want, gotResult); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

// errDown is a synthetic outage of GUS or NBP.
var errDown = errors.New("synthetic outage")

func TestIndexSyncReadingsFails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		watermark string
		cpiErr    error
		ratesErr  error
		wantErr   error
	}{
		{name: "a watermark it never wrote", watermark: "2026-10-28", wantErr: treasurybonds.ErrBadWatermark},
		{name: "GUS fails", cpiErr: errDown, wantErr: errDown},
		{name: "NBP fails", ratesErr: errDown, wantErr: errDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := treasurybonds.IndexSync{
				CPI: &fakeCPI{err: tt.cpiErr}, Rates: &fakeRates{changes: changes(t), err: tt.ratesErr}, Start: mustDate(t, "2025-01-01"),
			}
			_, _, err := s.Readings(context.Background(), tt.watermark, mustDate(t, "2026-10-29"))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("error (-want +got):\n%s", diff)
			}
		})
	}
}
