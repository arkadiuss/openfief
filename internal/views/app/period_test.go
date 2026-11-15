package app_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// periodFigures is what the tests check of a period's change, as text: the
// opening day, the gross start, end and change with its percent, the return,
// the deposits and withdrawals, and the liquidation value's start, end and
// percent.
type periodFigures struct {
	Opening, Start, End, Change, Percent, Return, Deposits, Withdrawals string
	LiquidationStart, LiquidationEnd, LiquidationPercent                string
}

// figuresOf writes a period's change as text.
func figuresOf(c *app.PeriodChange) periodFigures {
	pct := ""
	if c.Gross.HasPercent {
		pct = c.Gross.Percent.String()
	}
	return periodFigures{
		c.Opening.String(), c.Gross.Start.Amount().String(), c.Gross.End.Amount().String(), c.Gross.Change.Amount().String(), pct,
		c.Return.String(), c.Deposits.Amount().String(), c.Withdrawals.Amount().String(),
		c.Liquidation.Start.Amount().String(), c.Liquidation.End.Amount().String(), c.Liquidation.Percent.String(),
	}
}

// newPeriodFixture is the performance fixture with each holding's net value
// its gross and its liquidation value 10 PLN less, as a sale's commission
// would leave it.
func newPeriodFixture(t *testing.T) *fixture {
	t.Helper()
	f := newPerformanceFixture(t)
	for i := range f.valuations.vals {
		v := &f.valuations.vals[i]
		v.Net = v.Reporting
		var err error
		if v.Liquidation, err = v.Reporting.Sub(money(t, "10", "PLN")); err != nil {
			t.Fatalf("Sub: %v", err)
		}
	}
	return f
}

// TestPeriodChange runs the performance fixture's three days (acc-1: 1000
// PLN in at 03-01, valued 1000, 1100 and 1050; acc-2: 100 USD in at 03-02,
// worth 400 PLN) through GetPeriodChange's rules: the start is the end of
// the day before from, and the period never reaches before the record.
func TestPeriodChange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		from, to string
		want     periodFigures
	}{
		{
			name: "the start is the end of the day before from, and the day's deposit counts", from: "2026-03-02", to: "2026-03-03",
			want: periodFigures{"2026-03-01", "1000", "1450", "450", "45", "6.33", "400", "0", "990", "1440", "45.45"},
		},
		{
			name: "from the first day, the opening day is the first itself and its own deposit is in the start", from: "2026-03-01", to: "2026-03-03",
			want: periodFigures{"2026-03-01", "1000", "1450", "450", "45", "6.33", "400", "0", "990", "1440", "45.45"},
		},
		{
			name: "a from before the record starts on the first day", from: "2026-01-01", to: "2026-03-03",
			want: periodFigures{"2026-03-01", "1000", "1450", "450", "45", "6.33", "400", "0", "990", "1440", "45.45"},
		},
		{
			name: "one day falling", from: "2026-03-03", to: "2026-03-03",
			want: periodFigures{"2026-03-02", "1500", "1450", "-50", "-3.33", "-3.33", "0", "0", "1490", "1440", "-3.36"},
		},
		{
			name: "the first day alone", from: "2026-03-01", to: "2026-03-01",
			want: periodFigures{"2026-03-01", "1000", "1000", "0", "0", "0", "0", "0", "990", "990", "0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newPeriodFixture(t).views(t).PeriodChange(reader(t), date(t, tt.from), date(t, tt.to))
			if err != nil {
				t.Fatalf("PeriodChange() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, figuresOf(&got)); diff != "" {
				t.Errorf("PeriodChange() mismatch (-want +got):\n%s", diff)
			}
			if !got.Net.Known || !got.Liquidation.Known || !got.Pending {
				t.Errorf("PeriodChange() net known %v, liquidation known %v, pending %v, want all: no account lacks a value and acc-2's revaluation is pending",
					got.Net.Known, got.Liquidation.Known, got.Pending)
			}
			if got.From != date(t, tt.from) || got.To != date(t, tt.to) {
				t.Errorf("PeriodChange() period = %s to %s, want %s to %s as asked", got.From, got.To, tt.from, tt.to)
			}
		})
	}
}

// TestPeriodChangeRefusals checks what the series cannot reach is refused:
// a period ending after the valued-through date, entirely after or before
// the record, a reversed one, one over 3700 days, nothing recorded and an
// account never valued.
func TestPeriodChangeRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		from, to string
		change   func(f *fixture)
		wantErr  error
	}{
		{name: "ending after the valued-through date", from: "2026-03-02", to: "2026-03-04", wantErr: views.ErrPeriodUnvalued},
		{name: "entirely after it", from: "2026-04-01", to: "2026-04-30", wantErr: views.ErrPeriodUnvalued},
		{name: "entirely before the record", from: "2026-01-01", to: "2026-01-31", wantErr: views.ErrPeriodUnvalued},
		{name: "an end before the start", from: "2026-03-03", to: "2026-03-02", wantErr: views.ErrInvalidRange},
		{name: "a range too long", from: "2010-01-01", to: "2026-03-03", wantErr: views.ErrInvalidRange},
		{
			name: "nothing recorded", from: "2026-03-01", to: "2026-03-03", wantErr: views.ErrPeriodUnvalued,
			change: func(f *fixture) { f.ledger.txs = nil },
		},
		{
			name: "an account never valued", from: "2026-03-02", to: "2026-03-03", wantErr: views.ErrPeriodUnvalued,
			change: func(f *fixture) { delete(f.statuses, "acc-2") },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newPeriodFixture(t)
			if tt.change != nil {
				tt.change(f)
			}
			if _, err := f.views(t).PeriodChange(reader(t), date(t, tt.from), date(t, tt.to)); !errors.Is(err, tt.wantErr) {
				t.Errorf("PeriodChange() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
