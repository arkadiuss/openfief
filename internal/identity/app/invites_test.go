package app_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
)

// inviteID returns the ID of the enrollment code in an invitation's code text.
func inviteID(t *testing.T, code string) string {
	t.Helper()
	id, _, err := identity.ParseEnrollmentCode(code)
	if err != nil {
		t.Fatalf("ParseEnrollmentCode: %v", err)
	}
	return string(id)
}

// registerCancelling is register with between run after the ceremony began and
// before it finishes.
func (f *authFixture) registerCancelling(t *testing.T, code string, between func()) (*virtual.Authenticator, app.Registered, error) {
	t.Helper()
	dev, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("virtual.New: %v", err)
	}
	st, err := f.auth.BeginRegistration(t.Context(), local, app.RegistrationGrant{EnrollmentCode: code})
	must(t, "BeginRegistration", err)
	between()
	cred, err := dev.Create(st.Options, local)
	must(t, "Create", err)
	r, err := f.auth.FinishRegistration(t.Context(), local, st.CeremonyID, cred, "")
	if err != nil {
		return nil, app.Registered{}, fmt.Errorf("FinishRegistration: %w", err)
	}
	return dev, r, nil
}

// TestInviteLifetimeDefaultsToAnHour checks that a viewer's link lasts an hour
// when the owner chooses nothing, while 15 minutes and 24 hours are accepted,
// more than 24 hours is refused, and the owner's own enrollment link keeps its
// shorter default (D-183).
func TestInviteLifetimeDefaultsToAnHour(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tests := []struct {
		name string
		ttl  time.Duration
		want time.Duration
	}{
		{"nothing chosen", 0, time.Hour},
		{"15 minutes", 15 * time.Minute, 15 * time.Minute},
		{"24 hours", 24 * time.Hour, 24 * time.Hour},
	}
	for _, tt := range tests {
		l, err := f.auth.InviteViewer(t.Context(), local, owner.Token, "Alice", nil, "", tt.ttl)
		must(t, tt.name, err)
		if got := l.ExpiresAt.Sub(f.clock.now()); got != tt.want {
			t.Errorf("%s: the link lasts %s, want %s", tt.name, got, tt.want)
		}
	}
	if _, err := f.auth.InviteViewer(t.Context(), local, owner.Token, "Alice", nil, "", 25*time.Hour); !errors.Is(err, app.ErrInvalidRequest) {
		t.Errorf("a link for 25 hours = %v, want %v", err, app.ErrInvalidRequest)
	}
	l, err := f.auth.CreateEnrollmentLink(t.Context(), local, owner.Token, "", 0)
	must(t, "CreateEnrollmentLink", err)
	if got := l.ExpiresAt.Sub(f.clock.now()); got != identity.DefaultEnrollmentTTL {
		t.Errorf("the owner's own link lasts %s, want %s", got, identity.DefaultEnrollmentTTL)
	}
}

// inviteNames returns the names of invites, in order.
func inviteNames(invites []app.Invite) []string {
	out := make([]string, 0, len(invites))
	for i := range invites {
		out = append(out, invites[i].Name)
	}
	return out
}

// TestPendingInvitesAndCancel checks that the owner lists the invitations not
// used, cancelled or expired, each with its name, grants and expiry; and that
// cancelling one needs no passkey check, is an audit entry, leaves the others
// and is not found for what is gone.
func TestPendingInvitesAndCancel(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	alice := f.invite(t, &owner, "Alice", sleeve(sleeveCore))
	f.clock.set(f.clock.now().Add(time.Second))
	bob := f.invite(t, &owner, "Bob")
	f.clock.set(f.clock.now().Add(time.Second))
	f.invite(t, &owner, "Carol", sleeve(sleeveCore))
	f.clock.set(f.clock.now().Add(10 * time.Minute)) // the passkey check is stale now

	got, err := f.auth.Invites(t.Context(), local, owner.Token)
	must(t, "Invites", err)
	if diff := cmp.Diff([]string{"Alice", "Bob", "Carol"}, inviteNames(got)); diff != "" {
		t.Errorf("pending invitations (-want +got):\n%s", diff)
	}
	if got[0].ID != identity.EnrollmentCodeID(inviteID(t, alice)) || len(got[0].Targets) != 1 || got[0].ExpiresAt.Sub(got[0].CreatedAt) != time.Hour {
		t.Errorf("Alice's invitation = %+v, want her code's ID, her sleeve and an hour", got[0])
	}
	if _, inviteErr := f.auth.InviteViewer(t.Context(), local, owner.Token, "Dave", nil, "", 0); !errors.Is(inviteErr, app.ErrPasskeyCheckRequired) {
		t.Fatalf("InviteViewer with a stale check = %v, want %v", inviteErr, app.ErrPasskeyCheckRequired)
	}

	skip := len(f.store.entries())
	must(t, "CancelInvite", f.auth.CancelInvite(t.Context(), local, owner.Token, inviteID(t, bob)))
	if diff := cmp.Diff([]string{"owner enrollment_link_cancelled ok enrollment_code"}, f.trail(skip)); diff != "" {
		t.Errorf("audit trail (-want +got):\n%s", diff)
	}
	got, err = f.auth.Invites(t.Context(), local, owner.Token)
	must(t, "Invites", err)
	if diff := cmp.Diff([]string{"Alice", "Carol"}, inviteNames(got)); diff != "" {
		t.Errorf("pending invitations after cancelling Bob (-want +got):\n%s", diff)
	}
	for name, id := range map[string]string{"cancelled already": inviteID(t, bob), "unknown": "0123456789abcdef"} {
		if cancelErr := f.auth.CancelInvite(t.Context(), local, owner.Token, id); !errors.Is(cancelErr, app.ErrNotFound) {
			t.Errorf("cancelling %s = %v, want %v", name, cancelErr, app.ErrNotFound)
		}
	}
}

// TestCancelledAndExpiredLinksAreRefusedAlike checks that a cancelled link is
// refused with the error an expired one gets, naming nobody, and the refusal is
// recorded by the link's ID; that nothing is pending afterwards; and that an
// expired invitation can't be cancelled.
func TestCancelledAndExpiredLinksAreRefusedAlike(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	bob := f.invite(t, &owner, "Bob")
	carol := f.invite(t, &owner, "Carol", sleeve(sleeveCore))
	f.clock.set(f.clock.now().Add(10 * time.Minute))
	must(t, "CancelInvite", f.auth.CancelInvite(t.Context(), local, owner.Token, inviteID(t, bob)))
	_, _, cancelled := f.register(t, local, app.RegistrationGrant{EnrollmentCode: bob})
	f.clock.set(f.clock.now().Add(2 * time.Hour))
	_, _, expired := f.register(t, local, app.RegistrationGrant{EnrollmentCode: carol})
	for name, err := range map[string]error{"cancelled": cancelled, "expired": expired} {
		if !errors.Is(err, app.ErrForbidden) {
			t.Errorf("a %s link = %v, want %v", name, err, app.ErrForbidden)
		}
	}
	if strings.Contains(cancelled.Error(), "Bob") || strings.Contains(expired.Error(), "Carol") {
		t.Errorf("errors %q and %q name the invited viewer", cancelled, expired)
	}
	last := f.store.entries()[len(f.store.entries())-1]
	if last.Action != identity.ActionEnrollmentLinkUsed || last.Outcome != identity.OutcomeRefused || last.Actor.Kind != identity.ActorEnrollmentCode {
		t.Errorf("last entry = %+v, want the expired link's refusal", last)
	}
	left, err := f.auth.Invites(t.Context(), local, owner.Token)
	must(t, "Invites", err)
	if len(left) != 0 {
		t.Errorf("invitations %v are pending after the cancelled and the expired", inviteNames(left))
	}
	if err := f.auth.CancelInvite(t.Context(), local, owner.Token, inviteID(t, carol)); !errors.Is(err, app.ErrNotFound) {
		t.Errorf("cancelling an expired invitation = %v, want %v", err, app.ErrNotFound)
	}
}

// TestCancelledInviteCannotBeUsedMidCeremony checks that an invitation the owner
// cancels after the viewer began enrolling and before they finished creates no
// viewer: the link counts as spent when the registration finishes.
func TestCancelledInviteCannotBeUsedMidCeremony(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	code := f.invite(t, &owner, "Alice", sleeve(sleeveCore))
	_, _, err := f.registerCancelling(t, code, func() {
		must(t, "CancelInvite", f.auth.CancelInvite(t.Context(), local, owner.Token, inviteID(t, code)))
	})
	if !errors.Is(err, app.ErrForbidden) {
		t.Fatalf("finishing a cancelled invitation = %v, want %v", err, app.ErrForbidden)
	}
	people, peopleErr := f.auth.People(t.Context(), local, owner.Token)
	must(t, "People", peopleErr)
	if len(people) != 1 {
		t.Errorf("%d people after a cancelled invitation, want the owner alone", len(people))
	}
}

// TestInvitesAreTheOwners checks that listing and cancelling invitations take an
// owner session: a viewer's, a demo session's and no session are refused.
func TestInvitesAreTheOwners(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{on: true, marked: true})
	_, owner := f.enroll(t)
	code := f.invite(t, &owner, "Alice")
	_, viewer := f.enrollViewer(t, &owner)
	demo, err := f.auth.Login(t.Context(), f.auth.Demo(), local, app.Proof{})
	must(t, "demo Login", err)
	for name, token := range map[string]string{"a viewer": viewer.Session.Token, "a demo session": demo.Token, "none": "", "a made-up one": "ofsess_x"} {
		if _, listErr := f.auth.Invites(t.Context(), local, token); listErr == nil {
			t.Errorf("%s listed invitations", name)
		}
		if cancelErr := f.auth.CancelInvite(t.Context(), local, token, inviteID(t, code)); cancelErr == nil {
			t.Errorf("%s cancelled an invitation", name)
		}
	}
	if pending, listErr := f.auth.Invites(t.Context(), local, owner.Token); listErr != nil || len(pending) != 1 {
		t.Errorf("Invites = %v, %v, want the one invitation still pending", pending, listErr)
	}
}
