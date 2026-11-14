package main

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"
	"k8s.io/apimachinery/pkg/util/wait"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

// realCore is the owner's session on a real core and the clients of the
// services the maintainer's checks use, each presenting the view plugin's
// token.
type realCore struct {
	session    string
	maintainer portfoliov1connect.MaintainerServiceClient
	settings   modelsv1connect.ModelSettingsServiceClient
	auth       authAPI
	model      modelAPI
}

// newRealCore starts a core on PostgreSQL and signs the owner in.
func newRealCore(t *testing.T) *realCore {
	t.Helper()
	oc := ownerSession(t)
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+uiCaller)
			return next(ctx, req)
		}
	})
	return &realCore{
		session:    oc.session,
		maintainer: portfoliov1connect.NewMaintainerServiceClient(oc.srv.Client(), oc.srv.URL, connect.WithInterceptors(bearer)),
		settings:   modelsv1connect.NewModelSettingsServiceClient(oc.srv.Client(), oc.srv.URL, connect.WithInterceptors(bearer)),
		auth:       authClient(oc.srv.URL, oc.srv.Client(), uiCaller),
		model:      modelClient(oc.srv, uiCaller),
	}
}

// turnOn saves a model portfolio and turns the switch, the maintainer's toggle
// and its route on.
func (c *realCore) turnOn(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	saved := &portfoliov1.ModelPortfolio{Enabled: true, Targets: []*portfoliov1.Target{
		{Kind: portfoliov1.TargetKind_TARGET_KIND_INSTRUMENT, Isin: "IETEST000026", Symbol: "ETF1", WeightPercent: "70"},
		{Kind: portfoliov1.TargetKind_TARGET_KIND_INDEX, Index: "FAKE20", Currency: "PLN", WeightPercent: "30"},
	}}
	if _, err := c.model.SaveModelPortfolio(ctx, connect.NewRequest(&portfoliov1.SaveModelPortfolioRequest{Origin: localUI, SessionToken: c.session, Model: saved})); err != nil {
		t.Fatalf("SaveModelPortfolio: %v", err)
	}
	if _, err := c.settings.SetAgentsEnabled(ctx, connect.NewRequest(&modelsv1.SetAgentsEnabledRequest{Origin: localUI, SessionToken: c.session, Enabled: true})); err != nil {
		t.Fatalf("SetAgentsEnabled: %v", err)
	}
	if _, err := c.settings.SetFeatureEnabled(ctx, connect.NewRequest(&modelsv1.SetFeatureEnabledRequest{Origin: localUI, SessionToken: c.session, Feature: "maintainer", Enabled: true})); err != nil {
		t.Fatalf("SetFeatureEnabled: %v", err)
	}
	if _, err := c.settings.SetRoute(ctx, connect.NewRequest(&modelsv1.SetRouteRequest{Origin: localUI, SessionToken: c.session, Role: "maintainer", Models: []string{"llm/synthetic"}})); err != nil {
		t.Fatalf("SetRoute: %v", err)
	}
}

// get reads the maintainer's standing as the owner.
func (c *realCore) get(ctx context.Context) (*portfoliov1.GetMaintainerResponse, error) {
	resp, err := c.maintainer.GetMaintainer(ctx, connect.NewRequest(&portfoliov1.GetMaintainerRequest{Origin: localUI, SessionToken: c.session}))
	if err != nil {
		return nil, fmt.Errorf("GetMaintainer: %w", err)
	}
	return resp.Msg, nil
}

// waitForFailure waits until the latest run has ended without a note and returns
// how it ended.
func (c *realCore) waitForFailure(t *testing.T) *portfoliov1.RunFailure {
	t.Helper()
	var (
		last    *portfoliov1.RunFailure
		lastErr error
	)
	err := wait.PollUntilContextTimeout(t.Context(), 20*time.Millisecond, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		got, getErr := c.get(ctx)
		lastErr, last = getErr, got.GetLastFailure()
		return getErr == nil && last != nil && !got.GetRunning(), nil
	})
	if err != nil {
		t.Fatalf("the run never ended: %v (the last read: %v)", err, lastErr)
	}
	return last
}

// actions lists the audit log's actions of the models and the maintainer,
// oldest first.
func (c *realCore) actions(t *testing.T) []identityv1.AuditAction {
	t.Helper()
	resp, err := c.auth.ListAuditEntries(t.Context(), connect.NewRequest(&identityv1.ListAuditEntriesRequest{
		Origin: localUI, SessionToken: c.session, PageSize: 200,
	}))
	if err != nil {
		t.Fatalf("ListAuditEntries: %v", err)
	}
	watched := []identityv1.AuditAction{
		identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_RUN, identityv1.AuditAction_AUDIT_ACTION_MODEL_CALL,
		identityv1.AuditAction_AUDIT_ACTION_MODEL_CALL_RESULT, identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_PROPOSAL_DROPPED,
	}
	var out []identityv1.AuditAction
	for _, e := range slices.Backward(resp.Msg.GetEntries()) {
		if slices.Contains(watched, e.GetAction()) {
			out = append(out, e.GetAction())
		}
	}
	return out
}

// TestMaintainerOffRunsNothingOnARealCore: with the switch off, an ask is
// refused and writes no audit entry, so nothing ran and nothing called a model
// (D-212).
func TestMaintainerOffRunsNothingOnARealCore(t *testing.T) {
	t.Parallel()
	c := newRealCore(t)
	_, err := c.maintainer.AskMaintainer(t.Context(), connect.NewRequest(&portfoliov1.AskMaintainerRequest{Origin: localUI, SessionToken: c.session}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("AskMaintainer with everything off: %v, want FailedPrecondition", err)
	}
	if got := c.actions(t); len(got) != 0 {
		t.Errorf("an ask refused for the switch left audit entries %v", got)
	}
}

// TestMaintainerRunsOnARealCore drives a run through River: with the switch,
// the toggle, a route and the model portfolio on, an ask queues a run that the
// core works, and with nothing valued yet the run ends without a note and
// without asking any model, as one audited run with its class.
func TestMaintainerRunsOnARealCore(t *testing.T) {
	t.Parallel()
	c := newRealCore(t)
	ctx := t.Context()
	c.turnOn(t)
	got, err := c.get(ctx)
	if err != nil || !got.GetAvailable() || got.GetNote() != nil {
		t.Fatalf("GetMaintainer = %v, %v; want available with no note", got, err)
	}
	started, err := c.maintainer.AskMaintainer(ctx, connect.NewRequest(&portfoliov1.AskMaintainerRequest{Origin: localUI, SessionToken: c.session}))
	if err != nil || !started.Msg.GetStarted() {
		t.Fatalf("AskMaintainer = %v, %v; want a run queued", started, err)
	}
	failure := c.waitForFailure(t)
	if failure.GetErrorClass() != "no_figures" || failure.GetTrigger() != portfoliov1.RunTrigger_RUN_TRIGGER_OWNER {
		t.Errorf("the run ended as %v, want no_figures asked by the owner", failure)
	}
	want := []identityv1.AuditAction{identityv1.AuditAction_AUDIT_ACTION_MAINTAINER_RUN}
	if got := c.actions(t); !slices.Equal(got, want) {
		t.Errorf("audit actions = %v, want one maintainer_run and no model call: no model was asked", got)
	}
}
