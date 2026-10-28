package identity

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Errors about origins.
var (
	// ErrInvalidOrigin reports text that is not an origin OpenFief accepts:
	// https://host[:port], or http://localhost[:port].
	ErrInvalidOrigin = errors.New("invalid origin")
	// ErrOriginNotAllowed reports an origin that is not on the allow-list.
	ErrOriginNotAllowed = errors.New("origin not allowed")
)

const (
	schemeHTTPS = "https"
	schemeHTTP  = "http"
	// localhost is the one host plain http is allowed on: browsers treat
	// http://localhost as a secure context, so WebAuthn works there (D-60).
	localhost = "localhost"
	// maxHostLen is the longest DNS name.
	maxHostLen = 253
	// maxPort is the highest TCP port.
	maxPort = 65535
	// maxLabelLen is the longest DNS label.
	maxLabelLen = 63
)

// Origin is a web origin the owner's browser uses to reach OpenFief:
// scheme, host and port. Its host is the WebAuthn relying party ID, so a
// passkey created on one origin's host works only there (D-60).
type Origin struct {
	scheme string
	host   string
	// port is 0 for the scheme's default port.
	port int
}

// ParseOrigin parses an origin as a browser sends it in the Origin header:
// "https://" or "http://", a lowercase DNS host name, and an optional port,
// with no path, query, fragment or user info. Plain http is accepted only
// for localhost, the one host browsers treat as secure without TLS. IP
// addresses are refused: WebAuthn relying party IDs are domains.
func ParseOrigin(text string) (Origin, error) {
	scheme, rest, ok := strings.Cut(text, "://")
	if !ok || (scheme != schemeHTTPS && scheme != schemeHTTP) {
		return Origin{}, fmt.Errorf("%w: %q: want https:// or http://", ErrInvalidOrigin, text)
	}
	if strings.ContainsAny(rest, "/?#@\\ ") {
		return Origin{}, fmt.Errorf("%w: %q: want scheme, host and port only", ErrInvalidOrigin, text)
	}
	host, portText, hasPort := strings.Cut(rest, ":")
	o := Origin{scheme: scheme, host: host}
	if hasPort {
		p, ok := parsePort(portText)
		if !ok {
			return Origin{}, fmt.Errorf("%w: %q: bad port", ErrInvalidOrigin, text)
		}
		if p != defaultPort(scheme) {
			o.port = p
		}
	}
	if err := checkHost(scheme, host); err != nil {
		return Origin{}, fmt.Errorf("%w: %q: %w", ErrInvalidOrigin, text, err)
	}
	return o, nil
}

// Reasons checkHost gives.
var (
	errNotDNSName = errors.New("want a lowercase DNS host name")
	errPlainHTTP  = errors.New("plain http only on localhost")
)

// checkHost checks that host is a DNS name, and localhost when scheme is
// plain http.
func checkHost(scheme, host string) error {
	if !validHost(host) {
		return errNotDNSName
	}
	if scheme == schemeHTTP && host != localhost {
		return errPlainHTTP
	}
	return nil
}

// parsePort parses a TCP port, 1 to 65535, written without leading zeros.
func parsePort(text string) (int, bool) {
	p, err := strconv.Atoi(text)
	return p, err == nil && p >= 1 && p <= maxPort && strconv.Itoa(p) == text
}

// defaultPort is the scheme's port when an origin names none.
func defaultPort(scheme string) int {
	const httpsPort, httpPort = 443, 80
	if scheme == schemeHTTPS {
		return httpsPort
	}
	return httpPort
}

// validHost reports whether host is a lowercase DNS name with at least one
// letter, so that no IP address passes.
func validHost(host string) bool {
	if host == "" || len(host) > maxHostLen {
		return false
	}
	letter := false
	for label := range strings.SplitSeq(host, ".") {
		if !validLabel(label) {
			return false
		}
		letter = letter || strings.ContainsFunc(label, func(r rune) bool { return r >= 'a' && r <= 'z' })
	}
	return letter
}

// validLabel reports whether label is a DNS label of lowercase letters,
// digits and inner hyphens.
func validLabel(label string) bool {
	if label == "" || len(label) > maxLabelLen || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
		return false
	}
	for _, r := range label {
		if (r < 'a' || r > 'z') && !strings.ContainsRune(labelOthers, r) {
			return false
		}
	}
	return true
}

// labelOthers are the characters besides lowercase letters a DNS label may
// hold.
const labelOthers = "0123456789-"

// String returns the origin as a browser serializes it, such as
// "https://dev.openfief.dev" or "http://localhost:8080".
func (o Origin) String() string {
	if o.port == 0 {
		return o.scheme + "://" + o.host
	}
	return o.scheme + "://" + o.host + ":" + strconv.Itoa(o.port)
}

// RPID returns the WebAuthn relying party ID: the origin's host.
func (o Origin) RPID() string { return o.host }

// Equal reports whether o and p are the same origin.
func (o Origin) Equal(p Origin) bool { return o == p }

// IsZero reports whether o is the zero Origin.
func (o Origin) IsZero() bool { return o.scheme == "" }

// Origins is the allow-list of origins the core serves the owner on.
type Origins struct {
	list []Origin
}

// NewOrigins parses texts into an allow-list, refusing an invalid origin and
// an empty list.
func NewOrigins(texts []string) (Origins, error) {
	if len(texts) == 0 {
		return Origins{}, fmt.Errorf("%w: the allow-list is empty", ErrInvalidOrigin)
	}
	out := make([]Origin, 0, len(texts))
	for _, t := range texts {
		o, err := ParseOrigin(t)
		if err != nil {
			return Origins{}, err
		}
		if !slices.Contains(out, o) {
			out = append(out, o)
		}
	}
	return Origins{list: out}, nil
}

// Allowed parses text and returns it when it is on the allow-list, or an
// error wrapping ErrInvalidOrigin or ErrOriginNotAllowed.
func (a Origins) Allowed(text string) (Origin, error) {
	o, err := ParseOrigin(text)
	if err != nil {
		return Origin{}, err
	}
	if !slices.Contains(a.list, o) {
		return Origin{}, fmt.Errorf("%w: %s", ErrOriginNotAllowed, o)
	}
	return o, nil
}

// Equal reports whether a and b allow the same origins in the same order.
func (a Origins) Equal(b Origins) bool { return slices.Equal(a.list, b.list) }

// List returns the allowed origins, in the order given.
func (a Origins) List() []Origin { return slices.Clone(a.list) }
