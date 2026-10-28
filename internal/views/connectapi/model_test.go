package connectapi_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// fakeModels reports report while on.
type fakeModels struct {
	on     bool
	report portfolioapp.Report
}

func (f *fakeModels) Report(context.Context, kernel.Date) (portfolioapp.Report, bool, error) {
	return f.report, f.on, nil
}

// syntheticReport is a 60/40 model of ETF1 and the index FAKE20 backtested
// from 2026-03-01 (before go-live, 2026-03-02) to 2026-03-03, with a
// target that has no FX rate.
func syntheticReport(t *testing.T) portfolioapp.Report {
	t.Helper()
	etf := portfolio.Target{Kind: portfolio.TargetInstrument, Name: isinIE + "/ETF1", Weight: dec(t, "60")}
	idx := portfolio.Target{Kind: portfolio.TargetIndex, Name: "FAKE20", Currency: cur(t, "PLN"), Weight: dec(t, "40")}
	point := func(d, v string) portfolio.Point { return portfolio.Point{Date: date(t, d), Value: dec(t, v)} }
	return portfolioapp.Report{
		Model: portfolio.Model{Enabled: true, Targets: []portfolio.Target{etf, idx}},
		Backtest: portfolio.BacktestSeries{Days: []portfolio.Point{
			point("2026-03-01", "100"), point("2026-03-02", "120"), point("2026-03-03", "90"),
		}},
		Drawdown: portfolio.Drawdown{
			Peak: date(t, "2026-03-02"), Trough: date(t, "2026-03-03"), PeakValue: dec(t, "120"), TroughValue: dec(t, "90"), Depth: dec(t, "0.25"),
		},
		Targets: []portfolioapp.TargetReport{
			{Target: etf, First: date(t, "2026-03-01"), Last: date(t, "2026-03-03"), Drawdown: portfolio.Drawdown{
				Peak: date(t, "2026-03-01"), Trough: date(t, "2026-03-01"), PeakValue: dec(t, "40"), TroughValue: dec(t, "40"), Depth: dec(t, "0"),
			}},
			{Target: idx, Missing: portfolio.MissingRates},
		},
		Reporting: cur(t, "PLN"),
	}
}

// modelHandler serves the read API over newLedger with the model portfolio
// on or off.
func modelHandler(t *testing.T, on bool) http.Handler {
	t.Helper()
	v := newViews(t, newLedger(t), app.WithModel(&fakeModels{on: on, report: syntheticReport(t)}))
	h, err := connectapi.New(v, fakeVerifier{},
		connectapi.WithLogger(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	return h
}

func TestConnectModelPortfolio(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(modelHandler(t, true))
	t.Cleanup(srv.Close)
	c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
	got, err := c.GetModelPortfolio(t.Context(), authed(goodToken, &viewsv1.GetModelPortfolioRequest{}))
	if err != nil {
		t.Fatalf("GetModelPortfolio: %v", err)
	}
	etf := &viewsv1.ModelTarget{Kind: viewsv1.ModelTargetKind_MODEL_TARGET_KIND_INSTRUMENT, Instrument: &viewsv1.Instrument{Isin: isinIE, Symbol: "ETF1"}, WeightPercent: "60"}
	idx := &viewsv1.ModelTarget{Kind: viewsv1.ModelTargetKind_MODEL_TARGET_KIND_INDEX, Index: "FAKE20", Currency: "PLN", WeightPercent: "40"}
	// Go-live is 2026-03-02: the line starts there at 100.
	want := &viewsv1.GetModelPortfolioResponse{
		From: "2026-03-02", To: "2026-03-03", Targets: []*viewsv1.ModelTarget{etf, idx},
		BacktestStart: "2026-03-01", BacktestEnd: "2026-03-03",
		Days:     []*viewsv1.BacktestDay{{Date: "2026-03-02", Index: "100"}, {Date: "2026-03-03", Index: "75"}},
		Drawdown: &viewsv1.Drawdown{PeakDate: "2026-03-02", TroughDate: "2026-03-03", PeakValue: "120", TroughValue: "90", Depth: "0.25"},
		TargetHistories: []*viewsv1.ModelTargetHistory{
			{Target: etf, FirstDate: "2026-03-01", LastDate: "2026-03-03", Drawdown: &viewsv1.Drawdown{
				PeakDate: "2026-03-01", TroughDate: "2026-03-01", PeakValue: "40", TroughValue: "40", Depth: "0",
			}},
			{Target: idx, Gap: viewsv1.ModelHistoryGap_MODEL_HISTORY_GAP_NO_FX_RATE},
		},
		ReportingCurrency: "PLN",
	}
	if diff := cmp.Diff(want, got.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("GetModelPortfolio (-want +got):\n%s", diff)
	}
}

func TestModelPortfolioOff(t *testing.T) {
	t.Parallel()
	for _, h := range []http.Handler{modelHandler(t, false), newAPI(t, newLedger(t)).handler} {
		srv := httptest.NewServer(h)
		c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
		_, err := c.GetModelPortfolio(t.Context(), authed(goodToken, &viewsv1.GetModelPortfolioRequest{}))
		if connect.CodeOf(err) != connect.CodeNotFound || err.Error() != "not_found: no model portfolio" {
			t.Errorf("GetModelPortfolio of an off model = %v, want not_found", err)
		}
		srv.Close()
	}
	rec := httptest.NewRecorder()
	modelHandler(t, false).ServeHTTP(rec, authedGet(t, "/v1/model-portfolio"))
	if rec.Code != http.StatusNotFound || bytes.Contains(rec.Body.Bytes(), []byte("FAKE20")) {
		t.Errorf("GET /v1/model-portfolio of an off model = %d %s, want 404 naming nothing", rec.Code, rec.Body)
	}
}

func TestRESTModelPortfolio(t *testing.T) {
	t.Parallel()
	req := authedGet(t, "/v1/model-portfolio?from=2026-03-03")
	rec := httptest.NewRecorder()
	modelHandler(t, true).ServeHTTP(rec, req.Clone(req.Context()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/model-portfolio = %d %s", rec.Code, rec.Body)
	}
	loadContract(t).check(t, req, rec)
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"days":[{"date":"2026-03-03","index":"100"}]`)) {
		t.Errorf("a later from restarts the line at 100: %s", rec.Body)
	}
}

func TestModelEnumsMatch(t *testing.T) {
	t.Parallel()
	for _, k := range []portfolio.TargetKind{portfolio.TargetInstrument, portfolio.TargetIndex} {
		checkEnum(t, viewsv1.ModelTargetKind(0).Descriptor(), "MODEL_TARGET_KIND_", int(k), k.String())
	}
	if int(viewsv1.ModelHistoryGap_MODEL_HISTORY_GAP_NO_CLOSES) != int(portfolio.MissingCloses) ||
		int(viewsv1.ModelHistoryGap_MODEL_HISTORY_GAP_NO_FX_RATE) != int(portfolio.MissingRates) {
		t.Error("ModelHistoryGap and portfolio.MissingData are numbered differently")
	}
}

// authedGet is a GET of target with goodToken.
func authedGet(t *testing.T, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody)
	req.Header.Set("Authorization", "Bearer "+goodToken)
	return req
}
