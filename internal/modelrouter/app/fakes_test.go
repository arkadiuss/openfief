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

func (f *fakeStore) SaveTierModels(_ context.Context, tier modelrouter.Tier, allowed []modelrouter.Ref) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	policy := modelrouter.TierPolicy{}
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
	entries map[string][]modelrouter.Entry // by provider
	answers map[string]error               // by "provider/model": the error Generate returns; nil answers
	listErr error
	called  []string
}

func (f *fakeProviders) Describe(context.Context, string) (modelrouter.Provider, error) {
	return modelrouter.Provider{Location: modelrouter.LocationThirdPartyCloud}, nil
}

func (f *fakeProviders) Models(_ context.Context, provider string) ([]modelrouter.Entry, error) {
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

// fakeAudit records calls, and fails when err is set.
type fakeAudit struct {
	ev   *events
	err  error
	recs []app.CallRecord
}

func (f *fakeAudit) RecordCall(_ context.Context, rec *app.CallRecord) error {
	if f.err != nil {
		return f.err
	}
	f.ev.add("audit " + string(rec.Outcome) + " " + rec.Model.String())
	f.recs = append(f.recs, *rec)
	return nil
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
