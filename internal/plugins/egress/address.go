// Package egress holds the egress proxy's destination checks that need
// net/netip, which the domain packages may not import. The proxy adapter
// applies CheckAddress to every address it is about to dial, after it has
// resolved the allowed hostname itself (D-32).
package egress

import (
	"errors"
	"fmt"
	"net/netip"
)

// ErrForbiddenAddress reports a destination in a private or internal range.
var ErrForbiddenAddress = errors.New("forbidden destination address")

// Byte offsets of an IPv4 address embedded in an IPv6 address.
const (
	nat64V4Offset  = 12 // 64:ff9b::/96, RFC 6052
	sixToFourStart = 2  // 2002::/16, RFC 3056
	ipv4Len        = 4
)

// ForbiddenPrefixes returns the address ranges the egress proxy refuses:
// private, loopback, link-local (including the cloud metadata address
// 169.254.169.254), carrier-grade NAT, unique-local, multicast, unspecified,
// documentation, benchmarking and reserved ranges, for IPv4 and IPv6.
// IPv4-mapped IPv6 addresses (::ffff:0:0/96) are not listed: CheckAddress
// checks them as the IPv4 address they map to. The proxy adapter can also
// hand this list to Smokescreen as its deny ranges.
func ForbiddenPrefixes() []netip.Prefix {
	return []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),       // "this network", including unspecified
		netip.MustParsePrefix("10.0.0.0/8"),      // RFC 1918 private
		netip.MustParsePrefix("100.64.0.0/10"),   // RFC 6598 carrier-grade NAT
		netip.MustParsePrefix("127.0.0.0/8"),     // loopback
		netip.MustParsePrefix("169.254.0.0/16"),  // link-local, cloud metadata
		netip.MustParsePrefix("172.16.0.0/12"),   // RFC 1918 private
		netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
		netip.MustParsePrefix("192.0.2.0/24"),    // documentation (TEST-NET-1)
		netip.MustParsePrefix("192.88.99.0/24"),  // 6to4 relay anycast
		netip.MustParsePrefix("192.168.0.0/16"),  // RFC 1918 private
		netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking
		netip.MustParsePrefix("198.51.100.0/24"), // documentation (TEST-NET-2)
		netip.MustParsePrefix("203.0.113.0/24"),  // documentation (TEST-NET-3)
		netip.MustParsePrefix("224.0.0.0/4"),     // multicast
		netip.MustParsePrefix("240.0.0.0/4"),     // reserved, including broadcast
		netip.MustParsePrefix("::/96"),           // unspecified, loopback, IPv4-compatible
		netip.MustParsePrefix("64:ff9b:1::/48"),  // local-use NAT64
		netip.MustParsePrefix("100::/64"),        // discard-only
		netip.MustParsePrefix("2001::/23"),       // IETF protocol assignments, Teredo
		netip.MustParsePrefix("2001:db8::/32"),   // documentation
		netip.MustParsePrefix("3fff::/20"),       // documentation
		netip.MustParsePrefix("fc00::/7"),        // unique-local (ULA)
		netip.MustParsePrefix("fe80::/10"),       // link-local
		netip.MustParsePrefix("fec0::/10"),       // site-local (deprecated)
		netip.MustParsePrefix("ff00::/8"),        // multicast
	}
}

// CheckAddress returns nil when the proxy may dial addr, and an error wrapping
// ErrForbiddenAddress otherwise. IPv4-mapped IPv6 addresses are checked as
// IPv4, and so are the IPv4 addresses embedded in NAT64 (64:ff9b::/96) and
// 6to4 (2002::/16) addresses, so a private IPv4 address cannot hide inside an
// IPv6 one. Invalid and zoned addresses are refused.
func CheckAddress(addr netip.Addr) error {
	if !addr.IsValid() || addr.Zone() != "" {
		return fmt.Errorf("%w: %q", ErrForbiddenAddress, addr.String())
	}
	if embedded, ok := embeddedIPv4(addr); ok {
		if err := CheckAddress(embedded); err != nil {
			return fmt.Errorf("%w: %q embeds %q", ErrForbiddenAddress, addr.String(), embedded.String())
		}
	}
	plain := addr.Unmap()
	for _, p := range ForbiddenPrefixes() {
		if p.Contains(plain) || p.Contains(addr) {
			return fmt.Errorf("%w: %q is in %s", ErrForbiddenAddress, addr.String(), p)
		}
	}
	if !plain.IsGlobalUnicast() {
		return fmt.Errorf("%w: %q is not global unicast", ErrForbiddenAddress, addr.String())
	}
	return nil
}

// embeddedIPv4 returns the IPv4 address inside a NAT64 or 6to4 address.
func embeddedIPv4(addr netip.Addr) (netip.Addr, bool) {
	if !addr.Is6() || addr.Is4In6() {
		return netip.Addr{}, false
	}
	b := addr.As16()
	var start int
	switch {
	case netip.MustParsePrefix("64:ff9b::/96").Contains(addr):
		start = nat64V4Offset
	case netip.MustParsePrefix("2002::/16").Contains(addr):
		start = sixToFourStart
	default:
		return netip.Addr{}, false
	}
	return netip.AddrFrom4([ipv4Len]byte(b[start : start+ipv4Len])), true
}
