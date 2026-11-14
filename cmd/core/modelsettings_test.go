package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
)

// TestSettingsAuditRecordsAChangeAsAnEntryNamingTheSetting checks that a
// change to the agents-and-models settings becomes a valid entry of the
// audit log (D-211): the owner's session, the action model_settings_changed,
// ok, and the one setting, never a model list; and that a setting the
// identity context doesn't know, or a log that fails, is an error, which
// makes the settings refuse the change.
func TestSettingsAuditRecordsAChangeAsAnEntryNamingTheSetting(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, time.November, 13, 12, 0, 0, 0, time.UTC)
	const session = "0123456789abcdef"
	for _, setting := range []string{"switch:on", "switch:off", "feature:maintainer:on", "feature:mcp:off", "route:ask", "tier:T0"} {
		t.Run(setting, func(t *testing.T) {
			t.Parallel()
			log := &recordingLog{}
			if err := (settingsAudit{out: log}).RecordChange(t.Context(), &modelsapp.SettingChange{At: at, Session: session, Setting: setting}); err != nil {
				t.Fatalf("RecordChange: %v", err)
			}
			want := identity.AuditEntry{
				At: at, Actor: identity.Actor{Kind: identity.ActorOwner, ID: session}, Action: identity.ActionModelSettingsChanged,
				Outcome: identity.OutcomeOK, Refs: []identity.Ref{{Kind: identity.RefModelSetting, ID: setting}},
			}
			if diff := cmp.Diff([]identity.AuditEntry{want}, log.entries); diff != "" {
				t.Errorf("entries (-want +got):\n%s", diff)
			}
		})
	}
	t.Run("a setting nobody knows", func(t *testing.T) {
		t.Parallel()
		log := &recordingLog{}
		err := (settingsAudit{out: log}).RecordChange(t.Context(), &modelsapp.SettingChange{At: at, Session: session, Setting: "route:trader"})
		if !errors.Is(err, identity.ErrInvalidAuditEntry) || len(log.entries) != 0 {
			t.Errorf("RecordChange = %v with %d entries, want ErrInvalidAuditEntry and none", err, len(log.entries))
		}
	})
}

// TestModelProvidersWaitForDiscovery checks that the settings can't list the
// model providers before discovery found the plugins, which the settings take
// as an empty list, and that once it did, the list is discovery's.
func TestModelProvidersWaitForDiscovery(t *testing.T) {
	t.Parallel()
	p := &modelProviderInstances{}
	if _, err := p.ModelProviders(t.Context()); !errors.Is(err, modelsapp.ErrPluginsNotFound) {
		t.Fatalf("before discovery: %v, want ErrPluginsNotFound", err)
	}
	p.setDirectory(func(context.Context) ([]string, error) { return []string{"cloud", "home"}, nil })
	got, err := p.ModelProviders(t.Context())
	if err != nil || cmp.Diff([]string{"cloud", "home"}, got) != "" {
		t.Errorf("ModelProviders = %v, %v", got, err)
	}
}

// TestOwnerSessionNamesTheOwnersSessionForTheAuditEntry checks that the models
// settings learn the ID of the owner's session, which their audit entries name
// as the actor, and are refused for any other session or an origin off the list.
func TestOwnerSessionNamesTheOwnersSessionForTheAuditEntry(t *testing.T) {
	t.Parallel()
	origins, err := identity.NewOrigins([]string{"http://localhost:8080"})
	if err != nil {
		t.Fatal(err)
	}
	o := &ownerSessions{auth: sessionsByToken{}, origins: origins, noSession: modelsapp.ErrNoSession, notOwner: modelsapp.ErrNotOwner}
	for _, tt := range []struct {
		name, origin, token, want string
		wantErr                   error
	}{
		{"the owner", "http://localhost:8080", "owner", "owner-id", nil},
		{"a demo session", "http://localhost:8080", "demo", "", modelsapp.ErrNotOwner},
		{"no session", "http://localhost:8080", "nope", "", modelsapp.ErrNoSession},
		{"an origin off the list", "https://other.example.com", "owner", "", modelsapp.ErrNotOwner},
	} {
		got, err := o.OwnerSession(t.Context(), tt.origin, tt.token)
		if got != tt.want || !errors.Is(err, tt.wantErr) {
			t.Errorf("%s: OwnerSession = %q, %v; want %q, %v", tt.name, got, err, tt.want, tt.wantErr)
		}
	}
}
