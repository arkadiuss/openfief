package connectapi_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	"github.com/arkadiuss/openfief/internal/modelrouter/connectapi"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// memStore is an in-memory app.SettingsStore.
type memStore struct {
	mu sync.Mutex
	s  modelrouter.Settings
}

func (m *memStore) Settings(context.Context) (modelrouter.Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.s, nil
}

func (m *memStore) SaveEnabled(ctx context.Context, on bool, within func(context.Context) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := runWithin(ctx, within); err != nil {
		return err
	}
	m.s.Switch = m.s.Switch.WithEnabled(on)
	return nil
}

func (m *memStore) SaveToggle(ctx context.Context, f modelrouter.Feature, on bool, within func(context.Context) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := runWithin(ctx, within); err != nil {
		return err
	}
	sw, err := m.s.Switch.WithToggle(f, on)
	if err != nil {
		return fmt.Errorf("toggling a feature: %w", err)
	}
	m.s.Switch = sw
	return nil
}

func (m *memStore) SaveRoute(ctx context.Context, role modelrouter.Role, refs []modelrouter.Ref, within func(context.Context) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := runWithin(ctx, within); err != nil {
		return err
	}
	routes := map[modelrouter.Role]modelrouter.Route{role: {Role: role, Models: refs}}
	for r, v := range m.s.Routes {
		if r != role {
			routes[r] = v
		}
	}
	m.s.Routes = routes
	return nil
}

func (m *memStore) SaveDataTierModels(ctx context.Context, tier modelrouter.DataTier, refs []modelrouter.Ref, within func(context.Context) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := runWithin(ctx, within); err != nil {
		return err
	}
	m.s.Policy = modelrouter.DataTierPolicy{tier: refs}
	return nil
}

// runWithin runs the hook a store runs inside its transaction, when there is
// one.
func runWithin(ctx context.Context, within func(context.Context) error) error {
	if within == nil {
		return nil
	}
	return within(ctx)
}

// owners admits the session token "owner" only.
type owners struct{}

func (o owners) OwnerSession(ctx context.Context, origin, token string) (string, error) {
	if err := o.CheckOwner(ctx, origin, token); err != nil {
		return "", err
	}
	return "session-owner", nil
}

func (owners) CheckOwner(_ context.Context, _, token string) error {
	switch token {
	case "owner":
		return nil
	case "read":
		return app.ErrNotOwner
	}
	return app.ErrNoSession
}

// passThrough is a caller interceptor that admits everyone: the caller check
// is the interceptor's, tested where it lives.
type passThrough struct{}

func (passThrough) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc { return next }
func (passThrough) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (passThrough) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// settingsClient is the generated client of the ModelSettingsService.
type settingsClient struct {
	modelsv1connect.ModelSettingsServiceClient
}

func newClient(t *testing.T) (*settingsClient, *memStore) {
	t.Helper()
	store := &memStore{}
	path, h := connectapi.NewSettings(app.NewSettings(store, owners{}), passThrough{}, slog.New(slog.DiscardHandler))
	mux := http.NewServeMux()
	mux.Handle(path, h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &settingsClient{modelsv1connect.NewModelSettingsServiceClient(srv.Client(), srv.URL)}, store
}

func TestGetModelSettingsDefaultsToOff(t *testing.T) {
	t.Parallel()
	c, _ := newClient(t)
	resp, err := c.GetModelSettings(t.Context(), connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: "o", SessionToken: "owner"}))
	if err != nil {
		t.Fatal(err)
	}
	want := &modelsv1.GetModelSettingsResponse{
		Features: []*modelsv1.FeatureToggle{{Feature: "ask"}, {Feature: "maintainer"}, {Feature: "mcp"}},
		Routes:   []*modelsv1.Route{{Role: "ask"}, {Role: "maintainer"}},
		Tiers:    []*modelsv1.TierModels{{Tier: "T0"}, {Tier: "TF"}, {Tier: "T1"}, {Tier: "T2"}},
	}
	if diff := cmp.Diff(want, resp.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("settings (-want +got):\n%s", diff)
	}
}

func TestOwnerChangesTheSwitchAndTheToggles(t *testing.T) {
	t.Parallel()
	c, store := newClient(t)
	ctx := t.Context()
	if _, err := c.SetAgentsEnabled(ctx, connect.NewRequest(&modelsv1.SetAgentsEnabledRequest{Origin: "o", SessionToken: "owner", Enabled: true})); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SetFeatureEnabled(ctx, connect.NewRequest(&modelsv1.SetFeatureEnabledRequest{Origin: "o", SessionToken: "owner", Feature: "ask", Enabled: true})); err != nil {
		t.Fatal(err)
	}
	if !store.s.Switch.Allows(modelrouter.FeatureAsk) || store.s.Switch.Allows(modelrouter.FeatureMCP) {
		t.Errorf("switch = %+v", store.s.Switch)
	}
	resp, err := c.GetModelSettings(ctx, connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: "o", SessionToken: "owner"}))
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Msg.GetAgentsEnabled() || !resp.Msg.GetFeatures()[0].GetEnabled() || resp.Msg.GetFeatures()[2].GetEnabled() {
		t.Errorf("response = %v", resp.Msg)
	}
}

func TestRoutesAndTiersRoundTrip(t *testing.T) {
	t.Parallel()
	c, _ := newClient(t)
	ctx := t.Context()
	route := []string{"openrouter/openai/gpt-oss-120b", "ollama/qwen3:8b"}
	if _, err := c.SetRoute(ctx, connect.NewRequest(&modelsv1.SetRouteRequest{Origin: "o", SessionToken: "owner", Role: "ask", Models: route})); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SetTierModels(ctx, connect.NewRequest(&modelsv1.SetTierModelsRequest{Origin: "o", SessionToken: "owner", Tier: "T2", Models: route[:1]})); err != nil {
		t.Fatal(err)
	}
	resp, err := c.GetModelSettings(ctx, connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: "o", SessionToken: "owner"}))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(route, resp.Msg.GetRoutes()[0].GetModels(), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("route (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(route[:1], resp.Msg.GetTiers()[3].GetModels(), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("T2's models (-want +got):\n%s", diff)
	}
}

func TestOnlyTheOwnerMayCall(t *testing.T) {
	t.Parallel()
	calls := map[string]func(c *settingsClient, token string) connect.Code{
		"get": func(c *settingsClient, token string) connect.Code {
			_, err := c.GetModelSettings(t.Context(), connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: "o", SessionToken: token}))
			return connect.CodeOf(err)
		},
		"switch": func(c *settingsClient, token string) connect.Code {
			_, err := c.SetAgentsEnabled(t.Context(), connect.NewRequest(&modelsv1.SetAgentsEnabledRequest{Origin: "o", SessionToken: token, Enabled: true}))
			return connect.CodeOf(err)
		},
		"toggle": func(c *settingsClient, token string) connect.Code {
			_, err := c.SetFeatureEnabled(t.Context(), connect.NewRequest(&modelsv1.SetFeatureEnabledRequest{Origin: "o", SessionToken: token, Feature: "ask", Enabled: true}))
			return connect.CodeOf(err)
		},
		"route": func(c *settingsClient, token string) connect.Code {
			_, err := c.SetRoute(t.Context(), connect.NewRequest(&modelsv1.SetRouteRequest{Origin: "o", SessionToken: token, Role: "ask", Models: []string{"p/m"}}))
			return connect.CodeOf(err)
		},
		"tier": func(c *settingsClient, token string) connect.Code {
			_, err := c.SetTierModels(t.Context(), connect.NewRequest(&modelsv1.SetTierModelsRequest{Origin: "o", SessionToken: token, Tier: "T2", Models: []string{"p/m"}}))
			return connect.CodeOf(err)
		},
	}
	for name, call := range calls {
		for token, want := range map[string]connect.Code{"": connect.CodeUnauthenticated, "read": connect.CodePermissionDenied} {
			t.Run(name+" with "+token, func(t *testing.T) {
				t.Parallel()
				c, store := newClient(t)
				if got := call(c, token); got != want {
					t.Fatalf("code = %v, want %v", got, want)
				}
				if diff := cmp.Diff(modelrouter.Settings{}, store.s, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("a refused call changed the settings (-want +got):\n%s", diff)
				}
			})
		}
	}
}

func TestInvalidInputIsRefused(t *testing.T) {
	t.Parallel()
	c, store := newClient(t)
	ctx := t.Context()
	codes := []connect.Code{
		func() connect.Code {
			_, err := c.SetFeatureEnabled(ctx, connect.NewRequest(&modelsv1.SetFeatureEnabledRequest{Origin: "o", SessionToken: "owner", Feature: "trade", Enabled: true}))
			return connect.CodeOf(err)
		}(),
		func() connect.Code {
			_, err := c.SetRoute(ctx, connect.NewRequest(&modelsv1.SetRouteRequest{Origin: "o", SessionToken: "owner", Role: "ask", Models: []string{"no-slash"}}))
			return connect.CodeOf(err)
		}(),
		func() connect.Code {
			_, err := c.SetTierModels(ctx, connect.NewRequest(&modelsv1.SetTierModelsRequest{Origin: "o", SessionToken: "owner", Tier: "T9", Models: []string{"p/m"}}))
			return connect.CodeOf(err)
		}(),
	}
	for i, code := range codes {
		if code != connect.CodeInvalidArgument {
			t.Errorf("call %d: code = %v, want InvalidArgument", i, code)
		}
	}
	if diff := cmp.Diff(modelrouter.Settings{}, store.s, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("settings changed (-want +got):\n%s", diff)
	}
}
