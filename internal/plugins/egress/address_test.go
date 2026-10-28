package egress_test

import (
	"net/netip"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins/egress"
)

func TestCheckAddress(t *testing.T) {
	t.Parallel()
	forbidden := egress.ErrForbiddenAddress
	tests := []struct {
		name    string
		addr    string
		wantErr error
	}{
		// Public addresses (well-known public resolvers) are allowed.
		{name: "public IPv4", addr: "9.9.9.9"},
		{name: "public IPv4 near private 172.16/12", addr: "172.32.0.1"},
		{name: "public IPv4 near CGNAT", addr: "100.128.0.1"},
		{name: "public IPv6", addr: "2620:fe::fe"},
		{name: "NAT64 of a public IPv4", addr: "64:ff9b::909:909"},
		{name: "6to4 of a public IPv4", addr: "2002:909:909::1"},
		// IPv4 private and internal ranges.
		{name: "unspecified IPv4", addr: "0.0.0.0", wantErr: forbidden},
		{name: "this network", addr: "0.1.2.3", wantErr: forbidden},
		{name: "RFC 1918 10/8", addr: "10.1.2.3", wantErr: forbidden},
		{name: "RFC 1918 172.16/12", addr: "172.31.255.254", wantErr: forbidden},
		{name: "RFC 1918 192.168/16", addr: "192.168.1.1", wantErr: forbidden},
		{name: "loopback", addr: "127.0.0.1", wantErr: forbidden},
		{name: "loopback range", addr: "127.255.0.1", wantErr: forbidden},
		{name: "link-local", addr: "169.254.1.1", wantErr: forbidden},
		{name: "cloud metadata", addr: "169.254.169.254", wantErr: forbidden},
		{name: "carrier-grade NAT", addr: "100.64.0.1", wantErr: forbidden},
		{name: "carrier-grade NAT top", addr: "100.127.255.254", wantErr: forbidden},
		{name: "IETF protocol assignments", addr: "192.0.0.8", wantErr: forbidden},
		{name: "documentation", addr: "192.0.2.10", wantErr: forbidden},
		{name: "benchmarking", addr: "198.18.0.1", wantErr: forbidden},
		{name: "multicast", addr: "224.0.0.251", wantErr: forbidden},
		{name: "reserved", addr: "240.0.0.1", wantErr: forbidden},
		{name: "broadcast", addr: "255.255.255.255", wantErr: forbidden},
		// IPv6 private and internal ranges.
		{name: "unspecified IPv6", addr: "::", wantErr: forbidden},
		{name: "loopback IPv6", addr: "::1", wantErr: forbidden},
		{name: "unique-local", addr: "fd12:3456:789a::1", wantErr: forbidden},
		{name: "unique-local metadata", addr: "fd00:ec2::254", wantErr: forbidden},
		{name: "link-local IPv6", addr: "fe80::1", wantErr: forbidden},
		{name: "link-local IPv6 with zone", addr: "fe80::1%eth0", wantErr: forbidden},
		{name: "site-local", addr: "fec0::1", wantErr: forbidden},
		{name: "multicast IPv6", addr: "ff02::1", wantErr: forbidden},
		{name: "documentation IPv6", addr: "2001:db8::1", wantErr: forbidden},
		{name: "Teredo", addr: "2001::1", wantErr: forbidden},
		{name: "discard-only", addr: "100::1", wantErr: forbidden},
		{name: "local-use NAT64", addr: "64:ff9b:1::a00:1", wantErr: forbidden},
		// Private IPv4 hidden inside IPv6.
		{name: "IPv4-mapped loopback", addr: "::ffff:127.0.0.1", wantErr: forbidden},
		{name: "IPv4-mapped metadata", addr: "::ffff:169.254.169.254", wantErr: forbidden},
		{name: "IPv4-mapped public", addr: "::ffff:9.9.9.9"},
		{name: "IPv4-compatible private", addr: "::10.0.0.1", wantErr: forbidden},
		{name: "NAT64 of a private IPv4", addr: "64:ff9b::a00:1", wantErr: forbidden},
		{name: "NAT64 of metadata", addr: "64:ff9b::a9fe:a9fe", wantErr: forbidden},
		{name: "6to4 of a private IPv4", addr: "2002:c0a8:101::1", wantErr: forbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			addr, err := netip.ParseAddr(tt.addr)
			if err != nil {
				t.Fatalf("ParseAddr(%q): %v", tt.addr, err)
			}
			if diff := cmp.Diff(tt.wantErr, egress.CheckAddress(addr), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("CheckAddress(%s) (-want +got):\n%s", tt.addr, diff)
			}
		})
	}
}

func TestCheckAddressInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		addr netip.Addr
	}{
		{name: "zero value", addr: netip.Addr{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(egress.ErrForbiddenAddress, egress.CheckAddress(tt.addr), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("CheckAddress() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestForbiddenPrefixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prefix string
	}{
		{name: "RFC 1918 10/8", prefix: "10.0.0.0/8"},
		{name: "RFC 1918 172.16/12", prefix: "172.16.0.0/12"},
		{name: "RFC 1918 192.168/16", prefix: "192.168.0.0/16"},
		{name: "loopback", prefix: "127.0.0.0/8"},
		{name: "link-local and metadata", prefix: "169.254.0.0/16"},
		{name: "carrier-grade NAT", prefix: "100.64.0.0/10"},
		{name: "unique-local", prefix: "fc00::/7"},
		{name: "link-local IPv6", prefix: "fe80::/10"},
		{name: "multicast IPv6", prefix: "ff00::/8"},
	}
	got := egress.ForbiddenPrefixes()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if want := netip.MustParsePrefix(tt.prefix); !slices.Contains(got, want) {
				t.Errorf("ForbiddenPrefixes() lacks %s", want)
			}
		})
	}
}
