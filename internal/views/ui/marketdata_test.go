package ui

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
)

// Errors the fake MarketDataService answers with.
var (
	errNotOwnerFake = errors.New("not the owner")
	errTwiceFake    = errors.New(`invalid source ranking: "a" is ranked twice`)
)

// fakeMarketData is an in-memory MarketDataService with two price
// sources, EODHD's in its default first place. Like the core, it wants the
// owner's session and refuses a source named twice.
type fakeMarketData struct {
	valuationv1connect.UnimplementedMarketDataServiceHandler

	mu     sync.Mutex
	prices []*valuationv1.RankedSource
	set    [][]string
}

func newFakeMarketData() *fakeMarketData {
	return &fakeMarketData{prices: []*valuationv1.RankedSource{
		{Name: "prices-eu", Institution: "eodhd"}, {Name: "ibkr-flex-a", Institution: "ibkr"},
	}}
}

func (f *fakeMarketData) ListSourceRankings(_ context.Context, r *connect.Request[valuationv1.ListSourceRankingsRequest]) (*connect.Response[valuationv1.ListSourceRankingsResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return connect.NewResponse(&valuationv1.ListSourceRankingsResponse{Rankings: []*valuationv1.SourceRanking{
		{Kind: valuationv1.MarketDataKind_MARKET_DATA_KIND_PRICES, Sources: f.prices},
		{Kind: valuationv1.MarketDataKind_MARKET_DATA_KIND_FX_RATES},
		{Kind: valuationv1.MarketDataKind_MARKET_DATA_KIND_INDEX_READINGS},
	}}), nil
}

func (f *fakeMarketData) SetSourceRanking(_ context.Context, r *connect.Request[valuationv1.SetSourceRankingRequest]) (*connect.Response[valuationv1.SetSourceRankingResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	names := r.Msg.GetSources()
	if len(slices.Compact(slices.Sorted(slices.Values(names)))) != len(names) {
		return nil, codeErr(connect.CodeInvalidArgument, errTwiceFake)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.set = append(f.set, names)
	return connect.NewResponse(&valuationv1.SetSourceRankingResponse{}), nil
}

func (f *fakeMarketData) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

func TestSettingsShowsSourceRankings(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `id="market-data-title"`, "Prices", "Exchange rates", "prices-eu", "eodhd", "ibkr-flex-a",
		`name="up" value="ibkr-flex-a"`, "default place", "No connector has delivered any yet.")
}

func TestRankSourcesForm(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantSet    [][]string
	}{
		{name: "move a source up", body: "kind=prices&source=prices-eu&source=ibkr-flex-a&up=ibkr-flex-a", wantStatus: http.StatusSeeOther, wantSet: [][]string{{"ibkr-flex-a", "prices-eu"}}},
		{name: "move a source down", body: "kind=prices&source=prices-eu&source=ibkr-flex-a&down=prices-eu", wantStatus: http.StatusSeeOther, wantSet: [][]string{{"ibkr-flex-a", "prices-eu"}}},
		{name: "the top source up stays", body: "kind=prices&source=prices-eu&source=ibkr-flex-a&up=prices-eu", wantStatus: http.StatusSeeOther, wantSet: [][]string{{"prices-eu", "ibkr-flex-a"}}},
		{name: "back to the defaults", body: "kind=fx_rates&source=nbp-fx&reset=1", wantStatus: http.StatusSeeOther, wantSet: [][]string{nil}},
		{name: "an unknown kind", body: "kind=levels&source=a&up=a", wantStatus: http.StatusBadRequest},
		{name: "no button", body: "kind=prices&source=a", wantStatus: http.StatusBadRequest},
		{name: "the core refuses: shown with its reason", body: "kind=prices&source=a&source=a&up=a", wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			rec := h.serve(t, withSession(form("/settings/market-data", tt.body), ownerSession))
			wantStatus(t, rec, tt.wantStatus)
			var set [][]string
			h.marketData.locked(func() { set = h.marketData.set })
			if len(set) != len(tt.wantSet) || (len(set) == 1 && !slices.Equal(set[0], tt.wantSet[0])) {
				t.Errorf("rankings set = %q, want %q", set, tt.wantSet)
			}
			if strings.Contains(tt.name, "refuses") {
				wantBody(t, rec, "ranked twice")
			}
		})
	}
}
