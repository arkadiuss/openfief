package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// newPerformanceFixture is a synthetic portfolio that goes live on
// 2026-03-01: acc-1 deposits 1000 PLN and buys 10 ETF1 with it, valued 1000,
// 1100 and 1050 PLN on 03-01 to 03-03; acc-2 deposits 100 USD on 03-02, at
// 4.00 PLN. Both are valued through 03-03, acc-2 with a revaluation pending.
// acc-3 has no transactions and was never valued.
func newPerformanceFixture(t *testing.T) *fixture {
	t.Helper()
	f := newFixture(t)
	etf := instrument(t, isinIE, "ETF1")
	header := func(acc ledger.AccountID, id, day string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: acc, Date: date(t, day)}
	}
	deposit, err := ledger.NewDeposit(header("acc-1", "dep-1", "2026-03-01"), money(t, "1000", "PLN"))
	if err != nil {
		t.Fatalf("NewDeposit: %v", err)
	}
	buy, err := ledger.NewBuy(header("acc-1", "buy-1", "2026-03-01"), etf, kernel.NewQuantity(dec(t, "10")), money(t, "1000", "PLN"), kernel.Money{})
	if err != nil {
		t.Fatalf("NewBuy: %v", err)
	}
	usd, err := ledger.NewDeposit(header("acc-2", "dep-2", "2026-03-02"), money(t, "100", "USD"))
	if err != nil {
		t.Fatalf("NewDeposit: %v", err)
	}
	f.ledger.accounts = append(f.ledger.accounts, ledger.Account{ID: "acc-3", Institution: "broker-c", Name: "Synthetic C", Sleeve: "core"})
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-1": {deposit, buy}, "acc-2": {usd}}
	f.valuations.vals = nil
	for day, pln := range map[string]string{"2026-03-01": "1000", "2026-03-02": "1100", "2026-03-03": "1050"} {
		f.valuations.vals = append(f.valuations.vals, (&valRow{"acc-1", etf, day, "10", pln}).build(t))
	}
	through := date(t, "2026-03-03")
	f.statuses = fakeStatuses{"acc-1": {ValuedThrough: through}, "acc-2": {ValuedThrough: through, Pending: true}}
	return f
}

// performanceDay is what the tests check of a day: value, flow and index.
type performanceDay struct {
	Date               kernel.Date
	Value, Flow, Index kernel.Decimal
}

func performanceDays(days []views.PerformanceDay) []performanceDay {
	out := make([]performanceDay, 0, len(days))
	for i := range days {
		out = append(out, performanceDay{Date: days[i].Date, Value: days[i].Value.Amount(), Flow: days[i].Flow.Amount(), Index: days[i].Index})
	}
	return out
}

func TestPerformance(t *testing.T) {
	t.Parallel()
	threeDays := func(t *testing.T) []performanceDay {
		t.Helper()
		// 03-02: (1100 + 400 - 400) / 1000 = 1.1; 03-03: 1450 / 1500
		// rounds to 0.966666666667, and 1.1 times it to 1.063333333334.
		return []performanceDay{
			{date(t, "2026-03-01"), dec(t, "1000"), dec(t, "1000"), dec(t, "100")},
			{date(t, "2026-03-02"), dec(t, "1500"), dec(t, "400"), dec(t, "110")},
			{date(t, "2026-03-03"), dec(t, "1450"), dec(t, "0"), dec(t, "106.3333333334")},
		}
	}
	tests := []struct {
		name           string
		from, to       string
		change         func(t *testing.T, f *fixture)
		wantFrom, want string
		days           func(t *testing.T) []performanceDay
		gap            views.PerformanceGap
	}{
		{
			name: "from go-live to the valued-through date by default", wantFrom: "2026-03-01", want: "2026-03-03", days: threeDays,
		},
		{
			name: "a from before go-live starts at go-live, a to after the valued-through date stops there",
			from: "2026-01-01", to: "2026-03-05", wantFrom: "2026-03-01", want: "2026-03-05", days: threeDays, gap: views.GapUnvalued,
		},
		{
			name: "a later from starts the index there", from: "2026-03-02", wantFrom: "2026-03-02", want: "2026-03-03",
			days: func(t *testing.T) []performanceDay {
				t.Helper()
				return []performanceDay{
					{date(t, "2026-03-02"), dec(t, "1500"), dec(t, "400"), dec(t, "100")},
					{date(t, "2026-03-03"), dec(t, "1450"), dec(t, "0"), dec(t, "96.6666666667")},
				}
			},
		},
		{
			name: "an account with transactions that was never valued stops the series before it starts",
			change: func(_ *testing.T, f *fixture) {
				delete(f.statuses, "acc-2")
			},
			wantFrom: "2026-03-01", want: "2026-03-03", gap: views.GapUnvalued,
			days: func(t *testing.T) []performanceDay {
				t.Helper()
				return []performanceDay{{date(t, "2026-03-01"), dec(t, "1000"), dec(t, "1000"), dec(t, "100")}}
			},
		},
		{
			name:     "nothing valued after from",
			from:     "2026-03-04",
			wantFrom: "2026-03-04", want: "2026-03-04", gap: views.GapUnvalued,
			days: func(*testing.T) []performanceDay { return []performanceDay{} },
		},
		{
			name: "no transactions, no days",
			change: func(_ *testing.T, f *fixture) {
				f.ledger.txs = nil
			},
			days: func(*testing.T) []performanceDay { return []performanceDay{} },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newPerformanceFixture(t)
			if tt.change != nil {
				tt.change(t, f)
			}
			got, err := f.views(t).Performance(context.Background(), optionalDate(t, tt.from), optionalDate(t, tt.to))
			if err != nil {
				t.Fatalf("Performance() error = %v", err)
			}
			want := app.Performance{
				From: optionalDate(t, tt.wantFrom), To: optionalDate(t, tt.want),
				Series:  views.PerformanceSeries{Gap: tt.gap},
				Pending: len(f.ledger.txs) > 0 && f.statuses["acc-2"].Pending,
			}
			got.Series.Transfers = nil
			if diff := cmp.Diff(tt.days(t), performanceDays(got.Series.Days)); diff != "" {
				t.Errorf("Performance() days mismatch (-want +got):\n%s", diff)
			}
			got.Series.Days = nil
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Performance() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// optionalDate parses s, or returns the zero date for "".
func optionalDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	if s == "" {
		return kernel.Date{}
	}
	return date(t, s)
}

func TestPerformanceFailures(t *testing.T) {
	t.Parallel()
	t.Run("a reversed range", func(t *testing.T) {
		t.Parallel()
		f := newPerformanceFixture(t)
		if _, err := f.views(t).Performance(context.Background(), date(t, "2026-03-03"), date(t, "2026-03-02")); !errors.Is(err, views.ErrInvalidRange) {
			t.Errorf("Performance(reversed) error = %v, want %v", err, views.ErrInvalidRange)
		}
	})
	t.Run("a failing valuation store", func(t *testing.T) {
		t.Parallel()
		f := newPerformanceFixture(t)
		f.valuations.fail = errStore
		if _, err := f.views(t).Performance(context.Background(), kernel.Date{}, kernel.Date{}); !errors.Is(err, errStore) {
			t.Errorf("Performance() error = %v, want %v", err, errStore)
		}
	})
	t.Run("a failing ledger", func(t *testing.T) {
		t.Parallel()
		f := newPerformanceFixture(t)
		f.ledger.fail = errStore
		if _, err := f.views(t).Performance(context.Background(), kernel.Date{}, kernel.Date{}); !errors.Is(err, errStore) {
			t.Errorf("Performance() error = %v, want %v", err, errStore)
		}
	})
	t.Run("a failing status store", func(t *testing.T) {
		t.Parallel()
		f := newPerformanceFixture(t)
		v, err := app.New(f.ledger, f.valuations, failingStatuses{}, f.rates, cur(t, "PLN"))
		if err != nil {
			t.Fatalf("app.New: %v", err)
		}
		if _, err := v.Performance(context.Background(), kernel.Date{}, kernel.Date{}); !errors.Is(err, errStore) {
			t.Errorf("Performance() error = %v, want %v", err, errStore)
		}
	})
}

// failingStatuses fails every status query.
type failingStatuses struct{}

func (failingStatuses) Status(context.Context, ledger.AccountID) (valuationapp.ValuationStatus, error) {
	return valuationapp.ValuationStatus{}, errStore
}

// valRow is a valued holding valuation of qty units of in, in acc on day,
// worth pln PLN.
type valRow struct {
	acc           ledger.AccountID
	in            ledger.Instrument
	day, qty, pln string
}

func (r *valRow) build(t *testing.T) valuation.HoldingValuation {
	t.Helper()
	return valuation.HoldingValuation{
		Account: r.acc, Instrument: r.in, Date: date(t, r.day), Basis: valuation.Market, Source: "test",
		Quantity: kernel.NewQuantity(dec(t, r.qty)), Value: money(t, r.pln, "PLN"), PLN: money(t, r.pln, "PLN"),
	}
}
