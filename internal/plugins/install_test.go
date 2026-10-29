package plugins_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// imageOther is a third synthetic image, in a repository whose last path
// component is not a valid instance name.
const imageOther = "registry.example.com/openfief/sample_plugin@" + digestB

// digestC is a third synthetic digest.
const digestC = "sha256:c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3c3"

// installCatalog has one entry for "sample", pinning imageA, and two
// versions of "twice".
func installCatalog(t *testing.T) plugins.Catalog {
	t.Helper()
	pub := plugins.Signer{Name: "example", Kind: plugins.SignerPublisher, Identity: keyless(issuerExample, subjectCI)}
	entry := func(name, image string) plugins.CatalogEntry {
		img, err := plugins.ParseImageRef(image)
		if err != nil {
			t.Fatalf("image: %v", err)
		}
		return plugins.CatalogEntry{Name: name, Image: img, Publisher: pub}
	}
	return plugins.Catalog{Entries: []plugins.CatalogEntry{
		entry("sample", imageA),
		entry("twice", "registry.example.com/openfief/twice@"+digestC),
		entry("twice", "registry.example.com/openfief/twice@"+digestC[:len(digestC)-1]+"4"),
	}}
}

// installOutcome is what a resolved source says, for comparison.
type installOutcome struct {
	Image string
	Entry string
	Name  string
}

func TestResolveInstallSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		arg     string
		want    installOutcome
		wantErr error
	}{
		{name: "catalog name", arg: "sample", want: installOutcome{Image: imageA, Entry: "sample", Name: "sample"}},
		{
			name: "image the catalog pins is verified by its entry",
			arg:  "mirror.example.org/copy/renamed@" + digestA,
			want: installOutcome{Image: "mirror.example.org/copy/renamed@" + digestA, Entry: "sample", Name: "sample"},
		},
		{name: "image the catalog does not pin", arg: imageB, want: installOutcome{Image: imageB, Name: "sample-connector"}},
		{name: "repository name is no instance name", arg: imageOther, want: installOutcome{Image: imageOther}},
		{name: "tag refused", arg: "registry.example.com/openfief/sample:1.0", wantErr: plugins.ErrImageNotPinned},
		{name: "unknown name", arg: "missing", wantErr: plugins.ErrNotInCatalog},
		{name: "name of several versions", arg: "twice", wantErr: plugins.ErrAmbiguousCatalogName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := installCatalog(t)
			src, err := plugins.ResolveInstallSource(&c, tt.arg)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			got := installOutcome{Image: src.Image.String(), Name: plugins.DefaultInstanceName(&src)}
			if src.Entry != nil {
				got.Entry = src.Entry.Name
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("source (-want +got):\n%s", diff)
			}
		})
	}
}

// signedSample is samplePlugin's image, signed, with its manifest label.
func signedSample(t *testing.T) plugins.SignedImage {
	t.Helper()
	p := samplePlugin()
	img, err := plugins.ParseImageRef(p.Image)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	return plugins.SignedImage{Image: img, Manifest: p.Manifest}
}

func TestPlanInstall(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		instance string
		config   []byte
	}{
		{name: "named instance with config", instance: "sample-main", config: []byte(`{"interval":"24h"}`)},
		{name: "no config", instance: "sample"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			signed := signedSample(t)
			p, g, err := plugins.PlanInstall(tt.instance, &signed, tt.config)
			if err != nil {
				t.Fatalf("PlanInstall: %v", err)
			}
			want := samplePlugin()
			want.Name, want.Config = tt.instance, tt.config
			if diff := cmp.Diff(want, p); diff != "" {
				t.Errorf("plugin (-want +got):\n%s", diff)
			}
			wantGrant := plugins.Grant{Plugin: tt.instance, Permissions: want.Manifest.Requested()}
			if diff := cmp.Diff(wantGrant, g); diff != "" {
				t.Errorf("grant (-want +got):\n%s", diff)
			}
			if pending, pendErr := plugins.NeedsApproval(&g, &p); pendErr != nil || !pending.IsEmpty() {
				t.Errorf("the grant leaves %+v pending (error %v); it must approve everything requested", pending, pendErr)
			}
		})
	}
}

func TestPlanInstallRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		instance   string
		config     string
		wantFields []string
	}{
		{name: "name too long for the CRD", instance: "a23456789012345678901234567890123456789012345678901234567", wantFields: []string{"name"}},
		{name: "invalid name", instance: "Sample", wantFields: []string{"name", "plugin"}},
		{name: "config not JSON", instance: "sample", config: `{`, wantFields: []string{"config"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			signed := signedSample(t)
			_, _, err := plugins.PlanInstall(tt.instance, &signed, []byte(tt.config))
			var got []string
			for _, fe := range fieldErrors(errors.Unwrap(err)) {
				got = append(got, fe.Field)
			}
			if diff := cmp.Diff(tt.wantFields, got, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("invalid fields (-want +got):\n%s\nerror: %v", diff, err)
			}
		})
	}
}

func TestSecretsToDelete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		instance []string
		others   []string
		wantDel  []string
		wantKept []string
	}{
		{name: "nothing shared", instance: []string{"b-token", "a-cert"}, wantDel: []string{"a-cert", "b-token"}},
		{name: "named by manifest and grant", instance: []string{"a-cert", "a-cert"}, wantDel: []string{"a-cert"}},
		{
			name: "another instance of the image still names one", instance: []string{"a-cert", "shared"},
			others: []string{"shared", "unrelated"}, wantDel: []string{"a-cert"}, wantKept: []string{"shared"},
		},
		{name: "no secrets"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := plugins.SecretsToDelete(tt.instance, tt.others)
			if diff := cmp.Diff(tt.wantDel, got.Delete); diff != "" {
				t.Errorf("deleted (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantKept, got.Keep); diff != "" {
				t.Errorf("kept (-want +got):\n%s", diff)
			}
		})
	}
}
