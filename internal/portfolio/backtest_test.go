package portfolio_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// mustDate parses a YYYY-MM-DD date or fails the test.
func mustDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// closesOf reads "YYYY-MM-DD amount CUR" lines into closes.
func closesOf(t *testing.T, lines []string) []portfolio.Close {
	t.Helper()
	out := make([]portfolio.Close, 0, len(lines))
	for _, l := range lines {
		f := strings.Fields(l)
		out = append(out, portfolio.Close{Date: mustDate(t, f[0]), Value: kernel.NewMoney(mustDec(t, f[1]), mustCurrency(t, f[2]))})
	}
	return out
}

// ratesOf reads "CUR YYYY-MM-DD rate" lines into rates.
func ratesOf(t *testing.T, lines []string) []portfolio.Rate {
	t.Helper()
	out := make([]portfolio.Rate, 0, len(lines))
	for _, l := range lines {
		f := strings.Fields(l)
		out = append(out, portfolio.Rate{Currency: mustCurrency(t, f[0]), Date: mustDate(t, f[1]), Rate: mustDec(t, f[2])})
	}
	return out
}

// pointsOf reads "YYYY-MM-DD value" lines into points.
func pointsOf(t *testing.T, lines []string) []portfolio.Point {
	t.Helper()
	out := make([]portfolio.Point, 0, len(lines))
	for _, l := range lines {
		f := strings.Fields(l)
		out = append(out, portfolio.Point{Date: mustDate(t, f[0]), Value: mustDec(t, f[1])})
	}
	return out
}

// backtestCase is one backtest table row, as text.
type backtestCase struct {
	name    string
	targets []targetSpec
	closes  map[string][]string // by target name
	rates   []string
	through string
	want    []string // "date index"
	missing []portfolio.MissingData
	wantErr error
}

// input builds the case's BacktestInput, reporting in PLN.
func (c *backtestCase) input(t *testing.T) *portfolio.BacktestInput {
	t.Helper()
	in := &portfolio.BacktestInput{
		Targets: buildTargets(t, c.targets), Closes: map[string][]portfolio.Close{},
		Rates: ratesOf(t, c.rates), Reporting: mustCurrency(t, "PLN"), Through: mustDate(t, c.through),
	}
	for i := range in.Targets {
		in.Closes[in.Targets[i].Key()] = closesOf(t, c.closes[in.Targets[i].Name])
	}
	return in
}

func TestBacktest(t *testing.T) {
	t.Parallel()
	inst := func(name, weight string) targetSpec {
		return targetSpec{kind: portfolio.TargetInstrument, name: name, weight: weight}
	}
	for _, tt := range backtestCases(inst) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := portfolio.Backtest(tt.input(t))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Backtest() error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(pointsOf(t, tt.want), got.Days, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Backtest() days (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantMissing(), missingOf(&got)); diff != "" {
				t.Errorf("Backtest() missing (-want +got):\n%s", diff)
			}
		})
	}
}

// wantMissing is what the case expects missing per target: none unless
// it says.
func (c *backtestCase) wantMissing() []portfolio.MissingData {
	if c.missing != nil {
		return c.missing
	}
	return make([]portfolio.MissingData, len(c.targets))
}

// missingOf lists why each target of s has no history.
func missingOf(s *portfolio.BacktestSeries) []portfolio.MissingData {
	out := make([]portfolio.MissingData, 0, len(s.Targets))
	for _, h := range s.Targets {
		out = append(out, h.Missing)
	}
	return out
}

// backtestCases are hand-computed backtests; inst builds an instrument
// target.
func backtestCases(inst func(name, weight string) targetSpec) []backtestCase {
	usd := targetSpec{kind: portfolio.TargetInstrument, name: etfB, weight: "100"}
	return []backtestCase{
		{
			name: "one PLN target follows its close, carried over days without one", targets: []targetSpec{inst(etfA, "100")},
			closes:  map[string][]string{etfA: {"2024-01-05 40 PLN", "2024-01-02 50 PLN", "2024-01-03 55 PLN"}},
			through: "2024-01-06",
			want:    []string{"2024-01-02 100", "2024-01-03 110", "2024-01-04 110", "2024-01-05 80"},
		},
		{
			name: "two targets rebalance at the end of the year", targets: []targetSpec{inst(etfA, "50"), inst(etfB, "50")},
			closes: map[string][]string{
				etfA: {"2023-12-30 10 PLN", "2023-12-31 20 PLN", "2024-01-01 20 PLN"},
				etfB: {"2023-12-30 10 PLN", "2023-12-31 10 PLN", "2024-01-01 5 PLN"},
			},
			through: "2024-01-01",
			// 5 units each, worth 150 on Dec 31; rebalanced to 3.75 of A and
			// 7.5 of B, worth 112.5 on Jan 1 (125 without rebalancing).
			want: []string{"2023-12-30 100", "2023-12-31 150", "2024-01-01 112.5"},
		},
		{
			name: "no rebalance on a last day that is Dec 31", targets: []targetSpec{inst(etfA, "50"), inst(etfB, "50")},
			closes: map[string][]string{
				etfA: {"2023-12-30 10 PLN", "2023-12-31 20 PLN"},
				etfB: {"2023-12-30 10 PLN", "2023-12-31 10 PLN"},
			},
			through: "2023-12-31",
			want:    []string{"2023-12-30 100", "2023-12-31 150"},
		},
		{
			name: "a USD target at the latest rate", targets: []targetSpec{usd},
			closes:  map[string][]string{etfB: {"2024-01-02 10 USD", "2024-01-04 11 USD"}},
			rates:   []string{"USD 2024-01-03 4.2", "USD 2024-01-01 4.0", "EUR 2024-01-01 4.5"},
			through: "2024-01-05",
			want:    []string{"2024-01-02 100", "2024-01-03 105", "2024-01-04 115.5"},
		},
		{
			name: "a USD target starts with its first rate", targets: []targetSpec{usd},
			closes:  map[string][]string{etfB: {"2024-01-01 10 USD", "2024-01-03 12 USD"}},
			rates:   []string{"USD 2024-01-02 4"},
			through: "2024-01-03",
			want:    []string{"2024-01-02 100", "2024-01-03 120"},
		},
		{
			name: "an index in its own currency", targets: []targetSpec{{kind: portfolio.TargetIndex, name: index1, currency: "PLN", weight: "100"}},
			closes:  map[string][]string{index1: {"2024-01-01 2000 PLN", "2024-01-02 1900 PLN"}},
			through: "2024-01-02",
			want:    []string{"2024-01-01 100", "2024-01-02 95"},
		},
		{
			name: "starts when every target has data and ends with the first to stop", targets: []targetSpec{inst(etfA, "60"), inst(etfB, "40")},
			closes: map[string][]string{
				etfA: {"2024-01-01 10 PLN", "2024-01-03 10 PLN", "2024-01-04 11 PLN", "2024-01-06 12 PLN"},
				etfB: {"2024-01-03 20 PLN", "2024-01-04 18 PLN"},
			},
			through: "2024-01-08",
			// 6 units of A and 2 of B: 66 + 36 = 102 on Jan 4.
			want: []string{"2024-01-03 100", "2024-01-04 102"},
		},
		{
			name: "closes after through are left out", targets: []targetSpec{inst(etfA, "100")},
			closes:  map[string][]string{etfA: {"2024-01-01 10 PLN", "2024-01-02 11 PLN", "2024-01-03 12 PLN"}},
			through: "2024-01-02",
			want:    []string{"2024-01-01 100", "2024-01-02 110"},
		},
		{
			name: "the index keeps 12 places, half-even", targets: []targetSpec{inst(etfA, "100")},
			closes:  map[string][]string{etfA: {"2024-01-01 3 PLN", "2024-01-02 1 PLN"}},
			through: "2024-01-02",
			want:    []string{"2024-01-01 100", "2024-01-02 33.333333333333"},
		},
		{
			name: "a target with no closes stops the backtest", targets: []targetSpec{inst(etfA, "50"), inst(etfB, "50")},
			closes:  map[string][]string{etfA: {"2024-01-01 10 PLN"}, etfB: {"2024-02-01 10 PLN"}},
			through: "2024-01-31",
			missing: []portfolio.MissingData{0, portfolio.MissingCloses},
		},
		{
			name: "a foreign target with no rate stops the backtest", targets: []targetSpec{usd},
			closes:  map[string][]string{etfB: {"2024-01-01 10 USD"}},
			rates:   []string{"USD 2024-01-02 4"},
			through: "2024-01-31",
			missing: []portfolio.MissingData{portfolio.MissingRates},
		},
		{
			name: "histories that don't overlap give no days", targets: []targetSpec{inst(etfA, "50"), inst(etfB, "50")},
			closes:  map[string][]string{etfA: {"2024-01-01 10 PLN"}, etfB: {"2024-01-02 10 PLN"}},
			through: "2024-01-31",
		},
		{name: "no targets give no days", through: "2024-01-31"},
		{
			name: "two closes on one date", targets: []targetSpec{inst(etfA, "100")},
			closes:  map[string][]string{etfA: {"2024-01-01 10 PLN", "2024-01-01 11 PLN"}},
			through: "2024-01-31", wantErr: portfolio.ErrInvalidHistory,
		},
		{
			name: "closes in two currencies", targets: []targetSpec{inst(etfA, "100")},
			closes:  map[string][]string{etfA: {"2024-01-01 10 PLN", "2024-01-02 11 EUR"}},
			through: "2024-01-31", wantErr: portfolio.ErrInvalidHistory,
		},
		{
			name: "a close that is not positive", targets: []targetSpec{inst(etfA, "100")},
			closes:  map[string][]string{etfA: {"2024-01-01 0 PLN"}},
			through: "2024-01-31", wantErr: portfolio.ErrInvalidHistory,
		},
		{
			name: "an index level in another currency than the index's", targets: []targetSpec{{kind: portfolio.TargetIndex, name: index1, currency: "USD", weight: "100"}},
			closes:  map[string][]string{index1: {"2024-01-01 2000 PLN"}},
			through: "2024-01-31", wantErr: portfolio.ErrInvalidHistory,
		},
		{
			name: "two rates on one date", targets: []targetSpec{usd},
			closes:  map[string][]string{etfB: {"2024-01-01 10 USD"}},
			rates:   []string{"USD 2024-01-01 4", "USD 2024-01-01 4.1"},
			through: "2024-01-31", wantErr: portfolio.ErrInvalidHistory,
		},
		{
			name: "a rate that is not positive", targets: []targetSpec{usd},
			closes:  map[string][]string{etfB: {"2024-01-01 10 USD"}},
			rates:   []string{"USD 2024-01-01 -4"},
			through: "2024-01-31", wantErr: portfolio.ErrInvalidHistory,
		},
	}
}

func TestBacktestTargetHistories(t *testing.T) {
	t.Parallel()
	c := backtestCase{
		targets: []targetSpec{{kind: portfolio.TargetInstrument, name: etfA, weight: "50"}, {kind: portfolio.TargetInstrument, name: etfB, weight: "50"}},
		closes: map[string][]string{
			etfA: {"2024-01-01 10 PLN", "2024-01-03 12 PLN"},
			etfB: {"2024-01-02 3 USD"},
		},
		rates:   []string{"USD 2024-01-01 4"},
		through: "2024-01-03",
	}
	got, err := portfolio.Backtest(c.input(t))
	if err != nil {
		t.Fatalf("Backtest: %v", err)
	}
	want := [][]portfolio.Point{
		pointsOf(t, []string{"2024-01-01 10", "2024-01-02 10", "2024-01-03 12"}),
		pointsOf(t, []string{"2024-01-02 12"}), // ends with its last close
	}
	for i, h := range got.Targets {
		if diff := cmp.Diff(want[i], h.Days); diff != "" {
			t.Errorf("target %s days (-want +got):\n%s", h.Target.Key(), diff)
		}
	}
}

func TestMissingDataString(t *testing.T) {
	t.Parallel()
	for m, want := range map[portfolio.MissingData]string{
		portfolio.MissingCloses: "closes", portfolio.MissingRates: "rates", 0: "MissingData(0)",
	} {
		if got := m.String(); got != want {
			t.Errorf("MissingData(%d).String() = %q, want %q", int(m), got, want)
		}
	}
}
