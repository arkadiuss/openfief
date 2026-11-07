package openfief_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// serviceName is the in-cluster name the service's certificate is issued
// for; plugins have no DNS, so it never resolves.
const serviceName = "fixture.e2e-test.svc"

// TestEndpointClusterClient checks the in-cluster path (D-74): every request
// goes straight to the service's address, with no proxy and the base URL's
// host in the Host header, its certificate is
// verified for the base URL's host against the system's roots plus the
// bundle, and requests outside the base URL are refused.
func TestEndpointClusterClient(t *testing.T) {
	t.Parallel()
	system, extra := newTestCA(t, "Synthetic System CA"), newTestCA(t, "Synthetic Extra CA")
	srv := serviceServer(t, extra, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != serviceName {
			t.Errorf("Host = %q, want %q", r.Host, serviceName)
		}
		w.WriteHeader(http.StatusOK)
	}))
	addr := strings.TrimPrefix(srv.URL, "https://")
	tests := []struct {
		name    string
		base    string
		bundle  []byte
		outside bool
		wantErr func(error) bool
	}{
		{
			name: "the service, verified for the base URL's host", base: "https://" + serviceName + "/nbp", bundle: extra.pem,
			wantErr: func(err error) bool { return err == nil },
		},
		{
			name: "without the bundle, the service's CA is unknown", base: "https://" + serviceName + "/nbp",
			wantErr: func(err error) bool { var u x509.UnknownAuthorityError; return errors.As(err, &u) },
		},
		{
			name: "a host the certificate doesn't name", base: "https://other.e2e-test.svc/nbp", bundle: extra.pem,
			wantErr: func(err error) bool { var h x509.HostnameError; return errors.As(err, &h) },
		},
		{
			name: "outside the base URL", base: "https://" + serviceName + "/nbp", bundle: extra.pem, outside: true,
			wantErr: func(err error) bool { return errors.Is(err, openfief.ErrOutsideEndpoint) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := clusterGet(t, tt.base, tt.bundle, addr, system.pool(), tt.outside)
			if !tt.wantErr(err) {
				t.Errorf("GET = %v, not what the case expects", err)
			}
		})
	}
	if _, err := (openfief.Endpoint{BaseURL: "https://" + serviceName, ClusterService: "no-port"}).ClusterClient(nil); !errors.Is(err, openfief.ErrClusterService) {
		t.Errorf("ClusterClient without a port = %v, want ErrClusterService", err)
	}
}

// clusterGet GETs a path below base, or one outside it, with the cluster
// client of an endpoint at base trusting bundle, whose service is at addr.
func clusterGet(t *testing.T, base string, bundle []byte, addr string, system *x509.CertPool, outside bool) error {
	t.Helper()
	e, err := openfief.NewEndpoint(base, "", bundle)
	if err != nil {
		t.Fatalf("NewEndpoint: %v", err)
	}
	e.ClusterService = addr
	client, err := e.ClusterClient(system)
	if err != nil {
		t.Fatalf("ClusterClient: %v", err)
	}
	u := e.BaseURL + "/api"
	if outside {
		u = "https://" + serviceName + "/eodhd"
	}
	return getVia(t, client, u)
}

// serviceServer is an HTTPS server on 127.0.0.1 standing in for an
// in-cluster service, with a certificate ca signs for serviceName only.
func serviceServer(t *testing.T, ca *testCA, h http.Handler) *httptest.Server {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(3), Subject: pkix.Name{CommonName: serviceName},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		DNSNames: []string{serviceName}, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
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
	return srv
}
