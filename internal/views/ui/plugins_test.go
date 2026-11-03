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
				"ui-local", "requests view", "Blocked", "Refused", "UntrustedSigner", "SignatureNotVerified",
				"Signed only by a signer you don't trust:", "the image may be forged",
				"https://example.com/stranger/.github/workflows/sign.yml@refs/heads/main (issuer https://token.actions.githubusercontent.com)")
			if forms := strings.Count(rec.Body.String(), "<form"); forms != 1 {
				t.Errorf("%d forms on the plugins page, want only sign-out", forms)
			}
			wantNoBody(t, rec, "<button type=\"submit\">", `action="/plugins`)
			// The session reads the settings and the open attention items, then the plugins.
			if got := seen(h.read, func() []string { return h.read.bearers }); len(got) != 3 || got[0] != "Bearer "+token || got[1] != got[0] || got[2] != got[0] {
				t.Errorf("read API bearers = %q, want the session's token three times", got)
			}
		})
	}
}

// TestPluginsPageMarketData checks that the plugins page shows the owner
// which plugin is a market-data connector and so learns every instrument
// they hold (D-123), and only that one.
func TestPluginsPageMarketData(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/plugins", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	body := rec.Body.String()
	_, after, found := strings.Cut(body, "prices-main")
	row, _, _ := strings.Cut(after, "</tr>")
	if !found || !strings.Contains(row, "Market data") || !strings.Contains(row, "every instrument you hold") {
		t.Errorf("prices-main's row doesn't say it learns what the owner holds:\n%s", row)
	}
	if !strings.Contains(row, `Books for <span class="mono ident">eodhd</span> only.`) {
		t.Errorf("the market-data connector's row doesn't name the institution it books for (D-182):\n%s", row)
	}
	if !strings.Contains(row, `Asks to book for <span class="mono ident">eodhd-eu</span>; not approved`) {
		t.Errorf("the market-data connector's row doesn't name the institution its grant doesn't approve (D-185):\n%s", row)
	}
	if n := strings.Count(body, ">Market data<"); n != 1 {
		t.Errorf("%d plugins marked as market data, want 1", n)
	}
}

// TestPluginsPageDeliveries checks that the plugins page shows the owner
// which bond terms and index levels a connector delivers are stored, as
// its grant approves, and which are dropped for want of approval (D-179).
func TestPluginsPageDeliveries(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/plugins", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	body := rec.Body.String()
	_, after, found := strings.Cut(body, "prices-main")
	row, _, _ := strings.Cut(after, "</tr>")
	if !found || !strings.Contains(row, ">Index levels<") || !strings.Contains(row, "Delivers treasury bond series terms; not approved, so they are dropped.") {
		t.Errorf("prices-main's row doesn't show its index levels stored and its bond terms dropped:\n%s", row)
	}
	if n := strings.Count(body, ">Bond terms<"); n != 0 {
		t.Errorf("%d plugins marked as delivering bond terms, want none approved", n)
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
