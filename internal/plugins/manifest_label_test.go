package plugins_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// sampleLabel is samplePlugin's manifest as an image label, with its lists in
// another order.
const sampleLabel = `{"capabilities":["connector"],` +
	`"egress":["data.example.org:8443","api.example.com"],` +
	`"secrets":[{"role":"cert","mountPath":"/etc/openfief/cert"},{"role":"token","mountPath":"/etc/openfief/token"}],` +
	`"dataTiers":["T1"],"configSchema":{"type":"object"}}`

func TestParseManifestLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    plugins.Manifest
		wantErr error
	}{
		{
			name: "smallest manifest", in: `{"capabilities":["view"]}`,
			want: plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityView}},
		},
		{
			name: "null config schema is none", in: `{"capabilities":["view"],"configSchema":null}`,
			want: plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityView}},
		},
		{
			name: "every field", in: sampleLabel,
			want: plugins.Manifest{
				Capabilities: []plugins.Capability{plugins.CapabilityConnector},
				Egress:       []plugins.EgressHost{host("data.example.org", 8443), host("api.example.com", 443)},
				Secrets: []plugins.SecretMount{
					{Role: "cert", MountPath: "/etc/openfief/cert"},
					{Role: "token", MountPath: "/etc/openfief/token"},
				},
				DataTiers:    []plugins.DataTier{"T1"},
				ConfigSchema: []byte(`{"type":"object"}`),
			},
		},
		{
			name: "config host and in-cluster service",
			in:   `{"capabilities":["view"],"egressFromConfig":["url"],"clusterService":{"description":"the owner's Ghostfolio"}}`,
			want: plugins.Manifest{
				Capabilities:     []plugins.Capability{plugins.CapabilityView},
				EgressFromConfig: []string{"url"},
				ClusterService:   &plugins.ClusterServiceRequest{Description: "the owner's Ghostfolio"},
			},
		},
		{
			name: "a connector declaring cross-check prices and statement files (D-112, D-109)",
			in:   `{"capabilities":["connector"],"crossCheckPrices":true,"statementFiles":true}`,
			want: plugins.Manifest{
				Capabilities: []plugins.Capability{plugins.CapabilityConnector}, CrossCheckPrices: true, StatementFiles: true,
			},
		},
		{name: "bad config field", in: `{"capabilities":["view"],"egressFromConfig":["a.b"]}`, wantErr: plugins.ErrInvalidName},
		{name: "no capability", in: `{}`, wantErr: plugins.ErrNoCapability},
		{name: "unknown capability", in: `{"capabilities":["trader"]}`, wantErr: plugins.ErrUnknownCapability},
		{name: "IP egress", in: `{"capabilities":["connector"],"egress":["10.0.0.1"]}`, wantErr: plugins.ErrIPLiteral},
		{name: "reserved mount path", in: `{"capabilities":["connector"],"secrets":[{"role":"s","mountPath":"/var/run/secrets/x"}]}`, wantErr: plugins.ErrInvalidMountPath},
		{name: "unknown field", in: `{"capabilities":["view"],"privileged":true}`, wantErr: plugins.ErrInvalidJSON},
		{name: "not JSON", in: `capabilities: [view]`, wantErr: plugins.ErrInvalidJSON},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.ParseManifestLabel(tt.in)
			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseManifestLabel error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseManifestLabel: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ParseManifestLabel (-want +got):\n%s", diff)
			}
		})
	}
}

func TestManifestEqual(t *testing.T) {
	t.Parallel()
	label, err := plugins.ParseManifestLabel(sampleLabel)
	if err != nil {
		t.Fatalf("label: %v", err)
	}
	tests := []struct {
		name   string
		change func(m *plugins.Manifest)
		want   bool
	}{
		{name: "same, in another order", change: func(*plugins.Manifest) {}, want: true},
		{
			name:   "same schema, other spacing and key order",
			change: func(m *plugins.Manifest) { m.ConfigSchema = []byte(` { "type" : "object" } `) },
			want:   true,
		},
		{name: "extra egress host", change: func(m *plugins.Manifest) { m.Egress = append(m.Egress, host("exfil.example.net", 443)) }},
		{name: "other port", change: func(m *plugins.Manifest) { m.Egress[0] = host("api.example.com", 8443) }},
		{name: "extra capability", change: func(m *plugins.Manifest) { m.Capabilities = append(m.Capabilities, plugins.CapabilityView) }},
		{name: "other mount path", change: func(m *plugins.Manifest) { m.Secrets[0].MountPath = "/etc/elsewhere" }},
		{name: "extra data tier", change: func(m *plugins.Manifest) { m.DataTiers = append(m.DataTiers, "T2") }},
		{name: "other schema", change: func(m *plugins.Manifest) { m.ConfigSchema = []byte(`{"type":"string"}`) }},
		{name: "no schema", change: func(m *plugins.Manifest) { m.ConfigSchema = nil }},
		{name: "invalid schema", change: func(m *plugins.Manifest) { m.ConfigSchema = []byte(`{`) }},
		{name: "extra config host field", change: func(m *plugins.Manifest) { m.EgressFromConfig = []string{"url"} }},
		{name: "extra in-cluster service", change: func(m *plugins.Manifest) { m.ClusterService = &plugins.ClusterServiceRequest{} }},
		{name: "declares cross-check prices", change: func(m *plugins.Manifest) { m.CrossCheckPrices = true }},
		{name: "declares statement files", change: func(m *plugins.Manifest) { m.StatementFiles = true }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			declared := samplePlugin().Manifest
			tt.change(&declared)
			if got := declared.Equal(&label); got != tt.want {
				t.Errorf("Equal = %v, want %v", got, tt.want)
			}
		})
	}
}
