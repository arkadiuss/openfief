//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
)

// viewerName is what the owner calls the suite's viewer; its accounts, if any, would carry the suite's prefix.
const viewerName = suiteAccountPrefix + "viewer"

// viewerOrigin is the origin the suite's owner and viewer sign in on.
const viewerOrigin = "http://" + enrollHost

// TestViewersAndTokensReadOnlyTheirGrants checks D-65, D-209 and D-210 on the core that answers from SpiceDB: a
// script token the owner creates with one sleeve grant reads that sleeve's accounts alone; the owner invites a viewer
// with the same one grant through InviteViewer, the viewer enrolls a passkey through the link, signs in with it, and
// sees only that sleeve's accounts through the read API and through the UI's pages, is refused a change, and, once
// the owner revokes them, their session stops working (SCEN-e2e-viewers-read-only-their-grants).
//
// It is not parallel, like TestGrantsLimitWhatAScriptTokenViews: the owner's session must still be within five
// minutes of its passkey check (D-183) when it creates the token and the link, so Go runs it before the parallel
// tests start. It creates no account.
//
//nolint:paralleltest // see above
func TestViewersAndTokensReadOnlyTheirGrants(t *testing.T) {
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	waitSeedComplete(t, k)
	waitSpiceDBAvailable(t, k)
	waitPluginReady(t, k, uiPlugin)
	waitLogLine(t, k, coreDeployment, reconciledLine)
	owner := newOwnerAPI(t, k)
	sleeve := []*identityv1.GrantTarget{{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, Id: string(demo.DefensiveSleeve)}}
	bonds := []string{string(demo.BondsAccount)}

	// Both need the owner's recent passkey check, which the session made with its enrollment has.
	tokenResp, err := owner.auth.CreateScriptToken(t.Context(), connectrpc.NewRequest(&identityv1.CreateScriptTokenRequest{
		Origin: viewerOrigin, SessionToken: owner.session, Name: "e2e-sleeve", Ttl: durationpb.New(time.Hour), Grants: sleeve,
	}))
	if err != nil {
		t.Fatalf("CreateScriptToken with one sleeve grant: %v", err)
	}
	invite, err := owner.auth.InviteViewer(t.Context(), connectrpc.NewRequest(&identityv1.InviteViewerRequest{
		Origin: viewerOrigin, SessionToken: owner.session, Name: viewerName, Grants: sleeve, Ttl: durationpb.New(time.Hour),
	}))
	if err != nil {
		t.Fatalf("InviteViewer: %v", err)
	}

	// The token reads the sleeve's accounts alone.
	secret := tokenResp.Msg.GetSecret()
	if got := waitAccounts(t, k, secret, func(ids []string) bool { return len(ids) > 0 }); !slices.Equal(got, bonds) {
		t.Fatalf("a script token created with one sleeve grant lists %v, want %v", got, bonds)
	}
	expectStatus(t, k, secret, accountPath(demo.XTBAccount), http.StatusNotFound, "an account outside the token's sleeve")

	viewer := enrollViewer(t, owner, invite.Msg.GetUrl())
	checkViewerReads(t, k, owner, viewer, bonds)
	checkRevokedViewer(t, owner, viewer)
}

// enrollViewer makes a viewer's passkey through the invitation link with a virtual authenticator, as their browser
// would, checks the enrollment's session is a viewer's read session and the link is spent, signs in with the passkey
// and returns the session token of that sign-in.
func enrollViewer(t *testing.T, o *ownerAPI, link string) string {
	t.Helper()
	_, code, ok := strings.Cut(link, "#code=")
	if !ok || code == "" {
		t.Fatalf("the invitation link %q has no code", link)
	}
	enrolled := registerViewer(t, o, code)
	who, err := o.auth.WhoAmI(t.Context(), connectrpc.NewRequest(&identityv1.WhoAmIRequest{Origin: viewerOrigin, SessionToken: enrolled.session}))
	if err != nil || who.Msg.GetPerson().GetKind() != identityv1.PersonKind_PERSON_KIND_VIEWER || who.Msg.GetSession().GetScope() != identityv1.SessionScope_SESSION_SCOPE_READ {
		t.Fatalf("the enrollment's session is %v (%v), want a viewer's read session", who.Msg, err)
	}
	// The link was used up: opening it again is refused.
	if _, err = o.auth.BeginRegistration(t.Context(), connectrpc.NewRequest(&identityv1.BeginRegistrationRequest{
		Origin: viewerOrigin, Authorization: &identityv1.BeginRegistrationRequest_EnrollmentCode{EnrollmentCode: code},
	})); err == nil {
		t.Error("a spent invitation link began another registration")
	}
	login, err := o.auth.BeginLogin(t.Context(), connectrpc.NewRequest(&identityv1.BeginLoginRequest{Origin: viewerOrigin}))
	if err != nil {
		t.Fatalf("BeginLogin: %v", err)
	}
	assertion, err := enrolled.device.Get([]byte(login.Msg.GetOptionsJson()), viewerOrigin)
	if err != nil {
		t.Fatalf("signing in with the viewer's passkey: %v", err)
	}
	done, err := o.auth.FinishLogin(t.Context(), connectrpc.NewRequest(&identityv1.FinishLoginRequest{
		Origin: viewerOrigin, CeremonyId: login.Msg.GetCeremonyId(), CredentialJson: string(assertion),
	}))
	if err != nil {
		t.Fatalf("FinishLogin as the viewer: %v", err)
	}
	return done.Msg.GetSessionToken()
}

// enrollment is a viewer's passkey on its virtual authenticator and the session the enrollment signed them in with.
type enrollment struct {
	device  *virtual.Authenticator
	session string
}

// registerViewer enrolls a passkey on a new virtual authenticator with the invitation's code and returns the
// authenticator and the session the enrollment signed the viewer in with.
func registerViewer(t *testing.T, o *ownerAPI, code string) enrollment {
	t.Helper()
	begin, err := o.auth.BeginRegistration(t.Context(), connectrpc.NewRequest(&identityv1.BeginRegistrationRequest{
		Origin: viewerOrigin, Authorization: &identityv1.BeginRegistrationRequest_EnrollmentCode{EnrollmentCode: code},
	}))
	if err != nil {
		t.Fatalf("BeginRegistration with the invitation: %v", err)
	}
	if begin.Msg.GetInvitedName() != viewerName {
		t.Errorf("the invitation greets %q, want %q", begin.Msg.GetInvitedName(), viewerName)
	}
	device, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("a virtual authenticator: %v", err)
	}
	credential, err := device.Create([]byte(begin.Msg.GetOptionsJson()), viewerOrigin)
	if err != nil {
		t.Fatalf("creating the viewer's passkey: %v", err)
	}
	finish, err := o.auth.FinishRegistration(t.Context(), connectrpc.NewRequest(&identityv1.FinishRegistrationRequest{
		Origin: viewerOrigin, CeremonyId: begin.Msg.GetCeremonyId(), CredentialJson: string(credential), Name: "e2e-viewer",
	}))
	if err != nil {
		t.Fatalf("FinishRegistration with the invitation: %v", err)
	}
	return enrollment{device: device, session: finish.Msg.GetSessionToken()}
}

// checkViewerReads checks that the viewer lists the granted sleeve's accounts alone through the read API, sees only
// them in the UI, and is refused a change.
func checkViewerReads(t *testing.T, k *kube, o *ownerAPI, viewer string, want []string) {
	t.Helper()
	err := waitFor(t.Context(), spicedbTimeout, func(ctx context.Context) (string, bool, error) {
		ids, accErr := o.accountsOf(ctx, viewer)
		return fmt.Sprint(ids), slices.Equal(ids, want), accErr
	})
	if err != nil {
		t.Fatalf("waiting for the viewer to list only the granted sleeve's accounts: %v", err)
	}
	checkViewerUI(t, k, viewer)
	// A viewer changes nothing: the core refuses an owner call made with their session.
	_, err = o.accounts.CreateAccount(t.Context(), connectrpc.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: viewerOrigin, SessionToken: viewer, Name: suiteAccountPrefix + "viewer-account", Institution: "treasury-bonds", Sleeve: string(demo.DefensiveSleeve),
	}))
	if connectrpc.CodeOf(err) != connectrpc.CodePermissionDenied {
		t.Fatalf("CreateAccount with a viewer's session: %v, want PermissionDenied", err)
	}
}

// checkRevokedViewer checks that the owner lists the viewer with their grant and a session, revokes them, and that
// their session then stops reading and is unknown.
func checkRevokedViewer(t *testing.T, o *ownerAPI, viewer string) {
	t.Helper()
	people, err := o.auth.ListPeople(t.Context(), connectrpc.NewRequest(&identityv1.ListPeopleRequest{Origin: viewerOrigin, SessionToken: o.session}))
	if err != nil {
		t.Fatalf("ListPeople: %v", err)
	}
	var id string
	for _, p := range people.Msg.GetPeople() {
		if p.GetPerson().GetName() == viewerName && len(p.GetGrants()) == 1 && p.GetActiveSessions() > 0 {
			id = p.GetPerson().GetId()
		}
	}
	if id == "" {
		t.Fatalf("ListPeople lists no %s with one grant and a session: %v", viewerName, people.Msg.GetPeople())
	}
	if _, err = o.auth.RevokePerson(t.Context(), connectrpc.NewRequest(&identityv1.RevokePersonRequest{Origin: viewerOrigin, SessionToken: o.session, PersonId: id})); err != nil {
		t.Fatalf("RevokePerson: %v", err)
	}
	err = waitFor(t.Context(), spicedbTimeout, func(ctx context.Context) (string, bool, error) {
		_, listErr := o.accountsOf(ctx, viewer)
		return fmt.Sprint(listErr), listErr != nil, nil
	})
	if err != nil {
		t.Fatalf("waiting for the revoked viewer's session to stop reading: %v", err)
	}
	_, err = o.auth.WhoAmI(t.Context(), connectrpc.NewRequest(&identityv1.WhoAmIRequest{Origin: viewerOrigin, SessionToken: viewer}))
	if connectrpc.CodeOf(err) != connectrpc.CodeUnauthenticated {
		t.Errorf("WhoAmI with the revoked viewer's session: %v, want Unauthenticated", err)
	}
}

// checkViewerUI checks that the built-in UI, reached through the core as a browser reaches it with the viewer's
// session cookie, lists the granted sleeve's account on its accounts page and no other, and has no Access page.
func checkViewerUI(t *testing.T, k *kube, session string) {
	t.Helper()
	headers := []string{"Cookie: __Host-openfief_session=" + session, "Host: " + enrollHost}
	page := func(path string) reply {
		r, err := get(t.Context(), k, clientPod, coreURL+path, headers...)
		if err != nil {
			t.Fatalf("GET %s as the viewer: %v", path, err)
		}
		return r
	}
	accounts := page("/accounts")
	if accounts.status != http.StatusOK || !bytes.Contains(accounts.body, []byte("Demo treasury bonds (IKE)")) {
		t.Fatalf("the viewer's accounts page: status %d, want 200 listing the granted account", accounts.status)
	}
	for _, other := range []string{"Demo XTB", "Demo IBKR"} {
		if bytes.Contains(accounts.body, []byte(other)) {
			t.Errorf("the viewer's accounts page lists %q, outside their grant", other)
		}
	}
	if r := page("/access"); r.status != http.StatusForbidden {
		t.Errorf("the viewer's GET /access = %d, want 403", r.status)
	}
}
