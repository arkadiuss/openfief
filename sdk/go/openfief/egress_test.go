package openfief_test

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// fakeProxy is an HTTP CONNECT proxy that records each CONNECT's target and
// Proxy-Authorization, refuses one without a bearer token with 407, and
// tunnels the rest to the target.
type fakeProxy struct {
	mu       sync.Mutex
	connects []string
	auths    []string
	others   int
	// errs collects the tunnel's I/O errors, which end every tunnel when a
	// side closes; tests don't inspect them.
	errs []error
}

func (p *fakeProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	if r.Method != http.MethodConnect {
		p.others++
		p.mu.Unlock()
		http.Error(w, "only CONNECT", http.StatusMethodNotAllowed)
		return
	}
	auth := r.Header.Get("Proxy-Authorization")
	p.connects = append(p.connects, r.Host)
	p.auths = append(p.auths, auth)
	p.mu.Unlock()
	if auth == "" {
		w.WriteHeader(http.StatusProxyAuthRequired)
		return
	}
	upstream, err := (&net.Dialer{}).DialContext(r.Context(), "tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	conn, rw, err := http.NewResponseController(w).Hijack()
	if err != nil {
		p.note(errors.Join(err, upstream.Close()))
		return
	}
	_, err = rw.WriteString("HTTP/1.1 200 Connection established\r\n\r\n")
	p.note(err)
	p.note(rw.Flush())
	go p.pipe(upstream, conn)
	go p.pipe(conn, upstream)
}

// note records an I/O error of the tunnel.
func (p *fakeProxy) note(err error) {
	if err == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.errs = append(p.errs, err)
}

// pipe copies src to dst until either side closes, then closes dst.
func (p *fakeProxy) pipe(dst, src net.Conn) {
	_, err := io.Copy(dst, src)
	p.note(errors.Join(err, dst.Close()))
}

// proxyLog is what a fakeProxy received.
type proxyLog struct {
	connects []string
	auths    []string
	others   int
}

func (p *fakeProxy) seen() proxyLog {
	p.mu.Lock()
	defer p.mu.Unlock()
	return proxyLog{connects: slices.Clone(p.connects), auths: slices.Clone(p.auths), others: p.others}
}

// egressFixture is a TLS "external host", a fake egress proxy in front of it
// and an EgressTransport that trusts the host and reads its token from a file.
type egressFixture struct {
	host      *httptest.Server
	proxy     *fakeProxy
	tokenPath string
	transport *openfief.EgressTransport
}

func newEgressFixture(t *testing.T) *egressFixture {
	t.Helper()
	host := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(host.Close)
	proxy := &fakeProxy{}
	proxySrv := httptest.NewServer(proxy)
	t.Cleanup(proxySrv.Close)
	proxyURL, err := url.Parse(proxySrv.URL)
	if err != nil {
		t.Fatalf("parse proxy URL: %v", err)
	}
	hostTLS, ok := host.Client().Transport.(*http.Transport)
	if !ok {
		t.Fatal("httptest TLS client has no *http.Transport")
	}
	path := filepath.Join(t.TempDir(), "token")
	return &egressFixture{
		host: host, proxy: proxy, tokenPath: path,
		transport: &openfief.EgressTransport{
			Proxy: proxyURL, Tokens: openfief.TokenFile{Path: path},
			TLSClientConfig: &tls.Config{RootCAs: hostTLS.TLSClientConfig.RootCAs, MinVersion: tls.VersionTLS12},
		},
	}
}

func TestEgressSendsCurrentTokenOnEveryConnect(t *testing.T) {
	t.Parallel()
	f := newEgressFixture(t)
	client := &http.Client{Transport: f.transport}
	for _, token := range []string{"synthetic-token-1", "synthetic-token-2"} {
		writeToken(t, f.tokenPath, token)
		if err := get(t, client, f.host.URL); err != nil {
			t.Fatalf("GET through the proxy: %v", err)
		}
		client.CloseIdleConnections() // the next request opens a new tunnel
	}
	got := f.proxy.seen()
	hostAddr := f.host.Listener.Addr().String()
	if want := []string{hostAddr, hostAddr}; !slices.Equal(got.connects, want) {
		t.Errorf("CONNECT targets = %q, want %q", got.connects, want)
	}
	if want := []string{basic("synthetic-token-1"), basic("synthetic-token-2")}; !slices.Equal(got.auths, want) {
		t.Errorf("Proxy-Authorization = %q, want %q", got.auths, want)
	}
	if got.others != 0 {
		t.Errorf("the proxy got %d requests that were not CONNECT", got.others)
	}
}

func TestEgressRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		target  func(f *egressFixture) string
		noProxy bool
		noToken bool
		wantErr error
	}{
		{
			name:    "plain http is never sent",
			target:  func(f *egressFixture) string { return "http://" + f.host.Listener.Addr().String() },
			wantErr: openfief.ErrNotHTTPS,
		},
		{
			name:    "no proxy configured",
			target:  func(f *egressFixture) string { return f.host.URL },
			noProxy: true,
			wantErr: openfief.ErrNoEgressProxy,
		},
		{
			name:    "no token",
			target:  func(f *egressFixture) string { return f.host.URL },
			noToken: true,
			wantErr: openfief.ErrEmptyToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newEgressFixture(t)
			if tt.noProxy {
				f.transport.Proxy = nil
			}
			token := "synthetic-token-1"
			if tt.noToken {
				token = ""
			}
			writeToken(t, f.tokenPath, token)
			err := get(t, &http.Client{Transport: f.transport}, tt.target(f))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GET error = %v, want %v", err, tt.wantErr)
			}
			if got := f.proxy.seen(); len(got.connects) != 0 || got.others != 0 {
				t.Errorf("the proxy got %d CONNECTs and %d other requests, want none", len(got.connects), got.others)
			}
		})
	}
}

// basic returns the Proxy-Authorization value of a proxy URL whose password
// is token (D-63).
func basic(token string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(openfief.EgressProxyUser+":"+token))
}

func TestNewEgressClient(t *testing.T) {
	t.Setenv(openfief.EnvHTTPSProxy, "")
	t.Setenv(openfief.EnvEgressIdentityFile, "/var/run/secrets/openfief/egress-token")
	if _, err := openfief.NewEgressClient(); !errors.Is(err, openfief.ErrNoEgressProxy) {
		t.Errorf("without HTTPS_PROXY: error = %v, want %v", err, openfief.ErrNoEgressProxy)
	}
	t.Setenv(openfief.EnvHTTPSProxy, "http://10.0.0.1:4750")
	t.Setenv(openfief.EnvEgressIdentityFile, "")
	if _, err := openfief.NewEgressClient(); !errors.Is(err, openfief.ErrNoEgressToken) {
		t.Errorf("without OPENFIEF_EGRESS_TOKEN_FILE: error = %v, want %v", err, openfief.ErrNoEgressToken)
	}
	t.Setenv(openfief.EnvEgressIdentityFile, "/var/run/secrets/openfief/egress-token")
	client, err := openfief.NewEgressClient()
	if err != nil {
		t.Fatalf("NewEgressClient: %v", err)
	}
	tr, ok := client.Transport.(*openfief.EgressTransport)
	if !ok {
		t.Fatalf("transport is %T, want *openfief.EgressTransport", client.Transport)
	}
	if got := tr.Proxy.String(); got != "http://10.0.0.1:4750" {
		t.Errorf("proxy = %q, want http://10.0.0.1:4750", got)
	}
	if got, want := tr.Tokens, (openfief.TokenFile{Path: "/var/run/secrets/openfief/egress-token"}); got != want {
		t.Errorf("tokens = %v, want %v", got, want)
	}
}
