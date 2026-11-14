package identity_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Synthetic IDs.
const (
	auditSession = "0123456789abcdef"
	auditTarget  = "fedcba9876543210"
	auditCode    = "00112233445566ff"
	callID       = "00112233445566aa"
	auditHandle  = "dGVzdC1jcmVkZW50aWFs" // base64url of "test-credential"
)

func auditAt() time.Time { return time.Date(2026, 11, 12, 9, 30, 0, 0, time.UTC) }

func TestNewAuditEntry(t *testing.T) {
	t.Parallel()
	owner := identity.Actor{Kind: identity.ActorOwner, ID: auditSession}
	tests := []struct {
		name    string
		actor   identity.Actor
		action  identity.AuditAction
		outcome identity.AuditOutcome
		refs    []identity.Ref
		wantErr bool
	}{
		{
			name: "sign-in", actor: owner, action: identity.ActionSignIn, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefSession, ID: auditSession}},
		},
		{
			name: "refused sign-in by an anonymous caller", actor: identity.Actor{Kind: identity.ActorAnonymous},
			action: identity.ActionSignIn, outcome: identity.OutcomeRefused,
		},
		{
			name: "script token created", actor: owner, action: identity.ActionScriptTokenCreated, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefScriptToken, ID: auditTarget}},
		},
		{
			name: "passkey added through an enrollment code", actor: identity.Actor{Kind: identity.ActorEnrollmentCode, ID: auditCode},
			action: identity.ActionEnrollmentLinkUsed, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefEnrollmentCode, ID: auditCode}, {Kind: identity.RefPasskey, ID: auditHandle}},
		},
		{
			name: "link issued by the core", actor: identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID},
			action: identity.ActionEnrollmentLinkIssued, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefEnrollmentCode, ID: auditCode}},
		},
		{
			name: "the entry before a model call", actor: owner, action: identity.ActionModelCall, outcome: identity.OutcomeSent,
			refs: []identity.Ref{
				{Kind: identity.RefModelRole, ID: "ask"},
				{Kind: identity.RefModel, ID: "openrouter/openai/gpt-oss-120b"},
				{Kind: identity.RefModelCall, ID: callID},
			},
		},
		{
			name: "the result of a call that was answered", actor: owner, action: identity.ActionModelCallResult, outcome: identity.OutcomeOK,
			refs: []identity.Ref{
				{Kind: identity.RefModelRole, ID: "ask"},
				{Kind: identity.RefModel, ID: "openrouter/openai/gpt-oss-120b"},
				{Kind: identity.RefModelCall, ID: callID},
			},
		},
		{
			name: "the result of a call of the system that failed", actor: identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID},
			action: identity.ActionModelCallResult, outcome: identity.OutcomeFailed,
			refs: []identity.Ref{
				{Kind: identity.RefModelRole, ID: "maintainer"},
				{Kind: identity.RefModel, ID: "ollama/qwen3:8b"},
				{Kind: identity.RefModelCall, ID: callID},
				{Kind: identity.RefModelError, ID: "unavailable"},
			},
		},
		{
			name: "a refusal before a model was chosen names the role and the call", actor: owner,
			action: identity.ActionModelCallResult, outcome: identity.OutcomeRefused,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}, {Kind: identity.RefModelCall, ID: callID}},
		},
		{
			name: "model reference that is a sentence", actor: owner, action: identity.ActionModelCall, outcome: identity.OutcomeSent,
			refs: []identity.Ref{
				{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModel, ID: "what is my net worth/x"}, {Kind: identity.RefModelCall, ID: callID},
			}, wantErr: true,
		},
		{
			name: "model reference without a provider", actor: owner, action: identity.ActionModelCall, outcome: identity.OutcomeSent,
			refs: []identity.Ref{
				{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModel, ID: "gpt-oss"}, {Kind: identity.RefModelCall, ID: callID},
			}, wantErr: true,
		},
		{
			name: "model role that is not a role", actor: owner, action: identity.ActionModelCallResult, outcome: identity.OutcomeRefused,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "my question"}, {Kind: identity.RefModelCall, ID: callID}}, wantErr: true,
		},
		{
			name: "a call ID that is a sentence", actor: owner, action: identity.ActionModelCallResult, outcome: identity.OutcomeRefused,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModelCall, ID: "what is my net worth"}}, wantErr: true,
		},
		{
			name: "an error class that is the provider's message", actor: owner, action: identity.ActionModelCallResult, outcome: identity.OutcomeFailed,
			refs: []identity.Ref{
				{Kind: identity.RefModelRole, ID: "ask"},
				{Kind: identity.RefModel, ID: "ollama/qwen3:8b"},
				{Kind: identity.RefModelCall, ID: callID},
				{Kind: identity.RefModelError, ID: "rate limited: key sk-123"},
			}, wantErr: true,
		},
		{
			name: "the entry before a call without its call ID", actor: owner, action: identity.ActionModelCall, outcome: identity.OutcomeSent,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModel, ID: "ollama/qwen3:8b"}}, wantErr: true,
		},
		{
			name: "the entry before a call that is not sent", actor: owner, action: identity.ActionModelCall, outcome: identity.OutcomeOK,
			refs: []identity.Ref{
				{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModel, ID: "ollama/qwen3:8b"}, {Kind: identity.RefModelCall, ID: callID},
			}, wantErr: true,
		},
		{
			name: "the entry before a call without a model", actor: owner, action: identity.ActionModelCall, outcome: identity.OutcomeSent,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModelCall, ID: callID}}, wantErr: true,
		},
		{
			name: "a result that is sent", actor: owner, action: identity.ActionModelCallResult, outcome: identity.OutcomeSent,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModel, ID: "ollama/qwen3:8b"}, {Kind: identity.RefModelCall, ID: callID}}, wantErr: true,
		},
		{
			name: "an answered result without its model", actor: owner, action: identity.ActionModelCallResult, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModelCall, ID: callID}}, wantErr: true,
		},
		{
			name: "a failure without its error class", actor: owner, action: identity.ActionModelCallResult, outcome: identity.OutcomeFailed,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefModel, ID: "ollama/qwen3:8b"}, {Kind: identity.RefModelCall, ID: callID}}, wantErr: true,
		},
		{
			name: "an error class on an answered call", actor: owner, action: identity.ActionModelCallResult, outcome: identity.OutcomeOK,
			refs: []identity.Ref{
				{Kind: identity.RefModelRole, ID: "ask"},
				{Kind: identity.RefModel, ID: "ollama/qwen3:8b"},
				{Kind: identity.RefModelCall, ID: callID},
				{Kind: identity.RefModelError, ID: "internal"},
			}, wantErr: true,
		},
		{name: "sign out everywhere names no record", actor: owner, action: identity.ActionSignOutEverywhere, outcome: identity.OutcomeOK},
		{
			name: "failed outcome", actor: owner, action: identity.ActionPasskeyRemoved, outcome: identity.OutcomeFailed,
			refs: []identity.Ref{{Kind: identity.RefPasskey, ID: auditHandle}},
		},

		{
			name: "a maintainer run that left a note", actor: identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID},
			action: identity.ActionMaintainerRun, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}, {Kind: identity.RefMaintainerRun, ID: callID}},
		},
		{
			name: "a maintainer run that was refused, with its class", actor: owner,
			action: identity.ActionMaintainerRun, outcome: identity.OutcomeRefused,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}, {Kind: identity.RefMaintainerRun, ID: callID}, {Kind: identity.RefModelError, ID: "denied"}},
		},
		{
			name: "a maintainer run that failed on invalid output", actor: owner,
			action: identity.ActionMaintainerRun, outcome: identity.OutcomeFailed,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}, {Kind: identity.RefMaintainerRun, ID: callID}, {Kind: identity.RefModelError, ID: "invalid_output"}},
		},
		{
			name: "a proposal dropped", actor: identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID},
			action: identity.ActionMaintainerProposalDropped, outcome: identity.OutcomeRefused,
			refs: []identity.Ref{{Kind: identity.RefMaintainerRun, ID: callID}},
		},
		{
			name: "a proposal applied", actor: owner, action: identity.ActionMaintainerProposalApplied, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefMaintainerRun, ID: callID}},
		},
		{
			name: "a proposal applied but stale", actor: owner, action: identity.ActionMaintainerProposalApplied, outcome: identity.OutcomeRefused,
			refs: []identity.Ref{{Kind: identity.RefMaintainerRun, ID: callID}},
		},
		{
			name: "a proposal dismissed", actor: owner, action: identity.ActionMaintainerProposalDismissed, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefMaintainerRun, ID: callID}},
		},
		{
			name: "a maintainer run refused without its class", actor: owner, action: identity.ActionMaintainerRun, outcome: identity.OutcomeRefused,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}, {Kind: identity.RefMaintainerRun, ID: callID}}, wantErr: true,
		},
		{
			name: "a maintainer run that is ok with a class", actor: owner, action: identity.ActionMaintainerRun, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}, {Kind: identity.RefMaintainerRun, ID: callID}, {Kind: identity.RefModelError, ID: "denied"}}, wantErr: true,
		},
		{
			name: "a maintainer run for the ask role", actor: owner, action: identity.ActionMaintainerRun, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "ask"}, {Kind: identity.RefMaintainerRun, ID: callID}}, wantErr: true,
		},
		{
			name: "a maintainer run that is sent", actor: owner, action: identity.ActionMaintainerRun, outcome: identity.OutcomeSent,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}, {Kind: identity.RefMaintainerRun, ID: callID}}, wantErr: true,
		},
		{
			name: "a maintainer run without its run", actor: owner, action: identity.ActionMaintainerRun, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelRole, ID: "maintainer"}}, wantErr: true,
		},
		{
			name: "a maintainer run named by a sentence", actor: owner, action: identity.ActionMaintainerProposalDismissed, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefMaintainerRun, ID: "trim ETFA to 55"}}, wantErr: true,
		},
		{
			name: "a proposal dropped as ok", actor: owner, action: identity.ActionMaintainerProposalDropped, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefMaintainerRun, ID: callID}}, wantErr: true,
		},
		{
			name: "a proposal dismissed as failed", actor: owner, action: identity.ActionMaintainerProposalDismissed, outcome: identity.OutcomeFailed,
			refs: []identity.Ref{{Kind: identity.RefMaintainerRun, ID: callID}}, wantErr: true,
		},
		{
			name: "a proposal applied that also names a model", actor: owner, action: identity.ActionMaintainerProposalApplied, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefMaintainerRun, ID: callID}, {Kind: identity.RefModelRole, ID: "maintainer"}}, wantErr: true,
		},
		{
			name: "a models setting changed: switch:on", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "switch:on"}},
		},
		{
			name: "a models setting changed: switch:off", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "switch:off"}},
		},
		{
			name: "a models setting changed: feature:ask:on", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "feature:ask:on"}},
		},
		{
			name: "a models setting changed: feature:mcp:off", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "feature:mcp:off"}},
		},
		{
			name: "a models setting changed: route:maintainer", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "route:maintainer"}},
		},
		{
			name: "a models setting changed: tier:TF", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "tier:TF"}},
		},
		{
			name: "a models setting named wrongly: switch:maybe", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "switch:maybe"}}, wantErr: true,
		},
		{
			name: "a models setting named wrongly: feature:trade:on", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "feature:trade:on"}}, wantErr: true,
		},
		{
			name: "a models setting named wrongly: feature:ask", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "feature:ask"}}, wantErr: true,
		},
		{
			name: "a models setting named wrongly: route:trader", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "route:trader"}}, wantErr: true,
		},
		{
			name: "a models setting named wrongly: tier:T9", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "tier:T9"}}, wantErr: true,
		},
		{
			name: "a models setting named wrongly: route:ask with openrouter/gpt", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "route:ask with openrouter/gpt"}}, wantErr: true,
		},
		{
			name: "a settings change that failed", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeFailed,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "switch:on"}}, wantErr: true,
		},
		{
			name: "a settings change naming no setting", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK, wantErr: true,
		},
		{
			name: "a settings change naming a model besides", actor: owner, action: identity.ActionModelSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: "route:ask"}, {Kind: identity.RefModel, ID: "p/m"}}, wantErr: true,
		},
		{name: "unknown action", actor: owner, action: "transfer 100 PLN", outcome: identity.OutcomeOK, wantErr: true},
		{name: "empty action", actor: owner, outcome: identity.OutcomeOK, wantErr: true},
		{name: "unknown outcome", actor: owner, action: identity.ActionSignOut, outcome: "ok, with a note", wantErr: true},
		{name: "no outcome", actor: owner, action: identity.ActionSignOut, wantErr: true},
		{
			name: "unknown actor kind", actor: identity.Actor{Kind: "guest", ID: auditSession},
			action: identity.ActionSignOut, outcome: identity.OutcomeOK, wantErr: true,
		},
		{
			name: "owner named by a name", actor: identity.Actor{Kind: identity.ActorOwner, ID: "Jan Kowalski"},
			action: identity.ActionSignOut, outcome: identity.OutcomeOK, wantErr: true,
		},
		{
			name: "owner named by an email address", actor: identity.Actor{Kind: identity.ActorOwner, ID: "jan@example.org"},
			action: identity.ActionSignOut, outcome: identity.OutcomeOK, wantErr: true,
		},
		{
			name: "anonymous with an ID", actor: identity.Actor{Kind: identity.ActorAnonymous, ID: auditSession},
			action: identity.ActionSignIn, outcome: identity.OutcomeRefused, wantErr: true,
		},
		{
			name: "system with another ID", actor: identity.Actor{Kind: identity.ActorSystem, ID: "cron"},
			action: identity.ActionSignIn, outcome: identity.OutcomeOK, wantErr: true,
		},
		{
			name: "owner without an ID", actor: identity.Actor{Kind: identity.ActorOwner},
			action: identity.ActionSignOut, outcome: identity.OutcomeOK, wantErr: true,
		},
		{
			name: "reference kind the action does not name", actor: owner, action: identity.ActionSignOut, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefScriptToken, ID: auditTarget}}, wantErr: true,
		},
		{
			name: "unknown reference kind", actor: owner, action: identity.ActionSignOut, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: "account", ID: auditSession}}, wantErr: true,
		},
		{
			name: "reference that is a sentence", actor: owner, action: identity.ActionScriptTokenCreated, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefScriptToken, ID: "my broker token"}}, wantErr: true,
		},
		{
			name: "reference twice", actor: owner, action: identity.ActionScriptTokenRevoked, outcome: identity.OutcomeOK,
			refs: []identity.Ref{{Kind: identity.RefScriptToken, ID: auditTarget}, {Kind: identity.RefScriptToken, ID: auditTarget}}, wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := identity.NewAuditEntry(auditAt(), tc.actor, tc.action, tc.outcome, tc.refs...)
			if tc.wantErr {
				if !errors.Is(err, identity.ErrInvalidAuditEntry) {
					t.Fatalf("NewAuditEntry() error = %v, want %v", err, identity.ErrInvalidAuditEntry)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewAuditEntry() error = %v", err)
			}
			want := identity.AuditEntry{At: auditAt(), Actor: tc.actor, Action: tc.action, Outcome: tc.outcome, Refs: tc.refs}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("NewAuditEntry() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// rejectedValues are the secrets and texts GR-12 keeps out of the log.
func rejectedValues() map[string]string {
	key := strings.Repeat("A", 43)
	return map[string]string{
		"script token secret":  identity.SecretPrefix + auditTarget + "_" + key,
		"session token":        identity.SessionPrefix + auditSession + "_" + key,
		"enrollment code":      identity.EnrollmentPrefix + auditCode + "_" + key,
		"amount":               "1234.56 PLN",
		"question text":        "Should I sell my bonds?",
		"broker account":       "PL61109010140000071219812874",
		"hex key of 64 digits": strings.Repeat("ab", 32),
		"long text":            strings.Repeat("x", 5000),
		"newline":              auditSession + "\n",
		"empty":                "",
	}
}

// checkRefRejected fails unless a reference of kind with ID v is refused
// without the error quoting v.
func checkRefRejected(t *testing.T, kind identity.RefKind, name, v string) {
	t.Helper()
	if kind == identity.RefPasskey && (name == "broker account" || name == "hex key of 64 digits") {
		// A credential ID is any base64url text: only the callers' types, never the shape, keep these out.
		return
	}
	err := identity.Ref{Kind: kind, ID: v}.Validate()
	if !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Errorf("Ref{%s}.Validate() = %v, want %v", kind, err, identity.ErrInvalidAuditEntry)
	} else if len(v) > 8 && strings.Contains(err.Error(), v) {
		t.Errorf("the error quotes the rejected value: %v", err)
	}
}

// TestAuditEntryRejectsSecretsAndFreeText feeds every kind of secret and text
// GR-12 keeps out of the log into each ID position and checks that none gets
// in and the error does not quote it.
func TestAuditEntryRejectsSecretsAndFreeText(t *testing.T) {
	t.Parallel()
	kinds := []identity.RefKind{identity.RefSession, identity.RefPasskey, identity.RefEnrollmentCode, identity.RefScriptToken, identity.RefModelRole, identity.RefModel, identity.RefModelCall, identity.RefModelError}
	for name, v := range rejectedValues() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, kind := range kinds {
				checkRefRejected(t, kind, name, v)
			}
			a := identity.Actor{Kind: identity.ActorOwner, ID: v}
			if err := a.Validate(); !errors.Is(err, identity.ErrInvalidAuditEntry) {
				t.Errorf("Actor.Validate() = %v, want %v", err, identity.ErrInvalidAuditEntry)
			}
		})
	}
}

func TestAuditActionsAreClosed(t *testing.T) {
	t.Parallel()
	for _, a := range identity.AuditActions() {
		got, err := identity.ParseAuditAction(string(a))
		if err != nil || got != a {
			t.Errorf("ParseAuditAction(%q) = %q, %v, want it back", a, got, err)
		}
	}
	for _, s := range []string{"", "SIGN_IN", "sign_in ", "grant_changed"} {
		if _, err := identity.ParseAuditAction(s); !errors.Is(err, identity.ErrInvalidAuditEntry) {
			t.Errorf("ParseAuditAction(%q) error = %v, want %v", s, err, identity.ErrInvalidAuditEntry)
		}
	}
}

func TestAuditFilterValidate(t *testing.T) {
	t.Parallel()
	from := auditAt()
	tests := []struct {
		name    string
		f       identity.AuditFilter
		wantErr bool
	}{
		{name: "zero selects everything"},
		{name: "actor kind", f: identity.AuditFilter{ActorKind: identity.ActorOwner}},
		{name: "actor", f: identity.AuditFilter{ActorKind: identity.ActorOwner, ActorID: auditSession}},
		{name: "action", f: identity.AuditFilter{Action: identity.ActionSignIn}},
		{name: "actions", f: identity.AuditFilter{Actions: []identity.AuditAction{identity.ActionGrantCreated, identity.ActionGrantRevoked}}},
		{name: "outcome", f: identity.AuditFilter{Outcome: identity.OutcomeRefused}},
		{name: "unknown action among actions", f: identity.AuditFilter{Actions: []identity.AuditAction{identity.ActionSignIn, "other"}}, wantErr: true},
		{name: "unknown outcome", f: identity.AuditFilter{Outcome: "maybe"}, wantErr: true},
		{name: "range", f: identity.AuditFilter{From: from, To: from.Add(time.Hour)}},
		{name: "open range", f: identity.AuditFilter{From: from}},
		{name: "largest page", f: identity.AuditFilter{Limit: identity.MaxAuditPage, Before: 10}},
		{name: "actor ID without kind", f: identity.AuditFilter{ActorID: auditSession}, wantErr: true},
		{name: "bad actor ID", f: identity.AuditFilter{ActorKind: identity.ActorOwner, ActorID: "Jan"}, wantErr: true},
		{name: "unknown actor kind", f: identity.AuditFilter{ActorKind: "guest"}, wantErr: true},
		{name: "unknown action", f: identity.AuditFilter{Action: "other"}, wantErr: true},
		{name: "empty range", f: identity.AuditFilter{From: from, To: from}, wantErr: true},
		{name: "reversed range", f: identity.AuditFilter{From: from, To: from.Add(-time.Hour)}, wantErr: true},
		{name: "page too large", f: identity.AuditFilter{Limit: identity.MaxAuditPage + 1}, wantErr: true},
		{name: "negative page", f: identity.AuditFilter{Limit: -1}, wantErr: true},
		{name: "negative cursor", f: identity.AuditFilter{Before: -1}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.f.Validate()
			if tc.wantErr != errors.Is(err, identity.ErrInvalidAuditFilter) {
				t.Errorf("Validate() = %v, want an error: %v", err, tc.wantErr)
			}
		})
	}
}

func TestAuditFilterMatches(t *testing.T) {
	t.Parallel()
	owner := identity.Actor{Kind: identity.ActorOwner, ID: auditSession}
	e := identity.AuditEntry{Seq: 7, At: auditAt(), Actor: owner, Action: identity.ActionSignIn, Outcome: identity.OutcomeOK}
	tests := []struct {
		name string
		f    identity.AuditFilter
		want bool
	}{
		{name: "everything", want: true},
		{name: "same actor", f: identity.AuditFilter{ActorKind: identity.ActorOwner, ActorID: auditSession}, want: true},
		{name: "other actor kind", f: identity.AuditFilter{ActorKind: identity.ActorSystem}},
		{name: "other actor", f: identity.AuditFilter{ActorKind: identity.ActorOwner, ActorID: auditTarget}},
		{name: "same action", f: identity.AuditFilter{Action: identity.ActionSignIn}, want: true},
		{name: "other action", f: identity.AuditFilter{Action: identity.ActionSignOut}},
		{name: "one of the actions", f: identity.AuditFilter{Actions: []identity.AuditAction{identity.ActionSignOut, identity.ActionSignIn}}, want: true},
		{name: "none of the actions", f: identity.AuditFilter{Actions: []identity.AuditAction{identity.ActionSignOut, identity.ActionGrantCreated}}},
		{name: "the action and one of the actions", f: identity.AuditFilter{Action: identity.ActionSignIn, Actions: []identity.AuditAction{identity.ActionSignIn}}, want: true},
		{name: "the action but not one of the actions", f: identity.AuditFilter{Action: identity.ActionSignIn, Actions: []identity.AuditAction{identity.ActionSignOut}}},
		{name: "same outcome", f: identity.AuditFilter{Outcome: identity.OutcomeOK}, want: true},
		{name: "other outcome", f: identity.AuditFilter{Outcome: identity.OutcomeRefused}},
		{name: "from is inclusive", f: identity.AuditFilter{From: auditAt()}, want: true},
		{name: "from after", f: identity.AuditFilter{From: auditAt().Add(time.Second)}},
		{name: "to is exclusive", f: identity.AuditFilter{To: auditAt()}},
		{name: "to after", f: identity.AuditFilter{To: auditAt().Add(time.Second)}, want: true},
		{name: "cursor above", f: identity.AuditFilter{Before: 8}, want: true},
		{name: "cursor at the entry", f: identity.AuditFilter{Before: 7}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.f.Matches(&e); got != tc.want {
				t.Errorf("Matches() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAuditFilterPageSize(t *testing.T) {
	t.Parallel()
	if got := (&identity.AuditFilter{}).PageSize(); got != identity.DefaultAuditPage {
		t.Errorf("PageSize() of the zero filter = %d, want %d", got, identity.DefaultAuditPage)
	}
	if got := (&identity.AuditFilter{Limit: 5}).PageSize(); got != 5 {
		t.Errorf("PageSize() = %d, want 5", got)
	}
}
