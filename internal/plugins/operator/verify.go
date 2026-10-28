package operator

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// ImageVerifier checks a plugin image's cosign signature against the
// owner's signing policy (D-17). The operator deploys an image only when its
// Verification has Status True.
type ImageVerifier interface {
	VerifyImage(ctx context.Context, image plugins.ImageRef) Verification
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

// ReasonNotImplemented is the SignatureVerified reason while no verifier
// exists.
const ReasonNotImplemented = "NotImplemented"

// UnimplementedVerifier is the seam where cosign verification goes. It
// verifies nothing and reports every image as Unknown, so the operator
// deploys no plugin until a real verifier replaces it: D-17 requires
// verified signatures, and a verifier that always succeeded would hide that.
type UnimplementedVerifier struct{}

// VerifyImage reports image as not verified.
func (UnimplementedVerifier) VerifyImage(context.Context, plugins.ImageRef) Verification {
	return Verification{
		Status: metav1.ConditionUnknown,
		Reason: ReasonNotImplemented,
		Message: "cosign signature verification (D-17) is not implemented yet; " +
			"the operator deploys no plugin image until it is",
	}
}
