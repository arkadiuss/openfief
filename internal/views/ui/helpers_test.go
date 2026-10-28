package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// request builds a browser request to the UI at testOrigin, behind a proxy
// that terminates TLS. A state-changing one carries Sec-Fetch-Site:
// same-origin, as a browser sends it from the UI's own page.
func request(method, target, body string) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	r.Host = testHost
	r.Header.Set("X-Forwarded-Proto", "https")
	if !safeMethod(method) {
		r.Header.Set("Sec-Fetch-Site", "same-origin")
	}
	return r
}

// form builds a same-origin form POST.
func form(target, body string) *http.Request {
	r := request(http.MethodPost, target, body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

// withSession adds a session cookie to r.
func withSession(r *http.Request, token string) *http.Request {
	r.AddCookie(&http.Cookie{Name: cookieName, Value: token, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	return r
}

// htmx marks r as an htmx request.
func htmx(r *http.Request) *http.Request {
	r.Header.Set(hxRequest, "true")
	return r
}

// serve runs r through the UI.
func (h *harness) serve(t *testing.T, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, r)
	return rec
}

// wantStatus fails t unless rec has status.
func wantStatus(t *testing.T, rec *httptest.ResponseRecorder, status int) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d; body:\n%s", rec.Code, status, rec.Body.String())
	}
}

// wantBody fails t unless the body contains every one of parts.
func wantBody(t *testing.T, rec *httptest.ResponseRecorder, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(rec.Body.String(), p) {
			t.Errorf("body lacks %q; body:\n%s", p, rec.Body.String())
		}
	}
}

// wantNoBody fails t if the body contains any of parts.
func wantNoBody(t *testing.T, rec *httptest.ResponseRecorder, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if strings.Contains(rec.Body.String(), p) {
			t.Errorf("body has %q", p)
		}
	}
}

// locker is a fake whose state tests reach under its lock.
type locker interface{ locked(fn func()) }

// seen reads a list a fake recorded, under its lock.
func seen(f locker, get func() []string) []string {
	var v []string
	f.locked(func() { v = append(v, get()...) })
	return v
}

// seenOne reads one value a fake recorded, under its lock.
func seenOne(f locker, get func() string) string {
	var v string
	f.locked(func() { v = get() })
	return v
}

// responseCookie returns the session cookie the response sets, or nil.
func responseCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, line := range rec.Header().Values("Set-Cookie") {
		c, err := http.ParseSetCookie(line)
		if err == nil && c.Name == cookieName {
			return c
		}
	}
	return nil
}
