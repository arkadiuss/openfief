package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

// TestAuditEntriesFilterByOutcomeAndActions checks, for the owner's session,
// that the log can be filtered by outcome and by a set of actions as the
// Access page's history and the Settings audit log ask (D-211): the invitation
// and the viewer's enrollment are listed by the set, a refused sign-in by its
// outcome, and an invalid filter is refused.
func TestAuditEntriesFilterByOutcomeAndActions(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	code := f.invite(t, &owner, "Alice", sleeve(sleeveCore))
	_, _, enrollErr := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code})
	must(t, "enrolling", enrollErr)
	if _, _, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code}); err == nil {
		t.Fatal("a spent link was accepted")
	}
	history := []identity.AuditAction{identity.ActionEnrollmentLinkIssued, identity.ActionEnrollmentLinkUsed, identity.ActionPersonCreated, identity.ActionGrantCreated}
	page, err := f.auth.AuditEntries(t.Context(), local, owner.Token, &identity.AuditFilter{Actions: history})
	must(t, "AuditEntries by actions", err)
	got := make([]string, 0, len(page.Entries))
	for i := range page.Entries {
		got = append(got, string(page.Entries[i].Action))
	}
	want := []string{"enrollment_link_used", "grant_created", "person_created", "enrollment_link_used", "enrollment_link_issued", "enrollment_link_used", "enrollment_link_issued"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries of the history actions, newest first (-want +got):\n%s", diff)
	}
	page, err = f.auth.AuditEntries(t.Context(), local, owner.Token, &identity.AuditFilter{Outcome: identity.OutcomeRefused})
	must(t, "AuditEntries by outcome", err)
	if len(page.Entries) != 1 {
		t.Errorf("%d refused entries, want the spent link's one", len(page.Entries))
	}
	for i := range page.Entries {
		if page.Entries[i].Outcome != identity.OutcomeRefused {
			t.Errorf("entry %+v is not refused", page.Entries[i])
		}
	}
	if _, err = f.auth.AuditEntries(t.Context(), local, owner.Token, &identity.AuditFilter{Outcome: "maybe"}); err == nil {
		t.Error("an unknown outcome was accepted")
	}
}

// TestAuditEntriesKeepNamingAViewerWhoseSessionEnded checks that an entry a
// viewer made in a session still names the person after the session ended and
// after the owner revoked the person: sessions and people are never deleted
// (migration 0002), so the log can name them for ever, by an ID that the people
// list, revoked ones included, resolves to the name (D-211).
func TestAuditEntriesKeepNamingAViewerWhoseSessionEnded(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	_, r := f.enrollViewer(t, &owner, sleeve(sleeveCore))
	person, session := r.Session.Session.Person(), r.Session.Session.ID
	if _, err := f.auth.RevokePerson(t.Context(), local, owner.Token, person); err != nil {
		t.Fatalf("RevokePerson: %v", err)
	}
	if _, err := f.auth.Session(t.Context(), origin(t, local), r.Session.Token, identity.SessionScopeRead); err == nil {
		t.Fatal("the revoked viewer's session still works")
	}
	page, err := f.auth.AuditEntries(t.Context(), local, owner.Token, &identity.AuditFilter{ActorKind: identity.ActorViewer, ActorID: string(session)})
	must(t, "AuditEntries", err)
	if len(page.Entries) == 0 || page.Persons[session] != person {
		t.Fatalf("entries of the ended session = %d, persons %v; want its entries and %s as the person", len(page.Entries), page.Persons, person)
	}
	people, err := f.auth.People(t.Context(), local, owner.Token)
	must(t, "People", err)
	name := ""
	for i := range people {
		if people[i].Person.ID == person {
			name = people[i].Person.Name
		}
	}
	if name != "Alice" {
		t.Errorf("the revoked person's name = %q, want Alice: the log must still name them", name)
	}
}

// TestAuditEntriesNameViewerSessions checks that a viewer's session, which is
// the actor of their entries, comes with the ID of the person it is of, so the
// log can name them; an owner's or a link's actor has none.
func TestAuditEntriesNameViewerSessions(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	_, r := f.enrollViewer(t, &owner, sleeve(sleeveCore))
	page, err := f.auth.AuditEntries(t.Context(), local, owner.Token, &identity.AuditFilter{})
	must(t, "AuditEntries", err)
	want := map[identity.SessionID]string{r.Session.Session.ID: r.Session.Session.Person()}
	if diff := cmp.Diff(want, page.Persons); diff != "" {
		t.Errorf("persons of the viewer sessions on the page (-want +got):\n%s", diff)
	}
}
