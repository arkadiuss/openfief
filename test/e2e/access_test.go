//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	connectrpc "connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
	"github.com/arkadiuss/openfief/internal/plugins"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

const (
	// enrollHost is the host `core enroll-link` makes a link for: the origin the core allows by default.
	enrollHost = "localhost:8080"
)

var (
	// errNoEnrollLink reports `core enroll-link` output with no link in it.
	errNoEnrollLink = errors.New("core enroll-link printed no link")
	// errNoTokenID reports `core token issue` output with no token ID in it.
	errNoTokenID = errors.New("core token issue printed no token ID")
)

// ownerAPI is the owner's session as the built-in UI holds it: the AuthService and the AccountService, called
// as the UI plugin through a port-forward to the core's plugin port, with the session token of an owner session
// made from an enrollment link by a virtual authenticator, as a browser's would be.
type ownerAPI struct {
	auth     identityv1connect.AuthServiceClient
	accounts ingestv1connect.AccountServiceClient
	session  string
	// base is the address of the port-forward to the core's plugin port.
	base string
}

// newOwnerAPI makes an owner session: it asks the core's pod for an enrollment link (`core enroll-link`, as the
// owner's first run does), creates a passkey on a virtual authenticator through the AuthService and keeps the session
// that comes with it. The calls carry the UI plugin's projected token, which the core accepts from a view plugin.
func newOwnerAPI(t *testing.T, k *kube) *ownerAPI {
	t.Helper()
	pod, err := k.runningPod(t.Context(), namespace, "openfief-core")
	if err != nil {
		t.Fatal(err)
	}
	out, err := k.exec(t.Context(), namespace, pod, "core", nil, "/core", "enroll-link", "--host", enrollHost)
	if err != nil {
		t.Fatalf("core enroll-link: %v", err)
	}
	_, code, ok := strings.Cut(strings.SplitN(string(out), "\n", 2)[0], "#code=")
	if !ok || code == "" {
		t.Fatal(errNoEnrollLink)
	}
	base := "http://" + k.forwardPod(t, namespace, podPort{pod: pod, port: corePluginPort})
	bearerUI := connectrpc.WithInterceptors(withBearer(k.token(t, plugins.ServiceAccountName(uiPlugin), plugins.CoreAudience)))
	o := &ownerAPI{
		auth:     identityv1connect.NewAuthServiceClient(http.DefaultClient, base, bearerUI),
		accounts: ingestv1connect.NewAccountServiceClient(http.DefaultClient, base, bearerUI),
		base:     base,
	}
	origin := "http://" + enrollHost
	begin, err := o.auth.BeginRegistration(t.Context(), connectrpc.NewRequest(&identityv1.BeginRegistrationRequest{
		Origin: origin, Authorization: &identityv1.BeginRegistrationRequest_EnrollmentCode{EnrollmentCode: code},
	}))
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	device, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("a virtual authenticator: %v", err)
	}
	credential, err := device.Create([]byte(begin.Msg.GetOptionsJson()), origin)
	if err != nil {
		t.Fatalf("creating a passkey: %v", err)
	}
	finish, err := o.auth.FinishRegistration(t.Context(), connectrpc.NewRequest(&identityv1.FinishRegistrationRequest{
		Origin: origin, CeremonyId: begin.Msg.GetCeremonyId(), CredentialJson: string(credential), Name: "e2e",
	}))
	if err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}
	if o.session = finish.Msg.GetSessionToken(); o.session == "" {
		t.Fatal("registering with an enrollment link issued no session")
	}
	return o
}

// accountsOf returns the IDs of the accounts the read API lists for token on the core's plugin port, where a view
// plugin calls it, sorted.
func (o *ownerAPI) accountsOf(ctx context.Context, token string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.base+"/v1/accounts", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("GET /v1/accounts: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET /v1/accounts: %w", err)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, readAPIBodyLimit))
	if closeErr := resp.Body.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, fmt.Errorf("GET /v1/accounts: reading the body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: GET /v1/accounts: status %d", errRetry, resp.StatusCode)
	}
	var accounts viewsv1.ListAccountsResponse
	if err := protojson.Unmarshal(body, &accounts); err != nil {
		return nil, fmt.Errorf("decoding the accounts: %w", err)
	}
	ids := make([]string, 0, len(accounts.GetAccounts()))
	for _, a := range accounts.GetAccounts() {
		ids = append(ids, a.GetId())
	}
	slices.Sort(ids)
	return ids, nil
}

// grant lets the holder view the target, as the owner does.
func (o *ownerAPI) grant(t *testing.T, holder identityv1.GrantHolderKind, holderID string, target identityv1.GrantTargetKind, targetID string) {
	t.Helper()
	_, err := o.auth.CreateGrant(t.Context(), connectrpc.NewRequest(&identityv1.CreateGrantRequest{
		Origin: "http://" + enrollHost, SessionToken: o.session, HolderKind: holder, HolderId: holderID, TargetKind: target, TargetId: targetID,
	}))
	if err != nil {
		t.Fatalf("CreateGrant(%v %s, %v %s): %v", holder, holderID, target, targetID, err)
	}
}

// revoke ends a grant, as the owner does.
func (o *ownerAPI) revoke(t *testing.T, holder identityv1.GrantHolderKind, holderID string, target identityv1.GrantTargetKind, targetID string) {
	t.Helper()
	_, err := o.auth.RevokeGrant(t.Context(), connectrpc.NewRequest(&identityv1.RevokeGrantRequest{
		Origin: "http://" + enrollHost, SessionToken: o.session, HolderKind: holder, HolderId: holderID, TargetKind: target, TargetId: targetID,
	}))
	if err != nil {
		t.Fatalf("RevokeGrant(%v %s, %v %s): %v", holder, holderID, target, targetID, err)
	}
}

// auditEntries returns the audit log's entries of action, newest first.
func (o *ownerAPI) auditEntries(t *testing.T, action identityv1.AuditAction) []*identityv1.AuditEntry {
	t.Helper()
	resp, err := o.auth.ListAuditEntries(t.Context(), connectrpc.NewRequest(&identityv1.ListAuditEntriesRequest{
		Origin: "http://" + enrollHost, SessionToken: o.session, Action: action, PageSize: identityAuditPage,
	}))
	if err != nil {
		t.Fatalf("ListAuditEntries: %v", err)
	}
	return resp.Msg.GetEntries()
}

// identityAuditPage is the largest page of the audit log.
const identityAuditPage = 200

// createAccount creates a hand-entered bond account in sleeve and returns its ID.
func (o *ownerAPI) createAccount(t *testing.T, name, sleeve string) string {
	t.Helper()
	resp, err := o.accounts.CreateAccount(t.Context(), connectrpc.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: "http://" + enrollHost, SessionToken: o.session, Name: name, Institution: "treasury-bonds", Sleeve: sleeve,
	}))
	if err != nil {
		t.Fatalf("CreateAccount(%s): %v", name, err)
	}
	return resp.Msg.GetAccount().GetId()
}

// deleteAccount deletes an account that holds no transactions.
func (o *ownerAPI) deleteAccount(t *testing.T, id string) {
	t.Helper()
	_, err := o.accounts.DeleteAccount(context.WithoutCancel(t.Context()), connectrpc.NewRequest(&ingestv1.DeleteAccountRequest{
		Origin: "http://" + enrollHost, SessionToken: o.session, AccountId: id,
	}))
	if err != nil {
		t.Errorf("DeleteAccount(%s): %v", id, err)
	}
}

// coreAuthzFlags returns the flags that make the core's own commands pass a grant on to SpiceDB when the core
// answers from it, and none while it runs the policy: the commands run in the core's pod, where SpiceDB's endpoint
// and key are in the environment already.
func coreAuthzFlags(t *testing.T, k *kube) []string {
	t.Helper()
	d, err := k.clients.AppsV1().Deployments(namespace).Get(t.Context(), coreDeployment, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting Deployment %s: %v", coreDeployment, err)
	}
	if coreRunsSpiceDB(d) {
		return []string{authzSpiceDBFlag, authzInsecureFlag}
	}
	return nil
}

// coreRunsSpiceDB reports whether the core Deployment runs with --authz=spicedb.
func coreRunsSpiceDB(d *appsv1.Deployment) bool {
	for i := range d.Spec.Template.Spec.Containers {
		c := &d.Spec.Template.Spec.Containers[i]
		if c.Name != "core" {
			continue
		}
		if slices.Contains(c.Args, authzSpiceDBFlag) {
			return true
		}
	}
	return false
}

// issuedToken is a script token the suite issued: its secret, and its ID.
type issuedToken struct {
	secret, id string
}

// issueTokenWith issues a script token with the core's own CLI inside its pod, with extra flags, and returns its
// secret and its ID, from the first two lines of the output. The output is never logged.
func issueTokenWith(t *testing.T, k *kube, extra ...string) issuedToken {
	t.Helper()
	pod, err := k.runningPod(t.Context(), namespace, "openfief-core")
	if err != nil {
		t.Fatal(err)
	}
	args := append([]string{"/core", "token", "issue", "--name", "e2e", "--ttl", "1h"}, extra...)
	out, err := k.exec(t.Context(), namespace, pod, "core", nil, args...)
	if err != nil {
		t.Fatalf("issuing a script token: %v", err)
	}
	first, rest, _ := bytes.Cut(out, []byte("\n"))
	if len(first) == 0 {
		t.Fatal(errNoSecret)
	}
	fields := strings.Fields(string(bytes.SplitN(rest, []byte("\n"), 2)[0]))
	if len(fields) < 2 || fields[0] != "id" {
		t.Fatal(errNoTokenID)
	}
	return issuedToken{secret: string(first), id: strings.TrimSuffix(fields[1], ",")}
}

// grantHousehold lets the view plugin instance view every account, as the owner does from the Access page: a plugin
// installed after the core started has no grant, and under SpiceDB no account.
func grantHousehold(t *testing.T, k *kube, instance string) {
	t.Helper()
	newOwnerAPI(t, k).grant(t, identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, instance, identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, "")
}
