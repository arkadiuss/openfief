package egress

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"slices"
	"strings"
)

// ErrInvalidExemption reports a refused-range exemption that doesn't parse.
var ErrInvalidExemption = errors.New("invalid refused-range exemption")

// maxExemptHostLen is the longest DNS name, RFC 1035.
const maxExemptHostLen = 253

// exemptHostName is what an exempt host may be: dotted DNS labels of letters,
// numbers and hyphens, so neither an IP address nor a wildcard.
var exemptHostName = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$`)

// Exemption exempts exactly one host from the proxy's refused address ranges
// (D-215): the named hostname, when it resolves to exactly the exempt
// addresses. It is for the end-to-end suite alone; the zero value exempts
// nothing, and a shipped install never sets one. There is no range or
// wildcard form: another host, the same host resolving to any other address
// (a further one included), another address in the same range and an IP
// address without the hostname all stay refused.
type Exemption struct {
	host  string
	addrs []netip.Addr
}

// ParseExemption reads "host=address[,address...]": a DNS name and the one
// or more addresses it must resolve to, such as
// "egress-fixture.fixtures.svc.cluster.local=10.43.0.12". Blank text is no
// exemption. The host is matched case-insensitively, without a trailing dot;
// an IP address or a wildcard is not a host. Each address is exempted as
// itself, in its IPv4 form when it is IPv4-mapped.
func ParseExemption(s string) (Exemption, error) {
	if s = strings.TrimSpace(s); s == "" {
		return Exemption{}, nil
	}
	host, list, ok := strings.Cut(s, "=")
	host = normalizeHost(host)
	if !ok || len(host) > maxExemptHostLen || !exemptHostName.MatchString(host) {
		return Exemption{}, fmt.Errorf("%w: %q is not host=address[,address...] with a DNS name as host", ErrInvalidExemption, s)
	}
	if _, err := netip.ParseAddr(host); err == nil {
		return Exemption{}, fmt.Errorf("%w: the host %q is an IP address", ErrInvalidExemption, host)
	}
	var addrs []netip.Addr
	for item := range strings.SplitSeq(list, ",") {
		a, err := netip.ParseAddr(strings.TrimSpace(item))
		if err != nil || a.Zone() != "" {
			return Exemption{}, fmt.Errorf("%w: %q is not an address (host %q)", ErrInvalidExemption, item, host)
		}
		addrs = append(addrs, a.Unmap())
	}
	return Exemption{host: host, addrs: sortedUnique(addrs)}, nil
}

// IsSet reports whether e exempts a host.
func (e Exemption) IsSet() bool { return e.host != "" }

// Host returns the exempt host name, lower-case, or "" for no exemption.
func (e Exemption) Host() string { return e.host }

// Addrs returns the addresses the exempt host must resolve to, sorted.
func (e Exemption) Addrs() []netip.Addr { return slices.Clone(e.addrs) }

// String returns e as ParseExemption reads it, or "" for no exemption.
func (e Exemption) String() string {
	if !e.IsSet() {
		return ""
	}
	parts := make([]string, len(e.addrs))
	for i, a := range e.addrs {
		parts[i] = a.String()
	}
	return e.host + "=" + strings.Join(parts, ",")
}

// Covers reports whether host, having resolved to resolved, is the exempt
// host at exactly the exempt addresses: the same host name, with every
// resolved address exempt and every exempt address resolved. A host with no
// addresses, or with one more than the exemption lists, is not covered.
func (e Exemption) Covers(host string, resolved []netip.Addr) bool {
	if !e.IsSet() || normalizeHost(host) != e.host || len(resolved) == 0 {
		return false
	}
	got := make([]netip.Addr, len(resolved))
	for i, a := range resolved {
		got[i] = a.Unmap()
	}
	return slices.Equal(sortedUnique(got), e.addrs)
}

// CheckResolved returns nil when the proxy may dial host, which resolved to
// resolved, and an error wrapping ErrForbiddenAddress otherwise: every
// address must pass CheckAddressDenying with the extra ranges, unless
// exempt covers host at those addresses (D-215). With no exemption it is
// exactly that check on each address, so one refused address refuses the
// host whatever else it resolves to, and a host with no address is refused.
func CheckResolved(host string, resolved []netip.Addr, extra []netip.Prefix, exempt Exemption) error {
	if len(resolved) == 0 {
		return fmt.Errorf("%w: %q resolved to no address", ErrForbiddenAddress, host)
	}
	if exempt.Covers(host, resolved) {
		return nil
	}
	for _, a := range resolved {
		if err := CheckAddressDenying(a.Unmap(), extra); err != nil {
			return err
		}
	}
	return nil
}

// normalizeHost lower-cases host and drops one trailing dot.
func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

// sortedUnique returns addrs sorted, each once.
func sortedUnique(addrs []netip.Addr) []netip.Addr {
	out := slices.Clone(addrs)
	slices.SortFunc(out, netip.Addr.Compare)
	return slices.Compact(out)
}
