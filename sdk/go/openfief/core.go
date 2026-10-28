package openfief

import (
	"errors"
	"fmt"
	"net/http"
)

// CoreHTTPClient returns an HTTP client for a plugin's calls to the core,
// such as a view plugin's calls to the read API. Every request carries
// "Authorization: Bearer <token>" with the token tokens returns at that
// moment (D-27). It never uses a proxy: plugins reach the core directly.
func CoreHTTPClient(tokens TokenSource) *http.Client {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	t := base.Clone()
	t.Proxy = nil
	return &http.Client{Transport: &bearerTransport{base: t, tokens: tokens}}
}

// bearerTransport sets the Authorization header of every request it sends.
type bearerTransport struct {
	base   http.RoundTripper
	tokens TokenSource
}

// RoundTrip sends a copy of req with the current token; req itself is not
// changed.
func (b *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := b.tokens.Token()
	if err != nil {
		return nil, closeBody(req, fmt.Errorf("plugin token: %w", err))
	}
	out := req.Clone(req.Context())
	out.Header.Set("Authorization", "Bearer "+token)
	resp, err := b.base.RoundTrip(out)
	if err != nil {
		return nil, fmt.Errorf("openfief core: %w", err)
	}
	return resp, nil
}

// closeBody closes the body of a request that will not be sent, as a
// RoundTripper must, and returns err joined with any close error.
func closeBody(req *http.Request, err error) error {
	if req.Body == nil {
		return err
	}
	return errors.Join(err, req.Body.Close())
}
