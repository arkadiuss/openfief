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
		{
			name: "repository name holds the Secret name separator (D-89)", arg: "registry.example.com/someone/sample--plugin@" + digestB,
			want: installOutcome{Image: "registry.example.com/someone/sample--plugin@" + digestB},
		},
		{
			name: "repository name is reserved for the platform", arg: "registry.example.com/someone/openfief@" + digestB,
			want: installOutcome{Image: "registry.example.com/someone/openfief@" + digestB},
		},
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
			p, g, err := plugins.PlanInstall(tt.instance, &signed, tt.config, nil)
			if err != nil {
				t.Fatalf("PlanInstall: %v", err)
			}
			want := samplePlugin()
			want.Name, want.Config = tt.instance, tt.config
			if diff := cmp.Diff(want, p); diff != "" {
				t.Errorf("plugin (-want +got):\n%s", diff)
			}
			wantGrant := plugins.Grant{Plugin: tt.instance, Permissions: want.Requested()}
			if diff := cmp.Diff(wantGrant, g); diff != "" {
				t.Errorf("grant (-want +got):\n%s", diff)
			}
			if pending, pendErr := plugins.NeedsApproval(&g, &p); pendErr != nil || !pending.IsEmpty() {
				t.Errorf("the grant leaves %+v pending (error %v); it must approve everything requested", pending, pendErr)
			}
		})
	}
}

func TestPlanInstallOwnerPicks(t *testing.T) {
	t.Parallel()
	img, err := plugins.ParseImageRef(imageA)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	signed := plugins.SignedImage{Image: img, Manifest: plugins.Manifest{
		Capabilities:     []plugins.Capability{plugins.CapabilityView},
		EgressFromConfig: []string{"url"},
		ClusterService:   &plugins.ClusterServiceRequest{Description: "the owner's Ghostfolio"},
	}}
	gs := ghostfolioService()
	p, g, err := plugins.PlanInstall("gf", &signed, []byte(`{"url":"https://gf.example.com"}`), &gs)
	if err != nil {
		t.Fatalf("PlanInstall: %v", err)
	}
	want := plugins.Permissions{
		Capabilities:    []plugins.Capability{plugins.CapabilityView},
		Egress:          []plugins.EgressHost{host("gf.example.com", 443)},
		ClusterServices: []plugins.ClusterService{gs},
	}
	if diff := cmp.Diff(want, g.Permissions); diff != "" {
		t.Errorf("grant (-want +got):\n%s", diff)
	}
	if p.ClusterService == nil || *p.ClusterService != gs {
		t.Errorf("plugin's in-cluster service = %v, want %v", p.ClusterService, gs)
	}
	if _, _, err := plugins.PlanInstall("gf", &signed, []byte(`{"url":"https://192.0.2.7"}`), nil); !errors.Is(err, plugins.ErrConfigHost) {
		t.Errorf("an IP address in url: error = %v, want ErrConfigHost", err)
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
		{name: "name reserved for the platform", instance: "openfief", wantFields: []string{"name", "plugin"}},
		{name: "name with the platform's prefix", instance: "openfief-db", wantFields: []string{"name", "plugin"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			signed := signedSample(t)
			_, _, err := plugins.PlanInstall(tt.instance, &signed, []byte(tt.config), nil)
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
	owned := func(names ...string) map[string]string {
		out := map[string]string{}
		for _, n := range names {
			out[n] = "a"
		}
		return out
	}
	tests := []struct {
		name         string
		named        []string
		others       []string
		owners       map[string]string
		wantDel      []string
		wantKept     []string
		wantNotOwned []string
	}{
		{name: "nothing shared", named: []string{"a--token", "a--cert"}, owners: owned("a--token", "a--cert"), wantDel: []string{"a--cert", "a--token"}},
		{name: "named by manifest and grant", named: []string{"a--cert", "a--cert"}, owners: owned("a--cert"), wantDel: []string{"a--cert"}},
		{
			name: "another instance of the image still names one", named: []string{"a--cert", "shared"},
			others: []string{"shared", "unrelated"}, owners: owned("a--cert", "shared"), wantDel: []string{"a--cert"}, wantKept: []string{"shared"},
		},
		{name: "a named Secret that doesn't exist", named: []string{"a--cert", "a--gone"}, owners: owned("a--cert"), wantDel: []string{"a--cert"}},
		{
			name: "unlabelled and another instance's Secrets are left alone", named: []string{"a--cert", "a--key", "openfief-db-app"},
			owners: map[string]string{"a--cert": "a", "a--key": "", "openfief-db-app": ""}, wantDel: []string{"a--cert"},
			wantNotOwned: []string{"a--key", "openfief-db-app"},
		},
		{name: "labelled for another instance", named: []string{"a--c"}, owners: map[string]string{"a--c": "b"}, wantNotOwned: []string{"a--c"}},
		{name: "no secrets"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := plugins.SecretsToDelete("a", tt.named, tt.others, tt.owners)
			want := plugins.UninstallSecrets{Delete: tt.wantDel, Keep: tt.wantKept, NotOwned: tt.wantNotOwned}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("uninstall secrets (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReservedName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want bool
	}{
		{name: "openfief", want: true},
		{name: "openfief-db", want: true},
		{name: "openfief-db-app", want: true},
		{name: "openfiefs"},
		{name: "my-openfief"},
		{name: "ibkr-main"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := plugins.ReservedName(tt.name); got != tt.want {
				t.Errorf("ReservedName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
