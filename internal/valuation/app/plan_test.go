package app_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// twoAccounts is acc-1 holding the ETF from 2026-03-02 and a bond from
// 2026-03-03, acc-2 holding another instrument from 2026-03-05, and acc-3
// with an empty log.
func twoAccounts(t *testing.T) *fakeLedger {
	t.Helper()
	return &fakeLedger{logs: map[ledger.AccountID][]ledger.VersionedTransaction{
		acc1: {
			buy(t, acc1, "b2", "2026-03-03", instrument(t, bondISIN, "EDO0336"), "10", "1000"),
			buy(t, acc1, "b1", "2026-03-02", instrument(t, etfISIN, "SYNTH"), "1", "100"),
		},
		acc2:    {buy(t, acc2, "b3", "2026-03-05", instrument(t, otherISIN, "OTHER"), "1", "10")},
		"acc-3": nil,
	}}
}

func revals(t *testing.T, pairs ...string) []valuation.Revaluation {
	t.Helper()
	var out []valuation.Revaluation
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, valuation.Revaluation{Account: ledger.AccountID(pairs[i]), From: date(t, pairs[i+1])})
	}
	return out
}

func TestPlanner(t *testing.T) {
	t.Parallel()
	p := app.NewPlanner(twoAccounts(t), twoAccounts(t))
	etf := instrument(t, etfISIN, "SYNTH").String()
	bond, err := ledger.ParseISIN(bondISIN)
	if err != nil {
		t.Fatal(err)
	}
	store := newFakeStore()
	store.through[acc1] = date(t, "2026-03-09")
	store.through[acc2] = date(t, "2026-03-01")
	tests := []struct {
		name string
		plan func() ([]valuation.Revaluation, error)
		want []valuation.Revaluation
	}{
		{
			name: "a price: accounts that held the instrument, from its date",
			plan: func() ([]valuation.Revaluation, error) {
				return p.ForInstrument(t.Context(), etf, date(t, "2026-03-04"))
			},
			want: revals(t, "acc-1", "2026-03-04"),
		},
		{
			name: "bond terms: accounts that held the ISIN, from their first transaction",
			plan: func() ([]valuation.Revaluation, error) { return p.ForISIN(t.Context(), bond) },
			want: revals(t, "acc-1", "2026-03-02"),
		},
		{
			name: "an FX rate or index reading: every account with a log",
			plan: func() ([]valuation.Revaluation, error) { return p.ForAll(t.Context(), date(t, "2026-03-04")) },
			want: revals(t, "acc-1", "2026-03-04", "acc-2", "2026-03-04"),
		},
		{
			name: "the bond basis: every account from its first transaction",
			plan: func() ([]valuation.Revaluation, error) { return p.FromStart(t.Context(), "") },
			want: revals(t, "acc-1", "2026-03-02", "acc-2", "2026-03-05"),
		},
		{
			name: "a tax wrapper: that account from its first transaction",
			plan: func() ([]valuation.Revaluation, error) { return p.FromStart(t.Context(), acc2) },
			want: revals(t, "acc-2", "2026-03-05"),
		},
		{
			name: "end of day: from the valued-through date, never before the first transaction",
			plan: func() ([]valuation.Revaluation, error) { return p.EndOfDay(t.Context(), store) },
			want: revals(t, "acc-1", "2026-03-09", "acc-2", "2026-03-05"),
		},
		{
			name: "an import: the earliest affected date per account",
			plan: func() ([]valuation.Revaluation, error) {
				return app.ForImport(map[ledger.AccountID]kernel.Date{acc2: date(t, "2026-03-07"), acc1: date(t, "2026-03-03")}), nil
			},
			want: revals(t, "acc-1", "2026-03-03", "acc-2", "2026-03-07"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.plan()
			if err != nil {
				t.Fatalf("plan: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlannerErrors(t *testing.T) {
	t.Parallel()
	broken := twoAccounts(t)
	broken.err = errStoreDown
	p := app.NewPlanner(broken, broken)
	if _, err := p.ForAll(t.Context(), date(t, "2026-03-01")); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("ForAll: error = %v", err)
	}
	store := newFakeStore()
	store.err = errStoreDown
	if _, err := app.NewPlanner(twoAccounts(t), twoAccounts(t)).EndOfDay(t.Context(), store); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("EndOfDay: error = %v", err)
	}
}

func TestRecordInputs(t *testing.T) {
	t.Parallel()
	bond, err := ledger.ParseISIN(bondISIN)
	if err != nil {
		t.Fatal(err)
	}
	series, err := pl.ParseSeries("EDO0336")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		record func(r *app.InputRecorder) error
		want   recorded
	}{
		{
			name: "prices coalesce to the earliest date per account",
			record: func(r *app.InputRecorder) error {
				return r.RecordPrices(t.Context(), []valuation.Price{
					usdPrice(t, "2026-03-06", "1"), usdPrice(t, "2026-03-04", "1"),
					{Instrument: instrument(t, otherISIN, "OTHER").String(), Date: date(t, "2026-03-08"), Close: money(t, "1", "PLN")},
				})
			},
			want: recorded{what: "prices", revals: revals(t, "acc-1", "2026-03-04", "acc-2", "2026-03-08")},
		},
		{
			name: "FX rates revalue every account from the earliest",
			record: func(r *app.InputRecorder) error {
				return r.RecordFXRates(t.Context(), []valuation.FXRate{rate(t, "USD", "2026-03-05", "4"), rate(t, "EUR", "2026-03-04", "4.3")})
			},
			want: recorded{what: "fx", revals: revals(t, "acc-1", "2026-03-04", "acc-2", "2026-03-04")},
		},
		{
			name: "index readings revalue every account from the month's first day",
			record: func(r *app.InputRecorder) error {
				return r.RecordIndexReadings(t.Context(), []valuation.IndexReading{{Kind: valuation.CPI, Year: 2026, Month: time.February, Value: dec(t, "0.03")}})
			},
			want: recorded{what: "index", revals: revals(t, "acc-1", "2026-02-01", "acc-2", "2026-02-01")},
		},
		{
			name: "bond terms revalue its holders from the start",
			record: func(r *app.InputRecorder) error {
				return r.RecordBondTerms(t.Context(), bond, &pl.Terms{Series: series, EarlyRedemptionFee: money(t, "2", "PLN")})
			},
			want: recorded{what: "terms", revals: revals(t, "acc-1", "2026-03-02")},
		},
		{
			name:   "a tax wrapper revalues its account from the start",
			record: func(r *app.InputRecorder) error { return r.RecordTaxWrapper(t.Context(), acc1, pl.IKE) },
			want:   recorded{what: "wrapper", revals: revals(t, "acc-1", "2026-03-02")},
		},
		{
			name:   "end of day enqueues from the valued-through dates",
			record: func(r *app.InputRecorder) error { return r.EndOfDay(t.Context(), newFakeStore()) },
			want:   recorded{what: "enqueue", revals: revals(t, "acc-1", "2026-03-02", "acc-2", "2026-03-05")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &fakeRecorder{}
			l := twoAccounts(t)
			if err := tt.record(app.NewInputRecorder(app.NewPlanner(l, l), rec)); err != nil {
				t.Fatalf("record: %v", err)
			}
			if diff := cmp.Diff([]recorded{tt.want}, rec.calls, cmp.AllowUnexported(recorded{})); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestRecordRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	bond, err := ledger.ParseISIN(bondISIN)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		record func(r *app.InputRecorder) error
	}{
		{name: "price of no instrument", record: func(r *app.InputRecorder) error {
			return r.RecordPrices(t.Context(), []valuation.Price{{Date: date(t, "2026-03-02"), Close: money(t, "1", "USD")}})
		}},
		{name: "zero price", record: func(r *app.InputRecorder) error {
			return r.RecordPrices(t.Context(), []valuation.Price{usdPrice(t, "2026-03-02", "0")})
		}},
		{name: "a PLN rate", record: func(r *app.InputRecorder) error {
			return r.RecordFXRates(t.Context(), []valuation.FXRate{rate(t, "PLN", "2026-03-02", "1")})
		}},
		{name: "a negative rate", record: func(r *app.InputRecorder) error {
			return r.RecordFXRates(t.Context(), []valuation.FXRate{rate(t, "USD", "2026-03-02", "-1")})
		}},
		{name: "an unknown index kind", record: func(r *app.InputRecorder) error {
			return r.RecordIndexReadings(t.Context(), []valuation.IndexReading{{Year: 2026, Month: time.March}})
		}},
		{name: "a reading for month 13", record: func(r *app.InputRecorder) error {
			return r.RecordIndexReadings(t.Context(), []valuation.IndexReading{{Kind: valuation.CPI, Year: 2026, Month: 13}})
		}},
		{name: "bond terms without an ISIN", record: func(r *app.InputRecorder) error {
			return r.RecordBondTerms(t.Context(), ledger.ISIN{}, &pl.Terms{})
		}},
		{name: "bond terms without a series", record: func(r *app.InputRecorder) error {
			return r.RecordBondTerms(t.Context(), bond, &pl.Terms{})
		}},
		{name: "an unknown tax wrapper", record: func(r *app.InputRecorder) error {
			return r.RecordTaxWrapper(t.Context(), acc1, pl.TaxWrapper(9))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &fakeRecorder{}
			l := twoAccounts(t)
			err := tt.record(app.NewInputRecorder(app.NewPlanner(l, l), rec))
			if !cmp.Equal(app.ErrInvalidInput, err, cmpopts.EquateErrors()) {
				t.Errorf("error = %v, want ErrInvalidInput", err)
			}
			if len(rec.calls) != 0 {
				t.Errorf("stored %v, want nothing", rec.calls)
			}
		})
	}
}

func TestApplyBondBasis(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		configured valuation.Basis
		changed    bool
		want       []recorded
		wantErr    error
	}{
		{
			name: "unset is net if redeemed; a change revalues everything from the start", changed: true,
			want: []recorded{{what: "basis net_if_redeemed", revals: revals(t, "acc-1", "2026-03-02", "acc-2", "2026-03-05")}},
		},
		{
			name: "gross accrued, unchanged", configured: valuation.GrossAccrued,
			want: []recorded{{what: "basis gross_accrued", revals: revals(t, "acc-1", "2026-03-02", "acc-2", "2026-03-05")}},
		},
		{name: "market is not a bond basis", configured: valuation.Market, wantErr: valuation.ErrBasisNotForBonds},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &fakeRecorder{changed: tt.changed}
			l := twoAccounts(t)
			changed, err := app.NewInputRecorder(app.NewPlanner(l, l), rec).ApplyBondBasis(t.Context(), tt.configured)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if changed != tt.changed {
				t.Errorf("changed = %v, want %v", changed, tt.changed)
			}
			if diff := cmp.Diff(tt.want, rec.calls, cmp.AllowUnexported(recorded{})); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestRecordPassesStoreFailuresOn(t *testing.T) {
	t.Parallel()
	rec := &fakeRecorder{err: errStoreDown}
	l := twoAccounts(t)
	r := app.NewInputRecorder(app.NewPlanner(l, l), rec)
	if err := r.RecordFXRates(t.Context(), []valuation.FXRate{rate(t, "USD", "2026-03-02", "4")}); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("error = %v", err)
	}
}
