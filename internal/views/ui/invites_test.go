package ui

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
)

// errNoSuchInvite is the fake core's refusal of cancelling an invitation that is
// gone.
var errNoSuchInvite = errors.New("no such pending invitation")

// TestAccessPendingInvites checks that the People section lists each pending
// invitation with the name, what the viewer will be granted in words (a target
// the lists no longer have by its ID) and when the link stops working, with a
// button to cancel it, and never the link itself.
func TestAccessPendingInvites(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"Pending invitations", "Dana", "Sleeve Core", "Account Synthetic brokerage (synthetic-broker)", "Account acc-deleted-1",
		"2026-03-02 12:30 UTC", `action="/access/invites/`+pendingInvite+`/cancel"`, "Cancel invitation",
	)
	wantNoBody(t, rec, "synthetic-invite-code", "#code=")
}

// TestCancelInvite checks that cancelling reaches the core as one CancelInvite
// of the invitation named, needing no passkey check, goes back to the People
// section where the invitation is gone, and shows the core's refusal when it
// is gone already.
func TestCancelInvite(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.stale = true })
	rec := h.serve(t, withSession(form("/access/invites/"+pendingInvite+"/cancel", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/access#people-title" {
		t.Errorf("Location = %q, want /access#people-title", got)
	}
	if diff := cmp.Diff([]string{pendingInvite}, seen(h.auth, func() []string { return h.auth.cancelledInvites })); diff != "" {
		t.Errorf("cancelled invitations (-want +got):\n%s", diff)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantBody(t, rec, "No invitations are waiting to be used")
	wantNoBody(t, rec, "Cancel invitation")

	h.auth.locked(func() { h.auth.cancelInviteErr = codeErr(connect.CodeNotFound, errNoSuchInvite) })
	rec = h.serve(t, withSession(form("/access/invites/"+pendingInvite+"/cancel", ""), ownerSession))
	wantStatus(t, rec, http.StatusNotFound)
	wantBody(t, rec, "no such pending invitation", "People")
}

// TestEnrollmentLinkNoLongerValid checks that a link the core refuses (expired,
// cancelled, used or unknown) gets one plain sentence that reveals nothing about
// the household and does not repeat the core's words, on the enrollment page's
// way in.
func TestEnrollmentLinkNoLongerValid(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.codeRefused = true })
	rec := h.serve(t, request(http.MethodPost, "/enroll/begin", `{"code":"synthetic-code"}`))
	wantStatus(t, rec, http.StatusForbidden)
	wantBody(t, rec, "This invite is no longer valid.")
	wantNoBody(t, rec, "household", "expired", "forbidden")
	if n := strings.Count(rec.Body.String(), "error"); n != 1 {
		t.Errorf("the answer has %d mentions of an error, want its one field", n)
	}
}
