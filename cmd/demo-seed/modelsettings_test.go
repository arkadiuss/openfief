package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	modelspg "github.com/arkadiuss/openfief/internal/modelrouter/postgres"
)

// committedModelSettings is the demo's model layer's settings
// (deploy/demo/model), which the demo-seed Job stores once the layer is on.
const committedModelSettings = "../../deploy/demo/model/settings.json"

func refs(t *testing.T, ss ...string) []modelrouter.Ref {
	t.Helper()
	out, err := modelrouter.ParseRefs(ss)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestParseModelSettings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		json       string
		wantRoutes map[modelrouter.Role][]modelrouter.Ref
		wantTiers  map[modelrouter.DataTier][]modelrouter.Ref
	}{
		{
			name: "a route and a tier",
			json: `{"routes":{"ask":["or/vendor/a","or/vendor/b"]},"tiers":{"T2":["or/vendor/a"]}}`,
			wantRoutes: map[modelrouter.Role][]modelrouter.Ref{
				modelrouter.RoleAsk: {{Instance: "or", Model: "vendor/a"}, {Instance: "or", Model: "vendor/b"}},
			},
			wantTiers: map[modelrouter.DataTier][]modelrouter.Ref{modelrouter.DataTierT2: {{Instance: "or", Model: "vendor/a"}}},
		},
		{name: "nothing", json: `{}`, wantRoutes: map[modelrouter.Role][]modelrouter.Ref{}, wantTiers: map[modelrouter.DataTier][]modelrouter.Ref{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseModelSettings([]byte(tt.json))
			if err != nil {
				t.Fatalf("parseModelSettings: %v", err)
			}
			if diff := cmp.Diff(tt.wantRoutes, got.routes); diff != "" {
				t.Errorf("routes (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantTiers, got.tiers); diff != "" {
				t.Errorf("tiers (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseModelSettingsRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, json string }{
		{"a switch is not the seed's to turn", `{"switch":{"enabled":true}}`},
		{"a toggle is not the seed's to turn", `{"toggles":{"ask":true}}`},
		{"not JSON", `routes`},
		{"an unknown role", `{"routes":{"trader":["or/m"]}}`},
		{"a route with no model", `{"routes":{"ask":[]}}`},
		{"a route naming a model twice", `{"routes":{"ask":["or/m","or/m"]}}`},
		{"a reference without a provider", `{"routes":{"ask":["m"]}}`},
		{"a provider that is no DNS label", `{"routes":{"ask":["Open_Router/m"]}}`},
		{"an unknown tier", `{"tiers":{"T9":["or/m"]}}`},
		{"a tier naming a model twice", `{"tiers":{"T0":["or/m","or/m"]}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got, err := parseModelSettings([]byte(tt.json)); !errors.Is(err, errBadModelSettings) {
				t.Fatalf("parseModelSettings = %+v, %v; want an error wrapping errBadModelSettings", got, err)
			}
		})
	}
}

// errNoAuditHook is what the fake store answers a save that carries an audit
// hook: the seed has none to run.
var errNoAuditHook = errors.New("the seed has no audit hook to run")

// fakeSettingsStore records what the seed saves.
type fakeSettingsStore struct {
	settings modelrouter.Settings
	routes   map[modelrouter.Role][]modelrouter.Ref
	tiers    map[modelrouter.DataTier][]modelrouter.Ref
	saves    int
}

func (f *fakeSettingsStore) Settings(context.Context) (modelrouter.Settings, error) {
	return f.settings, nil
}

func (f *fakeSettingsStore) SaveRoute(_ context.Context, role modelrouter.Role, route []modelrouter.Ref, within func(context.Context) error) error {
	f.saves++
	if within != nil {
		return errNoAuditHook
	}
	if f.routes == nil {
		f.routes = map[modelrouter.Role][]modelrouter.Ref{}
	}
	f.routes[role] = route
	return nil
}

func (f *fakeSettingsStore) SaveDataTierModels(_ context.Context, tier modelrouter.DataTier, allowed []modelrouter.Ref, within func(context.Context) error) error {
	f.saves++
	if within != nil {
		return errNoAuditHook
	}
	if f.tiers == nil {
		f.tiers = map[modelrouter.DataTier][]modelrouter.Ref{}
	}
	f.tiers[tier] = allowed
	return nil
}

func TestSeedModelSettings(t *testing.T) {
	t.Parallel()
	ms, err := parseModelSettings([]byte(`{"routes":{"ask":["or/vendor/a"]},"tiers":{"T0":["or/vendor/a"],"T2":["or/vendor/a"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	owners := modelrouter.Settings{Routes: map[modelrouter.Role]modelrouter.Route{
		modelrouter.RoleAsk: {Role: modelrouter.RoleAsk, Models: refs(t, "mine/own")},
	}}
	tests := []struct {
		name        string
		store       *fakeSettingsStore
		wantApplied bool
		wantSaves   int
	}{
		{name: "an empty store gets the routes and tiers", store: &fakeSettingsStore{}, wantApplied: true, wantSaves: 3},
		{name: "a store with the owner's route is left alone", store: &fakeSettingsStore{settings: owners}},
		{
			name:  "a store with the owner's tier list is left alone",
			store: &fakeSettingsStore{settings: modelrouter.Settings{Policy: modelrouter.DataTierPolicy{modelrouter.DataTierT1: refs(t, "mine/own")}}},
		},
		{
			name: "a store with the switch saved as off and nothing else still gets them",
			store: &fakeSettingsStore{settings: modelrouter.Settings{
				Switch: modelrouter.Switch{Enabled: false, Toggles: map[modelrouter.Feature]bool{}},
			}},
			wantApplied: true, wantSaves: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			applied, err := seedModelSettings(t.Context(), tt.store, &ms)
			if err != nil {
				t.Fatalf("seedModelSettings: %v", err)
			}
			if applied != tt.wantApplied || tt.store.saves != tt.wantSaves {
				t.Errorf("applied %v with %d saves, want %v with %d", applied, tt.store.saves, tt.wantApplied, tt.wantSaves)
			}
		})
	}
}

// TestTheDemosModelSettingsRouteOnlyTheAskBoxToItsModel checks the committed
// file of the demo's model layer: it parses the way the loader parses it, routes
// the Ask box to the demo's model on OpenRouter, routes nothing for the
// maintainer (the demo has no maintainer route), allows that model for every
// data tier (D-206, the demo's data is synthetic) and carries no switch.
func TestTheDemosModelSettingsRouteOnlyTheAskBoxToItsModel(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(committedModelSettings)
	if err != nil {
		t.Fatal(err)
	}
	ms, err := parseModelSettings(data)
	if err != nil {
		t.Fatalf("the committed settings: %v", err)
	}
	wantRoute := refs(t, "openrouter/openai/gpt-oss-120b")
	if diff := cmp.Diff(map[modelrouter.Role][]modelrouter.Ref{modelrouter.RoleAsk: wantRoute}, ms.routes); diff != "" {
		t.Errorf("routes (-want +got):\n%s", diff)
	}
	for _, tier := range modelrouter.DataTiers() {
		if diff := cmp.Diff(wantRoute, ms.tiers[tier]); diff != "" {
			t.Errorf("models allowed for %s (-want +got):\n%s", tier, diff)
		}
	}
}

func TestParseConfigTakesTheModelSettingsFileFromItsFlag(t *testing.T) {
	t.Parallel()
	cfg, err := parseConfig([]string{"--model-settings", "/m/settings.json"}, "postgres://demo", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ModelSettings != "/m/settings.json" {
		t.Errorf("ModelSettings = %q, want the flag's", cfg.ModelSettings)
	}
	cfg, err = parseConfig(nil, "postgres://demo", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ModelSettings != "" {
		t.Errorf("ModelSettings = %q, want none without the flag: main fills it from %s", cfg.ModelSettings, modelSettingsEnv)
	}
}

// writeSettingsFile writes the committed settings into a temporary directory
// and returns the file's path.
func writeSettingsFile(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(committedModelSettings)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := root.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	if err := root.WriteFile("settings.json", data, 0o600); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "settings.json")
}

// checkSeededSettings checks what a seed of the committed file leaves: the
// Ask box's route to the demo's model and every tier allowing it, no
// maintainer route, and the switch off with no toggle on, so no model has
// been turned on.
func checkSeededSettings(t *testing.T, got *modelrouter.Settings) {
	t.Helper()
	if got.Switch.Enabled || len(got.Switch.Toggles) != 0 {
		t.Errorf("switch = %+v, want it off with no toggle on", got.Switch)
	}
	if len(got.Routes[modelrouter.RoleAsk].Models) != 1 {
		t.Errorf("the Ask box's route = %v, want the demo's model", got.Routes[modelrouter.RoleAsk].Models)
	}
	if _, ok := got.Routes[modelrouter.RoleMaintainer]; ok {
		t.Error("the maintainer has a route, want none")
	}
	if got.HasModel(modelrouter.RoleAsk) {
		t.Error("the Ask box has a model with the switch off")
	}
	if len(got.Policy) != len(modelrouter.DataTiers()) {
		t.Errorf("tiers with models = %d, want every tier", len(got.Policy))
	}
}

// TestApplyModelSettingsOnARealDatabase stores the committed file into a fresh
// database and again after the owner changed the route: the first stores the
// route and the tiers and leaves the switch off, the second changes nothing.
func TestApplyModelSettingsOnARealDatabase(t *testing.T) {
	t.Parallel()
	pool := emptyPool(t)
	ctx := t.Context()
	path := writeSettingsFile(t)
	if err := applyModelSettings(ctx, pool, path, quiet()); err != nil {
		t.Fatalf("applyModelSettings: %v", err)
	}
	store := modelspg.New(pool)
	seeded, err := store.Settings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	checkSeededSettings(t, &seeded)
	owners := refs(t, "mine/own")
	if saveErr := store.SaveRoute(ctx, modelrouter.RoleAsk, owners, nil); saveErr != nil {
		t.Fatal(saveErr)
	}
	if applyErr := applyModelSettings(ctx, pool, path, quiet()); applyErr != nil {
		t.Fatalf("applyModelSettings again: %v", applyErr)
	}
	again, err := store.Settings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(owners, again.Routes[modelrouter.RoleAsk].Models); diff != "" {
		t.Errorf("the owner's route after a second seed (-want +got):\n%s", diff)
	}
}
