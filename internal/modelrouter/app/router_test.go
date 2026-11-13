package app_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

func mustRef(t *testing.T, s string) modelrouter.Ref {
	t.Helper()
	r, err := modelrouter.ParseRef(s)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// routerFixture is a router over fakes, with the ask feature on and a route
// of a cloud model then a local one, the cloud model allowed for public data
// only and the local one for everything.
type routerFixture struct {
	t         *testing.T
	ev        *events
	store     *fakeStore
	providers *fakeProviders
	audit     *fakeAudit
	clock     *clock
	router    *app.Router
	cloud     modelrouter.Ref
	local     modelrouter.Ref
}

func newRouterFixture(t *testing.T) *routerFixture {
	t.Helper()
	ev := &events{}
	f := &routerFixture{
		t: t, ev: ev,
		cloud: mustRef(t, "openrouter/qwen"), local: mustRef(t, "ollama/qwen3:8b"),
		clock: &clock{t: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
	}
	route, err := modelrouter.NewRoute(modelrouter.RoleAsk, []modelrouter.Ref{f.cloud, f.local})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := modelrouter.NewTierPolicy(map[modelrouter.Tier][]modelrouter.Ref{
		modelrouter.TierT0: {f.local}, modelrouter.TierTF: {f.local}, modelrouter.TierT1: {f.local},
		modelrouter.TierT2: {f.cloud, f.local},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.store = &fakeStore{settings: modelrouter.Settings{
		Switch: modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}},
		Routes: map[modelrouter.Role]modelrouter.Route{modelrouter.RoleAsk: route},
		Policy: policy,
	}}
	f.providers = &fakeProviders{ev: ev, answers: map[string]error{}, entries: map[string][]modelrouter.Entry{
		"openrouter": {{ID: "qwen", ToolCalling: true, StructuredOut: true}},
		"ollama":     {{ID: "qwen3:8b", ToolCalling: true}},
	}}
	f.audit = &fakeAudit{ev: ev}
	f.build()
	return f
}

func (f *routerFixture) build() {
	f.router = app.NewRouter(f.store, f.providers, f.audit, f.clock.now)
}

// owner is the actor of the owner's calls.
func owner() app.Actor {
	return app.Actor{Kind: app.ActorOwner, ID: "0123456789abcdef0123456789abcdef"}
}

func request(tier modelrouter.Tier) *modelrouter.Request {
	return &modelrouter.Request{Messages: []modelrouter.Message{{Role: modelrouter.MessageUser, Content: "SECRET-QUESTION-TEXT", Tier: tier}}}
}

func (f *routerFixture) generate(req *modelrouter.Request) (modelrouter.Result, error) {
	f.t.Helper()
	res, err := f.router.Generate(f.t.Context(), modelrouter.RoleAsk, owner(), req)
	if err != nil {
		return modelrouter.Result{}, fmt.Errorf("generating: %w", err)
	}
	return res, nil
}

func TestRouterSwitchOffCallsNothing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mod     func(s *modelrouter.Settings)
		wantErr error
	}{
		{"the zero settings", func(s *modelrouter.Settings) { *s = modelrouter.Settings{} }, modelrouter.ErrSwitchOff},
		{"the switch off with the toggle on", func(s *modelrouter.Settings) { s.Switch.Enabled = false }, modelrouter.ErrSwitchOff},
		{"the switch on, the role's toggle off", func(s *modelrouter.Settings) { s.Switch.Toggles = nil }, modelrouter.ErrFeatureOff},
		{"the switch on, only another toggle on", func(s *modelrouter.Settings) {
			s.Switch.Toggles = map[modelrouter.Feature]bool{modelrouter.FeatureMCP: true, modelrouter.FeatureMaintainer: true}
		}, modelrouter.ErrFeatureOff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newRouterFixture(t)
			tt.mod(&f.store.settings)
			_, err := f.generate(request(modelrouter.TierT2))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			// Nothing reached a provider: no call, not even a listing.
			if diff := cmp.Diff([]string{"audit refused /"}, f.ev.all()); diff != "" {
				t.Errorf("events (-want +got):\n%s", diff)
			}
			if got := f.audit.recs[0]; got.Outcome != app.OutcomeRefused || got.Role != modelrouter.RoleAsk || got.Actor != owner() || !got.Model.IsZero() {
				t.Errorf("refusal record = %+v", got)
			}
		})
	}
}

func TestRouterSwitchOffRefusesEvenAnInvalidRequest(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.store.settings.Switch.Enabled = false
	if _, err := f.generate(&modelrouter.Request{}); !errors.Is(err, modelrouter.ErrSwitchOff) {
		t.Errorf("error = %v, want ErrSwitchOff", err)
	}
}

func TestRouterFailsClosedWithoutSettings(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.store.err = errBoom
	if _, err := f.generate(request(modelrouter.TierT2)); !errors.Is(err, errBoom) {
		t.Fatalf("error = %v, want the store's", err)
	}
	if len(f.ev.all()) != 0 {
		t.Errorf("events = %v, want none", f.ev.all())
	}
}

func TestRouterAuditsBeforeCalling(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	res, err := f.generate(request(modelrouter.TierT2))
	if err != nil {
		t.Fatal(err)
	}
	if res.Model != f.cloud || res.Content != "answer from openrouter/qwen" || res.Usage.OutputTokens != 2 {
		t.Errorf("result = %+v", res)
	}
	want := []string{"list openrouter", "list ollama", "audit ok openrouter/qwen", "generate openrouter/qwen"}
	if diff := cmp.Diff(want, f.ev.all()); diff != "" {
		t.Errorf("events (-want +got):\n%s", diff)
	}
	wantRec := app.CallRecord{At: f.clock.t, Actor: owner(), Role: modelrouter.RoleAsk, Model: f.cloud, Outcome: app.OutcomeCalled}
	if diff := cmp.Diff([]app.CallRecord{wantRec}, f.audit.recs); diff != "" {
		t.Errorf("records (-want +got):\n%s", diff)
	}
}

func TestRouterRefusesWhenTheEntryCannotBeWritten(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.audit.err = errBoom
	_, err := f.generate(request(modelrouter.TierT2))
	if !errors.Is(err, app.ErrAuditUnavailable) || !errors.Is(err, errBoom) {
		t.Fatalf("error = %v, want ErrAuditUnavailable", err)
	}
	if len(f.providers.called) != 0 {
		t.Errorf("a model was called without its audit entry: %v", f.providers.called)
	}
}

func TestRouterRefusalStandsWhenItsEntryCannotBeWritten(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.store.settings.Switch.Enabled = false
	f.audit.err = errBoom
	_, err := f.generate(request(modelrouter.TierT2))
	if !errors.Is(err, modelrouter.ErrSwitchOff) || !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want the refusal joined with the audit failure", err)
	}
	if len(f.providers.called) != 0 {
		t.Errorf("called %v", f.providers.called)
	}
}

func TestRouterTiers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		req       *modelrouter.Request
		wantModel modelrouter.Ref
		wantErr   error
	}{
		{"public data goes to the primary", request(modelrouter.TierT2), mustRef(t, "openrouter/qwen"), nil},
		{"portfolio data skips the cloud model for the local one", request(modelrouter.TierT0), mustRef(t, "ollama/qwen3:8b"), nil},
		{"every tier of the request must be allowed", &modelrouter.Request{Messages: []modelrouter.Message{
			{Role: modelrouter.MessageSystem, Tier: modelrouter.TierT2},
			{Role: modelrouter.MessageUser, Tier: modelrouter.TierT0},
		}}, mustRef(t, "ollama/qwen3:8b"), nil},
		{"untagged data is refused", &modelrouter.Request{Messages: []modelrouter.Message{{Role: modelrouter.MessageUser, Content: "x"}}}, modelrouter.Ref{}, modelrouter.ErrUntagged},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newRouterFixture(t)
			res, err := f.generate(tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if res.Model != tt.wantModel {
				t.Errorf("answered by %v, want %v", res.Model, tt.wantModel)
			}
			if tt.wantErr != nil && len(f.providers.called) != 0 {
				t.Errorf("a refused request reached %v", f.providers.called)
			}
		})
	}
}

func TestRouterDeniesByDefault(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mod  func(s *modelrouter.Settings)
	}{
		{"no policy at all", func(s *modelrouter.Settings) { s.Policy = nil }},
		{"the tier's list is empty", func(s *modelrouter.Settings) { delete(s.Policy, modelrouter.TierT2) }},
		{"the route's models are not listed for the tier", func(s *modelrouter.Settings) {
			s.Policy = modelrouter.TierPolicy{modelrouter.TierT2: {mustRef(t, "other/model")}}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newRouterFixture(t)
			tt.mod(&f.store.settings)
			_, err := f.generate(request(modelrouter.TierT2))
			if !errors.Is(err, modelrouter.ErrTierDenied) {
				t.Fatalf("error = %v, want ErrTierDenied", err)
			}
			if len(f.providers.called) != 0 {
				t.Errorf("called %v", f.providers.called)
			}
			if got := f.audit.recs[len(f.audit.recs)-1].Outcome; got != app.OutcomeRefused {
				t.Errorf("last record outcome = %q, want refused", got)
			}
		})
	}
}

func TestRouterNoRoute(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.store.settings.Routes = nil
	if _, err := f.generate(request(modelrouter.TierT2)); !errors.Is(err, modelrouter.ErrNoRoute) {
		t.Errorf("error = %v, want ErrNoRoute", err)
	}
}

func TestRouterFallsBackOnUnavailable(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.providers.answers["openrouter/qwen"] = app.ErrProviderUnavailable
	res, err := f.generate(request(modelrouter.TierT2))
	if err != nil {
		t.Fatal(err)
	}
	if res.Model != f.local {
		t.Errorf("answered by %v, want the fallback", res.Model)
	}
	want := []string{
		"list openrouter", "list ollama",
		"audit ok openrouter/qwen", "generate openrouter/qwen", "audit failed openrouter/qwen",
		"audit ok ollama/qwen3:8b", "generate ollama/qwen3:8b",
	}
	if diff := cmp.Diff(want, f.ev.all()); diff != "" {
		t.Errorf("events (-want +got):\n%s", diff)
	}
}

func TestRouterFallbackStaysWithinTheTiers(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	// Portfolio data may go to the local model only. When it is down there
	// is no fallback to the cloud model.
	f.providers.answers["ollama/qwen3:8b"] = app.ErrProviderUnavailable
	_, err := f.generate(request(modelrouter.TierT0))
	if !errors.Is(err, app.ErrNoAnswer) {
		t.Fatalf("error = %v, want ErrNoAnswer", err)
	}
	if diff := cmp.Diff([]string{"ollama/qwen3:8b"}, f.providers.called); diff != "" {
		t.Errorf("called (-want +got):\n%s", diff)
	}
}

func TestRouterNoAnswerWhenEveryModelIsDown(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.providers.answers["openrouter/qwen"] = app.ErrProviderUnavailable
	f.providers.answers["ollama/qwen3:8b"] = app.ErrProviderUnavailable
	if _, err := f.generate(request(modelrouter.TierT2)); !errors.Is(err, app.ErrNoAnswer) || !errors.Is(err, app.ErrProviderUnavailable) {
		t.Errorf("error = %v, want ErrNoAnswer", err)
	}
}

func TestRouterRejectionEndsTheCall(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.providers.answers["openrouter/qwen"] = app.ErrProviderRejected
	_, err := f.generate(request(modelrouter.TierT2))
	if !errors.Is(err, app.ErrProviderRejected) {
		t.Fatalf("error = %v, want ErrProviderRejected", err)
	}
	if diff := cmp.Diff([]string{"openrouter/qwen"}, f.providers.called); diff != "" {
		t.Errorf("called (-want +got):\n%s", diff)
	}
}

func TestRouterSkipsModelsWithoutWhatTheRequestNeeds(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	req := request(modelrouter.TierT2)
	req.ResponseSchemaJSON = `{"type":"object"}`
	res, err := f.generate(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Model != f.cloud { // the cloud entry has structured output
		t.Errorf("answered by %v", res.Model)
	}
	f = newRouterFixture(t)
	f.store.settings.Policy[modelrouter.TierT2] = []modelrouter.Ref{f.local} // only the model without it
	req = request(modelrouter.TierT2)
	req.ResponseSchemaJSON = `{"type":"object"}`
	if _, err := f.generate(req); !errors.Is(err, modelrouter.ErrNoCapableModel) {
		t.Errorf("error = %v, want ErrNoCapableModel", err)
	}
}

func TestRouterCachesEntriesUntilTheyExpire(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	for range 2 {
		if _, err := f.generate(request(modelrouter.TierT2)); err != nil {
			t.Fatal(err)
		}
	}
	count := func() int { return strings.Count(strings.Join(f.ev.all(), ","), "list openrouter") }
	if count() != 1 {
		t.Errorf("listed openrouter %d times within the TTL, want 1", count())
	}
	f.clock.t = f.clock.t.Add(6 * time.Minute)
	if _, err := f.generate(request(modelrouter.TierT2)); err != nil {
		t.Fatal(err)
	}
	if count() != 2 {
		t.Errorf("listed openrouter %d times after the TTL, want 2", count())
	}
}

func TestRouterWorksWhenAProviderCannotListItsModels(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.providers.listErr = errBoom
	res, err := f.generate(request(modelrouter.TierT2))
	if err != nil {
		t.Fatal(err)
	}
	if res.Model != f.cloud {
		t.Errorf("answered by %v", res.Model)
	}
}

// TestRouterKeepsTextOutOfRecordsAndErrors checks GR-12 and D-211: a call's
// record has no field for the question or the answer, and no error the
// router returns quotes the question.
func TestRouterKeepsTextOutOfRecordsAndErrors(t *testing.T) {
	t.Parallel()
	const secret = "SECRET-QUESTION-TEXT"
	f := newRouterFixture(t)
	f.providers.answers["openrouter/qwen"] = app.ErrProviderUnavailable
	f.providers.answers["ollama/qwen3:8b"] = app.ErrProviderUnavailable
	_, err := f.generate(request(modelrouter.TierT2))
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Errorf("error = %v", err)
	}
	for _, rec := range f.audit.recs {
		if strings.Contains(strings.Join([]string{string(rec.Actor.Kind), rec.Actor.ID, string(rec.Role), rec.Model.String(), string(rec.Outcome)}, " "), secret) {
			t.Errorf("a record holds the question: %+v", rec)
		}
	}
}
