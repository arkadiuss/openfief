package app_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
)

// passErr returns err: it lets a table of calls return the error of one that
// also returns a value, without handing a call's error on unwrapped.
func passErr[T any](_ T, err error) error { return err }

// invite has the owner invite a viewer called name with targets, and returns
// the code from the link.
func (f *grantFixture) invite(t *testing.T, owner *app.NewSession, name string, targets ...identity.GrantTarget) string {
	t.Helper()
	l, err := f.auth.InviteViewer(t.Context(), local, owner.Token, name, targets, "", 0)
	must(t, "InviteViewer", err)
	code, ok := strings.CutPrefix(l.URL, local+identity.EnrollPath+"#code=")
	if !ok {
		t.Fatalf("link %q is not an enrollment link on %s", l.URL, local)
	}
	return code
}

// enrollViewer invites Alice with targets and has her enroll a passkey,
// returning her authenticator and what the registration returned.
func (f *grantFixture) enrollViewer(t *testing.T, owner *app.NewSession, targets ...identity.GrantTarget) (*virtual.Authenticator, app.Registered) {
	t.Helper()
	dev, r, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: f.invite(t, owner, "Alice", targets...)})
	must(t, "enrolling a viewer", err)
	if r.Session == nil {
		t.Fatal("a viewer's enrollment issued no session")
	}
	return dev, r
}

// trail returns the audit entries after the first skip, as "actor action
// outcome refs", a reference to a passkey, a session or a link by its kind only.
func (f *grantFixture) trail(skip int) []string {
	entries := f.store.entries()[skip:]
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		words := []string{string(e.Actor.Kind), string(e.Action), string(e.Outcome)}
		for _, ref := range e.Refs {
			switch ref.Kind {
			case identity.RefPasskey, identity.RefSession, identity.RefEnrollmentCode:
				words = append(words, string(ref.Kind))
			case identity.RefPerson, identity.RefScriptToken, identity.RefViewPlugin, identity.RefAccount, identity.RefSleeve, identity.RefHousehold,
				identity.RefModelRole, identity.RefModel, identity.RefModelCall, identity.RefModelError, identity.RefMaintainerRun:
				words = append(words, string(ref.Kind)+":"+ref.ID)
			}
		}
		out = append(out, strings.Join(words, " "))
	}
	return out
}

// TestViewerEnrolls checks that a viewer opens the owner's link, enrolls
// their own passkey and is signed in with a read session, with their person
// created and their grants made, every step an audit entry (D-65, D-210,
// D-211).
func TestViewerEnrolls(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	before := len(f.store.entries())
	_, r := f.enrollViewer(t, &owner, sleeve(sleeveCore))
	pid := r.Session.Session.Person()
	if !r.Session.Session.OfViewer() || r.Session.Session.Scope != identity.SessionScopeRead || r.Passkey.PersonID != pid {
		t.Fatalf("session %+v, passkey of %q: want a read session of the viewer who owns the passkey", r.Session.Session, r.Passkey.PersonID)
	}
	if r.NotProjected != nil {
		t.Errorf("NotProjected = %v", r.NotProjected)
	}
	f.checkAlice(t, r.Session.Token, pid)
	want := []string{
		"owner enrollment_link_issued ok enrollment_code",
		"enrollment_code enrollment_link_used ok enrollment_code passkey person:" + pid,
		"enrollment_code person_created ok person:" + pid + " enrollment_code",
		"enrollment_code passkey_added ok passkey person:" + pid,
		"enrollment_code grant_created ok person:" + pid + " sleeve:" + sleeveCore,
		"viewer sign_in ok session person:" + pid,
	}
	if diff := cmp.Diff(want, f.trail(before)); diff != "" {
		t.Errorf("audit trail (-want +got):\n%s", diff)
	}
	if lines := f.model.lines(); len(lines) != 1 || !strings.Contains(lines[0], pid) {
		t.Errorf("model = %v, want the viewer's one grant projected", lines)
	}
}

// checkAlice checks that the session of token is of the viewer Alice, with
// the ID pid and no WebAuthn handle shown.
func (f *grantFixture) checkAlice(t *testing.T, token, pid string) {
	t.Helper()
	_, _, person, err := f.auth.Identify(t.Context(), local, token)
	must(t, "Identify", err)
	if person.ID != pid || person.Name != "Alice" || !person.IsViewer() || len(person.Handle) != 0 {
		t.Errorf("Identify = %+v, want Alice the viewer, without a handle", person)
	}
}

// TestViewerLinkIsSpentAndTheViewerSignsInAgain checks that the link is spent
// by the enrollment, refused and recorded the second time, and that the viewer
// signs in again with their own passkey and reads as the viewer holder.
func TestViewerLinkIsSpentAndTheViewerSignsInAgain(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	code := f.invite(t, &owner, "Alice", sleeve(sleeveCore))
	dev, r, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code})
	must(t, "enrolling", err)
	pid := r.Session.Session.Person()
	if _, _, again := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code}); !errors.Is(again, app.ErrForbidden) {
		t.Errorf("a spent link = %v, want %v", again, app.ErrForbidden)
	}
	last := f.store.entries()[len(f.store.entries())-1]
	if last.Action != identity.ActionEnrollmentLinkUsed || last.Outcome != identity.OutcomeRefused || last.Actor.Kind != identity.ActorEnrollmentCode {
		t.Errorf("last entry = %+v, want the spent link's refusal", last)
	}
	ns, err := f.login(t, dev, local)
	must(t, "the viewer's login", err)
	if ns.Session.Person() != pid || ns.Session.Scope != identity.SessionScopeRead {
		t.Errorf("login session = %+v, want a read session of the viewer", ns.Session)
	}
	p, err := f.auth.Authenticate(t.Context(), ns.Token)
	must(t, "Authenticate", err)
	if c := p.Caller(); c.Holder != identity.HolderViewer || c.ID != pid {
		t.Errorf("Caller = %+v, want the viewer %s", c, pid)
	}
}

// TestExpiredViewerLinkRefused checks that an expired link is refused, plainly
// and with an entry, and creates nobody.
func TestExpiredViewerLinkRefused(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	code := f.invite(t, &owner, "Bob")
	f.clock.set(authStart().Add(identity.DefaultInviteTTL + time.Second))
	if _, _, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code}); !errors.Is(err, app.ErrForbidden) || !errors.Is(err, identity.ErrEnrollmentCodeExpired) {
		t.Errorf("an expired link = %v, want %v for %v", err, app.ErrForbidden, identity.ErrEnrollmentCodeExpired)
	}
	last := f.store.entries()[len(f.store.entries())-1]
	if last.Action != identity.ActionEnrollmentLinkUsed || last.Outcome != identity.OutcomeRefused {
		t.Errorf("last entry = %+v, want the expired link's refusal", last)
	}
	people, err := f.store.People(t.Context())
	must(t, "People", err)
	if len(people) != 1 {
		t.Errorf("%d people after a refused link, want only the owner", len(people))
	}
}

// TestInviteViewerRefusals checks that only an owner session with a passkey
// check invites, and that a bad name or a target that isn't there stores no
// link.
func TestInviteViewerRefusals(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	_, viewer := f.enrollViewer(t, &owner)
	for _, tt := range []struct {
		name    string
		token   string
		who     string
		targets []identity.GrantTarget
		want    error
	}{
		{"a viewer's session", viewer.Session.Token, "Eve", nil, app.ErrForbidden},
		{"no session", "", "Eve", nil, app.ErrInvalidSession},
		{"no name", owner.Token, "", nil, app.ErrInvalidRequest},
		{"an unprintable name", owner.Token, "Eve\x00", nil, app.ErrInvalidRequest},
		{"an unknown sleeve", owner.Token, "Eve", []identity.GrantTarget{sleeve("nowhere")}, app.ErrNotFound},
	} {
		before := len(f.store.codes)
		if _, err := f.auth.InviteViewer(t.Context(), local, tt.token, tt.who, tt.targets, "", 0); !errors.Is(err, tt.want) {
			t.Errorf("InviteViewer with %s = %v, want %v", tt.name, err, tt.want)
		}
		if len(f.store.codes) != before {
			t.Errorf("InviteViewer with %s stored a code", tt.name)
		}
	}
	f.clock.set(authStart().Add(identity.PasskeyCheckWindow + time.Minute))
	if _, err := f.auth.InviteViewer(t.Context(), local, owner.Token, "Eve", nil, "", 0); !errors.Is(err, app.ErrPasskeyCheckRequired) {
		t.Errorf("InviteViewer without a recent passkey check = %v, want %v", err, app.ErrPasskeyCheckRequired)
	}
}

// TestViewerSessionChangesNothing checks that a viewer's session is refused by
// every owner call of the identity service, which is what the central scope
// check gives every other service too.
func TestViewerSessionChangesNothing(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	_, viewer := f.enrollViewer(t, &owner, account(accountMain))
	tok := f.token(t)
	ctx, token, a := t.Context(), viewer.Session.Token, f.auth
	all := []identity.GrantTarget{identity.Household()}
	calls := map[string]func() error{
		"CreateScriptToken": func() error {
			_, _, err := a.CreateScriptToken(ctx, local, token, app.NewToken{Name: "x"})
			return passErr(0, err)
		},
		"ListScriptTokens":     func() error { return passErr(a.ScriptTokens(ctx, local, token)) },
		"RevokeScriptToken":    func() error { return passErr(0, a.RevokeScriptToken(ctx, local, token, tok.ID)) },
		"ListPasskeys":         func() error { return passErr(a.Passkeys(ctx, local, token)) },
		"RemovePasskey":        func() error { return passErr(a.RemovePasskey(ctx, local, token, "x")) },
		"CreateEnrollmentLink": func() error { return passErr(a.CreateEnrollmentLink(ctx, local, token, "", 0)) },
		"InviteViewer":         func() error { return passErr(a.InviteViewer(ctx, local, token, "Eve", nil, "", 0)) },
		"BeginPasskeyCheck":    func() error { return passErr(a.BeginPasskeyCheck(ctx, local, token)) },
		"SignOutEverywhere":    func() error { return passErr(a.SignOutEverywhere(ctx, local, token)) },
		"ListAuditEntries":     func() error { return passErr(a.AuditEntries(ctx, local, token, &identity.AuditFilter{})) },
		"CreateGrant":          func() error { return passErr(a.CreateGrant(ctx, local, token, tok, identity.Household())) },
		"RevokeGrant":          func() error { return passErr(0, a.RevokeGrant(ctx, local, token, tok, identity.Household())) },
		"ListGrants":           func() error { return passErr(a.ListGrants(ctx, local, token, identity.GrantMatch{}, true)) },
		"UpdateGrants":         func() error { return passErr(a.UpdateGrants(ctx, local, token, tok, all, nil)) },
		"ListViewPlugins":      func() error { return passErr(a.ViewPlugins(ctx, local, token)) },
		"ListPeople":           func() error { return passErr(a.People(ctx, local, token)) },
		"RevokePerson":         func() error { return passErr(a.RevokePerson(ctx, local, token, viewer.Session.Session.Person())) },
	}
	for name, call := range calls {
		if err := call(); !errors.Is(err, app.ErrForbidden) {
			t.Errorf("%s in a viewer's session = %v, want %v", name, err, app.ErrForbidden)
		}
	}
	if err := a.Logout(ctx, local, token); err != nil {
		t.Errorf("Logout in a viewer's session = %v, want it allowed: sign-out is all a viewer changes", err)
	}
}

// TestOwnerPasskeyCheckRefusesAViewersPasskey checks that a viewer's passkey
// can't confirm the owner's session, and that the owner's passkey list leaves
// viewers' passkeys out.
func TestOwnerPasskeyCheckRefusesAViewersPasskey(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	viewerDev, _ := f.enrollViewer(t, &owner)
	st, err := f.auth.BeginPasskeyCheck(t.Context(), local, owner.Token)
	must(t, "BeginPasskeyCheck", err)
	assertion, err := viewerDev.Get(st.Options, local)
	must(t, "the viewer's assertion", err)
	if _, checkErr := f.auth.FinishPasskeyCheck(t.Context(), local, owner.Token, st.CeremonyID, assertion); !errors.Is(checkErr, app.ErrAuthenticationFailed) {
		t.Errorf("a passkey check with a viewer's passkey = %v, want %v", checkErr, app.ErrAuthenticationFailed)
	}
	keys, err := f.auth.Passkeys(t.Context(), local, owner.Token)
	must(t, "Passkeys", err)
	if len(keys) != 1 || keys[0].Person() != identity.OwnerPersonID {
		t.Errorf("the owner's passkeys = %d, want only their own", len(keys))
	}
}

// TestPeopleListedWithTheirAccess checks that the owner lists the people with
// their grants, passkeys and sessions (D-210).
func TestPeopleListedWithTheirAccess(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	_, viewer := f.enrollViewer(t, &owner, sleeve(sleeveCore), account(accountMain))
	people, err := f.auth.People(t.Context(), local, owner.Token)
	must(t, "People", err)
	if len(people) != 2 || people[0].Person.ID != identity.OwnerPersonID || people[0].Person.Name != "Owner" || people[0].Passkeys != 1 {
		t.Fatalf("People = %+v, want the owner first with a passkey", people)
	}
	if a := people[1]; a.Person.ID != viewer.Session.Session.Person() || a.Person.Name != "Alice" || a.Passkeys != 1 || a.ActiveSessions != 1 || len(a.Grants) != 2 {
		t.Errorf("Alice = %+v, want one passkey, one session and two grants", a)
	}
}

// TestRevokeViewer checks that revoking a viewer ends their sessions, disables
// their passkeys and ends their grants, each audited and removed from the
// model, keeping the person (D-210, D-211, Q8).
func TestRevokeViewer(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	dev, viewer := f.enrollViewer(t, &owner, sleeve(sleeveCore), account(accountMain))
	pid := viewer.Session.Session.Person()
	r, err := f.auth.RevokePerson(t.Context(), local, owner.Token, pid)
	must(t, "RevokePerson", err)
	if r != (app.Revocation{Sessions: 1, Passkeys: 1, Grants: 2}) {
		t.Errorf("Revocation = %+v, want one session, one passkey, two grants", r)
	}
	if _, sessionErr := f.auth.Session(t.Context(), identity.Origin{}, viewer.Session.Token, identity.SessionScopeRead); !errors.Is(sessionErr, app.ErrInvalidSession) {
		t.Errorf("a revoked viewer's session = %v, want %v", sessionErr, app.ErrInvalidSession)
	}
	if _, loginErr := f.login(t, dev, local); !errors.Is(loginErr, app.ErrAuthenticationFailed) {
		t.Errorf("a revoked viewer's login = %v, want %v", loginErr, app.ErrAuthenticationFailed)
	}
	people, err := f.auth.People(t.Context(), local, owner.Token)
	must(t, "People after the revocation", err)
	if a := people[1]; a.Person.ID != pid || a.Person.Active() || a.Passkeys != 0 || a.ActiveSessions != 0 || len(a.Grants) != 0 {
		t.Errorf("revoked Alice = %+v, want her record kept, with nothing working", a)
	}
	f.checkRevocationEntries(t)
	if got := f.model.lines(); len(got) != 4 {
		t.Errorf("model = %v, want the two grants projected and unprojected", got)
	}
}

// checkRevocationEntries checks that revoking Alice left one person_revoked,
// one passkey_removed and two grant_revoked entries.
func (f *grantFixture) checkRevocationEntries(t *testing.T) {
	t.Helper()
	counts := map[identity.AuditAction]int{}
	for _, e := range f.store.entries() {
		counts[e.Action]++
	}
	if counts[identity.ActionPersonRevoked] != 1 || counts[identity.ActionPasskeyRemoved] != 1 || counts[identity.ActionGrantRevoked] != 2 {
		t.Errorf("entries: %d person_revoked, %d passkey_removed, %d grant_revoked, want 1, 1, 2",
			counts[identity.ActionPersonRevoked], counts[identity.ActionPasskeyRemoved], counts[identity.ActionGrantRevoked])
	}
}

// TestRevokePersonRefusals checks what can't be revoked: a viewer revoked
// already, an unknown one, the owner and a malformed ID.
func TestRevokePersonRefusals(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	_, viewer := f.enrollViewer(t, &owner)
	pid := viewer.Session.Session.Person()
	if _, err := f.auth.RevokePerson(t.Context(), local, owner.Token, pid); err != nil {
		t.Fatalf("RevokePerson: %v", err)
	}
	for name, tt := range map[string]struct {
		id   string
		want error
	}{
		"a viewer revoked already": {pid, app.ErrNotFound},
		"an unknown viewer":        {"0123456789abcdef", app.ErrNotFound},
		"the owner":                {identity.OwnerPersonID, app.ErrInvalidRequest},
		"a malformed ID":           {"x", app.ErrInvalidRequest},
	} {
		if _, err := f.auth.RevokePerson(t.Context(), local, owner.Token, tt.id); !errors.Is(err, tt.want) {
			t.Errorf("RevokePerson of %s = %v, want %v", name, err, tt.want)
		}
	}
}

// TestUpdateViewerGrants checks that the owner changes an active viewer's
// grants as a token's, and can't grant a revoked one.
func TestUpdateViewerGrants(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	_, viewer := f.enrollViewer(t, &owner, sleeve(sleeveCore))
	holder := identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: viewer.Session.Session.Person()}
	got, err := f.auth.UpdateGrants(t.Context(), local, owner.Token, holder, []identity.GrantTarget{account(accountMain)}, []identity.GrantTarget{sleeve(sleeveCore)})
	must(t, "UpdateGrants", err)
	if len(got) != 1 || got[0].Target != account(accountMain) {
		t.Errorf("grants = %+v, want only the account", got)
	}
	if _, err = f.auth.RevokePerson(t.Context(), local, owner.Token, holder.ID); err != nil {
		t.Fatalf("RevokePerson: %v", err)
	}
	if _, err = f.auth.UpdateGrants(t.Context(), local, owner.Token, holder, []identity.GrantTarget{identity.Household()}, nil); !errors.Is(err, app.ErrNotFound) {
		t.Errorf("UpdateGrants of a revoked viewer = %v, want %v", err, app.ErrNotFound)
	}
}
