package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

// modelAPI is a generated ModelPortfolioService client.
type modelAPI struct {
	portfoliov1connect.ModelPortfolioServiceClient
}

// modelClient returns a ModelPortfolioService client of srv that
// presents token as the view plugin's.
func modelClient(srv *httptest.Server, token string) modelAPI {
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
	return modelAPI{portfoliov1connect.NewModelPortfolioServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer))}
}

// ownerCore is a fresh core and the owner's session token on it.
type ownerCore struct {
	srv     *httptest.Server
	session string
}

// ownerSession enrolls a passkey on a fresh core and logs in with it.
func ownerSession(t *testing.T) ownerCore {
	t.Helper()
	dsn := emptyDatabase(t)
	out := enrollCmd(t, dsn)
	link, _, _ := strings.Cut(out, "\n")
	code, ok := strings.CutPrefix(link, localUI+"/enroll#code=")
	if !ok {
		t.Fatalf("enroll-link printed %q", out)
	}
	_, srv, _ := startCoreWith(t, dsn, func(cfg *serveConfig) { cfg.Auth.DevCallerToken = uiCaller })
	c := authClient(srv.URL, srv.Client(), uiCaller)
	return ownerCore{srv: srv, session: passkeyLogin(t, c, passkeyEnroll(t, c, code))}
}

func TestModelPortfolioService(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	srv, session := oc.srv, oc.session
	ctx := t.Context()
	c := modelClient(srv, uiCaller)
	got, err := c.GetModelPortfolio(ctx, connect.NewRequest(&portfoliov1.GetModelPortfolioRequest{Origin: localUI, SessionToken: session}))
	if err != nil || got.Msg.GetModel().GetEnabled() || len(got.Msg.GetModel().GetTargets()) != 0 {
		t.Fatalf("a new core's model = %v, %v; want off and empty (D-52)", got, err)
	}
	checkModelRead(t, oc, http.StatusNotFound, "")
	model := &portfoliov1.ModelPortfolio{Enabled: true, Targets: []*portfoliov1.Target{
		{Kind: portfoliov1.TargetKind_TARGET_KIND_INSTRUMENT, Isin: "IETEST000026", Symbol: "ETF1", WeightPercent: "70"},
		{Kind: portfoliov1.TargetKind_TARGET_KIND_INDEX, Index: "FAKE20", Currency: "PLN", WeightPercent: "30"},
	}}
	if _, saveErr := c.SaveModelPortfolio(ctx, connect.NewRequest(&portfoliov1.SaveModelPortfolioRequest{Origin: localUI, SessionToken: session, Model: model})); saveErr != nil {
		t.Fatalf("SaveModelPortfolio: %v", saveErr)
	}
	got, err = c.GetModelPortfolio(ctx, connect.NewRequest(&portfoliov1.GetModelPortfolioRequest{Origin: localUI, SessionToken: session}))
	if err != nil || !got.Msg.GetModel().GetEnabled() || len(got.Msg.GetModel().GetTargets()) != 2 {
		t.Errorf("the saved model = %v, %v", got, err)
	}
	checkModelRead(t, oc, http.StatusOK, `"FAKE20"`)
	checkModelRefusals(t, oc)
}

// checkModelRefusals checks that the core refuses saving a model to a
// caller that is not a view plugin, an origin off the allow-list and an
// unknown session.
func checkModelRefusals(t *testing.T, oc ownerCore) {
	t.Helper()
	srv, session, ctx := oc.srv, oc.session, t.Context()
	refusals := []struct {
		name, caller, origin, session string
		want                          connect.Code
	}{
		{name: "not a view plugin", caller: "wrong", origin: localUI, session: session, want: connect.CodePermissionDenied},
		{name: "an origin off the allow-list", caller: uiCaller, origin: "https://other.example.com", session: session, want: connect.CodePermissionDenied},
		{name: "an unknown session", caller: uiCaller, origin: localUI, session: "ofsess_nope_nope", want: connect.CodeUnauthenticated},
	}
	for _, r := range refusals {
		_, err := modelClient(srv, r.caller).SaveModelPortfolio(ctx, connect.NewRequest(&portfoliov1.SaveModelPortfolioRequest{
			Origin: r.origin, SessionToken: r.session, Model: &portfoliov1.ModelPortfolio{},
		}))
		if connect.CodeOf(err) != r.want {
			t.Errorf("%s: SaveModelPortfolio = %v, want %s", r.name, err, r.want)
		}
	}
}

// checkModelRead GETs /v1/model-portfolio with the owner's session and
// wants status and a body containing want: 404 while the model is off
// (D-52).
func checkModelRead(t *testing.T, oc ownerCore, status int, want string) {
	t.Helper()
	if r := get(t, oc.srv, "/v1/model-portfolio", oc.session); r.status != status || !strings.Contains(string(r.body), want) {
		t.Errorf("GET /v1/model-portfolio = %d %s, want %d with %s", r.status, r.body, status, want)
	}
}
