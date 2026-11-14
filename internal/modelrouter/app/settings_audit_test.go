package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

// fakeSettingsAudit records the entries of changes to the settings, and fails
// when err is set.
type fakeSettingsAudit struct {
	err     error
	changes []app.SettingChange
}

func (f *fakeSettingsAudit) RecordChange(_ context.Context, c *app.SettingChange) error {
	if f.err != nil {
		return f.err
	}
	f.changes = append(f.changes, *c)
	return nil
}

// fakeDirectory lists the instances that run, or fails.
type fakeDirectory struct {
	names []string
	err   error
}

func (f fakeDirectory) ModelProviders(context.Context) ([]string, error) { return f.names, f.err }

// settingsClock is the time the audited settings are stamped with.
func settingsClock() time.Time { return time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC) }

func auditedSettings(audit *fakeSettingsAudit) (*app.Settings, *fakeStore) {
	store := &fakeStore{}
	s := app.NewSettings(store, fakeOwners{token: ownerToken}).WithAudit(audit, settingsClock)
	return s, store
}

func TestSettingsEveryChangeIsAuditedWithTheOwnersSession(t *testing.T) {
	t.Parallel()
	audit := &fakeSettingsAudit{}
	s, _ := auditedSettings(audit)
	ctx := t.Context()
	steps := []error{
		s.SetEnabled(ctx, origin, ownerToken, true),
		s.SetFeature(ctx, origin, ownerToken, "ask", true),
		s.SetFeature(ctx, origin, ownerToken, "mcp", false),
		s.SetRoute(ctx, origin, ownerToken, "maintainer", []string{"p/m", "q/n"}),
		s.SetDataTierModels(ctx, origin, ownerToken, "T0", []string{"p/m"}),
		s.SetEnabled(ctx, origin, ownerToken, false),
	}
	for i, err := range steps {
		if err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
	got := make([]string, 0, len(audit.changes))
	for _, c := range audit.changes {
		if c.Session != "session-"+ownerToken || !c.At.Equal(settingsClock()) {
			t.Errorf("entry = %+v, want the owner's session and the clock's time", c)
		}
		got = append(got, c.Setting)
	}
	want := []string{"switch:on", "feature:ask:on", "feature:mcp:off", "route:maintainer", "tier:T0", "switch:off"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("audited settings (-want +got):\n%s", diff)
	}
}

func TestSettingsAChangeTheLogCannotRecordIsNotMade(t *testing.T) {
	t.Parallel()
	audit := &fakeSettingsAudit{err: errBoom}
	s, store := auditedSettings(audit)
	ctx := t.Context()
	for name, err := range map[string]error{
		"switch": s.SetEnabled(ctx, origin, ownerToken, true),
		"toggle": s.SetFeature(ctx, origin, ownerToken, "ask", true),
		"route":  s.SetRoute(ctx, origin, ownerToken, "ask", []string{"p/m"}),
		"tier":   s.SetDataTierModels(ctx, origin, ownerToken, "T2", []string{"p/m"}),
	} {
		if !errors.Is(err, errBoom) {
			t.Errorf("%s: error = %v, want the log's", name, err)
		}
	}
	if store.settings.Switch.Enabled || len(store.settings.Routes) != 0 || len(store.settings.Policy) != 0 || store.settings.Switch.Allows(modelrouter.FeatureAsk) {
		t.Errorf("settings = %+v, want nothing changed", store.settings)
	}
}

func TestSettingsARefusedChangeWritesNoEntry(t *testing.T) {
	t.Parallel()
	audit := &fakeSettingsAudit{}
	s, _ := auditedSettings(audit)
	ctx := t.Context()
	for _, err := range []error{
		s.SetEnabled(ctx, origin, "read", true),
		s.SetFeature(ctx, origin, ownerToken, "trade", true),
		s.SetRoute(ctx, origin, ownerToken, "ask", []string{"p/m", "p/m"}),
	} {
		if err == nil {
			t.Error("a refused or invalid change was no error")
		}
	}
	if len(audit.changes) != 0 {
		t.Errorf("entries = %+v, want none for refused or invalid changes", audit.changes)
	}
}

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestSettingsProvidersListTermsAndModelsForTheOwner(t *testing.T) {
	t.Parallel()
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatal(err)
	}
	entry := modelrouter.ModelEntry{
		ID: "gpt-x", HasPrice: true, InputPrice: dec(t, "0.15"), OutputPrice: dec(t, "0.6"), PriceCurrency: usd,
		ContextTokens: 128000, MaxOutputTokens: 8000, ToolCalling: true, StructuredOut: true,
	}
	providers := &fakeProviders{ev: &events{}, entries: map[string][]modelrouter.ModelEntry{"cloud": {entry}}}
	s := app.NewSettings(&fakeStore{}, fakeOwners{token: ownerToken}).WithProviders(fakeDirectory{names: []string{"cloud", "local"}}, providers)
	got, err := s.Providers(t.Context(), origin, ownerToken)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Instance != "cloud" || !got[0].Available || got[0].Terms.Location != modelrouter.LocationThirdPartyCloud {
		t.Fatalf("providers = %+v", got)
	}
	if diff := cmp.Diff([]modelrouter.ModelEntry{entry}, got[0].Models, cmp.Comparer(func(a, b kernel.Decimal) bool { return a.Equal(b) })); diff != "" {
		t.Errorf("models (-want +got):\n%s", diff)
	}
	if !got[1].Available || len(got[1].Models) != 0 {
		t.Errorf("local = %+v, want available with no models", got[1])
	}
}

func TestSettingsProvidersMarkAnInstanceThatDoesNotAnswerUnavailable(t *testing.T) {
	t.Parallel()
	providers := &fakeProviders{ev: &events{}, listErr: app.ErrProviderUnavailable}
	s := app.NewSettings(&fakeStore{}, fakeOwners{token: ownerToken}).WithProviders(fakeDirectory{names: []string{"down"}}, providers)
	got, err := s.Providers(t.Context(), origin, ownerToken)
	if err != nil || len(got) != 1 || got[0].Available || len(got[0].Models) != 0 {
		t.Errorf("providers = %+v, %v; want one unavailable, with no models", got, err)
	}
}

func TestSettingsProvidersAreTheOwnersAndEmptyBeforeThePluginsAreFound(t *testing.T) {
	t.Parallel()
	s := app.NewSettings(&fakeStore{}, fakeOwners{token: ownerToken}).
		WithProviders(fakeDirectory{err: app.ErrPluginsNotFound}, &fakeProviders{ev: &events{}})
	for _, tt := range []struct {
		token string
		want  error
	}{{"nobody", app.ErrNoSession}, {"read", app.ErrNotOwner}} {
		if _, err := s.Providers(t.Context(), origin, tt.token); !errors.Is(err, tt.want) {
			t.Errorf("Providers as %s: %v, want %v", tt.token, err, tt.want)
		}
	}
	if got, err := s.Providers(t.Context(), origin, ownerToken); err != nil || len(got) != 0 {
		t.Errorf("Providers before discovery = %+v, %v; want none", got, err)
	}
	bare := app.NewSettings(&fakeStore{}, fakeOwners{token: ownerToken})
	if got, err := bare.Providers(t.Context(), origin, ownerToken); err != nil || len(got) != 0 {
		t.Errorf("Providers without a directory = %+v, %v; want none", got, err)
	}
}

func TestSettingsBrokenRoutesCountEveryRouteInUseOrNot(t *testing.T) {
	t.Parallel()
	providers := &fakeProviders{ev: &events{}, instances: map[string]error{
		"gone": app.ErrProviderMissing, "plug": app.ErrNotModelProvider, "slow": app.ErrProviderUnavailable,
	}}
	store := &fakeStore{}
	s := app.NewSettings(store, fakeOwners{token: ownerToken}).WithProviders(fakeDirectory{}, providers)
	ctx := t.Context()
	// The switch is off: no route is in use, and the page still shows what is broken.
	for role, refs := range map[string][]string{"ask": {"gone/m", "ok/m", "slow/m"}, "maintainer": {"plug/m"}} {
		if err := s.SetRoute(ctx, origin, ownerToken, role, refs); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.BrokenRoutes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := make([]string, 0, len(got))
	for _, p := range got {
		found = append(found, string(p.Role)+" "+p.Model.String()+" "+string(p.Reason))
	}
	slices.Sort(found)
	want := []string{"ask gone/m missing", "maintainer plug/m not-model-provider"}
	if diff := cmp.Diff(want, found); diff != "" {
		t.Errorf("broken (-want +got):\n%s", diff)
	}
	providers.instances["ok"] = app.ErrPluginsNotFound
	if got, err := s.BrokenRoutes(ctx); err != nil || len(got) != 0 {
		t.Errorf("BrokenRoutes before discovery = %+v, %v; want none", got, err)
	}
}
