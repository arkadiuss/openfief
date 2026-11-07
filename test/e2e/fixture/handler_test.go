package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
)

// The synthetic Flex and EODHD keys the fixture is started with; not real.
const (
	flexAccepted  = "synthetic-flex-token"
	eodhdAccepted = "synthetic-eodhd-token"
)

// fixtureServer serves the fixture's handler over TLS on 127.0.0.1.
func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	h := newHandler(credentials{flexToken: flexAccepted, eodhdToken: eodhdAccepted}, slog.New(slog.DiscardHandler))
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// TestFixtureServesFlex checks that the connectors' own Flex Web Service
// client gets the synthetic statement from the fixture with the right
// token, and IBKR's invalid-token error with another.
func TestFixtureServesFlex(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	tests := []struct {
		name    string
		token   string
		want    []byte
		wantErr string
	}{
		{name: "the accepted token gets the statement", token: flexAccepted, want: flexStatement},
		{name: "another token is invalid", token: "synthetic-other-token", wantErr: "1015"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := ibkrflex.Client{
				HTTP: srv.Client(), Token: tt.token, QueryID: "000000",
				SendRequestURL:  srv.URL + FlexPath + "/SendRequest",
				GetStatementURL: srv.URL + FlexPath + "/GetStatement",
				Wait:            func(context.Context, int) error { return nil },
			}
			got, err := c.FetchStatement(t.Context())
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("FetchStatement = %v, want error code %s", err, tt.wantErr)
				}
				return
			}
			if err != nil || !bytes.Equal(got, tt.want) {
				t.Errorf("FetchStatement = %d bytes, %v; want the synthetic statement", len(got), err)
			}
		})
	}
}

// TestFixtureServesEODHD checks that the connectors' own EODHD client finds
// a synthetic instrument by ISIN, reads its weekday closes within the
// fixture's window, and is refused with another API token.
func TestFixtureServesEODHD(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	c := eodhd.New(srv.Client(), srv.URL+EODHDPath, eodhdAccepted)
	isin, err := ledger.ParseISIN("IETEST000026")
	if err != nil {
		t.Fatal(err)
	}
	found, err := c.Listings(t.Context(), isin)
	if err != nil || len(found) != 1 || found[0].Code != "TSTB" || found[0].Exchange != "XETRA" {
		t.Errorf("Listings = %+v, %v; want TSTB on XETRA", found, err)
	}
	checkCloses(t, c)
	if _, err := c.Closes(t.Context(), "NONE.XETRA", day(t, "2026-03-06"), day(t, "2026-03-10")); !errors.Is(err, eodhd.ErrNotFound) {
		t.Errorf("Closes of an unknown ticker = %v, want ErrNotFound", err)
	}
	other := eodhd.New(srv.Client(), srv.URL+EODHDPath, "synthetic-other-token")
	if _, err := other.Listings(t.Context(), isin); !errors.Is(err, eodhd.ErrUnexpectedStatus) {
		t.Errorf("Listings with another token = %v, want an unexpected status", err)
	}
}

// checkCloses checks the weekday closes of the synthetic TSTB.XETRA from a
// Friday to the Tuesday after, with c.
func checkCloses(t *testing.T, c *eodhd.Client) {
	t.Helper()
	closes, err := c.Closes(t.Context(), "TSTB.XETRA", day(t, "2026-03-06"), day(t, "2026-03-10"))
	if err != nil {
		t.Fatalf("Closes: %v", err)
	}
	got := make([]string, 0, len(closes))
	for _, cl := range closes {
		got = append(got, cl.Date.String()+" "+cl.Close.String())
	}
	want := []string{"2026-03-06 107", "2026-03-09 107", "2026-03-10 107"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("closes (-want +got):\n%s", diff)
	}
}

// TestFixtureServesNBP checks that the connectors' own NBP client reads one
// table A per weekday of the range, and nothing for a weekend.
func TestFixtureServesNBP(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	c := nbp.New(srv.Client(), srv.URL+NBPPath)
	tests := []struct {
		name     string
		from, to string
		want     []string
	}{
		{
			name: "Friday to Monday", from: "2026-03-06", to: "2026-03-09",
			want: []string{
				"2026-03-06 USD 3.9512", "2026-03-06 EUR 4.2731",
				"2026-03-09 USD 3.9512", "2026-03-09 EUR 4.2731",
			},
		},
		{name: "a weekend", from: "2026-03-07", to: "2026-03-08"},
		{name: "after the window", from: "2026-04-01", to: "2026-04-03"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rates, err := c.Rates(t.Context(), day(t, tt.from), day(t, tt.to))
			if err != nil {
				t.Fatalf("Rates: %v", err)
			}
			var got []string
			for _, r := range rates {
				got = append(got, r.Date.String()+" "+r.Currency.String()+" "+r.Rate.String())
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("rates (-want +got):\n%s", diff)
			}
		})
	}
}

// TestFixtureRefusesMalformed checks the answers the connectors never
// trigger: a Flex query without its version, an NBP range over 93 days or
// without JSON, an unknown path.
func TestFixtureRefusesMalformed(t *testing.T) {
	t.Parallel()
	srv := fixtureServer(t)
	tests := []struct {
		name     string
		path     string
		wantCode int
		wantBody string
	}{
		{name: "Flex without a version", path: FlexPath + "/SendRequest?t=" + flexAccepted + "&q=1", wantCode: http.StatusOK, wantBody: "1014"},
		{name: "Flex statement of another reference", path: FlexPath + "/GetStatement?t=" + flexAccepted + "&q=1&v=3", wantCode: http.StatusOK, wantBody: "1014"},
		{name: "NBP over 93 days", path: NBPPath + "/api/exchangerates/tables/A/2026-01-01/2026-04-30/?format=json", wantCode: http.StatusBadRequest},
		{name: "NBP without JSON", path: NBPPath + "/api/exchangerates/tables/A/2026-01-05/2026-01-06/", wantCode: http.StatusBadRequest},
		{name: "EODHD with a bad date", path: EODHDPath + "/api/eod/TSTB.XETRA?api_token=" + eodhdAccepted + "&fmt=json&from=x", wantCode: http.StatusUnprocessableEntity},
		{name: "an unknown path", path: "/elsewhere", wantCode: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fetch(t, srv.Client(), srv.URL+tt.path)
			if got.code != tt.wantCode || !strings.Contains(got.body, tt.wantBody) {
				t.Errorf("GET %s = %d %q, want %d with %q", tt.path, got.code, got.body, tt.wantCode, tt.wantBody)
			}
		})
	}
}

// TestRun checks that run serves the handler over TLS with the certificate
// and key files, refuses a missing key pair, and stops when its context
// ends.
func TestRun(t *testing.T) {
	t.Parallel()
	kp := keyPairFiles(t)
	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	go func() { done <- run(ctx, ln, kp.certFile, kp.keyFile, h) }()
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: kp.roots, MinVersion: tls.VersionTLS12}}}
	if got := fetch(t, client, "https://"+ln.Addr().String()+"/"); got.code != http.StatusTeapot {
		t.Errorf("GET = %d, want %d", got.code, http.StatusTeapot)
	}
	cancel()
	if runErr := <-done; runErr != nil {
		t.Errorf("run = %v, want nil after its context ended", runErr)
	}
	other, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if runErr := run(t.Context(), other, filepath.Join(t.TempDir(), "none.crt"), kp.keyFile, h); runErr == nil {
		t.Error("run without a certificate file = nil, want an error")
	}
}

// answer is an HTTP status and body.
type answer struct {
	code int
	body string
}

// fetch GETs u with client and returns the status and body.
func fetch(t *testing.T, client *http.Client, u string) answer {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, u, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", u, err)
	}
	body, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
	return answer{code: resp.StatusCode, body: string(body)}
}

// day parses an ISO date.
func day(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// keyPair is a certificate's and its key's files and a pool trusting it.
type keyPair struct {
	certFile, keyFile string
	roots             *x509.CertPool
}

// keyPairFiles writes a self-signed certificate for 127.0.0.1 and its key,
// made for the test, and returns their files and a pool trusting it.
func keyPairFiles(t *testing.T) keyPair {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Synthetic fixture"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)}, IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certFile, keyFile := filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key")
	if writeErr := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	if writeErr := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	return keyPair{certFile: certFile, keyFile: keyFile, roots: roots}
}
