package ui

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"

	"connectrpc.com/connect"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

var (
	errBadRef = errors.New("a model is named provider/model")
	errDupRef = errors.New("a model may be on a route once")
)

// fakeModelSettings is an in-memory ModelSettingsService: it wants the
// plugin's token and the owner's session, keeps what the owner sets, and
// refuses what the core would (a reference that is not provider/model, an
// unknown role), recording each call.
type fakeModelSettings struct {
	modelsv1connect.UnimplementedModelSettingsServiceHandler

	mu       sync.Mutex
	enabled  bool
	toggles  map[string]bool
	routes   map[string][]string
	tiers    map[string][]string
	broken   []*modelsv1.BrokenEntry
	provider []*modelsv1.ModelProvider
	// refuse, when set, is the reason every change fails with FAILED_PRECONDITION.
	refuse error
	// calls lists the changes made, as "switch on", "toggle ask off", "route ask p/m,q/n" and "tier T0 p/m".
	calls []string
	// unavailable makes every call fail as a core that is down does.
	unavailable bool
}

// newFakeModelSettings is a core with a cloud and a local provider, a route of
// each role, one entry of which is broken, and nothing allowed or on.
func newFakeModelSettings() *fakeModelSettings {
	return &fakeModelSettings{
		toggles: map[string]bool{},
		routes:  map[string][]string{"ask": {"cloud/openai/gpt-x", "gone/m"}, "maintainer": {"home/qwen3:8b"}},
		tiers:   map[string][]string{"T2": {"cloud/openai/gpt-x"}},
		broken:  []*modelsv1.BrokenEntry{{Role: "ask", Model: "gone/m", Reason: "missing"}},
		provider: []*modelsv1.ModelProvider{
			{
				Instance: "cloud", Available: true, Location: modelsv1.Location_LOCATION_THIRD_PARTY_CLOUD,
				Retention: &modelsv1.Retention{Days: 30},
				Models: []*modelsv1.ModelEntry{{
					Id: "openai/gpt-x", InputPrice: "0.15", OutputPrice: "0.6", PriceCurrency: "USD",
					ContextTokens: 128000, MaxOutputTokens: 8000, ToolCalling: true, StructuredOutput: true,
				}},
			},
			{
				Instance: "home", Available: true, Location: modelsv1.Location_LOCATION_OWNER_HARDWARE, Retention: &modelsv1.Retention{},
				Models: []*modelsv1.ModelEntry{{Id: "qwen3:8b", ContextTokens: 32000, ToolCalling: true}},
			},
			{Instance: "sleepy"},
		},
	}
}

func (f *fakeModelSettings) GetModelSettings(_ context.Context, r *connect.Request[modelsv1.GetModelSettingsRequest]) (*connect.Response[modelsv1.GetModelSettingsResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := &modelsv1.GetModelSettingsResponse{AgentsEnabled: f.enabled, Broken: f.broken}
	for _, feature := range []string{"ask", "maintainer", "mcp"} {
		out.Features = append(out.Features, &modelsv1.FeatureToggle{Feature: feature, Enabled: f.toggles[feature]})
	}
	for _, role := range []string{"ask", "maintainer"} {
		out.Routes = append(out.Routes, &modelsv1.Route{Role: role, Models: f.routes[role]})
	}
	for _, tier := range []string{"T0", "TF", "T1", "T2"} {
		out.Tiers = append(out.Tiers, &modelsv1.TierModels{Tier: tier, Models: f.tiers[tier]})
	}
	return connect.NewResponse(out), nil
}

func (f *fakeModelSettings) ListModelProviders(_ context.Context, r *connect.Request[modelsv1.ListModelProvidersRequest]) (*connect.Response[modelsv1.ListModelProvidersResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return connect.NewResponse(&modelsv1.ListModelProvidersResponse{Providers: f.provider}), nil
}

func (f *fakeModelSettings) SetAgentsEnabled(_ context.Context, r *connect.Request[modelsv1.SetAgentsEnabledRequest]) (*connect.Response[modelsv1.SetAgentsEnabledResponse], error) {
	if err := f.change(r.Header(), r.Msg.GetSessionToken(), "switch "+onWord(r.Msg.GetEnabled())); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enabled = r.Msg.GetEnabled()
	return connect.NewResponse(&modelsv1.SetAgentsEnabledResponse{}), nil
}

func (f *fakeModelSettings) SetFeatureEnabled(_ context.Context, r *connect.Request[modelsv1.SetFeatureEnabledRequest]) (*connect.Response[modelsv1.SetFeatureEnabledResponse], error) {
	if err := f.change(r.Header(), r.Msg.GetSessionToken(), "toggle "+r.Msg.GetFeature()+" "+onWord(r.Msg.GetEnabled())); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.toggles[r.Msg.GetFeature()] = r.Msg.GetEnabled()
	return connect.NewResponse(&modelsv1.SetFeatureEnabledResponse{}), nil
}

func (f *fakeModelSettings) SetRoute(_ context.Context, r *connect.Request[modelsv1.SetRouteRequest]) (*connect.Response[modelsv1.SetRouteResponse], error) {
	if err := f.change(r.Header(), r.Msg.GetSessionToken(), "route "+r.Msg.GetRole()+" "+joinRefs(r.Msg.GetModels())); err != nil {
		return nil, err
	}
	for _, m := range r.Msg.GetModels() {
		if !strings.Contains(m, "/") {
			return nil, codeErr(connect.CodeInvalidArgument, errBadRef)
		}
	}
	if seen := map[string]bool{}; slices.ContainsFunc(r.Msg.GetModels(), func(m string) bool { dup := seen[m]; seen[m] = true; return dup }) {
		return nil, codeErr(connect.CodeInvalidArgument, errDupRef)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routes[r.Msg.GetRole()] = r.Msg.GetModels()
	return connect.NewResponse(&modelsv1.SetRouteResponse{}), nil
}

func (f *fakeModelSettings) SetTierModels(_ context.Context, r *connect.Request[modelsv1.SetTierModelsRequest]) (*connect.Response[modelsv1.SetTierModelsResponse], error) {
	if err := f.change(r.Header(), r.Msg.GetSessionToken(), "tier "+r.Msg.GetTier()+" "+joinRefs(r.Msg.GetModels())); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tiers[r.Msg.GetTier()] = r.Msg.GetModels()
	return connect.NewResponse(&modelsv1.SetTierModelsResponse{}), nil
}

// admit checks the plugin's token and that the session is the owner's.
func (f *fakeModelSettings) admit(h http.Header, token string) error {
	if h.Get("Authorization") != "Bearer "+pluginToken {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	if _, err := scopeOf(token); err != nil {
		return err
	}
	if err := owner(token); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unavailable {
		return codeErr(connect.CodeUnavailable, nil)
	}
	return nil
}

// change admits a change, records it, and fails it when the core refuses.
func (f *fakeModelSettings) change(h http.Header, token, what string) error {
	if err := f.admit(h, token); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, what)
	if f.refuse != nil {
		return codeErr(connect.CodeFailedPrecondition, f.refuse)
	}
	return nil
}

func (f *fakeModelSettings) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// onWord spells a switch's state as the fake records it.
func onWord(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

// joinRefs lists references with commas, in order.
func joinRefs(refs []string) string { return strings.Join(refs, ",") }
