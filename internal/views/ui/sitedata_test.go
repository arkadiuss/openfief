package ui

import (
	"net/http"
	"strings"
	"testing"
)

// TestSessionEndClearsSiteData checks the server's side of wiping what the
// browser keeps of the owner's record when a session ends (D-151): sign-out
// and a session the core rejects, for a page or an htmx request, answer
// with Clear-Site-Data "storage"; a page with a valid session doesn't.
func TestSessionEndClearsSiteData(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, tc := range []struct {
		name  string
		req   *http.Request
		clear string
	}{
		{name: "sign-out", req: withSession(request(http.MethodPost, "/logout", ""), ownerSession), clear: `"storage"`},
		{name: "expired or revoked session", req: withSession(request(http.MethodGet, "/", ""), "session-expired"), clear: `"storage"`},
		{name: "expired session, htmx", req: htmx(withSession(request(http.MethodGet, "/allocation", ""), "session-expired")), clear: `"storage"`},
		{name: "overview", req: withSession(request(http.MethodGet, "/", ""), ownerSession)},
	} {
		rec := h.serve(t, tc.req)
		if got := rec.Header().Get("Clear-Site-Data"); got != tc.clear {
			t.Errorf("%s: Clear-Site-Data = %q, want %q", tc.name, got, tc.clear)
		}
	}
}

// TestHTMXKeepsNoHistory checks that every page turns htmx's history cache
// off (D-151): with hx-push-url it would keep snapshots of the pages it
// pushed into the history, the owner's record among them, in
// sessionStorage; a page missing from the history is fetched again.
func TestHTMXKeepsNoHistory(t *testing.T) {
	t.Parallel()
	p, err := parsePages()
	if err != nil {
		t.Fatal(err)
	}
	for name, set := range p.sets {
		src := set.Lookup("layout.html").Tree.Root.String()
		if !strings.Contains(src, `"historyCacheSize":0,"refreshOnHistoryMiss":true`) {
			t.Errorf("page %s lets htmx keep pages in its history cache", name)
		}
	}
}
