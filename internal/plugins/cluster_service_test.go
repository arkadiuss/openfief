package plugins_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// ghostfolioService is a synthetic in-cluster Ghostfolio.
func ghostfolioService() plugins.ClusterService {
	return plugins.ClusterService{Namespace: "ghostfolio", Service: "ghostfolio", Port: 3333}
}

// inCluster is a view plugin whose manifest asks for an in-cluster service
// (D-74) and which names cs.
func inCluster(cs *plugins.ClusterService) plugins.Plugin {
	return plugins.Plugin{
		Name:  "gf",
		Image: imageA,
		Manifest: plugins.Manifest{
			Capabilities:   []plugins.Capability{plugins.CapabilityView},
			ClusterService: &plugins.ClusterServiceRequest{Description: "the owner's Ghostfolio"},
		},
		ClusterService: cs,
	}
}

func TestParseClusterService(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    plugins.ClusterService
		wantErr error
	}{
		{in: "ghostfolio/ghostfolio:3333", want: ghostfolioService()},
		{in: "openfief/openfief-core:8081", wantErr: plugins.ErrInvalidClusterService},
		{in: "ghostfolio/ghostfolio", wantErr: plugins.ErrInvalidClusterService},
		{in: "ghostfolio:3333", wantErr: plugins.ErrInvalidClusterService},
		{in: "Ghostfolio/ghostfolio:3333", wantErr: plugins.ErrInvalidClusterService},
		{in: "ghostfolio/ghostfolio.svc:3333", wantErr: plugins.ErrInvalidClusterService},
		{in: "ghostfolio/ghostfolio:0", wantErr: plugins.ErrInvalidClusterService},
		{in: "ghostfolio/ghostfolio:70000", wantErr: plugins.ErrInvalidClusterService},
		{in: "ghostfolio/ghostfolio:http", wantErr: plugins.ErrInvalidClusterService},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.ParseClusterService(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseClusterService (-want +got):\n%s", diff)
			}
			if err == nil && got.String() != tt.in {
				t.Errorf("String() = %q, want %q", got.String(), tt.in)
			}
		})
	}
}

func TestPluginClusterServiceValidate(t *testing.T) {
	t.Parallel()
	gs := ghostfolioService()
	openfief := plugins.ClusterService{Namespace: "openfief", Service: "openfief-core", Port: 8081}
	unrequested := inCluster(&gs)
	unrequested.Manifest.ClusterService = nil
	twoLines := inCluster(nil)
	twoLines.Manifest.ClusterService.Description = "line one\nline two"
	tests := []struct {
		name   string
		plugin plugins.Plugin
		want   []fieldError
	}{
		{name: "picked at install", plugin: inCluster(&gs)},
		{name: "none picked", plugin: inCluster(nil)},
		{
			name: "in the openfief namespace", plugin: inCluster(&openfief),
			want: []fieldError{{"clusterService", plugins.ErrInvalidClusterService}},
		},
		{
			name: "not requested by the manifest", plugin: unrequested,
			want: []fieldError{{"clusterService", plugins.ErrClusterServiceNotRequested}},
		},
		{
			name: "description on two lines", plugin: twoLines,
			want: []fieldError{{"manifest.clusterService.description", plugins.ErrInvalidName}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, fieldErrors(tt.plugin.Validate()), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClusterServicePermissions(t *testing.T) {
	t.Parallel()
	gs := ghostfolioService()
	p := inCluster(&gs)
	if diff := cmp.Diff([]plugins.ClusterService{gs}, p.Requested().ClusterServices); diff != "" {
		t.Errorf("Requested().ClusterServices (-want +got):\n%s", diff)
	}
	other := plugins.ClusterService{Namespace: "ghostfolio", Service: "ghostfolio", Port: 80}
	tests := []struct {
		name        string
		granted     []plugins.ClusterService
		wantPending []plugins.ClusterService
		wantRun     bool
	}{
		{name: "approved exactly", granted: []plugins.ClusterService{gs}, wantRun: true},
		{name: "not approved", wantPending: []plugins.ClusterService{gs}},
		{name: "another port approved", granted: []plugins.ClusterService{other}, wantPending: []plugins.ClusterService{gs}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := plugins.Grant{Plugin: "gf", Permissions: plugins.Permissions{
				Capabilities: []plugins.Capability{plugins.CapabilityView}, ClusterServices: tt.granted,
			}}
			pending, err := plugins.NeedsApproval(&g, &p)
			if err != nil {
				t.Fatalf("NeedsApproval: %v", err)
			}
			if diff := cmp.Diff(tt.wantPending, pending.ClusterServices); diff != "" {
				t.Errorf("pending (-want +got):\n%s", diff)
			}
			checkEffectiveClusterService(t, &p, &g, tt.wantRun)
		})
	}
}

// checkEffectiveClusterService fails t unless the effective permissions of
// p under g hold its in-cluster service exactly when want.
func checkEffectiveClusterService(t *testing.T, p *plugins.Plugin, g *plugins.Grant, want bool) {
	t.Helper()
	e, err := plugins.Effective(p, g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	cs, ok := e.ClusterService()
	if ok != want || (ok && cs != *p.ClusterService) {
		t.Errorf("ClusterService() = %v, %v; want it only when approved", cs, ok)
	}
}

func TestCompileClusterService(t *testing.T) {
	t.Parallel()
	gs := ghostfolioService()
	p := inCluster(&gs)
	p.Manifest.Egress = []plugins.EgressHost{host("api.example.com", 443)}
	g := plugins.Grant{Plugin: "gf", Permissions: plugins.Permissions{
		Capabilities:    []plugins.Capability{plugins.CapabilityView},
		Egress:          []plugins.EgressHost{host("api.example.com", 443)},
		ClusterServices: []plugins.ClusterService{gs},
	}}
	e, err := plugins.Effective(&p, &g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	if _, compileErr := plugins.Compile(e, topology()); !errors.Is(compileErr, plugins.ErrInvalidTopology) {
		t.Errorf("Compile without the service's address: error = %v, want ErrInvalidTopology", compileErr)
	}
	topo := topology()
	topo.ClusterService = &plugins.Service{Address: "192.0.2.30", Port: 3333, TargetPort: 3000}
	w, err := plugins.Compile(e, topo)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	wantRules := []plugins.NetworkRule{
		{Peer: plugins.PeerCore, Port: 8080},
		{Peer: plugins.PeerEgressProxy, Port: 4750},
		{Peer: plugins.PeerClusterService, Port: 3000},
	}
	if diff := cmp.Diff(wantRules, w.NetworkPolicy.Egress); diff != "" {
		t.Errorf("egress rules (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(&gs, w.ClusterService); diff != "" {
		t.Errorf("ClusterService (-want +got):\n%s", diff)
	}
	checkClusterServiceEnv(t, &w)
}

// checkClusterServiceEnv fails t unless w hands the plugin the service's
// ClusterIP, never a DNS name, and keeps it off the proxy (D-74).
func checkClusterServiceEnv(t *testing.T, w *plugins.Workload) {
	t.Helper()
	env := map[string]string{}
	for _, v := range w.Env {
		env[v.Name] = v.Value
	}
	if env["OPENFIEF_CLUSTER_SERVICE_ADDRESS"] != "192.0.2.30:3333" || env["NO_PROXY"] != "192.0.2.10,192.0.2.30" {
		t.Errorf("env = %v, want the service's ClusterIP and port, kept off the proxy", env)
	}
	for _, h := range w.EgressAllowList.Hosts {
		if h.Host == "192.0.2.30" {
			t.Errorf("the in-cluster service is on the egress allow-list")
		}
	}
}
