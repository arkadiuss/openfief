package app_test

import (
	"context"
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
	attention *fakeAttention
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
	policy, err := modelrouter.NewDataTierPolicy(map[modelrouter.DataTier][]modelrouter.Ref{
		modelrouter.DataTierT0: {f.local}, modelrouter.DataTierTF: {f.local}, modelrouter.DataTierT1: {f.local},
		modelrouter.DataTierT2: {f.cloud, f.local},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.store = &fakeStore{settings: modelrouter.Settings{
		Switch: modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}},
		Routes: map[modelrouter.Role]modelrouter.Route{modelrouter.RoleAsk: route},
		Policy: policy,
	}}
	f.providers = &fakeProviders{ev: ev, answers: map[string]error{}, entries: map[string][]modelrouter.ModelEntry{
		"openrouter": {{ID: "qwen", ToolCalling: true, StructuredOut: true}},
		"ollama":     {{ID: "qwen3:8b", ToolCalling: true}},
	}}
	f.audit = &fakeAudit{ev: ev}
	f.attention = &fakeAttention{ev: ev}
	f.build()
	return f
}

func (f *routerFixture) build() {
	f.router = app.NewRouter(&app.RouterPorts{
		Settings: f.store, Providers: f.providers, Audit: f.audit, Attention: f.attention, CallIDs: callIDs(), Now: f.clock.now,
	})
}

// owner is the actor of the owner's calls.
func owner() app.Actor {
	return app.Actor{Kind: app.ActorOwner, ID: "0123456789abcdef0123456789abcdef"}
}

func request(tier modelrouter.DataTier) *modelrouter.Request {
	return &modelrouter.Request{Messages: []modelrouter.Message{{Role: modelrouter.MessageUser, Content: "SECRET-QUESTION-TEXT", DataTier: tier}}}
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
			_, err := f.generate(request(modelrouter.DataTierT2))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			// Nothing reached a provider: no call, not even a listing.
			if diff := cmp.Diff([]string{"audit refused /"}, f.ev.all()); diff != "" {
				t.Errorf("events (-want +got):\n%s", diff)
			}
			if len(f.audit.sent) != 0 {
				t.Errorf("a refusal wrote a sent entry: %+v", f.audit.sent)
			}
			if got := f.audit.results[0]; got.Outcome != app.OutcomeRefused || got.Role != modelrouter.RoleAsk || got.Actor != owner() || !got.Model.IsZero() || got.CallID == "" {
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
	if _, err := f.generate(request(modelrouter.DataTierT2)); !errors.Is(err, errBoom) {
		t.Fatalf("error = %v, want the store's", err)
	}
	if len(f.ev.all()) != 0 {
		t.Errorf("events = %v, want none", f.ev.all())
	}
}

func TestRouterAuditsBeforeCalling(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	res, err := f.generate(request(modelrouter.DataTierT2))
	if err != nil {
		t.Fatal(err)
	}
	if res.Model != f.cloud || res.Content != "answer from openrouter/qwen" || res.Usage.OutputTokens != 2 {
		t.Errorf("result = %+v", res)
	}
	want := []string{"list openrouter", "list ollama", "audit sent openrouter/qwen", "generate openrouter/qwen", "audit ok openrouter/qwen"}
	if diff := cmp.Diff(want, f.ev.all()); diff != "" {
		t.Errorf("events (-want +got):\n%s", diff)
	}
	const id = "0000000000000001"
	wantSent := app.CallSent{At: f.clock.t, Actor: owner(), Role: modelrouter.RoleAsk, Model: f.cloud, CallID: id}
	if diff := cmp.Diff([]app.CallSent{wantSent}, f.audit.sent); diff != "" {
		t.Errorf("sent entries (-want +got):\n%s", diff)
	}
	wantResult := app.CallResult{At: f.clock.t, Actor: owner(), Role: modelrouter.RoleAsk, Model: f.cloud, CallID: id, Outcome: app.OutcomeOK}
	if diff := cmp.Diff([]app.CallResult{wantResult}, f.audit.results); diff != "" {
		t.Errorf("result entries (-want +got):\n%s", diff)
	}
}

// TestRouterLinksEachAttemptsEntries checks that each model tried has its
// own pair of entries under its own call ID, and that a failure records its
// error class, never the error's text.
func TestRouterLinksEachAttemptsEntries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		err       error
		wantClass app.ErrorClass
		wantNext  bool
	}{
		{"unavailable falls through", app.ErrProviderUnavailable, app.ClassUnavailable, true},
		{"misconfigured falls through", app.ErrProviderMissing, app.ClassMisconfigured, true},
		{"a plugin that is no model provider falls through", app.ErrNotModelProvider, app.ClassMisconfigured, true},
		{"rejected ends the call", app.ErrProviderRejected, app.ClassRejected, false},
		{"canceled ends the call", context.Canceled, app.ClassCanceled, false},
		{"anything else ends the call", errBoom, app.ClassInternal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newRouterFixture(t)
			f.providers.answers["openrouter/qwen"] = fmt.Errorf("%w: SECRET-PROVIDER-REPLY", tt.err)
			_, err := f.generate(request(modelrouter.DataTierT2))
			if tt.wantNext != (err == nil) {
				t.Fatalf("error = %v, want the fallback to answer: %v", err, tt.wantNext)
			}
			f.wantPairs(tt.wantClass, tt.wantNext)
		})
	}
}

// wantPairs checks the entries of a call whose primary failed with class:
// a sent entry and a failed result under call ID 1 and, when the fallback
// was tried, a sent entry and an ok result under call ID 2.
func (f *routerFixture) wantPairs(class app.ErrorClass, fallback bool) {
	f.t.Helper()
	wantSent := []app.CallSent{{At: f.clock.t, Actor: owner(), Role: modelrouter.RoleAsk, Model: f.cloud, CallID: "0000000000000001"}}
	wantResults := []app.CallResult{{
		At: f.clock.t, Actor: owner(), Role: modelrouter.RoleAsk, Model: f.cloud, CallID: "0000000000000001",
		Outcome: app.OutcomeFailed, Class: class,
	}}
	if fallback {
		wantSent = append(wantSent, app.CallSent{At: f.clock.t, Actor: owner(), Role: modelrouter.RoleAsk, Model: f.local, CallID: "0000000000000002"})
		wantResults = append(wantResults, app.CallResult{
			At: f.clock.t, Actor: owner(), Role: modelrouter.RoleAsk, Model: f.local, CallID: "0000000000000002", Outcome: app.OutcomeOK,
		})
	}
	if diff := cmp.Diff(wantSent, f.audit.sent); diff != "" {
		f.t.Errorf("sent entries (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(wantResults, f.audit.results); diff != "" {
		f.t.Errorf("result entries (-want +got):\n%s", diff)
	}
}

// TestRouterRefusalHasOnlyTheResultEntry checks that a call the router
// refuses before it chooses a model, by the data tiers here, has one entry:
// the result, refused, with no model, and nothing was sent.
func TestRouterRefusalHasOnlyTheResultEntry(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.store.settings.Policy = nil
	if _, err := f.generate(request(modelrouter.DataTierT0)); !errors.Is(err, modelrouter.ErrDataTierDenied) {
		t.Fatalf("error = %v, want ErrDataTierDenied", err)
	}
	if len(f.audit.sent) != 0 {
		t.Errorf("sent entries = %+v, want none", f.audit.sent)
	}
	want := []app.CallResult{{At: f.clock.t, Actor: owner(), Role: modelrouter.RoleAsk, CallID: "0000000000000001", Outcome: app.OutcomeRefused}}
	if diff := cmp.Diff(want, f.audit.results); diff != "" {
		t.Errorf("result entries (-want +got):\n%s", diff)
	}
}

// TestRouterWithholdsAnAnswerItCannotRecord checks that when the entry of an
// answered call's result can't be written, the caller gets the error and not
// the answer: the log fails closed.
func TestRouterWithholdsAnAnswerItCannotRecord(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.audit.resultErr = errBoom
	res, err := f.generate(request(modelrouter.DataTierT2))
	if !errors.Is(err, app.ErrAuditUnavailable) || !errors.Is(err, errBoom) {
		t.Fatalf("error = %v, want ErrAuditUnavailable", err)
	}
	if res.Content != "" {
		t.Errorf("the answer was returned: %+v", res)
	}
	if len(f.audit.sent) != 1 || len(f.providers.called) != 1 {
		t.Errorf("sent %d entries and called %v, want one each", len(f.audit.sent), f.providers.called)
	}
}

// TestRouterRecordsTheResultOfACanceledCall checks that the result entry of a
// call is written even when the caller's context ended: the call was sent.
func TestRouterRecordsTheResultOfACanceledCall(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.providers.answers["openrouter/qwen"] = context.Canceled
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	f.audit.failsOnEndedContext = true
	cancel()
	_, err := f.router.Generate(ctx, modelrouter.RoleAsk, owner(), request(modelrouter.DataTierT2))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want canceled", err)
	}
	if len(f.audit.results) != 1 || f.audit.results[0].Class != app.ClassCanceled {
		t.Errorf("results = %+v, want a canceled one", f.audit.results)
	}
}

// TestRouterRaisesAnItemForABrokenRouteEntry checks D-196: a route entry whose
// provider instance is missing or is no model provider falls through to the
// next model and raises one attention item naming the route and the
// instance.
func TestRouterRaisesAnItemForABrokenRouteEntry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		err        error
		wantDetail string
	}{
		{"a missing provider instance", app.ErrProviderMissing, "the ask route names provider instance openrouter: no plugin has that name"},
		{"a plugin that is no model provider", app.ErrNotModelProvider, "the ask route names provider instance openrouter: that plugin is not a model provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newRouterFixture(t)
			f.providers.answers["openrouter/qwen"] = tt.err
			res, err := f.generate(request(modelrouter.DataTierT2))
			if err != nil || res.Model != f.local {
				t.Fatalf("result = %+v, %v; want the fallback's answer", res, err)
			}
			want := map[string]string{"ask/openrouter/qwen": tt.wantDetail}
			if diff := cmp.Diff(want, f.attention.raised); diff != "" {
				t.Errorf("raised items (-want +got):\n%s", diff)
			}
			if got := f.attention.kinds["ask/openrouter/qwen"]; got != modelrouter.AttentionKind {
				t.Errorf("kind = %q, want %q", got, modelrouter.AttentionKind)
			}
		})
	}
}

// TestRouterRaisesNoItemForATransientFailure checks that a provider that is
// only down or slow, or that rejects a request, falls through or ends the
// call and raises no item: the owner has nothing to fix.
func TestRouterRaisesNoItemForATransientFailure(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct{ err, wantErr error }{
		"unavailable": {app.ErrProviderUnavailable, nil}, "rejected": {app.ErrProviderRejected, app.ErrProviderRejected},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := newRouterFixture(t)
			f.providers.answers["openrouter/qwen"] = tt.err
			if _, err := f.generate(request(modelrouter.DataTierT2)); !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if len(f.attention.raised) != 0 {
				t.Errorf("raised %v, want no item", f.attention.raised)
			}
		})
	}
}

// TestRouterFallsThroughWhenTheItemCannotBeRaised checks that a failure to
// raise the item is not the call's: the next model still answers.
func TestRouterFallsThroughWhenTheItemCannotBeRaised(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.attention.err = errBoom
	f.providers.answers["openrouter/qwen"] = app.ErrProviderMissing
	if res, err := f.generate(request(modelrouter.DataTierT2)); err != nil || res.Model != f.local {
		t.Errorf("result = %+v, %v; want the fallback's answer", res, err)
	}
}

func TestRouterRefusesWhenTheEntryCannotBeWritten(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.audit.err = errBoom
	_, err := f.generate(request(modelrouter.DataTierT2))
	if !errors.Is(err, app.ErrAuditUnavailable) || !errors.Is(err, errBoom) {
		t.Fatalf("error = %v, want ErrAuditUnavailable", err)
	}
	if len(f.providers.called) != 0 {
		t.Errorf("a model was called without its audit entry: %v", f.providers.called)
	}
	if len(f.audit.results) != 0 {
		t.Errorf("result entries = %+v, want none for a call that was not made", f.audit.results)
	}
}

func TestRouterRefusalStandsWhenItsEntryCannotBeWritten(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.store.settings.Switch.Enabled = false
	f.audit.err = errBoom
	_, err := f.generate(request(modelrouter.DataTierT2))
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
		{"public data goes to the primary", request(modelrouter.DataTierT2), mustRef(t, "openrouter/qwen"), nil},
		{"portfolio data skips the cloud model for the local one", request(modelrouter.DataTierT0), mustRef(t, "ollama/qwen3:8b"), nil},
		{"every tier of the request must be allowed", &modelrouter.Request{Messages: []modelrouter.Message{
			{Role: modelrouter.MessageSystem, DataTier: modelrouter.DataTierT2},
			{Role: modelrouter.MessageUser, DataTier: modelrouter.DataTierT0},
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
		{"the tier's list is empty", func(s *modelrouter.Settings) { delete(s.Policy, modelrouter.DataTierT2) }},
		{"the route's models are not listed for the tier", func(s *modelrouter.Settings) {
			s.Policy = modelrouter.DataTierPolicy{modelrouter.DataTierT2: {mustRef(t, "other/model")}}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newRouterFixture(t)
			tt.mod(&f.store.settings)
			_, err := f.generate(request(modelrouter.DataTierT2))
			if !errors.Is(err, modelrouter.ErrDataTierDenied) {
				t.Fatalf("error = %v, want ErrDataTierDenied", err)
			}
			if len(f.providers.called) != 0 {
				t.Errorf("called %v", f.providers.called)
			}
			if got := f.audit.results[len(f.audit.results)-1].Outcome; got != app.OutcomeRefused {
				t.Errorf("last result outcome = %q, want refused", got)
			}
		})
	}
}

func TestRouterNoRoute(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.store.settings.Routes = nil
	if _, err := f.generate(request(modelrouter.DataTierT2)); !errors.Is(err, modelrouter.ErrNoRoute) {
		t.Errorf("error = %v, want ErrNoRoute", err)
	}
}

func TestRouterFallsBackOnUnavailable(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.providers.answers["openrouter/qwen"] = app.ErrProviderUnavailable
	res, err := f.generate(request(modelrouter.DataTierT2))
	if err != nil {
		t.Fatal(err)
	}
	if res.Model != f.local {
		t.Errorf("answered by %v, want the fallback", res.Model)
	}
	want := []string{
		"list openrouter", "list ollama",
		"audit sent openrouter/qwen", "generate openrouter/qwen", "audit failed openrouter/qwen",
		"audit sent ollama/qwen3:8b", "generate ollama/qwen3:8b", "audit ok ollama/qwen3:8b",
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
	_, err := f.generate(request(modelrouter.DataTierT0))
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
	if _, err := f.generate(request(modelrouter.DataTierT2)); !errors.Is(err, app.ErrNoAnswer) || !errors.Is(err, app.ErrProviderUnavailable) {
		t.Errorf("error = %v, want ErrNoAnswer", err)
	}
}

func TestRouterRejectionEndsTheCall(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.providers.answers["openrouter/qwen"] = app.ErrProviderRejected
	_, err := f.generate(request(modelrouter.DataTierT2))
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
	req := request(modelrouter.DataTierT2)
	req.ResponseSchemaJSON = `{"type":"object"}`
	res, err := f.generate(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Model != f.cloud { // the cloud entry has structured output
		t.Errorf("answered by %v", res.Model)
	}
	f = newRouterFixture(t)
	f.store.settings.Policy[modelrouter.DataTierT2] = []modelrouter.Ref{f.local} // only the model without it
	req = request(modelrouter.DataTierT2)
	req.ResponseSchemaJSON = `{"type":"object"}`
	if _, err := f.generate(req); !errors.Is(err, modelrouter.ErrNoCapableModel) {
		t.Errorf("error = %v, want ErrNoCapableModel", err)
	}
}

func TestRouterCachesEntriesUntilTheyExpire(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	for range 2 {
		if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
			t.Fatal(err)
		}
	}
	count := func() int { return strings.Count(strings.Join(f.ev.all(), ","), "list openrouter") }
	if count() != 1 {
		t.Errorf("listed openrouter %d times within the TTL, want 1", count())
	}
	f.clock.t = f.clock.t.Add(6 * time.Minute)
	if _, err := f.generate(request(modelrouter.DataTierT2)); err != nil {
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
	res, err := f.generate(request(modelrouter.DataTierT2))
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
	_, err := f.generate(request(modelrouter.DataTierT2))
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Errorf("error = %v", err)
	}
	for _, rec := range f.audit.sent {
		if strings.Contains(strings.Join([]string{string(rec.Actor.Kind), rec.Actor.ID, string(rec.Role), rec.Model.String(), rec.CallID}, " "), secret) {
			t.Errorf("a sent entry holds the question: %+v", rec)
		}
	}
	for _, rec := range f.audit.results {
		if strings.Contains(strings.Join([]string{string(rec.Actor.Kind), rec.Actor.ID, string(rec.Role), rec.Model.String(), rec.CallID, string(rec.Outcome), string(rec.Class)}, " "), secret) {
			t.Errorf("a result entry holds the question: %+v", rec)
		}
	}
}
