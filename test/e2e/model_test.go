//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/test/e2e/modelcheck"
)

// modelAttempts is how often the model portfolio is asked for again while
// the record changes between the answer and the hand calculation, as
// chartAttempts is for a page.
const modelAttempts = 5

// modelFirstDays returns the first days the model portfolio is asked from: the
// backtest's own first day, December 3, 2025, which is the later of the two
// index series' first days and before the data path's first transaction
// (the bond lot of October 6, 2025 is before it), so the answer spans the
// year end that rebalances the basket; and the go-live day, so it is
// rebased on a day after the rebalance.
func modelFirstDays() []string { return []string{modelcheck.BacktestStart, goLive} }

// checkModelPortfolio checks D-200's criterion 19 on the data path
// (SCEN-e2e-model-portfolio-off-shows-nothing,
// SCEN-e2e-model-portfolio-backtest-hand-calculation), after the checks of
// criteria 7 and 8, which run with the model portfolio off. The owner
// saves a model portfolio of two index targets in PLN, MODA at 60% and MODB
// at 40%, with the settings page's form, as the owner does (owner.mjs):
//
//   - off: with nothing saved, and again with the targets saved but the box
//     unchecked, the read API answers 404 and the overview's performance
//     chart draws the time-weighted return alone, with no model portfolio
//     section;
//   - on: once an eodhd sync, which the core now tells to fetch the two
//     indices' levels (D-93), has stored them, the read API's answer from
//     December 3, 2025 and from March 6, 2026 equals the hand calculation
//     of modelcheck over the fixture's levels (modelseries): the backtest's
//     index on each day across the yearly rebalance of December 31, to 12
//     places, the model's max drawdown, 0.35 from January 20 to February
//     16, and each target's, MODA's 0.25 from January 15 to February 2 and
//     MODB's 0.55 from December 12 to February 16, with their dates; and
//     the overview's performance chart draws the model as a second line
//     over the same days and the page lists the three max drawdowns, the
//     read API's answer worked out as the UI asks (modelcheck);
//   - off again: the model portfolio the owner turned off disappears from
//     the read API and the page again, its targets kept.
func (p *dataPath) checkModelPortfolio(t *testing.T) {
	t.Helper()
	format := p.numberFormat(t)
	p.checkModelOff(t, format, "never saved")
	p.saveModel(t, false)
	p.checkModelOff(t, format, "saved, unchecked")
	p.saveModel(t, true)
	p.waitModelBacktest(t)
	for _, from := range modelFirstDays() {
		p.checkModelBacktest(t, from)
	}
	page := &chartPage{
		path: "/?date=" + weekEnd + "&over=performance", expect: p.overviewPerformanceWithModel(weekEnd), need: needLine,
		checkModel: true, model: modelcheck.DrawdownRows(),
	}
	p.checkChartPage(t, page, format)
	p.saveModel(t, false)
	p.checkModelOff(t, format, "turned off again")
}

// modelPath is the read API's model portfolio from first through weekEnd.
func modelPath(from string) string {
	return "/v1/model-portfolio?" + url.Values{"from": {from}, "to": {weekEnd}}.Encode()
}

// modelFields are the settings page's model portfolio form's fields for
// the two index targets, in the order the form sends them: the box checked
// when enabled, then each target's kind, index, currency and weight.
func modelFields(enabled bool) [][2]string {
	var fields [][2]string
	if enabled {
		fields = append(fields, [2]string{"enabled", "on"})
	}
	for _, target := range [][2]string{{"MODA", modelcheck.WeightA}, {"MODB", modelcheck.WeightB}} {
		fields = append(fields,
			[2]string{"kind", "index"}, [2]string{"id", target[0]},
			[2]string{"currency", modelcheck.Reporting}, [2]string{"weight", target[1]})
	}
	return fields
}

// saveModel saves the model portfolio, on or off, through the settings
// page's form.
func (p *dataPath) saveModel(t *testing.T, enabled bool) {
	t.Helper()
	p.owner.postRequest(t, &ownerRequest{Path: "/settings/model", Fields: modelFields(enabled)})
}

// checkModelOff checks that the model portfolio, in the state state, is
// nowhere: the read API answers 404 for it, and the overview's performance
// chart draws the time-weighted return alone with no model portfolio
// section (D-52).
func (p *dataPath) checkModelOff(t *testing.T, format *viewsv1.NumberFormat, state string) {
	t.Helper()
	err := waitFor(t.Context(), dataPathTimeout, func(ctx context.Context) (string, bool, error) {
		r, err := call(ctx, p.k, clientPod, modelPath(goLive), p.token)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		return fmt.Sprintf("GET %s answers %d", modelPath(goLive), r.status), r.status == http.StatusNotFound, nil
	})
	if err != nil {
		t.Errorf("the model portfolio, %s: the read API should answer 404: %v", state, err)
	}
	page := &chartPage{
		path: "/?date=" + weekEnd + "&over=performance", expect: p.overviewPerformance(weekEnd), need: needLine,
		checkModel: true,
	}
	p.checkChartPage(t, page, format)
}

// waitModelBacktest syncs the eodhd instance until the read API answers the
// model portfolio from the backtest's first day with a backtest: the core
// asks the connector for the levels of MODA and MODB once the model is on,
// and the fixture serves them from December 2025.
func (p *dataPath) waitModelBacktest(t *testing.T) {
	t.Helper()
	err := waitFor(t.Context(), dataPathTimeout, func(ctx context.Context) (string, bool, error) {
		if _, err := p.syncOnce(ctx, dpEODHD); err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		var model viewsv1.GetModelPortfolioResponse
		if err := p.read(ctx, modelPath(modelcheck.BacktestStart), &model); err != nil {
			return "", false, err
		}
		return fmt.Sprintf("backtest from %q to %q over %d days", model.GetBacktestStart(), model.GetBacktestEnd(), len(model.GetDays())),
			model.GetBacktestStart() != "", nil
	})
	if err != nil {
		t.Fatalf("waiting for the model portfolio's backtest after the owner turned it on: %v", err)
	}
}

// checkModelBacktest checks the read API's model portfolio from first
// against the hand calculation, asked again for up to modelAttempts times
// while the record changes under it.
func (p *dataPath) checkModelBacktest(t *testing.T, first string) {
	t.Helper()
	want, err := modelcheck.Expected(first)
	if err != nil {
		t.Fatalf("the hand calculation from %s: %v", first, err)
	}
	diff := ""
	for attempt := 1; attempt <= modelAttempts; attempt++ {
		var model viewsv1.GetModelPortfolioResponse
		if err := p.read(t.Context(), modelPath(first), &model); err != nil {
			t.Fatalf("reading the model portfolio from %s: %v", first, err)
		}
		if diff = cmp.Diff(want, modelcheck.Summarize(&model)); diff == "" {
			break
		}
		t.Logf("model portfolio from %s, look %d: differs from the hand calculation; looking again", first, attempt)
	}
	if diff != "" {
		t.Errorf("the model portfolio from %s through %s (-hand calculation +read API):\n%s", first, weekEnd, diff)
		return
	}
	t.Logf("model portfolio: from %s, %d days of the backtest and the max drawdowns equal the hand calculation", first, len(want.Days))
}

// overviewPerformanceWithModel is the overview's performance chart on date
// while the model portfolio is on: the time-weighted return and the model's
// line from the performance's first day (modelcheck).
func (p *dataPath) overviewPerformanceWithModel(date string) func(context.Context) (drawnCharts, error) {
	return func(ctx context.Context) (drawnCharts, error) {
		perf, err := p.performance(ctx, "", date)
		if err != nil {
			return drawnCharts{}, err
		}
		chart, err := modelcheck.PerformanceChart(perf)
		if err != nil {
			return drawnCharts{}, fmt.Errorf("the model's line: %w", err)
		}
		return drawnCharts{Charts: drawnOnly(chart)}, nil
	}
}
