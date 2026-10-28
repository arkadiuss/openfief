package smokescreen_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

// Synthetic tokens and addresses. The public-looking address is never
// dialed: the test dialer sends every connection to the local TLS server.
const (
	sampleToken = "synthetic-token-sample"
	nolistToken = "synthetic-token-nolist"
	publicAddr  = "8.8.4.4"
	privateAddr = "10.0.0.7"
)

var errUnknownHost = errors.New("no such host in the fake resolver")

// fakeReviewer accepts sampleToken as plugin "sample" and nolistToken as
// plugin "nolist", which has no egress allow-list.
type fakeReviewer struct{}

func (fakeReviewer) ReviewToken(_ context.Context, token string) (string, error) {
	switch token {
	case sampleToken:
		return "system:serviceaccount:openfief:" + plugins.ServiceAccountName("sample"), nil
	case nolistToken:
		return "system:serviceaccount:openfief:" + plugins.ServiceAccountName("nolist"), nil
	}
	return "", app.ErrUnauthenticated
}

// swappableLists serves one allow-list that a test can replace while the
// proxy runs, as the operator's ConfigMap watch does.
type swappableLists struct {
	list atomic.Pointer[plugins.EgressAllowList]
}

func newLists(l plugins.EgressAllowList) *swappableLists {
	s := &swappableLists{}
	s.list.Store(&l)
	return s
}

func (s *swappableLists) AllowList(plugin string) (plugins.EgressAllowList, bool) {
	l := s.list.Load()
	return *l, plugin == l.Plugin
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
	lists     *swappableLists
	// stop ends Serve and returns its error; the test's cleanup calls it too.
	stop func() error
}

// reachedHeader marks a response from the TLS server behind the proxy.
const reachedHeader = "X-Reached"

// sampleList is plugin "sample"'s allow-list for destinations on port.
func sampleList(port uint16) plugins.EgressAllowList {
	return plugins.EgressAllowList{Plugin: "sample", Hosts: []plugins.EgressHost{
		{Host: "example.com", Port: port},
		{Host: "internal.example.com", Port: port},
		{Host: "mixed.example.com", Port: port},
	}}
}

// startProxy runs a proxy whose plugin "sample" may reach sampleList. With
// noSource, the egress use case has no AllowListSource at all.
func startProxy(t *testing.T, noSource bool) *proxyFixture {
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
	f := &proxyFixture{port: port, server: server, dials: &dialLog{}, lists: newLists(sampleList(uint16(portNum)))}
	egressCase := app.NewEgress(fakeReviewer{}, f.lists)
	if noSource {
		egressCase = app.NewEgress(fakeReviewer{}, nil)
	}
	r := serve(t, egressCase, smokescreen.Options{
		Resolver: fakeResolver{
			"example.com":          {publicAddr},
			"other.example.com":    {publicAddr},
			"internal.example.com": {privateAddr},
			"mixed.example.com":    {publicAddr, privateAddr},
		},
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			f.dials.add(address)
			return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
		},
		Healthcheck:  http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }),
		DrainTimeout: 100 * time.Millisecond,
	})
	f.proxyAddr, f.stop = r.addr, r.stop
	return f
}

// running is a served proxy: its address, and a stop function that ends
// Serve and returns its error; stopping twice is fine.
type running struct {
	addr string
	stop func() error
}

// serve builds a proxy on a fresh local listener and serves it until the
// test ends.
func serve(t *testing.T, e *app.Egress, opts smokescreen.Options) running {
	t.Helper()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	opts.Log = logger
	opts.Listener = listen(t)
	proxy, err := smokescreen.New(e, opts)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- proxy.Serve(ctx) }()
	var once sync.Once
	var serveErr error
	stop := func() error {
		once.Do(func() {
			cancel()
			serveErr = <-served
		})
		return serveErr
	}
	t.Cleanup(func() {
		if err := stop(); err != nil {
			t.Error(err)
		}
	})
	return running{addr: opts.Listener.Addr().String(), stop: stop}
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

// basic is the Proxy-Authorization header an HTTP client sends for a proxy
// URL with user and password (D-63).
func basic(user, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
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
	o := f.open(t, method, host, auth, extra)
	conn, status := o.conn, o.status
	defer closeOrFail(t, conn)
	if status != http.StatusOK || method != http.MethodConnect {
		return outcome{status: status}
	}
	return outcome{status: status, reached: f.getThroughTunnel(t, conn, host)}
}

// opened is a connection to the proxy, which the caller closes, and the
// proxy's status for the request sent on it.
type opened struct {
	conn   net.Conn
	status int
}

// open sends one request through the proxy.
func (f *proxyFixture) open(t *testing.T, method, host, auth string, extra http.Header) opened {
	t.Helper()
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", f.proxyAddr)
	if err != nil {
		t.Fatal(err)
	}
	req := f.proxyRequest(t, conn, method, host, auth, extra)
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, resp.Body)
	return opened{conn: conn, status: resp.StatusCode}
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

// roots trusts the TLS server behind the proxy.
func (f *proxyFixture) roots() *x509.CertPool {
	roots := x509.NewCertPool()
	roots.AddCert(f.server.Certificate())
	return roots
}

// getThroughTunnel speaks TLS to the server at the end of the tunnel and
// returns its reachedHeader.
func (f *proxyFixture) getThroughTunnel(t *testing.T, conn net.Conn, host string) string {
	t.Helper()
	tlsConn := tls.Client(conn, &tls.Config{RootCAs: f.roots(), ServerName: host, MinVersion: tls.VersionTLS12})
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
	f := startProxy(t, false)
	auth := basic("plugin", sampleToken)
	tests := []struct {
		name       string
		method     string
		host       string
		auth       string
		extra      http.Header
		wantStatus int
		wantBody   string
	}{
		{name: "listed host", method: http.MethodConnect, host: "example.com", auth: auth, wantStatus: http.StatusOK, wantBody: "reached"},
		{name: "scheme in lower case", method: http.MethodConnect, host: "example.com", auth: "basic " + base64.StdEncoding.EncodeToString([]byte("plugin:"+sampleToken)), wantStatus: http.StatusOK, wantBody: "reached"},
		{name: "no user name", method: http.MethodConnect, host: "example.com", auth: basic("", sampleToken), wantStatus: http.StatusOK, wantBody: "reached"},
		{name: "unlisted host", method: http.MethodConnect, host: "other.example.com", auth: auth, wantStatus: http.StatusProxyAuthRequired},
		{name: "listed host resolves to a private address", method: http.MethodConnect, host: "internal.example.com", auth: auth, wantStatus: http.StatusBadGateway},
		{name: "listed host resolves to public and private addresses", method: http.MethodConnect, host: "mixed.example.com", auth: auth, wantStatus: http.StatusBadGateway},
		{name: "no token", method: http.MethodConnect, host: "example.com", wantStatus: http.StatusProxyAuthRequired},
		{name: "token not accepted", method: http.MethodConnect, host: "example.com", auth: basic("plugin", "synthetic-token-unknown"), wantStatus: http.StatusProxyAuthRequired},
		{name: "token as the user name", method: http.MethodConnect, host: "example.com", auth: basic(sampleToken, ""), wantStatus: http.StatusProxyAuthRequired},
		{name: "bearer instead of basic", method: http.MethodConnect, host: "example.com", auth: "Bearer " + sampleToken, wantStatus: http.StatusProxyAuthRequired},
		{name: "basic credentials not base64", method: http.MethodConnect, host: "example.com", auth: "Basic %%%", wantStatus: http.StatusProxyAuthRequired},
		{name: "basic credentials without a colon", method: http.MethodConnect, host: "example.com", auth: "Basic " + base64.StdEncoding.EncodeToString([]byte(sampleToken)), wantStatus: http.StatusProxyAuthRequired},
		{name: "plugin without an allow-list", method: http.MethodConnect, host: "example.com", auth: basic("plugin", nolistToken), wantStatus: http.StatusProxyAuthRequired},
		{
			name: "upstream proxy requested", method: http.MethodConnect, host: "example.com", auth: auth,
			extra:      http.Header{"X-Upstream-Https-Proxy": {"https://proxy.example.net:3128"}},
			wantStatus: http.StatusProxyAuthRequired,
		},
		{name: "plain HTTP proxying", method: http.MethodGet, host: "example.com", auth: auth, wantStatus: http.StatusProxyAuthRequired},
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

// TestProxyGoClientProxyURL checks D-63's premise end to end: Go's HTTP
// client, given the token as the proxy URL's password, sends it in a Basic
// Proxy-Authorization header on CONNECT, and the proxy accepts it.
func TestProxyGoClientProxyURL(t *testing.T) {
	t.Parallel()
	f := startProxy(t, false)
	proxyURL := &url.URL{Scheme: "http", User: url.UserPassword("plugin", sampleToken), Host: f.proxyAddr}
	client := &http.Client{Transport: &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{RootCAs: f.roots(), MinVersion: tls.VersionTLS12},
	}}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+net.JoinHostPort("example.com", f.port)+"/", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, resp.Body)
	if got := resp.Header.Get(reachedHeader); got != "reached" {
		t.Errorf("through the proxy URL, %s = %q, want %q", reachedHeader, got, "reached")
	}
}

func TestProxyDialsOnlyCheckedAddresses(t *testing.T) {
	t.Parallel()
	f := startProxy(t, false)
	for _, host := range []string{"internal.example.com", "mixed.example.com", "other.example.com", "example.com"} {
		f.connect(t, http.MethodConnect, host, basic("plugin", sampleToken), nil)
	}
	want := net.JoinHostPort(publicAddr, f.port)
	got := f.dials.all()
	if len(got) != 1 || got[0] != want {
		t.Errorf("dialed %q, want only [%q]", got, want)
	}
}

// TestProxyWithoutAllowListSource checks that the proxy never falls back to
// Smokescreen's allow-all: with no allow-lists at all, every destination is
// refused (D-63).
func TestProxyWithoutAllowListSource(t *testing.T) {
	t.Parallel()
	f := startProxy(t, true)
	got := f.connect(t, http.MethodConnect, "example.com", basic("plugin", sampleToken), nil)
	if want := (outcome{status: http.StatusProxyAuthRequired}); got != want {
		t.Errorf("CONNECT example.com = %+v, want %+v", got, want)
	}
	if dials := f.dials.all(); len(dials) != 0 {
		t.Errorf("dialed %q, want nothing", dials)
	}
}

// TestProxyFollowsSwappedAllowList checks that a replaced allow-list applies
// to the next CONNECT while the proxy runs.
func TestProxyFollowsSwappedAllowList(t *testing.T) {
	t.Parallel()
	f := startProxy(t, false)
	auth := basic("plugin", sampleToken)
	original := *f.lists.list.Load()
	steps := []struct {
		name string
		list plugins.EgressAllowList
		want outcome
	}{
		{name: "original list", list: original, want: outcome{status: http.StatusOK, reached: "reached"}},
		{name: "list without the host", list: plugins.EgressAllowList{Plugin: "sample"}, want: outcome{status: http.StatusProxyAuthRequired}},
		{name: "list restored", list: original, want: outcome{status: http.StatusOK, reached: "reached"}},
	}
	for _, step := range steps {
		f.lists.list.Store(&step.list)
		if got := f.connect(t, http.MethodConnect, "example.com", auth, nil); got != step.want {
			t.Errorf("%s: CONNECT example.com = %+v, want %+v", step.name, got, step.want)
		}
	}
}

func TestProxyHealthcheck(t *testing.T) {
	t.Parallel()
	f := startProxy(t, false)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+f.proxyAddr+"/healthcheck", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, resp.Body)
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("GET /healthcheck = %d, want %d", resp.StatusCode, http.StatusTeapot)
	}
}

// TestProxyServeStops checks that Serve returns nil when its context ends,
// closes tunnels still open after the drain timeout, and stops accepting.
func TestProxyServeStops(t *testing.T) {
	t.Parallel()
	f := startProxy(t, false)
	o := f.open(t, http.MethodConnect, "example.com", basic("plugin", sampleToken), nil)
	conn, status := o.conn, o.status
	defer closeOrFail(t, conn)
	if status != http.StatusOK {
		t.Fatalf("CONNECT example.com = %d, want %d", status, http.StatusOK)
	}
	if err := f.stop(); err != nil {
		t.Fatalf("Serve() = %v, want nil", err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	_, err := conn.Read(make([]byte, 1))
	if ne, ok := errors.AsType[net.Error](err); err == nil || (ok && ne.Timeout()) {
		t.Errorf("read from a tunnel after shutdown = %v, want it closed", err)
	}
	if c, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", f.proxyAddr); err == nil {
		closeOrFail(t, c)
		t.Error("the proxy still accepts connections after shutdown")
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

func TestNewRefusesProxyEnvironment(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://proxy.example.net:3128")
	listener := listen(t)
	defer closeOrFail(t, listener)
	_, err := smokescreen.New(app.NewEgress(fakeReviewer{}, nil), smokescreen.Options{Listener: listener})
	if !errors.Is(err, smokescreen.ErrProxyEnvironment) {
		t.Errorf("New() error = %v, want %v", err, smokescreen.ErrProxyEnvironment)
	}
}

func TestNewNeedsListenerAndEgress(t *testing.T) {
	t.Parallel()
	listener := listen(t)
	t.Cleanup(func() { closeOrFail(t, listener) })
	tests := []struct {
		name    string
		egress  *app.Egress
		opts    smokescreen.Options
		wantErr error
	}{
		{name: "no listener", egress: app.NewEgress(fakeReviewer{}, nil), wantErr: smokescreen.ErrNoListener},
		{name: "no egress use case", opts: smokescreen.Options{Listener: listener}, wantErr: smokescreen.ErrNoEgress},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := smokescreen.New(tt.egress, tt.opts); !errors.Is(err, tt.wantErr) {
				t.Errorf("New() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
