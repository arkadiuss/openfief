package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestPluginValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		change func(p *plugins.Plugin)
		want   []fieldError
	}{
		{name: "valid", change: func(*plugins.Plugin) {}},
		{name: "no config", change: func(p *plugins.Plugin) { p.Config = nil }},
		{
			name:   "tagged image",
			change: func(p *plugins.Plugin) { p.Image = "registry.example.com/openfief/sample-connector:latest" },
			want:   []fieldError{{"image", plugins.ErrImageNotPinned}},
		},
		{
			name:   "bad name",
			change: func(p *plugins.Plugin) { p.Name = "Sample" },
			want:   []fieldError{{"name", plugins.ErrInvalidName}},
		},
		{
			name:   "name starts with a digit",
			change: func(p *plugins.Plugin) { p.Name = "1sample" },
			want:   []fieldError{{"name", plugins.ErrInvalidName}},
		},
		{
			name: "manifest errors keep their field",
			change: func(p *plugins.Plugin) {
				p.Manifest.Capabilities = nil
				p.Manifest.Egress = []plugins.EgressHost{host("*.example.com", 443)}
			},
			want: []fieldError{
				{"manifest.capabilities", plugins.ErrNoCapability},
				{"manifest.egress[0]", plugins.ErrWildcardHost},
			},
		},
		{
			name:   "config not JSON",
			change: func(p *plugins.Plugin) { p.Config = []byte("interval: 24h") },
			want:   []fieldError{{"config", plugins.ErrInvalidJSON}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := samplePlugin()
			tt.change(&p)
			if diff := cmp.Diff(tt.want, fieldErrors(p.Validate()), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGrantValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		change func(g *plugins.Grant)
		want   []fieldError
	}{
		{name: "valid", change: func(*plugins.Grant) {}},
		{name: "grants nothing", change: func(g *plugins.Grant) { g.Permissions = plugins.Permissions{} }},
		{
			name: "bad entries",
			change: func(g *plugins.Grant) {
				g.Plugin = ""
				g.Permissions.Capabilities = []plugins.Capability{"admin"}
				g.Permissions.Egress = []plugins.EgressHost{host("10.0.0.1", 443)}
				g.Permissions.Secrets = []string{"sample-cert", "sample-cert"}
				g.Permissions.DataTiers = []plugins.DataTier{"all tiers"}
			},
			want: []fieldError{
				{"plugin", plugins.ErrInvalidName},
				{"permissions.capabilities[0]", plugins.ErrUnknownCapability},
				{"permissions.egress[0]", plugins.ErrIPLiteral},
				{"permissions.secrets[1]", plugins.ErrDuplicate},
				{"permissions.dataTiers[0]", plugins.ErrInvalidDataTier},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := fullGrant()
			tt.change(&g)
			if diff := cmp.Diff(tt.want, fieldErrors(g.Validate()), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNeedsApproval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		grant   func(g *plugins.Grant)
		upgrade func(p *plugins.Plugin)
		want    plugins.Permissions
		wantErr error
	}{
		{name: "same request", grant: func(*plugins.Grant) {}, upgrade: func(*plugins.Plugin) {}},
		{
			name: "new image, same request", grant: func(*plugins.Grant) {},
			upgrade: func(p *plugins.Plugin) { p.Image = imageB },
		},
		{
			name: "smaller request", grant: func(*plugins.Grant) {},
			upgrade: func(p *plugins.Plugin) { p.Manifest.Egress, p.Manifest.Secrets = nil, nil },
		},
		{
			name: "secret moved to a new path is not new", grant: func(*plugins.Grant) {},
			upgrade: func(p *plugins.Plugin) { p.Manifest.Secrets[0].MountPath = "/etc/openfief/other" },
		},
		{
			name:  "new egress host and port",
			grant: func(*plugins.Grant) {},
			upgrade: func(p *plugins.Plugin) {
				p.Image = imageB
				p.Manifest.Egress = append(p.Manifest.Egress, host("new.example.net", 443), host("api.example.com", 8443))
			},
			want: plugins.Permissions{Egress: []plugins.EgressHost{host("api.example.com", 8443), host("new.example.net", 443)}},
		},
		{
			name:  "everything new",
			grant: func(g *plugins.Grant) { g.Permissions = plugins.Permissions{} },
			upgrade: func(p *plugins.Plugin) {
				p.Manifest.Capabilities = append(p.Manifest.Capabilities, plugins.CapabilityView)
			},
			want: plugins.Permissions{
				Capabilities: []plugins.Capability{plugins.CapabilityConnector, plugins.CapabilityView},
				Egress:       []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)},
				Secrets:      []string{"sample-cert", "sample-token"},
				DataTiers:    []plugins.DataTier{"T1"},
			},
		},
		{
			name:  "new secret and data tier",
			grant: func(*plugins.Grant) {},
			upgrade: func(p *plugins.Plugin) {
				p.Manifest.Secrets = append(p.Manifest.Secrets, plugins.SecretMount{Role: "key", MountPath: "/etc/openfief/key"})
				p.Manifest.DataTiers = append(p.Manifest.DataTiers, "T2")
			},
			want: plugins.Permissions{Secrets: []string{"sample-key"}, DataTiers: []plugins.DataTier{"T2"}},
		},
		{
			name: "grant for another plugin", grant: func(g *plugins.Grant) { g.Plugin = "other" },
			upgrade: func(*plugins.Plugin) {}, wantErr: plugins.ErrGrantMismatch,
		},
		{
			name: "invalid upgrade", grant: func(*plugins.Grant) {},
			upgrade: func(p *plugins.Plugin) { p.Image = "registry.example.com/sample:2.0.0" }, wantErr: plugins.ErrImageNotPinned,
		},
		{
			name: "invalid grant", grant: func(g *plugins.Grant) { g.Permissions.Secrets = []string{"A"} },
			upgrade: func(*plugins.Plugin) {}, wantErr: plugins.ErrInvalidName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g, p := fullGrant(), samplePlugin()
			tt.grant(&g)
			tt.upgrade(&p)
			got, err := plugins.NeedsApproval(&g, &p)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("NeedsApproval() error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NeedsApproval() (-want +got):\n%s", diff)
			}
			if got.IsEmpty() != cmp.Equal(tt.want, plugins.Permissions{}) {
				t.Errorf("IsEmpty() = %v for %+v", got.IsEmpty(), got)
			}
		})
	}
}

func TestNeedsApprovalNil(t *testing.T) {
	t.Parallel()
	g, p := fullGrant(), samplePlugin()
	tests := []struct {
		name  string
		grant *plugins.Grant
		plug  *plugins.Plugin
	}{
		{name: "nil grant", plug: &p},
		{name: "nil plugin", grant: &g},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := plugins.NeedsApproval(tt.grant, tt.plug)
			if diff := cmp.Diff(plugins.ErrMissing, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("NeedsApproval() error (-want +got):\n%s", diff)
			}
		})
	}
}
