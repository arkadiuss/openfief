package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// effectiveView is EffectivePermissions read through its accessors.
type effectiveView struct {
	Plugin       string
	Image        string
	Capabilities []plugins.Capability
	Egress       []plugins.EgressHost
	Secrets      []plugins.SecretMount
	DataTiers    []plugins.DataTier
	Config       string
}

func viewOf(e *plugins.EffectivePermissions) effectiveView {
	if e == nil {
		return effectiveView{}
	}
	return effectiveView{
		Plugin: e.Plugin(), Image: e.Image().String(), Capabilities: e.Capabilities(), Egress: e.Egress(),
		Secrets: e.Secrets(), DataTiers: e.DataTiers(), Config: string(e.Config()),
	}
}

func TestEffective(t *testing.T) {
	t.Parallel()
	full := effectiveView{
		Plugin:       "sample",
		Image:        imageA,
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Egress:       []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)},
		Secrets: []plugins.SecretMount{
			{Role: "cert", MountPath: "/etc/openfief/cert"},
			{Role: "token", MountPath: "/etc/openfief/token"},
		},
		DataTiers: []plugins.DataTier{"T1"},
		Config:    `{"interval":"24h"}`,
	}
	tests := []struct {
		name    string
		grant   func(g *plugins.Grant)
		plugin  func(p *plugins.Plugin)
		want    effectiveView
		wantErr error
	}{
		{name: "everything granted", grant: func(*plugins.Grant) {}, plugin: func(*plugins.Plugin) {}, want: full},
		{
			name: "grant approves more than requested",
			grant: func(g *plugins.Grant) {
				g.Permissions.Capabilities = append(g.Permissions.Capabilities, plugins.CapabilityView)
				g.Permissions.Egress = append(g.Permissions.Egress, host("extra.example.net", 443))
				g.Permissions.Secrets = append(g.Permissions.Secrets, "unrelated")
				g.Permissions.DataTiers = append(g.Permissions.DataTiers, "T2")
			},
			plugin: func(*plugins.Plugin) {}, want: full,
		},
		{
			name: "partial grant enforces only the overlap",
			grant: func(g *plugins.Grant) {
				g.Permissions.Egress = []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 443)}
				g.Permissions.Secrets = []string{"sample-token"}
				g.Permissions.DataTiers = nil
			},
			plugin: func(*plugins.Plugin) {},
			want: effectiveView{
				Plugin: "sample", Image: imageA,
				Capabilities: []plugins.Capability{plugins.CapabilityConnector},
				Egress:       []plugins.EgressHost{host("api.example.com", 443)},
				Secrets:      []plugins.SecretMount{{Role: "token", MountPath: "/etc/openfief/token"}},
				Config:       `{"interval":"24h"}`,
			},
		},
		{
			name:  "upgrade requesting more keeps only the granted part",
			grant: func(*plugins.Grant) {},
			plugin: func(p *plugins.Plugin) {
				p.Manifest.Egress = append(p.Manifest.Egress, host("new.example.net", 443))
			},
			want: full,
		},
		{
			name:   "no egress granted",
			grant:  func(g *plugins.Grant) { g.Permissions.Egress = nil },
			plugin: func(*plugins.Plugin) {},
			want: effectiveView{
				Plugin: "sample", Image: imageA, Capabilities: full.Capabilities, Secrets: full.Secrets,
				DataTiers: full.DataTiers, Config: full.Config,
			},
		},
		{
			name:    "no capability granted",
			grant:   func(g *plugins.Grant) { g.Permissions.Capabilities = []plugins.Capability{plugins.CapabilityView} },
			plugin:  func(*plugins.Plugin) {},
			wantErr: plugins.ErrNoCapabilityGranted,
		},
		{
			name: "grant for another plugin", grant: func(g *plugins.Grant) { g.Plugin = "other" },
			plugin: func(*plugins.Plugin) {}, wantErr: plugins.ErrGrantMismatch,
		},
		{
			name: "unpinned image", grant: func(*plugins.Grant) {},
			plugin: func(p *plugins.Plugin) { p.Image = "registry.example.com/sample:1.0.0" }, wantErr: plugins.ErrImageNotPinned,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g, p := fullGrant(), samplePlugin()
			tt.grant(&g)
			tt.plugin(&p)
			got, err := plugins.Effective(&p, &g)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Effective() error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, viewOf(got)); diff != "" {
				t.Errorf("Effective() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEffectiveEqual(t *testing.T) {
	t.Parallel()
	narrow := fullGrant()
	narrow.Permissions.Egress = nil
	tests := []struct {
		name  string
		other plugins.Grant
		want  bool
	}{
		{name: "same inputs", other: fullGrant(), want: true},
		{name: "different egress", other: narrow, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, g := samplePlugin(), fullGrant()
			a, errA := plugins.Effective(&p, &g)
			b, errB := plugins.Effective(&p, &tt.other)
			if errA != nil || errB != nil {
				t.Fatalf("Effective: %v, %v", errA, errB)
			}
			if got := a.Equal(b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
			if a.Equal(nil) {
				t.Errorf("Equal(nil) = true for a non-nil value")
			}
		})
	}
}

func TestEffectiveAccessorsCopy(t *testing.T) {
	t.Parallel()
	p, g := samplePlugin(), fullGrant()
	e, err := plugins.Effective(&p, &g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	p.Config[0] = 'X'
	e.Egress()[0] = host("evil.example.net", 443)
	e.Config()[0] = 'X'
	want := []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)}
	if diff := cmp.Diff(want, e.Egress()); diff != "" {
		t.Errorf("Egress() changed through a returned slice (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(`{"interval":"24h"}`, string(e.Config())); diff != "" {
		t.Errorf("Config() changed through the input or a returned slice (-want +got):\n%s", diff)
	}
}
