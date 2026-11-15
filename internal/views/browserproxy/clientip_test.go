package browserproxy_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/browserproxy"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

func TestParseClientIPConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, header, proxies string
		wantHeader            string
		wantErr               bool
	}{
		{name: "none believes no header"},
		{name: "X-Forwarded-For from the pod network", header: "x-forwarded-for", proxies: "10.42.0.0/16", wantHeader: "X-Forwarded-For"},
		{name: "a single-address header from two proxies", header: "CF-Connecting-IP", proxies: "10.42.0.0/16, 192.0.2.9", wantHeader: "Cf-Connecting-Ip"},
		{name: "an IPv6 proxy", header: "X-Forwarded-For", proxies: "fd00::/8", wantHeader: "X-Forwarded-For"},
		{name: "a header with no proxy to trust", header: "X-Forwarded-For", wantErr: true},
		{name: "proxies without a header", proxies: "10.0.0.0/8", wantErr: true},
		{name: "a proxy that is neither address nor network", header: "X-Forwarded-For", proxies: "ingress", wantErr: true},
		{name: "a network with a bad length", header: "X-Forwarded-For", proxies: "10.0.0.0/40", wantErr: true},
		{name: "Forwarded is not supported", header: "Forwarded", proxies: "10.0.0.0/8", wantErr: true},
		{name: "a header name with a space", header: "X Forwarded", proxies: "10.0.0.0/8", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := browserproxy.ParseClientIPConfig(tt.header, tt.proxies)
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, browserproxy.ErrClientIPConfig)) {
				t.Fatalf("error = %v, want error: %v", err, tt.wantErr)
			}
			if got.Header != tt.wantHeader {
				t.Errorf("header = %q, want %q", got.Header, tt.wantHeader)
			}
		})
	}
}

func TestResolveTheClientAddress(t *testing.T) {
	t.Parallel()
	forwarded := mustConfig(t, "X-Forwarded-For", "10.42.0.0/16")
	single := mustConfig(t, "CF-Connecting-IP", "10.42.0.0/16")
	tests := []struct {
		name    string
		cfg     browserproxy.ClientIPConfig
		remote  string
		headers map[string][]string
		want    string
	}{
		{name: "no setting: the connection's address", remote: "198.51.100.7:5000", want: "198.51.100.7"},
		{name: "no setting ignores every header", remote: "198.51.100.7:5000", headers: map[string][]string{"X-Forwarded-For": {"203.0.113.1"}}, want: "198.51.100.7"},
		{name: "an IPv6 connection", remote: "[2001:db8::1]:5000", want: "2001:db8::1"},
		{name: "an IPv4-mapped connection", remote: "[::ffff:198.51.100.7]:5000", want: "198.51.100.7"},
		{name: "an unreadable peer", remote: "not-an-address", want: ""},
		{
			name: "a trusted proxy's list: the browser's address", cfg: forwarded, remote: "10.42.0.9:8080",
			headers: map[string][]string{"X-Forwarded-For": {"198.51.100.7"}}, want: "198.51.100.7",
		},
		{
			name: "a browser that writes its own list gains nothing: the rightmost untrusted wins", cfg: forwarded, remote: "10.42.0.9:8080",
			headers: map[string][]string{"X-Forwarded-For": {"203.0.113.99, 198.51.100.7"}}, want: "198.51.100.7",
		},
		{
			name: "proxies behind proxies are skipped", cfg: forwarded, remote: "10.42.0.9:8080",
			headers: map[string][]string{"X-Forwarded-For": {"198.51.100.7, 10.42.1.1, 10.42.2.2"}}, want: "198.51.100.7",
		},
		{
			name: "the list across several header lines", cfg: forwarded, remote: "10.42.0.9:8080",
			headers: map[string][]string{"X-Forwarded-For": {"203.0.113.99", "198.51.100.7, 10.42.1.1"}}, want: "198.51.100.7",
		},
		{
			name: "an untrusted peer's header is never believed", cfg: forwarded, remote: "203.0.113.50:5000",
			headers: map[string][]string{"X-Forwarded-For": {"198.51.100.7"}}, want: "203.0.113.50",
		},
		{
			name: "a trusted proxy without the header is unknown, not the proxy", cfg: forwarded, remote: "10.42.0.9:8080", want: "",
		},
		{
			name: "a list of trusted proxies only is unknown", cfg: forwarded, remote: "10.42.0.9:8080",
			headers: map[string][]string{"X-Forwarded-For": {"10.42.1.1, 10.42.2.2"}}, want: "",
		},
		{
			name: "a malformed entry ends the search", cfg: forwarded, remote: "10.42.0.9:8080",
			headers: map[string][]string{"X-Forwarded-For": {"198.51.100.7, garbage"}}, want: "",
		},
		{
			name: "an IPv6 browser", cfg: forwarded, remote: "10.42.0.9:8080",
			headers: map[string][]string{"X-Forwarded-For": {"2001:db8:1::5"}}, want: "2001:db8:1::5",
		},
		{
			name: "a single-address header from a trusted proxy", cfg: single, remote: "10.42.0.9:8080",
			headers: map[string][]string{"Cf-Connecting-Ip": {"198.51.100.7"}}, want: "198.51.100.7",
		},
		{
			name: "a single-address header that is not an address", cfg: single, remote: "10.42.0.9:8080",
			headers: map[string][]string{"Cf-Connecting-Ip": {"unknown"}}, want: "",
		},
		{
			name: "a single-address header from an untrusted peer", cfg: single, remote: "203.0.113.50:5000",
			headers: map[string][]string{"Cf-Connecting-Ip": {"198.51.100.7"}}, want: "203.0.113.50",
		},
		{
			name: "the other header is not the configured one", cfg: single, remote: "10.42.0.9:8080",
			headers: map[string][]string{"X-Forwarded-For": {"198.51.100.7"}}, want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:8080/", http.NoBody)
			req.RemoteAddr = tt.remote
			for k, vs := range tt.headers {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
			if got := tt.cfg.Resolve(req); got != tt.want {
				t.Errorf("Resolve = %q, want %q", got, tt.want)
			}
		})
	}
}

func mustConfig(t *testing.T, header, proxies string) browserproxy.ClientIPConfig {
	t.Helper()
	c, err := browserproxy.ParseClientIPConfig(header, proxies)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// proxied is what came of one request through the proxy: the value of the SDK's
// ClientIPHeader the plugin saw, whether it saw the header at all, and the
// proxy's response.
type proxied struct {
	ip   string
	set  bool
	resp *httptest.ResponseRecorder
}

// proxiedTo serves one request through a proxy made with opts to a plugin that
// records the SDK's ClientIPHeader.
func proxiedTo(t *testing.T, remote string, headers map[string]string, opts ...browserproxy.Option) proxied {
	t.Helper()
	var got proxied
	plugin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, got.set = r.Header[http.CanonicalHeaderKey(openfief.ClientIPHeader)]
		got.ip = r.Header.Get(openfief.ClientIPHeader)
		if _, err := io.WriteString(w, "ok"); err != nil {
			t.Errorf("writing: %v", err)
		}
	}))
	t.Cleanup(plugin.Close)
	door := &app.FrontDoor{
		Views:  fakeViews{routes: []views.BrowserRoute{{Instance: "ui", Prefix: "/"}}, urls: map[string]string{"ui": plugin.URL}},
		Tokens: fakeTokens{},
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:8080/ask", http.NoBody)
	req.RemoteAddr = remote
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	got.resp = httptest.NewRecorder()
	browserproxy.New(door, nil, opts...).ServeHTTP(got.resp, req)
	return got
}

func TestProxyHandsThePluginTheAddressItFoundAndNoOtherValue(t *testing.T) {
	t.Parallel()
	behindIngress := browserproxy.WithClientIP(mustConfig(t, "X-Forwarded-For", "10.42.0.0/16"))
	tests := []struct {
		name    string
		remote  string
		headers map[string]string
		opts    []browserproxy.Option
		want    string
		wantSet bool
	}{
		{name: "no setting: the connection's address", remote: "198.51.100.7:5000", want: "198.51.100.7", wantSet: true},
		{
			name: "a browser's own value is replaced", remote: "198.51.100.7:5000",
			headers: map[string]string{openfief.ClientIPHeader: "203.0.113.99"}, want: "198.51.100.7", wantSet: true,
		},
		{
			name: "behind the trusted Ingress", remote: "10.42.0.9:8080", opts: []browserproxy.Option{behindIngress},
			headers: map[string]string{"X-Forwarded-For": "198.51.100.7"}, want: "198.51.100.7", wantSet: true,
		},
		{
			name: "a browser that reaches the core directly can't claim an address", remote: "203.0.113.50:5000", opts: []browserproxy.Option{behindIngress},
			headers: map[string]string{"X-Forwarded-For": "198.51.100.7", openfief.ClientIPHeader: "198.51.100.8"}, want: "203.0.113.50", wantSet: true,
		},
		{
			name: "a trusted proxy that says nothing leaves the header out, whatever the browser sent", remote: "10.42.0.9:8080",
			opts: []browserproxy.Option{behindIngress}, headers: map[string]string{openfief.ClientIPHeader: "198.51.100.8"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := proxiedTo(t, tt.remote, tt.headers, tt.opts...)
			if got.set != tt.wantSet || got.ip != tt.want {
				t.Errorf("the plugin saw %q (set: %v), want %q (set: %v)", got.ip, got.set, tt.want, tt.wantSet)
			}
		})
	}
}

// wantPolicyOnPagesAndRefusals checks that a page the proxy serves and its
// refusal of a path no plugin serves carry exactly the content security policy
// want, once, and no other.
func wantPolicyOnPagesAndRefusals(t *testing.T, want string, opts ...browserproxy.Option) {
	t.Helper()
	got := proxiedTo(t, "198.51.100.7:5000", nil, opts...).resp.Header().Values("Content-Security-Policy")
	if len(got) != 1 || got[0] != want {
		t.Errorf("Content-Security-Policy = %q, want %q", got, want)
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:8080/", http.NoBody)
	rec := httptest.NewRecorder()
	browserproxy.New(&app.FrontDoor{Views: fakeViews{}, Tokens: fakeTokens{}}, nil, opts...).ServeHTTP(rec, req)
	if got := rec.Header().Values("Content-Security-Policy"); len(got) != 1 || got[0] != want {
		t.Errorf("a refusal's Content-Security-Policy = %q, want %q", got, want)
	}
}

func TestProxyWithoutTurnstileKeepsTheStrictPolicy(t *testing.T) {
	t.Parallel()
	wantPolicyOnPagesAndRefusals(t, browserproxy.ViewContentSecurityPolicy)
}

func TestProxyWithTurnstileAllowsCloudflaresScriptAndFrameOnItsPagesAndRefusals(t *testing.T) {
	t.Parallel()
	wantPolicyOnPagesAndRefusals(t, browserproxy.ViewContentSecurityPolicyWithTurnstile, browserproxy.WithTurnstile())
}

// TestTurnstilePolicyNamesOneHostAndNoLooseKeyword checks that Turnstile's
// origin is the only host in the widened policy: no wildcard, no unsafe
// keyword and no other site.
func TestTurnstilePolicyNamesOneHostAndNoLooseKeyword(t *testing.T) {
	t.Parallel()
	for directive := range strings.SplitSeq(browserproxy.ViewContentSecurityPolicyWithTurnstile, ";") {
		for _, field := range strings.Fields(directive)[1:] {
			named := strings.Contains(field, "://") && field != browserproxy.TurnstileOrigin
			if named || field == "*" || strings.Contains(field, "'unsafe-") {
				t.Errorf("directive %q names %s", strings.TrimSpace(directive), field)
			}
		}
	}
}
