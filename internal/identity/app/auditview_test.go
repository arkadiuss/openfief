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
