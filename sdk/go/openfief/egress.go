package openfief

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
)

// EnvHTTPSProxy is the environment variable that holds the egress proxy's
// URL, with its ClusterIP since plugins have no DNS. OpenFief sets it only
// for a plugin with granted egress hosts (D-32).
const EnvHTTPSProxy = "HTTPS_PROXY"

// EgressProxyUser is the user name in the proxy URL; the token is its
// password (D-63).
const EgressProxyUser = "plugin"

// Egress errors.
var (
	// ErrNoEgressProxy reports a plugin without an egress proxy: it has no
	// granted egress hosts, so it can reach nothing outside the cluster.
	ErrNoEgressProxy = errors.New("no egress proxy (HTTPS_PROXY is not set)")
	// ErrNoEgressToken reports a plugin without a token for the egress
	// proxy (OPENFIEF_EGRESS_TOKEN_FILE is not set).
	ErrNoEgressToken = errors.New("no egress proxy token (OPENFIEF_EGRESS_TOKEN_FILE is not set)")
	// ErrNotHTTPS reports an egress request that is not https: the egress
	// proxy only opens HTTP CONNECT tunnels.
	ErrNotHTTPS = errors.New("egress requests must use https")
)

// EgressTransport is an http.RoundTripper for https requests to granted
// external hosts through the egress proxy (D-32, D-63). For every request
// it builds the proxy URL anew with the plugin's current token as the
// password, which Go sends as the CONNECT's Proxy-Authorization header; the
// proxy identifies the plugin by it. Tokens rotate, so a proxy URL with a
// fixed password would stop working.
//
// Set the fields before the first request and don't change them after.
type EgressTransport struct {
	// Proxy is the egress proxy's URL, as HTTPS_PROXY holds it. Required.
	Proxy *url.URL
	// Tokens returns the plugin's current token for the egress proxy
	// (audience openfief-egress). Required.
	Tokens TokenSource
	// TLSClientConfig configures TLS to the external hosts; nil means the
	// system's roots.
	TLSClientConfig *tls.Config

	once      sync.Once
	transport *http.Transport
}

// NewEgressClient returns an HTTP client whose requests go through the egress
// proxy named by HTTPS_PROXY, with the token in the file
// OPENFIEF_EGRESS_TOKEN_FILE names.
func NewEgressClient() (*http.Client, error) {
	raw := os.Getenv(EnvHTTPSProxy)
	if raw == "" {
		return nil, ErrNoEgressProxy
	}
	proxy, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoEgressProxy, err)
	}
	tokenFile := os.Getenv(EnvEgressIdentityFile)
	if tokenFile == "" {
		return nil, ErrNoEgressToken
	}
	return &http.Client{Transport: &EgressTransport{Proxy: proxy, Tokens: TokenFile{Path: tokenFile}}}, nil
}

// RoundTrip sends req through the egress proxy. It refuses a request that is
// not https, and any request when Proxy is not set, rather than send it
// without the proxy.
func (e *EgressTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch {
	case e.Proxy == nil:
		return nil, closeBody(req, ErrNoEgressProxy)
	case req.URL.Scheme != "https":
		return nil, closeBody(req, fmt.Errorf("%w: %q", ErrNotHTTPS, req.URL.Scheme))
	}
	resp, err := e.base().RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("openfief egress: %w", err)
	}
	return resp, nil
}

// CloseIdleConnections closes idle tunnels.
func (e *EgressTransport) CloseIdleConnections() {
	e.base().CloseIdleConnections()
}

// base returns the underlying transport, built once from the default one.
func (e *EgressTransport) base() *http.Transport {
	e.once.Do(func() {
		def, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			def = &http.Transport{}
		}
		t := def.Clone()
		t.Proxy = e.proxyURL
		if e.TLSClientConfig != nil {
			t.TLSClientConfig = e.TLSClientConfig.Clone()
		}
		e.transport = t
	})
	return e.transport
}

// proxyURL returns the proxy URL for one request, with the token current at
// that moment as its password.
func (e *EgressTransport) proxyURL(*http.Request) (*url.URL, error) {
	token, err := e.Tokens.Token()
	if err != nil {
		return nil, fmt.Errorf("egress proxy token: %w", err)
	}
	u := *e.Proxy
	u.User = url.UserPassword(EgressProxyUser, token)
	return &u, nil
}
