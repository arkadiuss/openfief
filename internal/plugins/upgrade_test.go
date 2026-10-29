package plugins_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Synthetic versions of one catalog plugin; the digests are not of any
// real image.
const (
	digestD = "sha256:c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3"
	imageD  = "registry.example.com/openfief/sample-connector@" + digestD
	mirrorA = "mirror.example.org/sample-connector@" + digestA
)

// versionedCatalog lists three reviewed versions of sample and one of
// another plugin of the same version number.
func versionedCatalog(t *testing.T) plugins.Catalog {
	t.Helper()
	entry := func(name, version, image string) plugins.CatalogEntry {
		img, err := plugins.ParseImageRef(image)
		if err != nil {
			t.Fatalf("image: %v", err)
		}
		return plugins.CatalogEntry{Name: name, Version: version, Image: img, Publisher: plugins.Signer{
			Name: "example", Kind: plugins.SignerPublisher, Identity: keyless(issuerExample, subjectCI),
		}}
	}
	return plugins.Catalog{Entries: []plugins.CatalogEntry{
		entry("sample", "2026.11.0", imageA),
		entry("sample", "2026.11.1", imageB),
		entry("sample", "2026.12.0", imageD),
		entry("other", "2026.11.1", "registry.example.com/openfief/other@"+digestD),
	}}
}

func TestResolveUpgradeSource(t *testing.T) {
	t.Parallel()
	c := versionedCatalog(t)
	tests := []struct {
		name, current, arg string
		want               string
		wantErr            error
	}{
		{name: "a catalog version of the same plugin", current: imageA, arg: "2026.11.1", want: imageB},
		{name: "found by repository when the digest is unpinned", current: "registry.example.com/openfief/sample-connector@" + "sha256:" + strings.Repeat("d", 64), arg: "2026.12.0", want: imageD},
		{name: "from a mirror of a pinned digest", current: mirrorA, arg: "2026.12.0", want: imageD},
		{name: "an image reference", current: imageA, arg: imageD, want: imageD},
		{name: "a version the plugin doesn't have", current: imageA, arg: "2027.1.0", wantErr: plugins.ErrNoUpgradeVersion},
		{name: "a tag", current: imageA, arg: "registry.example.com/openfief/sample-connector:latest", wantErr: plugins.ErrImageNotPinned},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			current, err := plugins.ParseImageRef(tt.current)
			if err != nil {
				t.Fatalf("current: %v", err)
			}
			got, err := plugins.ResolveUpgradeSource(&c, current, tt.arg)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if err == nil && got.Image.String() != tt.want {
				t.Errorf("image = %s, want %s", got.Image, tt.want)
			}
		})
	}
}

func TestResolveInstallSourceListsVersions(t *testing.T) {
	t.Parallel()
	c := versionedCatalog(t)
	_, err := plugins.ResolveInstallSource(&c, "sample")
	if !errors.Is(err, plugins.ErrAmbiguousCatalogName) {
		t.Fatalf("error = %v, want ErrAmbiguousCatalogName", err)
	}
	for _, v := range []string{"sample 2026.11.0 (", "sample 2026.11.1 (", "sample 2026.12.0 ("} {
		if !strings.Contains(err.Error(), v) {
			t.Errorf("error %q doesn't list %q", err, v)
		}
	}
}

func TestPlanUpgrade(t *testing.T) {
	t.Parallel()
	current := samplePlugin()
	current.Name = "main"
	img, err := plugins.ParseImageRef(imageB)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	next := samplePlugin().Manifest
	next.Egress = append(next.Egress[:1:1], host("new.example.net", 443))
	signed := plugins.SignedImage{Image: img, Manifest: next}
	p, g, err := plugins.PlanUpgrade(&current, &signed, nil)
	if err != nil {
		t.Fatalf("PlanUpgrade: %v", err)
	}
	if p.Name != "main" || p.Image != imageB || !bytes.Equal(p.Config, current.Config) {
		t.Errorf("plugin %s %s %s, want main on the new image with the current config", p.Name, p.Image, p.Config)
	}
	if diff := cmp.Diff(p.Requested(), g.Permissions); diff != "" {
		t.Errorf("grant (-want +got):\n%s", diff)
	}
	from, to := current.Requested(), p.Requested()
	want := plugins.PermissionsDiff{
		Added:   plugins.Permissions{Egress: []plugins.EgressHost{host("new.example.net", 443)}},
		Removed: plugins.Permissions{Egress: []plugins.EgressHost{host("data.example.org", 8443)}},
	}
	if diff := cmp.Diff(want, plugins.DiffPermissions(&from, &to)); diff != "" {
		t.Errorf("diff (-want +got):\n%s", diff)
	}
	same := plugins.DiffPermissions(&from, &from)
	if !same.IsEmpty() {
		t.Errorf("diff of a version with itself = %+v, want empty", same)
	}
	if _, _, err := plugins.PlanUpgrade(&current, &signed, []byte(`{`)); err == nil {
		t.Errorf("PlanUpgrade with a config that isn't JSON succeeded")
	}
}

func TestPlanUpgradeKeepsClusterService(t *testing.T) {
	t.Parallel()
	gs := ghostfolioService()
	current := inCluster(&gs)
	img, err := plugins.ParseImageRef(imageB)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	kept, _, err := plugins.PlanUpgrade(&current, &plugins.SignedImage{Image: img, Manifest: current.Manifest}, nil)
	if err != nil || kept.ClusterService == nil || *kept.ClusterService != gs {
		t.Errorf("upgrade still asking for it: %v, %v; want the service kept", kept.ClusterService, err)
	}
	dropped, g, err := plugins.PlanUpgrade(&current, &plugins.SignedImage{Image: img, Manifest: plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityView},
	}}, nil)
	if err != nil || dropped.ClusterService != nil || len(g.Permissions.ClusterServices) != 0 {
		t.Errorf("upgrade no longer asking: %v, %v, %v; want none", dropped.ClusterService, g.Permissions.ClusterServices, err)
	}
}
