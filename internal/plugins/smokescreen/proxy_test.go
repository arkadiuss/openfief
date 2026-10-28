package smokescreen_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	ss "github.com/stripe/smokescreen/pkg/smokescreen"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

// Synthetic tokens and addresses. The public-looking address is never
// dialed: the test dialer sends every connection to the local TLS server.
const (
	sampleToken = "synthetic-token-sample"
	publicAddr  = "8.8.4.4"
	privateAddr = "10.0.0.7"
)

var errUnknownHost = errors.New("no such host in the fake resolver")

// fakeReviewer accepts only sampleToken, as plugin "sample".
type fakeReviewer struct{}

func (fakeReviewer) ReviewToken(_ context.Context, token string) (string, error) {
	if token != sampleToken {
		return "", app.ErrUnauthenticated
	}
	return "system:serviceaccount:openfief:" + plugins.ServiceAccountName("sample"), nil
}

// fakeLists serves one fixed allow-list.
type fakeLists struct{ list plugins.EgressAllowList }

func (f fakeLists) AllowList(plugin string) (plugins.EgressAllowList, bool) {
	return f.list, plugin == f.list.Plugin
}

// fakeResolver maps hostnames to fixed addresses, with no network.
type fakeResolver map[string][]string

func (fakeResolver) LookupPort(_ context.Context, _, service string) (int, error) {
	port, err := strconv.Atoi(service)
	if err != nil {
		return 0, fmt.Errorf("fake port lookup: %w", err)
	}
	return port, nil
}

func (f fakeResolver) LookupIP(_ context.Context, _, host string) ([]net.IP, error) {
	addrs, ok := f[host]
	if !ok {
		return nil, fmt.Errorf("%w: %q", errUnknownHost, host)
	}
	ips := make([]net.IP, len(addrs))
	for i, a := range addrs {
		ips[i] = net.ParseIP(a)
	}
	return ips, nil
}

// dialLog records the addresses the proxy dialed.
type dialLog struct {
	mu    sync.Mutex
	addrs []string
}

func (d *dialLog) add(addr string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.addrs = append(d.addrs, addr)
}

func (d *dialLog) all() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.addrs...)
}

// proxyFixture is a running egress proxy in front of one local TLS server.
type proxyFixture struct {
	proxyAddr string
	port      string // the TLS server's port, used for every destination
	server    *httptest.Server
	dials     *dialLog
}

// reachedHeader marks a response from the TLS server behind the proxy.
const reachedHeader = "X-Reached"

func startProxy(t *testing.T) *proxyFixture {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(reachedHeader, "reached")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	_, port, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		t.Fatal(err)
	}
	list := plugins.EgressAllowList{Plugin: "sample", Hosts: []plugins.EgressHost{
		{Host: "example.com", Port: uint16(portNum)},
		{Host: "internal.example.com", Port: uint16(portNum)},
		{Host: "mixed.example.com", Port: uint16(portNum)},
	}}
	resolver := fakeResolver{
		"example.com":          {publicAddr},
		"other.example.com":    {publicAddr},
		"internal.example.com": {privateAddr},
		"mixed.example.com":    {publicAddr, privateAddr},
	}
	dials := &dialLog{}
	dial := func(ctx context.Context, network, address string) (net.Conn, error) {
		dials.add(address)
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	listener := listen(t)
	cfg, err := smokescreen.NewConfig(app.NewEgress(fakeReviewer{}, fakeLists{list: list}), smokescreen.Options{
		Listener: listener, Resolver: resolver, Dial: dial, Log: logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	quit, done := make(chan any), make(chan struct{})
	go func() {
		defer close(done)
		ss.StartWithConfig(cfg, quit)
	}()
	t.Cleanup(func() {
		close(quit)
		<-done
	})
	return &proxyFixture{proxyAddr: listener.Addr().String(), port: port, server: server, dials: dials}
}

// listen opens a local listener for the proxy.
func listen(t *testing.T) net.Listener {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return listener
}

// closeOrFail closes c, failing the test on any error but "already closed".
func closeOrFail(t *testing.T, c io.Closer) {
	t.Helper()
	if err := c.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		t.Error(err)
	}
}

// outcome is the proxy's status and, when a tunnel opened, what the TLS
// server behind it answered.
type outcome struct {
	status  int
	reached string
}

// connect sends one request for host through the proxy, a CONNECT or a
// plain proxied request, with the given Proxy-Authorization header (none
// when empty) and extra headers. It returns the status and, when a tunnel
// opened, what the TLS server behind it answered.
func (f *proxyFixture) connect(t *testing.T, method, host, auth string, extra http.Header) outcome {
	t.Helper()
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", f.proxyAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, conn)
	req := f.proxyRequest(t, conn, method, host, auth, extra)
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, resp.Body)
	if resp.StatusCode != http.StatusOK || method != http.MethodConnect {
		return outcome{status: resp.StatusCode}
	}
	return outcome{status: resp.StatusCode, reached: f.getThroughTunnel(t, conn, host)}
}

// proxyRequest writes the request to the proxy and returns it.
func (f *proxyFixture) proxyRequest(t *testing.T, conn net.Conn, method, host, auth string, extra http.Header) *http.Request {
	t.Helper()
	target := net.JoinHostPort(host, f.port)
	req, err := http.NewRequestWithContext(t.Context(), method, "http://"+target+"/", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	if auth != "" {
		req.Header.Set("Proxy-Authorization", auth)
	}
	maps.Copy(req.Header, extra)
	if method == http.MethodConnect {
		req.URL.Path, req.Host = "", target
		err = req.Write(conn)
	} else {
		err = req.WriteProxy(conn)
	}
	if err != nil {
		t.Fatal(err)
	}
	return req
}

// getThroughTunnel speaks TLS to the server at the end of the tunnel and
// returns its reachedHeader.
func (f *proxyFixture) getThroughTunnel(t *testing.T, conn net.Conn, host string) string {
	t.Helper()
	roots := x509.NewCertPool()
	roots.AddCert(f.server.Certificate())
	tlsConn := tls.Client(conn, &tls.Config{RootCAs: roots, ServerName: host, MinVersion: tls.VersionTLS12})
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+host+"/", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	if err = req.Write(tlsConn); err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(tlsConn), req)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, resp.Body)
	return resp.Header.Get(reachedHeader)
}

func TestProxyConnect(t *testing.T) {
	t.Parallel()
	f := startProxy(t)
	bearer := "Bearer " + sampleToken
	tests := []struct {
		name       string
		method     string
		host       string
		auth       string
		extra      http.Header
		wantStatus int
		wantBody   string
	}{
		{name: "listed host", method: http.MethodConnect, host: "example.com", auth: bearer, wantStatus: http.StatusOK, wantBody: "reached"},
		{name: "scheme in lower case", method: http.MethodConnect, host: "example.com", auth: "bearer " + sampleToken, wantStatus: http.StatusOK, wantBody: "reached"},
		{name: "unlisted host", method: http.MethodConnect, host: "other.example.com", auth: bearer, wantStatus: http.StatusProxyAuthRequired},
		{name: "listed host resolves to a private address", method: http.MethodConnect, host: "internal.example.com", auth: bearer, wantStatus: http.StatusBadGateway},
		{name: "listed host resolves to public and private addresses", method: http.MethodConnect, host: "mixed.example.com", auth: bearer, wantStatus: http.StatusBadGateway},
		{name: "no token", method: http.MethodConnect, host: "example.com", wantStatus: http.StatusProxyAuthRequired},
		{name: "token not accepted", method: http.MethodConnect, host: "example.com", auth: "Bearer synthetic-token-unknown", wantStatus: http.StatusProxyAuthRequired},
		{name: "basic auth instead of bearer", method: http.MethodConnect, host: "example.com", auth: "Basic c2FtcGxlOnRva2Vu", wantStatus: http.StatusProxyAuthRequired},
		{
			name: "upstream proxy requested", method: http.MethodConnect, host: "example.com", auth: bearer,
			extra:      http.Header{"X-Upstream-Https-Proxy": {"https://proxy.example.net:3128"}},
			wantStatus: http.StatusProxyAuthRequired,
		},
		{name: "plain HTTP proxying", method: http.MethodGet, host: "example.com", auth: bearer, wantStatus: http.StatusProxyAuthRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := f.connect(t, tt.method, tt.host, tt.auth, tt.extra)
			if want := (outcome{status: tt.wantStatus, reached: tt.wantBody}); got != want {
				t.Errorf("%s %s = %+v, want %+v", tt.method, tt.host, got, want)
			}
		})
	}
}

func TestProxyDialsOnlyCheckedAddresses(t *testing.T) {
	t.Parallel()
	f := startProxy(t)
	for _, host := range []string{"internal.example.com", "mixed.example.com", "other.example.com", "example.com"} {
		f.connect(t, http.MethodConnect, host, "Bearer "+sampleToken, nil)
	}
	want := net.JoinHostPort(publicAddr, f.port)
	got := f.dials.all()
	if len(got) != 1 || got[0] != want {
		t.Errorf("dialed %q, want only [%q]", got, want)
	}
}

// TestMain keeps the test hermetic: a developer's or CI's HTTPS_PROXY must
// not reach the proxy under test, which refuses to start with it set.
func TestMain(m *testing.M) {
	for _, name := range []string{"HTTPS_PROXY", "https_proxy"} {
		if err := os.Unsetenv(name); err != nil {
			panic(err)
		}
	}
	os.Exit(m.Run())
}

func TestNewConfigRefusesProxyEnvironment(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://proxy.example.net:3128")
	listener := listen(t)
	defer closeOrFail(t, listener)
	_, err := smokescreen.NewConfig(app.NewEgress(fakeReviewer{}, fakeLists{}), smokescreen.Options{Listener: listener})
	if !errors.Is(err, smokescreen.ErrProxyEnvironment) {
		t.Errorf("NewConfig() error = %v, want %v", err, smokescreen.ErrProxyEnvironment)
	}
}

func TestNewConfigNeedsListener(t *testing.T) {
	t.Parallel()
	_, err := smokescreen.NewConfig(app.NewEgress(fakeReviewer{}, fakeLists{}), smokescreen.Options{})
	if !errors.Is(err, smokescreen.ErrNoListener) {
		t.Errorf("NewConfig() error = %v, want %v", err, smokescreen.ErrNoListener)
	}
}
