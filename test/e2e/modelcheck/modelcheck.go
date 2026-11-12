// Package modelcheck works out, for the end-to-end suite's check of D-200's
// criterion 19, what the read API's model portfolio and the built-in UI must
// answer for the two synthetic index series of modelseries, by hand: the
// basket's value on each stretch of days is written down as a table worked
// out from the levels (one yearly rebalance on December 31, D-53), not
// computed by the backtest the core runs, and the max drawdowns (D-54) are
// the series' known peak and trough days. From that table it builds the
// answer GetModelPortfolio must give from a first day, the model line the
// performance chart must draw, and the drawdown rows the page must show. It
// only reads messages; the suite (test/e2e, model_test.go) calls the core
// and loads the pages. It is never released.
package modelcheck

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/test/e2e/chartcheck"
	"github.com/arkadiuss/openfief/test/e2e/modelseries"
)

// The model portfolio the suite saves: two index targets in the reporting
// currency, weighted 60% and 40%, and where its backtest starts and ends.
const (
	Reporting     = "PLN"
	BacktestStart = "2025-12-03"
	BacktestEnd   = modelseries.End
	// WeightA and WeightB are MODA's and MODB's weights in percent.
	WeightA = "60"
	WeightB = "40"
)

// The days the basket and the series peak and trough on, as text.
const (
	dec12 = "2025-12-12"
	jan15 = "2026-01-15"
	jan20 = "2026-01-20"
	feb02 = "2026-02-02"
	feb16 = "2026-02-16"
)

// baseIndex is the backtest's index on its first day.
const baseIndex = 100

// indexPlaces is how many decimal places the backtest's index keeps,
// rounded half-even, as the core keeps them (portfolio.BacktestPlaces).
const indexPlaces int32 = 12

var (
	// ErrBadDate reports a date that is not YYYY-MM-DD or lies outside the
	// backtest.
	ErrBadDate = errors.New("bad date")
	// ErrBadDecimal reports a value that is no decimal.
	ErrBadDecimal = errors.New("bad decimal")
)

// segment is the basket's value from a day until the next segment's, as
// text. They are worked out from modelseries' levels: on December 3 the
// basket holds 0.6 units of MODA (60 of 100 at 100) and 0.8 of MODB (40 of
// 100 at 50), worth 0.6 A + 0.8 B; at the end of December 31, when A is 110
// and B 45 and the basket is worth 66 + 36 = 102, it is rebalanced to
// 60:40, so it holds 61.2/110 units of MODA and 40.8/45 of MODB, worth
// 61.2 a + 40.8 b, with a = A/110 and b = B/45.
//
//	Dec 3  A 100 B 50          0.6*100 + 0.8*50          = 100
//	Dec 10 A 120 B 50          0.6*120 + 0.8*50          = 112
//	Dec 12 A 120 B 60          72 + 48                   = 120
//	Dec 22 A  96 B 60          57.6 + 48                 = 105.6
//	Dec 29 A  96 B 45          57.6 + 36                 = 93.6
//	Dec 31 A 110 B 45          66 + 36                   = 102, rebalanced
//	Jan 15 A 132 B 45          61.2*1.2 + 40.8           = 114.24
//	Jan 20 A 132 B 54          73.44 + 40.8*1.2          = 122.4
//	Feb 2  A  99 B 54          61.2*0.9 + 48.96          = 104.04
//	Feb 16 A  99 B 27          55.08 + 40.8*0.6          = 79.56
//	Mar 2  A 121 B 27          61.2*1.1 + 24.48          = 91.8
//	Mar 9  A 121 B 45          67.32 + 40.8              = 108.12
//
// Without the rebalance the basket would be worth 0.6 A + 0.8 B to the
// end, 81 on February 16 and 108.6 on March 13.
type segment struct {
	from, value string
}

// segments are the basket's values, in date order.
func segments() []segment {
	return []segment{
		{"2025-12-03", "100"},
		{"2025-12-10", "112"},
		{dec12, "120"},
		{"2025-12-22", "105.6"},
		{"2025-12-29", "93.6"},
		{"2025-12-31", "102"},
		{jan15, "114.24"},
		{jan20, "122.4"},
		{feb02, "104.04"},
		{feb16, "79.56"},
		{"2026-03-02", "91.8"},
		{"2026-03-09", "108.12"},
	}
}

// Day is the backtest at the end of one day, its index as a canonical
// decimal.
type Day struct {
	Date, Index string
}

// Line returns the backtest's index for each calendar day from the first
// day of the backtest on or after from to its last, 100 on the first, each
// the day's value times 100 over the first day's, rounded half-even to 12
// places. The backtest rebased on a first day is what GetModelPortfolio
// answers from that day (D-53).
func Line(from string) ([]Day, error) {
	first, err := kernel.ParseDate(from)
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %w", ErrBadDate, from, err)
	}
	start, err := kernel.ParseDate(BacktestStart)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBadDate, err)
	}
	end, err := kernel.ParseDate(BacktestEnd)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBadDate, err)
	}
	if first.Before(start) {
		first = start
	}
	if first.After(end) {
		return nil, fmt.Errorf("%w: %s is after the backtest's end %s", ErrBadDate, first, end)
	}
	base, err := valueOn(first)
	if err != nil {
		return nil, err
	}
	var days []Day
	for d := first; !d.After(end); d = d.AddDays(1) {
		v, err := valueOn(d)
		if err != nil {
			return nil, err
		}
		index, err := v.Mul(kernel.NewDecimal(baseIndex, 0)).Quo(base, indexPlaces, kernel.RoundHalfEven)
		if err != nil {
			return nil, fmt.Errorf("rebasing %s: %w", d, err)
		}
		days = append(days, Day{Date: d.String(), Index: index.String()})
	}
	return days, nil
}

// valueOn returns the basket's value on d, a day of the backtest: the
// latest segment's on or before it.
func valueOn(d kernel.Date) (kernel.Decimal, error) {
	var out kernel.Decimal
	found := false
	for _, s := range segments() {
		at, err := kernel.ParseDate(s.from)
		if err != nil {
			return kernel.Decimal{}, fmt.Errorf("%w: %w", ErrBadDate, err)
		}
		if at.After(d) {
			break
		}
		if out, err = kernel.ParseDecimal(s.value); err != nil {
			return kernel.Decimal{}, fmt.Errorf("%w: segment of %s: %w", ErrBadDecimal, s.from, err)
		}
		found = true
	}
	if !found {
		return kernel.Decimal{}, fmt.Errorf("%w: %s is before the backtest's start", ErrBadDate, d)
	}
	return out, nil
}

// Aligned returns the model's line from the first day on or after from on
// each of labels, the days of a performance chart: its index on a day the
// backtest has, "" on one it hasn't, as the page draws the second line.
func Aligned(labels []string, from string) ([]string, error) {
	days, err := Line(from)
	if err != nil {
		return nil, err
	}
	byDate := make(map[string]string, len(days))
	for _, d := range days {
		byDate[d.Date] = d.Index
	}
	out := make([]string, len(labels))
	for i, l := range labels {
		out[i] = byDate[l]
	}
	return out, nil
}

// PerformanceChart is the overview's performance chart of GetPerformance's
// answer while the model portfolio is on: the time-weighted return's line
// and the model's, rebased on the performance's first day (D-53, D-106).
func PerformanceChart(p *viewsv1.GetPerformanceResponse) (chartcheck.Chart, error) {
	c := chartcheck.PerformanceChart(p)
	model, err := Aligned(c.Labels, p.GetFrom())
	if err != nil {
		return chartcheck.Chart{}, err
	}
	c.Series = append(c.Series, model)
	return c, nil
}

// Target is a model target as the read API names it, with its weight.
type Target struct {
	Kind, Index, Currency, Weight string
}

// Drawdown is a max drawdown's peak and trough with their values and the
// fall as a fraction of the peak, each value as a canonical decimal.
type Drawdown struct {
	Peak, Trough, PeakValue, TroughValue, Depth string
}

// History is one target's own history in the reporting currency.
type History struct {
	Target      Target
	First, Last string
	Gap         string
	Drawdown    Drawdown
}

// Summary is GetModelPortfolio's answer reduced to what the check
// compares, its decimals canonical.
type Summary struct {
	From, To, BacktestStart, BacktestEnd, Reporting string
	Targets                                         []Target
	Days                                            []Day
	Drawdown                                        Drawdown
	Histories                                       []History
}

// targets are the two targets as the read API lists them.
func targets() []Target {
	index := viewsv1.ModelTargetKind_MODEL_TARGET_KIND_INDEX.String()
	return []Target{
		{Kind: index, Index: "MODA", Currency: Reporting, Weight: WeightA},
		{Kind: index, Index: "MODB", Currency: Reporting, Weight: WeightB},
	}
}

// Expected is the answer GetModelPortfolio must give from the first day on
// or after from, through the backtest's end: the model's line rebased on
// that day, the model's max drawdown over the whole backtest, 122.4 on
// Tuesday, January 20, 2026 to 79.56 on Monday, February 16, a fall of 0.35
// (the basket's earlier fall, 120 on December 12 to 93.6 on December 29,
// is 0.22), and each target's own: MODA's from 132 on January 15 to 99 on
// February 2, 0.25 (its earlier fall, 120 to 96, is 0.2), MODB's from 60 on
// December 12 to 27 on February 16, 0.55 (its falls to 45 on December 29
// and from 54 on January 20 to 27 are 0.25 and 0.5). MODA has levels from
// December 1, MODB from December 3; both end on March 13.
func Expected(from string) (Summary, error) {
	days, err := Line(from)
	if err != nil {
		return Summary{}, err
	}
	ts := targets()
	return Summary{
		From: days[0].Date, To: days[len(days)-1].Date,
		BacktestStart: BacktestStart, BacktestEnd: BacktestEnd, Reporting: Reporting,
		Targets: ts, Days: days,
		Drawdown: Drawdown{Peak: jan20, Trough: feb16, PeakValue: "122.4", TroughValue: "79.56", Depth: "0.35"},
		Histories: []History{
			{Target: ts[0], First: "2025-12-01", Last: BacktestEnd, Gap: noGap(), Drawdown: Drawdown{
				Peak: jan15, Trough: feb02, PeakValue: "132", TroughValue: "99", Depth: "0.25",
			}},
			{Target: ts[1], First: BacktestStart, Last: BacktestEnd, Gap: noGap(), Drawdown: Drawdown{
				Peak: dec12, Trough: feb16, PeakValue: "60", TroughValue: "27", Depth: "0.55",
			}},
		},
	}, nil
}

// noGap is the gap of a target that has history.
func noGap() string { return viewsv1.ModelHistoryGap_MODEL_HISTORY_GAP_UNSPECIFIED.String() }

// Summarize reduces GetModelPortfolio's answer to a Summary.
func Summarize(r *viewsv1.GetModelPortfolioResponse) Summary {
	s := Summary{
		From: r.GetFrom(), To: r.GetTo(), BacktestStart: r.GetBacktestStart(), BacktestEnd: r.GetBacktestEnd(),
		Reporting: r.GetReportingCurrency(), Drawdown: drawdownOf(r.GetDrawdown()),
	}
	for _, t := range r.GetTargets() {
		s.Targets = append(s.Targets, targetOf(t))
	}
	for _, d := range r.GetDays() {
		s.Days = append(s.Days, Day{Date: d.GetDate(), Index: canonical(d.GetIndex())})
	}
	for _, h := range r.GetTargetHistories() {
		s.Histories = append(s.Histories, History{
			Target: targetOf(h.GetTarget()), First: h.GetFirstDate(), Last: h.GetLastDate(),
			Gap: h.GetGap().String(), Drawdown: drawdownOf(h.GetDrawdown()),
		})
	}
	return s
}

// targetOf reduces a model target.
func targetOf(t *viewsv1.ModelTarget) Target {
	return Target{Kind: t.GetKind().String(), Index: t.GetIndex(), Currency: t.GetCurrency(), Weight: canonical(t.GetWeightPercent())}
}

// drawdownOf reduces a drawdown; an unset one is empty.
func drawdownOf(d *viewsv1.Drawdown) Drawdown {
	return Drawdown{
		Peak: d.GetPeakDate(), Trough: d.GetTroughDate(), PeakValue: canonical(d.GetPeakValue()),
		TroughValue: canonical(d.GetTroughValue()), Depth: canonical(d.GetDepth()),
	}
}

// canonical writes a decimal without trailing zeros, as the core does; a
// string that is no decimal is kept, so a difference shows.
func canonical(s string) string {
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		return s
	}
	return d.String()
}

// DrawdownRows are the cells of the max drawdown table the performance
// chart's section shows while the model is on, a row per model and target
// (D-54): its name, the fall in percent rounded to one decimal place, the
// peak and trough days and the span measured, as the page writes them.
func DrawdownRows() [][]string {
	span := func(first, last string) string { return first + " to " + last }
	return [][]string{
		{"Model portfolio", "35%", jan20, feb16, span(BacktestStart, BacktestEnd)},
		{"MODA (index, PLN)", "25%", jan15, feb02, span("2025-12-01", BacktestEnd)},
		{"MODB (index, PLN)", "55%", dec12, feb16, span(BacktestStart, BacktestEnd)},
	}
}
