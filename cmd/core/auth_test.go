package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	authnv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
	ingestkube "github.com/arkadiuss/openfief/internal/ingest/kube"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
)

// uiCaller is the synthetic dev caller token of the tests.
const uiCaller = "synthetic-ui-caller"

// localUI is the origin the tests' browser uses.
const localUI = "http://localhost:8080"

// authAPI is a generated AuthService client.
type authAPI struct {
	identityv1connect.AuthServiceClient
}

// authClient returns an AuthService client of srv that presents token as
// the view plugin's.
func authClient(srvURL string, client *http.Client, token string) authAPI {
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
	return authAPI{identityv1connect.NewAuthServiceClient(client, srvURL, connect.WithInterceptors(bearer))}
}

// enrollCmd runs core enroll-link as the CLI does and returns its output.
func enrollCmd(t *testing.T, dsn string, args ...string) string {
	t.Helper()
	cmd, err := parseArgs(append([]string{"enroll-link"}, args...), func(string) string { return dsn })
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	var out bytes.Buffer
	if err := run(t.Context(), &cmd, &out, quiet()); err != nil {
		t.Fatalf("core enroll-link: %v", err)
	}
	return out.String()
}

func TestPasskeyLoginEndToEnd(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	out := enrollCmd(t, dsn)
	link, _, _ := strings.Cut(out, "\n")
	code, ok := strings.CutPrefix(link, localUI+"/enroll#code=")
	if !ok {
		t.Fatalf("enroll-link printed %q", out)
	}
	_, srv, _ := startCoreWith(t, dsn, func(cfg *serveConfig) { cfg.Auth.DevCallerToken = uiCaller })
	ctx := t.Context()
	_, err := authClient(srv.URL, srv.Client(), "wrong").GetLoginMethods(ctx, connect.NewRequest(&identityv1.GetLoginMethodsRequest{Origin: localUI}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("a caller with the wrong token = %v, want permission_denied", err)
	}
	c := authClient(srv.URL, srv.Client(), uiCaller)
	dev := passkeyEnroll(t, c, code)
	session := passkeyLogin(t, c, dev)
	if r := get(t, srv, "/v1/accounts", session); r.status != http.StatusOK {
		t.Errorf("read API with a session = %d, want 200", r.status)
	}
	tok, err := c.CreateScriptToken(ctx, connect.NewRequest(&identityv1.CreateScriptTokenRequest{Origin: localUI, SessionToken: session, Name: "nightly"}))
	if err != nil {
		t.Fatalf("CreateScriptToken: %v", err)
	}
	if r := get(t, srv, "/v1/accounts", tok.Msg.GetSecret()); r.status != http.StatusOK {
		t.Errorf("read API with the new script token = %d, want 200", r.status)
	}
	if _, err := c.Logout(ctx, connect.NewRequest(&identityv1.LogoutRequest{Origin: localUI, SessionToken: session})); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	checkRefused(t, srv, session, "a revoked session")
}

// passkeyEnroll creates a passkey with the enrollment code through c and
// returns its authenticator.
func passkeyEnroll(t *testing.T, c authAPI, code string) *virtual.Authenticator {
	t.Helper()
	ctx := t.Context()
	dev, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("virtual.New: %v", err)
	}
	begun, err := c.BeginRegistration(ctx, connect.NewRequest(&identityv1.BeginRegistrationRequest{
		Origin: localUI, Authorization: &identityv1.BeginRegistrationRequest_EnrollmentCode{EnrollmentCode: code},
	}))
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	cred, err := dev.Create([]byte(begun.Msg.GetOptionsJson()), localUI)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	reg, err := c.FinishRegistration(ctx, connect.NewRequest(&identityv1.FinishRegistrationRequest{
		Origin: localUI, CeremonyId: begun.Msg.GetCeremonyId(), CredentialJson: string(cred), Name: "laptop",
	}))
	if err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}
	if reg.Msg.GetSession().GetScope() != identityv1.SessionScope_SESSION_SCOPE_OWNER || reg.Msg.GetPasskey().GetRpId() != "localhost" {
		t.Errorf("FinishRegistration = %v", reg.Msg)
	}
	return dev
}

// passkeyLogin logs in with dev through c and returns the session token.
func passkeyLogin(t *testing.T, c authAPI, dev *virtual.Authenticator) string {
	t.Helper()
	ctx := t.Context()
	begun, err := c.BeginLogin(ctx, connect.NewRequest(&identityv1.BeginLoginRequest{Origin: localUI}))
	if err != nil {
		t.Fatalf("BeginLogin: %v", err)
	}
	assertion, err := dev.Get([]byte(begun.Msg.GetOptionsJson()), localUI)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	done, err := c.FinishLogin(ctx, connect.NewRequest(&identityv1.FinishLoginRequest{
		Origin: localUI, CeremonyId: begun.Msg.GetCeremonyId(), CredentialJson: string(assertion),
	}))
	if err != nil {
		t.Fatalf("FinishLogin: %v", err)
	}
	return done.Msg.GetSessionToken()
}

func TestDemoSessionsNeedTheMarker(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	_, srv, pool := startCoreWith(t, dsn, func(cfg *serveConfig) {
		cfg.Auth.DevCallerToken = uiCaller
		cfg.Auth.DemoSessions = true
	})
	ctx := t.Context()
	c := authClient(srv.URL, srv.Client(), uiCaller)
	req := func() *connect.Request[identityv1.CreateDemoSessionRequest] {
		return connect.NewRequest(&identityv1.CreateDemoSessionRequest{Origin: localUI})
	}
	if _, err := c.CreateDemoSession(ctx, req()); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("CreateDemoSession on a database without the marker = %v, want permission_denied", err)
	}
	// What cmd/demo-seed writes.
	if _, err := pool.Exec(ctx, `CREATE SCHEMA demo;
		CREATE TABLE demo.marker (demo boolean PRIMARY KEY CHECK (demo), seed text NOT NULL, loaded_on date NOT NULL);
		INSERT INTO demo.marker VALUES (true, '1', '2026-10-28')`); err != nil {
		t.Fatalf("writing the demo marker: %v", err)
	}
	demo, err := c.CreateDemoSession(ctx, req())
	if err != nil {
		t.Fatalf("CreateDemoSession on a demo database: %v", err)
	}
	token := demo.Msg.GetSessionToken()
	if demo.Msg.GetSession().GetScope() != identityv1.SessionScope_SESSION_SCOPE_READ {
		t.Errorf("demo session scope = %v", demo.Msg.GetSession().GetScope())
	}
	if r := get(t, srv, "/v1/accounts", token); r.status != http.StatusOK {
		t.Errorf("read API with a demo session = %d, want 200", r.status)
	}
	_, err = c.CreateScriptToken(ctx, connect.NewRequest(&identityv1.CreateScriptTokenRequest{Origin: localUI, SessionToken: token, Name: "x"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("CreateScriptToken with a demo session = %v, want permission_denied", err)
	}
	_, err = c.CreateEnrollmentLink(ctx, connect.NewRequest(&identityv1.CreateEnrollmentLinkRequest{Origin: localUI, SessionToken: token}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("CreateEnrollmentLink with a demo session = %v, want permission_denied", err)
	}
}

func TestDemoSessionsOffByDefault(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	_, srv, pool := startCoreWith(t, dsn, func(cfg *serveConfig) { cfg.Auth.DevCallerToken = uiCaller })
	ctx := t.Context()
	if _, err := pool.Exec(ctx, `CREATE SCHEMA demo;
		CREATE TABLE demo.marker (demo boolean PRIMARY KEY CHECK (demo), seed text NOT NULL, loaded_on date NOT NULL);
		INSERT INTO demo.marker VALUES (true, '1', '2026-10-28')`); err != nil {
		t.Fatalf("writing the demo marker: %v", err)
	}
	_, err := authClient(srv.URL, srv.Client(), uiCaller).CreateDemoSession(ctx, connect.NewRequest(&identityv1.CreateDemoSessionRequest{Origin: localUI}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("CreateDemoSession without --demo-sessions = %v, want permission_denied", err)
	}
}

func TestDevCallers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, configured, presented string
		want                        bool
	}{
		{name: "matching token", configured: "abc", presented: "abc", want: true},
		{name: "other token", configured: "abc", presented: "abd"},
		{name: "no token configured", configured: "", presented: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			name, err := devCallers{token: tt.configured}.VerifyCaller(t.Context(), tt.presented)
			if tt.want != (err == nil) || (err != nil && !errors.Is(err, identityconnect.ErrCallerRefused)) {
				t.Errorf("VerifyCaller = %q, %v; want accepted %v", name, err, tt.want)
			}
		})
	}
}

func TestKubeCallers(t *testing.T) {
	t.Parallel()
	plugin := func(name string, caps ...any) runtime.Object {
		return &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "openfief.io/v1alpha1", "kind": "Plugin",
			"metadata": map[string]any{"name": name, "namespace": "openfief"},
			"status":   map[string]any{"effective": map[string]any{"capabilities": caps}},
		}}
	}
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{ingestkube.PluginResource(): "PluginList"},
		plugin("ui", "view"), plugin("ibkr-main", "connector"))
	client := fake.NewClientset()
	users := map[string]string{
		"ui-token": "system:serviceaccount:openfief:plugin-ui", "ibkr-token": "system:serviceaccount:openfief:plugin-ibkr-main",
		"ghost-token": "system:serviceaccount:openfief:plugin-ghost", "core-token": "system:serviceaccount:openfief:openfief-core",
	}
	client.PrependReactor("create", "tokenreviews", func(a k8stesting.Action) (bool, runtime.Object, error) {
		create, ok := a.(k8stesting.CreateAction)
		if !ok {
			return false, nil, nil
		}
		review, ok := create.GetObject().(*authnv1.TokenReview)
		if !ok {
			return false, nil, nil
		}
		user, known := users[review.Spec.Token]
		review.Status = authnv1.TokenReviewStatus{Authenticated: known, User: authnv1.UserInfo{Username: user}, Audiences: review.Spec.Audiences}
		return true, review, nil
	})
	cl := &cluster{client: client, dynamic: dyn, namespace: "openfief"}
	callers, err := newCallers(cl, &authConfig{}, time.Now, quiet())
	if err != nil {
		t.Fatalf("newCallers: %v", err)
	}
	tests := []struct {
		token, want string
		refused     bool
	}{
		{token: "ui-token", want: "ui"},
		{token: "ibkr-token", refused: true},   // a connector, not a view
		{token: "core-token", refused: true},   // not a plugin
		{token: "stolen-token", refused: true}, // not authenticated
		{token: "ghost-token", refused: true},  // no Plugin resource
	}
	for _, tt := range tests {
		got, err := callers.VerifyCaller(t.Context(), tt.token)
		if tt.refused != errors.Is(err, identityconnect.ErrCallerRefused) || got != tt.want {
			t.Errorf("VerifyCaller(%s) = %q, %v; want %q, refused %v", tt.token, got, err, tt.want, tt.refused)
		}
	}
}
