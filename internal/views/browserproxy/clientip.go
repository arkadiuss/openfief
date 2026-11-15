package browserproxy

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
)

// ErrClientIPConfig reports a client address setting the core can't use: a
// header with no proxy to trust, a proxy that isn't an address or network, or a
// header name that isn't one.
var ErrClientIPConfig = errors.New("invalid client address setting")

// forwardedFor is the header whose value is a list of addresses, one added by
// each proxy the request passed.
const forwardedFor = "X-Forwarded-For"

// ClientIPConfig says whose word the core takes for the browser's address
// (D-208). Each limit per address rests on it, so it believes a header only
// from the proxies the owner named, and never one from anyone else: a browser
// that reaches the core directly, or through a proxy that is not listed, can
// write any header it likes.
type ClientIPConfig struct {
	// Header is the request header that carries the browser's address when the
	// connection comes from one of Trusted: "X-Forwarded-For", whose list is read
	// from its right end, past the trusted proxies, so that a browser that
	// writes its own list at the start gains nothing; or a header that holds
	// one address, such as "CF-Connecting-IP". Empty trusts no header.
	Header string
	// Trusted are the addresses and networks of the proxies whose Header is
	// believed: the Ingress controller's pods, and, behind a CDN, its edge.
	Trusted []netip.Prefix
}

// ParseClientIPConfig returns the setting for header and a list of trusted
// proxies, each an address or a network in CIDR form, separated by commas. No
// header is the setting that believes none; a header needs at least one
// proxy, since a header trusted from nobody would be ignored without a word.
func ParseClientIPConfig(header, proxies string) (ClientIPConfig, error) {
	header = strings.TrimSpace(header)
	var trusted []netip.Prefix
	for p := range strings.SplitSeq(proxies, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		prefix, err := parseProxy(p)
		if err != nil {
			return ClientIPConfig{}, err
		}
		trusted = append(trusted, prefix)
	}
	switch {
	case header == "" && len(trusted) > 0:
		return ClientIPConfig{}, fmt.Errorf("%w: trusted proxies without a header to read", ErrClientIPConfig)
	case header == "":
		return ClientIPConfig{}, nil
	case !validHeaderName(header) || strings.EqualFold(header, "Forwarded"):
		return ClientIPConfig{}, fmt.Errorf("%w: %q is not a header that holds an address (X-Forwarded-For, or one with a single address)", ErrClientIPConfig, header)
	case len(trusted) == 0:
		return ClientIPConfig{}, fmt.Errorf("%w: a header needs the proxies it is trusted from", ErrClientIPConfig)
	}
	return ClientIPConfig{Header: http.CanonicalHeaderKey(header), Trusted: trusted}, nil
}

// parseProxy parses an address or a network.
func parseProxy(s string) (netip.Prefix, error) {
	if prefix, err := netip.ParsePrefix(s); err == nil {
		return prefix.Masked(), nil
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("%w: %q is not an address or a network", ErrClientIPConfig, s)
	}
	addr = addr.Unmap()
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

// validHeaderName reports whether s is an HTTP header name: token characters.
func validHeaderName(s string) bool {
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', strings.ContainsRune("!#$%&'*+-.^_`|~", c):
		default:
			return false
		}
	}
	return s != ""
}

// Resolve returns the browser's address for r, or "" when the core can't tell.
// With no trusted proxy, or when the connection's peer is not one, it is the
// connection's address. When the peer is a trusted proxy it is what the proxy
// says in Header: the rightmost address of X-Forwarded-For that is not itself
// a trusted proxy, or the single address of another header; a header that is
// missing or holds nothing usable gives "", since the connection's address is
// then the proxy's own and would put every browser behind it under one limit.
func (c ClientIPConfig) Resolve(r *http.Request) string {
	peer, ok := peerAddr(r.RemoteAddr)
	if !ok {
		return ""
	}
	if c.Header == "" || !c.trusts(peer) {
		return peer.String()
	}
	values := r.Header.Values(c.Header)
	if strings.EqualFold(c.Header, forwardedFor) {
		return c.rightmostUntrusted(values)
	}
	if len(values) == 0 {
		return ""
	}
	addr, err := netip.ParseAddr(strings.TrimSpace(values[len(values)-1]))
	if err != nil {
		return ""
	}
	return addr.Unmap().WithZone("").String()
}

// rightmostUntrusted returns the last address in the comma-separated lists
// values that is not a trusted proxy, or "" for none, or for one that is not an
// address: a malformed entry ends the search, since what lies beyond it can't be
// told from what the browser wrote.
func (c ClientIPConfig) rightmostUntrusted(values []string) string {
	entries := make([]string, 0, len(values))
	for _, v := range values {
		for entry := range strings.SplitSeq(v, ",") {
			entries = append(entries, entry)
		}
	}
	for _, entry := range slices.Backward(entries) {
		addr, err := netip.ParseAddr(strings.TrimSpace(entry))
		if err != nil {
			return ""
		}
		addr = addr.Unmap().WithZone("")
		if !c.trusts(addr) {
			return addr.String()
		}
	}
	return ""
}

// trusts reports whether addr is one of the trusted proxies.
func (c ClientIPConfig) trusts(addr netip.Addr) bool {
	for _, p := range c.Trusted {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// peerAddr returns the address of a connection's remote end, "host:port".
func peerAddr(remote string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap().WithZone(""), true
}
