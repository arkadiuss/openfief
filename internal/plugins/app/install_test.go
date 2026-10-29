package app_test

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// Synthetic images and signer subjects; none is real.
const (
	pinnedDigest   = "sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
	unpinnedDigest = "sha256:0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f"
	pinnedImage    = "registry.example.com/example/sample-connector@" + pinnedDigest
	unpinnedImage  = "registry.example.com/example/other-connector@" + unpinnedDigest
	unsignedImage  = "registry.example.com/example/unsigned@" + unpinnedDigest
	strangerImage  = "registry.example.com/example/stranger@" + unpinnedDigest
	brokenImage    = "registry.example.com/example/unreachable@" + unpinnedDigest
	badNameImage   = "registry.example.com/example/bad_name@" + unpinnedDigest
	publisherSubj  = "https://github.com/example/sample-connector/.github/workflows/release.yml@refs/heads/main"
	strangerSubj   = "https://github.com/stranger/plugin/.github/workflows/release.yml@refs/heads/main"
)

var (
	// errRegistryDown stands in for an unreachable registry.
	errRegistryDown = errors.New("registry unreachable")
	// errSchema stands in for a config schema refusing a config.
	errSchema = errors.New("schema says no")
)

// sampleManifest is what every fake image's signed label requests.
func sampleManifest() plugins.Manifest {
	return plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Egress:       []plugins.EgressHost{{Host: "api.example.com", Port: 443}},
		Secrets:      []plugins.SecretMount{{Role: "credentials", MountPath: "/etc/openfief/sample"}},
		ConfigSchema: []byte(`{"type":"object"}`),
	}
}

// fakeImages answers Inspect by image repository: signed by the publisher,
// by OpenFief's CI, by a stranger, unsigned, or a registry failure.
type fakeImages struct{}

func (fakeImages) Inspect(_ context.Context, image plugins.ImageRef) (plugins.SignedImage, error) {
	out := plugins.SignedImage{Image: image, Manifest: sampleManifest()}
	switch image.Repository() {
	case "registry.example.com/example/sample-connector":
		out.Signatures = []plugins.SignerIdentity{{Issuer: plugins.GitHubActionsIssuer, Subject: publisherSubj}}
	case "registry.example.com/example/stranger":
		out.Signatures = []plugins.SignerIdentity{{Issuer: plugins.GitHubActionsIssuer, Subject: strangerSubj}}
	case "registry.example.com/example/unsigned":
		return plugins.SignedImage{Image: image}, nil
	case "registry.example.com/example/unreachable":
		return plugins.SignedImage{}, errRegistryDown
	default:
		out.Signatures = []plugins.SignerIdentity{{Issuer: plugins.GitHubActionsIssuer, Subject: plugins.OpenFiefCISubject}}
	}
	return out, nil
}

// installer trusts OpenFief's CI and one catalog publisher, which pins
// pinnedImage as "sample"; its schema check refuses the config `{"bad":1}`.
func installer(t *testing.T) *app.Installer {
	t.Helper()
	img, err := plugins.ParseImageRef(pinnedImage)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	c := plugins.Catalog{Entries: []plugins.CatalogEntry{{Name: "sample", Image: img, Publisher: plugins.Signer{
		Name: "example", Kind: plugins.SignerPublisher,
		Identity: plugins.SignerIdentity{Issuer: plugins.GitHubActionsIssuer, Subject: publisherSubj},
	}}}}
	policy, err := plugins.NewTrustPolicy(c, nil)
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	return &app.Installer{
		Images: fakeImages{}, Policy: &policy, Catalog: &c,
		ValidateConfig: func(_, config []byte) error {
			if string(config) == `{"bad":1}` {
				return errSchema
			}
			return nil
		},
	}
}

// proposed is the part of a Proposal a test compares.
type proposed struct {
	Name, Image, Signer string
	Trust               plugins.Trust
	Entry               bool
	Grant               plugins.Permissions
}

func TestInstallerPrepare(t *testing.T) {
	t.Parallel()
	grant := func(name string) plugins.Permissions {
		p := plugins.Plugin{Name: name, Manifest: sampleManifest()}
		return p.Requested()
	}
	tests := []struct {
		name    string
		req     app.InstallRequest
		want    proposed
		wantErr error
	}{
		{
			name: "verified catalog plugin",
			req:  app.InstallRequest{Arg: "sample"},
			want: proposed{Name: "sample", Image: pinnedImage, Signer: "example", Trust: plugins.Verified, Entry: true, Grant: grant("sample")},
		},
		{
			name: "image signed by OpenFief's CI, named after its repository",
			req:  app.InstallRequest{Arg: unpinnedImage},
			want: proposed{Name: "other-connector", Image: unpinnedImage, Signer: "openfief-ci", Trust: plugins.Unverified, Grant: grant("other-connector")},
		},
		{
			name: "owner names the instance",
			req:  app.InstallRequest{Arg: "sample", Name: "sample-second", Config: []byte(`{"ok":true}`)},
			want: proposed{Name: "sample-second", Image: pinnedImage, Signer: "example", Trust: plugins.Verified, Entry: true, Grant: grant("sample-second")},
		},
		{name: "unsigned image", req: app.InstallRequest{Arg: unsignedImage}, wantErr: plugins.ErrUnsigned},
		{name: "unsigned image is not trusted", req: app.InstallRequest{Arg: unsignedImage}, wantErr: app.ErrNotTrusted},
		{name: "untrusted signer", req: app.InstallRequest{Arg: strangerImage}, wantErr: plugins.ErrUntrustedSigner},
		{name: "registry unreachable", req: app.InstallRequest{Arg: brokenImage}, wantErr: errRegistryDown},
		{name: "config the schema refuses", req: app.InstallRequest{Arg: "sample", Config: []byte(`{"bad":1}`)}, wantErr: errSchema},
		{name: "no name to derive", req: app.InstallRequest{Arg: badNameImage}, wantErr: app.ErrNoInstanceName},
		{name: "tag", req: app.InstallRequest{Arg: "registry.example.com/example/sample:1.0"}, wantErr: plugins.ErrImageNotPinned},
		{name: "not in the catalog", req: app.InstallRequest{Arg: "missing"}, wantErr: plugins.ErrNotInCatalog},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := installer(t).Prepare(context.Background(), tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			got := proposed{
				Name: p.Plugin.Name, Image: p.Plugin.Image, Signer: p.Decision.Signer.Name, Trust: p.Decision.Trust,
				Entry: p.Entry != nil, Grant: p.Grant.Permissions,
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("proposal (-want +got):\n%s", diff)
			}
			if requested := sampleManifest(); !p.Plugin.Manifest.Equal(&requested) || !bytes.Equal(p.Plugin.Config, tt.req.Config) {
				t.Errorf("plugin requests %+v with config %s, want the signed manifest and the given config", p.Plugin.Manifest, p.Plugin.Config)
			}
		})
	}
}

// fakeStore is an in-memory InstanceStore.
type fakeStore struct {
	instances []app.Instance
	deleted   map[string][]string
}

func (s *fakeStore) List(context.Context) ([]app.Instance, error) {
	return slices.Clone(s.instances), nil
}

func (s *fakeStore) Create(_ context.Context, p *plugins.Plugin, g *plugins.Grant) error {
	if slices.ContainsFunc(s.instances, func(i app.Instance) bool { return i.Name == p.Name }) {
		return app.ErrInstanceExists
	}
	s.instances = append(s.instances, app.Instance{
		Name: p.Name, Image: p.Image, HasPlugin: true, HasGrant: true, Secrets: g.Permissions.Secrets,
	})
	return nil
}

func (s *fakeStore) Delete(_ context.Context, name string, secrets []string) error {
	if s.deleted == nil {
		s.deleted = map[string][]string{}
	}
	s.deleted[name] = secrets
	s.instances = slices.DeleteFunc(s.instances, func(i app.Instance) bool { return i.Name == name })
	return nil
}

// installSamples installs the catalog's sample twice, as sample-a and
// sample-b, next to an unrelated instance, and returns the store.
func installSamples(t *testing.T) *fakeStore {
	t.Helper()
	ctx := context.Background()
	in := installer(t)
	store := &fakeStore{instances: []app.Instance{
		{Name: "unrelated", HasPlugin: true, HasGrant: true, Secrets: []string{"other-credentials"}},
	}}
	for _, name := range []string{"sample-b", "sample-a"} {
		p, err := in.Prepare(ctx, app.InstallRequest{Arg: "sample", Name: name})
		if err != nil {
			t.Fatalf("prepare %s: %v", name, err)
		}
		if err := app.Install(ctx, store, &p); err != nil {
			t.Fatalf("install %s: %v", name, err)
		}
	}
	return store
}

func TestInstall(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := installSamples(t)
	again, err := installer(t).Prepare(ctx, app.InstallRequest{Arg: "sample", Name: "sample-a"})
	if err != nil {
		t.Fatalf("prepare again: %v", err)
	}
	if installErr := app.Install(ctx, store, &again); !errors.Is(installErr, app.ErrInstanceExists) {
		t.Errorf("installing over sample-a: error = %v, want ErrInstanceExists", installErr)
	}
	// A grant naming sample-b's Secret, as "sample" with a role
	// "b-credentials" would spell it, is refused.
	clash, err := installer(t).Prepare(ctx, app.InstallRequest{Arg: "sample", Name: "sample"})
	if err != nil {
		t.Fatalf("prepare clash: %v", err)
	}
	clash.Grant.Permissions.Secrets = []string{"sample-b-credentials"}
	if installErr := app.Install(ctx, store, &clash); !errors.Is(installErr, app.ErrSecretInUse) {
		t.Errorf("installing with sample-b's Secret: error = %v, want ErrSecretInUse", installErr)
	}
	listed, err := app.Instances(ctx, store)
	if err != nil {
		t.Fatalf("instances: %v", err)
	}
	names := make([]string, 0, len(listed))
	for _, i := range listed {
		names = append(names, i.Name)
	}
	if diff := cmp.Diff([]string{"sample-a", "sample-b", "unrelated"}, names); diff != "" {
		t.Errorf("listed (-want +got):\n%s", diff)
	}
}

// uninstall plans and carries out the uninstall of name, and returns the
// plan's Secrets.
func uninstall(t *testing.T, store *fakeStore, name string) plugins.UninstallSecrets {
	t.Helper()
	plan, err := app.PlanUninstall(context.Background(), store, name)
	if err != nil {
		t.Fatalf("plan uninstall %s: %v", name, err)
	}
	if err := app.Uninstall(context.Background(), store, &plan); err != nil {
		t.Fatalf("uninstall %s: %v", name, err)
	}
	return plan.Secrets
}

func TestUninstall(t *testing.T) {
	t.Parallel()
	store := installSamples(t)
	// Each instance of one image mounts its own Secret (D-89), so each
	// uninstall deletes exactly its own.
	if diff := cmp.Diff(plugins.UninstallSecrets{Delete: []string{"sample-a-credentials"}}, uninstall(t, store, "sample-a")); diff != "" {
		t.Errorf("first uninstall's Secrets (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(plugins.UninstallSecrets{Delete: []string{"sample-b-credentials"}}, uninstall(t, store, "sample-b")); diff != "" {
		t.Errorf("second uninstall's Secrets (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(map[string][]string{"sample-a": {"sample-a-credentials"}, "sample-b": {"sample-b-credentials"}}, store.deleted); diff != "" {
		t.Errorf("deleted Secrets (-want +got):\n%s", diff)
	}
	if _, err := app.PlanUninstall(context.Background(), store, "sample-a"); !errors.Is(err, app.ErrNotInstalled) {
		t.Errorf("uninstalling sample-a twice: error = %v, want ErrNotInstalled", err)
	}
}
