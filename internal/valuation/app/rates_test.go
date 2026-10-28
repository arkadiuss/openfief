package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

var errProviderDown = errors.New("synthetic provider failure")

func TestSplitRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		from    string
		to      string
		maxDays int
		want    [][2]string
		wantErr error
	}{
		{name: "one day", from: "2025-01-02", to: "2025-01-02", maxDays: 93, want: [][2]string{{"2025-01-02", "2025-01-02"}}},
		{name: "exactly the limit", from: "2025-01-01", to: "2025-04-03", maxDays: 93, want: [][2]string{{"2025-01-01", "2025-04-03"}}},
		{name: "one day over the limit", from: "2025-01-01", to: "2025-04-04", maxDays: 93, want: [][2]string{
			{"2025-01-01", "2025-04-03"}, {"2025-04-04", "2025-04-04"},
		}},
		{name: "a leap year in three chunks", from: "2024-01-01", to: "2024-12-31", maxDays: 150, want: [][2]string{
			{"2024-01-01", "2024-05-29"}, {"2024-05-30", "2024-10-26"}, {"2024-10-27", "2024-12-31"},
		}},
		{name: "across a year end", from: "2025-12-30", to: "2026-01-02", maxDays: 2, want: [][2]string{
			{"2025-12-30", "2025-12-31"}, {"2026-01-01", "2026-01-02"},
		}},
		{name: "end before start", from: "2025-01-03", to: "2025-01-02", maxDays: 93, wantErr: app.ErrInvalidRange},
		{name: "zero span limit", from: "2025-01-02", to: "2025-01-03", maxDays: 0, wantErr: app.ErrInvalidRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := app.SplitRange(date(t, tt.from), date(t, tt.to), tt.maxDays)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(ranges(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestFetchRates(t *testing.T) {
	t.Parallel()
	// Synthetic provider: returns its rates in reverse order to show sorting.
	p := &fakeProvider{rates: []valuation.FXRate{
		rate(t, "USD", "2025-04-04", "3.9001"),
		rate(t, "EUR", "2025-04-03", "4.2002"),
		rate(t, "USD", "2025-04-03", "3.9003"),
		rate(t, "USD", "2025-01-02", "4.1004"),
	}}
	got, err := app.FetchRates(context.Background(), p, date(t, "2025-01-01"), date(t, "2025-04-04"), 93)
	if err != nil {
		t.Fatalf("FetchRates: %v", err)
	}
	want := []valuation.FXRate{
		rate(t, "USD", "2025-01-02", "4.1004"),
		rate(t, "EUR", "2025-04-03", "4.2002"),
		rate(t, "USD", "2025-04-03", "3.9003"),
		rate(t, "USD", "2025-04-04", "3.9001"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rates (-want +got):\n%s", diff)
	}
	wantCalls := ranges(t, [][2]string{{"2025-01-01", "2025-04-03"}, {"2025-04-04", "2025-04-04"}})
	if diff := cmp.Diff(wantCalls, p.calls); diff != "" {
		t.Errorf("calls (-want +got):\n%s", diff)
	}
}

func TestFetchRatesWithNoTables(t *testing.T) {
	t.Parallel()
	p := &fakeProvider{}
	got, err := app.FetchRates(context.Background(), p, date(t, "2025-01-04"), date(t, "2025-01-06"), 93)
	if err != nil {
		t.Fatalf("FetchRates: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d rates, want none", len(got))
	}
}

func TestFetchRatesStopsAtTheFirstFailure(t *testing.T) {
	t.Parallel()
	p := &fakeProvider{err: errProviderDown}
	_, err := app.FetchRates(context.Background(), p, date(t, "2025-01-01"), date(t, "2025-12-31"), 93)
	if !errors.Is(err, errProviderDown) {
		t.Fatalf("got error %v, want %v", err, errProviderDown)
	}
	if len(p.calls) != 1 {
		t.Errorf("got %d calls, want 1", len(p.calls))
	}
	_, err = app.FetchRates(context.Background(), p, date(t, "2025-01-02"), date(t, "2025-01-01"), 93)
	if !errors.Is(err, app.ErrInvalidRange) {
		t.Errorf("got error %v, want %v", err, app.ErrInvalidRange)
	}
}

// fakeProvider records each call and returns its rates dated in the range.
type fakeProvider struct {
	rates []valuation.FXRate
	err   error
	calls []app.DateRange
}

func (f *fakeProvider) Rates(_ context.Context, from, to kernel.Date) ([]valuation.FXRate, error) {
	f.calls = append(f.calls, app.DateRange{From: from, To: to})
	if f.err != nil {
		return nil, f.err
	}
	var out []valuation.FXRate
	for _, r := range slices.Backward(f.rates) {
		if !r.Date.Before(from) && !r.Date.After(to) {
			out = append(out, r)
		}
	}
	return out, nil
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func ranges(t *testing.T, pairs [][2]string) []app.DateRange {
	t.Helper()
	out := make([]app.DateRange, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, app.DateRange{From: date(t, p[0]), To: date(t, p[1])})
	}
	return out
}

func rate(t *testing.T, code, on, value string) valuation.FXRate {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	d, err := kernel.ParseDecimal(value)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", value, err)
	}
	return valuation.FXRate{Currency: c, Date: date(t, on), Rate: d, Source: "nbp-a"}
}
