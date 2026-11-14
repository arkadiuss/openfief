package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/identity"
	ingestkube "github.com/arkadiuss/openfief/internal/ingest/kube"
	modelsdomain "github.com/arkadiuss/openfief/internal/modelrouter"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
)

// errLogDown is the error of an audit log that is down.
var errLogDown = errors.New("log down")

// recordingLog keeps the audit entries it is given, or fails.
type recordingLog struct {
	entries []identity.AuditEntry
	err     error
}

func (r *recordingLog) AppendAudit(_ context.Context, e *identity.AuditEntry) error {
	if r.err != nil {
		return r.err
	}
	r.entries = append(r.entries, *e)
	return nil
}

// Enqueue keeps e as the outbox would enqueue it, or fails.
func (r *recordingLog) Enqueue(ctx context.Context, e *identity.AuditEntry) error {
	return r.AppendAudit(ctx, e)
}

// TestModelAuditEntersAViewersQuestionAsTheViewers checks that the call a
// viewer's question makes is entered as the viewer's, by the session's ID, and
// never as the owner's (D-211).
func TestModelAuditEntersAViewersQuestionAsTheViewers(t *testing.T) {
	t.Parallel()
	const session = "0123456789abcdef"
	log := &recordingLog{}
	sent := modelsapp.CallSent{
		At: time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC), Actor: modelsapp.Actor{Kind: modelsapp.ActorViewer, ID: session},
		Role: modelsdomain.RoleAsk, Model: modelsdomain.Ref{Instance: "openrouter", Model: "vendor/synthetic-120b"}, CallID: "00112233445566aa",
	}
	if err := (modelAudit{log: log}).RecordSent(t.Context(), &sent); err != nil {
		t.Fatalf("RecordSent: %v", err)
	}
	if len(log.entries) != 1 || log.entries[0].Actor != (identity.Actor{Kind: identity.ActorViewer, ID: session}) {
		t.Errorf("entries = %+v, want one by the viewer, never recorded as the owner", log.entries)
	}
}

// TestModelAuditRecordsCallsAsAuditEntries checks that a model call becomes
// valid entries of the audit log (D-211): model_call before it, outcome sent,
// and model_call_result after, linked by the call ID, with the actor, the
// role, the model by name, an error class for a failure, and nothing else.
func TestModelAuditRecordsCallsAsAuditEntries(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC)
	session := "0123456789abcdef"
	const call = "00112233445566aa"
	model := modelsdomain.Ref{Instance: "openrouter", Model: "vendor/synthetic-120b"}
	system := identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
	owner := modelsapp.Actor{Kind: modelsapp.ActorOwner, ID: session}
	core := modelsapp.Actor{Kind: modelsapp.ActorSystem, ID: "core"}
	roleAsk := identity.Ref{Kind: identity.RefModelRole, ID: "ask"}
	roleMaintainer := identity.Ref{Kind: identity.RefModelRole, ID: "maintainer"}
	modelRef := identity.Ref{Kind: identity.RefModel, ID: model.String()}
	callRef := identity.Ref{Kind: identity.RefModelCall, ID: call}
	t.Run("sent", func(t *testing.T) {
		t.Parallel()
		log := &recordingLog{}
		sent := modelsapp.CallSent{At: at, Actor: owner, Role: modelsdomain.RoleAsk, Model: model, CallID: call}
		if err := (modelAudit{log: log}).RecordSent(t.Context(), &sent); err != nil {
			t.Fatalf("RecordSent: %v", err)
		}
		want := identity.AuditEntry{
			At: at, Actor: identity.Actor{Kind: identity.ActorOwner, ID: session}, Action: identity.ActionModelCall,
			Outcome: identity.OutcomeSent, Refs: []identity.Ref{roleAsk, modelRef, callRef},
		}
		if diff := cmp.Diff([]identity.AuditEntry{want}, log.entries); diff != "" {
			t.Errorf("entries (-want +got):\n%s", diff)
		}
	})
	tests := []struct {
		name string
		res  modelsapp.CallResult
		want identity.AuditEntry
	}{
		{
			name: "an answered call",
			res:  modelsapp.CallResult{At: at, Actor: owner, Role: modelsdomain.RoleAsk, Model: model, CallID: call, Outcome: modelsapp.OutcomeOK},
			want: identity.AuditEntry{
				At: at, Actor: identity.Actor{Kind: identity.ActorOwner, ID: session}, Action: identity.ActionModelCallResult,
				Outcome: identity.OutcomeOK, Refs: []identity.Ref{roleAsk, modelRef, callRef},
			},
		},
		{
			name: "a viewer's call is the viewer's, not the owner's",
			res: modelsapp.CallResult{
				At: at, Actor: modelsapp.Actor{Kind: modelsapp.ActorViewer, ID: session}, Role: modelsdomain.RoleAsk, Model: model, CallID: call,
				Outcome: modelsapp.OutcomeOK,
			},
			want: identity.AuditEntry{
				At: at, Actor: identity.Actor{Kind: identity.ActorViewer, ID: session}, Action: identity.ActionModelCallResult,
				Outcome: identity.OutcomeOK, Refs: []identity.Ref{roleAsk, modelRef, callRef},
			},
		},
		{
			name: "a refusal names the role and the call only",
			res:  modelsapp.CallResult{At: at, Actor: core, Role: modelsdomain.RoleMaintainer, CallID: call, Outcome: modelsapp.OutcomeRefused},
			want: identity.AuditEntry{
				At: at, Actor: system, Action: identity.ActionModelCallResult,
				Outcome: identity.OutcomeRefused, Refs: []identity.Ref{roleMaintainer, callRef},
			},
		},
		{
			name: "a failed call names its error class",
			res: modelsapp.CallResult{
				At: at, Actor: core, Role: modelsdomain.RoleMaintainer, Model: model, CallID: call,
				Outcome: modelsapp.OutcomeFailed, Class: modelsapp.ClassMisconfigured,
			},
			want: identity.AuditEntry{
				At: at, Actor: system, Action: identity.ActionModelCallResult, Outcome: identity.OutcomeFailed,
				Refs: []identity.Ref{roleMaintainer, modelRef, callRef, {Kind: identity.RefModelError, ID: "misconfigured"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			log := &recordingLog{}
			if err := (modelAudit{log: log}).RecordResult(t.Context(), &tt.res); err != nil {
				t.Fatalf("RecordResult: %v", err)
			}
			if diff := cmp.Diff([]identity.AuditEntry{tt.want}, log.entries); diff != "" {
				t.Errorf("entries (-want +got):\n%s", diff)
			}
		})
	}
}

// TestModelAuditFailsWhenTheLogDoes checks that a call whose entry can't be
// written is an error, which the router answers by refusing the call.
func TestModelAuditFailsWhenTheLogDoes(t *testing.T) {
	t.Parallel()
	sent := modelsapp.CallSent{
		At: time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC), Actor: modelsapp.Actor{Kind: modelsapp.ActorSystem, ID: "core"},
		Role: modelsdomain.RoleAsk, Model: modelsdomain.Ref{Instance: "p", Model: "m"}, CallID: "00112233445566aa",
	}
	if err := (modelAudit{log: &recordingLog{err: errLogDown}}).RecordSent(t.Context(), &sent); !errors.Is(err, errLogDown) {
		t.Errorf("RecordSent = %v, want the log's error", err)
	}
	// An owner without a session ID is no valid entry either.
	sent.Actor = modelsapp.Actor{Kind: modelsapp.ActorOwner}
	if err := (modelAudit{log: &recordingLog{}}).RecordSent(t.Context(), &sent); !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Errorf("RecordSent of an owner without a session = %v, want ErrInvalidAuditEntry", err)
	}
	// So is a result that names neither its call nor a valid outcome.
	res := modelsapp.CallResult{At: sent.At, Actor: modelsapp.Actor{Kind: modelsapp.ActorSystem, ID: "core"}, Role: modelsdomain.RoleAsk, Outcome: "sent"}
	if err := (modelAudit{log: &recordingLog{}}).RecordResult(t.Context(), &res); !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Errorf("RecordResult with outcome sent = %v, want ErrInvalidAuditEntry", err)
	}
	res.Outcome = modelsapp.OutcomeFailed
	if err := (modelAudit{log: &recordingLog{}}).RecordResult(t.Context(), &res); !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Errorf("RecordResult of a failure without a class = %v, want ErrInvalidAuditEntry", err)
	}
}

// TestNewCallID checks that call IDs are the 16 hex digits the audit log
// takes, and differ.
func TestNewCallID(t *testing.T) {
	t.Parallel()
	a, err := newCallID()
	if err != nil {
		t.Fatal(err)
	}
	b, err := newCallID()
	if err != nil {
		t.Fatal(err)
	}
	ref := identity.Ref{Kind: identity.RefModelCall, ID: a}
	if err := ref.Validate(); err != nil || a == b {
		t.Errorf("call IDs %q, %q: Validate = %v", a, b, err)
	}
}

// TestModelProvidersBeforeDiscovery checks that until discovery has found
// the plugins, no provider is reachable: every call is unavailable, so the
// router tries the route's next model or answers that none did.
func TestModelProvidersBeforeDiscovery(t *testing.T) {
	t.Parallel()
	p := &modelProviderInstances{}
	req := &modelsdomain.Request{}
	if _, err := p.Describe(t.Context(), "llm"); !errors.Is(err, modelsapp.ErrProviderUnavailable) {
		t.Errorf("Describe = %v, want ErrProviderUnavailable", err)
	}
	if _, err := p.Models(t.Context(), "llm"); !errors.Is(err, modelsapp.ErrProviderUnavailable) {
		t.Errorf("Models = %v, want ErrProviderUnavailable", err)
	}
	if _, err := p.Generate(t.Context(), "llm", "m", req); !errors.Is(err, modelsapp.ErrProviderUnavailable) {
		t.Errorf("Generate = %v, want ErrProviderUnavailable", err)
	}
	// The route check can't tell a missing instance from a plugin not found
	// yet, so it skips its round instead of clearing items.
	if err := p.Instance(t.Context(), "llm"); !errors.Is(err, modelsapp.ErrPluginsNotFound) || !errors.Is(err, modelsapp.ErrProviderUnavailable) {
		t.Errorf("Instance = %v, want ErrPluginsNotFound, unavailable", err)
	}
}

// TestProviderAddressNamesTheDiscoverysErrors checks that a name no plugin
// has and a plugin that is no model provider reach the router as a broken
// route, and that a provider that does not run now, or anything else, stays
// an unavailable one.
func TestProviderAddressNamesTheDiscoverysErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		err     error
		want    error
		wantNot error
	}{
		{name: "it runs"},
		{name: "no plugin has the name", err: fmt.Errorf("%w: x", ingestkube.ErrNoSuchInstance), want: modelsapp.ErrProviderMissing, wantNot: modelsapp.ErrProviderUnavailable},
		{name: "a connector", err: fmt.Errorf("%w: x", ingestkube.ErrNotModelProvider), want: modelsapp.ErrNotModelProvider, wantNot: modelsapp.ErrProviderUnavailable},
		{name: "not running", err: fmt.Errorf("%w: x", ingestkube.ErrNotRunning), wantNot: modelsapp.ErrProviderMissing},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := providerAddress(func(string) (string, error) { return "http://192.0.2.1:8080", tt.err })("x")
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want %v", err, tt.want)
			}
			if tt.wantNot != nil && errors.Is(err, tt.wantNot) {
				t.Errorf("error = %v, must not be %v", err, tt.wantNot)
			}
			if tt.err == nil && err != nil {
				t.Errorf("error = %v, want none", err)
			}
		})
	}
}

// instancesFunc is the ProviderInstances of a function that says whether an
// instance runs; the other methods are never called.
type instancesFunc func(instance string) error

func (f instancesFunc) Instance(_ context.Context, instance string) error { return f(instance) }

func (instancesFunc) Describe(context.Context, string) (modelsdomain.ProviderInstance, error) {
	panic("not called")
}

func (instancesFunc) Models(context.Context, string) ([]modelsdomain.ModelEntry, error) {
	panic("not called")
}

func (instancesFunc) Generate(context.Context, string, string, *modelsdomain.Request) (modelsdomain.Result, error) {
	panic("not called")
}

// fakeRouteStore is a models settings store with one route and the switch on.
type fakeRouteStore struct {
	modelsapp.SettingsStore

	settings modelsdomain.Settings
}

func (f fakeRouteStore) Settings(context.Context) (modelsdomain.Settings, error) {
	return f.settings, nil
}

// TestRouteCheckFindsBrokenEntries checks the attention check of the model
// routes (D-196): the kind is one the Inbox puts among the decisions, each
// broken route entry is one condition naming the route and the instance, and
// a provider instance that runs is none.
func TestRouteCheckFindsBrokenEntries(t *testing.T) {
	t.Parallel()
	missing := modelsdomain.Ref{Instance: "gone", Model: "m"}
	fine := modelsdomain.Ref{Instance: "llm", Model: "m"}
	route, err := modelsdomain.NewRoute(modelsdomain.RoleAsk, []modelsdomain.Ref{missing, fine})
	if err != nil {
		t.Fatal(err)
	}
	store := fakeRouteStore{settings: modelsdomain.Settings{
		Switch: modelsdomain.Switch{Enabled: true, Toggles: map[modelsdomain.Feature]bool{modelsdomain.FeatureAsk: true}},
		Routes: map[modelsdomain.Role]modelsdomain.Route{modelsdomain.RoleAsk: route},
	}}
	check := routeCheckOver(store, instancesFunc(func(instance string) error {
		if instance == "gone" {
			return modelsapp.ErrProviderMissing
		}
		return nil
	}))
	if check.Kind != "model-route-broken" {
		t.Errorf("the check's kind = %q, want the one the attention context groups and the UI names", check.Kind)
	}
	if got := attentiondomain.GroupOf(check.Kind); got != attentiondomain.GroupDecisions {
		t.Errorf("the check's kind %s is in group %s, want the decisions", check.Kind, got)
	}
	got, err := check.Find(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	want := []attentiondomain.Condition{{
		Kind: check.Kind, Subject: "ask/gone/m", Detail: "the ask route names provider instance gone: no plugin has that name",
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("conditions (-want +got):\n%s", diff)
	}
}
