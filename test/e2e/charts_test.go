//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/test/e2e/chartcheck"
)

// The chart check of D-200's criterion 8 on the data path's record
// (SCEN-e2e-charts-match-read-api, SCEN-e2e-charts-only-read-api).
const (
	// chartAttempts is how often a page is looked at again while the
	// record changes between the page and the read API's answer, as the
	// read API check retries its Connect calls.
	chartAttempts = 5
	// chartPageTimeout bounds the looks at one page.
	chartPageTimeout = 5 * time.Minute
	// logSlack is how far before the first page the core's log is read.
	logSlack = 2 * time.Second
	// chartPageCount is how many pages checkCharts looks at: the overview
	// three times, its allocation by four dimensions of three values, three
	// accounts and three instruments.
	chartPageCount = 3 + 4*3 + 3 + 3
)

// errChartsDiffer reports a page whose charts kept drawing something
// else than the read API answers.
var errChartsDiffer = errors.New("the page's charts differ from the read API")

// uiNonReadCalls are the calls the built-in UI may make to the core
// outside the read API while it draws a chart page: the AuthService, which
// checks the session (WhoAmI) and holds nothing of the record (D-69),
// StatementService's ListStatementAccounts, which an account page asks
// whether the account takes uploaded files (D-109), and TickerService's
// ListConfirmedTickers, which an instrument page asks for the tickers the
// owner confirmed (D-116). None of them answers a value, a price, a return
// or an allocation, so every number a chart draws comes from the read API.
func uiNonReadCalls() []string {
	return []string{
		"/" + identityv1connect.AuthServiceName + "/",
		ingestv1connect.StatementServiceListStatementAccountsProcedure,
		ingestv1connect.TickerServiceListConfirmedTickersProcedure,
	}
}

// chartPage is a page of the UI with charts, and how to work out from the
// read API what its charts must draw. need says what the page must draw
// for the check to mean something: a line chart with days, or bars. With
// checkModel, the page's model portfolio table must be model's rows, or,
// with none, the page must have no model portfolio section at all
// (model_test.go).
type chartPage struct {
	path       string
	expect     func(ctx context.Context) (drawnCharts, error)
	need       string
	checkModel bool
	model      [][]string
}

// drawnCharts is what a page's charts draw: its line charts in page order
// and its allocation's bars.
type drawnCharts struct {
	Charts []chartcheck.Chart
	Bars   []chartcheck.Bar
}

// pageLook is owner.mjs's answer to a page request.
type pageLook struct {
	Status int    `json:"status"`
	Path   string `json:"path"`
	Charts []struct {
		chartcheck.Chart

		Aria string `json:"aria"`
	} `json:"charts"`
	Bars []chartcheck.Bar `json:"bars"`
	// Model is the cells of the model portfolio's max drawdown table, row
	// by row; null when the page has no model portfolio section.
	Model [][]string `json:"model"`
}

// pageWindow is when the UI drew a page.
type pageWindow struct {
	path   string
	window chartcheck.Window
}

// The pages' needs.
const (
	needLine = "a line chart with days"
	needBars = "allocation bars"
	needNone = ""
)

// checkCharts checks criterion 8's charts on the data path's record, as
// the signed-in owner sees them in Chromium (owner.mjs): the overview's
// value over time, gross and all three values, its time-weighted return
// and its allocation by every dimension of every value, on weekEnd; each
// data path account's value and return, the IBKR and XTB accounts on
// weekEnd and the bond account on bondsAutumn; and each instrument's
// prices on weekEnd. What each Chart.js chart holds once app.js drew it,
// and each allocation bar's value, share and width, must equal what the
// read API answers the script token for the same days, worked out the way
// the UI asks (chartcheck), a page looked at again up to chartAttempts
// times while the record changes (SCEN-e2e-charts-match-read-api). Then
// the core's log, over the spans the UI drew each page in, shows the UI
// called nothing outside the read API but uiNonReadCalls and made read API
// requests, and the UI's own pod can reach nothing but the core's plugin
// port (SCEN-e2e-charts-only-read-api).
func (p *dataPath) checkCharts(t *testing.T) {
	t.Helper()
	format := p.numberFormat(t)
	started := time.Now()
	pages := p.chartPages()
	windows := make([]pageWindow, 0, len(pages))
	for _, page := range pages {
		windows = append(windows, p.checkChartPage(t, &page, format)...)
	}
	checkUICalls(t, p.k, started, windows)
	checkUIReachesOnlyCore(t, p.k)
}

// numberFormat reads the owner's number format, which the pages write
// amounts in (D-113).
func (p *dataPath) numberFormat(t *testing.T) *viewsv1.NumberFormat {
	t.Helper()
	var settings viewsv1.GetSettingsResponse
	err := waitFor(t.Context(), coreTimeout, func(ctx context.Context) (string, bool, error) {
		if err := p.read(ctx, "/v1/settings", &settings); err != nil {
			return "", false, err
		}
		return "read", true, nil
	})
	if err != nil {
		t.Fatalf("reading the settings: %v", err)
	}
	return settings.GetNumberFormat()
}

// chartPages are the pages checkCharts looks at.
func (p *dataPath) chartPages() []chartPage {
	gross := []viewsv1.ValueKind{viewsv1.ValueKind_VALUE_KIND_GROSS}
	pages := make([]chartPage, 0, chartPageCount)
	pages = append(pages, []chartPage{
		{path: "/?date=" + weekEnd, expect: p.overviewValue(weekEnd, gross), need: needLine},
		{path: "/?chart=all&date=" + weekEnd, expect: p.overviewValue(weekEnd, chartcheck.ValueKinds()), need: needLine},
		{path: "/?date=" + weekEnd + "&over=performance", expect: p.overviewPerformance(weekEnd), need: needLine},
	}...)
	dims := map[string]viewsv1.AllocationDimension{
		"asset-class": viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS,
		"currency":    viewsv1.AllocationDimension_ALLOCATION_DIMENSION_CURRENCY,
		"institution": viewsv1.AllocationDimension_ALLOCATION_DIMENSION_INSTITUTION,
		"sleeve":      viewsv1.AllocationDimension_ALLOCATION_DIMENSION_SLEEVE,
	}
	values := map[string]viewsv1.ValueKind{
		"gross": viewsv1.ValueKind_VALUE_KIND_GROSS, "net": viewsv1.ValueKind_VALUE_KIND_NET,
		"liquidation": viewsv1.ValueKind_VALUE_KIND_LIQUIDATION,
	}
	for _, by := range []string{"asset-class", "currency", "institution", "sleeve"} {
		for _, value := range []string{"gross", "net", "liquidation"} {
			q := url.Values{"by": {by}, "date": {weekEnd}, "over": {"allocation"}, "value": {value}}
			pages = append(pages, chartPage{path: "/?" + q.Encode(), expect: p.allocation(weekEnd, dims[by], values[value]), need: needBars})
		}
	}
	for _, a := range []struct{ id, date, need string }{
		{p.ibkr, weekEnd, needLine}, {p.xtb, weekEnd, needLine}, {p.bonds, bondsAutumn, needNone},
	} {
		pages = append(pages, chartPage{path: "/accounts/" + a.id + "?date=" + a.date, expect: p.accountCharts(a.id, a.date), need: a.need})
	}
	for _, isin := range []string{isinG, isinH, isinJ} {
		pages = append(pages, chartPage{path: "/instruments/" + isin + "?date=" + weekEnd, expect: p.instrumentCharts(isin, weekEnd), need: needLine})
	}
	return pages
}

// checkChartPage looks at page until what it draws equals what the read
// API answers, at most chartAttempts times within chartPageTimeout, and
// returns when the UI drew it each time.
func (p *dataPath) checkChartPage(t *testing.T, page *chartPage, format *viewsv1.NumberFormat) []pageWindow {
	t.Helper()
	var windows []pageWindow
	err := waitFor(t.Context(), chartPageTimeout, func(ctx context.Context) (string, bool, error) {
		from := time.Now()
		look := p.owner.look(t, page.path)
		windows = append(windows, pageWindow{path: page.path, window: chartcheck.Window{From: from, To: time.Now()}})
		problem := comparePage(ctx, page, &look, format)
		switch {
		case problem == "":
			return "drawn as the read API answers", true, nil
		case len(windows) >= chartAttempts:
			return "", false, fmt.Errorf("%w after %d looks: %s", errChartsDiffer, len(windows), problem)
		}
		t.Logf("charts: %s, look %d: %s; looking again", page.path, len(windows), problem)
		return problem, false, errRetry
	})
	if err != nil {
		t.Errorf("charts: %s: %v", page.path, err)
		return windows
	}
	t.Logf("charts: %s draws what the read API answers", page.path)
	return windows
}

// comparePage compares what look drew with what page's charts must draw,
// and returns what differs, or "" when nothing does.
func comparePage(ctx context.Context, page *chartPage, look *pageLook, format *viewsv1.NumberFormat) string {
	if look.Status != http.StatusOK || look.Path != page.path {
		return fmt.Sprintf("status %d at %s, want 200 at %s", look.Status, look.Path, page.path)
	}
	want, err := page.expect(ctx)
	if err != nil {
		return "asking the read API: " + err.Error()
	}
	got, err := look.drawn(format)
	if err != nil {
		return err.Error()
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		return "the page draws something else than the read API answers (-read API +page):\n" + diff
	}
	if page.checkModel {
		if diff := cmp.Diff(page.model, look.Model, cmpopts.EquateEmpty()); diff != "" {
			return "the page's model portfolio table differs (-want +page):\n" + diff
		}
	}
	return missing(page.need, &got)
}

// drawn is what l drew, each bar's value as the read API writes it.
func (l *pageLook) drawn(format *viewsv1.NumberFormat) (drawnCharts, error) {
	got := drawnCharts{}
	for _, c := range l.Charts {
		got.Charts = append(got.Charts, c.Chart)
	}
	for _, b := range l.Bars {
		value, err := chartcheck.PlainMoney(b.Value, format)
		if err != nil {
			return drawnCharts{}, fmt.Errorf("bar %s: %w", b.Key, err)
		}
		b.Value = value
		got.Bars = append(got.Bars, b)
	}
	return got, nil
}

// missing says what got lacks of need, or "" when it lacks nothing.
func missing(need string, got *drawnCharts) string {
	switch need {
	case needLine:
		if len(got.Charts) == 0 || len(got.Charts[0].Labels) == 0 {
			return "the page draws no " + needLine
		}
	case needBars:
		if len(got.Bars) == 0 {
			return "the page draws no " + needBars
		}
	}
	return ""
}

// drawnOnly keeps the charts that have days: the UI draws no canvas for a
// chart without any.
func drawnOnly(charts ...chartcheck.Chart) []chartcheck.Chart {
	var out []chartcheck.Chart
	for _, c := range charts {
		if len(c.Labels) > 0 {
			out = append(out, c)
		}
	}
	return out
}

// performance reads the time-weighted return up to date from the
// charts' start, of the portfolio or, with account, of that account.
func (p *dataPath) performance(ctx context.Context, account, date string) (*viewsv1.GetPerformanceResponse, error) {
	q := url.Values{"to": {date}}
	if account != "" {
		q.Set("account_id", account)
	}
	var perf viewsv1.GetPerformanceResponse
	if err := p.read(ctx, "/v1/performance?"+q.Encode(), &perf); err != nil {
		return nil, err
	}
	return &perf, nil
}

// overviewValue is the overview's value chart of kinds on date: the
// portfolio's totals from the performance's first day (D-138).
func (p *dataPath) overviewValue(date string, kinds []viewsv1.ValueKind) func(context.Context) (drawnCharts, error) {
	return func(ctx context.Context) (drawnCharts, error) {
		perf, err := p.performance(ctx, "", date)
		if err != nil {
			return drawnCharts{}, err
		}
		first := ""
		if days := perf.GetDays(); len(days) > 0 {
			first = days[0].GetDate()
		}
		from, err := chartcheck.ChartFrom(first, date)
		if err != nil {
			return drawnCharts{}, fmt.Errorf("the value chart's start: %w", err)
		}
		var totals viewsv1.ListValuationTotalsResponse
		if err := p.read(ctx, "/v1/valuations?"+url.Values{"from": {from}, "to": {date}}.Encode(), &totals); err != nil {
			return drawnCharts{}, err
		}
		return drawnCharts{Charts: drawnOnly(chartcheck.ValueChart(totals.GetTotals(), kinds))}, nil
	}
}

// overviewPerformance is the overview's performance chart on date; the
// data path's model portfolio is off, so it draws no backtest (D-52).
func (p *dataPath) overviewPerformance(date string) func(context.Context) (drawnCharts, error) {
	return func(ctx context.Context) (drawnCharts, error) {
		perf, err := p.performance(ctx, "", date)
		if err != nil {
			return drawnCharts{}, err
		}
		return drawnCharts{Charts: drawnOnly(chartcheck.PerformanceChart(perf))}, nil
	}
}

// allocation is the overview's allocation of value on date by dimension.
func (p *dataPath) allocation(date string, by viewsv1.AllocationDimension, value viewsv1.ValueKind) func(context.Context) (drawnCharts, error) {
	return func(ctx context.Context) (drawnCharts, error) {
		var resp viewsv1.GetAllocationResponse
		q := url.Values{"date": {date}, "by": {by.String()}, "value": {value.String()}}
		if err := p.read(ctx, "/v1/allocation?"+q.Encode(), &resp); err != nil {
			return drawnCharts{}, err
		}
		bars, err := chartcheck.Allocation(&resp)
		if err != nil {
			return drawnCharts{}, fmt.Errorf("the allocation's bars: %w", err)
		}
		return drawnCharts{Bars: bars}, nil
	}
}

// accountCharts are an account page's charts on date: its three values
// from its performance's start, then its time-weighted return.
func (p *dataPath) accountCharts(account, date string) func(context.Context) (drawnCharts, error) {
	return func(ctx context.Context) (drawnCharts, error) {
		perf, err := p.performance(ctx, account, date)
		if err != nil {
			return drawnCharts{}, err
		}
		from, err := chartcheck.ChartFrom(perf.GetFrom(), date)
		if err != nil {
			return drawnCharts{}, fmt.Errorf("the account chart's start: %w", err)
		}
		var vals viewsv1.ListAccountValuationsResponse
		path := "/v1/accounts/" + account + "/valuations?" + url.Values{"from": {from}, "to": {date}}.Encode()
		if err := p.read(ctx, path, &vals); err != nil {
			return drawnCharts{}, err
		}
		return drawnCharts{Charts: drawnOnly(
			chartcheck.ValueChart(vals.GetTotals(), chartcheck.ValueKinds()), chartcheck.PerformanceChart(perf),
		)}, nil
	}
}

// instrumentCharts are an instrument page's price charts on date, one per
// listing, over the year before it.
func (p *dataPath) instrumentCharts(isin, date string) func(context.Context) (drawnCharts, error) {
	return func(ctx context.Context) (drawnCharts, error) {
		from, err := chartcheck.DaysBefore(date, chartcheck.InstrumentPriceDays)
		if err != nil {
			return drawnCharts{}, fmt.Errorf("the price chart's start: %w", err)
		}
		var prices viewsv1.ListInstrumentPricesResponse
		path := "/v1/instruments/" + isin + "/prices?" + url.Values{"from": {from}, "to": {date}}.Encode()
		if err := p.read(ctx, path, &prices); err != nil {
			return drawnCharts{}, err
		}
		return drawnCharts{Charts: drawnOnly(chartcheck.PriceCharts(prices.GetPrices())...)}, nil
	}
}

// checkUICalls reads the core's log since started and checks, for each
// span the UI drew a page in, that the UI plugin called the core outside
// the read API only for uiNonReadCalls, and that the read API logged
// requests in it. The core logs every call of a service only plugins
// reach with its procedure and the calling plugin, and every read API
// request without either (cmd/core); the UI reaches nothing else
// (checkUIReachesOnlyCore).
func checkUICalls(t *testing.T, k *kube, started time.Time, windows []pageWindow) {
	t.Helper()
	calls, err := chartcheck.Calls(coreLogs(t, k, started.Add(-logSlack)))
	if err != nil {
		t.Fatal(err)
	}
	allowed := uiNonReadCalls()
	for _, w := range windows {
		procs := chartcheck.PluginCalls(calls, uiPlugin, w.window)
		if bad := chartcheck.Disallowed(procs, allowed); len(bad) > 0 {
			t.Errorf("while drawing %s the UI called %v outside the read API; it may call only %v", w.path, bad, allowed)
		}
		if chartcheck.ReadRequests(calls, w.window) == 0 {
			t.Errorf("while drawing %s the core logged no read API request", w.path)
		}
	}
	t.Logf("charts: the core's log shows the UI calling, outside the read API, only %v over %d page loads", allowed, len(windows))
}

// coreLogs returns the log lines every core pod wrote since since.
func coreLogs(t *testing.T, k *kube, since time.Time) []byte {
	t.Helper()
	pods, err := k.clients.CoreV1().Pods(namespace).List(t.Context(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=" + coreDeployment})
	if err != nil {
		t.Fatalf("listing the core's pods: %v", err)
	}
	start := metav1.NewTime(since)
	var logs []byte
	for i := range pods.Items {
		out, logErr := k.clients.CoreV1().Pods(namespace).GetLogs(pods.Items[i].Name, &corev1.PodLogOptions{SinceTime: &start}).DoRaw(t.Context())
		if logErr != nil {
			t.Fatalf("reading the log of the core's pod %s: %v", pods.Items[i].Name, logErr)
		}
		logs = append(logs, out...)
	}
	return logs
}

// checkUIReachesOnlyCore checks what the operator made for the built-in
// UI: its NetworkPolicy lets it reach the core's plugin port alone, and it
// has no proxy settings, no egress token, no host on its allow-list and no
// effective egress host (D-32, D-63), so the core is the only place its
// pages' data can come from.
func checkUIReachesOnlyCore(t *testing.T, k *kube) {
	t.Helper()
	core := k.getService(t, coreService)
	want := &pluginNetwork{
		Egress:       []string{fmt.Sprintf("pods %s: TCP/%d", selectorString(core.Spec.Selector), podTargetPort(t, core, "connect"))},
		EgressPolicy: true, TokenAudiences: []string{plugins.CoreAudience}, Listed: true,
	}
	if diff := cmp.Diff(want, networkOf(t, k, uiPlugin), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("the UI plugin's network (-want +got):\n%s", diff)
	}
}

// look opens path in the owner's signed-in tab and returns what its charts
// draw.
func (o *owner) look(t *testing.T, path string) pageLook {
	t.Helper()
	var l pageLook
	if err := json.Unmarshal([]byte(o.roundTrip(t, &ownerRequest{Page: path})), &l); err != nil {
		t.Fatalf("GET %s: decoding the answer: %v", path, err)
	}
	return l
}
