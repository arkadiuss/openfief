package plugins_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestParseEgressHost(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", 63)
	tests := []struct {
		name    string
		in      string
		want    plugins.EgressHost
		wantErr error
	}{
		{name: "hostname defaults to 443", in: "api.example.com", want: host("api.example.com", 443)},
		{name: "explicit port", in: "data.example.org:8443", want: host("data.example.org", 8443)},
		{name: "IDNA A-label", in: "xn--bcher-kva.example", want: host("xn--bcher-kva.example", 443)},
		{name: "hyphen and digits inside labels", in: "eu-1.api2.example.com", want: host("eu-1.api2.example.com", 443)},
		{name: "63-character label", in: long + ".example.com", want: host(long+".example.com", 443)},
		{name: "upper case", in: "API.example.com", wantErr: plugins.ErrInvalidHostname},
		{name: "unicode instead of A-label", in: "bücher.example", wantErr: plugins.ErrInvalidHostname},
		{name: "single label", in: "localhost", wantErr: plugins.ErrInvalidHostname},
		{name: "cluster-internal short name", in: "core", wantErr: plugins.ErrInvalidHostname},
		{name: "trailing dot", in: "api.example.com.", wantErr: plugins.ErrInvalidHostname},
		{name: "empty label", in: "api..example.com", wantErr: plugins.ErrInvalidHostname},
		{name: "leading hyphen", in: "-api.example.com", wantErr: plugins.ErrInvalidHostname},
		{name: "trailing hyphen", in: "api-.example.com", wantErr: plugins.ErrInvalidHostname},
		{name: "reserved double hyphen", in: "ab--cd.example.com", wantErr: plugins.ErrInvalidHostname},
		{name: "underscore", in: "api_v1.example.com", wantErr: plugins.ErrInvalidHostname},
		{name: "64-character label", in: long + "a.example.com", wantErr: plugins.ErrInvalidHostname},
		{name: "too long", in: strings.Repeat(long+".", 4) + "com", wantErr: plugins.ErrInvalidHostname},
		{name: "empty", in: "", wantErr: plugins.ErrInvalidHostname},
		{name: "IPv4 literal", in: "192.0.2.10", wantErr: plugins.ErrIPLiteral},
		{name: "IPv4 literal with port", in: "192.0.2.10:443", wantErr: plugins.ErrIPLiteral},
		{name: "short IPv4 form", in: "127.1", wantErr: plugins.ErrIPLiteral},
		{name: "IPv6 literal", in: "2001:db8::1", wantErr: plugins.ErrIPLiteral},
		{name: "bracketed IPv6 with port", in: "[2001:db8::1]:443", wantErr: plugins.ErrIPLiteral},
		{name: "wildcard", in: "*.example.com", wantErr: plugins.ErrWildcardHost},
		{name: "port zero", in: "api.example.com:0", wantErr: plugins.ErrInvalidPort},
		{name: "port too large", in: "api.example.com:65536", wantErr: plugins.ErrInvalidPort},
		{name: "port not a number", in: "api.example.com:https", wantErr: plugins.ErrInvalidPort},
		{name: "empty port", in: "api.example.com:", wantErr: plugins.ErrInvalidPort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.ParseEgressHost(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseEgressHost(%q) error (-want +got):\n%s", tt.in, diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseEgressHost(%q) (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}

func TestEgressHostString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   plugins.EgressHost
		want string
	}{
		{name: "default port", in: host("api.example.com", 443), want: "api.example.com:443"},
		{name: "other port", in: host("data.example.org", 8443), want: "data.example.org:8443"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.in.String()); diff != "" {
				t.Errorf("String() (-want +got):\n%s", diff)
			}
		})
	}
}
