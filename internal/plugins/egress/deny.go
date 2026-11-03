package egress

import (
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"strings"
)

// ErrInvalidDenyList reports an extra deny range or node address that
// doesn't parse.
var ErrInvalidDenyList = errors.New("invalid egress deny list")

// ParseDenyList reads the ranges the egress proxy refuses beyond
// ForbiddenPrefixes (D-184): cidrs, comma-separated CIDR prefixes such as a
// cluster's pod or Service range when it isn't private, or its nodes' public
// or global IPv6 addresses; and nodeIPs, the comma-separated addresses of the
// node the proxy runs on, each refused as a single address. Blank entries
// are skipped; anything else that doesn't parse is an error. The result
// holds each prefix once, masked, in the order given.
func ParseDenyList(cidrs, nodeIPs string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	add := func(p netip.Prefix) {
		if p = p.Masked(); !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	for s := range strings.SplitSeq(cidrs, ",") {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		p, err := netip.ParsePrefix(s)
		if err != nil {
			return nil, fmt.Errorf("%w: range %q: %w", ErrInvalidDenyList, s, err)
		}
		add(p)
	}
	for s := range strings.SplitSeq(nodeIPs, ",") {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		a, err := netip.ParseAddr(s)
		if err != nil || a.Zone() != "" {
			return nil, fmt.Errorf("%w: node address %q", ErrInvalidDenyList, s)
		}
		a = a.Unmap()
		add(netip.PrefixFrom(a, a.BitLen()))
	}
	return out, nil
}

// CheckAddressDenying is CheckAddress that also refuses addresses in extra,
// the ranges ParseDenyList read: checked as CheckAddress checks its own,
// through IPv4-mapped, NAT64 and 6to4 forms too.
func CheckAddressDenying(addr netip.Addr, extra []netip.Prefix) error {
	if err := CheckAddress(addr); err != nil {
		return err
	}
	candidates := []netip.Addr{addr.Unmap()}
	if embedded, ok := embeddedIPv4(addr); ok {
		candidates = append(candidates, embedded)
	}
	for _, a := range candidates {
		for _, p := range extra {
			if p.Contains(a) {
				return fmt.Errorf("%w: %q is in the denied range %s", ErrForbiddenAddress, addr.String(), p)
			}
		}
	}
	return nil
}
