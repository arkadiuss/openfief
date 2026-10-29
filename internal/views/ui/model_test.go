package ui

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errWeights is the fake core's reason for refusing a model whose weights
// don't sum to 100.
var errWeights = errors.New("the weights do not sum to 100%")

// errModelOff is the read API's answer while the model portfolio is off.
var errModelOff = errors.New("no model portfolio")

// fakeModel is an in-memory ModelPortfolioService. Like the core, it wants
// the plugin's token as the caller and an owner session in the request,
// and refuses weights that don't sum to 100.
type fakeModel struct {
	portfoliov1connect.UnimplementedModelPortfolioServiceHandler

	mu      sync.Mutex
	model   *portfoliov1.ModelPortfolio
	bearers []string
}

// newFakeModel is the fake with a model never saved: off and empty.
func newFakeModel() *fakeModel { return &fakeModel{model: &portfoliov1.ModelPortfolio{}} }

func (f *fakeModel) GetModelPortfolio(_ context.Context, r *connect.Request[portfoliov1.GetModelPortfolioRequest]) (*connect.Response[portfoliov1.GetModelPortfolioResponse], error) {
	if err := f.check(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return connect.NewResponse(&portfoliov1.GetModelPortfolioResponse{Model: f.model}), nil
}

func (f *fakeModel) SaveModelPortfolio(_ context.Context, r *connect.Request[portfoliov1.SaveModelPortfolioRequest]) (*connect.Response[portfoliov1.SaveModelPortfolioResponse], error) {
	if err := f.check(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	sum := 0
	for _, t := range r.Msg.GetModel().GetTargets() {
		w, err := strconv.Atoi(t.GetWeightPercent()) // whole percents are enough here
		if err != nil {
			return nil, codeErr(connect.CodeInvalidArgument, err)
		}
		sum += w
	}
	if sum != 100 && len(r.Msg.GetModel().GetTargets()) > 0 {
		return nil, codeErr(connect.CodeInvalidArgument, errWeights)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.model = r.Msg.GetModel()
	return connect.NewResponse(&portfoliov1.SaveModelPortfolioResponse{Model: f.model}), nil
}

// check notes the caller and checks it and the session.
func (f *fakeModel) check(h http.Header, session string) error {
	f.mu.Lock()
	f.bearers = append(f.bearers, h.Get("Authorization"))
	f.mu.Unlock()
	if h.Get("Authorization") != "Bearer "+pluginToken {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	return owner(session)
}

// locked runs fn with the fake's lock held, for reading what it recorded.
func (f *fakeModel) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// stored returns the stored model.
func (f *fakeModel) stored() *portfoliov1.ModelPortfolio {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.model
}

func (f *fakeRead) GetPerformance(_ context.Context, r *connect.Request[viewsv1.GetPerformanceRequest]) (*connect.Response[viewsv1.GetPerformanceResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.GetPerformanceResponse{
		From: "2026-03-30", To: r.Msg.GetTo(), ReportingCurrency: "PLN", RevaluationPending: true,
		Days: []*viewsv1.PerformanceDay{
			{Date: "2026-03-30", Index: "100", ValueReporting: pln("1490.00")},
			{Date: "2026-03-31", Index: "101.5", ValueReporting: pln("1512.34")},
		},
	}), nil
}

func (f *fakeRead) GetModelPortfolio(_ context.Context, r *connect.Request[viewsv1.GetModelPortfolioRequest]) (*connect.Response[viewsv1.GetModelPortfolioResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	var on bool
	f.locked(func() { on = f.modelOn })
	if !on {
		return nil, codeErr(connect.CodeNotFound, errModelOff)
	}
	etf := &viewsv1.ModelTarget{Kind: viewsv1.ModelTargetKind_MODEL_TARGET_KIND_INSTRUMENT, Instrument: syntheticInstrument(), WeightPercent: "60"}
	idx := &viewsv1.ModelTarget{Kind: viewsv1.ModelTargetKind_MODEL_TARGET_KIND_INDEX, Index: "FAKE20", Currency: "PLN", WeightPercent: "40"}
	return connect.NewResponse(&viewsv1.GetModelPortfolioResponse{
		From: "2026-03-30", To: r.Msg.GetTo(), ReportingCurrency: "PLN", Targets: []*viewsv1.ModelTarget{etf, idx},
		BacktestStart: "2024-01-02", BacktestEnd: "2026-03-31",
		Days:     []*viewsv1.BacktestDay{{Date: "2026-03-30", Index: "100"}, {Date: "2026-03-31", Index: "99.25"}},
		Drawdown: &viewsv1.Drawdown{PeakDate: "2025-02-03", TroughDate: "2025-04-08", PeakValue: "130", TroughValue: "104", Depth: "0.2"},
		TargetHistories: []*viewsv1.ModelTargetHistory{
			{Target: etf, FirstDate: "2024-01-02", LastDate: "2026-03-31", Drawdown: &viewsv1.Drawdown{
				PeakDate: "2024-01-02", TroughDate: "2024-01-02", PeakValue: "50", TroughValue: "50", Depth: "0",
			}},
			{Target: idx, Gap: viewsv1.ModelHistoryGap_MODEL_HISTORY_GAP_NO_CLOSES},
		},
	}), nil
}

// TestOverviewModel checks the overview while the model portfolio is on:
// its backtest as the chart's second line on the portfolio's days, and the
// max drawdowns of the model and each target (D-53, D-54).
func TestOverviewModel(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.read.locked(func() { h.read.modelOn = true })
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`&#34;series&#34;:[{&#34;label&#34;:&#34;Model portfolio (backtest)&#34;,&#34;values&#34;:[&#34;100&#34;,&#34;99.25&#34;]}]`,
		"Model portfolio", "from 2024-01-02 to 2026-03-31", "20%", "2025-02-03", "2025-04-08",
		"SYN (XX0000000001)", "never fell", "FAKE20 (index, PLN)", "No history: no stored prices or index levels",
	)
}

// TestSettingsModel checks the model portfolio section of the settings
// page: the stored model with blank rows, a save that reaches the core with
// the plugin's token and the owner's session, and a refusal shown with the
// core's reason and the rows as posted.
func TestSettingsModel(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `id="model-title"`, `action="/settings/model"`, `name="enabled">`)
	if n := strings.Count(rec.Body.String(), `name="weight"`); n != blankTargetRows {
		t.Errorf("an empty model shows %d rows, want %d blank", n, blankTargetRows)
	}

	body := "enabled=on&kind=instrument&id=XX0000000001%2FSYN&currency=&weight=60&kind=index&id=FAKE20&currency=pln&weight=40" +
		"&kind=instrument&id=&currency=&weight="
	rec = h.serve(t, withSession(form("/settings/model", body), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	want := &portfoliov1.ModelPortfolio{Enabled: true, Targets: []*portfoliov1.Target{
		{Kind: portfoliov1.TargetKind_TARGET_KIND_INSTRUMENT, Isin: "XX0000000001", Symbol: "SYN", WeightPercent: "60"},
		{Kind: portfoliov1.TargetKind_TARGET_KIND_INDEX, Index: "FAKE20", Currency: "PLN", WeightPercent: "40"},
	}}
	if diff := cmp.Diff(want, h.model.stored(), protocmp.Transform()); diff != "" {
		t.Errorf("saved model (-want +got):\n%s", diff)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantBody(t, rec, `name="enabled" checked>`, `value="XX0000000001/SYN"`, `value="FAKE20"`, `<option value="index" selected>`)

	rec = h.serve(t, withSession(form("/settings/model", "kind=instrument&id=XX0000000001%2FSYN&currency=&weight=30"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The model portfolio was not saved", "the weights do not sum to 100%", `value="30"`)
	if bearers := seen(h.model, func() []string { return h.model.bearers }); len(bearers) == 0 || bearers[0] != "Bearer "+pluginToken {
		t.Errorf("ModelPortfolioService callers = %v, want the plugin's token", bearers)
	}
}

// TestSettingsModelRefusals checks that a read session can't reach the
// model's form, and that a malformed form never reaches the core.
func TestSettingsModelRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/settings/model", "enabled=on"), readSession))
	wantStatus(t, rec, http.StatusForbidden)
	rec = h.serve(t, withSession(form("/settings/model", "kind=index&id=FAKE20&weight=100"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The model portfolio form is malformed.")
	if len(h.model.stored().GetTargets()) != 0 {
		t.Errorf("a refused form stored %v", h.model.stored())
	}
}
