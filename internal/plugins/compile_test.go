package plugins_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// topology is a synthetic cluster: core and proxy at documentation-range
// ClusterIPs.
func topology() plugins.Topology {
	return plugins.Topology{
		Core:        plugins.Service{Address: "192.0.2.10", Port: 80, TargetPort: 8080},
		EgressProxy: plugins.Service{Address: "192.0.2.20", Port: 4750, TargetPort: 4750},
		PluginPort:  9090, CoreServiceAccount: "openfief-core",
	}
}

// coreAuthEnv is env plus the variables every plugin verifies the core's
// token with (D-59).
func coreAuthEnv(env ...plugins.EnvVar) []plugins.EnvVar {
	return append(env,
		plugins.EnvVar{Name: "OPENFIEF_CORE_SUBJECT", Value: "system:serviceaccount:openfief:openfief-core"},
		plugins.EnvVar{Name: "OPENFIEF_INSTANCE", Value: "sample"},
		plugins.EnvVar{Name: "OPENFIEF_SA_JWKS_DIR", Value: "/var/run/secrets/openfief-sa-jwks"},
	)
}

// identityParts are the parts of a Workload every plugin gets the same way.
func identityParts(w *plugins.Workload) {
	w.Namespace = "openfief"
	w.ServiceAccount = plugins.ServiceAccount{Name: "plugin-sample", Namespace: "openfief"}
	w.Token = plugins.ProjectedToken{
		Audience: "openfief-core", ExpirationSeconds: 3600, MountPath: "/var/run/secrets/openfief", Path: "token",
	}
	w.SigningKeys = plugins.ConfigMapVolume{ConfigMapName: "openfief-sa-jwks", MountPath: "/var/run/secrets/openfief-sa-jwks"}
	w.Env = coreAuthEnv(w.Env...)
}

func TestCompile(t *testing.T) {
	t.Parallel()
	image, err := plugins.ParseImageRef(imageA)
	if err != nil {
		t.Fatalf("ParseImageRef: %v", err)
	}
	withEgress := plugins.Workload{
		Plugin:       "sample",
		Image:        image,
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		DataTiers:    []plugins.DataTier{"T1"},
		Secrets: []plugins.SecretVolume{
			{SecretName: "sample-cert", MountPath: "/etc/openfief/cert", ReadOnly: true},
			{SecretName: "sample-token", MountPath: "/etc/openfief/token", ReadOnly: true},
		},
		Env: []plugins.EnvVar{
			{Name: "HTTPS_PROXY", Value: "http://192.0.2.20:4750"},
			{Name: "NO_PROXY", Value: "192.0.2.10"},
		},
		NetworkPolicy: plugins.NetworkPolicy{
			Ingress: []plugins.NetworkRule{{Peer: plugins.PeerCore, Port: 9090}},
			Egress: []plugins.NetworkRule{
				{Peer: plugins.PeerCore, Port: 8080},
				{Peer: plugins.PeerEgressProxy, Port: 4750},
			},
		},
		EgressAllowList: plugins.EgressAllowList{
			Plugin: "sample",
			Hosts:  []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)},
		},
		Config: []byte(`{"interval":"24h"}`),
		EgressToken: &plugins.ProjectedToken{
			Audience: "openfief-egress", ExpirationSeconds: 3600, MountPath: "/var/run/secrets/openfief-egress", Path: "token",
		},
	}
	identityParts(&withEgress)
	noEgress := plugins.Workload{
		Plugin:       "sample",
		Image:        image,
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		NetworkPolicy: plugins.NetworkPolicy{
			Ingress: []plugins.NetworkRule{{Peer: plugins.PeerCore, Port: 9090}},
			Egress:  []plugins.NetworkRule{{Peer: plugins.PeerCore, Port: 8080}},
		},
		EgressAllowList: plugins.EgressAllowList{Plugin: "sample"},
		Config:          []byte(`{"interval":"24h"}`),
	}
	identityParts(&noEgress)
	ipv6Proxy := withEgress
	ipv6Proxy.Env = coreAuthEnv(
		plugins.EnvVar{Name: "HTTPS_PROXY", Value: "http://[2001:db8::20]:4750"},
		plugins.EnvVar{Name: "NO_PROXY", Value: "2001:db8::10"},
	)

	tests := []struct {
		name     string
		grant    func(g *plugins.Grant)
		topology func(t *plugins.Topology)
		want     plugins.Workload
		wantErr  error
	}{
		{name: "connector with egress", grant: func(*plugins.Grant) {}, topology: func(*plugins.Topology) {}, want: withEgress},
		{
			name:     "no egress granted: no proxy route, no proxy env, no secrets",
			grant:    func(g *plugins.Grant) { g.Permissions = plugins.Permissions{Capabilities: g.Permissions.Capabilities} },
			topology: func(*plugins.Topology) {},
			want:     noEgress,
		},
		{
			name:  "IPv6 cluster addresses",
			grant: func(*plugins.Grant) {},
			topology: func(t *plugins.Topology) {
				t.Core.Address, t.EgressProxy.Address = "2001:db8::10", "2001:db8::20"
			},
			want: ipv6Proxy,
		},
		{
			name: "missing proxy address", grant: func(*plugins.Grant) {},
			topology: func(t *plugins.Topology) { t.EgressProxy.Address = "" }, wantErr: plugins.ErrInvalidTopology,
		},
		{
			name: "DNS-style URL instead of an address", grant: func(*plugins.Grant) {},
			topology: func(t *plugins.Topology) { t.Core.Address = "http://core" }, wantErr: plugins.ErrInvalidTopology,
		},
		{
			name: "missing target port", grant: func(*plugins.Grant) {},
			topology: func(t *plugins.Topology) { t.Core.TargetPort = 0 }, wantErr: plugins.ErrInvalidTopology,
		},
		{
			name: "missing core ServiceAccount", grant: func(*plugins.Grant) {},
			topology: func(t *plugins.Topology) { t.CoreServiceAccount = "" }, wantErr: plugins.ErrInvalidTopology,
		},
		{
			name: "missing plugin port", grant: func(*plugins.Grant) {},
			topology: func(t *plugins.Topology) { t.PluginPort = 0 }, wantErr: plugins.ErrInvalidTopology,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, g, topo := samplePlugin(), fullGrant(), topology()
			tt.grant(&g)
			tt.topology(&topo)
			e, effErr := plugins.Effective(&p, &g)
			if effErr != nil {
				t.Fatalf("Effective: %v", effErr)
			}
			got, compileErr := plugins.Compile(e, topo)
			if diff := cmp.Diff(tt.wantErr, compileErr, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Compile() error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Compile() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCompileNeverAllowsDNS(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		grant func(g *plugins.Grant)
	}{
		{name: "with egress", grant: func(*plugins.Grant) {}},
		{name: "without egress", grant: func(g *plugins.Grant) { g.Permissions.Egress = nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, g := samplePlugin(), fullGrant()
			tt.grant(&g)
			e, err := plugins.Effective(&p, &g)
			if err != nil {
				t.Fatalf("Effective: %v", err)
			}
			w, err := plugins.Compile(e, topology())
			if err != nil {
				t.Fatalf("Compile: %v", err)
			}
			checkOnlyCoreAndProxy(t, &w)
		})
	}
}

// checkOnlyCoreAndProxy fails t when w's egress reaches anything but the core
// and the egress proxy, allows DNS, or mounts a Kubernetes API token.
func checkOnlyCoreAndProxy(t *testing.T, w *plugins.Workload) {
	t.Helper()
	const dnsPort = 53
	for _, r := range w.NetworkPolicy.Egress {
		if r.Port == dnsPort || (r.Peer != plugins.PeerCore && r.Peer != plugins.PeerEgressProxy) {
			t.Errorf("egress rule %+v reaches beyond the core and the egress proxy", r)
		}
	}
	if w.ServiceAccount.AutomountAPIToken {
		t.Errorf("plugin ServiceAccount mounts a Kubernetes API token")
	}
}

func TestCompileDeterministic(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		reorder func(p *plugins.Plugin, g *plugins.Grant)
	}{
		{
			name: "reversed manifest and grant lists",
			reorder: func(p *plugins.Plugin, g *plugins.Grant) {
				m := &p.Manifest
				m.Egress[0], m.Egress[1] = m.Egress[1], m.Egress[0]
				m.Secrets[0], m.Secrets[1] = m.Secrets[1], m.Secrets[0]
				g.Permissions.Secrets[0], g.Permissions.Secrets[1] = g.Permissions.Secrets[1], g.Permissions.Secrets[0]
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p1, g1 := samplePlugin(), fullGrant()
			p2, g2 := samplePlugin(), fullGrant()
			tt.reorder(&p2, &g2)
			e1, err1 := plugins.Effective(&p1, &g1)
			e2, err2 := plugins.Effective(&p2, &g2)
			if err1 != nil || err2 != nil {
				t.Fatalf("Effective: %v, %v", err1, err2)
			}
			w1, err1 := plugins.Compile(e1, topology())
			w2, err2 := plugins.Compile(e2, topology())
			if err1 != nil || err2 != nil {
				t.Fatalf("Compile: %v, %v", err1, err2)
			}
			if diff := cmp.Diff(w1, w2); diff != "" {
				t.Errorf("Compile() depends on input order (-first +reordered):\n%s", diff)
			}
		})
	}
}

func TestCompileNil(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		e    *plugins.EffectivePermissions
	}{
		{name: "nil", e: nil},
		{name: "zero value", e: &plugins.EffectivePermissions{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := plugins.Compile(tt.e, topology())
			if diff := cmp.Diff(plugins.ErrNoCapabilityGranted, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Compile() error (-want +got):\n%s", diff)
			}
		})
	}
}
