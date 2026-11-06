package ui

import (
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// performanceView is the performance chart (D-30, D-106): the portfolio's
// or an account's time-weighted return (D-61) from the charts' start
// (D-138) and, while the model portfolio is on, its backtest as a second
// line, both 100 on the first day (D-53).
type performanceView struct {
	From, To string
	// Default is the charts' default start, where tracking began, which
	// the chart marks (D-138).
	Default  string
	Currency string
	Points   int
	// Value is the portfolio's value with cash on To.
	Value *viewsv1.Money
	// Gap says why the line ends before the date asked for; empty when it
	// doesn't.
	Gap     string
	Pending bool
	Chart   string
	// Return is the time-weighted return from From to To in percent;
	// empty without days.
	Return string
	// Model is the model portfolio; nil while it is off (D-52).
	Model *modelView
	// Suggested counts the suggested transfers the owner has not decided
	// on (D-105): until confirmed, they count as cash flows.
	Suggested int
}

// modelView is the model portfolio's targets and max drawdowns (D-54).
type modelView struct {
	BacktestStart, BacktestEnd string
	Targets                    []*viewsv1.ModelTarget
	Drawdowns                  []drawdownRow
}

// drawdownRow is one max drawdown: of the model or of a target.
type drawdownRow struct {
	Name          string
	Depth         string
	Peak, Trough  string
	First, Last   string
	Missing       string
	NeverFell     bool
	PeakValue     string
	TroughValue   string
	ValueCurrency string
}

// chartSeries is one more line on a line chart, its values aligned with
// the chart's labels; an empty value is a day the line has none.
type chartSeries struct {
	Label  string   `json:"label"`
	Values []string `json:"values"`
}

// targetLabel names a model target: its instrument, or its index with its
// currency.
func targetLabel(t *viewsv1.ModelTarget) string {
	if t.GetKind() == viewsv1.ModelTargetKind_MODEL_TARGET_KIND_INDEX {
		return t.GetIndex() + " (index, " + t.GetCurrency() + ")"
	}
	return instrumentLabel(t.GetInstrument())
}

// gapLabel says why a performance series or a target's history stops,
// from its enum name.
func gapLabel(name string) string {
	return map[string]string{
		viewsv1.PerformanceGap_PERFORMANCE_GAP_UNVALUED.String():      "the next day is not fully valued yet",
		viewsv1.PerformanceGap_PERFORMANCE_GAP_NO_FX_RATE.String():    "the next day has no FX rate yet",
		viewsv1.ModelHistoryGap_MODEL_HISTORY_GAP_NO_CLOSES.String():  "no stored prices or index levels",
		viewsv1.ModelHistoryGap_MODEL_HISTORY_GAP_NO_FX_RATE.String(): "no FX rate into the reporting currency",
	}[name]
}

// drawdownOf turns a drawdown into a row named name; a nil one, of no
// history, gives a row with only the name.
func drawdownOf(name string, d *viewsv1.Drawdown) drawdownRow {
	if d == nil {
		return drawdownRow{Name: name}
	}
	return drawdownRow{
		Name: name, Depth: share(d.GetDepth(), "1"), Peak: d.GetPeakDate(), Trough: d.GetTroughDate(),
		NeverFell: d.GetPeakDate() == d.GetTroughDate(), PeakValue: d.GetPeakValue(), TroughValue: d.GetTroughValue(),
	}
}

// buildModel turns GetModelPortfolio's answer into the drawdown table.
func buildModel(m *viewsv1.GetModelPortfolioResponse) *modelView {
	v := &modelView{BacktestStart: m.GetBacktestStart(), BacktestEnd: m.GetBacktestEnd(), Targets: m.GetTargets()}
	model := drawdownOf("Model portfolio", m.GetDrawdown())
	model.First, model.Last = m.GetBacktestStart(), m.GetBacktestEnd()
	v.Drawdowns = append(v.Drawdowns, model)
	for _, h := range m.GetTargetHistories() {
		row := drawdownOf(targetLabel(h.GetTarget()), h.GetDrawdown())
		row.First, row.Last, row.ValueCurrency = h.GetFirstDate(), h.GetLastDate(), m.GetReportingCurrency()
		if h.GetGap() != viewsv1.ModelHistoryGap_MODEL_HISTORY_GAP_UNSPECIFIED {
			row.Missing = gapLabel(h.GetGap().String())
		}
		v.Drawdowns = append(v.Drawdowns, row)
	}
	return v
}

// buildPerformance turns GetPerformance's answer, and the model portfolio's
// when it is on, into the performance chart: the time-weighted return
// index per day and the model's index on the same days, its numbers
// written in f.
func buildPerformance(p *viewsv1.GetPerformanceResponse, m *viewsv1.GetModelPortfolioResponse, f numberFormat) performanceView {
	return buildLine(p, m, f, "Portfolio (time-weighted return)")
}

// buildLine is buildPerformance with the time-weighted return's line
// named label: the portfolio's, or an account's.
func buildLine(p *viewsv1.GetPerformanceResponse, m *viewsv1.GetModelPortfolioResponse, f numberFormat, label string) performanceView {
	chart := chartData{
		Kind: chartLine, Label: label, Currency: "index", Group: f.Group, Decimal: f.Decimal,
		Mark: p.GetDefaultFrom(),
	}
	v := performanceView{Currency: p.GetReportingCurrency(), Pending: p.GetRevaluationPending(), Default: p.GetDefaultFrom()}
	for _, d := range p.GetDays() {
		chart.Labels = append(chart.Labels, d.GetDate())
		chart.Values = append(chart.Values, d.GetIndex())
		v.Value = d.GetValueReporting()
	}
	if p.GetGap() != viewsv1.PerformanceGap_PERFORMANCE_GAP_UNSPECIFIED {
		v.Gap = gapLabel(p.GetGap().String())
	}
	if m != nil {
		v.Model = buildModel(m)
		byDate := map[string]string{}
		for _, d := range m.GetDays() {
			byDate[d.GetDate()] = d.GetIndex()
		}
		model := chartSeries{Label: "Model portfolio (backtest)", Values: make([]string, len(chart.Labels))}
		for i, date := range chart.Labels {
			model.Values[i] = byDate[date]
		}
		chart.Series = []chartSeries{model}
	}
	v.Points = len(chart.Labels)
	if v.Points > 0 {
		v.From, v.To = chart.Labels[0], chart.Labels[v.Points-1]
		v.Return = returnPercent(chart.Values[0], chart.Values[v.Points-1], f)
	}
	v.Chart = chart.encode()
	return v
}

// performance reads the time-weighted return from start, or from the
// charts' default start when start is zero (D-138), up to date, while it
// is on the model portfolio over the same days, and how many suggested
// transfers wait for the owner. A model portfolio that is off answers not
// found: then the page shows none of it (D-52).
func (s *server) performance(r *http.Request, sess session, start, date kernel.Date) (performanceView, error) {
	perf, err := s.read.GetPerformance(r.Context(), bearer(&viewsv1.GetPerformanceRequest{From: optionalString(start), To: date.String()}, sess))
	if err != nil {
		return performanceView{}, fmt.Errorf("GetPerformance: %w", err)
	}
	// The model's line starts on the day the portfolio's does (D-53).
	model, err := s.read.GetModelPortfolio(r.Context(), bearer(&viewsv1.GetModelPortfolioRequest{From: perf.Msg.GetFrom(), To: date.String()}, sess))
	var on *viewsv1.GetModelPortfolioResponse
	switch {
	case err == nil:
		on = model.Msg
	case connect.CodeOf(err) != connect.CodeNotFound:
		return performanceView{}, fmt.Errorf("GetModelPortfolio: %w", err)
	}
	v := buildPerformance(perf.Msg, on, requestFormat(r))
	if v.Suggested, err = s.suggestedTransfers(r, sess); err != nil {
		return performanceView{}, err
	}
	return v, nil
}
