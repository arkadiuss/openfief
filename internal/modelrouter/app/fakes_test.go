package app_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

// events is the order of everything the fakes were asked to do.
type events struct {
	mu  sync.Mutex
	log []string
}

func (e *events) add(s string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.log = append(e.log, s)
}

func (e *events) all() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return slices.Clone(e.log)
}

// fakeStore is an in-memory SettingsStore.
type fakeStore struct {
	mu       sync.Mutex
	settings modelrouter.Settings
	err      error
}

func (f *fakeStore) Settings(context.Context) (modelrouter.Settings, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.settings, f.err
}

func (f *fakeStore) SaveEnabled(_ context.Context, on bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.settings.Switch = f.settings.Switch.WithEnabled(on)
	return f.err
}

func (f *fakeStore) SaveToggle(_ context.Context, ft modelrouter.Feature, on bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	sw, err := f.settings.Switch.WithToggle(ft, on)
	if err != nil {
		return fmt.Errorf("toggling a feature: %w", err)
	}
	f.settings.Switch = sw
	return f.err
}

func (f *fakeStore) SaveRoute(_ context.Context, role modelrouter.Role, route []modelrouter.Ref) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	routes := map[modelrouter.Role]modelrouter.Route{}
	maps.Copy(routes, f.settings.Routes)
	if len(route) == 0 {
		delete(routes, role)
	} else {
		routes[role] = modelrouter.Route{Role: role, Models: route}
	}
	f.settings.Routes = routes
	return f.err
}

func (f *fakeStore) SaveDataTierModels(_ context.Context, tier modelrouter.DataTier, allowed []modelrouter.Ref) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	policy := modelrouter.DataTierPolicy{}
	maps.Copy(policy, f.settings.Policy)
	if len(allowed) == 0 {
		delete(policy, tier)
	} else {
		policy[tier] = allowed
	}
	f.settings.Policy = policy
	return f.err
}

// fakeProviders answers from scripts, by "provider/model".
type fakeProviders struct {
	ev      *events
	entries map[string][]modelrouter.ModelEntry // by provider
	answers map[string]error                    // by "provider/model": the error Generate returns; nil answers
	// instances holds the error Instance returns for a provider instance;
	// nil, or none, for an instance that runs.
	instances map[string]error
	listErr   error
	called    []string
}

func (f *fakeProviders) Instance(_ context.Context, instance string) error {
	return f.instances[instance]
}

func (f *fakeProviders) Describe(context.Context, string) (modelrouter.ProviderInstance, error) {
	return modelrouter.ProviderInstance{Location: modelrouter.LocationThirdPartyCloud}, nil
}

func (f *fakeProviders) Models(_ context.Context, provider string) ([]modelrouter.ModelEntry, error) {
	f.ev.add("list " + provider)
	return f.entries[provider], f.listErr
}

func (f *fakeProviders) Generate(_ context.Context, provider, model string, _ *modelrouter.Request) (modelrouter.Result, error) {
	key := provider + "/" + model
	f.ev.add("generate " + key)
	f.called = append(f.called, key)
	if err := f.answers[key]; err != nil {
		return modelrouter.Result{}, err
	}
	return modelrouter.Result{Content: "answer from " + key, Finish: modelrouter.FinishStop, Usage: modelrouter.Usage{InputTokens: 3, OutputTokens: 2}}, nil
}

// fakeAudit records the two entries of calls, and fails when err is set (both
// entries) or resultErr is (the result entry only).
type fakeAudit struct {
	ev        *events
	err       error
	resultErr error
	sent      []app.CallSent
	results   []app.CallResult
	// failsOnEndedContext makes RecordResult fail when it is handed a
	// context that has ended, and not one detached from it.
	failsOnEndedContext bool
}

func (f *fakeAudit) RecordSent(_ context.Context, c *app.CallSent) error {
	if f.err != nil {
		return f.err
	}
	f.ev.add("audit sent " + c.Model.String())
	f.sent = append(f.sent, *c)
	return nil
}

func (f *fakeAudit) RecordResult(ctx context.Context, c *app.CallResult) error {
	if f.failsOnEndedContext && ctx.Err() != nil {
		return fmt.Errorf("the context ended: %w", ctx.Err())
	}
	if f.err != nil {
		return f.err
	}
	if f.resultErr != nil {
		return f.resultErr
	}
	f.ev.add("audit " + string(c.Outcome) + " " + c.Model.String())
	f.results = append(f.results, *c)
	return nil
}

// fakeAttention keeps the items the router raised and cleared, by subject.
type fakeAttention struct {
	ev      *events
	err     error
	raised  map[string]string // subject to detail
	kinds   map[string]string // subject to kind
	cleared []string
}

func (f *fakeAttention) Raise(_ context.Context, kind, subject, detail string) error {
	if f.err != nil {
		return f.err
	}
	f.ev.add("raise " + subject)
	if f.raised == nil {
		f.raised, f.kinds = map[string]string{}, map[string]string{}
	}
	f.raised[subject], f.kinds[subject] = detail, kind
	return nil
}

func (f *fakeAttention) Clear(_ context.Context, _, subject string) error {
	f.cleared = append(f.cleared, subject)
	return nil
}

// callIDs returns IDs of 16 hex digits that count up from 1.
func callIDs() app.CallIDs {
	n := 0
	return func() (string, error) {
		n++
		return fmt.Sprintf("%016x", n), nil
	}
}

// fakeOwners accepts one session token as the owner's.
type fakeOwners struct{ token string }

func (f fakeOwners) CheckOwner(_ context.Context, _, token string) error {
	switch token {
	case f.token:
		return nil
	case "read":
		return app.ErrNotOwner
	}
	return app.ErrNoSession
}

var errBoom = errors.New("boom")

// clock is a settable clock.
type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }
