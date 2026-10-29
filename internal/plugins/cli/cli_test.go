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
)

// errNoCluster stands in for an unreachable cluster.
var errNoCluster = errors.New("no cluster")

// fakeCluster is an in-memory Cluster; signers are the owner's.
type fakeCluster struct {
	instances []app.Instance
	signers   []plugins.Signer
	deleted   []string
}

func (c *fakeCluster) List(context.Context) ([]app.Instance, error) {
	return slices.Clone(c.instances), nil
}

func (c *fakeCluster) Create(_ context.Context, p *plugins.Plugin, g *plugins.Grant) error {
	c.instances = append(c.instances, app.Instance{Name: p.Name, Image: p.Image, HasPlugin: true, HasGrant: true, Secrets: g.Permissions.Secrets})
	return nil
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
	if image.String() == ownerImage {
		id = plugins.SignerIdentity{Issuer: plugins.GitHubActionsIssuer, Subject: ownerSubject}
	}
	return plugins.SignedImage{Image: image, Signatures: []plugins.SignerIdentity{id}, Manifest: plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityView},
		Secrets:      []plugins.SecretMount{{Name: "view-key", MountPath: "/etc/openfief/view"}},
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
	files := map[string]string{"config.json": `{"theme":"dark"}`, "signers.json": ownerSigners}
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
			wantOut: []string{"Instance:     sample", "unverified: signed by trusted signer openfief-ci", "view-key at /etc/openfief/view", `Config given: {"theme":"dark"}`, "Install sample? [y/N] ", "Installed sample."},
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
		{name: "unknown command", args: []string{"plugin", "upgrade", "x"}, wantErr: cli.ErrUsage},
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
