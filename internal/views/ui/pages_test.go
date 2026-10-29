package ui

import (
	"net/http"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// TestPagesNeedSession checks that every page of the record sends a browser
// without a session to the login page, and an htmx request gets
// HX-Redirect instead of a redirect it would swap in.
func TestPagesNeedSession(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, path := range []string{"/", "/accounts/" + accountID, "/assertions", "/settings", "/allocation?by=currency"} {
		rec := h.serve(t, request(http.MethodGet, path, ""))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != "/login" {
			t.Errorf("%s: Location = %q, want /login", path, got)
		}
	}
	rec := h.serve(t, htmx(request(http.MethodGet, "/allocation?by=currency", "")))
	wantStatus(t, rec, http.StatusUnauthorized)
	if got := rec.Header().Get(hxRedirect); got != "/login" {
		t.Errorf("HX-Redirect = %q, want /login", got)
	}
	if n := len(seen(h.read, func() []string { return h.read.bearers })); n != 0 {
		t.Errorf("the read API was called %d times without a session", n)
	}
}

// TestExpiredSession checks that a session the core no longer knows drops
// the cookie and goes to the login page (UNAUTHENTICATED in the contract).
func TestExpiredSession(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), "session-expired"))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
	if c := responseCookie(rec); c == nil || c.MaxAge >= 0 {
		t.Errorf("cookie = %v, want it deleted", c)
	}
}

// TestOverview checks the overview: the total in the reporting currency,
// each account with its valuation status (D-36) and holdings value, the
// allocation chart and table (D-30) on the latest valued-through date, the
// value-over-time chart as the time-weighted return from go-live (D-61),
// no model portfolio while it is off (D-52), and every read with the
// session token as the bearer token, never the plugin's.
func TestOverview(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"Total on 2026-03-31", "1 500.00 PLN",
		"Synthetic brokerage", `valued through <span class="mono">2026-03-31</span>`, "Revaluation pending", "1 000.00 PLN",
		`id="allocation"`, "66.7%", "33.3%",
		`data-chart="{&#34;kind&#34;:&#34;bar&#34;`, `&#34;values&#34;:[&#34;1000.00&#34;,&#34;500.00&#34;]`,
		`data-chart="{&#34;kind&#34;:&#34;line&#34;`, "From go-live (2026-03-30) to 2026-03-31",
		"time-weighted return as an index, 100 on 2026-03-30", `Its value with cash on 2026-03-31 is <span class="mono">1`,
		`&#34;values&#34;:[&#34;100&#34;,&#34;101.5&#34;]`,
		"/static/vendor/chart-4.5.1.umd.min.js", "/static/app.js",
		`<a href="/assertions">1 flagged balance assertion(s)</a>`,
	)
	wantNoBody(t, rec, "model portfolio", "Model", "TWR: not built", "series")
	bearers := seen(h.read, func() []string { return h.read.bearers })
	if len(bearers) == 0 || slices.ContainsFunc(bearers, func(b string) bool { return b != "Bearer "+ownerSession }) {
		t.Errorf("read API bearers = %v, want the session token on every call", bearers)
	}
	var dims []viewsv1.AllocationDimension
	h.read.locked(func() { dims = append(dims, h.read.dimensions...) })
	if diff := cmp.Diff([]viewsv1.AllocationDimension{viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS}, dims); diff != "" {
		t.Errorf("allocation dimensions (-want +got):\n%s", diff)
	}
}

// TestOverviewDate checks the date query parameter: it sets the
// allocation's date, and a malformed one is 400.
func TestOverviewDate(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/?date=2026-02-27&by=sleeve", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	got := seen(h.read, func() []string { return h.read.dates })
	if diff := cmp.Diff([]string{"2026-02-27"}, got); diff != "" {
		t.Errorf("allocation dates (-want +got):\n%s", diff)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/?date=27.02.2026", ""), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
}

// TestAllocationSwitch checks the dimension switch (D-30): each dimension's
// link, and the htmx fragment for another dimension, which is only the
// allocation section; without htmx the browser gets the whole overview.
func TestAllocationSwitch(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	for _, d := range dimensions() {
		wantBody(t, rec, `hx-get="/allocation?by=`+d.Key+`&amp;date=2026-03-31"`)
	}
	rec = h.serve(t, htmx(withSession(request(http.MethodGet, "/allocation?by=currency&date=2026-03-31", ""), ownerSession)))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<section id="allocation"`, `aria-current="true">Currency</a>`)
	wantNoBody(t, rec, "<html", "Value over time")
	var dims []viewsv1.AllocationDimension
	h.read.locked(func() { dims = append(dims, h.read.dimensions...) })
	if got := dims[len(dims)-1]; got != viewsv1.AllocationDimension_ALLOCATION_DIMENSION_CURRENCY {
		t.Errorf("dimension asked for = %v, want currency", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/allocation?by=currency&date=2026-03-31", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/?by=currency&date=2026-03-31" {
		t.Errorf("Location = %q", got)
	}
}

// TestAccountPage checks an account's page: holdings with their stored
// values, cash, open lots, the flagged assertion, and the first page of
// transactions with a "More" link to the next one.
func TestAccountPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"Synthetic brokerage", `valued through <span class="mono">2026-03-31</span>`,
		"SYN (XX0000000001)", "250.00 EUR", "1 000.00 PLN", "market",
		"12 345.67 EUR",
		"2026-01-02", "2 000.00 EUR", "opening lot",
		"cash EUR", "flagged",
		"Synthetic deposit", "deposit",
		`hx-get="/accounts/acc-synthetic-1/transactions?page_token=page-2"`,
	)
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts/unknown", ""), readSession))
	wantStatus(t, rec, http.StatusNotFound)
}

// TestTransactionsPaging checks "More": htmx gets only the next page's rows,
// with no further link on the last page, read with the page token.
func TestTransactionsPaging(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/accounts/"+accountID+"/transactions?page_token=page-2", ""), readSession)))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Synthetic deposit")
	wantNoBody(t, rec, "<html", "More transactions")
	if diff := cmp.Diff([]string{"page-2"}, seen(h.read, func() []string { return h.read.pageTokens })); diff != "" {
		t.Errorf("page tokens (-want +got):\n%s", diff)
	}
}

// TestAssertionsPage checks the list of flagged balance assertions.
func TestAssertionsPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/assertions", ""), demoSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Synthetic brokerage", "2026-03-15", "cash EUR", "100.00", "90.00", "flagged")
}

// TestReadScopeHidesOwnerActions checks that a read-scoped session, such as
// the demo's, sees no settings link or sign-in management and gets 403 on
// the owner's pages, which never reach the core; the owner sees the link.
func TestReadScopeHidesOwnerActions(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), demoSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "read access")
	wantNoBody(t, rec, `href="/settings"`)
	for _, r := range []*http.Request{
		request(http.MethodGet, "/settings", ""),
		form("/settings/tokens", "name=x"),
		form("/settings/passkeys/pk1/remove", ""),
		form("/settings/tokens/fedcba9876543210/revoke", ""),
		form("/settings/enrollment-links", ""),
	} {
		rec = h.serve(t, withSession(r, demoSession))
		wantStatus(t, rec, http.StatusForbidden)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantBody(t, rec, `href="/settings"`, "owner access")
}

// TestStaticFiles checks that the vendored libraries and the UI's own
// scripts and stylesheet are served from the UI with their content types.
func TestStaticFiles(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for path, ctype := range map[string]string{
		"/static/vendor/htmx-2.0.11.min.js":     "text/javascript; charset=utf-8",
		"/static/vendor/chart-4.5.1.umd.min.js": "text/javascript; charset=utf-8",
		"/static/app.js":                        "text/javascript; charset=utf-8",
		"/static/webauthn.js":                   "text/javascript; charset=utf-8",
		"/static/style.css":                     "text/css; charset=utf-8",
	} {
		rec := h.serve(t, request(http.MethodGet, path, ""))
		wantStatus(t, rec, http.StatusOK)
		if got := rec.Header().Get("Content-Type"); got != ctype {
			t.Errorf("%s: Content-Type = %q, want %q", path, got, ctype)
		}
	}
	wantStatus(t, h.serve(t, request(http.MethodGet, "/healthz", "")), http.StatusOK)
}
