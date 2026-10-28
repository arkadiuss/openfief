package plugins

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// EgressAllowList is the hostnames a plugin may reach through the egress
// proxy (D-32): the egress hosts of its effective permissions, sorted. The
// proxy finds the list by plugin name (PluginForServiceAccount) and refuses
// every destination not on it.
type EgressAllowList struct {
	Plugin string
	Hosts  []EgressHost
}

// Allows reports whether a CONNECT to host:port is on the list. ASCII letters
// compare case-insensitively and a single trailing dot is ignored, as in DNS.
// A host with any non-ASCII byte never matches (Unicode case folding would
// let, for example, the Kelvin sign stand in for "k"), and neither does an IP
// literal, because no entry is one.
func (l EgressAllowList) Allows(host string, port uint16) bool {
	for i := range len(host) {
		if host[i] >= utf8.RuneSelf {
			return false
		}
	}
	want := EgressHost{Host: strings.ToLower(strings.TrimSuffix(host, ".")), Port: port}
	return slices.Contains(l.Hosts, want)
}
