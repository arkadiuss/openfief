package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

var errReport = errors.New("report failed")

// fakeModels reports a backtest of 100, 105, 110, 99, 121 from 2026-02-27
// to 2026-03-03 while on, records the date it was asked through, and
// fails with err.
type fakeModels struct {
	on      bool
	err     error
	through kernel.Date
	days    []portfolio.Point
}

func (f *fakeModels) Report(_ context.Context, through kernel.Date) (portfolioapp.Report, bool, error) {
	f.through = through
	if f.err != nil || !f.on {
		return portfolioapp.Report{}, false, f.err
	}
	var days []portfolio.Point
	for _, p := range f.days {
		if !p.Date.After(through) {
			days = append(days, p)
		}
	}
	return portfolioapp.Report{Backtest: portfolio.BacktestSeries{Days: days}}, true, nil
}

// newFakeModels is fakeModels with its five days.
func newFakeModels(t *testing.T, on bool) *fakeModels {
	t.Helper()
	f := &fakeModels{on: on}
	for i, v := range []string{"100", "105", "110", "99", "121"} {
		f.days = append(f.days, portfolio.Point{Date: date(t, "2026-02-27").AddDays(i), Value: dec(t, v)})
	}
	return f
}

// modelViews returns f's views serving models.
func modelViews(t *testing.T, f *fixture, models app.ModelReports) *app.Views {
	t.Helper()
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), app.WithModel(models))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestModelPortfolio(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, from, to string
		recorded       bool
		wantThrough    string
		want           []string // "date index"
	}{
		{
			name: "from go-live to the latest valued-through date", recorded: true, wantThrough: "2026-03-03",
			// 110, 99, 121 rebased on 110.
			want: []string{"2026-03-01 100", "2026-03-02 90", "2026-03-03 110"},
		},
		{
			name: "a later from restarts at 100", recorded: true, from: "2026-03-02", wantThrough: "2026-03-03",
			want: []string{"2026-03-02 100", "2026-03-03 122.222222222222"},
		},
		{
			name: "an earlier to", recorded: true, to: "2026-03-02", wantThrough: "2026-03-02",
			want: []string{"2026-03-01 100", "2026-03-02 90"},
		},
		{
			name: "nothing recorded: all stored history", wantThrough: "9999-12-31",
			want: []string{"2026-02-27 100", "2026-02-28 105", "2026-03-01 110", "2026-03-02 99", "2026-03-03 121"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := map[bool]func(*testing.T) *fixture{false: newFixture, true: newPerformanceFixture}[tt.recorded](t)
			models := newFakeModels(t, true)
			got, err := modelViews(t, f, models).ModelPortfolio(t.Context(), optionalDate(t, tt.from), optionalDate(t, tt.to))
			if err != nil {
				t.Fatalf("ModelPortfolio: %v", err)
			}
			want := points(t, tt.want)
			if diff := cmp.Diff(want, got.Line); diff != "" {
				t.Errorf("line (-want +got):\n%s", diff)
			}
			if !models.through.Equal(date(t, tt.wantThrough)) {
				t.Errorf("reported through %s, want %s", models.through, tt.wantThrough)
			}
			if !got.From.Equal(want[0].Date) || !got.To.Equal(want[len(want)-1].Date) {
				t.Errorf("range %s to %s", got.From, got.To)
			}
		})
	}
}

// points reads "YYYY-MM-DD value" lines.
func points(t *testing.T, lines []string) []portfolio.Point {
	t.Helper()
	out := make([]portfolio.Point, 0, len(lines))
	for _, w := range lines {
		out = append(out, portfolio.Point{Date: date(t, w[:10]), Value: dec(t, w[11:])})
	}
	return out
}

func TestModelPortfolioRefusals(t *testing.T) {
	t.Parallel()
	f := newPerformanceFixture(t)
	tests := []struct {
		name     string
		views    *app.Views
		from, to string
		want     error
	}{
		{name: "off", views: modelViews(t, f, newFakeModels(t, false)), want: app.ErrModelOff},
		{name: "no model served", views: f.views(t), want: app.ErrModelOff},
		{name: "the report fails", views: modelViews(t, f, &fakeModels{err: errReport}), want: errReport},
		{name: "to before from", views: modelViews(t, f, newFakeModels(t, true)), from: "2026-03-03", to: "2026-03-02", want: views.ErrInvalidRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.views.ModelPortfolio(t.Context(), optionalDate(t, tt.from), optionalDate(t, tt.to))
			if !cmp.Equal(tt.want, err, cmpopts.EquateErrors()) {
				t.Errorf("ModelPortfolio = %v, want %v", err, tt.want)
			}
		})
	}
}
