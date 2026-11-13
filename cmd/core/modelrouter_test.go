package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
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

// TestModelAuditRecordsCallsAsAuditEntries checks that a model call becomes
// a valid model_call entry of the audit log (D-211): the actor, the role and
// the model by name, the outcome, and nothing else.
func TestModelAuditRecordsCallsAsAuditEntries(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC)
	session := "0123456789abcdef"
	model := modelsdomain.Ref{Provider: "openrouter", Model: "vendor/synthetic-120b"}
	tests := []struct {
		name string
		rec  modelsapp.CallRecord
		want identity.AuditEntry
	}{
		{
			name: "an owner's call",
			rec: modelsapp.CallRecord{
				At: at, Actor: modelsapp.Actor{Kind: modelsapp.ActorOwner, ID: session},
				Role: modelsdomain.RoleAsk, Model: model, Outcome: modelsapp.OutcomeCalled,
			},
			want: identity.AuditEntry{
				At: at, Actor: identity.Actor{Kind: identity.ActorOwner, ID: session}, Action: identity.ActionModelCall,
				Outcome: identity.OutcomeOK, Refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModel, ID: model.String()}},
			},
		},
		{
			name: "a refusal names the role only",
			rec: modelsapp.CallRecord{
				At: at, Actor: modelsapp.Actor{Kind: modelsapp.ActorSystem, ID: "core"},
				Role: modelsdomain.RoleMaintainer, Outcome: modelsapp.OutcomeRefused,
			},
			want: identity.AuditEntry{
				At: at, Actor: identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}, Action: identity.ActionModelCall,
				Outcome: identity.OutcomeRefused, Refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}},
			},
		},
		{
			name: "a failed call",
			rec: modelsapp.CallRecord{
				At: at, Actor: modelsapp.Actor{Kind: modelsapp.ActorSystem, ID: "core"},
				Role: modelsdomain.RoleMaintainer, Model: model, Outcome: modelsapp.OutcomeFailed,
			},
			want: identity.AuditEntry{
				At: at, Actor: identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}, Action: identity.ActionModelCall,
				Outcome: identity.OutcomeFailed, Refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}, {Kind: identity.RefModel, ID: model.String()}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			log := &recordingLog{}
			if err := (modelAudit{log: log}).RecordCall(t.Context(), &tt.rec); err != nil {
				t.Fatalf("RecordCall: %v", err)
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
	rec := modelsapp.CallRecord{
		At: time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC), Actor: modelsapp.Actor{Kind: modelsapp.ActorSystem, ID: "core"},
		Role: modelsdomain.RoleAsk, Outcome: modelsapp.OutcomeCalled,
	}
	if err := (modelAudit{log: &recordingLog{err: errLogDown}}).RecordCall(t.Context(), &rec); !errors.Is(err, errLogDown) {
		t.Errorf("RecordCall = %v, want the log's error", err)
	}
	// An owner without a session ID is no valid entry either.
	rec.Actor = modelsapp.Actor{Kind: modelsapp.ActorOwner}
	if err := (modelAudit{log: &recordingLog{}}).RecordCall(t.Context(), &rec); !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Errorf("RecordCall of an owner without a session = %v, want ErrInvalidAuditEntry", err)
	}
}

// TestModelProvidersBeforeDiscovery checks that until discovery has found
// the plugins, no provider is reachable: every call is unavailable, so the
// router tries the route's next model or answers that none did.
func TestModelProvidersBeforeDiscovery(t *testing.T) {
	t.Parallel()
	p := &modelProviders{}
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
}
