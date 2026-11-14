package connectapi_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	"github.com/arkadiuss/openfief/internal/modelrouter/connectapi"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// instances is a fake of the plugins: the cloud instance answers, the local
// one is down, the connector is no model provider.
type instances struct{}

func (instances) Instance(_ context.Context, name string) error {
	switch name {
	case "gone":
		return app.ErrProviderMissing
	case "connector":
		return app.ErrNotModelProvider
	}
	return nil
}

func (instances) Describe(_ context.Context, name string) (modelrouter.ProviderInstance, error) {
	switch name {
	case "cloud":
		return modelrouter.ProviderInstance{Location: modelrouter.LocationThirdPartyCloud, RetentionDays: 30, MayTrain: false}, nil
	case "home":
		return modelrouter.ProviderInstance{Location: modelrouter.LocationOwnerHardware}, nil
	}
	return modelrouter.ProviderInstance{}, app.ErrProviderUnavailable
}

func (instances) Models(_ context.Context, name string) ([]modelrouter.ModelEntry, error) {
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		return nil, err //nolint:wrapcheck // a test fake
	}
	in, err := kernel.ParseDecimal("0.15")
	if err != nil {
		return nil, err //nolint:wrapcheck // a test fake
	}
	out, err := kernel.ParseDecimal("0.6")
	if err != nil {
		return nil, err //nolint:wrapcheck // a test fake
	}
	switch name {
	case "cloud":
		return []modelrouter.ModelEntry{{
			ID: "openai/gpt-x", HasPrice: true, InputPrice: in, OutputPrice: out, PriceCurrency: usd,
			ContextTokens: 128000, MaxOutputTokens: 8000, ToolCalling: true, StructuredOut: true,
		}}, nil
	case "home":
		return []modelrouter.ModelEntry{{ID: "qwen3:8b", ContextTokens: 32000}}, nil
	}
	return nil, app.ErrProviderUnavailable
}

func (instances) Generate(context.Context, string, string, *modelrouter.Request) (modelrouter.Result, error) {
	return modelrouter.Result{}, app.ErrProviderUnavailable
}

type directory []string

func (d directory) ModelProviders(context.Context) ([]string, error) { return d, nil }

func newProviderClient(t *testing.T) *settingsClient {
	t.Helper()
	store := &memStore{}
	settings := app.NewSettings(store, owners{}).WithProviders(directory{"cloud", "down", "home"}, instances{})
	path, h := connectapi.NewSettings(settings, passThrough{}, slog.New(slog.DiscardHandler))
	mux := http.NewServeMux()
	mux.Handle(path, h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &settingsClient{modelsv1connect.NewModelSettingsServiceClient(srv.Client(), srv.URL)}
}

func TestListModelProvidersShowsTermsAndModels(t *testing.T) {
	t.Parallel()
	c := newProviderClient(t)
	resp, err := c.ListModelProviders(t.Context(), connect.NewRequest(&modelsv1.ListModelProvidersRequest{Origin: "o", SessionToken: "owner"}))
	if err != nil {
		t.Fatal(err)
	}
	want := &modelsv1.ListModelProvidersResponse{Providers: []*modelsv1.ModelProvider{
		{
			Instance: "cloud", Available: true, Location: modelsv1.Location_LOCATION_THIRD_PARTY_CLOUD,
			Retention: &modelsv1.Retention{Days: 30},
			Models: []*modelsv1.ModelEntry{{
				Id: "openai/gpt-x", InputPrice: "0.15", OutputPrice: "0.6", PriceCurrency: "USD",
				ContextTokens: 128000, MaxOutputTokens: 8000, ToolCalling: true, StructuredOutput: true,
			}},
		},
		{Instance: "down"},
		{
			Instance: "home", Available: true, Location: modelsv1.Location_LOCATION_OWNER_HARDWARE, Retention: &modelsv1.Retention{},
			Models: []*modelsv1.ModelEntry{{Id: "qwen3:8b", ContextTokens: 32000}},
		},
	}}
	if diff := cmp.Diff(want, resp.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("providers (-want +got):\n%s", diff)
	}
}

func TestListModelProvidersIsTheOwnersAlone(t *testing.T) {
	t.Parallel()
	c := newProviderClient(t)
	for token, want := range map[string]connect.Code{"": connect.CodeUnauthenticated, "read": connect.CodePermissionDenied} {
		_, err := c.ListModelProviders(t.Context(), connect.NewRequest(&modelsv1.ListModelProvidersRequest{Origin: "o", SessionToken: token}))
		if connect.CodeOf(err) != want {
			t.Errorf("ListModelProviders with %q: %v, want %v", token, err, want)
		}
	}
}

func TestGetModelSettingsNamesTheBrokenRouteEntries(t *testing.T) {
	t.Parallel()
	c := newProviderClient(t)
	ctx := t.Context()
	for role, models := range map[string][]string{"ask": {"gone/m", "cloud/openai/gpt-x"}, "maintainer": {"connector/m"}} {
		if _, err := c.SetRoute(ctx, connect.NewRequest(&modelsv1.SetRouteRequest{Origin: "o", SessionToken: "owner", Role: role, Models: models})); err != nil {
			t.Fatal(err)
		}
	}
	resp, err := c.GetModelSettings(ctx, connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: "o", SessionToken: "owner"}))
	if err != nil {
		t.Fatal(err)
	}
	want := []*modelsv1.BrokenEntry{
		{Role: "ask", Model: "gone/m", Reason: "missing"},
		{Role: "maintainer", Model: "connector/m", Reason: "not-model-provider"},
	}
	if diff := cmp.Diff(want, resp.Msg.GetBroken(), protocmp.Transform()); diff != "" {
		t.Errorf("broken entries (-want +got):\n%s", diff)
	}
}
