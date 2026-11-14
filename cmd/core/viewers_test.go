package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// viewerEnroll opens the viewer's link code through c and creates their
// passkey, returning their authenticator and the read session it signs them
// in with.
func viewerEnroll(t *testing.T, c authAPI, code string) (*virtual.Authenticator, *identityv1.FinishRegistrationResponse) {
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
		t.Fatalf("BeginRegistration with a viewer's link: %v", err)
	}
	if begun.Msg.GetInvitedName() != "Alice" {
		t.Errorf("invited name = %q, want Alice", begun.Msg.GetInvitedName())
	}
	cred, err := dev.Create([]byte(begun.Msg.GetOptionsJson()), localUI)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	reg, err := c.FinishRegistration(ctx, connect.NewRequest(&identityv1.FinishRegistrationRequest{
		Origin: localUI, CeremonyId: begun.Msg.GetCeremonyId(), CredentialJson: string(cred), Name: "phone",
	}))
	if err != nil {
		t.Fatalf("FinishRegistration with a viewer's link: %v", err)
	}
	return dev, reg.Msg
}

// accountIDs returns the IDs GET /v1/accounts lists for token.
func accountIDs(t *testing.T, srv *httptest.Server, token string) []string {
	t.Helper()
	var accounts viewsv1.ListAccountsResponse
	getJSON(t, srv, "/v1/accounts", token, &accounts)
	ids := make([]string, 0, len(accounts.GetAccounts()))
	for _, a := range accounts.GetAccounts() {
		ids = append(ids, a.GetId())
	}
	slices.Sort(ids)
	return ids
}

// TestViewerSeesOnlyWhatTheyAreGranted runs the owner's invitation of a viewer
// end to end on the core, under the policy: the viewer enrolls their own
// passkey, reads only what their grant covers (accounts, sleeves, kinds),
// is refused every change, and loses everything when revoked (D-65, D-163,
// D-209, D-210).
func TestViewerSeesOnlyWhatTheyAreGranted(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	link, _, _ := strings.Cut(enrollCmd(t, dsn), "\n")
	ownerCode, ok := strings.CutPrefix(link, localUI+"/enroll#code=")
	if !ok {
		t.Fatalf("enroll-link printed %q", link)
	}
	_, srv, pool := startCoreForUI(t, dsn, func(*serveConfig) {})
	loadDemo(t, pool)
	c := authClient(srv.URL, srv.Client(), uiCaller)
	ownerSession := passkeyLogin(t, c, passkeyEnroll(t, c, ownerCode))
	all := []string{string(demo.BondsAccount), string(demo.IBKRAccount), string(demo.XTBAccount)}
	if got := accountIDs(t, srv, ownerSession); !slices.Equal(got, all) {
		t.Fatalf("the owner sees %v, want %v", got, all)
	}
	code := inviteAlice(t, c, ownerSession)
	dev, reg := viewerEnroll(t, c, code)
	viewer := reg.GetSessionToken()
	if reg.GetSession().GetScope() != identityv1.SessionScope_SESSION_SCOPE_READ || reg.GetSession().GetPersonId() == "owner" {
		t.Fatalf("the viewer's session = %v, want a read session of their own", reg.GetSession())
	}
	if err := beginWithSpentLink(t, c, code); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("the spent link = %v, want permission_denied", err)
	}
	checkViewerReads(t, srv, viewer)
	checkViewerRefusedEverywhere(t, srv, c, viewer)

	// The viewer signs in again with their own passkey.
	again := passkeyLogin(t, c, dev)
	if got := accountIDs(t, srv, again); !slices.Equal(got, []string{string(demo.BondsAccount)}) {
		t.Errorf("the viewer after a login sees %v, want only the bonds account", got)
	}
	revokeAlice(t, c, ownerSession)
	checkRefused(t, srv, viewer, "a revoked viewer's session")
	checkRefused(t, srv, again, "a revoked viewer's second session")
}

// inviteAlice has the owner of session invite Alice with the defensive sleeve
// and returns the code of the link.
func inviteAlice(t *testing.T, c authAPI, session string) string {
	t.Helper()
	invite, err := c.InviteViewer(t.Context(), connect.NewRequest(&identityv1.InviteViewerRequest{
		Origin: localUI, SessionToken: session, Name: "Alice",
		Grants: []*identityv1.GrantTarget{{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, Id: string(demo.DefensiveSleeve)}},
	}))
	if err != nil {
		t.Fatalf("InviteViewer: %v", err)
	}
	code, ok := strings.CutPrefix(invite.Msg.GetUrl(), localUI+"/enroll#code=")
	if !ok {
		t.Fatalf("the link is %q", invite.Msg.GetUrl())
	}
	return code
}

// revokeAlice has the owner of session list the people, find Alice with her
// grant, and revoke her.
func revokeAlice(t *testing.T, c authAPI, session string) {
	t.Helper()
	people, err := c.ListPeople(t.Context(), connect.NewRequest(&identityv1.ListPeopleRequest{Origin: localUI, SessionToken: session}))
	if err != nil || len(people.Msg.GetPeople()) != 2 {
		t.Fatalf("ListPeople = %v, %v, want the owner and Alice", people, err)
	}
	alice := people.Msg.GetPeople()[1]
	if alice.GetPerson().GetName() != "Alice" || len(alice.GetGrants()) != 1 {
		t.Fatalf("ListPeople's second person = %v, want Alice with her grant", alice)
	}
	_, err = c.RevokePerson(t.Context(), connect.NewRequest(&identityv1.RevokePersonRequest{Origin: localUI, SessionToken: session, PersonId: alice.GetPerson().GetId()}))
	if err != nil {
		t.Fatalf("RevokePerson: %v", err)
	}
}

// beginWithSpentLink begins a registration with code once more, as someone
// holding a spent link would, and returns the core's answer.
func beginWithSpentLink(t *testing.T, c authAPI, code string) error {
	t.Helper()
	_, err := c.BeginRegistration(t.Context(), connect.NewRequest(&identityv1.BeginRegistrationRequest{
		Origin: localUI, Authorization: &identityv1.BeginRegistrationRequest_EnrollmentCode{EnrollmentCode: code},
	}))
	return err //nolint:wrapcheck // the code of the core's refusal is what the test reads
}

// checkViewerReads checks that the read API shows the viewer only their
// sleeve's account, its sleeve and the sleeve's kind, and not another account.
func checkViewerReads(t *testing.T, srv *httptest.Server, viewer string) {
	t.Helper()
	if got := accountIDs(t, srv, viewer); !slices.Equal(got, []string{string(demo.BondsAccount)}) {
		t.Errorf("the viewer sees accounts %v, want only the bonds account", got)
	}
	if r := get(t, srv, "/v1/accounts/"+string(demo.XTBAccount), viewer); r.status != http.StatusNotFound {
		t.Errorf("GET an account not granted = %d, want 404: a viewer can't tell it from one not stored", r.status)
	}
	var sleeves viewsv1.ListSleevesResponse
	getJSON(t, srv, "/v1/sleeves", viewer, &sleeves)
	if len(sleeves.GetSleeves()) != 1 || sleeves.GetSleeves()[0].GetId() != string(demo.DefensiveSleeve) {
		t.Errorf("the viewer sees sleeves %v, want only the defensive one", sleeves.GetSleeves())
	}
	var kinds viewsv1.ListSleeveKindsResponse
	getJSON(t, srv, "/v1/sleeve-kinds", viewer, &kinds)
	if len(kinds.GetKinds()) != 1 {
		t.Errorf("the viewer sees %d sleeve kinds, want the one the sleeve has", len(kinds.GetKinds()))
	}
}

// checkViewerRefusedEverywhere checks that the viewer's session is refused
// the owner's calls: the identity service's and another service's.
func checkViewerRefusedEverywhere(t *testing.T, srv *httptest.Server, c authAPI, viewer string) {
	t.Helper()
	ctx := t.Context()
	if _, err := c.CreateScriptToken(ctx, connect.NewRequest(&identityv1.CreateScriptTokenRequest{Origin: localUI, SessionToken: viewer, Name: "x"})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("CreateScriptToken by a viewer = %v, want permission_denied", err)
	}
	if _, err := c.InviteViewer(ctx, connect.NewRequest(&identityv1.InviteViewerRequest{Origin: localUI, SessionToken: viewer, Name: "Eve"})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("InviteViewer by a viewer = %v, want permission_denied", err)
	}
	if _, err := c.ListPeople(ctx, connect.NewRequest(&identityv1.ListPeopleRequest{Origin: localUI, SessionToken: viewer})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("ListPeople by a viewer = %v, want permission_denied", err)
	}
	sleeves := portfoliov1connect.NewSleeveServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(callerToken(uiCaller)))
	_, err := sleeves.AddSleeve(ctx, connect.NewRequest(&portfoliov1.AddSleeveRequest{Origin: localUI, SessionToken: viewer, Name: "Mine", Kind: "bonds"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("AddSleeve by a viewer = %v, want permission_denied", err)
	}
}

// callerToken is an interceptor that presents token as the view plugin's.
func callerToken(token string) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
}
