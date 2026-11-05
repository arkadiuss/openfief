package plugins_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Synthetic signer identities; none is a real publisher's.
const (
	issuerExample   = "https://issuer.example.com"
	subjectCI       = "https://github.com/example/sample-plugin/.github/workflows/release.yml@refs/heads/main"
	subjectOtherRef = "https://github.com/example/sample-plugin/.github/workflows/release.yml@refs/heads/feature"
)

// newPublicKeyPEM returns a fresh ECDSA P-256 public key in PEM form.
func newPublicKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

// The owner's own CI in the policy fixture: the repository and tag its
// signing run comes from, to which its identity is pinned (D-187).
const (
	ownerRepository = "https://github.com/owner-example/plugin"
	ownerRef        = "refs/tags/v1.2.0"
)

func keyless(issuer, subject string) plugins.SignerIdentity {
	return plugins.SignerIdentity{Issuer: issuer, Subject: subject}
}

// fromSource is a keyless identity whose signing run came from repository
// at ref, as Fulcio's source repository extensions certify it.
func fromSource(issuer, subject, repository, ref string) plugins.SignerIdentity {
	return plugins.SignerIdentity{Issuer: issuer, Subject: subject, Source: &plugins.SignerSource{Repository: repository, Ref: ref}}
}

// openFiefCISigned is a signature by OpenFief's CI: its subject from its own
// repository's main branch.
func openFiefCISigned() plugins.SignerIdentity {
	return fromSource(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject, plugins.OpenFiefCISourceRepository, plugins.OpenFiefCISourceRef)
}

func TestSignerIdentityValidate(t *testing.T) {
	t.Parallel()
	keyPEM := newPublicKeyPEM(t)
	tests := []struct {
		name    string
		id      plugins.SignerIdentity
		wantErr error
	}{
		{name: "keyless exact subject", id: keyless(issuerExample, subjectCI)},
		{name: "keyless subject pattern", id: plugins.SignerIdentity{Issuer: issuerExample, SubjectPattern: `https://github\.com/example/.*`}},
		{name: "public key", id: plugins.SignerIdentity{PublicKey: keyPEM}},
		{name: "OpenFief CI", id: plugins.OpenFiefCI().Identity},
		{name: "keyless pinned to a source repository and ref", id: fromSource(issuerExample, subjectCI, "https://github.com/example/sample-plugin", "refs/heads/main")},
		{name: "source repository not https", id: fromSource(issuerExample, subjectCI, "http://github.com/example/sample-plugin", ""), wantErr: plugins.ErrInvalidSigner},
		{name: "source repository as owner/repo", id: fromSource(issuerExample, subjectCI, "example/sample-plugin", ""), wantErr: plugins.ErrInvalidSigner},
		{name: "source ref not a git ref", id: fromSource(issuerExample, subjectCI, "", "main"), wantErr: plugins.ErrInvalidSigner},
		{name: "source pinning nothing", id: fromSource(issuerExample, subjectCI, "", ""), wantErr: plugins.ErrInvalidSigner},
		{name: "repository pinned alone", id: fromSource(issuerExample, subjectCI, "https://github.com/example/sample-plugin", "")},
		{name: "ref pinned alone", id: fromSource(issuerExample, subjectCI, "", "refs/tags/v1.0.0")},
		{
			name: "source without a subject", id: plugins.SignerIdentity{Source: &plugins.SignerSource{Repository: "https://github.com/example/sample-plugin"}},
			wantErr: plugins.ErrInvalidSigner,
		},
		{
			name:    "public key with a source",
			id:      plugins.SignerIdentity{PublicKey: keyPEM, Source: &plugins.SignerSource{Ref: "refs/heads/main"}},
			wantErr: plugins.ErrInvalidSigner,
		},
		{name: "empty", id: plugins.SignerIdentity{}, wantErr: plugins.ErrInvalidSigner},
		{name: "keyless and key", id: plugins.SignerIdentity{Issuer: issuerExample, Subject: subjectCI, PublicKey: keyPEM}, wantErr: plugins.ErrInvalidSigner},
		{name: "subject without issuer", id: keyless("", subjectCI), wantErr: plugins.ErrInvalidSigner},
		{name: "http issuer", id: keyless("http://issuer.example.com", subjectCI), wantErr: plugins.ErrInvalidSigner},
		{name: "issuer without subject", id: keyless(issuerExample, ""), wantErr: plugins.ErrInvalidSigner},
		{
			name:    "subject and pattern",
			id:      plugins.SignerIdentity{Issuer: issuerExample, Subject: subjectCI, SubjectPattern: ".*"},
			wantErr: plugins.ErrInvalidSigner,
		},
		{name: "bad pattern", id: plugins.SignerIdentity{Issuer: issuerExample, SubjectPattern: "("}, wantErr: plugins.ErrInvalidSigner},
		{name: "not PEM", id: plugins.SignerIdentity{PublicKey: "not a key"}, wantErr: plugins.ErrInvalidSigner},
		{name: "private key block", id: plugins.SignerIdentity{PublicKey: strings.ReplaceAll(keyPEM, "PUBLIC", "PRIVATE")}, wantErr: plugins.ErrInvalidSigner},
		{name: "two PEM blocks", id: plugins.SignerIdentity{PublicKey: keyPEM + keyPEM}, wantErr: plugins.ErrInvalidSigner},
		// D-187: a GitHub-issued identity names a workflow any repository
		// may call, so it is pinned to its source repository and ref.
		{name: "GitHub pinned to a repository and ref", id: fromSource(plugins.GitHubActionsIssuer, subjectCI, "https://github.com/example/sample-plugin", "refs/heads/main")},
		{name: "GitHub by subject alone", id: keyless(plugins.GitHubActionsIssuer, subjectCI), wantErr: plugins.ErrInvalidSigner},
		{name: "GitHub pinned to a repository alone", id: fromSource(plugins.GitHubActionsIssuer, subjectCI, "https://github.com/example/sample-plugin", ""), wantErr: plugins.ErrInvalidSigner},
		{name: "GitHub pinned to a ref alone", id: fromSource(plugins.GitHubActionsIssuer, subjectCI, "", "refs/heads/main"), wantErr: plugins.ErrInvalidSigner},
		{
			name: "GitHub subject pattern without a source", id: plugins.SignerIdentity{Issuer: plugins.GitHubActionsIssuer, SubjectPattern: `https://github\.com/example/.*`},
			wantErr: plugins.ErrInvalidSigner,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, tt.id.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSignerIdentityMatches(t *testing.T) {
	t.Parallel()
	keyA, keyB := newPublicKeyPEM(t), newPublicKeyPEM(t)
	pattern := plugins.SignerIdentity{Issuer: issuerExample, SubjectPattern: `https://github\.com/example/sample-plugin/.*@refs/heads/main`}
	tests := []struct {
		name   string
		id     plugins.SignerIdentity
		signed plugins.SignerIdentity
		want   bool
	}{
		{name: "same issuer and subject", id: keyless(issuerExample, subjectCI), signed: keyless(issuerExample, subjectCI), want: true},
		{name: "other subject", id: keyless(issuerExample, subjectCI), signed: keyless(issuerExample, subjectOtherRef)},
		{name: "other issuer", id: keyless(issuerExample, subjectCI), signed: keyless(plugins.GitHubActionsIssuer, subjectCI)},
		{name: "subject prefix only", id: keyless(issuerExample, subjectCI), signed: keyless(issuerExample, subjectCI+"x")},
		{name: "pattern matches whole subject", id: pattern, signed: keyless(issuerExample, subjectCI), want: true},
		{name: "pattern is anchored at the end", id: pattern, signed: keyless(issuerExample, subjectCI+"/evil")},
		{name: "pattern is anchored at the start", id: pattern, signed: keyless(issuerExample, "x"+subjectCI)},
		{name: "pattern on other ref", id: pattern, signed: keyless(issuerExample, subjectOtherRef)},
		{name: "same key", id: plugins.SignerIdentity{PublicKey: keyA}, signed: plugins.SignerIdentity{PublicKey: keyA}, want: true},
		{
			name: "same key, other PEM whitespace", id: plugins.SignerIdentity{PublicKey: keyA},
			signed: plugins.SignerIdentity{PublicKey: "\n" + keyA + "\n"}, want: true,
		},
		{name: "other key", id: plugins.SignerIdentity{PublicKey: keyA}, signed: plugins.SignerIdentity{PublicKey: keyB}},
		{name: "key against keyless", id: plugins.SignerIdentity{PublicKey: keyA}, signed: keyless(issuerExample, subjectCI)},
		{name: "keyless against key", id: keyless(issuerExample, subjectCI), signed: plugins.SignerIdentity{PublicKey: keyA}},
		{name: "pattern against empty subject", id: plugins.SignerIdentity{Issuer: issuerExample, SubjectPattern: ".*"}, signed: keyless(issuerExample, "")},
		{name: "invalid identity matches nothing", id: plugins.SignerIdentity{}, signed: plugins.SignerIdentity{}},
		{name: "OpenFief CI", id: plugins.OpenFiefCI().Identity, signed: openFiefCISigned(), want: true},
		{
			name: "OpenFief CI subject without a source", id: plugins.OpenFiefCI().Identity,
			signed: keyless(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject),
		},
		{
			name: "OpenFief CI subject called from another repository", id: plugins.OpenFiefCI().Identity,
			signed: fromSource(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject, "https://github.com/attacker/openfief-caller", plugins.OpenFiefCISourceRef),
		},
		{
			name: "OpenFief CI subject called from another ref", id: plugins.OpenFiefCI().Identity,
			signed: fromSource(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject, plugins.OpenFiefCISourceRepository, "refs/heads/feature"),
		},
		{
			name: "unpinned source accepts any repository", id: keyless(issuerExample, subjectCI),
			signed: fromSource(issuerExample, subjectCI, "https://github.com/example/other", "refs/heads/dev"), want: true,
		},
		{
			name: "a repository pinned alone accepts any ref", id: fromSource(issuerExample, subjectCI, "https://github.com/example/sample-plugin", ""),
			signed: fromSource(issuerExample, subjectCI, "https://github.com/example/sample-plugin", "refs/tags/v1.0.0"), want: true,
		},
		{
			name: "a repository pinned alone refuses another repository", id: fromSource(issuerExample, subjectCI, "https://github.com/example/sample-plugin", ""),
			signed: fromSource(issuerExample, subjectCI, "https://github.com/example/other", "refs/tags/v1.0.0"),
		},
		{
			name: "OpenFief CI from a fork", id: plugins.OpenFiefCI().Identity,
			signed: keyless(plugins.GitHubActionsIssuer, "https://github.com/fork/openfief/.github/workflows/images.yml@refs/heads/main"),
		},
		{
			name: "OpenFief CI from another branch", id: plugins.OpenFiefCI().Identity,
			signed: keyless(plugins.GitHubActionsIssuer, "https://github.com/arkadiuss/openfief/.github/workflows/images.yml@refs/heads/dev"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.id.Matches(tt.signed); got != tt.want {
				t.Errorf("%s.Matches(%s) = %v, want %v", tt.id, tt.signed, got, tt.want)
			}
		})
	}
}

func TestSignerIdentityString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		id   plugins.SignerIdentity
		want string
	}{
		{name: "keyless", id: keyless(issuerExample, subjectCI), want: subjectCI + " (issuer " + issuerExample + ")"},
		{name: "pattern", id: plugins.SignerIdentity{Issuer: issuerExample, SubjectPattern: "a.*"}, want: "~a.* (issuer " + issuerExample + ")"},
		{
			name: "keyless with a source", id: fromSource(issuerExample, subjectCI, "https://github.com/example/sample-plugin", "refs/heads/main"),
			want: subjectCI + " (issuer " + issuerExample + ", source https://github.com/example/sample-plugin, ref refs/heads/main)",
		},
		{name: "bad key", id: plugins.SignerIdentity{PublicKey: "x"}, want: "invalid public key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.id.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
	key := plugins.SignerIdentity{PublicKey: newPublicKeyPEM(t)}.String()
	if !strings.HasPrefix(key, "public key sha256:") || len(key) != len("public key sha256:")+64 {
		t.Errorf("key String() = %q, want a SHA-256 fingerprint", key)
	}
}

// policyFixture is a trust policy with one catalog entry (imageA, published
// by a keyless publisher) and two owner signers: a CI pattern and a key.
type policyFixture struct {
	policy    plugins.TrustPolicy
	publisher plugins.Signer
	ownerCI   plugins.Signer
	ownerKey  plugins.Signer
}

func newPolicyFixture(t *testing.T) policyFixture {
	t.Helper()
	img, err := plugins.ParseImageRef(imageA)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	f := policyFixture{
		publisher: plugins.Signer{Name: "example-publisher", Kind: plugins.SignerPublisher, Identity: keyless(issuerExample, subjectCI)},
		ownerCI: plugins.Signer{Name: "owner-ci", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{
			Issuer: plugins.GitHubActionsIssuer, SubjectPattern: `https://github\.com/owner-example/.*@refs/tags/v.*`,
			Source: &plugins.SignerSource{Repository: ownerRepository, Ref: ownerRef},
		}},
		ownerKey: plugins.Signer{Name: "owner-laptop", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{PublicKey: newPublicKeyPEM(t)}},
	}
	catalog := plugins.Catalog{Entries: []plugins.CatalogEntry{{Name: "sample", Image: img, Publisher: f.publisher}}}
	f.policy, err = plugins.NewTrustPolicy(catalog, []plugins.Signer{f.ownerCI, f.ownerKey})
	if err != nil {
		t.Fatalf("NewTrustPolicy: %v", err)
	}
	return f
}

func TestTrustPolicyDecide(t *testing.T) {
	t.Parallel()
	f := newPolicyFixture(t)
	ci := openFiefCISigned()
	publisher := keyless(issuerExample, subjectCI)
	ownerTag := fromSource(plugins.GitHubActionsIssuer, "https://github.com/owner-example/plugin/.github/workflows/release.yml@refs/tags/v1.2.0",
		ownerRepository, ownerRef)
	stranger := keyless(plugins.GitHubActionsIssuer, "https://github.com/stranger/plugin/.github/workflows/release.yml@refs/heads/main")
	manifest := samplePlugin().Manifest
	widened := samplePlugin().Manifest
	widened.Egress = append(widened.Egress, host("exfil.example.net", 443))
	tests := []struct {
		name     string
		image    string
		sigs     []plugins.SignerIdentity
		declared *plugins.Manifest
		want     plugins.TrustDecision
		wantErr  error
	}{
		{
			name: "catalog pins the digest and its publisher signed it", image: imageA, sigs: []plugins.SignerIdentity{publisher},
			want: plugins.TrustDecision{Trust: plugins.Verified, Signer: f.publisher},
		},
		{
			name: "a mirror of the pinned digest is verified too", image: "mirror.example.org/sample@" + digestA,
			sigs: []plugins.SignerIdentity{publisher}, want: plugins.TrustDecision{Trust: plugins.Verified, Signer: f.publisher},
		},
		{
			name: "verified wins over another trusted signature", image: imageA, sigs: []plugins.SignerIdentity{ci, publisher},
			want: plugins.TrustDecision{Trust: plugins.Verified, Signer: f.publisher},
		},
		{
			name: "OpenFief CI signed an image the catalog does not pin", image: imageB, sigs: []plugins.SignerIdentity{ci},
			want: plugins.TrustDecision{Trust: plugins.Unverified, Signer: plugins.OpenFiefCI()},
		},
		{
			name: "pinned digest signed by a trusted signer that is not its publisher", image: imageA, sigs: []plugins.SignerIdentity{ci},
			want: plugins.TrustDecision{Trust: plugins.Unverified, Signer: plugins.OpenFiefCI()},
		},
		{
			name: "publisher signed a version the catalog does not pin", image: imageB, sigs: []plugins.SignerIdentity{publisher},
			want: plugins.TrustDecision{Trust: plugins.Unverified, Signer: f.publisher},
		},
		{
			name: "owner's CI identity", image: imageB, sigs: []plugins.SignerIdentity{ownerTag},
			want: plugins.TrustDecision{Trust: plugins.Unverified, Signer: f.ownerCI},
		},
		{
			name: "owner's key", image: imageB, sigs: []plugins.SignerIdentity{f.ownerKey.Identity},
			want: plugins.TrustDecision{Trust: plugins.Unverified, Signer: f.ownerKey},
		},
		{
			name: "declared manifest equals the signed label", image: imageB, sigs: []plugins.SignerIdentity{ci}, declared: &manifest,
			want: plugins.TrustDecision{Trust: plugins.Unverified, Signer: plugins.OpenFiefCI()},
		},
		{name: "no signature", image: imageA, wantErr: plugins.ErrUnsigned},
		{name: "untrusted signer", image: imageA, sigs: []plugins.SignerIdentity{stranger}, wantErr: plugins.ErrUntrustedSigner},
		{
			name: "OpenFief CI's reusable workflow called from another repository", image: imageB,
			sigs: []plugins.SignerIdentity{fromSource(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject,
				"https://github.com/attacker/openfief-caller", plugins.OpenFiefCISourceRef)},
			wantErr: plugins.ErrUntrustedSigner,
		},
		{
			name: "OpenFief CI's reusable workflow called from another ref", image: imageB,
			sigs: []plugins.SignerIdentity{fromSource(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject,
				plugins.OpenFiefCISourceRepository, "refs/heads/feature")},
			wantErr: plugins.ErrUntrustedSigner,
		},
		{
			name: "untrusted key", image: imageA, sigs: []plugins.SignerIdentity{{PublicKey: newPublicKeyPEM(t)}},
			wantErr: plugins.ErrUntrustedSigner,
		},
		{
			name: "declared manifest asks for more than the signed label", image: imageB, sigs: []plugins.SignerIdentity{ci},
			declared: &widened, wantErr: plugins.ErrManifestMismatch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			img, err := plugins.ParseImageRef(tt.image)
			if err != nil {
				t.Fatalf("image: %v", err)
			}
			signed := plugins.SignedImage{Image: img, Signatures: tt.sigs, Manifest: samplePlugin().Manifest}
			got, err := f.policy.Decide(&signed, tt.declared)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Decide error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Decide (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewTrustPolicy(t *testing.T) {
	t.Parallel()
	img, err := plugins.ParseImageRef(imageA)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	imgB, err := plugins.ParseImageRef(imageB)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	publisher := plugins.Signer{Name: "p", Kind: plugins.SignerPublisher, Identity: keyless(issuerExample, subjectCI)}
	other := plugins.Signer{Name: "q", Kind: plugins.SignerPublisher, Identity: keyless(issuerExample, subjectOtherRef)}
	namesake := plugins.Signer{Name: "p", Kind: plugins.SignerPublisher, Identity: keyless(issuerExample, subjectOtherRef)}
	// pinned and pinnedAgain are one publisher whose sources are separate
	// pointers to equal values: the policy compares sources by value.
	pinned := plugins.Signer{Name: "r", Kind: plugins.SignerPublisher, Identity: fromSource(issuerExample, subjectCI, ownerRepository, ownerRef)}
	pinnedAgain := plugins.Signer{Name: "r", Kind: plugins.SignerPublisher, Identity: fromSource(issuerExample, subjectCI, ownerRepository, ownerRef)}
	owner := plugins.Signer{Name: "o", Kind: plugins.SignerOwner, Identity: keyless(issuerExample, subjectOtherRef)}
	// entries is a catalog with one entry per publisher, on alternating images.
	entries := func(publishers ...plugins.Signer) plugins.Catalog {
		imgs := []plugins.ImageRef{img, imgB}
		var c plugins.Catalog
		for i, s := range publishers {
			c.Entries = append(c.Entries, plugins.CatalogEntry{Name: "sample", Image: imgs[i%len(imgs)], Publisher: s})
		}
		return c
	}
	tests := []struct {
		name        string
		catalog     plugins.Catalog
		owner       []plugins.Signer
		wantSigners []string
		wantErr     error
	}{
		{name: "CI alone", wantSigners: []string{"openfief-ci"}},
		{
			name:        "CI, then publishers, then owner signers",
			catalog:     plugins.Catalog{Entries: []plugins.CatalogEntry{{Name: "sample", Image: img, Publisher: publisher}}},
			owner:       []plugins.Signer{owner},
			wantSigners: []string{"openfief-ci", "p", "o"},
		},
		{
			name:        "a publisher of several entries is trusted once",
			catalog:     entries(publisher, publisher, publisher),
			wantSigners: []string{"openfief-ci", "p"},
		},
		{
			name:        "a publisher whose source pins are equal copies is trusted once",
			catalog:     entries(pinned, pinnedAgain),
			wantSigners: []string{"openfief-ci", "r"},
		},
		{
			name:        "two publishers are trusted once each, in catalog order",
			catalog:     entries(other, publisher, other, publisher),
			wantSigners: []string{"openfief-ci", "q", "p"},
		},
		{
			name:        "publishers sharing only a name are both trusted",
			catalog:     entries(publisher, namesake),
			wantSigners: []string{"openfief-ci", "p", "p"},
		},
		{
			name:        "an owner signer listed twice is trusted once",
			owner:       []plugins.Signer{owner, owner},
			wantSigners: []string{"openfief-ci", "o"},
		},
		{
			name:        "an owner signer with a publisher's identity is kept apart",
			catalog:     entries(publisher),
			owner:       []plugins.Signer{{Name: "p", Kind: plugins.SignerOwner, Identity: publisher.Identity}},
			wantSigners: []string{"openfief-ci", "p", "p"},
		},
		{name: "owner signer posing as a publisher", owner: []plugins.Signer{publisher}, wantErr: plugins.ErrInvalidSigner},
		{
			name:    "invalid owner signer",
			owner:   []plugins.Signer{{Name: "o", Kind: plugins.SignerOwner, Identity: keyless("", subjectCI)}},
			wantErr: plugins.ErrInvalidSigner,
		},
		{
			name:    "invalid catalog",
			catalog: plugins.Catalog{Entries: []plugins.CatalogEntry{{Name: "sample", Image: img, Publisher: owner}}},
			wantErr: plugins.ErrInvalidCatalog,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := plugins.NewTrustPolicy(tt.catalog, tt.owner)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("NewTrustPolicy error (-want +got):\n%s", diff)
			}
			var names []string
			for _, s := range p.Signers() {
				names = append(names, s.Name)
			}
			if diff := cmp.Diff(tt.wantSigners, names); diff != "" {
				t.Errorf("Signers (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSignerValidate(t *testing.T) {
	t.Parallel()
	id := keyless(issuerExample, subjectCI)
	tests := []struct {
		name    string
		signer  plugins.Signer
		wantErr error
	}{
		{name: "valid", signer: plugins.Signer{Name: "s", Kind: plugins.SignerOwner, Identity: id}},
		{name: "no name", signer: plugins.Signer{Kind: plugins.SignerOwner, Identity: id}, wantErr: plugins.ErrInvalidSigner},
		{name: "unknown kind", signer: plugins.Signer{Name: "s", Kind: "Anyone", Identity: id}, wantErr: plugins.ErrInvalidSigner},
		{name: "bad identity", signer: plugins.Signer{Name: "s", Kind: plugins.SignerOwner}, wantErr: plugins.ErrInvalidSigner},
		{name: "OpenFief CI pins its source", signer: plugins.OpenFiefCI()},
		{
			name:    "OpenFief CI without a source repository",
			signer:  plugins.Signer{Name: "ci", Kind: plugins.SignerOpenFiefCI, Identity: fromSource(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject, "", plugins.OpenFiefCISourceRef)},
			wantErr: plugins.ErrInvalidSigner,
		},
		{
			name:    "OpenFief CI without a source ref",
			signer:  plugins.Signer{Name: "ci", Kind: plugins.SignerOpenFiefCI, Identity: fromSource(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject, plugins.OpenFiefCISourceRepository, "")},
			wantErr: plugins.ErrInvalidSigner,
		},
		{
			name:    "OpenFief CI by subject alone",
			signer:  plugins.Signer{Name: "ci", Kind: plugins.SignerOpenFiefCI, Identity: keyless(plugins.GitHubActionsIssuer, plugins.OpenFiefCISubject)},
			wantErr: plugins.ErrInvalidSigner,
		},
		{
			name:    "OpenFief CI as a key",
			signer:  plugins.Signer{Name: "ci", Kind: plugins.SignerOpenFiefCI, Identity: plugins.SignerIdentity{PublicKey: newPublicKeyPEM(t)}},
			wantErr: plugins.ErrInvalidSigner,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, tt.signer.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() (-want +got):\n%s", diff)
			}
		})
	}
}
