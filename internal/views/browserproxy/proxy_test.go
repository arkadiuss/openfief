package browserproxy_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/browserproxy"
)

// Synthetic failures of the fakes.
var (
	errNoToken   = errors.New("synthetic: no token")
	errNoAddress = errors.New("synthetic: no address")
)

// fakeViews serves fixed routes and addresses.
type fakeViews struct {
	routes []views.BrowserRoute
	urls   map[string]string
	err    error
}

func (f fakeViews) BrowserRoutes(context.Context) ([]views.BrowserRoute, error) {
	return f.routes, f.err
}

func (f fakeViews) BaseURL(instance string) (string, error) {
	u, ok := f.urls[instance]
	if !ok {
		return "", fmt.Errorf("%w for %s", errNoAddress, instance)
	}
	return u, nil
}

// fakeTokens issues "core-token-for-<instance>", or fails.
type fakeTokens struct{ err error }

func (f fakeTokens) TokenFor(_ context.Context, instance string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return "core-token-for-" + instance, nil
}

// seen is what the plugin received.
type seen struct {
	Path, Query, Host, Auth, Proto, FwdHost, FwdFor, Forwarded, Cookie, Body string
}

// pluginServer records each request and answers with a cookie and a body.
func pluginServer(t *testing.T, got *seen) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading the body: %v", err)
		}
		*got = seen{
			Path: r.URL.Path, Query: r.URL.RawQuery, Host: r.Host, Auth: r.Header.Get("Authorization"),
			Proto: r.Header.Get("X-Forwarded-Proto"), FwdHost: r.Header.Get("X-Forwarded-Host"),
			FwdFor: r.Header.Get("X-Forwarded-For"), Forwarded: r.Header.Get("Forwarded"),
			Cookie: r.Header.Get("Cookie"), Body: string(body),
		}
		http.SetCookie(w, &http.Cookie{Name: "__Host-openfief_session", Value: "synthetic", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
		w.WriteHeader(http.StatusTeapot)
		if _, err := io.WriteString(w, "from the plugin"); err != nil {
			t.Errorf("writing: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestProxyForwards(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		headers map[string]string
		want    seen
	}{
		{
			name: "through kubectl port-forward",
			headers: map[string]string{
				"Authorization": "Bearer stolen", "X-Forwarded-Host": "evil.example.com", "Forwarded": "proto=https",
				"X-Forwarded-For": "192.0.2.7", "Cookie": "__Host-openfief_session=ofsess_x",
			},
			want: seen{
				Path: "/accounts/a-1", Query: "date=2026-10-28", Host: "localhost:8080", Auth: "Bearer core-token-for-ui",
				Proto: "http", FwdHost: "localhost:8080", Cookie: "__Host-openfief_session=ofsess_x", Body: "mode=transactions",
			},
		},
		{
			name:    "behind the owner's Ingress on https",
			headers: map[string]string{"X-Forwarded-Proto": "https"},
			want: seen{
				Path: "/accounts/a-1", Query: "date=2026-10-28", Host: "localhost:8080", Auth: "Bearer core-token-for-ui",
				Proto: "https", FwdHost: "localhost:8080", Body: "mode=transactions",
			},
		},
		{
			name:    "an unknown scheme",
			headers: map[string]string{"X-Forwarded-Proto": "gopher"},
			want: seen{
				Path: "/accounts/a-1", Query: "date=2026-10-28", Host: "localhost:8080", Auth: "Bearer core-token-for-ui",
				Proto: "http", FwdHost: "localhost:8080", Body: "mode=transactions",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got seen
			plugin := pluginServer(t, &got)
			door := &app.FrontDoor{
				Views:  fakeViews{routes: []views.BrowserRoute{{Instance: "ui", Prefix: "/"}}, urls: map[string]string{"ui": plugin.URL}},
				Tokens: fakeTokens{},
			}
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://localhost:8080/accounts/a-1?date=2026-10-28",
				strings.NewReader("mode=transactions"))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			browserproxy.New(door, nil).ServeHTTP(rec, req)
			if rec.Code != http.StatusTeapot || rec.Body.String() != "from the plugin" || rec.Header().Get("Set-Cookie") == "" {
				t.Errorf("answer = %d %q, Set-Cookie %q; want the plugin's", rec.Code, rec.Body.String(), rec.Header().Get("Set-Cookie"))
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("the plugin saw (-want +got):\n%s", diff)
			}
		})
	}
}

func TestProxyRefuses(t *testing.T) {
	t.Parallel()
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	route := views.BrowserRoute{Instance: "ui", Prefix: "/"}
	ui := []views.BrowserRoute{route}
	tests := []struct {
		name string
		door *app.FrontDoor
		want int
	}{
		{name: "no route", door: &app.FrontDoor{Views: fakeViews{}, Tokens: fakeTokens{}}, want: http.StatusNotFound},
		{
			name: "two plugins on one prefix",
			door: &app.FrontDoor{Views: fakeViews{routes: []views.BrowserRoute{route, {Instance: "twin", Prefix: "/"}}}, Tokens: fakeTokens{}},
			want: http.StatusServiceUnavailable,
		},
		{
			name: "routes not read yet",
			door: &app.FrontDoor{Views: fakeViews{err: app.ErrViewsUnknown}, Tokens: fakeTokens{}},
			want: http.StatusServiceUnavailable,
		},
		{name: "no address", door: &app.FrontDoor{Views: fakeViews{routes: ui}, Tokens: fakeTokens{}}, want: http.StatusBadGateway},
		{
			name: "no token",
			door: &app.FrontDoor{Views: fakeViews{routes: ui, urls: map[string]string{"ui": closed.URL}}, Tokens: fakeTokens{err: errNoToken}},
			want: http.StatusBadGateway,
		},
		{
			name: "the plugin does not answer",
			door: &app.FrontDoor{Views: fakeViews{routes: ui, urls: map[string]string{"ui": closed.URL}}, Tokens: fakeTokens{}},
			want: http.StatusBadGateway,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:8080/login", http.NoBody)
			rec := httptest.NewRecorder()
			browserproxy.New(tt.door, nil).ServeHTTP(rec, req)
			if rec.Code != tt.want || rec.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("status = %d, Cache-Control %q; want %d, no-store", rec.Code, rec.Header().Get("Cache-Control"), tt.want)
			}
		})
	}
}

func TestBrowserScheme(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, header, want string
		tls                bool
	}{
		{name: "plain", want: "http"},
		{name: "https from the Ingress", header: "https", want: "https"},
		{name: "http from the Ingress", header: "http", want: "http"},
		{name: "junk", header: "https, http", want: "http"},
		{name: "TLS", tls: true, header: "http", want: "https"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:8080/", http.NoBody)
			if tt.header != "" {
				req.Header.Set("X-Forwarded-Proto", tt.header)
			}
			if tt.tls {
				req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "https://localhost:8443/", http.NoBody)
				req.Header.Set("X-Forwarded-Proto", tt.header)
			}
			if got := browserproxy.BrowserScheme(req); got != tt.want {
				t.Errorf("BrowserScheme = %q, want %q", got, tt.want)
			}
		})
	}
}
