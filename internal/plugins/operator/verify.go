package operator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// ImageVerifier checks a plugin image's cosign signature against the trust
// policy (D-17, D-45, D-56) and, when declared is not nil, that the image's
// signed manifest label is the manifest the Plugin declares (D-42). The
// operator deploys an image only when its Verification has Status True.
type ImageVerifier interface {
	VerifyImage(ctx context.Context, image plugins.ImageRef, declared *plugins.Manifest) Verification
}

// Verification is the outcome of an image signature check, in the shape of
// the SignatureVerified condition: True verified, False rejected, Unknown
// not checked or not checkable.
type Verification struct {
	Status  metav1.ConditionStatus
	Reason  string
	Message string
}

// Verified reports whether the image may be deployed.
func (v Verification) Verified() bool { return v.Status == metav1.ConditionTrue }

// Reasons of the SignatureVerified condition.
const (
	// ReasonNotImplemented: no verifier is configured.
	ReasonNotImplemented = "NotImplemented"
	// ReasonVerifiedPlugin: the catalog pins the image and its publisher
	// signed it (D-44). The condition is True.
	ReasonVerifiedPlugin = "Verified"
	// ReasonTrustedSigner: a trusted signer signed the image, but it is not
	// a verified plugin (D-45). The condition is True.
	ReasonTrustedSigner = "SignedByTrustedSigner"
	// ReasonUnsigned: no valid signature by a trusted signer.
	ReasonUnsigned = "Unsigned"
	// ReasonUntrustedSigner: signed, but by no trusted signer.
	ReasonUntrustedSigner = "UntrustedSigner"
	// ReasonManifestMismatch: spec.manifest is not the image's signed
	// manifest label.
	ReasonManifestMismatch = "ManifestMismatch"
	// ReasonInvalidManifestLabel: the signed image has no readable manifest
	// label.
	ReasonInvalidManifestLabel = "InvalidManifestLabel"
	// ReasonRegistryUnavailable: the registry could not be read; the check
	// is retried.
	ReasonRegistryUnavailable = "RegistryUnavailable"
)

// UnimplementedVerifier verifies nothing and reports every image as Unknown,
// so an operator built with it deploys no plugin: D-17 requires verified
// signatures, and a verifier that always succeeded would hide that.
type UnimplementedVerifier struct{}

// VerifyImage reports image as not verified.
func (UnimplementedVerifier) VerifyImage(context.Context, plugins.ImageRef, *plugins.Manifest) Verification {
	return Verification{
		Status: metav1.ConditionUnknown,
		Reason: ReasonNotImplemented,
		Message: "no signature verifier is configured; " +
			"the operator deploys no plugin image without one",
	}
}

// SignedImageSource reads an image from its registry: the trusted
// signatures over its digest that verify, and its signed manifest label.
// cosign.Verifier implements it.
type SignedImageSource interface {
	Inspect(ctx context.Context, image plugins.ImageRef) (plugins.SignedImage, error)
}

// maxCachedImages bounds the verified-image cache; it is emptied when full.
const maxCachedImages = 256

// PolicyVerifier is the operator's ImageVerifier: it reads an image through
// a SignedImageSource and lets the trust policy decide (D-45, D-56, D-57).
// There is no way to deploy an image the policy refuses.
//
// A signed image's registry content is cached by digest, so a running plugin
// keeps its verified status while the registry is unreachable, and a
// reconcile does not refetch it. The policy is fixed for the operator's
// lifetime, so the cache never goes stale; an unsigned or unreadable image is
// never cached.
type PolicyVerifier struct {
	source SignedImageSource
	policy *plugins.TrustPolicy

	mu    sync.Mutex
	cache map[string]plugins.SignedImage
}

// NewPolicyVerifier returns a PolicyVerifier that reads images through
// source and decides with policy.
func NewPolicyVerifier(source SignedImageSource, policy *plugins.TrustPolicy) *PolicyVerifier {
	return &PolicyVerifier{source: source, policy: policy, cache: map[string]plugins.SignedImage{}}
}

// VerifyImage checks image against the trust policy and, when declared is
// not nil, its signed manifest label against declared.
func (v *PolicyVerifier) VerifyImage(ctx context.Context, image plugins.ImageRef, declared *plugins.Manifest) Verification {
	signed, err := v.inspect(ctx, image)
	if err != nil {
		if errors.Is(err, plugins.ErrManifestLabel) {
			return Verification{Status: metav1.ConditionFalse, Reason: ReasonInvalidManifestLabel, Message: err.Error()}
		}
		return Verification{Status: metav1.ConditionUnknown, Reason: ReasonRegistryUnavailable, Message: err.Error()}
	}
	d, err := v.policy.Decide(&signed, declared)
	if err != nil {
		return Verification{Status: metav1.ConditionFalse, Reason: refusalReason(err), Message: err.Error()}
	}
	switch d.Trust {
	case plugins.Verified:
		return Verification{
			Status: metav1.ConditionTrue, Reason: ReasonVerifiedPlugin,
			Message: "verified plugin: the catalog pins this image, signed by its publisher " + d.Signer.Name,
		}
	case plugins.Unverified:
	}
	return Verification{
		Status: metav1.ConditionTrue, Reason: ReasonTrustedSigner,
		Message: "unverified plugin signed by trusted signer " + d.Signer.Name + " (" + string(d.Signer.Kind) + ")",
	}
}

// refusalReason names why the trust policy refused an image.
func refusalReason(err error) string {
	switch {
	case errors.Is(err, plugins.ErrManifestMismatch):
		return ReasonManifestMismatch
	case errors.Is(err, plugins.ErrUntrustedSigner):
		return ReasonUntrustedSigner
	default:
		return ReasonUnsigned
	}
}

// inspect reads image through the cache.
func (v *PolicyVerifier) inspect(ctx context.Context, image plugins.ImageRef) (plugins.SignedImage, error) {
	key := image.String()
	v.mu.Lock()
	signed, ok := v.cache[key]
	v.mu.Unlock()
	if ok {
		return signed, nil
	}
	signed, err := v.source.Inspect(ctx, image)
	if err != nil {
		return plugins.SignedImage{}, fmt.Errorf("check image: %w", err)
	}
	if len(signed.Signatures) > 0 {
		v.mu.Lock()
		if len(v.cache) >= maxCachedImages {
			clear(v.cache)
		}
		v.cache[key] = signed
		v.mu.Unlock()
	}
	return signed, nil
}
