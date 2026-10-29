package cli_test

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// Synthetic owner-signed image and owner signer.
const (
	ownerImage   = "registry.example.com/owner/mine@sha256:0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f"
	ownerSubject = "https://github.com/owner/mine/.github/workflows/release.yml@refs/heads/main"
	ownerSigners = `{"signers":[{"name":"my-ci","issuer":"https://token.actions.githubusercontent.com","subject":"` + ownerSubject + `"}]}`
	// gfImage is signed by OpenFief's CI; its manifest asks for the host of
	// config field url (D-75) and an in-cluster service (D-74).
	gfImage = "registry.example.com/example/gf@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
)

// errNoCluster stands in for an unreachable cluster.
var errNoCluster = errors.New("no cluster")

// fakeCluster is an in-memory Cluster; signers are the owner's.
type fakeCluster struct {
	instances []app.Instance
	signers   []plugins.Signer
	deleted   []string
	// granted is what each Create's grant approves.
	granted []plugins.Permissions
}

func (c *fakeCluster) List(context.Context) ([]app.Instance, error) {
	return slices.Clone(c.instances), nil
}

func (c *fakeCluster) Create(_ context.Context, p *plugins.Plugin, g *plugins.Grant) error {
	c.instances = append(c.instances, app.Instance{
		Name: p.Name, Image: p.Image, HasPlugin: true, HasGrant: true, Secrets: g.Permissions.Secrets,
		ClusterService: p.ClusterService,
	})
	c.granted = append(c.granted, g.Permissions)
	return nil
}

// ClusterService finds any service but one named absent; namespace guarded
// has NetworkPolicies of its own.
func (c *fakeCluster) ClusterService(_ context.Context, cs plugins.ClusterService) (app.ClusterServiceInfo, error) {
	if cs.Service == "absent" {
		return app.ClusterServiceInfo{}, app.ErrClusterService
	}
	return app.ClusterServiceInfo{
		TargetPort: "3000", Selector: map[string]string{"app": cs.Service}, NamespaceHasNetworkPolicies: cs.Namespace == "guarded",
	}, nil
}

func (c *fakeCluster) Delete(_ context.Context, name string, secrets []string) error {
	c.deleted = append(append(c.deleted, name), secrets...)
	c.instances = slices.DeleteFunc(c.instances, func(i app.Instance) bool { return i.Name == name })
	return nil
}

func (c *fakeCluster) OwnerSigners(context.Context) ([]plugins.Signer, error) { return c.signers, nil }

// fakeImages signs sampleImage by OpenFief's CI and ownerImage by the
// owner's CI.
type fakeImages struct{}

func (fakeImages) Inspect(_ context.Context, image plugins.ImageRef) (plugins.SignedImage, error) {
	id := plugins.OpenFiefCI().Identity
	if image.String() == gfImage {
		return plugins.SignedImage{Image: image, Signatures: []plugins.SignerIdentity{id}, Manifest: plugins.Manifest{
			Capabilities:     []plugins.Capability{plugins.CapabilityView},
			EgressFromConfig: []string{"url"},
			ClusterService:   &plugins.ClusterServiceRequest{Description: "the owner's Ghostfolio"},
		}}, nil
	}
	if image.String() == ownerImage {
		id = plugins.SignerIdentity{Issuer: plugins.GitHubActionsIssuer, Subject: ownerSubject}
	}
	return plugins.SignedImage{Image: image, Signatures: []plugins.SignerIdentity{id}, Manifest: plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityView},
		Secrets:      []plugins.SecretMount{{Role: "key", MountPath: "/etc/openfief/view"}},
	}}, nil
}

// runCase is one command line run.
type runCase struct {
	name string
	args []string
	// in is what the owner types.
	in string
	// cluster is the cluster's state; nil means it can't be reached.
	cluster *fakeCluster
	wantErr error
	// wantOut must appear in Out, wantErrOut in Err.
	wantOut, wantErrOut []string
	// wantNames are the instances afterwards, wantDeleted what was deleted.
	wantNames, wantDeleted []string
}

// env is an Env over the fakes and c, reading files from a fixed map.
func env(c *fakeCluster, in string, out, errOut *bytes.Buffer) *cli.Env {
	files := map[string]string{
		"config.json": `{"theme":"dark"}`, "signers.json": ownerSigners, "gf.json": `{"url":"https://gf.example.com"}`,
	}
	return &cli.Env{
		In: strings.NewReader(in), Out: out, Err: errOut,
		ReadFile: func(path string) ([]byte, error) {
			if s, ok := files[path]; ok {
				return []byte(s), nil
			}
			return nil, fs.ErrNotExist
		},
		Cluster: func(string, string) (cli.Cluster, error) {
			if c == nil {
				return nil, errNoCluster
			}
			return c, nil
		},
		Images: func(*plugins.TrustPolicy) (app.ImageInspector, error) { return fakeImages{}, nil },
		Render: func(p *plugins.Plugin, _ *plugins.Grant) ([]byte, error) {
			return []byte("kind: Plugin\nname: " + p.Name + "\n"), nil
		},
		IngressRule: func(instance string, cs plugins.ClusterService, info *app.ClusterServiceInfo) ([]byte, error) {
			return []byte("kind: NetworkPolicy\nnamespace: " + cs.Namespace + "\nfrom: " + instance + "\nport: " + info.TargetPort + "\n"), nil
		},
	}
}

func installed(names ...string) *fakeCluster {
	c := &fakeCluster{}
	for _, n := range names {
		c.instances = append(c.instances, app.Instance{Name: n, HasPlugin: true, HasGrant: true, Secrets: []string{"view-key"}})
	}
	return c
}

func TestRun(t *testing.T) {
	t.Parallel()
	owner := []plugins.Signer{{Name: "my-ci", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{Issuer: plugins.GitHubActionsIssuer, Subject: ownerSubject}}}
	tests := []runCase{
		{
			name: "install confirmed", args: []string{"plugin", "install", sampleImage, "--config", "config.json"}, in: "y\n",
			cluster: installed(), wantNames: []string{"sample"},
			wantOut: []string{"Instance:     sample", "unverified: signed by trusted signer openfief-ci", "sample-key (role key) at /etc/openfief/view", `Config given: {"theme":"dark"}`, "Install sample? [y/N] ", "Installed sample."},
		},
		{
			name: "install declined changes nothing", args: []string{"plugin", "install", sampleImage}, in: "n\n",
			cluster: installed(), wantErr: cli.ErrAborted, wantOut: []string{"Install sample? [y/N] "},
		},
		{
			name: "no answer is no", args: []string{"plugin", "install", sampleImage},
			cluster: installed(), wantErr: cli.ErrAborted,
		},
		{
			name: "--yes and --name, flags before the argument", args: []string{"plugin", "install", "--yes", "--name", "second", sampleImage},
			cluster: installed("sample"), wantNames: []string{"sample", "second"}, wantOut: []string{"Installed second."},
		},
		{
			name: "owner's signer from the cluster", args: []string{"plugin", "install", ownerImage, "--yes"},
			cluster: &fakeCluster{signers: owner}, wantNames: []string{"mine"}, wantOut: []string{"signed by trusted signer my-ci (Owner"},
		},
		{
			name: "owner's signer unknown to the cluster", args: []string{"plugin", "install", ownerImage, "--yes"},
			cluster: installed(), wantErr: plugins.ErrUntrustedSigner,
		},
		{
			name: "--print needs no cluster", args: []string{"plugin", "install", sampleImage, "--print"},
			wantOut: []string{"kind: Plugin\nname: sample\n"}, wantErrOut: []string{"Trusting OpenFief's CI and the catalog only", "Instance:     sample"},
		},
		{
			name: "--print with the owner's signers file", args: []string{"plugin", "install", ownerImage, "--print", "--trusted-signers", "signers.json"},
			wantOut: []string{"name: mine\n"}, wantErrOut: []string{"signed by trusted signer my-ci"},
		},
		{name: "cluster unreachable", args: []string{"plugin", "install", sampleImage}, wantErr: errNoCluster},
		{name: "missing config file", args: []string{"plugin", "install", sampleImage, "--config", "nope.json"}, cluster: installed(), wantErr: fs.ErrNotExist},
		{
			name: "uninstall confirmed", args: []string{"plugin", "uninstall", "sample"}, in: "yes\n",
			cluster: installed("sample", "other"), wantNames: []string{"other"}, wantDeleted: []string{"sample"},
			wantOut: []string{"Keeps Secrets another instance uses: view-key", "Uninstall sample? [y/N] ", "Uninstalled sample."},
		},
		{
			name: "uninstall the last user of a Secret", args: []string{"plugin", "uninstall", "sample", "--yes"},
			cluster: installed("sample"), wantDeleted: []string{"sample", "view-key"}, wantOut: []string{"Secret view-key"},
		},
		{name: "uninstall declined", args: []string{"plugin", "uninstall", "sample"}, in: "\n", cluster: installed("sample"), wantNames: []string{"sample"}, wantErr: cli.ErrAborted},
		{name: "uninstall unknown", args: []string{"plugin", "uninstall", "nope", "--yes"}, cluster: installed(), wantErr: app.ErrNotInstalled},
		{name: "list", args: []string{"plugin", "list"}, cluster: installed("b", "a"), wantNames: []string{"b", "a"}, wantOut: []string{"NAME", "a  ", "b  "}},
		{name: "no command", args: nil, wantErr: cli.ErrUsage},
		{name: "unknown command", args: []string{"plugin", "frobnicate", "x"}, wantErr: cli.ErrUsage},
		{name: "two arguments", args: []string{"plugin", "install", "a", "b"}, wantErr: cli.ErrUsage},
		{name: "unknown flag", args: []string{"plugin", "list", "--yes"}, wantErr: cli.ErrUsage},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkRun(t, tt)
		})
	}
}

func TestRunInstallOwnerPicks(t *testing.T) {
	t.Parallel()
	ghostfolio := plugins.ClusterService{Namespace: "ghostfolio", Service: "gf", Port: 3333}
	tests := []struct {
		runCase

		wantGrant plugins.Permissions
	}{
		{
			runCase: runCase{
				name: "asks for the service, shows it and the config host, warns", args: []string{"plugin", "install", gfImage, "--config", "gf.json"},
				in: "ghostfolio\ngf\n3333\ny\n", cluster: installed(), wantNames: []string{"gf"},
				wantOut: []string{
					"  Namespace (empty for none): ", "  Service: ", "  Port: ",
					"Egress hosts: gf.example.com:443 (the host of config field url)",
					"In-cluster:   namespace ghostfolio, service gf, port 3333 (the owner's Ghostfolio)",
					"not through the egress proxy", "WARNING:      traffic to that service may be plain HTTP",
				},
			},
			wantGrant: plugins.Permissions{
				Capabilities: []plugins.Capability{plugins.CapabilityView}, Egress: []plugins.EgressHost{{Host: "gf.example.com", Port: 443}},
				ClusterServices: []plugins.ClusterService{ghostfolio},
			},
		},
		{
			runCase: runCase{
				name: "flag, and a namespace with its own NetworkPolicies", args: []string{"plugin", "install", gfImage, "--cluster-service", "guarded/gf:3333", "--yes"},
				cluster: installed(), wantNames: []string{"gf"},
				wantOut: []string{"Namespace guarded has NetworkPolicies of its own", "kind: NetworkPolicy\nnamespace: guarded\nfrom: gf\nport: 3000\n"},
			},
			wantGrant: plugins.Permissions{
				Capabilities:    []plugins.Capability{plugins.CapabilityView},
				ClusterServices: []plugins.ClusterService{{Namespace: "guarded", Service: "gf", Port: 3333}},
			},
		},
		{
			runCase: runCase{
				name: "no namespace picks none", args: []string{"plugin", "install", gfImage, "--config", "gf.json"},
				in: "\ny\n", cluster: installed(), wantNames: []string{"gf"},
				wantOut: []string{"In-cluster:   none picked (it asks for the owner's Ghostfolio)"},
			},
			wantGrant: plugins.Permissions{
				Capabilities: []plugins.Capability{plugins.CapabilityView}, Egress: []plugins.EgressHost{{Host: "gf.example.com", Port: 443}},
			},
		},
		{runCase: runCase{
			name: "a service that isn't there", args: []string{"plugin", "install", gfImage, "--cluster-service", "ghostfolio/absent:3333", "--yes"},
			cluster: installed(), wantErr: app.ErrClusterService,
		}},
		{runCase: runCase{
			name: "the openfief namespace", args: []string{"plugin", "install", gfImage, "--cluster-service", "openfief/openfief-core:8081", "--yes"},
			cluster: installed(), wantErr: plugins.ErrInvalidClusterService,
		}},
		{runCase: runCase{
			name: "a plugin that asks for none", args: []string{"plugin", "install", sampleImage, "--cluster-service", "ghostfolio/ghostfolio:3333", "--yes"},
			cluster: installed(), wantErr: plugins.ErrClusterServiceNotRequested,
		}},
		{runCase: runCase{
			name: "--print says what an ingress rule must admit", args: []string{"plugin", "install", gfImage, "--cluster-service", "ghostfolio/ghostfolio:3333", "--print"},
			wantOut: []string{"name: gf\n"}, wantErrOut: []string{"If namespace ghostfolio has NetworkPolicies of its own", "openfief.io/plugin=gf"},
		}},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkRun(t, &tt.runCase)
			if tt.cluster == nil || tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff([]plugins.Permissions{tt.wantGrant}, tt.cluster.granted); diff != "" {
				t.Errorf("grant (-want +got):\n%s", diff)
			}
		})
	}
}

// checkRun runs one case and checks its error, output and cluster.
func checkRun(t *testing.T, tt *runCase) {
	t.Helper()
	var out, errOut bytes.Buffer
	err := cli.Run(context.Background(), tt.args, env(tt.cluster, tt.in, &out, &errOut))
	if !errors.Is(err, tt.wantErr) {
		t.Fatalf("error = %v, want %v\nout:\n%s", err, tt.wantErr, out.String())
	}
	for _, s := range tt.wantOut {
		if !strings.Contains(out.String(), s) {
			t.Errorf("out lacks %q:\n%s", s, out.String())
		}
	}
	for _, s := range tt.wantErrOut {
		if !strings.Contains(errOut.String(), s) {
			t.Errorf("err lacks %q:\n%s", s, errOut.String())
		}
	}
	if tt.cluster == nil {
		return
	}
	var names []string
	for _, i := range tt.cluster.instances {
		names = append(names, i.Name)
	}
	if diff := cmp.Diff(tt.wantNames, names); diff != "" {
		t.Errorf("instances (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(tt.wantDeleted, tt.cluster.deleted); diff != "" {
		t.Errorf("deleted (-want +got):\n%s", diff)
	}
}
