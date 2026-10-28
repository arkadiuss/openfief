package cosign

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// SignPredicateType is the in-toto predicate type of a `cosign sign`
// signature in the bundle format. A bundle with any other predicate, such as
// an SBOM attestation by the same signer, is not a signature of the image.
const SignPredicateType = "https://sigstore.dev/cosign/sign/v1"

// sha256Prefix starts every image digest plugins.ImageRef accepts.
const sha256Prefix = "sha256:"

// Verifier finds and verifies an image's signatures and reads its manifest
// label. It is safe for concurrent use.
type Verifier struct {
	// keyless verifies certificate-signed bundles against the trust root.
	keyless *verify.Verifier
	// identities are the policy's keyless signers, for sigstore-go's
	// certificate identity check.
	identities []verify.CertificateIdentity
	// keys verify bundles signed with an owner's or publisher's key.
	keys   []keyVerifier
	remote []remote.Option
}

// keyVerifier verifies bundles signed with one trusted public key.
type keyVerifier struct {
	identity plugins.SignerIdentity
	verifier *verify.Verifier
}

// New returns a Verifier for the policy's signers, checking keyless
// signatures against trusted (normally TrustedRoot). The remote options set
// how the registry is reached, such as its transport.
func New(policy *plugins.TrustPolicy, trusted root.TrustedMaterial, opts ...remote.Option) (*Verifier, error) {
	return newVerifier(policy, trusted, []verify.VerifierOption{
		verify.WithSignedCertificateTimestamps(1),
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1),
	}, opts)
}

// newVerifier builds a Verifier with the given keyless verifier options.
func newVerifier(policy *plugins.TrustPolicy, trusted root.TrustedMaterial, keylessOpts []verify.VerifierOption, opts []remote.Option) (*Verifier, error) {
	keyless, err := verify.NewVerifier(trusted, keylessOpts...)
	if err != nil {
		return nil, fmt.Errorf("keyless verifier: %w", err)
	}
	v := &Verifier{keyless: keyless, remote: opts}
	for _, s := range policy.Signers() {
		if s.Identity.IsKey() {
			kv, err := newKeyVerifier(s.Identity)
			if err != nil {
				return nil, fmt.Errorf("signer %q: %w", s.Name, err)
			}
			v.keys = append(v.keys, kv)
			continue
		}
		id, err := certificateIdentity(s.Identity)
		if err != nil {
			return nil, fmt.Errorf("signer %q: %w", s.Name, err)
		}
		v.identities = append(v.identities, id)
	}
	return v, nil
}

// certificateIdentity is a keyless signer as sigstore-go's identity check:
// the exact issuer, and the exact subject or the anchored subject pattern.
func certificateIdentity(id plugins.SignerIdentity) (verify.CertificateIdentity, error) {
	pattern := ""
	if id.SubjectPattern != "" {
		pattern = `^(?:` + id.SubjectPattern + `)$`
	}
	ci, err := verify.NewShortCertificateIdentity(id.Issuer, "", id.Subject, pattern)
	if err != nil {
		return verify.CertificateIdentity{}, fmt.Errorf("certificate identity: %w", err)
	}
	return ci, nil
}

// newKeyVerifier builds a verifier that trusts one public key, with no
// expiry, and needs no timestamp: an owner's key signature need not be in a
// transparency log.
func newKeyVerifier(id plugins.SignerIdentity) (keyVerifier, error) {
	pub, err := cryptoutils.UnmarshalPEMToPublicKey([]byte(id.PublicKey))
	if err != nil {
		return keyVerifier{}, fmt.Errorf("public key: %w", err)
	}
	sv, err := signature.LoadDefaultVerifier(pub)
	if err != nil {
		return keyVerifier{}, fmt.Errorf("public key: %w", err)
	}
	key := root.NewExpiringKey(sv, time.Time{}, time.Time{})
	material := root.NewTrustedPublicKeyMaterial(func(string) (root.TimeConstrainedVerifier, error) { return key, nil })
	v, err := verify.NewVerifier(material, verify.WithNoObserverTimestamps())
	if err != nil {
		return keyVerifier{}, fmt.Errorf("key verifier: %w", err)
	}
	return keyVerifier{identity: id, verifier: v}, nil
}

// Inspect reads an image from its registry: the identities of the trusted
// signers whose signatures over its digest verify, and its manifest label.
// The label is read only from a signed image. A missing or invalid label
// fails with plugins.ErrManifestLabel; a registry that cannot be read fails
// with another error, worth retrying.
func (v *Verifier) Inspect(ctx context.Context, image plugins.ImageRef) (plugins.SignedImage, error) {
	d, err := name.NewDigest(image.String())
	if err != nil {
		return plugins.SignedImage{}, fmt.Errorf("image %s: %w", image, err)
	}
	opts := append([]remote.Option{remote.WithContext(ctx)}, v.remote...)
	digest, err := hex.DecodeString(strings.TrimPrefix(image.Digest(), sha256Prefix))
	if err != nil {
		return plugins.SignedImage{}, fmt.Errorf("image %s: %w", image, err)
	}
	bundles, err := readBundles(d, opts)
	if err != nil {
		return plugins.SignedImage{}, err
	}
	out := plugins.SignedImage{Image: image}
	for _, b := range bundles {
		if id, ok := v.verifyBundle(b, digest); ok {
			out.Signatures = append(out.Signatures, id)
		}
	}
	if len(out.Signatures) == 0 {
		// Unsigned: what it asks for does not matter.
		return out, nil
	}
	if out.Manifest, err = readManifestLabel(d, opts); err != nil {
		return plugins.SignedImage{}, err
	}
	return out, nil
}

// verifyBundle verifies one bundle as a `cosign sign` signature over the
// image digest by a trusted signer, and returns who made it.
func (v *Verifier) verifyBundle(b *bundle.Bundle, digest []byte) (plugins.SignerIdentity, bool) {
	content, err := b.VerificationContent()
	if err != nil {
		return plugins.SignerIdentity{}, false
	}
	artifact := verify.WithArtifactDigest("sha256", digest)
	if content.Certificate() != nil {
		return v.verifyKeyless(b, artifact)
	}
	for _, k := range v.keys {
		res, err := k.verifier.Verify(b, verify.NewPolicy(artifact, verify.WithKey()))
		if err == nil && isSignature(res) {
			return k.identity, true
		}
	}
	return plugins.SignerIdentity{}, false
}

// verifyKeyless verifies a certificate-signed bundle against the trust root
// and the policy's keyless signers, and returns the certificate's identity.
func (v *Verifier) verifyKeyless(b *bundle.Bundle, artifact verify.ArtifactPolicyOption) (plugins.SignerIdentity, bool) {
	if len(v.identities) == 0 {
		return plugins.SignerIdentity{}, false
	}
	opts := make([]verify.PolicyOption, 0, len(v.identities))
	for i := range v.identities {
		opts = append(opts, verify.WithCertificateIdentity(v.identities[i]))
	}
	res, err := v.keyless.Verify(b, verify.NewPolicy(artifact, opts...))
	if err != nil || !isSignature(res) || res.Signature == nil || res.Signature.Certificate == nil {
		return plugins.SignerIdentity{}, false
	}
	cert := res.Signature.Certificate
	return plugins.SignerIdentity{Issuer: cert.Issuer, Subject: cert.SubjectAlternativeName}, true
}

// isSignature reports whether a verified bundle is a `cosign sign`
// signature rather than some other statement about the image.
func isSignature(res *verify.VerificationResult) bool {
	return res != nil && res.Statement != nil && res.Statement.GetPredicateType() == SignPredicateType
}

// errRegistry wraps every failure to read the registry.
var errRegistry = errors.New("registry")
