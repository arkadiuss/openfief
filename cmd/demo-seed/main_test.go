package main

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/riverqueue/river/rivertype"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
)

func mustDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func TestParseConfig(t *testing.T) {
	t.Parallel()
	// 23:30 UTC on 28 October is already 29 October in Warsaw (UTC+1).
	now := time.Date(2026, time.October, 28, 23, 30, 0, 0, time.UTC)
	tests := []struct {
		name      string
		args      []string
		dsn       string
		wantToday string
		wantReset bool
		wantSeed  uint64
		wantErr   bool
	}{
		{name: "defaults: today in Warsaw, the default seed", dsn: "postgres://demo", wantToday: "2026-10-29", wantSeed: defaultSeed},
		{name: "flags", args: []string{"--reset-demo", "--today", "2026-10-01", "--seed", "7"}, dsn: "postgres://demo", wantToday: "2026-10-01", wantReset: true, wantSeed: 7},
		{name: "no DSN", wantErr: true},
		{name: "bad date", args: []string{"--today", "28.10.2026"}, dsn: "postgres://demo", wantErr: true},
		{name: "unknown flag", args: []string{"--force"}, dsn: "postgres://demo", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseConfig(tt.args, tt.dsn, now)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseConfig = %+v, want an error", got)
				}
				return
			}
			if err != nil || got.Today.String() != tt.wantToday || got.Reset != tt.wantReset || got.Seed != tt.wantSeed || got.DSN != tt.dsn {
				t.Errorf("parseConfig = %+v, %v; want today %s, reset %v, seed %d", got, err, tt.wantToday, tt.wantReset, tt.wantSeed)
			}
		})
	}
}

func TestCommittedRatesParse(t *testing.T) {
	t.Parallel()
	rates, err := parseRates(committedRates, mustDate(t, "2025-10-01"), mustDate(t, "2026-10-28"))
	if err != nil {
		t.Fatalf("parseRates(committed): %v", err)
	}
	seen := map[string]int{}
	for _, r := range rates {
		seen[r.Currency.String()]++
		if r.Source != "nbp-a" || r.Date.Before(mustDate(t, "2025-10-01")) {
			t.Errorf("rate %s %s from %s", r.Currency, r.Date, r.Source)
		}
	}
	if seen["USD"] < 250 || seen["EUR"] < 250 || len(seen) != 2 {
		t.Errorf("committed rates per currency = %v, want a year of USD and EUR", seen)
	}
}

func TestParseRates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		data    string
		want    []string
		wantErr bool
	}{
		{name: "comments and range", data: "# c\n\n2026-01-01,USD,4.0\n2026-01-02,USD,4.1\n2026-02-01,EUR,4.2\n", want: []string{"2026-01-02 USD 4.1"}},
		{name: "too few fields", data: "2026-01-02,USD\n", wantErr: true},
		{name: "bad currency", data: "2026-01-02,usd,4.1\n", wantErr: true},
		{name: "zero rate", data: "2026-01-02,USD,0\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rates, err := parseRates([]byte(tt.data), mustDate(t, "2026-01-02"), mustDate(t, "2026-01-31"))
			if tt.wantErr {
				if !errors.Is(err, errBadRates) {
					t.Errorf("err = %v, want errBadRates", err)
				}
				return
			}
			got := make([]string, 0, len(rates))
			for _, r := range rates {
				got = append(got, r.Date.String()+" "+r.Currency.String()+" "+r.Rate.String())
			}
			if err != nil || !cmp.Equal(tt.want, got) {
				t.Errorf("parseRates = %v, %v; want %v", got, err, tt.want)
			}
		})
	}
}

func TestOnlyDemoCurrencies(t *testing.T) {
	t.Parallel()
	rates := make([]valuation.FXRate, 0, 3)
	for _, code := range []string{"USD", "CHF", "EUR"} {
		c, err := kernel.ParseCurrency(code)
		if err != nil {
			t.Fatalf("ParseCurrency: %v", err)
		}
		rates = append(rates, valuation.FXRate{Currency: c})
	}
	got := onlyDemoCurrencies(rates)
	if len(got) != 2 || got[0].Currency.String() != "USD" || got[1].Currency.String() != "EUR" {
		t.Errorf("onlyDemoCurrencies kept %v, want USD and EUR", got)
	}
}

func TestAllValued(t *testing.T) {
	t.Parallel()
	today := mustDate(t, "2026-10-28")
	ds := &demo.Dataset{Today: today, Accounts: demo.Accounts()}
	valued := func(id ledger.AccountID) valuationapp.ValuationStatus {
		return valuationapp.ValuationStatus{Account: id, ValuedThrough: today}
	}
	all := []valuationapp.ValuationStatus{valued(demo.IBKRAccount), valued(demo.XTBAccount), valued(demo.BondsAccount)}
	pending := append([]valuationapp.ValuationStatus(nil), all...)
	pending[1].Pending = true
	behind := append([]valuationapp.ValuationStatus(nil), all...)
	behind[2].ValuedThrough = today.AddDays(-1)
	tests := []struct {
		name     string
		statuses []valuationapp.ValuationStatus
		want     bool
	}{
		{name: "every account valued through today", statuses: all, want: true},
		{name: "one pending", statuses: pending},
		{name: "one behind", statuses: behind},
		{name: "one missing", statuses: all[:2]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := allValued(tt.statuses, ds); got != tt.want {
				t.Errorf("allValued = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	t.Parallel()
	in := ledgerapp.ImportSummary{
		New: 2, NewVersions: 1, Unchanged: 3, AlreadySeeded: []ledger.AccountID{demo.IBKRAccount},
		Checks: []ledgerapp.AssertionCheck{{Status: ledgerapp.AssertionMatched}, {Status: ledgerapp.AssertionFlagged}},
	}
	got := summary(&in)
	if got.New != 2 || got.NewVersions != 1 || got.Unchanged != 3 || got.FailedAssertions != 1 || len(got.AlreadySeeded) != 1 {
		t.Errorf("summary = %+v", got)
	}
}

func TestLastError(t *testing.T) {
	t.Parallel()
	job := &rivertype.JobRow{Errors: []rivertype.AttemptError{{Error: "first"}, {Error: "second"}}}
	if got := lastError(job); got != "second" {
		t.Errorf("lastError = %q, want the latest", got)
	}
	if got := lastError(&rivertype.JobRow{}); got != "" {
		t.Errorf("lastError of a job without errors = %q", got)
	}
}
