package connectapi_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
)

// The synthetic caller tokens and origin.
const (
	viewCaller   = "view-plugin-token"
	brokenCaller = "token-the-reviewer-cannot-check"
	local        = "http://localhost:8080"
)

var errReviewer = errors.New("token reviewer unavailable")

// callers accepts viewCaller as the view plugin "ui".
type callers struct{}

func (callers) VerifyCaller(_ context.Context, token string) (string, error) {
	switch token {
	case viewCaller:
		return "ui", nil
	case brokenCaller:
		return "", errReviewer
	}
	return "", connectapi.ErrCallerRefused
}

// sessionStore is the part of app.AuthStore that sessions and the demo
// authenticator use; any other method panics through the nil interface.
type sessionStore struct {
	app.AuthStore

	mu       sync.Mutex
	owner    *identity.Owner
	sessions map[identity.SessionID]identity.Session
	hashes   map[identity.SessionID][]byte
}

func (s *sessionStore) Owner(context.Context) (identity.Owner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.owner == nil {
		return identity.Owner{}, app.ErrNoOwner
	}
	return *s.owner, nil
}

func (s *sessionStore) EnsureOwner(_ context.Context, o identity.Owner) (identity.Owner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.owner == nil {
		s.owner = &o
	}
	return *s.owner, nil
}

func (*sessionStore) Passkeys(context.Context) ([]identity.Passkey, error) { return nil, nil }

func (*sessionStore) TakeCeremony(context.Context, string) (identity.Ceremony, error) {
	return identity.Ceremony{}, app.ErrUnknownCeremony
}

func (s *sessionStore) SaveSession(_ context.Context, ss *identity.Session, hash []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[ss.ID], s.hashes[ss.ID] = *ss, hash
	return nil
}

func (s *sessionStore) Session(_ context.Context, id identity.SessionID) (identity.Session, []byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ss, ok := s.sessions[id]
	if !ok {
		return identity.Session{}, nil, app.ErrUnknownSession
	}
	return ss, s.hashes[id], nil
}

func (*sessionStore) TouchSession(context.Context, identity.SessionID, time.Time) error { return nil }

func (s *sessionStore) RevokeSession(_ context.Context, id identity.SessionID, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ss := s.sessions[id]
	ss.RevokedAt = at
	s.sessions[id] = ss
	return nil
}

// marked is a demo database.
type marked struct{}

func (marked) DemoMarked(context.Context) (bool, error) { return true, nil }

// newClient serves the AuthService over a demo database with demo
// sessions on, and returns a client that presents token as the caller.
// apiClient is a generated AuthService client.
type apiClient struct {
	identityv1connect.AuthServiceClient
}

func newClient(t *testing.T, token string) apiClient {
	t.Helper()
	origins, err := identity.NewOrigins([]string{local})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	now := func() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) }
	store := &sessionStore{sessions: map[identity.SessionID]identity.Session{}, hashes: map[identity.SessionID][]byte{}}
	auth, err := app.NewAuth(store, webauthn.New("OpenFief"), nil, marked{},
		app.AuthConfig{Origins: origins, Sessions: identity.DefaultSessionPolicy(), DemoSessions: true}, now, nil)
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	path, handler := connectapi.New(auth, callers{}, slog.New(slog.DiscardHandler))
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	var opts []connect.ClientOption
	if token != "" {
		opts = append(opts, connect.WithInterceptors(bearer(token)))
	}
	return apiClient{identityv1connect.NewAuthServiceClient(srv.Client(), srv.URL, opts...)}
}

// bearer sets the Authorization header of every call.
func bearer(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// wantCode fails unless err has code.
func wantCode(t *testing.T, what string, err error, code connect.Code) {
	t.Helper()
	if connect.CodeOf(err) != code {
		t.Errorf("%s = %v, want %s", what, err, code)
	}
}

func TestCallers(t *testing.T) {
	t.Parallel()
	req := connect.NewRequest(&identityv1.GetLoginMethodsRequest{Origin: local})
	tests := []struct {
		name  string
		token string
		want  connect.Code
	}{
		{name: "no caller token", token: "", want: connect.CodePermissionDenied},
		{name: "not a view plugin", token: "connector-token", want: connect.CodePermissionDenied},
		{name: "reviewer down", token: brokenCaller, want: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := newClient(t, tt.token).GetLoginMethods(t.Context(), connect.NewRequest(&identityv1.GetLoginMethodsRequest{Origin: local}))
			wantCode(t, "GetLoginMethods", err, tt.want)
		})
	}
	got, err := newClient(t, viewCaller).GetLoginMethods(t.Context(), req)
	if err != nil || !got.Msg.GetDemoSessions() || got.Msg.GetPasskeyRegistered() {
		t.Errorf("GetLoginMethods = %v, %v; want demo sessions and no passkey", got, err)
	}
}

// demoSession creates a demo session through c and checks its shape.
func demoSession(t *testing.T, c apiClient) string {
	t.Helper()
	demo, err := c.CreateDemoSession(t.Context(), connect.NewRequest(&identityv1.CreateDemoSessionRequest{Origin: local}))
	if err != nil {
		t.Fatalf("CreateDemoSession: %v", err)
	}
	s := demo.Msg.GetSession()
	if s.GetScope() != identityv1.SessionScope_SESSION_SCOPE_READ || s.GetAuthenticator() != identityv1.Authenticator_AUTHENTICATOR_DEMO ||
		s.GetRpId() != "localhost" || s.GetExpiresAt() == nil || s.GetIdleExpiresAt() == nil {
		t.Errorf("demo session = %v", s)
	}
	who, err := c.WhoAmI(t.Context(), connect.NewRequest(&identityv1.WhoAmIRequest{Origin: local, SessionToken: demo.Msg.GetSessionToken()}))
	if err != nil || who.Msg.GetSession().GetId() != s.GetId() || who.Msg.GetOwner().GetCreatedAt() == nil {
		t.Errorf("WhoAmI = %v, %v", who, err)
	}
	return demo.Msg.GetSessionToken()
}

func TestDemoSessionOverConnect(t *testing.T) {
	t.Parallel()
	c := newClient(t, viewCaller)
	ctx := t.Context()
	_, err := c.CreateDemoSession(ctx, connect.NewRequest(&identityv1.CreateDemoSessionRequest{Origin: "https://evil.example.org"}))
	wantCode(t, "CreateDemoSession on an origin off the list", err, connect.CodePermissionDenied)
	token := demoSession(t, c)
	_, err = c.CreateScriptToken(ctx, connect.NewRequest(&identityv1.CreateScriptTokenRequest{Origin: local, SessionToken: token, Name: "x"}))
	wantCode(t, "CreateScriptToken with a read session", err, connect.CodePermissionDenied)
	_, err = c.ListPasskeys(ctx, connect.NewRequest(&identityv1.ListPasskeysRequest{Origin: local, SessionToken: token}))
	wantCode(t, "ListPasskeys with a read session", err, connect.CodePermissionDenied)
	_, err = c.CreateEnrollmentLink(ctx, connect.NewRequest(&identityv1.CreateEnrollmentLinkRequest{Origin: local, SessionToken: token}))
	wantCode(t, "CreateEnrollmentLink with a read session", err, connect.CodePermissionDenied)
	_, err = c.FinishLogin(ctx, connect.NewRequest(&identityv1.FinishLoginRequest{Origin: local, CeremonyId: "nope", CredentialJson: "{}"}))
	wantCode(t, "FinishLogin of an unknown ceremony", err, connect.CodeFailedPrecondition)
	if _, logoutErr := c.Logout(ctx, connect.NewRequest(&identityv1.LogoutRequest{Origin: local, SessionToken: token})); logoutErr != nil {
		t.Fatalf("Logout: %v", logoutErr)
	}
	_, err = c.WhoAmI(ctx, connect.NewRequest(&identityv1.WhoAmIRequest{Origin: local, SessionToken: token}))
	wantCode(t, "WhoAmI after Logout", err, connect.CodeUnauthenticated)
	_, err = c.WhoAmI(ctx, connect.NewRequest(&identityv1.WhoAmIRequest{Origin: local}))
	wantCode(t, "WhoAmI without a session", err, connect.CodeUnauthenticated)
}
