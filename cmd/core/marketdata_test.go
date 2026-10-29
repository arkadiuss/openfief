package main

import (
	"context"
	"net/http/httptest"
	"slices"
	"testing"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
)

// marketDataAPI is a generated MarketDataService client.
type marketDataAPI struct {
	valuationv1connect.MarketDataServiceClient
}

// marketDataClient returns a MarketDataService client of srv that presents
// token as the view plugin's.
func marketDataClient(srv *httptest.Server, token string) marketDataAPI {
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
	return marketDataAPI{valuationv1connect.NewMarketDataServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer))}
}

// TestMarketDataService checks that the core serves the owner's ranking of
// market-data sources to view plugins (D-94, D-79): one ranking per kind,
// the owner's order stored, a source ranked twice and other callers
// refused.
func TestMarketDataService(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	ctx := t.Context()
	c := marketDataClient(oc.srv, uiCaller)
	list, err := c.ListSourceRankings(ctx, connect.NewRequest(&valuationv1.ListSourceRankingsRequest{Origin: localUI, SessionToken: oc.session}))
	if err != nil || len(list.Msg.GetRankings()) != 3 {
		t.Fatalf("ListSourceRankings = %v, %v; want one ranking per kind", list, err)
	}
	set, err := c.SetSourceRanking(ctx, connect.NewRequest(&valuationv1.SetSourceRankingRequest{
		Origin: localUI, SessionToken: oc.session, Kind: valuationv1.MarketDataKind_MARKET_DATA_KIND_FX_RATES, Sources: []string{"ecb-fx", "nbp-fx"},
	}))
	if err != nil {
		t.Fatalf("SetSourceRanking: %v", err)
	}
	names := make([]string, 0, len(set.Msg.GetRanking().GetSources()))
	for _, s := range set.Msg.GetRanking().GetSources() {
		names = append(names, s.GetName())
	}
	if !slices.Equal(names, []string{"ecb-fx", "nbp-fx"}) {
		t.Errorf("the stored FX ranking = %v", names)
	}
	refusals := []struct {
		name, caller, session string
		sources               []string
		want                  connect.Code
	}{
		{name: "a source twice", caller: uiCaller, session: oc.session, sources: []string{"a", "a"}, want: connect.CodeInvalidArgument},
		{name: "not a view plugin", caller: "wrong", session: oc.session, want: connect.CodePermissionDenied},
		{name: "an unknown session", caller: uiCaller, session: "ofsess_nope_nope", want: connect.CodeUnauthenticated},
	}
	for _, r := range refusals {
		_, err := marketDataClient(oc.srv, r.caller).SetSourceRanking(ctx, connect.NewRequest(&valuationv1.SetSourceRankingRequest{
			Origin: localUI, SessionToken: r.session, Kind: valuationv1.MarketDataKind_MARKET_DATA_KIND_PRICES, Sources: r.sources,
		}))
		if connect.CodeOf(err) != r.want {
			t.Errorf("%s: SetSourceRanking = %v, want %s", r.name, err, r.want)
		}
	}
}
