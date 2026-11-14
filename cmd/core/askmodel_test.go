package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	modelsdomain "github.com/arkadiuss/openfief/internal/modelrouter"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// scopeSessions knows the sessions "owner" and "read", of those scopes, and
// keeps the scope each check asked for.
type scopeSessions struct{ asked []identity.SessionScope }

func (s *scopeSessions) Session(_ context.Context, _ identity.Origin, token string, want identity.SessionScope) (identity.Session, error) {
	s.asked = append(s.asked, want)
	var have identity.SessionScope
	switch token {
	case "owner":
		have = identity.SessionScopeOwner
	case "read":
		have = identity.SessionScopeRead
	case "viewer":
		return identity.Session{ID: "viewer-id", Scope: identity.SessionScopeRead, PersonID: "0123456789abcdef0123456789abcdef"}, nil
	default:
		return identity.Session{}, fmt.Errorf("%w: unknown", identityapp.ErrInvalidSession)
	}
	if !have.Includes(want) {
		return identity.Session{}, fmt.Errorf("%w: scope", identityapp.ErrForbidden)
	}
	return identity.Session{ID: identity.SessionID(token + "-id"), Scope: have}, nil
}

// TestAskSessionsAdmitAnyValidSession checks the Ask box's session port
// (D-203): a read-scoped session may ask as an owner's may, since what the
// model sees is limited by what the session reads; a session that is unknown
// or an origin off the allow-list is no session.
func TestAskSessionsAdmitAnyValidSession(t *testing.T) {
	t.Parallel()
	origins, err := identity.NewOrigins([]string{"http://localhost:8080"})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	const local = "http://localhost:8080"
	tests := []struct {
		name, origin, token string
		want                modelsapp.Actor
		wantErr             error
	}{
		{name: "the owner", origin: local, token: "owner", want: modelsapp.Actor{Kind: modelsapp.ActorOwner, ID: "owner-id"}},
		{name: "a read-scoped session of the owner", origin: local, token: "read", want: modelsapp.Actor{Kind: modelsapp.ActorOwner, ID: "read-id"}},
		{name: "a viewer is the viewer, not the owner", origin: local, token: "viewer", want: modelsapp.Actor{Kind: modelsapp.ActorViewer, ID: "viewer-id"}},
		{name: "no session", origin: local, token: "nope", wantErr: modelsapp.ErrNoSession},
		{name: "an origin off the list", origin: "https://other.example.com", token: "owner", wantErr: modelsapp.ErrNoSession},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sessions := &scopeSessions{}
			got, err := askSessions{auth: sessions, origins: origins}.Asker(t.Context(), tt.origin, tt.token)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Asker error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Asker = %+v, want %+v", got, tt.want)
			}
			for _, scope := range sessions.asked {
				if scope != identity.SessionScopeRead {
					t.Errorf("asked for scope %v, want read: any valid session may ask", scope)
				}
			}
		})
	}
}

// ownerlessSettings is a settings store that holds only what is read.
type ownerlessSettings struct {
	modelsapp.SettingsStore

	s modelsdomain.Settings
}

func (o ownerlessSettings) Settings(context.Context) (modelsdomain.Settings, error) { return o.s, nil }

// uiCallers admits the bearer token "ui" as a view plugin.
type uiCallers struct{}

func (uiCallers) VerifyCaller(_ context.Context, token string) (string, error) {
	if token == "ui" {
		return "ui", nil
	}
	return "", identityconnect.ErrCallerRefused
}

// onlySession admits the session token "s" only.
type onlySession struct{}

func (onlySession) Asker(_ context.Context, _, token string) (modelsapp.Actor, error) {
	if token == "s" {
		return modelsapp.Actor{Kind: modelsapp.ActorOwner, ID: "session-s"}, nil
	}
	return modelsapp.Actor{}, modelsapp.ErrNoSession
}

// TestAskModelServiceIsMountedForViewPluginsOnly checks that the core serves
// the Ask box's door next to the other plugin-port services (D-203): a view
// plugin's call is answered with nothing stored, a call without the plugin's
// token is refused, and every other path goes on to the handler it was
// mounted over.
func TestAskModelServiceIsMountedForViewPluginsOnly(t *testing.T) {
	t.Parallel()
	settings := ownerlessSettings{s: modelsdomain.Settings{
		Switch: modelsdomain.Switch{Enabled: true, Toggles: map[modelsdomain.Feature]bool{modelsdomain.FeatureAsk: true}},
		Routes: map[modelsdomain.Role]modelsdomain.Route{modelsdomain.RoleAsk: {Role: modelsdomain.RoleAsk, Models: []modelsdomain.Ref{{Instance: "p", Model: "m"}}}},
	}}
	other := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	h := withAskModelService(other, modelsapp.NewAsk(onlySession{}, settings, nil), uiCallers{}, slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	call := func(token string) (*connect.Response[modelsv1.GetAskModelResponse], error) {
		c := modelsv1connect.NewAskModelServiceClient(srv.Client(), srv.URL)
		req := connect.NewRequest(&modelsv1.GetAskModelRequest{Origin: "o", SessionToken: "s"})
		if token != "" {
			req.Header().Set("Authorization", "Bearer "+token)
		}
		resp, err := c.GetAskModel(t.Context(), req)
		if err != nil {
			return nil, fmt.Errorf("GetAskModel: %w", err)
		}
		return resp, nil
	}
	resp, err := call("ui")
	if err != nil || !resp.Msg.GetAvailable() {
		t.Fatalf("a view plugin's call = %v, %v; want available", resp, err)
	}
	if got := resp.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store: nothing may keep the answer", got)
	}
	for _, token := range []string{"", "stranger"} {
		if _, err := call(token); connect.CodeOf(err) != connect.CodePermissionDenied {
			t.Errorf("a call with the token %q: code = %v, want PermissionDenied", token, connect.CodeOf(err))
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/anything", http.NoBody))
	if rec.Code != http.StatusTeapot {
		t.Errorf("another path = %d, want the handler it was mounted over", rec.Code)
	}
}
