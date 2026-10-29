package connectapi_test

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/testing/protocmp"

	gocmp "github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/portfolio/connectapi"
	"github.com/arkadiuss/openfief/internal/valuation"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

// The synthetic caller token, sessions and origin.
const (
	viewCaller   = "view-plugin-token"
	ownerSession = "ofsess_owner"
	readSession  = "ofsess_read"
	local        = "http://localhost:8080"
)

var (
	errDown    = errors.New("store down")
	errNotView = errors.New("not a view plugin")
)

// memStore keeps the model in memory, or fails with err.
type memStore struct {
	model portfolio.Model
	err   error
}

func (m *memStore) Model(context.Context) (portfolio.Model, error) { return m.model, m.err }

func (m *memStore) SaveModel(_ context.Context, model portfolio.Model) error {
	m.model = model
	return m.err
}

// noMarket has no history; these tests never backtest.
type noMarket struct{}

func (noMarket) Prices(context.Context, string, kernel.Date) ([]valuation.Price, error) {
	return nil, nil
}

func (noMarket) IndexLevels(context.Context, string, kernel.Date) ([]valuation.IndexLevel, error) {
	return nil, nil
}

func (noMarket) FXRates(context.Context, kernel.Currency, kernel.Date) ([]valuation.FXRate, error) {
	return nil, nil
}

// owners accepts ownerSession, refuses readSession as not the owner's and
// any other token as no session.
type owners struct{}

func (owners) CheckOwner(_ context.Context, _, token string) error {
	switch token {
	case ownerSession:
		return nil
	case readSession:
		return app.ErrNotOwner
	}
	return app.ErrNoSession
}

// viewOnly admits calls bearing viewCaller, like cmd/core's view-plugin
// check.
func viewOnly() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if req.Header().Get("Authorization") != "Bearer "+viewCaller {
				return nil, connect.NewError(connect.CodePermissionDenied, errNotView)
			}
			return next(ctx, req)
		}
	}
}

// bearer sets the Authorization header of every call.
func bearer(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// newClient serves the service over store and returns a client calling
// it with the caller token.
// client is the generated client, as a struct so helpers return no
// interface.
type client struct {
	portfoliov1connect.ModelPortfolioServiceClient
}

func newClient(t *testing.T, store *memStore, caller string) client {
	t.Helper()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	models, err := app.New(store, noMarket{}, owners{}, pln)
	if err != nil {
		t.Fatal(err)
	}
	_, handler := connectapi.New(models, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return client{portfoliov1connect.NewModelPortfolioServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(caller)))}
}

// sample is a valid model with an instrument and an index.
func sample() *portfoliov1.ModelPortfolio {
	return &portfoliov1.ModelPortfolio{Enabled: true, Targets: []*portfoliov1.Target{
		{Kind: portfoliov1.TargetKind_TARGET_KIND_INSTRUMENT, Isin: "IETEST000026", Symbol: "ETF1", WeightPercent: "60.5"},
		{Kind: portfoliov1.TargetKind_TARGET_KIND_INDEX, Index: "FAKE20", Currency: "PLN", WeightPercent: "39.5"},
	}}
}

func TestSaveAndGet(t *testing.T) {
	t.Parallel()
	c := newClient(t, &memStore{}, viewCaller)
	ctx := t.Context()
	got, err := c.GetModelPortfolio(ctx, connect.NewRequest(&portfoliov1.GetModelPortfolioRequest{Origin: local, SessionToken: ownerSession}))
	if err != nil || got.Msg.GetModel().GetEnabled() || len(got.Msg.GetModel().GetTargets()) != 0 {
		t.Fatalf("GetModelPortfolio before a save = %v, %v; want off and empty", got, err)
	}
	saved, err := c.SaveModelPortfolio(ctx, connect.NewRequest(&portfoliov1.SaveModelPortfolioRequest{Origin: local, SessionToken: ownerSession, Model: sample()}))
	if err != nil {
		t.Fatalf("SaveModelPortfolio: %v", err)
	}
	if diff := gocmp.Diff(sample(), saved.Msg.GetModel(), protocmp.Transform()); diff != "" {
		t.Errorf("saved (-want +got):\n%s", diff)
	}
	got, err = c.GetModelPortfolio(ctx, connect.NewRequest(&portfoliov1.GetModelPortfolioRequest{Origin: local, SessionToken: ownerSession}))
	if err != nil {
		t.Fatalf("GetModelPortfolio: %v", err)
	}
	if diff := gocmp.Diff(sample(), got.Msg.GetModel(), protocmp.Transform()); diff != "" {
		t.Errorf("read back (-want +got):\n%s", diff)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	edit := func(f func(m *portfoliov1.ModelPortfolio)) *portfoliov1.ModelPortfolio {
		m := sample()
		f(m)
		return m
	}
	tests := []struct {
		name, caller, session string
		store                 *memStore
		model                 *portfoliov1.ModelPortfolio
		want                  connect.Code
	}{
		{name: "not a view plugin", caller: "connector-token", session: ownerSession, want: connect.CodePermissionDenied},
		{name: "no session", caller: viewCaller, session: "ofsess_gone", want: connect.CodeUnauthenticated},
		{name: "a read session", caller: viewCaller, session: readSession, want: connect.CodePermissionDenied},
		{
			name: "weights not summing to 100", caller: viewCaller, session: ownerSession, want: connect.CodeInvalidArgument,
			model: edit(func(m *portfoliov1.ModelPortfolio) { m.Targets[0].WeightPercent = "50" }),
		},
		{
			name: "a malformed weight", caller: viewCaller, session: ownerSession, want: connect.CodeInvalidArgument,
			model: edit(func(m *portfoliov1.ModelPortfolio) { m.Targets[0].WeightPercent = "60,5" }),
		},
		{
			name: "an index without a currency", caller: viewCaller, session: ownerSession, want: connect.CodeInvalidArgument,
			model: edit(func(m *portfoliov1.ModelPortfolio) { m.Targets[1].Currency = "" }),
		},
		{
			name: "no kind", caller: viewCaller, session: ownerSession, want: connect.CodeInvalidArgument,
			model: edit(func(m *portfoliov1.ModelPortfolio) { m.Targets[1].Kind = 0 }),
		},
		{
			name: "a bad ISIN", caller: viewCaller, session: ownerSession, want: connect.CodeInvalidArgument,
			model: edit(func(m *portfoliov1.ModelPortfolio) { m.Targets[0].Isin = "IETEST000027" }),
		},
		{name: "the store failing", caller: viewCaller, session: ownerSession, store: &memStore{err: errDown}, want: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store, model := cmp.Or(tt.store, &memStore{}), cmp.Or(tt.model, sample())
			c := newClient(t, store, tt.caller)
			_, err := c.SaveModelPortfolio(t.Context(), connect.NewRequest(&portfoliov1.SaveModelPortfolioRequest{Origin: local, SessionToken: tt.session, Model: model}))
			if connect.CodeOf(err) != tt.want {
				t.Errorf("SaveModelPortfolio = %v, want %s", err, tt.want)
			}
			if tt.model != nil || tt.store != nil {
				return
			}
			_, err = c.GetModelPortfolio(t.Context(), connect.NewRequest(&portfoliov1.GetModelPortfolioRequest{Origin: local, SessionToken: tt.session}))
			if connect.CodeOf(err) != tt.want {
				t.Errorf("GetModelPortfolio = %v, want %s", err, tt.want)
			}
		})
	}
}
