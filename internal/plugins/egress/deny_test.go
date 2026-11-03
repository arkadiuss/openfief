package egress_test

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins/egress"
)

func TestParseDenyList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cidrs   string
		nodeIPs string
		want    []string
		wantErr error
	}{
		{name: "nothing"},
		{
			name: "public cluster ranges and the node", cidrs: "9.9.0.0/16, 2620:fe::/48", nodeIPs: "9.9.9.9",
			want: []string{"9.9.0.0/16", "2620:fe::/48", "9.9.9.9/32"},
		},
		{name: "an IPv6 node", nodeIPs: "2620:fe::fe", want: []string{"2620:fe::fe/128"}},
		{name: "an IPv4-mapped node address", nodeIPs: "::ffff:9.9.9.9", want: []string{"9.9.9.9/32"}},
		{name: "masked and once", cidrs: "9.9.9.1/24,9.9.9.0/24,,", want: []string{"9.9.9.0/24"}},
		{name: "a range that doesn't parse", cidrs: "9.9.9.0/33", wantErr: egress.ErrInvalidDenyList},
		{name: "a node address that doesn't parse", nodeIPs: "node-1", wantErr: egress.ErrInvalidDenyList},
		{name: "a zoned node address", nodeIPs: "fe80::1%eth0", wantErr: egress.ErrInvalidDenyList},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := egress.ParseDenyList(tt.cidrs, tt.nodeIPs)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			var gotS []string
			for _, p := range got {
				gotS = append(gotS, p.String())
			}
			if diff := cmp.Diff(tt.want, gotS); diff != "" {
				t.Errorf("ranges (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckAddressDenying(t *testing.T) {
	t.Parallel()
	deny, err := egress.ParseDenyList("2620:fe::/48", "9.9.9.9")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		addr    string
		wantErr error
	}{
		{name: "another public address", addr: "8.8.4.4"},
		{name: "the node", addr: "9.9.9.9", wantErr: egress.ErrForbiddenAddress},
		{name: "the node, IPv4-mapped", addr: "::ffff:9.9.9.9", wantErr: egress.ErrForbiddenAddress},
		{name: "the node inside NAT64", addr: "64:ff9b::909:909", wantErr: egress.ErrForbiddenAddress},
		{name: "the node inside 6to4", addr: "2002:909:909::1", wantErr: egress.ErrForbiddenAddress},
		{name: "a global IPv6 cluster range", addr: "2620:fe::fe", wantErr: egress.ErrForbiddenAddress},
		{name: "still a private address", addr: "10.0.0.7", wantErr: egress.ErrForbiddenAddress},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := egress.CheckAddressDenying(netip.MustParseAddr(tt.addr), deny); !errors.Is(err, tt.wantErr) {
				t.Errorf("CheckAddressDenying(%s) = %v, want %v", tt.addr, err, tt.wantErr)
			}
		})
	}
}
