package modelcheck_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/test/e2e/chartcheck"
	"github.com/arkadiuss/openfief/test/e2e/modelcheck"
	"github.com/arkadiuss/openfief/test/e2e/modelseries"
)

// mustDecimal parses a decimal or fails the test.
func mustDecimal(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

// mustDate parses a date or fails the test.
func mustDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// backtestOfSeries runs the core's backtest (portfolio.Backtest) over the
// two series, as the core would from the levels the fixture served it, and
// returns it with each target's history.
func backtestOfSeries(t *testing.T) portfolio.BacktestSeries {
	t.Helper()
	pln, err := kernel.ParseCurrency(modelcheck.Reporting)
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	weights := map[string]string{"MODA": modelcheck.WeightA, "MODB": modelcheck.WeightB}
	in := &portfolio.BacktestInput{Closes: map[string][]portfolio.Close{}, Reporting: pln, Through: mustDate(t, modelseries.End)}
	for _, s := range modelseries.All() {
		target := portfolio.Target{Kind: portfolio.TargetIndex, Name: s.Index, Currency: pln, Weight: mustDecimal(t, weights[s.Index])}
		in.Targets = append(in.Targets, target)
		for _, c := range s.Closes(time.Time{}, time.Time{}) {
			in.Closes[target.Key()] = append(in.Closes[target.Key()], portfolio.Close{
				Date: mustDate(t, c.Date), Value: kernel.NewMoney(mustDecimal(t, c.Level), pln),
			})
		}
	}
	out, err := portfolio.Backtest(in)
	if err != nil {
		t.Fatalf("Backtest: %v", err)
	}
	return out
}

// TestHandTableEqualsBacktest checks the hand-worked table against the
// backtest the core runs over the same series: every day of the backtest,
// rebalanced at the year end, is the table's value, and the max drawdown of
// the basket and of each series is the one written down by hand. A false
// pass would need the backtest and the table to share a mistake, which two
// independent workings of the same levels rule out.
func TestHandTableEqualsBacktest(t *testing.T) {
	t.Parallel()
	series := backtestOfSeries(t)
	want, err := modelcheck.Line(modelcheck.BacktestStart)
	if err != nil {
		t.Fatalf("Line: %v", err)
	}
	got := make([]modelcheck.Day, 0, len(series.Days))
	for _, p := range series.Days {
		got = append(got, modelcheck.Day{Date: p.Date.String(), Index: p.Value.String()})
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the backtest's days (-hand table +portfolio.Backtest):\n%s", diff)
	}
	expected, err := modelcheck.Expected(modelcheck.BacktestStart)
	if err != nil {
		t.Fatalf("Expected: %v", err)
	}
	drawdowns := []struct {
		name   string
		points []portfolio.Point
		want   modelcheck.Drawdown
	}{
		{"the basket", series.Days, expected.Drawdown},
		{"MODA", series.Targets[0].Days, expected.Histories[0].Drawdown},
		{"MODB", series.Targets[1].Days, expected.Histories[1].Drawdown},
	}
	for _, d := range drawdowns {
		dd, err := portfolio.MaxDrawdown(d.points)
		if err != nil {
			t.Fatalf("MaxDrawdown of %s: %v", d.name, err)
		}
		gotDD := modelcheck.Drawdown{
			Peak: dd.Peak.String(), Trough: dd.Trough.String(), PeakValue: dd.PeakValue.String(),
			TroughValue: dd.TroughValue.String(), Depth: dd.Depth.String(),
		}
		if diff := cmp.Diff(d.want, gotDD); diff != "" {
			t.Errorf("max drawdown of %s (-hand +portfolio.MaxDrawdown):\n%s", d.name, diff)
		}
	}
}

func TestLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, from string
		first      modelcheck.Day
		count      int
		at         map[string]string
		wantErr    error
	}{
		{
			name: "from the backtest's start, 100 on December 3 and the rebalance at the year end",
			from: "2025-12-03", first: modelcheck.Day{Date: "2025-12-03", Index: "100"}, count: 101,
			at: map[string]string{
				"2025-12-12": "120", "2025-12-30": "93.6", "2025-12-31": "102", "2026-01-01": "102",
				"2026-01-20": "122.4", "2026-02-16": "79.56", "2026-03-13": "108.12",
			},
		},
		{
			name: "from the go-live day, rebased on it: 108.12 over 91.8",
			from: "2026-03-06", first: modelcheck.Day{Date: "2026-03-06", Index: "100"}, count: 8,
			at: map[string]string{"2026-03-08": "100", "2026-03-09": "117.777777777778", "2026-03-13": "117.777777777778"},
		},
		{
			name: "before the backtest, from its start",
			from: "2025-10-06", first: modelcheck.Day{Date: "2025-12-03", Index: "100"}, count: 101,
			at: map[string]string{"2025-12-10": "112"},
		},
		{name: "after its end", from: "2026-03-14", wantErr: modelcheck.ErrBadDate},
		{name: "no date", from: "March", wantErr: modelcheck.ErrBadDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := modelcheck.Line(tt.from)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Line(%q) error (-want +got):\n%s", tt.from, diff)
			}
			if tt.wantErr != nil {
				return
			}
			want := lineSummary{First: tt.first, Count: tt.count, At: tt.at}
			if diff := cmp.Diff(want, lineSummary{First: got[0], Count: len(got), At: pick(got, tt.at)}); diff != "" {
				t.Errorf("Line(%q) (-want +got):\n%s", tt.from, diff)
			}
		})
	}
}

// lineSummary is what TestLine compares: a line's first day, how many
// days it has and its index on some days.
type lineSummary struct {
	First modelcheck.Day
	Count int
	At    map[string]string
}

// pick returns the index of days on each date of dates.
func pick(days []modelcheck.Day, dates map[string]string) map[string]string {
	byDate := make(map[string]string, len(days))
	for _, d := range days {
		byDate[d.Date] = d.Index
	}
	out := make(map[string]string, len(dates))
	for date := range dates {
		out[date] = byDate[date]
	}
	return out
}

func TestAligned(t *testing.T) {
	t.Parallel()
	labels := []string{"2025-12-01", "2025-12-03", "2025-12-10", "2026-03-13", "2026-03-16"}
	got, err := modelcheck.Aligned(labels, "2025-10-06")
	if err != nil {
		t.Fatalf("Aligned: %v", err)
	}
	want := []string{"", "100", "112", "108.12", ""}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Aligned (-want +got):\n%s", diff)
	}
	if _, err := modelcheck.Aligned(labels, "2026-04-01"); err == nil {
		t.Error("Aligned from after the backtest's end succeeded, want an error")
	}
}

func TestPerformanceChart(t *testing.T) {
	t.Parallel()
	perf := &viewsv1.GetPerformanceResponse{From: "2026-03-06", Days: []*viewsv1.PerformanceDay{
		{Date: "2026-03-06", Index: "100"}, {Date: "2026-03-13", Index: "101.5"},
	}}
	got, err := modelcheck.PerformanceChart(perf)
	if err != nil {
		t.Fatalf("PerformanceChart: %v", err)
	}
	want := chartcheck.Chart{
		Labels: []string{"2026-03-06", "2026-03-13"},
		Series: [][]string{{"100", "101.5"}, {"100", "117.777777777778"}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PerformanceChart (-want +got):\n%s", diff)
	}
}

func TestExpectedAndSummarize(t *testing.T) {
	t.Parallel()
	want, err := modelcheck.Expected("2026-03-06")
	if err != nil {
		t.Fatalf("Expected: %v", err)
	}
	index := viewsv1.ModelTargetKind_MODEL_TARGET_KIND_INDEX
	resp := &viewsv1.GetModelPortfolioResponse{
		From: "2026-03-06", To: "2026-03-13", ReportingCurrency: "PLN",
		BacktestStart: "2025-12-03", BacktestEnd: "2026-03-13",
		Targets: []*viewsv1.ModelTarget{
			{Kind: index, Index: "MODA", Currency: "PLN", WeightPercent: "60"},
			{Kind: index, Index: "MODB", Currency: "PLN", WeightPercent: "40.00"},
		},
		Drawdown: &viewsv1.Drawdown{
			PeakDate: "2026-01-20", TroughDate: "2026-02-16", PeakValue: "122.400", TroughValue: "79.56", Depth: "0.350000000000",
		},
		TargetHistories: []*viewsv1.ModelTargetHistory{
			{
				Target:    &viewsv1.ModelTarget{Kind: index, Index: "MODA", Currency: "PLN", WeightPercent: "60"},
				FirstDate: "2025-12-01", LastDate: "2026-03-13",
				Drawdown: &viewsv1.Drawdown{PeakDate: "2026-01-15", TroughDate: "2026-02-02", PeakValue: "132", TroughValue: "99", Depth: "0.25"},
			},
			{
				Target:    &viewsv1.ModelTarget{Kind: index, Index: "MODB", Currency: "PLN", WeightPercent: "40"},
				FirstDate: "2025-12-03", LastDate: "2026-03-13",
				Drawdown: &viewsv1.Drawdown{PeakDate: "2025-12-12", TroughDate: "2026-02-16", PeakValue: "60", TroughValue: "27", Depth: "0.55"},
			},
		},
	}
	for _, d := range want.Days {
		padded := d.Index + "000"
		if !strings.Contains(d.Index, ".") {
			padded = d.Index + ".000"
		}
		resp.Days = append(resp.Days, &viewsv1.BacktestDay{Date: d.Date, Index: padded})
	}
	if diff := cmp.Diff(want, modelcheck.Summarize(resp)); diff != "" {
		t.Errorf("Summarize of an answer shaped as expected (-want +got):\n%s", diff)
	}
	resp.GetDays()[len(resp.GetDays())-1].Index = "108.12"
	if cmp.Equal(want, modelcheck.Summarize(resp)) {
		t.Error("Summarize of an answer with a wrong last day equals the expected one")
	}
}

func TestDrawdownRows(t *testing.T) {
	t.Parallel()
	want := [][]string{
		{"Model portfolio", "35%", "2026-01-20", "2026-02-16", "2025-12-03 to 2026-03-13"},
		{"MODA (index, PLN)", "25%", "2026-01-15", "2026-02-02", "2025-12-01 to 2026-03-13"},
		{"MODB (index, PLN)", "55%", "2025-12-12", "2026-02-16", "2025-12-03 to 2026-03-13"},
	}
	if diff := cmp.Diff(want, modelcheck.DrawdownRows()); diff != "" {
		t.Errorf("DrawdownRows (-want +got):\n%s", diff)
	}
}
