package operator_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

// Synthetic signer identities.

func ciSignature() plugins.SignerIdentity {
	return plugins.SignerIdentity{Issuer: plugins.GitHubActionsIssuer, Subject: plugins.OpenFiefCISubject}
}

func publisherSignature() plugins.SignerIdentity {
	return plugins.SignerIdentity{Issuer: "https://issuer.example.com", Subject: "https://example.com/release"}
}

func strangerSignature() plugins.SignerIdentity {
	return plugins.SignerIdentity{Issuer: "https://issuer.example.com", Subject: "https://example.com/stranger"}
}

// errUnreachable stands in for a registry that cannot be read.
var errUnreachable = errors.New("registry unreachable")

// fakeSource is a SignedImageSource that serves fixed images and counts
// reads; an image it does not know fails like an unreachable registry.
type fakeSource struct {
	mu     sync.Mutex
	images map[string]plugins.SignedImage
	errs   map[string]error
	reads  int
}

func (f *fakeSource) Inspect(_ context.Context, image plugins.ImageRef) (plugins.SignedImage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reads++
	if err, ok := f.errs[image.String()]; ok {
		return plugins.SignedImage{}, err
	}
	img, ok := f.images[image.String()]
	if !ok {
		return plugins.SignedImage{}, fmt.Errorf("%s: %w", image, errUnreachable)
	}
	return img, nil
}

func mustImage(t *testing.T, s string) plugins.ImageRef {
	t.Helper()
	img, err := plugins.ParseImageRef(s)
	if err != nil {
		t.Fatalf("image %q: %v", s, err)
	}
	return img
}

// viewManifest is the manifest every fake image carries.
func viewManifest() plugins.Manifest {
	return plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityView}}
}

// testPolicy pins imageA in the catalog, published by publisherSignature.
func testPolicy(t *testing.T) *plugins.TrustPolicy {
	t.Helper()
	p, err := plugins.NewTrustPolicy(plugins.Catalog{Entries: []plugins.CatalogEntry{{
		Name: "sample", Image: mustImage(t, imageA),
		Publisher: plugins.Signer{Name: "example-publisher", Kind: plugins.SignerPublisher, Identity: publisherSignature()},
	}}}, nil)
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	return &p
}

func TestPolicyVerifier(t *testing.T) {
	t.Parallel()
	connector := plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityConnector}}
	view := viewManifest()
	tests := []struct {
		name      string
		image     string
		sigs      []plugins.SignerIdentity
		sourceErr error
		declared  *plugins.Manifest
		want      operator.Verification
	}{
		{
			name: "verified plugin", image: imageA, sigs: []plugins.SignerIdentity{publisherSignature()}, declared: &view,
			want: operator.Verification{
				Status: metav1.ConditionTrue, Reason: operator.ReasonVerifiedPlugin,
				Message: "verified plugin: the catalog pins this image, signed by its publisher example-publisher",
			},
		},
		{
			name: "signed by OpenFief's CI", image: imageB, sigs: []plugins.SignerIdentity{ciSignature()}, declared: &view,
			want: operator.Verification{
				Status: metav1.ConditionTrue, Reason: operator.ReasonTrustedSigner,
				Message: "unverified plugin signed by trusted signer openfief-ci (OpenFiefCI)",
			},
		},
		{
			name: "unsigned", image: imageB,
			want: operator.Verification{Status: metav1.ConditionFalse, Reason: operator.ReasonUnsigned},
		},
		{
			name: "untrusted signer", image: imageB, sigs: []plugins.SignerIdentity{strangerSignature()},
			want: operator.Verification{Status: metav1.ConditionFalse, Reason: operator.ReasonUntrustedSigner},
		},
		{
			name: "declared manifest differs from the signed label", image: imageA, sigs: []plugins.SignerIdentity{publisherSignature()},
			declared: &connector,
			want:     operator.Verification{Status: metav1.ConditionFalse, Reason: operator.ReasonManifestMismatch},
		},
		{
			name: "signed image without a manifest label", image: imageB, sourceErr: plugins.ErrManifestLabel,
			want: operator.Verification{Status: metav1.ConditionFalse, Reason: operator.ReasonInvalidManifestLabel},
		},
		{
			name: "registry unreachable", image: imageB, sourceErr: errUnreachable,
			want: operator.Verification{Status: metav1.ConditionUnknown, Reason: operator.ReasonRegistryUnavailable},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			img := mustImage(t, tt.image)
			src := &fakeSource{
				images: map[string]plugins.SignedImage{tt.image: {Image: img, Signatures: tt.sigs, Manifest: viewManifest()}},
				errs:   map[string]error{},
			}
			if tt.sourceErr != nil {
				src.errs[tt.image] = tt.sourceErr
			}
			got := operator.NewPolicyVerifier(src, testPolicy(t)).VerifyImage(t.Context(), img, tt.declared)
			if tt.want.Message == "" {
				got.Message = ""
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("VerifyImage (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPolicyVerifierCache(t *testing.T) {
	t.Parallel()
	signed, unsigned := mustImage(t, imageA), mustImage(t, imageB)
	src := &fakeSource{images: map[string]plugins.SignedImage{
		imageA: {Image: signed, Signatures: []plugins.SignerIdentity{ciSignature()}, Manifest: viewManifest()},
		imageB: {Image: unsigned, Manifest: viewManifest()},
	}}
	v := operator.NewPolicyVerifier(src, testPolicy(t))
	for range 2 {
		if got := v.VerifyImage(t.Context(), signed, nil); !got.Verified() {
			t.Fatalf("signed image: %+v", got)
		}
		if got := v.VerifyImage(t.Context(), unsigned, nil); got.Verified() {
			t.Fatalf("unsigned image verified: %+v", got)
		}
	}
	if src.reads != 3 {
		t.Errorf("registry reads = %d, want 3: the signed image once, the unsigned one every time", src.reads)
	}
	// A running plugin stays verified while its registry is down.
	src.mu.Lock()
	src.images = nil
	src.mu.Unlock()
	if got := v.VerifyImage(t.Context(), signed, nil); !got.Verified() {
		t.Errorf("cached signed image while the registry is down: %+v", got)
	}
}

func TestUnimplementedVerifier(t *testing.T) {
	t.Parallel()
	got := operator.UnimplementedVerifier{}.VerifyImage(t.Context(), mustImage(t, imageA), nil)
	if got.Verified() || got.Status != metav1.ConditionUnknown || got.Reason != operator.ReasonNotImplemented {
		t.Errorf("UnimplementedVerifier = %+v, want Unknown/NotImplemented", got)
	}
}
