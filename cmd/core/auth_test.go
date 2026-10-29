package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
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

// uiCaller is the synthetic projected token of the tests' view plugin ui,
// which viewCluster's TokenReview accepts.
const uiCaller = "synthetic-ui-caller"

// viewCluster is a fake cluster with the view plugin ui installed and
// ready: the TokenReview API accepts uiCaller as ui's ServiceAccount, so
// the core admits it through the same check as in a real cluster, with no
// way around authentication in the binary (D-84).
func viewCluster(t *testing.T) *cluster {
	t.Helper()
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{ingestkube.PluginResource(): "PluginList"},
		&unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "openfief.io/v1alpha1", "kind": "Plugin",
			"metadata": map[string]any{"name": "ui", "namespace": "openfief"},
			"status": map[string]any{
				"conditions": []any{map[string]any{"type": "Ready", "status": "True"}},
				"effective":  map[string]any{"capabilities": []any{"view"}},
			},
		}})
	client := fake.NewClientset()
	client.PrependReactor("create", "tokenreviews", reviewTokens(map[string]string{uiCaller: "system:serviceaccount:openfief:plugin-ui"}))
	return &cluster{client: client, dynamic: dyn, namespace: "openfief"}
}

// reviewTokens is a TokenReview reactor that authenticates the tokens in
// users as their usernames and no other token.
func reviewTokens(users map[string]string) k8stesting.ReactionFunc {
	return func(a k8stesting.Action) (bool, runtime.Object, error) {
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
	}
}

// startCoreForUI is startCoreWith in viewCluster, so the view plugin ui
// may call the AuthService and the owner's entry services with uiCaller.
// The server it returns is the plugin port, where the view plugin calls
// them (D-115).
func startCoreForUI(t *testing.T, dsn string, edit func(cfg *serveConfig)) (*core, *httptest.Server, *pgxpool.Pool) {
	t.Helper()
	c, p, pool := startCorePorts(t, dsn, viewCluster(t), edit)
	return c, p.plugin, pool
}

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
func enrollCmd(t *testing.T, dsn string) string {
	t.Helper()
	cmd, err := parseArgs([]string{"enroll-link"}, func(string) string { return dsn })
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
	_, srv, _ := startCoreForUI(t, dsn, func(*serveConfig) {})
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
	_, srv, pool := startCoreForUI(t, dsn, func(cfg *serveConfig) { cfg.Auth.DemoSessions = true })
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
	_, srv, pool := startCoreForUI(t, dsn, func(*serveConfig) {})
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

func TestCallersOutsideACluster(t *testing.T) {
	t.Parallel()
	callers, err := newCallers(nil, time.Now, quiet())
	if err != nil {
		t.Fatalf("newCallers: %v", err)
	}
	for _, token := range []string{"", uiCaller, "anything"} {
		if name, err := callers.VerifyCaller(t.Context(), token); !errors.Is(err, identityconnect.ErrCallerRefused) {
			t.Errorf("VerifyCaller(%q) outside a cluster = %q, %v; want refused", token, name, err)
		}
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
	client.PrependReactor("create", "tokenreviews", reviewTokens(users))
	cl := &cluster{client: client, dynamic: dyn, namespace: "openfief"}
	callers, err := newCallers(cl, time.Now, quiet())
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
