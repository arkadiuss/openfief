//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	connectrpc "connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/plugins"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errNoCoreContainer reports a core Deployment without its container.
var errNoCoreContainer = errors.New("the core Deployment has no container named core")

// TestPolicyModeGrantsLimitTokensAndKeepThePagesWorking checks D-209, D-213 and D-165 on the core as the default
// install and the live demo run it, --authz=policy (SCEN-identity-grants-limit-tokens-under-policy): the suite
// runs under SpiceDB, so this test restarts the core without --authz and --spicedb-insecure, checks, and restores
// them. A script token granted one account views that account alone over /v1 and over /mcp (with the switch and the
// MCP toggle on) while the household's token views every account; no token is listed an attention item of an
// unmapped broker account, which is tied to no account and the owner's alone; and the built-in UI still renders the
// accounts for the owner's session and for a demo session.
//
// It is not parallel, on purpose: it restarts the core, which the other tests must not see, so Go runs it before
// any parallel test starts. The core is back on SpiceDB, with the switch off again, when it ends.
//
//nolint:paralleltest // see above: the test restarts the core
func TestPolicyModeGrantsLimitTokensAndKeepThePagesWorking(t *testing.T) {
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	waitSeedComplete(t, k)
	waitPluginReady(t, k, uiPlugin)
	waitReconciled(t, k)

	original := setCoreArgs(t, k, func(args []string) []string {
		return slices.DeleteFunc(slices.Clone(args), func(a string) bool {
			return strings.HasPrefix(a, "--authz") || strings.HasPrefix(a, "--spicedb")
		})
	})
	t.Cleanup(func() {
		setCoreArgs(t, k, func([]string) []string { return original })
		waitCoreRolledOut(t, k)
		waitSpiceDBAvailable(t, k)
		waitReconciled(t, k)
	})
	waitCoreRolledOut(t, k)

	base := forwardCore(t, k)
	owner := newOwnerAPI(t, k)
	enableMCP(t, k, owner)

	xtb := string(demo.XTBAccount)
	household := createMCPToken(t, owner, "e2e-policy-household", identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, "")
	narrow := createMCPToken(t, owner, "e2e-policy-narrow", identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, xtb)

	if got := accountsOver(t, base, household); !slices.Equal(got, demoAccountIDs()) {
		t.Fatalf("the household's token lists %v over /v1 under the policy, want every demo account", got)
	}
	if got := accountsOver(t, base, narrow); !slices.Equal(got, []string{xtb}) {
		t.Fatalf("a token granted one account lists %v over /v1 under the policy, want %s alone", got, xtb)
	}
	if r := bundleGet(t, base+"/v1/accounts/"+string(demo.IBKRAccount), narrow); r.status != http.StatusNotFound {
		t.Errorf("GET an account outside the grant = %d, want 404", r.status)
	}

	checkMCPGrants(t, base, household, narrow)

	checkUnboundItemsAreTheOwners(t, owner, base, household, narrow)

	checkPagesRender(t, owner, base)
}

// setCoreArgs replaces the args of the core container with what edit makes of them and returns the ones it had.
func setCoreArgs(t *testing.T, k *kube, edit func([]string) []string) []string {
	t.Helper()
	d, err := k.clients.AppsV1().Deployments(namespace).Get(t.Context(), coreDeployment, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting Deployment %s: %v", coreDeployment, err)
	}
	for i := range d.Spec.Template.Spec.Containers {
		c := &d.Spec.Template.Spec.Containers[i]
		if c.Name != "core" {
			continue
		}
		before := slices.Clone(c.Args)
		c.Args = edit(before)
		if _, err := k.clients.AppsV1().Deployments(namespace).Update(t.Context(), d, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("updating Deployment %s: %v", coreDeployment, err)
		}
		return before
	}
	t.Fatal(errNoCoreContainer)
	return nil
}

// enableMCP turns the one switch for agents and models and the MCP toggle on as the owner does, and off again
// when the test ends.
func enableMCP(t *testing.T, k *kube, owner *ownerAPI) {
	t.Helper()
	bearer := connectrpc.WithInterceptors(withBearer(k.token(t, plugins.ServiceAccountName(uiPlugin), plugins.CoreAudience)))
	settings := modelsv1connect.NewModelSettingsServiceClient(http.DefaultClient, owner.base, bearer)
	origin := "http://" + enrollHost
	set := func(on bool) {
		if _, err := settings.SetAgentsEnabled(t.Context(), connectrpc.NewRequest(&modelsv1.SetAgentsEnabledRequest{Origin: origin, SessionToken: owner.session, Enabled: on})); err != nil {
			t.Fatalf("SetAgentsEnabled(%v): %v", on, err)
		}
		if _, err := settings.SetFeatureEnabled(t.Context(), connectrpc.NewRequest(&modelsv1.SetFeatureEnabledRequest{Origin: origin, SessionToken: owner.session, Feature: "mcp", Enabled: on})); err != nil {
			t.Fatalf("SetFeatureEnabled(mcp, %v): %v", on, err)
		}
	}
	set(true)
	t.Cleanup(func() { set(false) })
}

// createMCPToken makes a script token that may use MCP and is granted one target, as the Access page does, and
// returns its secret.
func createMCPToken(t *testing.T, owner *ownerAPI, name string, kind identityv1.GrantTargetKind, id string) string {
	t.Helper()
	resp, err := owner.auth.CreateScriptToken(t.Context(), connectrpc.NewRequest(&identityv1.CreateScriptTokenRequest{
		Origin: "http://" + enrollHost, SessionToken: owner.session, Name: name, McpAllowed: true,
		Grants: []*identityv1.GrantTarget{{Kind: kind, Id: id}},
	}))
	if err != nil {
		t.Fatalf("CreateScriptToken(%s): %v", name, err)
	}
	return resp.Msg.GetSecret()
}

// accountsOver returns the IDs of the accounts the read API lists for token on the core's browser port, sorted.
func accountsOver(t *testing.T, base, token string) []string {
	t.Helper()
	r := bundleGet(t, base+"/v1/accounts", token)
	if r.status != http.StatusOK {
		t.Fatalf("GET /v1/accounts = %d %s, want 200", r.status, r.body)
	}
	var accounts viewsv1.ListAccountsResponse
	if err := protojson.Unmarshal(r.body, &accounts); err != nil {
		t.Fatalf("decoding the accounts: %v", err)
	}
	ids := make([]string, 0, len(accounts.GetAccounts()))
	for _, a := range accounts.GetAccounts() {
		ids = append(ids, a.GetId())
	}
	slices.Sort(ids)
	return ids
}

// postMCP sends one JSON-RPC message to /mcp on the core's browser port with token as the bearer token.
func postMCP(t *testing.T, base, token, message string) reply {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/mcp", strings.NewReader(message))
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	return doReply(t, req)
}

// uiPage returns the body of GET path on the core's browser port, which proxies it to the built-in UI, with
// session as the signed-in session, as a browser on the allowed origin sends it.
func uiPage(t *testing.T, base, path, session string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, http.NoBody)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	req.Host = enrollHost
	req.AddCookie(&http.Cookie{Name: "__Host-openfief_session", Value: session, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	r := doReply(t, req)
	if r.status != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", path, r.status)
	}
	return string(r.body)
}

// doReply sends req without following a redirect, so a page that sends the browser to the login page fails the
// caller's status check, and returns the status and the body.
func doReply(t *testing.T, req *http.Request) reply {
	t.Helper()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Timeout: 2 * requestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL.Path, err)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, readAPIBodyLimit))
	if closeErr := resp.Body.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatalf("%s %s: reading the body: %v", req.Method, req.URL.Path, err)
	}
	return reply{status: resp.StatusCode, body: body}
}

// checkUnboundItemsAreTheOwners checks that an item tied to no account, an unmapped broker account's, is
// listed to the owner when there is one and to no script token, whatever it is granted (D-165).
func checkUnboundItemsAreTheOwners(t *testing.T, owner *ownerAPI, base string, tokens ...string) {
	t.Helper()
	const unbound = "broker-account-unmapped"
	kinds := func(token string) []string {
		r := bundleGet(t, base+"/v1/attention-items", token)
		if r.status != http.StatusOK {
			t.Fatalf("GET /v1/attention-items = %d %s, want 200", r.status, r.body)
		}
		var items struct {
			Items []struct {
				Kind string `json:"kind"`
			} `json:"items"`
		}
		if err := json.Unmarshal(r.body, &items); err != nil {
			t.Fatalf("decoding the attention items: %v", err)
		}
		out := make([]string, 0, len(items.Items))
		for _, it := range items.Items {
			out = append(out, it.Kind)
		}
		return out
	}
	for _, token := range tokens {
		if slices.Contains(kinds(token), unbound) {
			t.Errorf("a script token is listed an attention item of kind %s, which is tied to no account and the owner's alone", unbound)
		}
	}
	if got := kinds(owner.session); slices.Contains(got, unbound) {
		t.Logf("the owner is listed %d items, one of them of kind %s, which no token is", len(got), unbound)
	}
}

// checkMCPGrants checks that over MCP the same grants limit what a tool sees.
func checkMCPGrants(t *testing.T, base, household, narrow string) {
	t.Helper()
	valueOf := func(token string) string {
		r := postMCP(t, base, token, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"value","arguments":{}}}`)
		if r.status != http.StatusOK || !strings.Contains(string(r.body), `"content"`) || strings.Contains(string(r.body), `"isError":true`) {
			t.Fatalf("tools/call value = %d %s, want a card", r.status, r.body)
		}
		return string(r.body)
	}
	if narrowCard, householdCard := valueOf(narrow), valueOf(household); narrowCard == householdCard {
		t.Errorf("MCP's value is the same for the one-account and the household token: %s", narrowCard)
	}
}

// checkPagesRender checks that the built-in UI renders the accounts for the owner's session and a demo session.
func checkPagesRender(t *testing.T, owner *ownerAPI, base string) {
	t.Helper()
	demoSession, err := owner.auth.CreateDemoSession(t.Context(), connectrpc.NewRequest(&identityv1.CreateDemoSessionRequest{Origin: "http://" + enrollHost}))
	if err != nil {
		t.Fatalf("CreateDemoSession: %v", err)
	}
	for name, session := range map[string]string{"the owner's": owner.session, "a demo": demoSession.Msg.GetSessionToken()} {
		if got := accountsOver(t, base, session); !slices.Equal(got, demoAccountIDs()) {
			t.Errorf("%s session lists %v over /v1 under the policy, want every demo account", name, got)
		}
		page := uiPage(t, base, "/accounts", session)
		for _, id := range demoAccountIDs() {
			if !strings.Contains(page, id) {
				t.Errorf("the accounts page for %s session does not show %s", name, id)
			}
		}
	}
}
