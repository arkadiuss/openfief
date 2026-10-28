package ui

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestSettingsPage checks the owner's settings: passkeys with remove,
// script tokens with revoke, the forms to create a token, a passkey on this
// host and an enrollment link.
func TestSettingsPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"laptop", testHost, `action="/settings/passkeys/pk1/remove"`,
		"backup-script", "active", `action="/settings/tokens/fedcba9876543210/revoke"`,
		`action="/settings/tokens"`, `data-webauthn="register"`, `action="/settings/enrollment-links"`,
	)
}

// TestCreateScriptToken checks that a new script token's secret is shown
// once, in the response to its creation, with the lifetime the form asked
// for; and that a lifetime out of range never reaches the core.
func TestCreateScriptToken(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/settings/tokens", "name=nightly&days=7"), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Script token nightly created", "ofs_0123456789abcdef_synthetic-secret", "shown only this once")
	var ttl time.Duration
	h.auth.locked(func() { ttl = h.auth.tokenTTL })
	if got := ttl; got != 7*24*time.Hour {
		t.Errorf("TTL = %v, want 7 days", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantNoBody(t, rec, "synthetic-secret")

	rec = h.serve(t, withSession(form("/settings/tokens", "name=nightly&days=400"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "from 1 to 365")
}

// TestRevokeAndRemove checks revoking a script token and removing a
// passkey, and that the core's refusal to remove the last passkey is shown
// on the settings page with its reason.
func TestRevokeAndRemove(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/settings/tokens/fedcba9876543210/revoke", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	rec = h.serve(t, withSession(form("/settings/passkeys/pk1/remove", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	got := append(seen(h.auth, func() []string { return h.auth.revoked }), seen(h.auth, func() []string { return h.auth.removed })...)
	if diff := cmp.Diff([]string{"fedcba9876543210", "pk1"}, got); diff != "" {
		t.Errorf("revoked and removed (-want +got):\n%s", diff)
	}

	h.auth.locked(func() { h.auth.lastPasskey = true })
	rec = h.serve(t, withSession(form("/settings/passkeys/pk1/remove", ""), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "the last passkey can&#39;t be removed", "Passkeys")
}

// TestCreateEnrollmentLink checks the link for a passkey on another host
// (D-60): the target origin goes to the core and the link is shown once.
func TestCreateEnrollmentLink(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/settings/enrollment-links", "target_origin=https%3A%2F%2Fother.example.com"), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Enrollment link created", "https://other.example.com/enroll#code=synthetic-code", "2026-03-01 12:00 UTC")
	if got := seenOne(h.auth, func() string { return h.auth.linkTarget }); got != "https://other.example.com" {
		t.Errorf("target origin = %q", got)
	}
}
