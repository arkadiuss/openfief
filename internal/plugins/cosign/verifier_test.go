package cosign

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/testing/ca"
	"github.com/sigstore/sigstore-go/pkg/verify"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Synthetic signer subjects; none is a real publisher's.
const (
	subjectPublisher = "https://github.com/example/sample-plugin/.github/workflows/release.yml@refs/heads/main"
	subjectFork      = "https://github.com/fork/openfief/.github/workflows/images.yml@refs/heads/main"
	issuerOther      = "https://issuer.example.com"
	// sourcePublisher is the publisher's own repository; sourceAttacker is
	// a repository whose workflow calls someone else's reusable workflow.
	sourcePublisher = "https://github.com/example/sample-plugin"
	sourceAttacker  = "https://github.com/attacker/caller"
	refFeature      = "refs/heads/feature"
)

// virtualTestVerifier is a Verifier whose keyless signatures chain to the
// virtual Sigstore vs. Its certificates carry no SCTs, so unlike New it does
// not require one; TestProductionVerifierRequiresSCT covers that.
func virtualTestVerifier(t *testing.T, policy *plugins.TrustPolicy, vs root.TrustedMaterial) *Verifier {
	t.Helper()
	v, err := newVerifier(policy, vs, []verify.VerifierOption{verify.WithTransparencyLog(1), verify.WithObserverTimestamps(1)}, nil)
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	return v
}

// world is one test's registry, virtual Sigstore and signers.
type world struct {
	reg *testRegistry
	vs  *ca.VirtualSigstore
	// gh certifies GitHub Actions runs with their source repository and
	// ref, as Fulcio does; its certificates are logged in vs.
	gh *githubCA
	// trusted is what the verifier trusts: vs and gh, unless a case
	// changes it.
	trusted   root.TrustedMaterial
	publisher plugins.Signer
	// untrusted are the key signers a case's setup made and expects
	// Inspect to name (D-169): the keys are fresh, so a case can't spell
	// them out.
	untrusted []plugins.UntrustedSigner
}

// foreignCA trusts everything of logs except its Fulcio CAs, which come
// from ca: a log, CT log and timestamp authority that are trusted, with a
// certificate chain that is not.
type foreignCA struct {
	root.TrustedMaterial

	ca root.TrustedMaterial
}

func (f foreignCA) FulcioCertificateAuthorities() []root.CertificateAuthority {
	return f.ca.FulcioCertificateAuthorities()
}

func newWorld(t *testing.T, referrersAPI bool, wrap func(http.Handler) http.Handler) *world {
	t.Helper()
	vs, err := ca.NewVirtualSigstore()
	if err != nil {
		t.Fatalf("virtual sigstore: %v", err)
	}
	gh := newGitHubCA(t)
	return &world{
		reg: newTestRegistry(t, referrersAPI, wrap), vs: vs, gh: gh, trusted: withCA{TrustedMaterial: vs, extra: gh.authority},
		publisher: plugins.Signer{
			Name: "example-publisher", Kind: plugins.SignerPublisher,
			Identity: plugins.SignerIdentity{
				Issuer: plugins.GitHubActionsIssuer, Subject: subjectPublisher,
				Source: &plugins.SignerSource{Repository: sourcePublisher, Ref: plugins.OpenFiefCISourceRef},
			},
		},
	}
}

// signPublisher attaches a keyless signature of image by the publisher's
// workflow, run from its own repository's main branch, to which its
// identity is pinned (D-187).
func (w *world) signPublisher(t *testing.T, image plugins.ImageRef) {
	t.Helper()
	w.signRun(t, image, githubRun{subject: subjectPublisher, issuer: plugins.GitHubActionsIssuer, repository: sourcePublisher, ref: plugins.OpenFiefCISourceRef})
}

// signCI attaches an OpenFief CI keyless signature of image.
func (w *world) signCI(t *testing.T, image plugins.ImageRef) {
	t.Helper()
	w.signRun(t, image, openFiefCIRun())
}

// signRun attaches a keyless signature of image made in the GitHub Actions
// run run.
func (w *world) signRun(t *testing.T, image plugins.ImageRef, run githubRun) {
	t.Helper()
	w.reg.attach(t, image, githubBundle(t, w.vs, w.gh, run, statement(t, image, SignPredicateType), time.Now().Add(5*time.Minute)))
}

// outcome is the trust decision for an inspected image.
type outcome struct {
	Trust  plugins.Trust
	Signer string
}

// inspectCase is one TestInspect case.
type inspectCase struct {
	name string
	// referrersAPI serves the OCI Referrers API; otherwise cosign's
	// referrers tag is used, as on GHCR.
	referrersAPI bool
	wrap         func(http.Handler) http.Handler
	// setup pushes and signs the image, and returns it, the catalog and
	// the owner's signers.
	setup    func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer)
	declared string
	want     outcome
	// wantErr is the Inspect or Decide error.
	wantErr error
	// wantUntrusted are the signers Inspect names that the policy doesn't
	// hold (D-169).
	wantUntrusted []plugins.UntrustedSigner
}

func TestInspect(t *testing.T) {
	t.Parallel()
	tests := []inspectCase{
		{
			name: "valid CI signature, referrers tag",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.signCI(t, img)
				return img, plugins.Catalog{}, nil
			},
			declared: labelView,
			want:     outcome{plugins.Unverified, "openfief-ci"},
		},
		{
			name: "valid CI signature, Referrers API", referrersAPI: true,
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.signCI(t, img)
				return img, plugins.Catalog{}, nil
			},
			declared: labelView,
			want:     outcome{plugins.Unverified, "openfief-ci"},
		},
		{
			name: "multi-platform index with an attestation manifest",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.pushIndex(t, labelledImage(t, labelConnector), labelledImage(t, labelConnector))
				w.signCI(t, img)
				return img, plugins.Catalog{}, nil
			},
			declared: labelConnector,
			want:     outcome{plugins.Unverified, "openfief-ci"},
		},
		{
			name: "catalog pins the digest and its publisher signed it",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.signPublisher(t, img)
				return img, plugins.Catalog{Entries: []plugins.CatalogEntry{{Name: "sample", Image: img, Publisher: w.publisher}}}, nil
			},
			declared: labelView,
			want:     outcome{plugins.Verified, "example-publisher"},
		},
		{
			name: "owner-added key",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				k := keyBundle(t, statement(t, img, SignPredicateType))
				w.reg.attach(t, img, k.bundle)
				return img, plugins.Catalog{}, []plugins.Signer{{Name: "laptop", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{PublicKey: k.publicKey}}}
			},
			declared: labelView,
			want:     outcome{plugins.Unverified, "laptop"},
		},
		{
			name: "owner's keyless CI identity, by pattern",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.signPublisher(t, img)
				return img, plugins.Catalog{}, []plugins.Signer{{Name: "my-ci", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{
					Issuer: plugins.GitHubActionsIssuer, SubjectPattern: `https://github\.com/example/.*`,
					Source: &plugins.SignerSource{Repository: sourcePublisher, Ref: plugins.OpenFiefCISourceRef},
				}}}
			},
			declared: labelView,
			want:     outcome{plugins.Unverified, "my-ci"},
		},
		{
			name: "unsigned image",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				return w.reg.push(t, labelledImage(t, labelView)), plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "wrong identity: OpenFief's workflow in a fork",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keylessBundle(t, w.vs, subjectFork, plugins.GitHubActionsIssuer,
					statement(t, img, SignPredicateType), time.Now().Add(5*time.Minute)))
				return img, plugins.Catalog{}, nil
			},
			wantErr:       plugins.ErrUntrustedSigner,
			wantUntrusted: []plugins.UntrustedSigner{{Issuer: plugins.GitHubActionsIssuer, Subject: subjectFork}},
		},
		{
			name: "wrong identity: OpenFief's reusable workflow called from another repository",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				run := openFiefCIRun()
				run.repository = sourceAttacker
				w.signRun(t, img, run)
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUntrustedSigner,
			wantUntrusted: []plugins.UntrustedSigner{{
				Issuer: plugins.GitHubActionsIssuer, Subject: plugins.OpenFiefCISubject,
				Source: &plugins.SignerSource{Repository: sourceAttacker, Ref: plugins.OpenFiefCISourceRef},
			}},
		},
		{
			name: "wrong identity: OpenFief's reusable workflow called from another ref",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				run := openFiefCIRun()
				run.ref = refFeature
				w.signRun(t, img, run)
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUntrustedSigner,
			wantUntrusted: []plugins.UntrustedSigner{{
				Issuer: plugins.GitHubActionsIssuer, Subject: plugins.OpenFiefCISubject,
				Source: &plugins.SignerSource{Repository: plugins.OpenFiefCISourceRepository, Ref: refFeature},
			}},
		},
		{
			name: "wrong identity: OpenFief's subject with no source repository",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keylessBundle(t, w.vs, plugins.OpenFiefCISubject, plugins.GitHubActionsIssuer,
					statement(t, img, SignPredicateType), time.Now().Add(5*time.Minute)))
				return img, plugins.Catalog{}, nil
			},
			wantErr:       plugins.ErrUntrustedSigner,
			wantUntrusted: []plugins.UntrustedSigner{{Issuer: plugins.GitHubActionsIssuer, Subject: plugins.OpenFiefCISubject}},
		},
		{
			name: "a broad owner pattern pinned to the caller's repository takes its signature as the owner's, never as OpenFief CI's",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				run := openFiefCIRun()
				run.repository = sourceAttacker
				w.signRun(t, img, run)
				return img, plugins.Catalog{}, []plugins.Signer{{Name: "my-ci", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{
					Issuer: plugins.GitHubActionsIssuer, SubjectPattern: `https://github\.com/.*`,
					Source: &plugins.SignerSource{Repository: sourceAttacker, Ref: plugins.OpenFiefCISourceRef},
				}}}
			},
			declared: labelView,
			want:     outcome{plugins.Unverified, "my-ci"},
		},
		{
			name: "owner's keyless identity pinned to its source repository and ref",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.signRun(t, img, githubRun{subject: subjectPublisher, issuer: plugins.GitHubActionsIssuer, repository: sourcePublisher, ref: plugins.OpenFiefCISourceRef})
				return img, plugins.Catalog{}, []plugins.Signer{{Name: "my-ci", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{
					Issuer: plugins.GitHubActionsIssuer, Subject: subjectPublisher, Source: &plugins.SignerSource{Repository: sourcePublisher, Ref: plugins.OpenFiefCISourceRef},
				}}}
			},
			declared: labelView,
			want:     outcome{plugins.Unverified, "my-ci"},
		},
		{
			name: "owner's keyless identity pinned to its source repository, signed from another",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.signRun(t, img, githubRun{subject: subjectPublisher, issuer: plugins.GitHubActionsIssuer, repository: sourceAttacker, ref: plugins.OpenFiefCISourceRef})
				return img, plugins.Catalog{}, []plugins.Signer{{Name: "my-ci", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{
					Issuer: plugins.GitHubActionsIssuer, Subject: subjectPublisher, Source: &plugins.SignerSource{Repository: sourcePublisher, Ref: plugins.OpenFiefCISourceRef},
				}}}
			},
			wantErr: plugins.ErrUntrustedSigner,
			wantUntrusted: []plugins.UntrustedSigner{{
				Issuer: plugins.GitHubActionsIssuer, Subject: subjectPublisher, Source: &plugins.SignerSource{Repository: sourceAttacker, Ref: plugins.OpenFiefCISourceRef},
			}},
		},
		{
			name: "wrong identity: right subject from another issuer",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keylessBundle(t, w.vs, plugins.OpenFiefCISubject, issuerOther,
					statement(t, img, SignPredicateType), time.Now().Add(5*time.Minute)))
				return img, plugins.Catalog{}, nil
			},
			wantErr:       plugins.ErrUntrustedSigner,
			wantUntrusted: []plugins.UntrustedSigner{{Issuer: issuerOther, Subject: plugins.OpenFiefCISubject}},
		},
		{
			name: "untrusted signer: a key nobody added",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				stranger := keyBundle(t, statement(t, img, SignPredicateType))
				w.reg.attach(t, img, stranger.bundle)
				w.untrusted = append(w.untrusted, plugins.UntrustedSigner{KeySHA256: stranger.fingerprint})
				other := keyBundle(t, statement(t, img, SignPredicateType)).publicKey
				return img, plugins.Catalog{}, []plugins.Signer{{Name: "laptop", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{PublicKey: other}}}
			},
			wantErr: plugins.ErrUntrustedSigner,
		},
		{
			name: "untrusted signer and a trusted one: the trusted one decides",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keylessBundle(t, w.vs, subjectFork, plugins.GitHubActionsIssuer,
					statement(t, img, SignPredicateType), time.Now().Add(5*time.Minute)))
				w.signCI(t, img)
				return img, plugins.Catalog{}, nil
			},
			declared:      labelView,
			want:          outcome{plugins.Unverified, "openfief-ci"},
			wantUntrusted: []plugins.UntrustedSigner{{Issuer: plugins.GitHubActionsIssuer, Subject: subjectFork}},
		},
		{
			name: "key signature over another image names nobody",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				other := w.reg.push(t, labelledImage(t, labelConnector))
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keyBundle(t, statement(t, other, SignPredicateType)).bundle)
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "key attestation is not a signature and names nobody",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keyBundle(t, statement(t, img, "https://spdx.dev/Document")).bundle)
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "forged key signature: another key under a trusted key's hint names nobody",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				trusted := keyBundle(t, statement(t, img, SignPredicateType))
				w.reg.attach(t, img, keyBundleWithHint(t, statement(t, img, SignPredicateType), trusted.hint))
				return img, plugins.Catalog{}, []plugins.Signer{{Name: "laptop", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{PublicKey: trusted.publicKey}}}
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "digest mismatch: another image's signature attached",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				signed := w.reg.push(t, labelledImage(t, labelView))
				img := w.reg.push(t, labelledImage(t, labelConnector))
				w.reg.attach(t, img, keylessBundle(t, w.vs, plugins.OpenFiefCISubject, plugins.GitHubActionsIssuer,
					statement(t, signed, SignPredicateType), time.Now().Add(5*time.Minute)))
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "attestation by a trusted signer is not a signature",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keylessBundle(t, w.vs, plugins.OpenFiefCISubject, plugins.GitHubActionsIssuer,
					statement(t, img, "https://spdx.dev/Document"), time.Now().Add(5*time.Minute)))
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "expired certificate: logged after it expired",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keylessBundle(t, w.vs, plugins.OpenFiefCISubject, plugins.GitHubActionsIssuer,
					statement(t, img, SignPredicateType), time.Now().Add(30*time.Minute)))
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "untrusted Sigstore: certificate and log entry from another instance",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				rogue, err := ca.NewVirtualSigstore()
				if err != nil {
					t.Fatalf("rogue sigstore: %v", err)
				}
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keylessBundle(t, rogue, plugins.OpenFiefCISubject, plugins.GitHubActionsIssuer,
					statement(t, img, SignPredicateType), time.Now().Add(5*time.Minute)))
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "invalid chain: trusted log, certificate from an untrusted CA",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				rogue, err := ca.NewVirtualSigstore()
				if err != nil {
					t.Fatalf("rogue sigstore: %v", err)
				}
				w.trusted = foreignCA{TrustedMaterial: rogue, ca: w.vs}
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, keylessBundle(t, rogue, plugins.OpenFiefCISubject, plugins.GitHubActionsIssuer,
					statement(t, img, SignPredicateType), time.Now().Add(5*time.Minute)))
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "garbage referrer is not a signature",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.reg.attach(t, img, []byte(`{"mediaType":"not a bundle"}`))
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrUnsigned,
		},
		{
			name: "declared manifest asks for more than the signed label",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.signCI(t, img)
				return img, plugins.Catalog{}, nil
			},
			declared: labelConnector,
			wantErr:  plugins.ErrManifestMismatch,
		},
		{
			name: "signed image without a manifest label",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, ""))
				w.signCI(t, img)
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrManifestLabel,
		},
		{
			name: "index whose platform images carry different manifests",
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.pushIndex(t, labelledImage(t, labelView), labelledImage(t, labelConnector))
				w.signCI(t, img)
				return img, plugins.Catalog{}, nil
			},
			wantErr: plugins.ErrManifestLabel,
		},
		{
			name: "tampered manifest label: the registry serves a widened config",
			wrap: tamperConfig(`\"capabilities\":[\"view\"]`, `\"capabilities\":[\"connector\"]`),
			setup: func(t *testing.T, w *world) (plugins.ImageRef, plugins.Catalog, []plugins.Signer) {
				t.Helper()
				img := w.reg.push(t, labelledImage(t, labelView))
				w.signCI(t, img)
				return img, plugins.Catalog{}, nil
			},
			declared: labelConnector,
			wantErr:  errRegistry,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkInspectCase(t, &tt)
		})
	}
}

// checkInspectCase runs one TestInspect case and checks its decision, its
// error and the untrusted signers Inspect named, the case's own and those
// its setup made with fresh keys (world.untrusted), each also named in the
// refusal (D-169).
func checkInspectCase(t *testing.T, tt *inspectCase) {
	t.Helper()
	w := newWorld(t, tt.referrersAPI, tt.wrap)
	got, err := runInspectCase(t, tt, w)
	wantUntrusted := append(slices.Clone(tt.wantUntrusted), w.untrusted...)
	if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
		t.Fatalf("Inspect and Decide error (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(tt.want, got.decided); diff != "" {
		t.Errorf("decision (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(wantUntrusted, got.untrusted, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("untrusted signers (-want +got):\n%s", diff)
	}
	for _, u := range wantUntrusted {
		if err != nil && !strings.Contains(err.Error(), u.String()) {
			t.Errorf("refusal %q does not name %s", err, u)
		}
	}
}

// inspected is what runInspectCase found: the decision, and the untrusted
// signers Inspect named.
type inspected struct {
	decided   outcome
	untrusted []plugins.UntrustedSigner
}

// runInspectCase sets the case up in w, inspects its image and lets the
// trust policy decide.
func runInspectCase(t *testing.T, tt *inspectCase, w *world) (inspected, error) {
	t.Helper()
	img, catalog, owner := tt.setup(t, w)
	policy, err := plugins.NewTrustPolicy(catalog, owner)
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	signed, err := virtualTestVerifier(t, &policy, w.trusted).Inspect(t.Context(), img)
	if err != nil {
		return inspected{}, err
	}
	var declared *plugins.Manifest
	if tt.declared != "" {
		m, parseErr := plugins.ParseManifestLabel(tt.declared)
		if parseErr != nil {
			t.Fatalf("declared manifest: %v", parseErr)
		}
		declared = &m
	}
	d, err := policy.Decide(&signed, declared)
	if err != nil {
		return inspected{untrusted: signed.Untrusted}, fmt.Errorf("decide: %w", err)
	}
	return inspected{decided: outcome{d.Trust, d.Signer.Name}, untrusted: signed.Untrusted}, nil
}

func TestProductionVerifierRequiresSCT(t *testing.T) {
	t.Parallel()
	w := newWorld(t, false, nil)
	img := w.reg.push(t, labelledImage(t, labelView))
	w.signCI(t, img)
	policy, err := plugins.NewTrustPolicy(plugins.Catalog{}, nil)
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	v, err := New(&policy, w.trusted)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	signed, err := v.Inspect(t.Context(), img)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if len(signed.Signatures) != 0 {
		t.Errorf("a certificate without an SCT verified: %v", signed.Signatures)
	}
}

func TestInspectUnreachableRegistry(t *testing.T) {
	t.Parallel()
	policy, err := plugins.NewTrustPolicy(plugins.Catalog{}, nil)
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	vs, err := ca.NewVirtualSigstore()
	if err != nil {
		t.Fatalf("virtual sigstore: %v", err)
	}
	reg := newTestRegistry(t, false, func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) })
	})
	img := mustImageRef(t, reg.host+"/openfief/sample@sha256:"+
		"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1")
	v, err := newVerifier(&policy, vs, []verify.VerifierOption{verify.WithObserverTimestamps(1)},
		[]remote.Option{remote.WithRetryStatusCodes()})
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	_, err = v.Inspect(t.Context(), img)
	if !errors.Is(err, errRegistry) {
		t.Errorf("Inspect error = %v, want a registry error", err)
	}
}

func TestTrustedRoot(t *testing.T) {
	t.Parallel()
	tr, err := TrustedRoot()
	if err != nil {
		t.Fatalf("TrustedRoot: %v", err)
	}
	if len(tr.FulcioCertificateAuthorities()) == 0 || len(tr.RekorLogs()) == 0 || len(tr.CTLogs()) == 0 {
		t.Errorf("bundled trust root lacks Fulcio CAs, Rekor logs or CT logs")
	}
	policy, err := plugins.NewTrustPolicy(plugins.Catalog{}, nil)
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	if _, err := New(&policy, tr); err != nil {
		t.Errorf("New with the bundled trust root: %v", err)
	}
}
