package ui

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestLoginPage checks the login page: the passkey button always, "View
// demo" only when the core says demo sessions work, and a page that still
// renders, saying so, when the core can't answer.
func TestLoginPage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		demo     bool
		down     bool
		want     []string
		wantNot  []string
		wantCall bool
	}{
		{name: "passkey only", want: []string{"Sign in with passkey", `data-webauthn="login"`}, wantNot: []string{"View demo"}},
		{name: "demo sessions offered", demo: true, want: []string{"Sign in with passkey", "View demo", `action="/login/demo"`}},
		{name: "core down", down: true, want: []string{"could not say how you can sign in"}, wantNot: []string{"View demo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.auth.locked(func() { h.auth.demo, h.auth.down = tt.demo, tt.down })
			rec := h.serve(t, request(http.MethodGet, "/login", ""))
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, tt.want...)
			wantNoBody(t, rec, tt.wantNot...)
			if diff := cmp.Diff([]string{testOrigin}, seen(h.auth, func() []string { return h.auth.origins })); diff != "" {
				t.Errorf("origins sent to the core (-want +got):\n%s", diff)
			}
		})
	}
}

// TestAuthServiceGetsPluginToken checks that every AuthService call carries
// the plugin's own token as the bearer, read from its file on each call so
// a rotated token is used at once, and never the session token.
func TestAuthServiceGetsPluginToken(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.serve(t, request(http.MethodGet, "/login", ""))
	h.writeToken(t, "plugin-token-rotated")
	h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	want := []string{"Bearer " + pluginToken, "Bearer plugin-token-rotated"}
	if diff := cmp.Diff(want, seen(h.auth, func() []string { return h.auth.bearers })); diff != "" {
		t.Errorf("Authorization headers at the AuthService (-want +got):\n%s", diff)
	}
}

// TestDemoLogin checks "View demo": a read-scoped demo session in a
// __Host- cookie that is HttpOnly, Secure and SameSite=Strict, then the
// overview; and an error page when the core offers no demo sessions.
func TestDemoLogin(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.demo = true })
	rec := h.serve(t, request(http.MethodPost, "/login/demo", ""))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/" {
		t.Errorf("Location = %q, want /", got)
	}
	c := responseCookie(rec)
	if c == nil {
		t.Fatalf("no %s cookie in %v", cookieName, rec.Header().Values("Set-Cookie"))
	}
	want := cookieAttrs{Value: demoSession, Path: "/", HTTPOnly: true, Secure: true, SameSite: http.SameSiteStrictMode}
	if diff := cmp.Diff(want, attrsOf(c)); diff != "" {
		t.Errorf("session cookie (-want +got):\n%s", diff)
	}

	h.auth.locked(func() { h.auth.demo = false })
	rec = h.serve(t, request(http.MethodPost, "/login/demo", ""))
	wantStatus(t, rec, http.StatusForbidden)
	if responseCookie(rec) != nil {
		t.Error("a refused demo login set a cookie")
	}
}

// cookieAttrs are the attributes of a cookie a test compares.
type cookieAttrs struct {
	Value    string
	Path     string
	HTTPOnly bool
	Secure   bool
	SameSite http.SameSite
}

func attrsOf(c *http.Cookie) cookieAttrs {
	return cookieAttrs{Value: c.Value, Path: c.Path, HTTPOnly: c.HttpOnly, Secure: c.Secure, SameSite: c.SameSite}
}

// TestLocalhostCookie checks the session cookie on http://localhost, the
// first login through kubectl port-forward (D-60): the same __Host- cookie,
// Secure, HttpOnly and SameSite=Strict, which browsers keep on localhost as
// a secure context; and the origin the core gets is the localhost one.
func TestLocalhostCookie(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.demo = true })
	r := request(http.MethodPost, "/login/demo", "")
	r.Host = "localhost:8080"
	r.Header.Del("X-Forwarded-Proto")
	rec := h.serve(t, r)
	wantStatus(t, rec, http.StatusSeeOther)
	c := responseCookie(rec)
	if c == nil {
		t.Fatalf("no %s cookie in %v", cookieName, rec.Header().Values("Set-Cookie"))
	}
	want := cookieAttrs{Value: demoSession, Path: "/", HTTPOnly: true, Secure: true, SameSite: http.SameSiteStrictMode}
	if diff := cmp.Diff(want, attrsOf(c)); diff != "" {
		t.Errorf("session cookie (-want +got):\n%s", diff)
	}
	origins := seen(h.auth, func() []string { return h.auth.origins })
	if got := origins[len(origins)-1]; got != localOrigin {
		t.Errorf("origin sent to the core = %q, want %q", got, localOrigin)
	}
}

// TestPasskeyLogin checks the passkey login relay: begin returns the
// core's ceremony and options untouched, finish hands the credential JSON
// to the core and keeps the new session in the cookie.
func TestPasskeyLogin(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, request(http.MethodPost, "/login/passkey/begin", "{}"))
	wantStatus(t, rec, http.StatusOK)
	var begin ceremonyStart
	if err := json.Unmarshal(rec.Body.Bytes(), &begin); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(ceremonyStart{CeremonyID: "c-login", Options: `{"challenge":"AAAA"}`}, begin); diff != "" {
		t.Errorf("begin (-want +got):\n%s", diff)
	}
	assertion := `{"id":"cred","response":{}}`
	body, err := json.Marshal(ceremonyEnd{CeremonyID: "c-login", Credential: assertion})
	if err != nil {
		t.Fatal(err)
	}
	rec = h.serve(t, request(http.MethodPost, "/login/passkey/finish", string(body)))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `"redirect":"/"`)
	if c := responseCookie(rec); c == nil || c.Value != newSession {
		t.Errorf("session cookie = %v, want the new session", c)
	}
	if diff := cmp.Diff([]string{assertion}, seen(h.auth, func() []string { return h.auth.credentials })); diff != "" {
		t.Errorf("credentials at the core (-want +got):\n%s", diff)
	}
}

// TestPasskeyLoginRefused checks that a refused assertion is a JSON error
// with the core's reason and sets no cookie, and a malformed body is 400.
func TestPasskeyLoginRefused(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, request(http.MethodPost, "/login/passkey/finish", `{"ceremonyId":"other","credential":"{}"}`))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, `"error"`)
	if responseCookie(rec) != nil {
		t.Error("a refused login set a cookie")
	}
	rec = h.serve(t, request(http.MethodPost, "/login/passkey/finish", `{"unknown":1}`))
	wantStatus(t, rec, http.StatusBadRequest)
}

// TestEnrollment checks the first passkey from an enrollment link (D-60):
// the page carries no code (it stays in the fragment), begin hands the code
// the script posts to BeginRegistration, and finish logs the owner in.
func TestEnrollment(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, request(http.MethodGet, "/enroll", ""))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `data-webauthn="register"`, `/static/webauthn.js`)

	rec = h.serve(t, request(http.MethodPost, "/enroll/begin", `{"code":"synthetic-code"}`))
	wantStatus(t, rec, http.StatusOK)
	code, sess := seenOne(h.auth, func() string { return h.auth.enrollCode }), seenOne(h.auth, func() string { return h.auth.regSession })
	if code != "synthetic-code" || sess != "" {
		t.Errorf("BeginRegistration got code %q and session %q, want only the code", code, sess)
	}
	rec = h.serve(t, request(http.MethodPost, "/enroll/finish", `{"ceremonyId":"c-reg","credential":"{}","name":"laptop"}`))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `"redirect":"/"`)
	if c := responseCookie(rec); c == nil || c.Value != newSession {
		t.Errorf("session cookie = %v, want the new session", c)
	}
}

// TestRegisterWithSession checks adding a passkey on this host from the
// owner's session: BeginRegistration gets the session, not a code, and
// finishing goes back to the settings without a new cookie; without a code
// or a session the request is refused before reaching the core.
func TestRegisterWithSession(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodPost, "/enroll/begin", `{}`), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	code, sess := seenOne(h.auth, func() string { return h.auth.enrollCode }), seenOne(h.auth, func() string { return h.auth.regSession })
	if sess != ownerSession || code != "" {
		t.Errorf("BeginRegistration got session %q and code %q, want only the session", sess, code)
	}
	rec = h.serve(t, withSession(request(http.MethodPost, "/enroll/finish", `{"ceremonyId":"c-reg","credential":"{}"}`), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `"redirect":"/settings"`)
	if responseCookie(rec) != nil {
		t.Error("a registration from a session set a new cookie")
	}

	rec = h.serve(t, request(http.MethodPost, "/enroll/begin", `{}`))
	wantStatus(t, rec, http.StatusBadRequest)
}

// TestLogout checks that logging out revokes the session at the core and
// drops the cookie, and drops it even without one.
func TestLogout(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodPost, "/logout", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
	if c := responseCookie(rec); c == nil || c.MaxAge >= 0 || c.Value != "" {
		t.Errorf("cookie = %v, want it deleted", c)
	}
	if diff := cmp.Diff([]string{ownerSession}, seen(h.auth, func() []string { return h.auth.loggedOut })); diff != "" {
		t.Errorf("sessions logged out (-want +got):\n%s", diff)
	}
}

// TestCrossOriginRefused checks the CSRF defense: a state-changing request
// another site sent, or one with neither Sec-Fetch-Site nor a matching
// Origin, is refused before any handler runs; a same-origin one passes.
func TestCrossOriginRefused(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		header map[string]string
		want   int
	}{
		{name: "cross-site", header: map[string]string{"Sec-Fetch-Site": "cross-site"}, want: http.StatusForbidden},
		{name: "same-site but another origin", header: map[string]string{"Sec-Fetch-Site": "same-site"}, want: http.StatusForbidden},
		{name: "foreign Origin only", header: map[string]string{"Origin": "https://evil.example.org"}, want: http.StatusForbidden},
		{name: "neither header", want: http.StatusForbidden},
		{name: "own Origin only", header: map[string]string{"Origin": testOrigin}, want: http.StatusSeeOther},
		{name: "same-origin", header: map[string]string{"Sec-Fetch-Site": "same-origin"}, want: http.StatusSeeOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			r := withSession(request(http.MethodPost, "/logout", ""), ownerSession)
			r.Header.Del("Sec-Fetch-Site")
			for k, v := range tt.header {
				r.Header.Set(k, v)
			}
			rec := h.serve(t, r)
			wantStatus(t, rec, tt.want)
			if called := len(seen(h.auth, func() []string { return h.auth.loggedOut })) > 0; called != (tt.want != http.StatusForbidden) {
				t.Errorf("Logout reached the core: %t", called)
			}
		})
	}
}

// TestSecurityHeaders checks the headers of a page and a static file: the
// strict CSP (self only, nothing inline, no framing), nosniff, no Referer,
// HSTS on HTTPS, and no caching of pages.
func TestSecurityHeaders(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, path := range []string{"/login", "/static/app.js", "/static/design/tokens.css"} {
		rec := h.serve(t, request(http.MethodGet, path, ""))
		wantStatus(t, rec, http.StatusOK)
		csp := rec.Header().Get("Content-Security-Policy")
		for _, d := range []string{"default-src 'none'", "script-src 'self'", "manifest-src 'self'", "style-src 'self'", "font-src 'self'", "img-src 'self'", "frame-ancestors 'none'"} {
			if !strings.Contains(csp, d) {
				t.Errorf("%s: CSP %q lacks %q", path, csp, d)
			}
		}
		if strings.Contains(csp, "unsafe") {
			t.Errorf("%s: CSP %q allows something unsafe", path, csp)
		}
		got := []string{
			rec.Header().Get("X-Content-Type-Options"), rec.Header().Get("X-Frame-Options"),
			rec.Header().Get("Referrer-Policy"), rec.Header().Get("Strict-Transport-Security"),
		}
		want := []string{"nosniff", "DENY", "no-referrer", hstsValue}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("%s: headers (-want +got):\n%s", path, diff)
		}
	}
	if got := h.serve(t, request(http.MethodGet, "/login", "")).Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("page Cache-Control = %q, want no-store", got)
	}
}

// TestNoInlineScriptsOrStyles checks every page template against the CSP:
// no inline script, no style element or attribute, no inline handler.
func TestNoInlineScriptsOrStyles(t *testing.T) {
	t.Parallel()
	p, err := parsePages()
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(p.sets))
	for name, set := range p.sets {
		names = append(names, name)
		for _, tmpl := range set.Templates() {
			src := tmpl.Tree.Root.String()
			for _, bad := range []string{"<style", " style=", " onclick=", " onload=", "javascript:", "<script>"} {
				if strings.Contains(src, bad) {
					t.Errorf("page %s, template %s has %q", name, tmpl.Name(), bad)
				}
			}
		}
	}
	slices.Sort(names)
	want := []string{"account", "accounts", "assertions", "enroll", "error", "login", "offline", "overview", "plugins", "price-checks", "settings", "setup", "sleeves"}
	if diff := cmp.Diff(want, names); diff != "" {
		t.Errorf("pages (-want +got):\n%s", diff)
	}
}
