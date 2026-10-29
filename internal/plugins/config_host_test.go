package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// viewWithURL is a view plugin whose manifest asks for the host of its
// config field url (D-75), with the given config.
func viewWithURL(config string) plugins.Plugin {
	return plugins.Plugin{
		Name:  "gf",
		Image: imageA,
		Manifest: plugins.Manifest{
			Capabilities:     []plugins.Capability{plugins.CapabilityView},
			EgressFromConfig: []string{"url"},
		},
		Config: []byte(config),
	}
}

func TestConfigHosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  string
		want    []plugins.ConfigHost
		wantErr []fieldError
	}{
		{
			name: "an https origin", config: `{"url":"https://ghostfolio.example.com"}`,
			want: []plugins.ConfigHost{{Field: "url", Host: host("ghostfolio.example.com", 443)}},
		},
		{
			name: "a port, a path and upper case", config: `{"url":"https://GF.Example.org:8443/app/"}`,
			want: []plugins.ConfigHost{{Field: "url", Host: host("gf.example.org", 8443)}},
		},
		{name: "field absent", config: `{"interval":"1h"}`},
		{name: "field null", config: `{"url":null}`},
		{name: "field empty", config: `{"url":""}`},
		{name: "no config", config: ``},
		{
			name: "plain http", config: `{"url":"http://gf.example.com"}`,
			wantErr: []fieldError{{"config.url", plugins.ErrConfigHost}},
		},
		{
			name: "an IPv4 address", config: `{"url":"https://192.0.2.7"}`,
			wantErr: []fieldError{{"config.url", plugins.ErrConfigHost}},
		},
		{
			name: "an IPv6 address", config: `{"url":"https://[2001:db8::7]:443"}`,
			wantErr: []fieldError{{"config.url", plugins.ErrConfigHost}},
		},
		{
			name: "a wildcard", config: `{"url":"https://*.example.com"}`,
			wantErr: []fieldError{{"config.url", plugins.ErrConfigHost}},
		},
		{
			name: "a single label", config: `{"url":"https://ghostfolio"}`,
			wantErr: []fieldError{{"config.url", plugins.ErrConfigHost}},
		},
		{
			name: "a user in the URL", config: `{"url":"https://someone@gf.example.com"}`,
			wantErr: []fieldError{{"config.url", plugins.ErrConfigHost}},
		},
		{
			name: "port zero", config: `{"url":"https://gf.example.com:0"}`,
			wantErr: []fieldError{{"config.url", plugins.ErrConfigHost}},
		},
		{
			name: "not a string", config: `{"url":42}`,
			wantErr: []fieldError{{"config.url", plugins.ErrConfigHost}},
		},
		{
			name: "config not an object", config: `["https://gf.example.com"]`,
			wantErr: []fieldError{{"config", plugins.ErrInvalidJSON}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := viewWithURL(tt.config)
			got, err := plugins.ConfigHosts(&p.Manifest, p.Config)
			if diff := cmp.Diff(tt.wantErr, fieldErrors(err), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ConfigHosts() errors (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ConfigHosts() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConfigHostPermissions(t *testing.T) {
	t.Parallel()
	gf := host("ghostfolio.example.com", 443)
	grant := plugins.Grant{Plugin: "gf", Permissions: plugins.Permissions{
		Capabilities: []plugins.Capability{plugins.CapabilityView},
		Egress:       []plugins.EgressHost{gf},
	}}
	p := viewWithURL(`{"url":"https://ghostfolio.example.com"}`)
	if diff := cmp.Diff([]plugins.EgressHost{gf}, p.Requested().Egress); diff != "" {
		t.Errorf("Requested().Egress (-want +got):\n%s", diff)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
	e, err := plugins.Effective(&p, &grant)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	w, err := plugins.Compile(e, topology())
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	// Through the egress proxy like any other host (D-75).
	if diff := cmp.Diff([]plugins.EgressHost{gf}, w.EgressAllowList.Hosts); diff != "" {
		t.Errorf("allow-list (-want +got):\n%s", diff)
	}
	if !hasRule(&w, plugins.PeerEgressProxy) {
		t.Errorf("no egress rule to the proxy: %+v", w.NetworkPolicy.Egress)
	}
	// The owner points the config elsewhere: that host waits for approval,
	// and nothing reaches it meanwhile.
	moved := viewWithURL(`{"url":"https://other.example.org"}`)
	pending, err := plugins.NeedsApproval(&grant, &moved)
	if err != nil {
		t.Fatalf("NeedsApproval: %v", err)
	}
	if diff := cmp.Diff(plugins.Permissions{Egress: []plugins.EgressHost{host("other.example.org", 443)}}, pending); diff != "" {
		t.Errorf("pending (-want +got):\n%s", diff)
	}
	bad := viewWithURL(`{"url":"https://192.0.2.7"}`)
	if diff := cmp.Diff([]fieldError{{"config.url", plugins.ErrConfigHost}}, fieldErrors(bad.Validate()), cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Validate() of an IP address (-want +got):\n%s", diff)
	}
}

func TestManifestEgressFromConfigFields(t *testing.T) {
	t.Parallel()
	m := viewWithURL("").Manifest
	m.EgressFromConfig = []string{"url", "9url", "", "url", "base_url"}
	want := []fieldError{
		{"egressFromConfig[1]", plugins.ErrInvalidName},
		{"egressFromConfig[2]", plugins.ErrInvalidName},
		{"egressFromConfig[3]", plugins.ErrDuplicate},
	}
	if diff := cmp.Diff(want, fieldErrors(m.Validate()), cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Validate() (-want +got):\n%s", diff)
	}
}

// hasRule reports whether w's NetworkPolicy has an egress rule to peer.
func hasRule(w *plugins.Workload, peer plugins.Peer) bool {
	for _, r := range w.NetworkPolicy.Egress {
		if r.Peer == peer {
			return true
		}
	}
	return false
}
