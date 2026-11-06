package openfief

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// The instance's optional CA bundle Secret. A connector's manifest asks for
// it with role CABundleRole at CABundleDir, marked optional; the owner
// creates the Secret <instance>--ca-bundle (D-89) with the key ca.crt only
// when an endpoint needs it, such as a self-hosted service or a corporate
// TLS proxy with a private CA.
const (
	// CABundleRole is the role of the CA bundle Secret in a manifest.
	CABundleRole = "ca-bundle"
	// CABundleDir is where a manifest mounts the CA bundle Secret.
	CABundleDir = "/etc/openfief/ca-bundle"
	// CABundleFile is the Secret's key, and so the file in CABundleDir,
	// that holds the PEM certificates.
	CABundleFile = "ca.crt"
	// maxCABundle bounds the bundle read from the Secret; a Secret holds at
	// most 1 MiB.
	maxCABundle = 1 << 20
)

// Endpoint errors.
var (
	// ErrBaseURL reports a base URL that is not https://host[:port][/path]
	// with no user, query or fragment.
	ErrBaseURL = errors.New("a base URL must be https://host[:port][/path], with no user, query or fragment")
	// ErrCABundle reports a CA bundle that holds no PEM certificate, or one
	// that doesn't parse.
	ErrCABundle = errors.New("the CA bundle holds no valid PEM certificate")
	// ErrOutsideEndpoint reports a request, a redirect included, to anywhere
	// but the endpoint's base URL and the paths below it.
	ErrOutsideEndpoint = errors.New("request outside the endpoint's base URL")
)

// Endpoint is an external HTTPS service a connector calls: the base URL the
// owner configured, or the provider's real one by default, and the extra CA
// certificates the client trusts on top of the system's roots, never instead
// of them.
type Endpoint struct {
	// BaseURL is the service's https base URL without a trailing slash,
	// such as https://api.nbp.pl, as BaseURL returns it.
	BaseURL string
	// CABundle holds PEM certificates trusted besides the system's roots;
	// nil for none.
	CABundle []byte
}

// BaseURL returns the base URL a connector calls: configured, or def when
// configured is empty. It must be https://host[:port][/path], with no user,
// query or fragment, so a plain http URL is refused; the result has no
// trailing slash, so a path appends to it as base + "/api/...".
func BaseURL(configured, def string) (string, error) {
	raw := configured
	if raw == "" {
		raw = def
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrBaseURL, raw)
	}
	if !validBaseURL(u) {
		return "", fmt.Errorf("%w: %q", ErrBaseURL, raw)
	}
	return "https://" + u.Host + strings.TrimRight(u.EscapedPath(), "/"), nil
}

// validBaseURL reports whether u is https://host[:port][/path], with no
// user, query or fragment.
func validBaseURL(u *url.URL) bool {
	return strings.EqualFold(u.Scheme, "https") && u.Opaque == "" && u.User == nil && u.Hostname() != "" &&
		u.RawQuery == "" && !u.ForceQuery && u.Fragment == "" && validPort(u)
}

// NewEndpoint returns the Endpoint at BaseURL(configured, def) that trusts
// bundle's certificates besides the system's roots; a nil or empty bundle
// adds none. A bundle without a valid PEM certificate is refused.
func NewEndpoint(configured, def string, bundle []byte) (Endpoint, error) {
	base, err := BaseURL(configured, def)
	if err != nil {
		return Endpoint{}, err
	}
	e := Endpoint{BaseURL: base}
	if strings.TrimSpace(string(bundle)) != "" {
		if _, err := appendBundle(x509.NewCertPool(), bundle); err != nil {
			return Endpoint{}, err
		}
		e.CABundle = bundle
	}
	return e, nil
}

// LoadEndpoint is NewEndpoint with the bundle the instance's CA bundle
// Secret holds at CABundleDir/CABundleFile; without that Secret, or that
// key, it adds no certificate.
func LoadEndpoint(configured, def string) (Endpoint, error) {
	bundle, err := ReadCABundle(filepath.Join(CABundleDir, CABundleFile))
	if err != nil {
		return Endpoint{}, err
	}
	return NewEndpoint(configured, def, bundle)
}

// ReadCABundle reads the CA bundle at path; a missing file, which is what an
// optional Secret the owner hasn't created leaves, reads as none.
func ReadCABundle(path string) ([]byte, error) {
	f, err := os.Open(filepath.Clean(path))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCABundle, err)
	}
	b, readErr := io.ReadAll(io.LimitReader(f, maxCABundle+1))
	if closeErr := f.Close(); readErr == nil {
		readErr = closeErr
	}
	switch {
	case readErr != nil:
		return nil, fmt.Errorf("%w: %w", ErrCABundle, readErr)
	case len(b) > maxCABundle:
		return nil, fmt.Errorf("%w: over %d bytes", ErrCABundle, maxCABundle)
	}
	return b, nil
}

// Roots returns a copy of system with the CABundle's certificates added; a
// nil system starts from an empty pool. system itself is not changed.
func (e Endpoint) Roots(system *x509.CertPool) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	if system != nil {
		pool = system.Clone()
	}
	return appendBundle(pool, e.CABundle)
}

// TLSConfig returns the TLS configuration for the endpoint: the roots of
// Roots(system), TLS 1.2 or later.
func (e Endpoint) TLSConfig(system *x509.CertPool) (*tls.Config, error) {
	roots, err := e.Roots(system)
	if err != nil {
		return nil, err
	}
	return &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}, nil
}

// Within returns a RoundTripper that sends through next only requests to the
// BaseURL or below it, and refuses every other one with ErrOutsideEndpoint,
// a redirect elsewhere included, so a credential sent to the endpoint never
// leaves it.
func (e Endpoint) Within(next http.RoundTripper) http.RoundTripper {
	return &endpointTransport{base: e.BaseURL, next: next}
}

// NewEgressClient returns an HTTP client for the endpoint: through the egress
// proxy like NewEgressClient, trusting the system's roots plus the
// CABundle, and confined to the BaseURL (Within).
func (e Endpoint) NewEgressClient() (*http.Client, error) {
	hc, err := NewEgressClient()
	if err != nil {
		return nil, err
	}
	et, ok := hc.Transport.(*EgressTransport)
	if !ok {
		return nil, ErrNoEgressProxy
	}
	if len(e.CABundle) > 0 {
		system, sysErr := x509.SystemCertPool()
		if sysErr != nil {
			return nil, fmt.Errorf("system roots: %w", sysErr)
		}
		cfg, cfgErr := e.TLSConfig(system)
		if cfgErr != nil {
			return nil, cfgErr
		}
		et.TLSClientConfig = cfg
	}
	hc.Transport = e.Within(et)
	return hc, nil
}

// endpointTransport is Endpoint.Within's RoundTripper.
type endpointTransport struct {
	base string
	next http.RoundTripper
}

// RoundTrip sends req through next when it is within base.
func (t *endpointTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !within(t.base, req.URL) {
		return nil, closeBody(req, fmt.Errorf("%w: %s", ErrOutsideEndpoint, req.URL.Redacted()))
	}
	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("openfief endpoint: %w", err)
	}
	return resp, nil
}

// CloseIdleConnections closes next's idle connections.
func (t *endpointTransport) CloseIdleConnections() {
	if c, ok := t.next.(interface{ CloseIdleConnections() }); ok {
		c.CloseIdleConnections()
	}
}

// within reports whether u is the https base URL or a path below it, on the
// same host and port; port 443 written or not is the same.
func within(base string, u *url.URL) bool {
	b, err := url.Parse(base)
	if err != nil || !strings.EqualFold(u.Scheme, "https") || u.User != nil || !sameHost(b, u) {
		return false
	}
	bp := b.EscapedPath()
	p := u.EscapedPath()
	return bp == "" || p == bp || strings.HasPrefix(p, bp+"/")
}

// sameHost reports whether a and b name the same host and port, with 443
// for none.
func sameHost(a, b *url.URL) bool {
	return strings.EqualFold(a.Hostname(), b.Hostname()) && portOr443(a) == portOr443(b)
}

// validPort reports whether u names no port, or one from 1 to 65535.
func validPort(u *url.URL) bool {
	if strings.HasSuffix(u.Host, ":") {
		return false
	}
	p := u.Port()
	if p == "" {
		return true
	}
	n, err := strconv.ParseUint(p, 10, 16)
	return err == nil && n > 0
}

// portOr443 is u's port, or "443" when it names none.
func portOr443(u *url.URL) string {
	if p := u.Port(); p != "" {
		return p
	}
	return "443"
}

// appendBundle adds bundle's PEM certificates to pool. A non-empty bundle
// with no certificate, or with one that doesn't parse, is refused.
func appendBundle(pool *x509.CertPool, bundle []byte) (*x509.CertPool, error) {
	if strings.TrimSpace(string(bundle)) == "" {
		return pool, nil
	}
	rest, n := bundle, 0
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCABundle, err)
		}
		pool.AddCert(cert)
		n++
	}
	if n == 0 {
		return nil, ErrCABundle
	}
	return pool, nil
}
