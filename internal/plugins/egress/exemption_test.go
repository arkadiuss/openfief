package egress_test

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins/egress"
)

// Synthetic names and addresses: the exempt host is a Service-like name at a
// private address, as the end-to-end suite's fixture is.
const (
	exemptHost = "fixture.tests.svc.cluster.local"
	exemptAddr = "10.43.0.12"
)

func TestParseExemption(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		in       string
		wantHost string
		want     string
		wantErr  error
	}{
		{name: "no exemption"},
		{name: "blank is none", in: "   "},
		{name: "one address", in: exemptHost + "=" + exemptAddr, wantHost: exemptHost, want: exemptHost + "=" + exemptAddr},
		{
			name: "two addresses, sorted and once", in: exemptHost + "=fd00::5,10.43.0.12, 10.43.0.12",
			wantHost: exemptHost, want: exemptHost + "=10.43.0.12,fd00::5",
		},
		{name: "host in any case, with a trailing dot", in: "Fixture.Tests.SVC.cluster.local.=" + exemptAddr, wantHost: exemptHost, want: exemptHost + "=" + exemptAddr},
		{name: "IPv4-mapped address is its IPv4 form", in: exemptHost + "=::ffff:10.43.0.12", wantHost: exemptHost, want: exemptHost + "=" + exemptAddr},
		{name: "no addresses", in: exemptHost + "=", wantErr: egress.ErrInvalidExemption},
		{name: "no equals sign", in: exemptHost, wantErr: egress.ErrInvalidExemption},
		{name: "an address that doesn't parse", in: exemptHost + "=10.43.0", wantErr: egress.ErrInvalidExemption},
		{name: "a zoned address", in: exemptHost + "=fe80::1%eth0", wantErr: egress.ErrInvalidExemption},
		{name: "a range", in: exemptHost + "=10.43.0.0/16", wantErr: egress.ErrInvalidExemption},
		{name: "a wildcard host", in: "*.tests.svc.cluster.local=" + exemptAddr, wantErr: egress.ErrInvalidExemption},
		{name: "an IP address as the host", in: exemptAddr + "=" + exemptAddr, wantErr: egress.ErrInvalidExemption},
		{name: "an IPv6 address as the host", in: "fd00::5=fd00::5", wantErr: egress.ErrInvalidExemption},
		{name: "a host with a port", in: exemptHost + ":443=" + exemptAddr, wantErr: egress.ErrInvalidExemption},
		{name: "an empty host", in: "=" + exemptAddr, wantErr: egress.ErrInvalidExemption},
		{name: "a second host", in: exemptHost + "=" + exemptAddr + ";other.tests.svc=" + exemptAddr, wantErr: egress.ErrInvalidExemption},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := egress.ParseExemption(tt.in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got.Host() != tt.wantHost || got.String() != tt.want || got.IsSet() != (tt.want != "") {
				t.Errorf("exemption = host %q, %q, set %t; want host %q, %q", got.Host(), got.String(), got.IsSet(), tt.wantHost, tt.want)
			}
		})
	}
}

func addrs(t *testing.T, list ...string) []netip.Addr {
	t.Helper()
	out := make([]netip.Addr, len(list))
	for i, s := range list {
		out[i] = netip.MustParseAddr(s)
	}
	return out
}

// TestCheckResolved covers D-215: the exempt host at exactly its addresses
// passes and everything else is refused as before.
func TestCheckResolved(t *testing.T) {
	t.Parallel()
	one, err := egress.ParseExemption(exemptHost + "=" + exemptAddr)
	if err != nil {
		t.Fatal(err)
	}
	two, err := egress.ParseExemption(exemptHost + "=" + exemptAddr + ",10.43.0.13")
	if err != nil {
		t.Fatal(err)
	}
	// The node's own address is an extra deny range, which the exemption
	// overrides for the exempt host and nothing else.
	deny, err := egress.ParseDenyList("", exemptAddr)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		host    string
		ips     []string
		exempt  egress.Exemption
		wantErr bool
	}{
		{name: "the exempt host at its address", host: exemptHost, ips: []string{exemptAddr}, exempt: one},
		{name: "the exempt host in another case, with a trailing dot", host: "FIXTURE.tests.svc.cluster.local.", ips: []string{exemptAddr}, exempt: one},
		{name: "the exempt host at its IPv4-mapped address", host: exemptHost, ips: []string{"::ffff:" + exemptAddr}, exempt: one},
		{name: "the exempt host listing its address twice", host: exemptHost, ips: []string{exemptAddr, exemptAddr}, exempt: one},
		{name: "the exempt host at both of its addresses", host: exemptHost, ips: []string{"10.43.0.13", exemptAddr}, exempt: two},
		{name: "another address in the same range", host: "other.tests.svc.cluster.local", ips: []string{"10.43.0.99"}, exempt: one, wantErr: true},
		{name: "another host at the exempt address", host: "other.tests.svc.cluster.local", ips: []string{exemptAddr}, exempt: one, wantErr: true},
		{name: "a parent domain of the exempt host", host: "tests.svc.cluster.local", ips: []string{exemptAddr}, exempt: one, wantErr: true},
		{name: "a subdomain of the exempt host", host: "a." + exemptHost, ips: []string{exemptAddr}, exempt: one, wantErr: true},
		{name: "the exempt host at another address in the range", host: exemptHost, ips: []string{"10.43.0.99"}, exempt: one, wantErr: true},
		{name: "the exempt host at its address and a second one", host: exemptHost, ips: []string{exemptAddr, "10.43.0.99"}, exempt: one, wantErr: true},
		{name: "the exempt host at its address and a public one", host: exemptHost, ips: []string{exemptAddr, "8.8.4.4"}, exempt: one, wantErr: true},
		{name: "the exempt host at only one of two exempt addresses", host: exemptHost, ips: []string{exemptAddr}, exempt: two, wantErr: true},
		{name: "the exempt host at no address", host: exemptHost, exempt: one, wantErr: true},
		{name: "the exempt address as an IP literal", host: exemptAddr, ips: []string{exemptAddr}, exempt: one, wantErr: true},
		{name: "no exemption: the private address", host: exemptHost, ips: []string{exemptAddr}, wantErr: true},
		{name: "no exemption: a public address", host: "example.org", ips: []string{"8.8.4.4"}},
		{name: "no exemption: public and private mixed", host: "example.org", ips: []string{"8.8.4.4", exemptAddr}, wantErr: true},
		{name: "an exemption leaves public hosts alone", host: "example.org", ips: []string{"8.8.4.4"}, exempt: one},
		{name: "an exemption leaves a loopback address refused", host: "example.org", ips: []string{"127.0.0.1"}, exempt: one, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := egress.CheckResolved(tt.host, addrs(t, tt.ips...), deny, tt.exempt)
			if tt.wantErr != errors.Is(err, egress.ErrForbiddenAddress) {
				t.Errorf("CheckResolved(%q, %v) = %v, want forbidden: %t", tt.host, tt.ips, err, tt.wantErr)
			}
		})
	}
}
