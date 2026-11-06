package ui

import (
	"net/http"
	"slices"
	"testing"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// TestOverviewChartStart checks the charts' start on the overview (D-138):
// a start the owner chose goes to GetPerformance, the model's line starts
// on the day the portfolio's does, the value chart over the same days, and
// a link leads back to the default; a malformed one is refused.
func TestOverviewChartStart(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.read.locked(func() { h.read.modelOn = true })
	rec := h.serve(t, withSession(request(http.MethodGet, "/?start=2026-03-31", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`name="start" type="date" value="2026-03-31"`,
		`Back to the default, <span class="mono">2026-03-30</span>`, `href="/?date=2026-03-31"`,
		`from=2026-03-31`, `start=2026-03-31`, `mark=2026-03-30`,
	)
	rec = h.serve(t, withSession(request(http.MethodGet, "/?start=2026-03-31&over=performance", ""), readSession))
	wantBody(t, rec, "time-weighted return as an index, 100 on 2026-03-31", `aria-current="page">Performance</a>`)
	h.read.locked(func() {
		chosen := slices.ContainsFunc(h.read.performance, func(p *viewsv1.GetPerformanceRequest) bool {
			return p.GetFrom() == "2026-03-31" && p.GetAccountId() == ""
		})
		if !chosen {
			t.Errorf("GetPerformance requests = %v, want the portfolio's from 2026-03-31", h.read.performance)
		}
		if n := len(h.read.model); n == 0 || h.read.model[n-1].GetFrom() != "2026-03-31" {
			t.Errorf("GetModelPortfolio requests = %v, want from 2026-03-31", h.read.model)
		}
	})

	rec = h.serve(t, withSession(request(http.MethodGet, "/", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "Back to the default")

	rec = h.serve(t, withSession(request(http.MethodGet, "/?start=31.03.2026", ""), readSession))
	wantStatus(t, rec, http.StatusBadRequest)
}

// TestAccountCharts checks an account's time charts (D-138): its own
// performance and value over time, by default from its first day, with
// the start control on its page.
func TestAccountCharts(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID+"?start=2026-03-31", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`<h2 id="account-charts-title">Over time</h2>`,
		`action="/accounts/acc-synthetic-1"`, `name="start" type="date" value="2026-03-31"`,
		"By default the charts start on the account's first day",
		`&#34;label&#34;:&#34;Account (time-weighted return)&#34;`,
		"the account's time-weighted return as an index, 100 on 2026-03-31",
		"The account's gross, net and liquidation values",
		"/static/vendor/chart-4.5.1.umd.min.js", "/static/app.js",
	)
	h.read.locked(func() {
		var asked *viewsv1.GetPerformanceRequest
		if n := len(h.read.performance); n > 0 {
			asked = h.read.performance[n-1]
		}
		if asked.GetAccountId() != accountID || asked.GetFrom() != "2026-03-31" {
			t.Errorf("GetPerformance request = %v, want %s from 2026-03-31", asked, accountID)
		}
	})
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID+"?start=yesterday", ""), readSession))
	wantStatus(t, rec, http.StatusBadRequest)
}
