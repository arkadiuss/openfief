package openfief_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

func TestBaseURL(t *testing.T) {
	t.Parallel()
	const def = "https://api.example.com"
	tests := []struct {
		name       string
		configured string
		want       string
		wantErr    bool
	}{
		{name: "nothing configured: the real host", want: def},
		{name: "a configured host", configured: "https://mirror.example.org", want: "https://mirror.example.org"},
		{name: "a port and a path", configured: "https://mirror.example.org:8443/flex/", want: "https://mirror.example.org:8443/flex"},
		{name: "upper-case scheme", configured: "HTTPS://mirror.example.org", want: "https://mirror.example.org"},
		{name: "plain http", configured: "http://mirror.example.org", wantErr: true},
		{name: "another scheme", configured: "ftp://mirror.example.org", wantErr: true},
		{name: "no host", configured: "https:///flex", wantErr: true},
		{name: "a user", configured: "https://owner@mirror.example.org", wantErr: true},
		{name: "a query", configured: "https://mirror.example.org/?key=x", wantErr: true},
		{name: "an empty query", configured: "https://mirror.example.org/?", wantErr: true},
		{name: "a fragment", configured: "https://mirror.example.org/#x", wantErr: true},
		{name: "port zero", configured: "https://mirror.example.org:0", wantErr: true},
		{name: "port too large", configured: "https://mirror.example.org:65536", wantErr: true},
		{name: "empty port", configured: "https://mirror.example.org:", wantErr: true},
		{name: "not a URL", configured: "mirror.example.org", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := openfief.BaseURL(tt.configured, def)
			if tt.wantErr {
				if !errors.Is(err, openfief.ErrBaseURL) {
					t.Fatalf("BaseURL(%q) = %q, %v; want ErrBaseURL", tt.configured, got, err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("BaseURL(%q) = %q, %v; want %q", tt.configured, got, err, tt.want)
			}
		})
	}
}

func TestNewEndpointBundle(t *testing.T) {
	t.Parallel()
	ca := newTestCA(t, "Synthetic Extra CA")
	tests := []struct {
		name    string
		bundle  []byte
		want    []byte
		wantErr bool
	}{
		{name: "no bundle"},
		{name: "an empty bundle", bundle: []byte("\n")},
		{name: "a certificate", bundle: ca.pem, want: ca.pem},
		{name: "no certificate in it", bundle: []byte("not a certificate"), wantErr: true},
		{name: "only a key", bundle: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{1}}), wantErr: true},
		{name: "a broken certificate", bundle: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte{1}}), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := openfief.NewEndpoint("", "https://api.example.com", tt.bundle)
			if tt.wantErr {
				if !errors.Is(err, openfief.ErrCABundle) {
					t.Fatalf("NewEndpoint = %v, want ErrCABundle", err)
				}
				return
			}
			want := openfief.Endpoint{BaseURL: "https://api.example.com", CABundle: tt.want}
			if err != nil {
				t.Fatalf("NewEndpoint: %v", err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("NewEndpoint (-want +got):\n%s", diff)
			}
		})
	}
	if _, err := openfief.NewEndpoint("http://api.example.com", "", ca.pem); !errors.Is(err, openfief.ErrBaseURL) {
		t.Errorf("NewEndpoint with plain http = %v, want ErrBaseURL", err)
	}
}

func TestReadCABundle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	present := filepath.Join(dir, openfief.CABundleFile)
	if err := os.WriteFile(present, []byte("synthetic bundle"), 0o600); err != nil {
		t.Fatal(err)
	}
	huge := filepath.Join(dir, "huge.crt")
	if err := os.WriteFile(huge, make([]byte, 1<<20+1), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "the Secret holds a bundle", path: present, want: "synthetic bundle"},
		{name: "no Secret: none", path: filepath.Join(dir, "missing", openfief.CABundleFile)},
		{name: "over a Secret's size", path: huge, wantErr: true},
		{name: "a directory", path: dir, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := openfief.ReadCABundle(tt.path)
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, openfief.ErrCABundle)) {
				t.Fatalf("ReadCABundle error = %v, want error %v", err, tt.wantErr)
			}
			if string(got) != tt.want {
				t.Errorf("ReadCABundle = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestEndpointTrustsExtraCA runs real TLS through a fake egress proxy: an
// endpoint trusts a server whose CA only its bundle holds, still trusts one
// the system's roots hold, and trusts neither without the right roots.
// "System" is a CA made for the test, since the real system roots can't sign
// a test server.
func TestEndpointTrustsExtraCA(t *testing.T) {
	t.Parallel()
	system, extra := newTestCA(t, "Synthetic System CA"), newTestCA(t, "Synthetic Extra CA")
	systemSrv, extraSrv := tlsServer(t, system), tlsServer(t, extra)
	tests := []struct {
		name   string
		server *httptest.Server
		bundle []byte
		wantOK bool
	}{
		{name: "the bundle's CA is trusted", server: extraSrv, bundle: extra.pem, wantOK: true},
		{name: "the system's roots still apply with a bundle", server: systemSrv, bundle: extra.pem, wantOK: true},
		{name: "the system's roots apply without a bundle", server: systemSrv, wantOK: true},
		{name: "a CA in neither is refused", server: extraSrv},
		{name: "a bundle doesn't trust another CA", server: extraSrv, bundle: newTestCA(t, "Synthetic Other CA").pem},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, err := openfief.NewEndpoint(tt.server.URL, "", tt.bundle)
			if err != nil {
				t.Fatalf("NewEndpoint: %v", err)
			}
			err = getVia(t, endpointClient(t, e, system.pool()), e.BaseURL+"/api/ping")
			if (err == nil) != tt.wantOK {
				t.Fatalf("GET = %v, want success %v", err, tt.wantOK)
			}
			var unknown x509.UnknownAuthorityError
			if !tt.wantOK && !errors.As(err, &unknown) {
				t.Errorf("GET error = %v, want an unknown authority", err)
			}
		})
	}
}

// TestEndpointRefusesRedirectElsewhere checks that the endpoint's client
// follows no redirect off its base URL, so a credential in a query never
// leaves the configured host.
func TestEndpointRefusesRedirectElsewhere(t *testing.T) {
	t.Parallel()
	ca := newTestCA(t, "Synthetic System CA")
	srv := tlsServerWith(t, ca, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://elsewhere.example.com/collect", http.StatusFound)
	}))
	e, err := openfief.NewEndpoint(srv.URL+"/api", "", nil)
	if err != nil {
		t.Fatalf("NewEndpoint: %v", err)
	}
	err = getVia(t, endpointClient(t, e, ca.pool()), e.BaseURL+"/eod?api_token=synthetic")
	if !errors.Is(err, openfief.ErrOutsideEndpoint) {
		t.Errorf("GET = %v, want ErrOutsideEndpoint", err)
	}
}

func TestEndpointWithin(t *testing.T) {
	t.Parallel()
	const base = "https://data.example.com/flex/v3"
	tests := []struct {
		name   string
		url    string
		wantOK bool
	}{
		{name: "the base itself", url: base, wantOK: true},
		{name: "below the base", url: base + "/SendRequest?t=x", wantOK: true},
		{name: "port 443 written", url: "https://data.example.com:443/flex/v3/GetStatement", wantOK: true},
		{name: "host in another case", url: "https://DATA.example.com/flex/v3/x", wantOK: true},
		{name: "beside the base path", url: "https://data.example.com/flex/v30", wantOK: false},
		{name: "above the base path", url: "https://data.example.com/flex", wantOK: false},
		{name: "another host", url: "https://other.example.com/flex/v3/x", wantOK: false},
		{name: "another port", url: "https://data.example.com:8443/flex/v3/x", wantOK: false},
		{name: "plain http", url: "http://data.example.com/flex/v3/x", wantOK: false},
	}
	e := openfief.Endpoint{BaseURL: base}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			next := &countingTransport{}
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, tt.url, http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := e.Within(next).RoundTrip(req)
			if err == nil {
				err = resp.Body.Close()
			}
			if (err == nil) != tt.wantOK || next.calls != map[bool]int{true: 1, false: 0}[tt.wantOK] {
				t.Errorf("RoundTrip(%s) = %v after %d calls, want success %v", tt.url, err, next.calls, tt.wantOK)
			}
			if !tt.wantOK && !errors.Is(err, openfief.ErrOutsideEndpoint) {
				t.Errorf("RoundTrip error = %v, want ErrOutsideEndpoint", err)
			}
		})
	}
}

func TestEndpointRoots(t *testing.T) {
	t.Parallel()
	system, extra := newTestCA(t, "Synthetic System CA"), newTestCA(t, "Synthetic Extra CA")
	sys := system.pool()
	before := sys.Clone()
	roots, err := openfief.Endpoint{BaseURL: "https://api.example.com", CABundle: extra.pem}.Roots(sys)
	if err != nil {
		t.Fatalf("Roots: %v", err)
	}
	for _, ca := range []*testCA{system, extra} {
		if _, verifyErr := ca.cert.Verify(x509.VerifyOptions{Roots: roots}); verifyErr != nil {
			t.Errorf("%s is not among the roots: %v", ca.cert.Subject.CommonName, verifyErr)
		}
	}
	if !sys.Equal(before) {
		t.Error("Roots changed the system pool")
	}
	none, err := openfief.Endpoint{CABundle: extra.pem}.Roots(nil)
	if err != nil {
		t.Fatalf("Roots(nil): %v", err)
	}
	if _, err := extra.cert.Verify(x509.VerifyOptions{Roots: none}); err != nil {
		t.Errorf("the bundle's CA is not among the roots without a system pool: %v", err)
	}
}

// countingTransport answers every request with 200 and counts them.
type countingTransport struct{ calls int }

func (c *countingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	c.calls++
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

// getVia sends a GET for u with client and returns its error, closing the body.
func getVia(t *testing.T, client *http.Client, u string) error {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, u, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GET: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	return nil
}

// endpointClient is an HTTP client for e through a fake egress proxy, whose
// system roots are system.
func endpointClient(t *testing.T, e openfief.Endpoint, system *x509.CertPool) *http.Client {
	t.Helper()
	proxySrv := httptest.NewServer(&fakeProxy{})
	t.Cleanup(proxySrv.Close)
	proxyURL, err := url.Parse(proxySrv.URL)
	if err != nil {
		t.Fatal(err)
	}
	tokenPath := filepath.Join(t.TempDir(), "token")
	if writeErr := os.WriteFile(tokenPath, []byte("synthetic-egress-token"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	cfg, err := e.TLSConfig(system)
	if err != nil {
		t.Fatalf("TLSConfig: %v", err)
	}
	et := &openfief.EgressTransport{Proxy: proxyURL, Tokens: openfief.TokenFile{Path: tokenPath}, TLSClientConfig: cfg}
	t.Cleanup(et.CloseIdleConnections)
	return &http.Client{Transport: e.Within(et), Timeout: 10 * time.Second}
}

// testCA is a certificate authority made for one test.
type testCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pem  []byte
}

// pool is a pool of the CA alone.
func (c *testCA) pool() *x509.CertPool {
	p := x509.NewCertPool()
	p.AddCert(c.cert)
	return p
}

// newTestCA makes a self-signed CA named name.
func newTestCA(t *testing.T, name string) *testCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: name},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return &testCA{cert: cert, key: key, pem: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})}
}

// tlsServer is an HTTPS server on 127.0.0.1 with a certificate ca signs,
// answering 200.
func tlsServer(t *testing.T, ca *testCA) *httptest.Server {
	t.Helper()
	return tlsServerWith(t, ca, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

// tlsServerWith is an HTTPS server on 127.0.0.1 with a certificate ca signs,
// serving h.
func tlsServerWith(t *testing.T, ca *testCA, h http.Handler) *httptest.Server {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "127.0.0.1"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)}, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewUnstartedServer(h)
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}, MinVersion: tls.VersionTLS12}
	srv.StartTLS()
	t.Cleanup(srv.Close)
	if !strings.HasPrefix(srv.URL, "https://127.0.0.1:") {
		t.Fatalf("server URL %s is not on 127.0.0.1", srv.URL)
	}
	return srv
}
