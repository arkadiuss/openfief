package plugins_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestEgressAllowListAllows(t *testing.T) {
	t.Parallel()
	list := plugins.EgressAllowList{
		Plugin: "sample",
		Hosts:  []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)},
	}
	tests := []struct {
		name string
		list plugins.EgressAllowList
		host string
		port uint16
		want bool
	}{
		{name: "listed host and port", list: list, host: "api.example.com", port: 443, want: true},
		{name: "listed non-default port", list: list, host: "data.example.org", port: 8443, want: true},
		{name: "upper case", list: list, host: "API.Example.COM", port: 443, want: true},
		{name: "trailing dot", list: list, host: "api.example.com.", port: 443, want: true},
		{name: "listed host, other port", list: list, host: "api.example.com", port: 80, want: false},
		{name: "unlisted host", list: list, host: "evil.example.net", port: 443, want: false},
		{name: "subdomain of a listed host", list: list, host: "x.api.example.com", port: 443, want: false},
		{name: "parent of a listed host", list: list, host: "example.com", port: 443, want: false},
		{name: "suffix trick", list: list, host: "api.example.com.evil.example.net", port: 443, want: false},
		{name: "IP literal", list: list, host: "192.0.2.10", port: 443, want: false},
		{name: "Kelvin sign that lower-cases to k", list: plugins.EgressAllowList{Plugin: "k", Hosts: []plugins.EgressHost{host("kelvin.example.com", 443)}}, host: "Kelvin.example.com", port: 443, want: false},
		{name: "two trailing dots", list: list, host: "api.example.com..", port: 443, want: false},
		{name: "empty list", list: plugins.EgressAllowList{Plugin: "xtb"}, host: "api.example.com", port: 443, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.list.Allows(tt.host, tt.port); got != tt.want {
				t.Errorf("Allows(%q, %d) = %v, want %v", tt.host, tt.port, got, tt.want)
			}
		})
	}
}
