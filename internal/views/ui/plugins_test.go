package ui

import (
	"net/http"
	"strings"
	"testing"
)

// TestPluginsPage checks that a session of either scope sees the installed
// plugins with their capabilities, trust, signer and status, reached from
// the navigation, and that the page offers nothing that changes a plugin
// (D-46): its only form is the header's sign-out.
func TestPluginsPage(t *testing.T) {
	t.Parallel()
	for _, token := range []string{ownerSession, readSession, demoSession} {
		t.Run(token, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			rec := h.serve(t, withSession(request(http.MethodGet, "/plugins", ""), token))
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, `href="/plugins"`, "Installed plugins",
				"ibkr-main", "connector", "Unreviewed", "Signed by OpenFief", "openfief-ci", "Running", "sha256:a1a1",
				"ui-local", "requests view", "Blocked", "Refused", "UntrustedSigner", "SignatureNotVerified")
			if forms := strings.Count(rec.Body.String(), "<form"); forms != 1 {
				t.Errorf("%d forms on the plugins page, want only sign-out", forms)
			}
			wantNoBody(t, rec, "<button type=\"submit\">", `action="/plugins`)
			// The session reads the settings, then the plugins.
			if got := seen(h.read, func() []string { return h.read.bearers }); len(got) != 2 || got[0] != "Bearer "+token || got[1] != got[0] {
				t.Errorf("read API bearers = %q, want the session's token twice", got)
			}
		})
	}
}

// TestPluginsPageUnavailable checks that a core that can't read its
// cluster leaves a sentence on the plugins and settings pages instead of
// failing them.
func TestPluginsPageUnavailable(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.read.locked(func() { h.read.pluginsUnavailable = true })
	for _, path := range []string{"/plugins", "/settings"} {
		rec := h.serve(t, withSession(request(http.MethodGet, path, ""), ownerSession))
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "Installed plugins", "be read now")
		wantNoBody(t, rec, "ibkr-main")
	}
}

// TestSettingsShowsPlugins checks that the owner's settings page lists the
// installed plugins next to the owner's other settings.
func TestSettingsShowsPlugins(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Installed plugins", "ibkr-main", "Signed by OpenFief", "ui-local", "Refused", "Script tokens")
}

// TestPluginsPageNeedsASession checks that /plugins sends a browser
// without a session to the login page before calling the core.
func TestPluginsPageNeedsASession(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, request(http.MethodGet, "/plugins", ""))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
	if got := seen(h.read, func() []string { return h.read.bearers }); len(got) != 0 {
		t.Errorf("read API called without a session: %q", got)
	}
}

// TestTransactionsMarkRemovedInstance checks that an account's transactions
// name the connector instance that brought each in and mark one that is no
// longer installed (D-58), on the page and in the next page's fragment.
func TestTransactionsMarkRemovedInstance(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, r := range []*http.Request{
		request(http.MethodGet, "/accounts/"+accountID, ""),
		htmx(request(http.MethodGet, "/accounts/"+accountID+"/transactions?page_token=page-2", "")),
	} {
		rec := h.serve(t, withSession(r, readSession))
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "synthetic", "ibkr-old", `class="pill pill-neutral removed"`, "removed instance")
		wantNoBody(t, rec, "plugin:ibkr-old:ibkr")
	}
}
